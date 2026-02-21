package main

import (
	"fst/backend/app/models"
	"fst/backend/app/plugins"
	"fst/backend/app/plugins/demo"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/internal/db"
	"fst/backend/internal/middleware"
	"fst/backend/routes"
	"log"

	"github.com/gin-gonic/gin"
)

// @title FST Platform API
// @version 1.0
// @description FST Platform 后端 API 接口文档
// @host localhost:8080
// @BasePath /api

func main() {
	config.InitConfig()
	db.InitDB()

	// 初始化邮件模板
	models.InitEmailTemplates()

	// 初始化验证码表（如果不存在）
	models.InitVerificationCodeTable()

	// 启动定时清理任务：间隔可通过 CLEANUP_INTERVAL_MINUTES 配置，默认10分钟
	// 清理状态仅在内存中记录，不输出周期性日志，可通过接口查询
	services.StartCleanupTask()

	router := gin.Default()
	router.SetTrustedProxies(nil) // 修复 "trusted all proxies" 警告
	router.Use(middleware.CorsMiddleware())
	routes.SetupRoutes(router)

	// 插件初始化
	pluginMgr := plugins.NewPluginManager()
	pluginMgr.Register(demo.NewPlugin())
	apiGroup := router.Group("/api/v1")
	for _, p := range pluginMgr.GetPlugins() {
		if err := p.Init(); err != nil {
			log.Printf("Plugin %s init failed: %v", p.Name(), err)
			continue
		}
		p.RegisterRoutes(apiGroup)
	}

	port := config.GlobalConfig.Port
	log.Printf("Server starting on port %s [dev backend only]...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
