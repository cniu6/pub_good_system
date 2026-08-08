package routes

import (
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// registerSystemRoutes 注册系统级路由（登录后可访问，与用户/管理员鉴权分离）。
func registerSystemRoutes(v1 *gin.RouterGroup) {
	system := v1.Group("/system")
	system.Use(middleware.AuthMiddlewareForGuard("user", "admin"))
	{
		system.GET("/cleanup-status", systemCtrl.GetCleanupStatus) // 获取清理状态
		system.POST("/ws-ticket", systemCtrl.CreatePresenceTicket) // 创建WebSocket票据（心跳上报在线状态）
	}
}
