package models

import (
	"fst/backend/internal/db"
	"log"
	"time"
)

// UserSession 用户会话模型
type UserSession struct {
	ID        uint64 `db:"id" json:"id"`
	UserID    uint64 `db:"user_id" json:"user_id"`
	TokenHash string `db:"token_hash" json:"-"`
	IP        string `db:"ip" json:"ip"`
	UserAgent string `db:"user_agent" json:"user_agent"`
	Device    string `db:"device" json:"device"`
	IsActive  bool   `db:"is_active" json:"is_active"`
	LoginAt   int64  `db:"login_at" json:"login_at"`
	ExpiresAt int64  `db:"expires_at" json:"expires_at"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

// InitUserSessionsTable 初始化用户会话表
func InitUserSessionsTable() {
	if db.CheckTableExists("user_sessions") {
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS user_sessions (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
		token_hash VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Token哈希',
		ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '登录IP',
		user_agent TEXT COMMENT '浏览器UA',
		device VARCHAR(100) NOT NULL DEFAULT '' COMMENT '设备信息',
		is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否活跃',
		login_at BIGINT NOT NULL DEFAULT 0 COMMENT '登录时间',
		expires_at BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间',
		created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		INDEX idx_user_id (user_id),
		INDEX idx_is_active (is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	_, err := db.DB.Exec(schema)
	if err != nil {
		log.Printf("[Init] Failed to create user_sessions table: %v", err)
	} else {
		log.Println("[Init] Created user_sessions table")
	}
}

// CreateUserSession 创建用户会话记录
func CreateUserSession(userID uint64, tokenHash, ip, userAgent, device string, expiresAt int64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec(
		`INSERT INTO user_sessions (user_id, token_hash, ip, user_agent, device, is_active, login_at, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)`,
		userID, tokenHash, ip, userAgent, device, now, expiresAt, now,
	)
	return err
}

// GetUserSessions 获取用户的活跃会话列表
func GetUserSessions(userID uint64) ([]UserSession, error) {
	var sessions []UserSession
	now := time.Now().Unix()
	err := db.DB.Select(&sessions,
		`SELECT id, user_id, ip, user_agent, device, is_active, login_at, expires_at, created_at
		 FROM user_sessions
		 WHERE user_id = ? AND is_active = 1 AND expires_at > ?
		 ORDER BY login_at DESC LIMIT 50`,
		userID, now,
	)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// RevokeUserSession 撤销指定会话
func RevokeUserSession(userID uint64, sessionID string) error {
	_, err := db.DB.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE id = ? AND user_id = ?",
		sessionID, userID,
	)
	return err
}

// RevokeAllUserSessions 撤销用户所有会话（除当前）
func RevokeAllUserSessions(userID uint64, currentTokenHash string) error {
	if currentTokenHash != "" {
		_, err := db.DB.Exec(
			"UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND token_hash != ?",
			userID, currentTokenHash,
		)
		return err
	}
	_, err := db.DB.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE user_id = ?",
		userID,
	)
	return err
}

// CleanupExpiredSessions 清理过期会话
func CleanupExpiredSessions() error {
	now := time.Now().Unix()
	_, err := db.DB.Exec(
		"DELETE FROM user_sessions WHERE expires_at < ? OR is_active = 0",
		now,
	)
	return err
}

// GetUserLoginCount 获取用户登录次数
func GetUserLoginCount(userID uint64) (int64, error) {
	var count int64
	err := db.DB.Get(&count,
		"SELECT COUNT(*) FROM user_sessions WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}
