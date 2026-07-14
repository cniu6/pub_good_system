package appinit

import (
	"fst/backend/internal/db"
	"fst/backend/internal/env"
	"fst/backend/internal/ginweb"
	"fst/backend/models"
	"log"
)

// Start 启动草稿应用栈（GORM + 独立路由）
// 注意：正式线上入口仍是 backend/cmd/main.go / 根 main.go，本包默认不被主程序 import
func Start() {
	log.Println("")
	log.Println("=========================================")
	log.Println("开始使用 F.st 平台（草稿栈）")
	log.Println("F.st - Think Fast, Run F.st")
	log.Println("=========================================")
	log.Println("")

	log.Println("[1/5] 初始化环境变量...")
	envConfig := env.InitEnv()
	// 同步 JWT 密钥到现网 config，保证 utils.GenerateToken / ParseToken 可用
	ensureJWTConfig(envConfig.JWTSecret)
	log.Printf("  - 数据库: %s:%d/%s", envConfig.DBHost, envConfig.DBPort, envConfig.DBName)
	log.Printf("  - API端口: %d", envConfig.APIPort)

	log.Println("[2/5] 初始化数据库...")
	db.InitDB()

	log.Println("[3/5] 自动迁移数据库表结构...")
	AutoMigrate()

	log.Println("[4/5] 初始化默认设置...")
	initDefaultSettings()

	log.Println("[5/5] 启动Web服务...")
	ginweb.InitGin()
}

// AutoMigrate 迁移草稿 models 表结构
func AutoMigrate() {
	gormDB := db.GetGormDB()
	err := gormDB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Category{},
		&models.Order{},
		&models.Settings{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("  - 已迁移: users, products, categories, orders, settings")
	log.Println("  - 数据库迁移完成")
}

// initDefaultSettings 占位：可写入默认站点配置
func initDefaultSettings() {
	log.Println("  - 默认设置初始化完成")
}
