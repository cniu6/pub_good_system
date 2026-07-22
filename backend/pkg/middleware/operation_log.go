package middleware

import (
	"bytes"
	"encoding/json"
	"fst/backend/app/models"
	"fst/backend/pkg/panicsafe"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleLogMiddleware 操作日志中间件：记录模块/动作/路径，以及请求内容、响应内容、处理函数名。
// GET 无 body 时把 query 写入 request_body；响应体会捕获并截断后入库（不做字段脱敏）。
func SimpleLogMiddleware(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 采集请求内容：优先 body；GET/无 body 时用 query；过大/二进制则给说明文案
		requestBody := captureOperationRequestContent(c)

		// 包装 ResponseWriter 以捕获响应体（上限与存库截断一致）
		blw := newResponseWriter(c.Writer, maxLogStoredBodyBytes)
		c.Writer = blw

		c.Next()

		duration := time.Since(startTime).Milliseconds()

		var userID uint64
		var username string
		if uid, exists := c.Get("userID"); exists {
			if parsedUID, ok := uid.(uint64); ok {
				userID = parsedUID
			}
		}
		if uname, exists := c.Get("username"); exists {
			if parsedUsername, ok := uname.(string); ok {
				username = parsedUsername
			}
		}

		// 响应内容：安全截断（复用 API 访问日志的类型判断；管理端可见，不做字段脱敏）
		responseContentType := normalizeContentType(blw.Header().Get("Content-Type"))
		responseBody := sanitizeResponseBodyByType(blw.body.String(), responseContentType, "http", c.Writer.Status())
		requestBody = sanitizeOperationRequestContent(requestBody)

		handlerName := truncateForLog(c.HandlerName(), 255)

		reqPtr := requestBody
		respPtr := responseBody

		record := &models.OperationLog{
			UserID:       userID,
			Username:     username,
			Module:       module,
			Action:       getActionByMethod(c.Request.Method),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			IP:           c.ClientIP(),
			UserAgent:    truncateForLog(c.Request.UserAgent(), 1000),
			HandlerName:  handlerName,
			RequestBody:  &reqPtr,
			ResponseBody: &respPtr,
			StatusCode:   c.Writer.Status(),
			Duration:     int(duration),
		}

		panicsafe.Go("OperationLog.write", func() {
			if err := models.CreateOperationLog(record); err != nil {
				log.Printf("[OperationLog] 保存失败: %v", err)
			}
		})
	}
}

// captureOperationRequestContent 读取请求内容（可读完后还原 Body 供后续使用）。
// GET 等无 body 场景返回 query JSON；有 body 则读 body。
func captureOperationRequestContent(c *gin.Context) string {
	method := strings.ToUpper(c.Request.Method)
	rawQuery := c.Request.URL.RawQuery

	// GET/HEAD 或明确无 body：用 query 作为「请求内容」
	if method == "GET" || method == "HEAD" || c.Request.Body == nil || c.Request.ContentLength == 0 {
		if strings.TrimSpace(rawQuery) == "" {
			if method == "GET" || method == "HEAD" {
				return `{"query":{}}`
			}
			return ""
		}
		qs := sanitizeQueryString(rawQuery)
		if qs == "" {
			return `{"query":{}}`
		}
		return `{"query":` + qs + `}`
	}

	contentType := strings.ToLower(c.GetHeader("Content-Type"))
	if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/octet-stream") {
		return "[request body omitted: multipart/binary content]"
	}
	if c.Request.ContentLength > maxLogReadableBodyBytes {
		return "[request body omitted: too large]"
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxLogReadableBodyBytes+1))
	if err != nil {
		return ""
	}
	// 读完后还原，供后续 handler / 其它中间件使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if int64(len(bodyBytes)) > maxLogReadableBodyBytes {
		return "[request body omitted: too large]"
	}

	body := string(bodyBytes)
	// 同时有 query 时一并附上，方便排查
	if strings.TrimSpace(rawQuery) != "" {
		qs := sanitizeQueryString(rawQuery)
		if qs != "" && strings.TrimSpace(body) != "" {
			return `{"query":` + qs + `,"body":` + wrapAsJSONValue(body) + `}`
		}
		if qs != "" {
			return `{"query":` + qs + `}`
		}
	}
	return body
}

// wrapAsJSONValue 若 body 已是 JSON 则原样嵌入；否则当字符串转义。
func wrapAsJSONValue(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return `""`
	}
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		return trimmed
	}
	b, err := json.Marshal(trimmed)
	if err != nil {
		return `""`
	}
	return string(b)
}

func sanitizeOperationRequestContent(raw string) string {
	if raw == "" {
		return raw
	}
	// 说明性占位文案只做长度截断
	if strings.HasPrefix(raw, "[") && strings.Contains(raw, "omitted") {
		return truncateForLog(raw, maxLogStoredBodyBytes)
	}
	return sanitizeLogBody(raw, maxLogStoredBodyBytes, true)
}

func getActionByMethod(method string) string {
	switch method {
	case "GET":
		return "查询"
	case "POST":
		return "创建"
	case "PUT":
		return "更新"
	case "DELETE":
		return "删除"
	default:
		return "操作"
	}
}
