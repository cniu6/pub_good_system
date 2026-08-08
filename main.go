//go:build !embedded

package main

import (
	"fst/backend/cmd/server"

	// ========================================
	// 插件自动导入区域
	// 运行 gen-swagger.bat 会自动扫描并更新此区域
	// ========================================
	// @plugins-start
	_ "fst/backend/app/plugins/pay_balance"
	_ "fst/backend/app/plugins/sms"
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
	// 开发/普通运行：仅后端 API，不托管前端
	server.Start(server.Options{BuildMode: "none"})
}
