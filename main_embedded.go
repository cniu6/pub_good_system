//go:build embedded

package main

import (
	"embed"
	"fst/backend/app/models"
	"fst/backend/app/plugins"
	"fst/backend/app/plugins/demo"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"fst/backend/pkg/middleware"
	"fst/backend/routes"
	"fst/backend/utils"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// BuildMode 由构建脚本在编译时注入: "embedded" 或 "external" 或 "none"
// go run main_embedded.go 时默认为 "embedded"，直接从二进制内嵌 FS 提供前端
var BuildMode = "embedded"

//go:embed dist/*
var frontendFS embed.FS

// @title FST Platform API
// @version 1.0
// @description FST Platform 后端 API 接口文档
// @host localhost:8080
// @BasePath /api

func main() {
	config.InitConfig()
	if config.IsProductionMode() {
		gin.SetMode(gin.ReleaseMode)
	}
	db.InitDB()

	// 初始化邮件模板
	models.InitEmailTemplates()

	// 初始化验证码表（如果不存在）
	models.InitVerificationCodeTable()

	// 初始化系统配置表
	models.InitSystemSettingsTable()

	// 初始化用户设置表
	models.InitUserSettingsTable()

	// 初始化用户会话表
	models.InitUserSessionsTable()

	// 初始化余额/积分变动日志表
	models.InitUserMoneyLogsTable()
	models.InitUserScoreLogsTable()
	models.InitOperationLogsTable()

	// 初始化支付订单表
	models.InitPaymentOrdersTable()

	// 初始化提现申请表
	models.InitWithdrawRequestsTable()

	// 初始化接口幂等键表
	models.InitIdempotencyKeysTable()

	// 初始化支付通道表
	models.InitPayGatewaysTable()

	// 初始化配置服务（缓存）
	services.InitSettingsService()

	// 启动定时清理任务：间隔可通过 CLEANUP_INTERVAL_MINUTES 配置，默认10分钟
	// 清理状态仅在内存中记录，不输出周期性日志，可通过接口查询
	services.StartCleanupTask()
	models.CleanupExpiredIdempotencyKeys()

	// 初始化短信服务
	services.InitSMSService()

	router := gin.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(gin.Logger(), gin.Recovery())
	router.SetTrustedProxies(nil) // 修复 "trusted all proxies" 警告
	router.Use(middleware.CorsMiddleware())
	routes.SetupRoutes(router)

	// 插件初始化
	pluginMgr := plugins.NewManager()
	plugins.AutoRegisterAll(pluginMgr)
	pluginMgr.Register(demo.NewPlugin())
	if err := pluginMgr.LoadAll(); err != nil {
		log.Printf("[Plugin] 插件加载失败: %v", err)
	}
	apiGroup := router.Group("/api/v1")
	pluginMgr.RegisterAllRoutes(apiGroup)

	// 前端资源处理：main_embedded.go 始终托管嵌入的前端
	var rawFS fs.FS           // 用于直接读取文件（避免 http.FileServer 重定向问题）
	var publicHTTPFS http.FileSystem // 用于静态资源服务

	if BuildMode == "external" {
		log.Println("[Mode] External: Serving from ./dist folder")
		rawFS = os.DirFS("dist")
		publicHTTPFS = http.Dir("dist")
	} else {
		// 默认 embedded 模式：从 go:embed 内嵌 FS 提供前端
		log.Println("[Mode] Embedded: Serving from binary internal FS")
		distFS, err := fs.Sub(frontendFS, "dist")
		if err != nil {
			log.Fatalf("Failed to load embedded frontend: %v", err)
		}
		rawFS = distFS
		publicHTTPFS = http.FS(distFS)
	}

	// 不再使用 http.FileServer（会产生 301 重定向），改为直接读取并返回
	_ = publicHTTPFS // 保留声明以兼容 external 模式扩展
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			utils.Fail(c, 404, "API not found")
			return
		}
		// 尝试直接读取静态文件（完全绕过 http.FileServer 的重定向逻辑）
		relPath := strings.TrimPrefix(path, "/")
		if relPath != "" && relPath != "index.html" {
			if data, err := fs.ReadFile(rawFS, relPath); err == nil {
				contentType := mime.TypeByExtension(filepath.Ext(relPath))
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				c.Data(http.StatusOK, contentType, data)
				return
			}
		}
		// 根路径、不存在的路径一律返回 index.html 供 SPA 路由处理
		indexData, err := fs.ReadFile(rawFS, "index.html")
		if err != nil {
			log.Printf("[Embedded] index.html 读取失败: %v", err)
			c.Status(http.StatusNotFound)
			return
		}
		log.Printf("[Embedded] 返回 index.html for path: %s", path)
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	port := config.GlobalConfig.Port
	server := utils.NewHTTPServer(":"+port, router)
	log.Printf("Server starting on port %s [%s Mode]...", port, BuildMode)
	if err := utils.ServeHTTPServer(server, pluginMgr.ShutdownAll); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
