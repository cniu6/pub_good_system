package server

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"time"

	"fst/backend/internal/appinit"
	"fst/backend/pkg/apilog"
	"fst/backend/pkg/config"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// Options 启动选项：由根目录 main / main_embedded 传入，区分开发与打包。
//
// BuildMode 取值：
//   - "none"     ：仅后端 API（开发默认，不托管前端）
//   - "embedded" ：从二进制内嵌 FS 托管前端
//   - "external" ：从 ./dist 目录托管前端
type Options struct {
	BuildMode  string
	FrontendFS embed.FS // 仅 embedded 模式需要；none/external 可零值
}

// Start 进程入口薄壳：编排交给 internal/appinit，本包只负责前端托管与 Listen。
func Start(opts Options) {
	if opts.BuildMode == "" {
		opts.BuildMode = "none"
	}

	log.Println("[Server] ===================启动中===================")

	// 1. 系统加载：配置 / DB / 自迁移 / 服务 / 定时任务
	appinit.Bootstrap()

	// 2. HTTP 引擎 + 中间件 / 路由 / 插件
	router := gin.New()
	disableSlashRedirect := opts.BuildMode != "none" // 打包托管前端时关闭尾斜杠重定向
	pluginMgr := appinit.SetupHTTP(router, disableSlashRedirect)

	// 3. 前端静态资源（仅 embedded / external）
	mountFrontend(router, opts)

	// 4. 监听与优雅关闭
	port := config.GlobalConfig.Port
	httpServer := utils.NewHTTPServer(":"+port, router)
	log.Printf("[Server] 服务启动，端口: %s，BuildMode=%s", port, opts.BuildMode)
	if opts.BuildMode == "none" {
		log.Printf("[Server] 服务启动成功，端口: %s，BuildMode=%s", port, opts.BuildMode)
		log.Printf("[Server] Scalar 文档: http://localhost:%s/scalar", port)
	}
	log.Printf("[Server] 已加载插件数量: %d", pluginMgr.Count())
	log.Printf("===================================================")

	if err := utils.ServeHTTPServer(httpServer, func() error {
		// HTTP 已停止接收新请求后再 flush，避免关闭期间继续向队列写入。
		// 预留一部分 HTTP 优雅关闭时间给插件，超时的内存日志会落入 WAL 等待下次启动重放。
		flushTimeout := 8 * time.Second
		if cfg := config.CloneGlobalConfig(); cfg != nil && cfg.HTTPShutdownTimeoutSeconds > 0 {
			flushTimeout = time.Duration(cfg.HTTPShutdownTimeoutSeconds) * time.Second * 8 / 10
		}
		ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
		defer cancel()
		return errors.Join(apilog.Stop(ctx), pluginMgr.ShutdownAll())
	}); err != nil { //启动服务，并监听插件关闭信号
		log.Fatalf("[Server] 启动失败: %v", err)
	}

	log.Printf("===================================================")
	log.Printf("")
	log.Printf("")

}

// mountFrontend 按 BuildMode 挂载前端静态托管；none 模式直接跳过。
func mountFrontend(router *gin.Engine, opts Options) {
	if opts.BuildMode == "none" {
		log.Printf("[Mode] 仅后端 API：BuildMode=%s，不托管前端", opts.BuildMode)
		return
	}

	var rawFS fs.FS
	switch opts.BuildMode {
	case "external":
		log.Println("[Mode] External：从 ./dist 目录托管前端")
		rawFS = openExternalDist()
	case "embedded":
		log.Println("[Mode] Embedded：从二进制内嵌 FS 托管前端")
		distFS, err := fs.Sub(opts.FrontendFS, "dist")
		if err != nil {
			log.Fatalf("[Mode] 加载内嵌前端失败: %v", err)
		}
		rawFS = distFS
	default:
		log.Printf("[Mode] 未知 BuildMode=%s，跳过前端托管", opts.BuildMode)
		return
	}

	if rawFS == nil {
		return
	}
	registerFrontendNoRoute(router, rawFS)
}
