package models

import (
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"time"
)

// OnlineHeartbeatGraceSeconds 心跳超过此秒数未上报即视为离线（默认值，实际以管理端「上报周期」设置为准，
// 见 services.GetGlobalOnlinePresenceRuntimeConfig；调用方应传入动态 graceSeconds，此常量仅作兜底）。
const OnlineHeartbeatGraceSeconds int64 = 90

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
	ClientType       string `db:"client_type" json:"client_type"`
	BrowserID        string `db:"browser_id" json:"browser_id"`
	IsActive         bool   `db:"is_active" json:"is_active"`
	LoginAt          int64  `db:"login_at" json:"login_at"`
	LastSeenAt       int64  `db:"last_seen_at" json:"last_seen_at"`
	ExpiresAt        int64  `db:"expires_at" json:"expires_at"`
	RefreshExpiresAt int64  `db:"refresh_expires_at" json:"-"`
	CreatedAt        int64  `db:"created_at" json:"created_at"`
	// 以下字段仅用于 API 返回，不能持久化。
	IsOnline  bool   `db:"-" json:"is_online"`
	IsCurrent bool   `db:"-" json:"is_current"`
	Username  string `db:"username" json:"username,omitempty"`
	Nickname  string `db:"nickname" json:"nickname,omitempty"`
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
			client_type VARCHAR(20) NOT NULL DEFAULT 'web' COMMENT '客户端类型 web/app',
			browser_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '浏览器实例ID（同浏览器多标签共用）',
			is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否活跃',
			login_at BIGINT NOT NULL DEFAULT 0 COMMENT '登录时间',
			last_seen_at BIGINT NOT NULL DEFAULT 0 COMMENT '最后在线心跳时间',
			expires_at BIGINT NOT NULL DEFAULT 0 COMMENT 'Access Token过期时间',
			refresh_expires_at BIGINT NOT NULL DEFAULT 0 COMMENT 'Refresh Token过期时间',
			created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			INDEX idx_user_id (user_id),
			INDEX idx_user_guard (user_id, auth_guard),
			INDEX idx_is_active (is_active),
			INDEX idx_user_token_active_expire (user_id, auth_guard, token_hash, is_active, expires_at),
			INDEX idx_user_refresh_active_expire (user_id, auth_guard, refresh_token_hash, is_active, refresh_expires_at),
			INDEX idx_user_active_login (user_id, auth_guard, is_active, login_at),
			INDEX idx_active_last_seen (is_active, last_seen_at),
			INDEX idx_client_last_seen (client_type, last_seen_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

		_, err := db.Exec(schema)
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
		"client_type":        "ALTER TABLE user_sessions ADD COLUMN client_type VARCHAR(20) NOT NULL DEFAULT 'web' COMMENT '客户端类型 web/app' AFTER device",
		"browser_id":         "ALTER TABLE user_sessions ADD COLUMN browser_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '浏览器实例ID（同浏览器多标签共用）' AFTER client_type",
		"last_seen_at":       "ALTER TABLE user_sessions ADD COLUMN last_seen_at BIGINT NOT NULL DEFAULT 0 COMMENT '最后在线心跳时间' AFTER login_at",
	}

	for column, alterSQL := range repairs {
		if !db.CheckColumnExists("user_sessions", column) {
			if _, err := db.Exec(alterSQL); err != nil {
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
		"idx_active_last_seen":           "ALTER TABLE user_sessions ADD INDEX idx_active_last_seen (is_active, last_seen_at)",
		"idx_client_last_seen":           "ALTER TABLE user_sessions ADD INDEX idx_client_last_seen (client_type, last_seen_at)",
		"idx_user_browser_active":        "ALTER TABLE user_sessions ADD INDEX idx_user_browser_active (user_id, auth_guard, browser_id, is_active)",
	}

	for indexName, alterSQL := range indexRepairs {
		db.EnsureIndex("user_sessions", indexName, alterSQL)
	}
}

// NormalizeClientType 将外部客户端类型归一化，未知值安全回退为 web。
func NormalizeClientType(clientType string) string {
	if strings.EqualFold(strings.TrimSpace(clientType), "app") {
		return "app"
	}
	return "web"
}

// SessionIsOnline 根据最后一次 Presence 心跳判断会话在线状态。graceSeconds 由调用方传入（来自动态设置）。
func SessionIsOnline(session UserSession, graceSeconds int64) bool {
	if graceSeconds <= 0 {
		graceSeconds = OnlineHeartbeatGraceSeconds
	}
	return session.IsActive && session.LastSeenAt >= time.Now().Unix()-graceSeconds
}

// TouchSessionLastSeen 更新心跳时间。仅更新尚未过期的活跃会话。
func TouchSessionLastSeen(sessionID uint64) error {
	_, err := db.Exec(`UPDATE user_sessions SET last_seen_at = ? WHERE id = ? AND is_active = 1`, time.Now().Unix(), sessionID)
	return err
}

// BindSessionBrowserID 给尚无 browser_id 的旧会话补上浏览器实例 ID（兼容升级前已存在的会话）。
func BindSessionBrowserID(sessionID uint64, browserID string) error {
	browserID = strings.TrimSpace(browserID)
	if browserID == "" {
		return nil
	}
	_, err := db.Exec(`UPDATE user_sessions SET browser_id = ? WHERE id = ? AND is_active = 1 AND (browser_id = '' OR browser_id IS NULL)`, browserID, sessionID)
	return err
}

// CreateUserSession 创建用户会话记录。
// browserID 非空时，会先撤销同一用户、同一认证上下文、同一浏览器实例下的其它活跃会话，
// 保证「一个浏览器只保留一条在线会话」（多标签重复登录不会堆出多行）。
func CreateUserSession(userID uint64, authGuard, tokenHash, refreshTokenHash, ip, userAgent, device, clientType, browserID string, expiresAt, refreshExpiresAt int64) error {
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	browserID = strings.TrimSpace(browserID)
	if browserID != "" {
		if _, err := RevokeOtherSessionsByBrowserID(userID, authGuard, browserID, ""); err != nil {
			return err
		}
	}
	_, err := db.Exec(
		`INSERT INTO user_sessions (user_id, auth_guard, token_hash, refresh_token_hash, ip, user_agent, device, client_type, browser_id, is_active, login_at, last_seen_at, expires_at, refresh_expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?)`,
		userID, authGuard, tokenHash, refreshTokenHash, ip, userAgent, device, NormalizeClientType(clientType), browserID, now, now, expiresAt, refreshExpiresAt, now,
	)
	return err
}

// RevokeOtherSessionsByBrowserID 撤销同用户/同 guard/同浏览器下、除 keepSessionID 以外的活跃会话。
// 返回被撤销的会话 ID 列表，便于 Presence Hub 踢掉旧连接。
func RevokeOtherSessionsByBrowserID(userID uint64, authGuard, browserID, keepSessionID string) ([]uint64, error) {
	browserID = strings.TrimSpace(browserID)
	if browserID == "" {
		return nil, nil
	}
	if authGuard == "" {
		authGuard = "user"
	}
	args := []interface{}{userID, authGuard, browserID}
	where := `user_id = ? AND auth_guard = ? AND browser_id = ? AND is_active = 1`
	if keepSessionID = strings.TrimSpace(keepSessionID); keepSessionID != "" {
		where += ` AND id != ?`
		args = append(args, keepSessionID)
	}
	ids := make([]uint64, 0)
	if err := db.DB.Select(&ids, `SELECT id FROM user_sessions WHERE `+where, args...); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	if _, err := db.Exec(`UPDATE user_sessions SET is_active = 0 WHERE `+where, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

// RevokeSiblingWebSessionsByUA 兼容升级前无 browser_id 的旧会话：
// 同一用户、同一 guard、同一 User-Agent 的其它 web 活跃会话一并撤销（用于多标签重复登录收口）。
func RevokeSiblingWebSessionsByUA(userID uint64, authGuard, userAgent, keepSessionID string) ([]uint64, error) {
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return nil, nil
	}
	if authGuard == "" {
		authGuard = "user"
	}
	args := []interface{}{userID, authGuard, userAgent}
	where := `user_id = ? AND auth_guard = ? AND client_type = 'web' AND user_agent = ? AND is_active = 1`
	if keepSessionID = strings.TrimSpace(keepSessionID); keepSessionID != "" {
		where += ` AND id != ?`
		args = append(args, keepSessionID)
	}
	ids := make([]uint64, 0)
	if err := db.DB.Select(&ids, `SELECT id FROM user_sessions WHERE `+where, args...); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	if _, err := db.Exec(`UPDATE user_sessions SET is_active = 0 WHERE `+where, args...); err != nil {
		return nil, err
	}
	return ids, nil
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
	result, err := db.Exec(
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
	return GetUserSessionsWithGuard(userID, "user", "")
}

// GetUserSessionsWithGuard 返回用户未过期的活动会话，不会向调用方泄露 token 哈希。
func GetUserSessionsWithGuard(userID uint64, authGuard, currentTokenHash string) ([]UserSession, error) {
	var sessions []UserSession
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Select(&sessions,
		`SELECT id, user_id, auth_guard, token_hash, ip, user_agent, device, client_type, is_active, login_at, last_seen_at, expires_at, created_at
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
	graceSeconds := GetOnlineHeartbeatGraceSeconds()
	for i := range sessions {
		sessions[i].IsOnline = SessionIsOnline(sessions[i], graceSeconds)
		sessions[i].IsCurrent = currentTokenHash != "" && sessions[i].TokenHash == currentTokenHash
		sessions[i].TokenHash = ""
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
	_, err := db.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE id = ? AND user_id = ? AND auth_guard = ?",
		sessionID, userID, authGuard,
	)
	return err
}

// RevokeSessionByTokenHash 按 token_hash 直接撤销会话，不检查是否仍在有效期内。
// 用于 access token 已过期时的"尽力而为"退出清理：此时 IsUserSessionActive/GetActiveSessionByTokenHash
// 的 expires_at 过滤条件已经命中不到这条记录了，只能靠 token_hash + user_id + auth_guard 精确匹配定位。
func RevokeSessionByTokenHash(userID uint64, authGuard, tokenHash string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	if tokenHash == "" {
		return nil
	}
	_, err := db.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND auth_guard = ? AND token_hash = ?",
		userID, authGuard, tokenHash,
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
		_, err := db.Exec(
			"UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND auth_guard = ? AND token_hash != ?",
			userID, authGuard, currentTokenHash,
		)
		return err
	}
	_, err := db.Exec(
		"UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND auth_guard = ?",
		userID, authGuard,
	)
	return err
}

// CleanupExpiredSessions 清理过期会话，返回删除行数
func CleanupExpiredSessions() (int64, error) {
	now := time.Now().Unix()
	tx, err := db.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.Exec(
		"DELETE FROM user_sessions WHERE is_active = 0 OR (refresh_expires_at > 0 AND refresh_expires_at <= ?) OR (refresh_expires_at = 0 AND expires_at <= ?)",
		now, now,
	)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return aff, nil
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

// ParseSessionID 解析并校验会话 ID。
func ParseSessionID(raw string) (uint64, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || id == 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

// GetActiveSessionByTokenHash 以 access token 哈希获取未过期的活动会话。
func GetActiveSessionByTokenHash(tokenHash string) (*UserSession, error) {
	var session UserSession
	now := time.Now().Unix()
	err := db.DB.Get(&session, `SELECT id, user_id, auth_guard, token_hash, ip, user_agent, device, client_type, browser_id, is_active, login_at, last_seen_at, expires_at, refresh_expires_at, created_at
		FROM user_sessions WHERE token_hash = ? AND is_active = 1 AND expires_at > ? LIMIT 1`, tokenHash, now)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserSessionByID 按 ID 获取单个会话（含已离线但仍有效的会话）。
func GetUserSessionByID(sessionID uint64) (*UserSession, error) {
	var session UserSession
	err := db.DB.Get(&session, `SELECT id, user_id, auth_guard, token_hash, ip, user_agent, device, client_type, browser_id, is_active, login_at, last_seen_at, expires_at, refresh_expires_at, created_at
		FROM user_sessions WHERE id = ?`, sessionID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// AdminRevokeSessionByID 管理端按会话 ID 强制撤销。
func AdminRevokeSessionByID(sessionID uint64) error {
	_, err := db.Exec(`UPDATE user_sessions SET is_active = 0 WHERE id = ?`, sessionID)
	return err
}

// AdminRevokeAllUserSessions 管理端撤销某用户指定认证上下文的全部会话。
func AdminRevokeAllUserSessions(userID uint64, authGuard string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	_, err := db.Exec(`UPDATE user_sessions SET is_active = 0 WHERE user_id = ? AND auth_guard = ?`, userID, authGuard)
	return err
}

// ListActiveSessionIDsExceptCurrent 返回被“踢出其它设备”操作影响的会话 ID。
func ListActiveSessionIDsExceptCurrent(userID uint64, authGuard, currentTokenHash string) ([]uint64, error) {
	if authGuard == "" {
		authGuard = "user"
	}
	ids := make([]uint64, 0)
	query := `SELECT id FROM user_sessions WHERE user_id = ? AND auth_guard = ? AND is_active = 1`
	args := []interface{}{userID, authGuard}
	if currentTokenHash != "" {
		query += ` AND token_hash != ?`
		args = append(args, currentTokenHash)
	}
	if err := db.DB.Select(&ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

// GetOnlineHeartbeatGraceSeconds 返回当前生效的在线判定容忍窗口（由 services 层的「上报周期」设置换算而来）。
// 用 var 函数指针而非直接依赖 services 包，避免 models → services 的循环引用；
// 由 services 包在包初始化时注入真实实现，此处仅提供兜底默认值。
var GetOnlineHeartbeatGraceSeconds = func() int64 { return OnlineHeartbeatGraceSeconds }

// CountOnlineSessions 统计心跳窗口内在线会话数。
func CountOnlineSessions() (int64, error) {
	var count int64
	err := db.DB.Get(&count, `SELECT COUNT(*) FROM user_sessions WHERE is_active = 1 AND last_seen_at >= ?`, time.Now().Unix()-GetOnlineHeartbeatGraceSeconds())
	return count, err
}

// CountOnlineUsers 统计至少有一个在线会话的用户数。
func CountOnlineUsers() (int64, error) {
	var count int64
	err := db.DB.Get(&count, `SELECT COUNT(DISTINCT user_id) FROM user_sessions WHERE is_active = 1 AND last_seen_at >= ?`, time.Now().Unix()-GetOnlineHeartbeatGraceSeconds())
	return count, err
}

// ListOnlineSessions 分页查询在线会话。关键字匹配用户名/昵称/用户 ID、IP、设备或 UA。
func ListOnlineSessions(keyword, clientType, authGuard string, page, pageSize int) ([]UserSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where := ` WHERE s.is_active = 1 AND s.last_seen_at >= ?`
	args := []interface{}{time.Now().Unix() - GetOnlineHeartbeatGraceSeconds()}
	if clientType = strings.TrimSpace(clientType); clientType != "" {
		where += ` AND s.client_type = ?`
		args = append(args, NormalizeClientType(clientType))
	}
	if authGuard = strings.TrimSpace(authGuard); authGuard != "" {
		where += ` AND s.auth_guard = ?`
		args = append(args, authGuard)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		where += ` AND (CAST(s.user_id AS CHAR) LIKE ? OR s.ip LIKE ? OR s.device LIKE ? OR COALESCE(s.user_agent, '') LIKE ? OR COALESCE(u.username, '') LIKE ? OR COALESCE(u.nickname, '') LIKE ?)`
		like := "%" + keyword + "%"
		args = append(args, like, like, like, like, like, like)
	}
	var total int64
	if err := db.DB.Get(&total, `SELECT COUNT(*) FROM user_sessions s LEFT JOIN users u ON u.id = s.user_id`+where, args...); err != nil {
		return nil, 0, err
	}
	sessions := make([]UserSession, 0)
	listArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	err := db.DB.Select(&sessions, `SELECT s.id, s.user_id, s.auth_guard, s.ip, s.user_agent, s.device, s.client_type, s.browser_id, s.is_active, s.login_at, s.last_seen_at, s.expires_at, s.created_at,
		COALESCE(u.username, '') AS username, COALESCE(u.nickname, '') AS nickname
		FROM user_sessions s
		LEFT JOIN users u ON u.id = s.user_id`+where+
		` ORDER BY s.last_seen_at DESC LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	for i := range sessions {
		sessions[i].IsOnline = true
	}
	return sessions, total, nil
}
