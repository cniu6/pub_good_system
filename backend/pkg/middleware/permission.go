package middleware

import (
	"fst/backend/app/models"
	"fst/backend/utils"
	"log"

	"github.com/gin-gonic/gin"
)

// RequirePermission 要求当前管理员具备指定权限码。
// 兼容：users.role == "admin" 时直接放行（现有管理员不受影响）。
// 否则查 user_roles → role_permissions。
func RequirePermission(permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get("role")
		roleStr, _ := roleVal.(string)
		// 系统级 admin（users.role）旁路：兼容旧逻辑
		if roleStr == utils.AdminAuthGuard {
			c.Next()
			return
		}

		userIDVal, ok := c.Get("userID")
		if !ok {
			utils.Fail(c, 401, "Unauthorized")
			c.Abort()
			return
		}
		userID, ok := userIDVal.(uint64)
		if !ok || userID == 0 {
			utils.Fail(c, 401, "Unauthorized")
			c.Abort()
			return
		}

		okPerm, err := models.UserHasPermissionCode(userID, permCode)
		if err != nil {
			log.Printf("[RBAC] permission check failed user_id=%d code=%s err=%v", userID, permCode, err)
			utils.Fail(c, 500, "Permission check failed")
			c.Abort()
			return
		}
		if !okPerm {
			utils.Fail(c, 403, "Insufficient permissions: "+permCode)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRecentTOTP 敏感操作可选二次校验。
// 若用户未启用 TOTP 则跳过；已启用时要求请求头 X-Totp-Code 验证通过。
// MVP：仅校验当前码，不做「近期已验证」会话缓存（可后续扩展）。
func RequireRecentTOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, ok := c.Get("userID")
		if !ok {
			c.Next()
			return
		}
		userID, ok := userIDVal.(uint64)
		if !ok || userID == 0 {
			c.Next()
			return
		}

		user, err := models.GetUserByID(userID)
		if err != nil || user == nil || !user.TotpEnabled {
			// 未启用 2FA：跳过
			c.Next()
			return
		}

		code := c.GetHeader("X-Totp-Code")
		if code == "" {
			utils.Fail(c, 403, "TOTP code required")
			c.Abort()
			return
		}
		if user.TotpSecret == nil || *user.TotpSecret == "" {
			utils.Fail(c, 403, "TOTP not configured")
			c.Abort()
			return
		}
		if !utils.ValidateTOTP(*user.TotpSecret, code) {
			utils.Fail(c, 403, "Invalid TOTP code")
			c.Abort()
			return
		}
		c.Next()
	}
}
