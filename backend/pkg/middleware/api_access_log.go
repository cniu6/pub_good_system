package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// API 访问日志 body 截断与操作日志共用常量（见 log_sanitize.go）：
// MEDIUMTEXT ≈ 16MB，实际存库上限 maxLogStoredBodyBytes = 64KB，并标记「已截断」。
const apiAccessLogCleanupInterval = 30 * time.Second

var apiAccessLogRequestSeq atomic.Uint64
var apiAccessLogCleanupNextAt atomic.Int64

// adminAPIBasePath 返回管理端 API 完整前缀 /api/v1{ADMIN_API_PATH}
func adminAPIBasePath() string {
	apiPath := "/admin"
	if config.GlobalConfig != nil {
		apiPath = config.NormalizeAdminAPIPath(config.GlobalConfig.AdminAPIPath)
	}
	return "/api/v1" + apiPath
}

func shouldSkipAPIAccessLog(path string) bool {
	if !strings.HasPrefix(path, "/api/") {
		return true
	}
	// pprof 二进制输出体积大，跳过访问日志
	if strings.HasPrefix(path, adminAPIBasePath()+"/debug/pprof") {
		return true
	}
	return false
}

func resolveAPIScene(path string) string {
	switch {
	case strings.HasPrefix(path, adminAPIBasePath()):
		return "admin"
	case strings.HasPrefix(path, "/api/v1/user"):
		return "user"
	case strings.HasPrefix(path, "/api/v1/public"):
		return "public"
	case strings.HasPrefix(path, "/api/v1/system"):
		return "system"
	default:
		return "plugin"
	}
}

func shouldReadRequestBody(c *gin.Context) bool {
	if c.Request.Body == nil {
		return false
	}
	if c.Request.ContentLength > int64(maxLogReadableBodyBytes) {
		return false
	}
	contentType := strings.ToLower(c.GetHeader("Content-Type"))
	if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/octet-stream") {
		return false
	}
	return true
}

func sanitizeQueryString(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return truncateForLog(raw, 1000)
	}
	payload := make(map[string]any, len(values))
	for key, items := range values {
		if len(items) == 1 {
			payload[key] = items[0]
			continue
		}
		arr := make([]any, 0, len(items))
		for _, item := range items {
			arr = append(arr, item)
		}
		payload[key] = arr
	}
	masked := maskSensitiveValues(payload)
	data, err := json.Marshal(masked)
	if err != nil {
		return truncateForLog(raw, 1000)
	}
	return truncateForLog(string(data), 1000)
}

func sanitizePathParams(params gin.Params) *string {
	if len(params) == 0 {
		return nil
	}
	payload := make(map[string]any, len(params))
	for _, item := range params {
		payload[item.Key] = item.Value
	}
	masked := maskSensitiveValues(payload)
	data, err := json.Marshal(masked)
	if err != nil {
		fallback := truncateForLog(fmt.Sprintf("%v", params), 1000)
		return &fallback
	}
	value := truncateForLog(string(data), 2000)
	return &value
}

func sanitizeRequestHeaders(headers map[string][]string) *string {
	if len(headers) == 0 {
		return nil
	}
	payload := make(map[string]any, len(headers))
	for key, items := range headers {
		if len(items) == 0 {
			payload[key] = ""
			continue
		}
		if isSensitiveLogField(key) {
			if len(items) == 1 {
				payload[key] = "***"
			} else {
				masked := make([]string, len(items))
				for i := range masked {
					masked[i] = "***"
				}
				payload[key] = masked
			}
			continue
		}
		if len(items) == 1 {
			payload[key] = truncateForLog(items[0], maxLogHeaderValueLength)
			continue
		}
		values := make([]string, 0, len(items))
		for _, item := range items {
			values = append(values, truncateForLog(item, maxLogHeaderValueLength))
		}
		payload[key] = values
	}
	masked := maskSensitiveValues(payload)
	data, err := json.Marshal(masked)
	if err != nil {
		fallback := truncateForLog("[request headers omitted: marshal failed]", maxLogStoredHeadersLength)
		return &fallback
	}
	value := truncateForLog(string(data), maxLogStoredHeadersLength)
	return &value
}

func normalizeContentType(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if idx := strings.Index(trimmed, ";"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return strings.TrimSpace(trimmed)
}

func isTruthyStreamFlag(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func requestEnablesStreaming(rawQuery, requestBody string) bool {
	values, err := url.ParseQuery(rawQuery)
	if err == nil {
		if isTruthyStreamFlag(values.Get("stream")) || isTruthyStreamFlag(values.Get("streaming")) {
			return true
		}
	}

	trimmed := strings.TrimSpace(requestBody)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[")) {
		return false
	}

	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return false
	}

	object, ok := payload.(map[string]any)
	if !ok {
		return false
	}

	for _, key := range []string{"stream", "streaming"} {
		value, exists := object[key]
		if !exists {
			continue
		}
		switch v := value.(type) {
		case bool:
			if v {
				return true
			}
		case string:
			if isTruthyStreamFlag(v) {
				return true
			}
		case float64:
			if v != 0 {
				return true
			}
		}
	}

	return false
}

func resolveAPITransport(c *gin.Context, requestBody, responseContentType string) string {
	requestContentType := normalizeContentType(c.GetHeader("Content-Type"))
	responseContentType = normalizeContentType(responseContentType)
	upgrade := strings.ToLower(strings.TrimSpace(c.GetHeader("Upgrade")))
	connection := strings.ToLower(strings.TrimSpace(c.GetHeader("Connection")))
	accept := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept")))
	responseUpgrade := strings.ToLower(strings.TrimSpace(c.Writer.Header().Get("Upgrade")))

	if strings.Contains(upgrade, "websocket") || strings.Contains(responseUpgrade, "websocket") || c.GetHeader("Sec-WebSocket-Key") != "" || (strings.Contains(connection, "upgrade") && c.Writer.Status() == 101) {
		return "websocket"
	}
	if responseContentType == "text/event-stream" || strings.Contains(accept, "text/event-stream") {
		return "sse"
	}
	if strings.Contains(responseContentType, "stream") || strings.Contains(requestContentType, "stream") || strings.Contains(accept, "stream") || requestEnablesStreaming(c.Request.URL.RawQuery, requestBody) {
		return "stream"
	}
	return "http"
}

func resolveSourceIP(remoteAddr string) string {
	trimmed := strings.TrimSpace(remoteAddr)
	if trimmed == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return host
	}
	return trimmed
}

func newAPIAccessRequestID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		binary.BigEndian.PutUint64(raw[:8], uint64(time.Now().UnixNano()))
		binary.BigEndian.PutUint64(raw[8:], apiAccessLogRequestSeq.Add(1))
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	hexbuf := make([]byte, 36)
	hex.Encode(hexbuf[0:8], raw[0:4])
	hexbuf[8] = '-'
	hex.Encode(hexbuf[9:13], raw[4:6])
	hexbuf[13] = '-'
	hex.Encode(hexbuf[14:18], raw[6:8])
	hexbuf[18] = '-'
	hex.Encode(hexbuf[19:23], raw[8:10])
	hexbuf[23] = '-'
	hex.Encode(hexbuf[24:36], raw[10:16])
	return string(hexbuf)
}

func scheduleAPIAccessLogRetentionCleanup() {
	maxCount := services.GetGlobalAPILogRuntimeConfig().MaxCount
	if maxCount <= 0 {
		return
	}

	now := time.Now().UnixNano()
	nextAt := apiAccessLogCleanupNextAt.Load()
	if nextAt > now {
		return
	}

	scheduledAt := now + int64(apiAccessLogCleanupInterval)
	if !apiAccessLogCleanupNextAt.CompareAndSwap(nextAt, scheduledAt) {
		return
	}

	if _, err := models.CleanExcessAPIAccessLogs(maxCount); err != nil {
		apiAccessLogCleanupNextAt.Store(now)
		log.Printf("[APIAccessLog] 自动清理超限日志失败: %v", err)
	}
}

func sanitizeResponseBodyByType(raw, contentType, transport string, statusCode int) string {
	trimmedType := normalizeContentType(contentType)
	if transport == "websocket" && statusCode == 101 {
		return "[websocket upgrade handshake established; frame messages are not captured in HTTP access logs]"
	}
	if trimmedType == "" || strings.Contains(trimmedType, "json") || strings.Contains(trimmedType, "text") || strings.Contains(trimmedType, "xml") || strings.Contains(trimmedType, "javascript") || strings.Contains(trimmedType, "x-www-form-urlencoded") || trimmedType == "text/event-stream" || transport == "sse" || transport == "stream" {
		return sanitizeLogBody(raw, maxLogStoredBodyBytes)
	}
	if raw == "" {
		return ""
	}
	return "[response body omitted: non-text content]"
}

// APIAccessLogMiddleware 记录所有 API 请求的访问日志。
func APIAccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if shouldSkipAPIAccessLog(path) {
			c.Next()
			return
		}

		if !services.GetGlobalAPILogRuntimeConfig().Enabled {
			c.Next()
			return
		}

		requestID := newAPIAccessRequestID()
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		startTime := time.Now()
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		requestBody := ""
		if shouldReadRequestBody(c) {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				requestSize = int64(len(bodyBytes))
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		if requestBody == "" {
			contentType := strings.ToLower(c.GetHeader("Content-Type"))
			if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/octet-stream") {
				requestBody = "[request body omitted: multipart/binary content]"
			} else if c.Request.ContentLength > int64(maxLogReadableBodyBytes) {
				requestBody = "[request body omitted: too large]"
			}
		}

		blw := newResponseWriter(c.Writer, maxLogStoredBodyBytes)
		c.Writer = blw
		c.Next()

		duration := time.Since(startTime).Milliseconds()
		requestContentType := truncateForLog(normalizeContentType(c.GetHeader("Content-Type")), 255)
		responseContentType := truncateForLog(normalizeContentType(blw.Header().Get("Content-Type")), 255)
		transport := resolveAPITransport(c, requestBody, responseContentType)
		responseBody := sanitizeResponseBodyByType(blw.body.String(), responseContentType, transport, c.Writer.Status())
		requestBody = sanitizeLogBody(requestBody, maxLogStoredBodyBytes)
		requestHeaders := sanitizeRequestHeaders(c.Request.Header)
		pathParams := sanitizePathParams(c.Params)

		var userID uint64
		var username string
		var role string
		if uid, exists := c.Get("userID"); exists {
			if parsed, ok := uid.(uint64); ok {
				userID = parsed
			}
		}
		if uname, exists := c.Get("username"); exists {
			if parsed, ok := uname.(string); ok {
				username = parsed
			}
		}
		if r, exists := c.Get("role"); exists {
			if parsed, ok := r.(string); ok {
				role = parsed
			}
		}

		routePath := c.FullPath()
		if routePath == "" {
			routePath = path
		}
		handlerName := truncateForLog(c.HandlerName(), 255)
		sourceIP := truncateForLog(resolveSourceIP(c.Request.RemoteAddr), 45)
		xip := truncateForLog(c.GetHeader("X-IP"), 45)
		xForwardedFor := truncateForLog(c.GetHeader("X-Forwarded-For"), 512)
		xRealIP := truncateForLog(c.GetHeader("X-Real-IP"), 45)

		entry := &models.APIAccessLog{
			RequestID:           requestID,
			UserID:              userID,
			Username:            username,
			Role:                role,
			Scene:               resolveAPIScene(path),
			Method:              c.Request.Method,
			Transport:           transport,
			Protocol:            truncateForLog(c.Request.Proto, 32),
			Path:                path,
			RoutePath:           routePath,
			HandlerName:         handlerName,
			RequestContentType:  requestContentType,
			ResponseContentType: responseContentType,
			QueryString:         sanitizeQueryString(c.Request.URL.RawQuery),
			PathParams:          pathParams,
			IP:                  truncateForLog(c.ClientIP(), 45),
			SourceIP:            sourceIP,
			XIP:                 xip,
			XForwardedFor:       xForwardedFor,
			XRealIP:             xRealIP,
			UserAgent:           truncateForLog(c.Request.UserAgent(), 1000),
			Referer:             truncateForLog(c.Request.Referer(), 500),
			RequestHeaders:      requestHeaders,
			RequestBody:         &requestBody,
			ResponseBody:        &responseBody,
			StatusCode:          c.Writer.Status(),
			Duration:            int(duration),
			RequestSize:         requestSize,
			ResponseSize:        blw.written,
		}

		go func(item *models.APIAccessLog) {
			if err := models.CreateAPIAccessLog(item); err != nil {
				log.Printf("[APIAccessLog] 保存失败: %v", err)
				return
			}
			if err := models.RecordAPIAccessLogAggregate(item); err != nil {
				log.Printf("[APIAccessLog] 汇总更新失败: %v", err)
			}
			scheduleAPIAccessLogRetentionCleanup()
		}(entry)
	}
}
