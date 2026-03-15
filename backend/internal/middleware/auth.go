package middleware

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return AuthMiddlewareForGuard(utils.UserAuthGuard)
}

// AuthMiddlewareForGuard 支持传入一个或多个可接受的 guard。
// 例如用户路由传 ("user","admin") 表示管理员token也能访问用户接口。
func AuthMiddlewareForGuard(acceptGuards ...string) gin.HandlerFunc {
	if len(acceptGuards) == 0 {
		acceptGuards = []string{utils.UserAuthGuard}
	}
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Fail(c, 401, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Fail(c, 401, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		// 依次尝试每种 guard 解析 token，匹配则通过
		var claims *utils.Claims
		var parseErr error
		for _, guard := range acceptGuards {
			claims, parseErr = utils.ParseTokenForGuard(parts[1], guard)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			utils.Fail(c, 401, "Invalid or expired token")
			c.Abort()
			return
		}

		// 用 token 中实际的 guard 查询会话
		actualGuard := claims.AuthGuard
		if actualGuard == "" {
			actualGuard = utils.UserAuthGuard
		}
		active, err := models.IsUserSessionActive(claims.UserID, actualGuard, utils.HashToken(parts[1]))
		if err != nil || !active {
			utils.Fail(c, 401, "Session expired or revoked")
			c.Abort()
			return
		}

		if user, err := models.GetUserByID(claims.UserID); err == nil && user != nil {
			c.Set("username", user.Username)
		}
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("authGuard", actualGuard)
		c.Next()
	}
}

// AdminOnly 验证用户是否为管理员
// 这是核心安全防护：即使前端守卫被绕过，后端也会拦截非管理员请求
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		authGuard, _ := c.Get("authGuard")
		if authGuard != utils.AdminAuthGuard {
			utils.Fail(c, 403, "Admin access only")
			c.Abort()
			return
		}

		role, exists := c.Get("role")
		if !exists || role != "admin" {
			// 记录可疑访问（可用于安全审计）
			userID, _ := c.Get("userID")
			path := c.Request.URL.Path
			method := c.Request.Method
			clientIP := c.ClientIP()

			// 输出安全警告日志
			gin.DefaultWriter.Write([]byte(
				fmt.Sprintf("[SECURITY WARNING] %s | Non-admin access attempt | UserID: %v | IP: %s | Method: %s | Path: %s\n",
					time.Now().Format(time.RFC3339), userID, clientIP, method, path),
			))

			utils.Fail(c, 403, "Admin access only")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 通用角色验证中间件
// 可用于验证多种角色权限
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.Fail(c, 403, "Role not found")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			utils.Fail(c, 403, "Invalid role type")
			c.Abort()
			return
		}

		allowed := false
		for _, r := range allowedRoles {
			if roleStr == r {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.Fail(c, 403, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
