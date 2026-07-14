package middleware

import (
	"errors"
	"fst/backend/internal/db"
	"fst/backend/models"
	"fst/backend/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware JWT 认证中间件（草稿 GORM 栈）
// 对齐现网：Bearer Token + ParseToken(userID, role) + 校验用户存在与状态
// 不使用 PasswordMD5；不强制 sqlx 的 user_sessions（草稿库可独立跑）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Fail(c, 401, "未授权访问")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Fail(c, 401, "未授权访问")
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			utils.Fail(c, 401, "未授权访问")
			c.Abort()
			return
		}

		// 使用现网 JWT 解析（2 参数 claims：user_id + role）
		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.Fail(c, 401, "未授权访问")
			c.Abort()
			return
		}

		// 从草稿 GORM 库加载用户
		dbInstance := db.GetDB()
		var user models.User
		err = dbInstance.Table("users").Where("id", "=", claims.UserID).First(&user)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Fail(c, 401, "用户不存在")
			} else {
				utils.Fail(c, 500, "数据库查询失败")
			}
			c.Abort()
			return
		}

		// 角色不一致则要求重新登录
		if string(user.Role) != claims.Role {
			utils.Fail(c, 401, "用户角色已变更，请重新登录")
			c.Abort()
			return
		}

		if user.Status == models.UserStatusInactive {
			utils.Fail(c, 403, "用户已被禁用")
			c.Abort()
			return
		}

		// 写入上下文：同时兼容现网键名 userID 与草稿 user_id
		c.Set(utils.ContextKeyUser, user)
		c.Set(utils.ContextKeyUserID, user.ID)
		c.Set("userID", uint64(user.ID))
		c.Set(utils.ContextKeyUsername, user.Username)
		c.Set(utils.ContextKeyRole, string(user.Role))
		c.Set(utils.ContextKeyStatus, user.Status)
		c.Set("role", string(user.Role))
		c.Set("username", user.Username)
		c.Set("authGuard", utils.UserAuthGuard)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件（需在 AuthMiddleware 之后）
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !utils.IsAdmin(c) {
			utils.Fail(c, 403, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORSMiddleware 跨域中间件（草稿默认 *，正式环境请走现网配置）
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, apikey")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件（默认空格式，避免刷屏）
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return ""
	})
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}
