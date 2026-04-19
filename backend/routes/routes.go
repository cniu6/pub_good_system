package routes

import (
	"fst/backend/app/controllers"
	"fst/backend/app/controllers/admin"
	"fst/backend/app/controllers/public"
	"fst/backend/app/controllers/user"
	_ "fst/backend/docs"
	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"

	// "fst/backend/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ========================================
// 初始化控制器（延迟初始化避免循环依赖）
// ========================================

var (
	publicAuthCtrl            *public.AuthController
	publicSettingsCtrl        *public.SettingsController
	publicPaymentCallbackCtrl *public.PaymentCallbackController
	userProfileCtrl           *user.ProfileController
	userPaymentCtrl           *user.PaymentController
	userRealnameCtrl          *user.RealnameController
	userWithdrawCtrl          *user.WithdrawController
	systemCtrl                *controllers.SystemController
	adminUserCtrl             *admin.UserController
	adminLogCtrl              *admin.LogController
	adminAPILogCtrl           *admin.APILogController
	adminEmailTplCtrl         *admin.EmailTemplateController
	adminEmailLogCtrl         *admin.EmailLogController
	adminSettingsCtrl         *admin.SettingsController
	adminDebugCtrl            *admin.DebugController
	adminMoneyScoreCtrl       *admin.UserMoneyScoreController
	adminPaymentCtrl          *admin.PaymentController
	adminRealnameCtrl         *admin.RealnameController
	adminWithdrawCtrl         *admin.WithdrawController
	adminSMSLogCtrl          *admin.SMSLogController
)

// initControllers 初始化所有控制器
func initControllers() {
	publicAuthCtrl = public.NewAuthController()
	publicSettingsCtrl = public.NewSettingsController()
	publicPaymentCallbackCtrl = public.NewPaymentCallbackController()
	userProfileCtrl = user.NewProfileController()
	userPaymentCtrl = user.NewPaymentController()
	userRealnameCtrl = user.NewRealnameController()
	userWithdrawCtrl = user.NewWithdrawController()
	systemCtrl = &controllers.SystemController{}
	adminUserCtrl = admin.NewUserController()
	adminLogCtrl = admin.NewLogController()
	adminAPILogCtrl = admin.NewAPILogController()
	adminEmailTplCtrl = admin.NewEmailTemplateController()
	adminEmailLogCtrl = admin.NewEmailLogController()
	adminSettingsCtrl = admin.NewSettingsController()
	adminDebugCtrl = admin.NewDebugController()
	adminMoneyScoreCtrl = admin.NewUserMoneyScoreController()
	adminPaymentCtrl = admin.NewPaymentController()
	adminRealnameCtrl = admin.NewRealnameController()
	adminWithdrawCtrl = admin.NewWithdrawController()
	adminSMSLogCtrl = admin.NewSMSLogController()
}

func SetupRoutes(router *gin.Engine) {
	// ========================================
	// 初始化控制器
	// ========================================
	initControllers()

	// ========================================
	// Swagger 文档
	// ========================================
	if config.GlobalConfig.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// ========================================
	// API 路由
	// ========================================
	api := router.Group("/api")
	{
		// ========================================
		// V1 版本 API
		// ========================================
		v1 := api.Group("/v1")
		{
			// ----------------------------------------
			// 公共接口 (无需登录)
			// ----------------------------------------
			publicGroup := v1.Group("/public")
			{
				publicAuthCtrl.RegisterRoutes(publicGroup)
				publicSettingsCtrl.RegisterRoutes(publicGroup)
				publicPaymentCallbackCtrl.RegisterRoutes(publicGroup)
			}

			// ----------------------------------------
			// 用户接口 (需要登录，管理员token也可访问)
			// ----------------------------------------
			userGroup := v1.Group("/user")
			userGroup.Use(middleware.AuthMiddlewareForGuard("user", "admin"))
			{
				userProfileCtrl.RegisterRoutes(userGroup)
				userPaymentCtrl.RegisterRoutes(userGroup)
				userRealnameCtrl.RegisterRoutes(userGroup)
				userWithdrawCtrl.RegisterRoutes(userGroup)
			}

			// ----------------------------------------
			// 系统状态接口 (需要登录，管理员token也可访问)
			// ----------------------------------------
			system := v1.Group("/system")
			system.Use(middleware.AuthMiddlewareForGuard("user", "admin"))
			{
				system.GET("/cleanup-status", systemCtrl.GetCleanupStatus)
			}

			// ----------------------------------------
			// 管理后台接口 (需要管理员权限)
			// ----------------------------------------
			adminGroup := v1.Group("/admin")
			adminGroup.Use(middleware.AuthMiddlewareForGuard("admin"))
			adminGroup.Use(middleware.AdminOnly())
			adminGroup.Use(middleware.DynamicAdminRateLimitMiddleware())
			{
				// 仪表盘
				adminGroup.GET("/dashboard", admin.GetDashboard)

				// ----- 用户管理 -----
				users := adminGroup.Group("/users")
				users.Use(middleware.SimpleLogMiddleware("用户管理"))
				{
					users.GET("", adminUserCtrl.List)
					users.GET("/:id", adminUserCtrl.Detail)
					users.POST("", adminUserCtrl.Create)
					users.POST("/batch-simple", adminUserCtrl.BatchGetSimpleInfo) // 批量获取用户简要信息
					users.PUT("/:id", adminUserCtrl.Update)
					users.DELETE("/:id", adminUserCtrl.Delete)
					users.PUT("/:id/status", adminUserCtrl.UpdateStatus)
					users.PUT("/:id/password", adminUserCtrl.ResetPassword)
					users.GET("/lookup", adminUserCtrl.LookupUser)
					users.POST("/:id/login-as", adminUserCtrl.LoginToUser)
					users.POST("/:id/reset-apikey", adminUserCtrl.ResetApiKey)
				}

				// ----- 操作日志 -----
				logs := adminGroup.Group("/logs")
				{
					logs.GET("", adminLogCtrl.List)
					logs.GET("/:id", adminLogCtrl.Detail)
					logs.POST("/clean", adminLogCtrl.Clean)
				}

				// ----- API接口日志 -----
				apiLogs := adminGroup.Group("/api-logs")
				{
					apiLogs.GET("", adminAPILogCtrl.List)
					apiLogs.GET("/stats", adminAPILogCtrl.Stats)
					apiLogs.GET("/:id", adminAPILogCtrl.Detail)
					apiLogs.POST("/clean", adminAPILogCtrl.Clean)
				}

				// ----- 邮件发件测试 -----
				adminGroup.POST("/email-send-test", adminEmailTplCtrl.SendTest)

				// ----- 邮件模板 -----
				emailTemplates := adminGroup.Group("/email-templates")
				{
					emailTemplates.GET("", adminEmailTplCtrl.List)
					emailTemplates.GET("/:id", adminEmailTplCtrl.Detail)
					emailTemplates.PUT("/:id", adminEmailTplCtrl.Update)
					emailTemplates.POST("/:id/preview", adminEmailTplCtrl.Preview)
					emailTemplates.POST("/:id/reset", adminEmailTplCtrl.Reset)
				}

			// ----- 邮件发送记录 -----
			emailLogs := adminGroup.Group("/email-logs")
			{
				emailLogs.GET("", adminEmailLogCtrl.List)
				emailLogs.GET("/stats", adminEmailLogCtrl.Stats)
				emailLogs.GET("/template-names", adminEmailLogCtrl.TemplateNames)
				emailLogs.GET("/:id", adminEmailLogCtrl.Detail)
				emailLogs.POST("/clean", adminEmailLogCtrl.Clean)
			}

			// ----- 短信发送记录 -----
			smsLogs := adminGroup.Group("/sms-logs")
			{
				smsLogs.GET("", adminSMSLogCtrl.List)
				smsLogs.GET("/stats", adminSMSLogCtrl.Stats)
				smsLogs.GET("/template-names", adminSMSLogCtrl.TemplateNames)
				smsLogs.GET("/:id", adminSMSLogCtrl.Detail)
				smsLogs.POST("/clean", adminSMSLogCtrl.Clean)
			}

				// ----- 余额/积分管理 -----
				adminMoneyScoreCtrl.RegisterRoutes(adminGroup)

				// ----- 系统配置 -----
				adminSettingsCtrl.RegisterRoutes(adminGroup)

			// ----- 支付订单管理 -----
			adminPaymentCtrl.RegisterPaymentRoutes(adminGroup)

			// ----- 实名认证管理 -----
			adminRealnameCtrl.RegisterRoutes(adminGroup)

			// ----- 提现管理 -----
			adminWithdrawCtrl.RegisterRoutes(adminGroup)

			// ----- 调试工具 -----
			adminDebugCtrl.RegisterRoutes(adminGroup)
		}
		}

		// ========================================
		// 兼容旧接口 (逐步废弃)
		// ========================================
		{
			// 认证相关 - 指向新的公共接口
			v1.POST("/login", func(c *gin.Context) {
				c.Redirect(307, "/api/v1/public/login")
			})
			v1.POST("/register", func(c *gin.Context) {
				c.Redirect(307, "/api/v1/public/register")
			})
			v1.POST("/updateToken", func(c *gin.Context) {
				c.Redirect(307, "/api/v1/public/refresh-token")
			})
			v1.GET("/userPage", systemCtrl.GetUserPage)

			// 提示使用正确的请求方法，多余请求直接404，不要回退
			// v1.GET("/login", func(c *gin.Context) {
			// 	utils.Fail(c, 405, "请使用 POST 方法登录")
			// })
			// v1.GET("/register", func(c *gin.Context) {
			// 	utils.Fail(c, 405, "请使用 POST 方法注册")
			// })
		}
	}
}

