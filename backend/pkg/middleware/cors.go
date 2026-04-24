package middleware

import (
	"fst/backend/pkg/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// 允许的跨域请求头白名单。
// 仅在放行时返回，避免完全反射客户端请求头带来被滥用的风险。
const defaultAllowHeaders = "Origin, Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-Idempotency-Key, X-Geetest-Lot-Number, X-Geetest-Captcha-Output, X-Geetest-Pass-Token, X-Geetest-Gen-Time, X-Geetest-Captcha-Id"

// isOriginAllowed 判断请求 Origin 是否在允许白名单内。
// 若配置为 "*" 表示允许任意 Origin（但此时不会附带 Allow-Credentials）。
func isOriginAllowed(origin, corsOrigins string) (allowed bool, wildcard bool) {
	if origin == "" {
		return false, false
	}
	corsOrigins = strings.TrimSpace(corsOrigins)
	if corsOrigins == "*" {
		return true, true
	}
	if corsOrigins == "" {
		// 未配置跨域白名单：仅在非生产模式放行，避免生产下出现 CSRF 风险
		return !config.IsProductionMode(), false
	}
	for _, allowedOrigin := range strings.Split(corsOrigins, ",") {
		if strings.TrimSpace(allowedOrigin) == origin {
			return true, false
		}
	}
	return false, false
}

// CorsMiddleware 处理跨域请求。
// 关键约束：只有命中白名单且非通配 Origin 时才会附带 Allow-Credentials，
// 避免 "*" / 任意反射 + 凭证并存这一经典 CSRF 风险。
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsOrigins := ""
		if cfg := config.GlobalConfig; cfg != nil {
			corsOrigins = cfg.CorsOrigins
		}
		origin := c.GetHeader("Origin")

		allowed, wildcard := isOriginAllowed(origin, corsOrigins)
		if allowed {
			if wildcard {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Header("Access-Control-Allow-Headers", defaultAllowHeaders)
		c.Header("Access-Control-Max-Age", "3600") // 预检请求缓存1小时

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

