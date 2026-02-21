package middleware

import (
	"fst/backend/internal/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware 处理跨域请求
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsOrigins := config.GlobalConfig.CorsOrigins
		origin := c.GetHeader("Origin")

		if origin != "" {
			if corsOrigins == "*" || corsOrigins == "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				allowedOrigins := strings.Split(corsOrigins, ",")
				for _, allowedOrigin := range allowedOrigins {
					if strings.TrimSpace(allowedOrigin) == origin {
						c.Header("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		
		// 动态允许所有请求头，解决极验等自定义头导致的跨域问题
		reqHeaders := c.GetHeader("Access-Control-Request-Headers")
		if reqHeaders != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-Geetest-Lot-Number, X-Geetest-Captcha-Output, X-Geetest-Pass-Token, X-Geetest-Gen-Time, X-Geetest-Captcha-Id")
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "3600") // 预检请求缓存1小时

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
