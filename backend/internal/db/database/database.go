package database

import (
	"fmt"
	"fst/backend/internal/env"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// db 全局 GORM 连接（草稿独立栈，与现网 sqlx 分离）
var db *gorm.DB

// InitDB 初始化 MySQL 连接
func InitDB() *gorm.DB {
	config := env.GetEnv()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		SkipDefaultTransaction: false,
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接池失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	log.Println("数据库连接成功")
	return db
}

// GetDB 获取 GORM 实例（未初始化则自动 InitDB）
func GetDB() *gorm.DB {
	if db == nil {
		return InitDB()
	}
	return db
}
