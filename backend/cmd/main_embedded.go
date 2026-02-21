package main

import (
	"embed"
	"fst/backend/app/models"
	"fst/backend/app/plugins"
	"fst/backend/app/plugins/demo"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/internal/db"
	"fst/backend/internal/middleware"
	"fst/backend/routes"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BuildMode 由构建脚本在编译时注入: "embedded" 或 "external" 或 "none"
// 默认值为 "none"，表示开发阶段不嵌入前端，仅提供后端 API
var BuildMode = "none"

//go:embed dist/*
var frontendFS embed.FS

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

	// 前端资源处理
	// 仅当 AppMode == "integrated" 且 BuildMode != "none" 时，才提供前端托管能力
	if config.GlobalConfig.AppMode == "integrated" && BuildMode != "none" {
		var publicFS http.FileSystem

		if BuildMode == "external" {
			log.Println("[Mode] External: Serving from ./dist folder")
			publicFS = http.Dir("dist")
		} else if BuildMode == "embedded" {
			log.Println("[Mode] Embedded: Serving from binary internal FS")
			distFS, err := fs.Sub(frontendFS, "dist")
			if err != nil {
				log.Fatalf("Failed to load embedded frontend: %v", err)
			}
			publicFS = http.FS(distFS)
		} else {
			log.Printf("[Mode] Unknown BuildMode=%s, frontend will not be served", BuildMode)
		}

		if publicFS != nil {
			staticFileServer := http.FileServer(publicFS)
			router.NoRoute(func(c *gin.Context) {
				path := c.Request.URL.Path
				if strings.HasPrefix(path, "/api") {
					c.JSON(404, gin.H{"error": "API not found"})
					return
				}
				// Serve static or index.html
				f, err := publicFS.Open(strings.TrimPrefix(path, "/"))
				if err == nil {
					f.Close()
					staticFileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
				c.FileFromFS("index.html", publicFS)
			})
		}
	} else {
		log.Printf("[Mode] Backend only: AppMode=%s, BuildMode=%s, frontend not served", config.GlobalConfig.AppMode, BuildMode)
	}

	port := config.GlobalConfig.Port
	log.Printf("Server starting on port %s [%s Mode]...", port, BuildMode)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
