package main

import (
	"fst/backend/app/models"
	"fst/backend/app/plugins"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/internal/db"
	"fst/backend/internal/middleware"
	"fst/backend/routes"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	// ========================================
	// 插件自动导入区域
	// 运行 gen-swagger.bat 会自动扫描并更新此区域
	// ========================================
	// @plugins-start
	_ "fst/backend/app/plugins/demo"
	// @plugins-end
)

// @title FST Platform API
// @version 1.0
// @description FST Platform 后端 API 接口文档
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT认证令牌，格式: Bearer {token}

func main() {
	// 1. 初始化配置
	config.InitConfig()

	// 2. 初始化数据库
	db.InitDB()

	// 3. 初始化邮件模板
	models.InitEmailTemplates()

	// 4. 初始化验证码表
	models.InitVerificationCodeTable()

	// 5. 初始化系统配置表
	models.InitSystemSettingsTable()

	// 5.1 初始化用户设置表
	models.InitUserSettingsTable()

	// 5.2 初始化用户会话表
	models.InitUserSessionsTable()

	// 6. 初始化配置服务（缓存）
	services.InitSettingsService()

	// 7. 启动定时清理任务
	services.StartCleanupTask()

	// 8. 创建路由
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(middleware.CorsMiddleware())

	// 9. 添加请求日志中间件
	router.Use(middleware.LoggerMiddleware())

	// 10. 注册路由
	routes.SetupRoutes(router)

	// 11. 初始化插件系统（使用新的管理器）
	pluginMgr := plugins.NewManager()

	// 自动注册所有导入的插件
	// 注意：插件通过 init() 函数自动注册到全局注册表
	plugins.AutoRegisterAll(pluginMgr)

	// 加载所有插件
	if err := pluginMgr.LoadAll(); err != nil {
		log.Printf("[Plugin] 插件加载失败: %v", err)
	}

	// 注册插件路由
	apiGroup := router.Group("/api/v1")
	pluginMgr.RegisterAllRoutes(apiGroup)

	// 12. 优雅关闭处理
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("[Server] 正在关闭...")

		// 关闭插件
		if err := pluginMgr.ShutdownAll(); err != nil {
			log.Printf("[Plugin] 关闭失败: %v", err)
		}

		log.Println("[Server] 已关闭")
		os.Exit(0)
	}()

	// 13. 启动服务
	port := config.GlobalConfig.Port
	log.Printf("[Server] 服务启动，端口: %s", port)
	log.Printf("[Server] Swagger 文档: http://localhost:%s/swagger/index.html", port)
	log.Printf("[Server] 已加载插件数量: %d", pluginMgr.Count())

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[Server] 启动失败: %v", err)
	}
}
