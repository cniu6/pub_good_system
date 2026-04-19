package middleware

import (
	"bytes"
	"encoding/json"
	"fst/backend/app/models"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxOperationLogCapturedResponseBytes = 8 * 1024

// responseWriter 自定义响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	maxCapture int
	written    int64
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.written += int64(len(b))
	if w.body != nil && w.maxCapture > 0 {
		remaining := w.maxCapture - w.body.Len()
		if remaining > 0 {
			if len(b) > remaining {
				_, _ = w.body.Write(b[:remaining])
			} else {
				_, _ = w.body.Write(b)
			}
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func newResponseWriter(writer gin.ResponseWriter, maxCapture int) *responseWriter {
	return &responseWriter{
		ResponseWriter: writer,
		body:           bytes.NewBufferString(""),
		maxCapture:     maxCapture,
	}
}

// sensitiveLogFields 审计日志中需要脱敏的字段。
// 注意：这里使用“包含匹配”以覆盖各种命名风格（如 new_password / oldPassword）。
var sensitiveLogFields = []string{
	"password", "passwd", "pwd",
	"token", "accesstoken", "refreshtoken", "apikey", "api_key", "authorization", "cookie", "session", "jwt",
	"secret", "sign", "signature",
	"captcha", "code", "verification",
	"mobile", "phone", "email",
	"certificate_no", "certificateno", "id_card", "idcard",
	"account_no", "accountno", "card_no", "cardno",
}

// sanitizeLogBody 对请求/响应体做关键字段脱敏，并控制长度。
// 优先尝试 JSON 脱敏；如果不是 JSON，则退化为正则式关键字段遮蔽。
func sanitizeLogBody(raw string, limit int) string {
	if raw == "" {
		return raw
	}

	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var payload any
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			masked := maskSensitiveValues(payload)
			if data, err := json.Marshal(masked); err == nil {
				return truncateForLog(string(data), limit)
			}
		}
	}

	return truncateForLog(raw, limit)
}

// maskSensitiveValues 递归遍历 JSON 结构，对敏感键值做遮蔽。
func maskSensitiveValues(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, inner := range v {
			if isSensitiveLogField(key) {
				out[key] = "***"
				continue
			}
			out[key] = maskSensitiveValues(inner)
		}
		return out
	case []any:
		for i := range v {
			v[i] = maskSensitiveValues(v[i])
		}
		return v
	default:
		return v
	}
}

// isSensitiveLogField 判断字段名是否涉及敏感数据。
func isSensitiveLogField(key string) bool {
	normalized := strings.ToLower(key)
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	for _, keyword := range sensitiveLogFields {
		clean := strings.ReplaceAll(keyword, "_", "")
		if clean == "" {
			continue
		}
		if strings.Contains(normalized, clean) {
			return true
		}
	}
	return false
}

func truncateForLog(raw string, limit int) string {
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	return raw[:limit] + "...(truncated)"
}

// OperationLogMiddleware 操作日志中间件
// 记录用户的操作日志
func OperationLogMiddleware(module string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start_time := time.Now()

		// 获取请求体
		var request_body string
		if c.Request.Body != nil {
			body_bytes, _ := io.ReadAll(c.Request.Body)
			request_body = string(body_bytes)
			// 重新设置请求体，以便后续处理
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body_bytes))
		}

		// 包装响应写入器
		blw := newResponseWriter(c.Writer, maxOperationLogCapturedResponseBytes)
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start_time).Milliseconds()

		// 获取用户信息
		var user_id uint64
		var username string
		if uid, exists := c.Get("userID"); exists {
			if parsedUID, ok := uid.(uint64); ok {
				user_id = parsedUID
			}
		}
		if uname, exists := c.Get("username"); exists {
			if parsedUsername, ok := uname.(string); ok {
				username = parsedUsername
			}
		}

		// 对请求/响应体做脱敏与长度限制，避免密码、token、验证码等敏感字段落库
		request_body = sanitizeLogBody(request_body, 2000)
		response_body := sanitizeLogBody(blw.body.String(), 2000)

		// 创建日志记录
		record := &models.OperationLog{
			UserID:       user_id,
			Username:     username,
			Module:       module,
			Action:       action,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestBody:  &request_body,
			ResponseBody: &response_body,
			StatusCode:   c.Writer.Status(),
			Duration:     int(duration),
		}

		// 异步保存日志
		go func(entry *models.OperationLog) {
			if err := models.CreateOperationLog(entry); err != nil {
				log.Printf("[OperationLog] 保存失败: %v", err)
			}
		}(record)
	}
}

// SimpleLogMiddleware 简单日志中间件
// 只记录基本信息，不记录请求/响应体
func SimpleLogMiddleware(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start_time := time.Now()

		c.Next()

		duration := time.Since(start_time).Milliseconds()

		var user_id uint64
		var username string
		if uid, exists := c.Get("userID"); exists {
			if parsedUID, ok := uid.(uint64); ok {
				user_id = parsedUID
			}
		}
		if uname, exists := c.Get("username"); exists {
			if parsedUsername, ok := uname.(string); ok {
				username = parsedUsername
			}
		}

		// 根据请求方法确定操作类型
		action := getActionByMethod(c.Request.Method)

		record := &models.OperationLog{
			UserID:     user_id,
			Username:   username,
			Module:     module,
			Action:     action,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   int(duration),
		}

		go func(entry *models.OperationLog) {
			if err := models.CreateOperationLog(entry); err != nil {
				log.Printf("[OperationLog] 保存失败: %v", err)
			}
		}(record)
	}
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

