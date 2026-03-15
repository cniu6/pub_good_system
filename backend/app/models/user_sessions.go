package models

import (
	"fst/backend/internal/db"
	"log"
	"time"
)

// UserSession 用户会话模型
type UserSession struct {
	ID               uint64 `db:"id" json:"id"`
	UserID           uint64 `db:"user_id" json:"user_id"`
	AuthGuard        string `db:"auth_guard" json:"auth_guard"`
	TokenHash        string `db:"token_hash" json:"-"`
	RefreshTokenHash string `db:"refresh_token_hash" json:"-"`
	IP               string `db:"ip" json:"ip"`
	UserAgent        string `db:"user_agent" json:"user_agent"`
	Device           string `db:"device" json:"device"`
	IsActive         bool   `db:"is_active" json:"is_active"`
	LoginAt          int64  `db:"login_at" json:"login_at"`
	ExpiresAt        int64  `db:"expires_at" json:"expires_at"`
	RefreshExpiresAt int64  `db:"refresh_expires_at" json:"-"`
	CreatedAt        int64  `db:"created_at" json:"created_at"`
}

// InitUserSessionsTable 初始化用户会话表
func InitUserSessionsTable() {
	if !db.CheckTableExists("user_sessions") {
		schema := `CREATE TABLE IF NOT EXISTS user_sessions (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
			auth_guard VARCHAR(50) NOT NULL DEFAULT 'user' COMMENT '认证上下文 user/admin',
			token_hash VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Access Token哈希',
			refresh_token_hash VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Refresh Token哈希',
			ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '登录IP',
			user_agent TEXT COMMENT '浏览器UA',
			device VARCHAR(100) NOT NULL DEFAULT '' COMMENT '设备信息',
			is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否活跃',
			login_at BIGINT NOT NULL DEFAULT 0 COMMENT '登录时间',
			expires_at BIGINT NOT NULL DEFAULT 0 COMMENT 'Access Token过期时间',
			refresh_expires_at BIGINT NOT NULL DEFAULT 0 COMMENT 'Refresh Token过期时间',
			created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			INDEX idx_user_id (user_id),
			INDEX idx_user_guard (user_id, auth_guard),
			INDEX idx_is_active (is_active),
			INDEX idx_user_token_active_expire (user_id, auth_guard, token_hash, is_active, expires_at),
			INDEX idx_user_refresh_active_expire (user_id, auth_guard, refresh_token_hash, is_active, refresh_expires_at),
			INDEX idx_user_active_login (user_id, auth_guard, is_active, login_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

		_, err := db.DB.Exec(schema)
		if err != nil {
			log.Printf("[Init] Failed to create user_sessions table: %v", err)
		} else {
			log.Println("[Init] Created user_sessions table")
		}
	}

	repairs := map[string]string{
		"auth_guard":         "ALTER TABLE user_sessions ADD COLUMN auth_guard VARCHAR(50) NOT NULL DEFAULT 'user' COMMENT '认证上下文 user/admin' AFTER user_id",
		"refresh_token_hash": "ALTER TABLE user_sessions ADD COLUMN refresh_token_hash VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Refresh Token哈希' AFTER token_hash",
		"refresh_expires_at": "ALTER TABLE user_sessions ADD COLUMN refresh_expires_at BIGINT NOT NULL DEFAULT 0 COMMENT 'Refresh Token过期时间' AFTER expires_at",
	}

	for column, alterSQL := range repairs {
		if !db.CheckColumnExists("user_sessions", column) {
			if _, err := db.DB.Exec(alterSQL); err != nil {
				log.Printf("[Init] Failed to add user_sessions.%s: %v", column, err)
			} else {
				log.Printf("[Init] Added user_sessions.%s", column)
			}
		}
	}

	indexRepairs := map[string]string{
		"idx_user_guard":                 "ALTER TABLE user_sessions ADD INDEX idx_user_guard (user_id, auth_guard)",
		"idx_user_token_active_expire":   "ALTER TABLE user_sessions ADD INDEX idx_user_token_active_expire (user_id, auth_guard, token_hash, is_active, expires_at)",
		"idx_user_refresh_active_expire": "ALTER TABLE user_sessions ADD INDEX idx_user_refresh_active_expire (user_id, auth_guard, refresh_token_hash, is_active, refresh_expires_at)",
		"idx_user_active_login":          "ALTER TABLE user_sessions ADD INDEX idx_user_active_login (user_id, auth_guard, is_active, login_at)",
	}

	for indexName, alterSQL := range indexRepairs {
		db.EnsureIndex("user_sessions", indexName, alterSQL)
	}
}

// CreateUserSession 创建用户会话记录
func CreateUserSession(userID uint64, authGuard, tokenHash, refreshTokenHash, ip, userAgent, device string, expiresAt, refreshExpiresAt int64) error {
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var lockedUserID uint64
	if err = tx.QueryRow("SELECT id FROM users WHERE id = ? FOR UPDATE", userID).Scan(&lockedUserID); err != nil {
		return err
	}

	if _, err = tx.Exec("DELETE FROM user_sessions WHERE user_id = ? AND auth_guard = ?", userID, authGuard); err != nil {
		return err
	}

	if _, err = tx.Exec(
		`INSERT INTO user_sessions (user_id, auth_guard, token_hash, refresh_token_hash, ip, user_agent, device, is_active, login_at, expires_at, refresh_expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)`,
		userID, authGuard, tokenHash, refreshTokenHash, ip, userAgent, device, now, expiresAt, refreshExpiresAt, now,
	); err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func IsUserSessionActive(userID uint64, authGuard, tokenHash string) (bool, error) {
	var count int
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Get(&count,
		`SELECT COUNT(*) FROM user_sessions
		 WHERE user_id = ? AND auth_guard = ? AND token_hash = ? AND is_active = 1 AND expires_at > ?`,
		userID, authGuard, tokenHash, now,
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func IsRefreshSessionActive(userID uint64, authGuard, refreshTokenHash string) (bool, error) {
	var count int
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Get(&count,
		`SELECT COUNT(*) FROM user_sessions
		 WHERE user_id = ? AND auth_guard = ? AND refresh_token_hash = ? AND is_active = 1 AND refresh_expires_at > ?`,
		userID, authGuard, refreshTokenHash, now,
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func RotateUserSessionTokens(userID uint64, authGuard, currentRefreshTokenHash, newTokenHash, newRefreshTokenHash, ip, userAgent, device string, expiresAt, refreshExpiresAt int64) (bool, error) {
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	result, err := db.DB.Exec(
		`UPDATE user_sessions
		 SET token_hash = ?, refresh_token_hash = ?, ip = ?, user_agent = ?, device = ?, expires_at = ?, refresh_expires_at = ?, login_at = ?
		 WHERE user_id = ? AND auth_guard = ? AND refresh_token_hash = ? AND is_active = 1 AND refresh_expires_at > ?`,
		newTokenHash, newRefreshTokenHash, ip, userAgent, device, expiresAt, refreshExpiresAt, now,
		userID, authGuard, currentRefreshTokenHash, now,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// GetUserSessions 获取用户的活跃会话列表
func GetUserSessions(userID uint64) ([]UserSession, error) {
	return GetUserSessionsWithGuard(userID, "user")
}

func GetUserSessionsWithGuard(userID uint64, authGuard string) ([]UserSession, error) {
	var sessions []UserSession
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Select(&sessions,
		`SELECT id, user_id, auth_guard, ip, user_agent, device, is_active, login_at, expires_at, created_at
		 FROM user_sessions
		 WHERE user_id = ? AND auth_guard = ? AND is_active = 1 AND ((refresh_expires_at > 0 AND refresh_expires_at > ?) OR (refresh_expires_at = 0 AND expires_at > ?))
		 ORDER BY login_at DESC LIMIT 50`,
		userID, authGuard, now, now,
	)
	if err != nil {
		return nil, err
	}
	if sessions == nil {
		sessions = []UserSession{}
	}
	return sessions, nil
}

// RevokeUserSession 撤销指定会话
func RevokeUserSession(userID uint64, sessionID string) error {
	return RevokeUserSessionWithGuard(userID, "user", sessionID)
}

func RevokeUserSessionWithGuard(userID uint64, authGuard, sessionID string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	_, err := db.DB.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE id = ? AND user_id = ? AND auth_guard = ?",
		sessionID, userID, authGuard,
	)
	return err
}

// RevokeAllUserSessions 撤销用户所有会话（除当前）
func RevokeAllUserSessions(userID uint64, currentTokenHash string) error {
	return RevokeAllUserSessionsWithGuard(userID, "user", currentTokenHash)
}

func RevokeAllUserSessionsWithGuard(userID uint64, authGuard, currentTokenHash string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	if currentTokenHash != "" {
		_, err := db.DB.Exec(
			"UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND auth_guard = ? AND token_hash != ?",
			userID, authGuard, currentTokenHash,
		)
		return err
	}
	_, err := db.DB.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND auth_guard = ?",
		userID, authGuard,
	)
	return err
}

// CleanupExpiredSessions 清理过期会话
func CleanupExpiredSessions() error {
	now := time.Now().Unix()
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(
		"DELETE FROM user_sessions WHERE is_active = 0 OR (refresh_expires_at > 0 AND refresh_expires_at <= ?) OR (refresh_expires_at = 0 AND expires_at <= ?)",
		now, now,
	); err != nil {
		return err
	}

	if _, err = tx.Exec(
		`DELETE stale FROM user_sessions stale
		 INNER JOIN user_sessions latest
		 	ON stale.user_id = latest.user_id
		 	AND stale.id <> latest.id
		 	AND latest.is_active = 1
		 	AND ((latest.refresh_expires_at > 0 AND latest.refresh_expires_at > ?) OR (latest.refresh_expires_at = 0 AND latest.expires_at > ?))
		 	AND (stale.login_at < latest.login_at OR (stale.login_at = latest.login_at AND stale.id < latest.id))
		 WHERE stale.is_active = 1
		 	AND ((stale.refresh_expires_at > 0 AND stale.refresh_expires_at > ?) OR (stale.refresh_expires_at = 0 AND stale.expires_at > ?))`,
		now, now, now, now,
	); err != nil {
		return err
	}

	err = tx.Commit()
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
