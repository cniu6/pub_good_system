package appinit

import (
	"fst/backend/internal/db"
	"fst/backend/internal/env"
	"fst/backend/internal/ginweb"
	"fst/backend/models"
	"log"
)

// Start 启动草稿应用栈（GORM + 独立路由）
//
// 注意：正式线上入口仅项目根目录 main.go / main_embedded.go，本包默认不被主程序 import。
// 【已注释禁用】电商商品/分类/订单半成品：路由与 AutoMigrate 相关项已注释，见 backend/留档.md。
func Start() {
	log.Println("")
	log.Println("=========================================")
	log.Println("开始使用 F.st 平台（草稿栈）")
	log.Println("【提示】电商半成品（商品/分类/商城订单）已注释禁用")
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
// 【已注释禁用】products / categories / orders 电商表不迁移，仅保留 users、settings 对照用
func AutoMigrate() {
	gormDB := db.GetGormDB()
	err := gormDB.AutoMigrate(
		&models.User{},
		// &models.Product{},   // 【已注释禁用】电商半成品
		// &models.Category{},  // 【已注释禁用】电商半成品
		// &models.Order{},     // 【已注释禁用】电商半成品
		&models.Settings{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("  - 已迁移: users, settings")
	log.Println("  - 已跳过(注释禁用): products, categories, orders")
	log.Println("  - 数据库迁移完成")
}

// initDefaultSettings 占位：可写入默认站点配置
func initDefaultSettings() {
	log.Println("  - 默认设置初始化完成")
}
