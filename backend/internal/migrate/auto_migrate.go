// Package migrate 数据库自迁移统一入口（系统内部骨架）。
//
// 只放「建表 / 修字段 / 修索引 / 种子 Init*」，不放业务 API。
//
// 调用链：
//
//	cmd/server → appinit.Bootstrap → db.InitDB（只连库）→ migrate.RunAutoMigrate（本包）
//
// 新增表时：model 里写好 Init*，再挂到 RunAutoMigrate 末尾即可。
package migrate

import (
	"log"

	"fst/backend/app/models"
	"fst/backend/pkg/db"
)

type columnRepair struct {
	Column   string
	AlterSQL string
}

type indexRepair struct {
	Index    string
	AlterSQL string
}

// RunAutoMigrate 统一执行全部数据库自迁移与种子初始化。
// 含：核心表 SQL 建表/补列/补索引 + 各业务表 Init*。
func RunAutoMigrate() {
	log.Println("[Migrate] 开始数据库自迁移...")

	migrateCoreSchemas()
	migrateBusinessTables()

	log.Println("[Migrate] 数据库自迁移全部完成")
}

// migrateCoreSchemas 核心表：CREATE IF NOT EXISTS + 历史字段/索引修复。
// 从原 pkg/db.Migrate 收拢至此，连接层不再顺带跑迁移。
func migrateCoreSchemas() {
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
			admin_remark TEXT COMMENT '管理员备注',
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
			user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联用户ID（匿名发送为0）',
			to_email VARCHAR(150) NOT NULL COMMENT '收件人',
			subject VARCHAR(255) NOT NULL COMMENT '主题',
			content TEXT NOT NULL COMMENT '内容',
			template_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '模板名称',
			status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0=失败,1=成功',
			error_msg TEXT COMMENT '错误信息',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			INDEX idx_email_logs_to (to_email),
			INDEX idx_email_logs_user_id (user_id),
			INDEX idx_email_logs_status_created (status, created_at),
			INDEX idx_email_logs_template_name (template_name),
			INDEX idx_email_logs_created_at (created_at)
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
			contact VARCHAR(255) NOT NULL COMMENT '联系方式(邮箱或手机号)',
			code VARCHAR(10) NOT NULL COMMENT '验证码',
			code_type VARCHAR(20) NOT NULL COMMENT '类型:register=注册,reset_password=重置密码',
			expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
			is_used TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已使用:0=未使用,1=已使用',
			is_deleted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否软删除:0=正常,1=已删除',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			INDEX idx_contact_type (contact, code_type),
			INDEX idx_contact_type_active_created (contact, code_type, is_used, is_deleted, created_at),
			INDEX idx_contact_code_type_active (contact, code, code_type, is_used, is_deleted),
			INDEX idx_expires_at (expires_at),
			INDEX idx_is_deleted (is_deleted)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS user_realname_verifications (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
			real_name VARCHAR(100) NOT NULL COMMENT '真实姓名',
			certificate_type TINYINT UNSIGNED NOT NULL COMMENT '证件类型:1=身份证,2=护照,3=军官证',
			certificate_no VARCHAR(50) NOT NULL COMMENT '证件号码',
			certificate_front VARCHAR(500) NOT NULL COMMENT '证件正面照',
			certificate_back VARCHAR(500) NOT NULL COMMENT '证件背面照',
			status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0=待审核,1=通过,2=拒绝',
			reject_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '拒绝原因',
			submitted_at BIGINT UNSIGNED NULL COMMENT '提交时间',
			reviewed_at BIGINT UNSIGNED NULL COMMENT '审核时间',
			reviewed_by BIGINT UNSIGNED NULL COMMENT '审核人ID',
			create_time BIGINT UNSIGNED NULL COMMENT '创建时间',
			update_time BIGINT UNSIGNED NULL COMMENT '更新时间',
			delete_time BIGINT UNSIGNED NULL COMMENT '删除时间',
			INDEX idx_user_id (user_id),
			INDEX idx_status (status),
			INDEX idx_submitted_at (submitted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS auto_job_definitions (
			job_code VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '任务业务主键',
			name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '名称',
			description VARCHAR(512) NOT NULL DEFAULT '' COMMENT '描述',
			category VARCHAR(32) NOT NULL DEFAULT 'maintenance' COMMENT '分类 cleanup/maintenance',
			handler_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '代码注册 handler key',
			cron_expr VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'cron 表达式(5段)',
			interval_seconds INT NOT NULL DEFAULT 0 COMMENT '间隔秒，>0 时优先于 cron',
			timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '时区',
			enabled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否启用',
			timeout_sec INT NOT NULL DEFAULT 300 COMMENT '超时秒',
			max_concurrency INT NOT NULL DEFAULT 1 COMMENT '最大并发',
			params_json TEXT COMMENT '任务参数JSON',
			last_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '最近状态',
			last_started_at BIGINT NOT NULL DEFAULT 0 COMMENT '最近开始',
			last_finished_at BIGINT NOT NULL DEFAULT 0 COMMENT '最近结束',
			last_error TEXT COMMENT '最近错误',
			lifetime_run_count DECIMAL(30,0) NOT NULL DEFAULT 0 COMMENT '终身执行次数',
			lifetime_success_count DECIMAL(30,0) NOT NULL DEFAULT 0 COMMENT '终身成功次数',
			lifetime_fail_count DECIMAL(30,0) NOT NULL DEFAULT 0 COMMENT '终身失败次数',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			update_time BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
			INDEX idx_auto_job_def_category (category),
			INDEX idx_auto_job_def_enabled (enabled),
			INDEX idx_auto_job_def_handler (handler_key)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动任务定义';`,
		`CREATE TABLE IF NOT EXISTS auto_job_runs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			run_uid CHAR(36) NOT NULL DEFAULT '' COMMENT '对外UUID',
			job_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '任务code',
			category VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分类快照',
			trigger_type VARCHAR(32) NOT NULL DEFAULT 'schedule' COMMENT 'schedule/manual',
			status VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'success/failed/timeout',
			started_at BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
			finished_at BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
			duration_ms BIGINT NOT NULL DEFAULT 0 COMMENT '耗时毫秒',
			message VARCHAR(512) NOT NULL DEFAULT '' COMMENT '短文案',
			detail_json MEDIUMTEXT COMMENT '详情JSON',
			error_text TEXT COMMENT '错误文本',
			keep_forever TINYINT(1) NOT NULL DEFAULT 0 COMMENT '清理时保留',
			operator VARCHAR(100) NOT NULL DEFAULT '' COMMENT '手动执行人',
			INDEX idx_auto_job_runs_started_id (started_at, id),
			INDEX idx_auto_job_runs_job_started (job_code, started_at),
			INDEX idx_auto_job_runs_status_started (status, started_at),
			INDEX idx_auto_job_runs_category_started (category, started_at),
			INDEX idx_auto_job_runs_run_uid (run_uid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动任务执行记录';`,
	}

	for _, schema := range schemas {
		// 走 db.Exec：SQLite 下自动把 MySQL DDL 适配成可执行语句
		if _, err := db.Exec(schema); err != nil {
			log.Fatalf("[Migrate] 执行建表 SQL 失败: %v", err)
		}
	}

	// users 表缺失字段自动补齐（按 AFTER 顺序）
	if db.CheckTableExists("users") {
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
			{"admin_remark", "ALTER TABLE users ADD COLUMN admin_remark TEXT COMMENT '管理员备注' AFTER motto"},
			{"apikey", "ALTER TABLE users ADD COLUMN apikey VARCHAR(255) NULL DEFAULT NULL COMMENT 'API密钥' AFTER status"},
			{"language", "ALTER TABLE users ADD COLUMN language VARCHAR(20) NOT NULL DEFAULT 'zh-CN' COMMENT '语言' AFTER apikey"},
			{"country", "ALTER TABLE users ADD COLUMN country VARCHAR(50) NOT NULL DEFAULT '' COMMENT '国家' AFTER language"},
			{"token", "ALTER TABLE users ADD COLUMN token VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Token' AFTER country"},
			{"update_time", "ALTER TABLE users ADD COLUMN update_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '更新时间' AFTER token"},
			{"create_time", "ALTER TABLE users ADD COLUMN create_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '创建时间' AFTER update_time"},
			{"delete_time", "ALTER TABLE users ADD COLUMN delete_time BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '删除时间' AFTER create_time"},
		}

		for _, r := range repairs {
			if !db.CheckColumnExists("users", r.Column) {
				log.Printf("[Migrate] users 补列 %s ...", r.Column)
				if _, err := db.Exec(r.AlterSQL); err != nil {
					log.Printf("[Migrate] users 补列 %s 失败: %v", r.Column, err)
				}
			}
		}

		if db.CheckColumnExists("users", "admin_remark") {
			if _, err := db.Exec("UPDATE users SET admin_remark = '' WHERE admin_remark IS NULL"); err != nil {
				log.Printf("[Migrate] 规范化 users.admin_remark 失败: %v", err)
			}
		}
	}

	if db.CheckTableExists("email_logs") {
		for _, r := range []indexRepair{
			{"idx_email_logs_status_created", "ALTER TABLE email_logs ADD INDEX idx_email_logs_status_created (status, created_at)"},
			{"idx_email_logs_template_name", "ALTER TABLE email_logs ADD INDEX idx_email_logs_template_name (template_name)"},
			{"idx_email_logs_created_at", "ALTER TABLE email_logs ADD INDEX idx_email_logs_created_at (created_at)"},
		} {
			db.EnsureIndex("email_logs", r.Index, r.AlterSQL)
		}
	}

	if db.CheckTableExists("verification_codes") {
		if db.CheckColumnExists("verification_codes", "email") && !db.CheckColumnExists("verification_codes", "contact") {
			if _, err := db.Exec("ALTER TABLE verification_codes CHANGE COLUMN email contact VARCHAR(255) NOT NULL COMMENT '联系方式(邮箱或手机号)'"); err != nil {
				log.Printf("[Migrate] verification_codes.email→contact 失败: %v", err)
			} else if !db.IsSQLite() {
				log.Println("[Migrate] 已重命名 verification_codes.email → contact")
			}
		}
		for _, r := range []indexRepair{
			{"idx_contact_type_active_created", "ALTER TABLE verification_codes ADD INDEX idx_contact_type_active_created (contact, code_type, is_used, is_deleted, created_at)"},
			{"idx_contact_code_type_active", "ALTER TABLE verification_codes ADD INDEX idx_contact_code_type_active (contact, code, code_type, is_used, is_deleted)"},
		} {
			db.EnsureIndex("verification_codes", r.Index, r.AlterSQL)
		}
	}

	log.Println("[Migrate] 核心表 SQL 迁移完成")
}

// migrateBusinessTables 业务表 Init*（各 model 自己负责 CREATE/补结构）。
func migrateBusinessTables() {
	// ---------- 邮件模板种子 ----------
	models.InitEmailTemplates()

	// ---------- 验证码 / 系统配置 / 用户设置 / 会话 ----------
	models.InitVerificationCodeTable()
	models.InitSystemSettingsTable()
	models.InitUserSettingsTable()
	models.InitUserSessionsTable()

	// ---------- 余额 / 积分 / 操作日志 / API 访问日志 ----------
	models.InitUserMoneyLogsTable()
	models.InitUserScoreLogsTable()
	models.InitOperationLogsTable()
	models.InitAPIAccessLogsTable()
	models.InitAPIAccessLogAggregateTables()

	// ---------- 短信日志 / 短信模板 ----------
	models.InitSMSTable()
	models.InitSMSTemplatesTable()
	models.InitSMSTemplates()

	// ---------- 支付 / 提现 / 幂等键 / 支付通道 ----------
	models.InitPaymentOrdersTable()
	models.InitWithdrawRequestsTable()
	models.InitIdempotencyKeysTable()
	models.InitPayGatewaysTable()

	log.Println("[Migrate] 业务表 Init* 完成")
}
