package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// 临时脚本：删除角色管理、权限管理、菜单管理相关的数据库表和数据

func main() {
	// 加载 .env 配置
	if err := godotenv.Load(".env"); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	// 数据库配置
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "root")
	dbName := getEnv("DB_NAME", "fst_platform")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	fmt.Println("========================================")
	fmt.Println("开始清理角色管理、权限管理、菜单管理相关数据...")
	fmt.Println("========================================")

	// 要删除的表列表
	tables := []string{
		"roles",
		"permissions",
		"menus",
		"role_permissions",
		"user_roles",
		"role_menus",
		"dicts",
		"dict_items",
	}

	// 删除表
	for _, table := range tables {
		result := db.MustExec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table))
		if result != nil {
			fmt.Printf("✅ 成功删除表: %s\n", table)
		}
	}

	// 检查 users 表中是否有 role_id 字段，如果有则删除
	var columnExists int
	err = db.Get(&columnExists, "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = 'users' AND column_name = 'role_id'", dbName)
	if err == nil && columnExists > 0 {
		db.MustExec("ALTER TABLE `users` DROP COLUMN `role_id`")
		fmt.Println("✅ 成功删除 users.role_id 字段")
	}

	// 删除可能存在的权限相关数据
	db.MustExec("DELETE FROM operation_logs WHERE module IN ('角色管理', '权限管理', '菜单管理', '字典管理')")
	fmt.Println("✅ 已清理操作日志中的相关记录")

	fmt.Println("========================================")
	fmt.Println("清理完成!")
	fmt.Println("========================================")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
