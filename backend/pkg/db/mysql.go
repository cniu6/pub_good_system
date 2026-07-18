package db

import (
	"fst/backend/pkg/config"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DB 全局 sqlx 连接。建表/自迁移请走 internal/migrate.RunAutoMigrate，不要在这里顺带 Migrate。
var DB *sqlx.DB

// InitDB 仅建立数据库连接与连接池，不执行表迁移。
func InitDB() {
	var err error
	cfg := config.GlobalConfig
	if cfg == nil {
		log.Fatalf("[数据库配置错误] 数据库配置未初始化")
	}
	if cfg.DBDriver != "mysql" {
		log.Fatalf("[数据库配置错误] 当前仅支持 mysql 数据库驱动，收到: %s", cfg.DBDriver)
	}
	DB, err = sqlx.Connect(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("[数据库连接错误] 无法连接数据库，请检查数据库服务和配置: %v", err)
	}

	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(10)

	log.Println("[DB] 数据库连接已建立")
}

// CheckTableExists 当前库是否存在指定表。
func CheckTableExists(tableName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
			  WHERE table_schema = DATABASE() AND table_name = ?`
	err := DB.Get(&count, query, tableName)
	return err == nil && count > 0
}

// CheckColumnExists 指定表是否存在指定列。
func CheckColumnExists(tableName, columnName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND column_name = ?`
	err := DB.Get(&count, query, tableName, columnName)
	return err == nil && count > 0
}

// CheckIndexExists 指定表是否存在指定索引。
func CheckIndexExists(tableName, indexName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.statistics 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND index_name = ?`
	err := DB.Get(&count, query, tableName, indexName)
	return err == nil && count > 0
}

// EnsureIndex 表存在且索引缺失时执行 ALTER 补索引。
func EnsureIndex(tableName, indexName, alterSQL string) {
	if !CheckTableExists(tableName) || CheckIndexExists(tableName, indexName) {
		return
	}
	_, err := DB.Exec(alterSQL)
	if err != nil {
		log.Printf("[DB] 补索引 '%s' on '%s' 失败: %v", indexName, tableName, err)
	} else {
		log.Printf("[DB] 已补索引 '%s' on '%s'", indexName, tableName)
	}
}

// GetDB 返回全局 sqlx 连接。
func GetDB() *sqlx.DB {
	return DB
}
