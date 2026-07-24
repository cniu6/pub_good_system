package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CtxRequestID 上下文中的请求 ID 键（与 API 访问日志共用）。
const CtxRequestID = "requestID"

// HeaderRequestID 对外统一响应头（也兼容读 X-Request-ID）。
const HeaderRequestID = "X-Request-Id"

// RequestIDMiddleware 轻量请求 ID：优先沿用客户端传入的 X-Request-Id / X-Request-ID，
// 否则生成 UUID，写入 context 与响应头，供访问日志、请求日志、panic 恢复引用。
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := strings.TrimSpace(c.GetHeader(HeaderRequestID))
		if rid == "" {
			rid = strings.TrimSpace(c.GetHeader("X-Request-ID"))
		}
		if rid == "" {
			rid = newAPIAccessRequestID()
		}
		c.Set(CtxRequestID, rid)
		c.Writer.Header().Set(HeaderRequestID, rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

// GetRequestID 从 gin.Context 取请求 ID，没有则空串。
func GetRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(CtxRequestID); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}
