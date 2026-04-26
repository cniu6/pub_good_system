//go:build embedded

package main

import (
	"embed"
	_ "fst/backend/app/plugins/demo"
	_ "fst/backend/app/plugins/sms"
	"fst/backend/app/models"
	"fst/backend/app/plugins"
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

func requestAcceptsGzip(r *http.Request) bool {
	if r == nil {
		return false
	}
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(strings.ToLower(acceptEncoding), "gzip")
}

func isStaticAssetRequest(relPath string) bool {
	return strings.TrimSpace(filepath.Ext(relPath)) != ""
}

func frontendContentType(relPath string) string {
	cleanPath := strings.TrimSuffix(relPath, ".gz")
	contentType := mime.TypeByExtension(filepath.Ext(cleanPath))
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

func serveFrontendFile(c *gin.Context, rawFS fs.FS, relPath string) bool {
	if c == nil || rawFS == nil {
		return false
	}

	if requestAcceptsGzip(c.Request) {
		gzipPath := relPath + ".gz"
		if data, err := fs.ReadFile(rawFS, gzipPath); err == nil {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
			c.Data(http.StatusOK, frontendContentType(gzipPath), data)
			return true
		}
	}

	data, err := fs.ReadFile(rawFS, relPath)
	if err != nil {
		return false
	}

	c.Data(http.StatusOK, frontendContentType(relPath), data)
	return true
}

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
	if err := pluginMgr.LoadAll(); err != nil {
		log.Printf("[Plugin] 插件加载失败: %v", err)
	}
	apiGroup := router.Group("/api/v1")
	pluginMgr.RegisterAllRoutes(apiGroup)

	// 前端资源处理：main_embedded.go 始终托管嵌入的前端
	var rawFS fs.FS

	if BuildMode == "external" {
		log.Println("[Mode] External: Serving from ./dist folder")
		rawFS = os.DirFS("dist")
	} else {
		// 默认 embedded 模式：从 go:embed 内嵌 FS 提供前端
		log.Println("[Mode] Embedded: Serving from binary internal FS")
		distFS, err := fs.Sub(frontendFS, "dist")
		if err != nil {
			log.Fatalf("Failed to load embedded frontend: %v", err)
		}
		rawFS = distFS
	}

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			utils.Fail(c, 404, "API not found")
			return
		}

		relPath := strings.TrimPrefix(path, "/")
		if relPath != "" {
			if serveFrontendFile(c, rawFS, relPath) {
				return
			}

			if isStaticAssetRequest(relPath) {
				c.Status(http.StatusNotFound)
				return
			}
		}

		if !serveFrontendFile(c, rawFS, "index.html") {
			log.Printf("[Embedded] index.html 读取失败")
			c.Status(http.StatusNotFound)
			return
		}

		log.Printf("[Embedded] 返回 index.html for path: %s", path)
	})

	port := config.GlobalConfig.Port
	server := utils.NewHTTPServer(":"+port, router)
	log.Printf("Server starting on port %s [%s Mode]...", port, BuildMode)
	if err := utils.ServeHTTPServer(server, pluginMgr.ShutdownAll); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
