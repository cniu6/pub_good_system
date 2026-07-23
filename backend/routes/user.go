package routes

import (
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// registerUserRoutes 用户侧接口（需登录；管理员 token 也可访问）
func registerUserRoutes(v1 *gin.RouterGroup) {
	userGroup := v1.Group("/user")
	userGroup.Use(middleware.AuthMiddlewareForGuard("user", "admin"))
	{
		userProfileCtrl.RegisterRoutes(userGroup)
		userPaymentCtrl.RegisterRoutes(userGroup)  //用户支付
		userRealnameCtrl.RegisterRoutes(userGroup) //实名认证
		userWithdrawCtrl.RegisterRoutes(userGroup) //用户提现

		// 站内公告（按用户已读）
		anns := userGroup.Group("/announcements")
		{
			anns.GET("", userAnnouncementCtrl.List)
			anns.GET("/unread-count", userAnnouncementCtrl.UnreadCount)
			anns.POST("/read-all", userAnnouncementCtrl.MarkAllRead)
			anns.GET("/:id", userAnnouncementCtrl.Detail)
			anns.POST("/:id/read", userAnnouncementCtrl.MarkRead)
		}
	}

	// 系统状态（登录后可查清理任务等）
	system := v1.Group("/system")
	system.Use(middleware.AuthMiddlewareForGuard("user", "admin"))
	{
		system.GET("/cleanup-status", systemCtrl.GetCleanupStatus)
		system.POST("/ws-ticket", systemCtrl.CreatePresenceTicket)
	}
}
