package routes

import (
	"fst/backend/app/controllers"
	"fst/backend/app/controllers/admin"
	_ "fst/backend/docs"
	"fst/backend/internal/config"
	"fst/backend/internal/middleware"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine) {
	// 控制器初始化
	authCtrl := &controllers.AuthController{}
	systemCtrl := &controllers.SystemController{}

	// Admin 控制器
	adminUserCtrl := admin.NewUserController()
	adminLogCtrl := admin.NewLogController()

	// Swagger 文档
	if config.GlobalConfig.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

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
			public := v1.Group("/public")
			{
				// 认证相关
				public.POST("/login", authCtrl.Login)
				public.POST("/register", authCtrl.Register)
				public.POST("/send-register-code", authCtrl.SendRegisterCode)
				public.POST("/forgot-password", authCtrl.SendResetEmail)
				public.POST("/reset-password", authCtrl.ResetPasswordConfirm)
			public.POST("/refresh-token", authCtrl.UpdateToken)
			}

			// ----------------------------------------
			// 用户接口 (需要登录)
			// ----------------------------------------
			user := v1.Group("/user")
			// 注册验证码接口支持 /api/v1/user/send-register-code (前端常用)
			user.POST("/send-register-code", authCtrl.SendRegisterCode)

			user.Use(middleware.AuthMiddleware())
			{
				// 个人信息
				user.GET("/profile", func(c *gin.Context) {
					userID, exists := c.Get("userID")
					if !exists {
						utils.Fail(c, 401, "用户未登录")
						return
					}
					utils.Success(c, gin.H{"userID": userID})
				})
				user.PUT("/profile", func(c *gin.Context) {
					// TODO: 更新个人资料
					utils.Success(c, nil)
				})
				user.PUT("/password", func(c *gin.Context) {
					// TODO: 修改密码
					utils.Success(c, nil)
				})

// 用户路由
				user.GET("/routes", authCtrl.GetUserRoutes)
			}

			// ----------------------------------------
			// 系统状态接口 (需要登录)
			// ----------------------------------------
			system := v1.Group("/system")
			system.Use(middleware.AuthMiddleware())
			{
				system.GET("/cleanup-status", systemCtrl.GetCleanupStatus)
			}

			// ----------------------------------------
			// 管理后台接口 (需要管理员权限)
			// 路径可通过 .env 配置: ADMIN_PATH
			// ----------------------------------------
			adminPath := config.GlobalConfig.AdminPath
			if adminPath == "" {
				adminPath = "/admin"
			}
			adminGroup := v1.Group(adminPath)
			adminGroup.Use(middleware.AuthMiddleware())
			adminGroup.Use(middleware.AdminOnly())
			{
				// 仪表盘
				adminGroup.GET("/dashboard", func(c *gin.Context) {
					utils.Success(c, gin.H{"message": "欢迎访问管理后台"})
				})

				// ----- 用户管理 -----
				users := adminGroup.Group("/users")
				users.Use(middleware.SimpleLogMiddleware("用户管理"))
				{
					users.GET("", adminUserCtrl.List)
					users.GET("/:id", adminUserCtrl.Detail)
					users.POST("", adminUserCtrl.Create)
					users.PUT("/:id", adminUserCtrl.Update)
				users.DELETE("/:id", adminUserCtrl.Delete)
					users.PUT("/:id/status", adminUserCtrl.UpdateStatus)
					users.PUT("/:id/password", adminUserCtrl.ResetPassword)
				}

				// ----- 操作日志 -----
				logs := adminGroup.Group("/logs")
				{
					logs.GET("", adminLogCtrl.List)
					logs.GET("/stats", adminLogCtrl.Stats)
					logs.POST("/clean", adminLogCtrl.Clean)
				}
			}
		}

		// ========================================
		// 兼容旧接口 (逐步废弃)
		// 统一为 /api/v1 前缀，保留旧路径行为
		// ========================================
		v1.POST("/login", authCtrl.Login)
		v1.POST("/register", authCtrl.Register)
		v1.POST("/updateToken", authCtrl.UpdateToken)
		v1.GET("/getUserRoutes", authCtrl.GetUserRoutes)
		v1.GET("/userPage", systemCtrl.GetUserPage)
		// 提示使用正确的请求方法
		v1.GET("/login", func(c *gin.Context) {
			utils.Fail(c, 405, "请使用 POST 方法登录")
		})
		v1.GET("/register", func(c *gin.Context) {
			utils.Fail(c, 405, "请使用 POST 方法注册")
		})
	}
}
