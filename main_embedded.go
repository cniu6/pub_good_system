//go:build embedded

package main

import (
	"embed"

	"fst/backend/cmd/server"

	// ========================================
	// 插件自动导入区域（与非 embedded 入口保持一致）
	// ========================================
	_ "fst/backend/app/plugins/demo"
	_ "fst/backend/app/plugins/pay_balance"
	_ "fst/backend/app/plugins/sms"
)

// BuildMode 由构建脚本在编译时注入: "embedded" 或 "external" 或 "none"
// 根目录入口默认 embedded；build.bat 会通过 -ldflags 覆盖
var BuildMode = "embedded"

//go:embed dist/*
var frontendFS embed.FS

// @title FST Platform API
// @version 1.0
// @description FST Platform 后端 API 接口文档
// @host localhost:8080
// @BasePath /api

func main() {
	// 打包运行：通过 BuildMode 区分内嵌 / 外置前端资源
	server.Start(server.Options{
		BuildMode:  BuildMode,
		FrontendFS: frontendFS,
	})
}
