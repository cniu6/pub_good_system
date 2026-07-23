// Package appinit 正式启动编排（系统内部骨架）。
//
// 调用链：
//
//	根 main.go / main_embedded.go
//	  → backend/cmd/server.Start（薄壳：前端托管 + Listen）
//	  → appinit.Bootstrap / SetupHTTP（本包）
//	  → internal/migrate.RunAutoMigrate（GORM AutoMigrate + 补丁/种子）
//
// 业务 API 仍在 backend/app 与 backend/routes，本包只负责「加载顺序」。
// pkg = 可复用零件（config/db连接/middleware）；internal = 开机组装。
package appinit

import (
	"log"

	"fst/backend/app/plugins"
	"fst/backend/app/services"
	"fst/backend/internal/migrate"
	"fst/backend/pkg/apilog"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"fst/backend/pkg/middleware"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

// Bootstrap 完成配置、数据库、业务表自迁移、服务与定时任务。
// 必须在创建 HTTP 引擎之前调用（有依赖顺序，勿打乱）。
func Bootstrap() {
	log.Println("[AppInit] ============系统加载中==============")

	// ---------- 1. 配置与 Gin 运行模式 ----------
	config.InitConfig()
	log.Println("[AppInit] 配置初始化完成")
	// 生产误配强警告（CORS=* / Swagger / 弱 JWT 等），不中断启动
	config.WarnProductionMisconfig()
	if config.IsProductionMode() {
		gin.SetMode(gin.ReleaseMode)
		log.Println("[AppInit] 生产模式")
	} else {
		gin.SetMode(gin.DebugMode)
		log.Println("[AppInit] 开发模式")
	}
	log.Println("[AppInit] GIN模式初始化完成")

	// ---------- 2. 数据库连接（只连库，不建表） ----------
	db.InitDB()
	log.Println("[AppInit] 数据库连接完成")

	// ---------- 3. 数据库自迁移（GORM AutoMigrate + 补丁/种子） ----------
	migrate.RunAutoMigrate()

	// ---------- 4. 业务服务（非定时任务） ----------
	services.InitSettingsService() // 系统配置缓存
	services.InitSMSService()      // 短信发送服务
	if err := apilog.Start(); err != nil {
		// WAL 目录不可用不能阻断业务服务启动；中间件会记录入队失败，运维可据此修复目录权限。
		log.Printf("[AppInit] API访问日志 writer 启动失败: %v", err)
	}

	// ---------- 5. 后台异步任务（统一走 internal/task 管理器） ----------
	services.StartBackgroundTasks()

	log.Println("[AppInit] Bootstrap 完成")
}

// SetupHTTP 挂载全局中间件、业务路由与插件。
// disableSlashRedirect：打包托管前端静态资源时建议 true，避免 SPA 尾斜杠 301 循环。
// 返回插件管理器，供外层优雅关闭时调用 ShutdownAll。
func SetupHTTP(router *gin.Engine, disableSlashRedirect bool) *plugins.Manager {
	if router == nil {
		log.Fatal("[AppInit] SetupHTTP: router 不能为空")
	}

	// 静态托管场景关闭尾斜杠自动重定向
	if disableSlashRedirect {
		router.RedirectTrailingSlash = false
		router.RedirectFixedPath = false
	}

	// ---------- 全局中间件 ----------
	// 请求日志统一走 LoggerMiddleware（可配置跳过路径），不再叠加 gin.Logger
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)
	router.Use(middleware.RequestIDMiddleware()) // 尽早写入 X-Request-Id，供日志 /ready 等使用
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CorsMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.APIAccessLogMiddleware())
	router.Use(middleware.DynamicGlobalRateLimitMiddleware())

	// ---------- 业务路由（public / user / admin） ----------
	routes.SetupRoutes(router)
	log.Println("[AppInit] 业务路由加载完成")

	// ---------- 插件系统 ----------
	// 插件通过根入口 blank import 的 init() 注册到全局表，此处统一装载
	pluginMgr := plugins.NewManager()
	plugins.AutoRegisterAll(pluginMgr)
	if err := pluginMgr.LoadAll(); err != nil {
		log.Printf("[Plugin] 插件加载失败: %v", err)
	}

	apiGroup := router.Group("/api/v1")
	pluginMgr.RegisterAllRoutes(apiGroup)
	log.Println("[AppInit] 插件路由加载完成")

	log.Println("[AppInit] Appinit 完成")

	return pluginMgr
}
