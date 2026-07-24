package middleware

import (
	"fst/backend/utils"
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 自定义 panic 恢复：在默认堆栈日志上附带 request_id / 路由信息，
// 并用统一 {code,message,data} 协议回写 500（避免 gin 默认纯文本/HTML 响应）。
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		rid := GetRequestID(c)
		method := ""
		path := ""
		ip := ""
		if c.Request != nil {
			method = c.Request.Method
			path = c.Request.URL.Path
			ip = c.ClientIP()
		}
		log.Printf("[PANIC] request_id=%s method=%s path=%s ip=%s err=%v\n%s",
			rid, method, path, ip, recovered, debug.Stack())

		// 响应可能已部分写出（如流式/WS 升级失败后），此时不要再写 JSON
		if c.Writer != nil && !c.Writer.Written() {
			utils.Fail(c, 500, "Internal Server Error")
		}
		c.Abort()
	})
}
