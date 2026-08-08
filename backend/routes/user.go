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
		userPaymentCtrl.RegisterRoutes(userGroup)      // 用户支付
		userRealnameCtrl.RegisterRoutes(userGroup)     // 实名认证
		userWithdrawCtrl.RegisterRoutes(userGroup)     // 用户提现
		userAnnouncementCtrl.RegisterRoutes(userGroup) // 站内公告
	}
}
