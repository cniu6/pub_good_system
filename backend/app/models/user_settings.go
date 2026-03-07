package models

import (
	"fst/backend/internal/db"
	"log"
	"time"
)

// UserSettings 用户设置模型
type UserSettings struct {
	ID          uint64 `db:"id" json:"id"`
	UserID      uint64 `db:"user_id" json:"user_id"`
	Theme       string `db:"theme" json:"theme"`
	NotifyEmail bool   `db:"notify_email" json:"notify_email"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
	UpdatedAt   int64  `db:"updated_at" json:"updated_at"`
}

// InitUserSettingsTable 初始化用户设置表
func InitUserSettingsTable() {
	if db.CheckTableExists("user_settings") {
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS user_settings (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
		theme VARCHAR(20) NOT NULL DEFAULT 'light' COMMENT '主题:light/dark',
		notify_email TINYINT(1) NOT NULL DEFAULT 1 COMMENT '邮件通知:0=关闭,1=开启',
		created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		updated_at BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
		UNIQUE KEY idx_user_id (user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	_, err := db.DB.Exec(schema)
	if err != nil {
		log.Printf("[Init] Failed to create user_settings table: %v", err)
	} else {
		log.Println("[Init] Created user_settings table")
	}
}

// GetUserSettings 获取用户设置
func GetUserSettings(userID uint64) (*UserSettings, error) {
	var settings UserSettings
	err := db.DB.Get(&settings, "SELECT * FROM user_settings WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// SaveUserSettings 保存用户设置（upsert）
func SaveUserSettings(settings *UserSettings) error {
	now := time.Now().Unix()
	settings.UpdatedAt = now

	// 尝试更新
	result, err := db.DB.Exec(
		"UPDATE user_settings SET theme = ?, notify_email = ?, updated_at = ? WHERE user_id = ?",
		settings.Theme, settings.NotifyEmail, now, settings.UserID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		// 不存在则插入
		settings.CreatedAt = now
		_, err = db.DB.Exec(
			"INSERT INTO user_settings (user_id, theme, notify_email, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			settings.UserID, settings.Theme, settings.NotifyEmail, now, now,
		)
		return err
	}

	return nil
}
