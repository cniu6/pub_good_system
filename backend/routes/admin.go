package routes

import (
	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// registerAdminRoutes 管理后台 REST（前缀可配置，需管理员权限）
// - API 前缀：.env 的 ADMIN_API_PATH（默认 /admin）→ 实际 /api/v1{ADMIN_API_PATH}
// - 前端页面入口：ADMIN_PATH / VITE_ADMIN_BASE_PATH，与 API 分离
func registerAdminRoutes(v1 *gin.RouterGroup) {
	adminAPIPath := "/admin"
	if cfg := config.GlobalConfig; cfg != nil {
		adminAPIPath = config.NormalizeAdminAPIPath(cfg.AdminAPIPath)
	}

	adminGroup := v1.Group(adminAPIPath)
	adminGroup.Use(middleware.AuthMiddlewareForGuard("admin"))
	adminGroup.Use(middleware.AdminOnly())
	adminGroup.Use(middleware.DynamicAdminRateLimitMiddleware())
	{
		adminGroup.GET("/dashboard", adminDashboardCtrl.GetDashboard)

		users := adminGroup.Group("/users")
		users.Use(middleware.SimpleLogMiddleware("用户管理"))
		{
			users.GET("", adminUserCtrl.List)
			users.GET("/:id", adminUserCtrl.Detail)
			users.POST("", adminUserCtrl.Create)
			users.POST("/batch-simple", adminUserCtrl.BatchGetSimpleInfo)
			users.PUT("/:id", adminUserCtrl.Update)
			users.DELETE("/:id", adminUserCtrl.Delete)
			users.PUT("/:id/status", adminUserCtrl.UpdateStatus)
			users.PUT("/:id/password", adminUserCtrl.ResetPassword)
			users.GET("/lookup", adminUserCtrl.LookupUser)
			users.POST("/:id/login-as", adminUserCtrl.LoginToUser)
			users.POST("/:id/reset-apikey", adminUserCtrl.ResetApiKey)
		}

		logs := adminGroup.Group("/logs")
		{
			logs.GET("", adminLogCtrl.List)
			logs.GET("/:id", adminLogCtrl.Detail)
			logs.POST("/clean", adminLogCtrl.Clean)
		}

		apiLogs := adminGroup.Group("/api-logs")
		{
			apiLogs.GET("", adminAPILogCtrl.List)
			apiLogs.GET("/stats", adminAPILogCtrl.Stats)
			apiLogs.GET("/:id", adminAPILogCtrl.Detail)
			apiLogs.POST("/clean", adminAPILogCtrl.Clean)
		}

		adminGroup.POST("/email-send-test", adminEmailTplCtrl.SendTest)

		emailTemplates := adminGroup.Group("/email-templates")
		{
			emailTemplates.GET("", adminEmailTplCtrl.List)
			emailTemplates.GET("/:id", adminEmailTplCtrl.Detail)
			emailTemplates.PUT("/:id", adminEmailTplCtrl.Update)
			emailTemplates.POST("/:id/preview", adminEmailTplCtrl.Preview)
			emailTemplates.POST("/:id/reset", adminEmailTplCtrl.Reset)
		}

		emailLogs := adminGroup.Group("/email-logs")
		{
			emailLogs.GET("", adminEmailLogCtrl.List)
			emailLogs.GET("/stats", adminEmailLogCtrl.Stats)
			emailLogs.GET("/template-names", adminEmailLogCtrl.TemplateNames)
			emailLogs.GET("/:id", adminEmailLogCtrl.Detail)
			emailLogs.POST("/clean", adminEmailLogCtrl.Clean)
		}

		smsLogs := adminGroup.Group("/sms-logs")
		{
			smsLogs.GET("", adminSMSLogCtrl.List)
			smsLogs.GET("/stats", adminSMSLogCtrl.Stats)
			smsLogs.GET("/template-names", adminSMSLogCtrl.TemplateNames)
			smsLogs.GET("/:id", adminSMSLogCtrl.Detail)
			smsLogs.POST("/clean", adminSMSLogCtrl.Clean)
		}

		adminMoneyScoreCtrl.RegisterRoutes(adminGroup)     //用户积分
		adminSettingsCtrl.RegisterRoutes(adminGroup)       //系统设置
		adminPaymentCtrl.RegisterPaymentRoutes(adminGroup) //支付
		adminRealnameCtrl.RegisterRoutes(adminGroup)       //实名认证
		adminWithdrawCtrl.RegisterRoutes(adminGroup)       //提现
		adminDebugCtrl.RegisterRoutes(adminGroup)          //调试
		adminAutoJobCtrl.RegisterRoutes(adminGroup)        //自动任务管理器
	}
}
