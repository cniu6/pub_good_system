package db

import (
	"fst/backend/internal/config"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitDB() {
	var err error
	DB, err = sqlx.Connect(config.GlobalConfig.DBDriver, config.GlobalConfig.DBDSN)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(10)

	log.Println("Database connection established")

	// Run migrations
	Migrate()
}

type columnRepair struct {
	Column   string
	AlterSQL string
}

func Migrate() {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			group_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分组ID',
			username VARCHAR(100) NOT NULL COMMENT '用户名',
			nickname VARCHAR(100) NOT NULL DEFAULT '' COMMENT '昵称',
			email VARCHAR(150) NOT NULL COMMENT '邮箱',
			mobile VARCHAR(50) NOT NULL DEFAULT '' COMMENT '手机',
			avatar VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像',
			back_ground VARCHAR(255) NOT NULL DEFAULT '' COMMENT '背景',
			gender TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '性别:0=未知,1=男,2=女',
			birthday BIGINT NULL DEFAULT NULL COMMENT '生日',
			money DECIMAL(10,2) NOT NULL DEFAULT '0.00' COMMENT '余额',
			score BIGINT NOT NULL DEFAULT 0 COMMENT '积分',
			level BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '用户等级',
			role VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色:user=普通用户,admin=管理员',
			last_login_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '上次登录时间',
			last_login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上次登录IP',
			login_failure TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '登录失败次数',
			lock_until BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '账户锁定到期时间',
			join_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '加入IP',
			join_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '加入时间',
			motto VARCHAR(255) NOT NULL DEFAULT '' COMMENT '签名',
			password VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
			apikey VARCHAR(255) NULL DEFAULT NULL COMMENT 'API密钥',
			language VARCHAR(20) NOT NULL DEFAULT 'zh-CN' COMMENT '语言',
			country VARCHAR(50) NOT NULL DEFAULT '' COMMENT '国家',
			token VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Token',
			update_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '更新时间',
			create_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '创建时间',
			delete_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '删除时间',
			UNIQUE KEY idx_users_username (username),
			UNIQUE KEY idx_users_email (email),
			UNIQUE KEY idx_users_api_key (apikey),
			INDEX idx_users_mobile (mobile),
			INDEX idx_users_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS email_logs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			to_email VARCHAR(150) NOT NULL COMMENT '收件人',
			subject VARCHAR(255) NOT NULL COMMENT '主题',
			content TEXT NOT NULL COMMENT '内容',
			template_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '模板名称',
			status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0=失败,1=成功',
			error_msg TEXT COMMENT '错误信息',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			INDEX idx_email_logs_to (to_email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS email_templates (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL COMMENT '模板标识',
			lang VARCHAR(20) NOT NULL DEFAULT 'zh-CN' COMMENT '语言',
			title VARCHAR(100) NOT NULL COMMENT '模板标题',
			subject VARCHAR(255) NOT NULL COMMENT '邮件主题',
			content TEXT NOT NULL COMMENT '邮件内容(支持HTML)',
			description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
			variables VARCHAR(500) NOT NULL DEFAULT '' COMMENT '可用变量说明',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY idx_tpl_name_lang (name, lang)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS verification_codes (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL COMMENT '邮箱地址',
			code VARCHAR(10) NOT NULL COMMENT '验证码',
			code_type VARCHAR(20) NOT NULL COMMENT '类型:register=注册,reset_password=重置密码',
			expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
			is_used TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已使用:0=未使用,1=已使用',
			is_deleted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否软删除:0=正常,1=已删除',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			INDEX idx_email_type (email, code_type),
			INDEX idx_expires_at (expires_at),
			INDEX idx_is_deleted (is_deleted)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, schema := range schemas {
		_, err := DB.Exec(schema)
		if err != nil {
			log.Fatalf("Error running migration: %v", err)
		}
	}

	// 自动修复 users 表缺失的字段 (按顺序排列确保 AFTER 可用)
	if CheckTableExists("users") {
		repairs := []columnRepair{
			{"group_id", "ALTER TABLE users ADD COLUMN group_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分组ID' AFTER id"},
			{"nickname", "ALTER TABLE users ADD COLUMN nickname VARCHAR(100) NOT NULL DEFAULT '' COMMENT '昵称' AFTER username"},
			{"mobile", "ALTER TABLE users ADD COLUMN mobile VARCHAR(50) NOT NULL DEFAULT '' COMMENT '手机' AFTER email"},
			{"avatar", "ALTER TABLE users ADD COLUMN avatar VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像' AFTER mobile"},
			{"back_ground", "ALTER TABLE users ADD COLUMN back_ground VARCHAR(255) NOT NULL DEFAULT '' COMMENT '背景' AFTER avatar"},
			{"gender", "ALTER TABLE users ADD COLUMN gender TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '性别:0=未知,1=男,2=女' AFTER back_ground"},
			{"birthday", "ALTER TABLE users ADD COLUMN birthday BIGINT NULL DEFAULT NULL COMMENT '生日' AFTER gender"},
			{"money", "ALTER TABLE users ADD COLUMN money DECIMAL(10,2) NOT NULL DEFAULT '0.00' COMMENT '余额' AFTER birthday"},
			{"score", "ALTER TABLE users ADD COLUMN score BIGINT NOT NULL DEFAULT 0 COMMENT '积分' AFTER money"},
			{"level", "ALTER TABLE users ADD COLUMN level BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '用户等级' AFTER score"},
			{"status", "ALTER TABLE users ADD COLUMN status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用' AFTER password"},
			{"last_login_time", "ALTER TABLE users ADD COLUMN last_login_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '上次登录时间' AFTER role"},
			{"last_login_ip", "ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上次登录IP' AFTER last_login_time"},
			{"login_failure", "ALTER TABLE users ADD COLUMN login_failure TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '登录失败次数' AFTER last_login_ip"},
			{"lock_until", "ALTER TABLE users ADD COLUMN lock_until BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '账户锁定到期时间' AFTER login_failure"},
			{"join_ip", "ALTER TABLE users ADD COLUMN join_ip VARCHAR(50) NOT NULL DEFAULT '' COMMENT '加入IP' AFTER lock_until"},
			{"join_time", "ALTER TABLE users ADD COLUMN join_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '加入时间' AFTER join_ip"},
			{"motto", "ALTER TABLE users ADD COLUMN motto VARCHAR(255) NOT NULL DEFAULT '' COMMENT '签名' AFTER join_time"},
			{"apikey", "ALTER TABLE users ADD COLUMN apikey VARCHAR(255) NULL DEFAULT NULL COMMENT 'API密钥' AFTER status"},
			{"language", "ALTER TABLE users ADD COLUMN language VARCHAR(20) NOT NULL DEFAULT 'zh-CN' COMMENT '语言' AFTER apikey"},
			{"country", "ALTER TABLE users ADD COLUMN country VARCHAR(50) NOT NULL DEFAULT '' COMMENT '国家' AFTER language"},
			{"token", "ALTER TABLE users ADD COLUMN token VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Token' AFTER country"},
			{"update_time", "ALTER TABLE users ADD COLUMN update_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '更新时间' AFTER token"},
			{"create_time", "ALTER TABLE users ADD COLUMN create_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '创建时间' AFTER update_time"},
			{"delete_time", "ALTER TABLE users ADD COLUMN delete_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '删除时间' AFTER create_time"},
		}

		for _, r := range repairs {
			if !CheckColumnExists("users", r.Column) {
				log.Printf("[Init] Adding missing column '%s' to 'users' table...", r.Column)
				_, err := DB.Exec(r.AlterSQL)
				if err != nil {
					log.Printf("[Init] Failed to add column '%s': %v", r.Column, err)
				} else {
					log.Printf("[Init] Successfully added column '%s'", r.Column)
				}
			}
		}
	}

	log.Println("Database migration completed")
}

func CheckTableExists(tableName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
			  WHERE table_schema = DATABASE() AND table_name = ?`
	err := DB.Get(&count, query, tableName)
	return err == nil && count > 0
}

func CheckColumnExists(tableName, columnName string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND column_name = ?`
	err := DB.Get(&count, query, tableName, columnName)
	return err == nil && count > 0
}

func GetDB() *sqlx.DB {
	return DB
}
