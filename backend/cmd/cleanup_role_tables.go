package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	// 从命令行获取数据库连接信息
	dsn := flag.String("dsn", "", "MySQL DSN (e.g., root:password@tcp(localhost:3306)/fst)")
	flag.Parse()

	if *dsn == "" {
		// 尝试从环境变量读取
		*dsn = os.Getenv("DB_DSN")
		if *dsn == "" {
			log.Fatal("请提供数据库DSN: -dsn=\"root:password@tcp(localhost:3306)/fst\"")
		}
	}

	db, err := sqlx.Connect("mysql", *dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	fmt.Println("=== 开始清理角色、权限、菜单相关表 ===")

	// 要删除的表列表
	tables := []string{
		"roles",
		"permissions",
		"menus",
		"role_permissions",
		"user_roles",
		"role_menus",
	}

	// 获取数据库名称
	var dbName string
	db.Get(&dbName, "SELECT DATABASE()")
	fmt.Printf("当前数据库: %s\n", dbName)

	// 删除表
	for _, table := range tables {
		if tableExists(db, table) {
			fmt.Printf("删除表: %s ... ", table)
			_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table))
			if err != nil {
				fmt.Printf("失败: %v\n", err)
			} else {
				fmt.Println("成功")
			}
		} else {
			fmt.Printf("表不存在: %s (跳过)\n", table)
		}
	}

	// 检查 users 表中是否有 role_id 字段，如果有则删除
	fmt.Println("\n检查 users 表中的 role_id 字段...")
	var columnExists int
	err = db.Get(&columnExists, "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = 'users' AND column_name = 'role_id'", dbName)
if err == nil && columnExists > 0 {
		fmt.Println("删除 users.role_id 字段 ... ")
		db.MustExec("ALTER TABLE `users` DROP COLUMN `role_id`")
		fmt.Println("成功")
	} else {
		fmt.Println("users 表中没有 role_id 字段 (跳过)")
	}

	// 检查 users 表中是否有 role 字段
	fmt.Println("\n检查 users 表中的 role 字段...")
	err = db.Get(&columnExists, "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = 'users' AND column_name = 'role'", dbName)
	if err == nil && columnExists > 0 {
		fmt.Println("保留 users.role 字段（用于简单的admin/user判断）")
	}

	fmt.Println("\n=== 清理完成 ===")
}

func tableExists(db *sqlx.DB, tableName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
			  WHERE table_schema = DATABASE() AND table_name = ?`
	err := db.Get(&count, query, tableName)
	return err == nil && count > 0
}