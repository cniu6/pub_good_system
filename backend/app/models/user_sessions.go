package models

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/pkg/db"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OnlineHeartbeatGraceSeconds 心跳超过此秒数未上报即视为离线（默认值，实际以管理端「上报周期」设置为准，
// 见 services.GetGlobalOnlinePresenceRuntimeConfig；调用方应传入动态 graceSeconds，此常量仅作兜底）。
const OnlineHeartbeatGraceSeconds int64 = 90

// ImpersonationDeviceLabel 管理员代登录(login-as)会话在 user_sessions.device 中的固定标记。
const ImpersonationDeviceLabel = "Admin Impersonation"

// UserSession 用户会话模型
type UserSession struct {
	ID               uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID           uint64 `gorm:"column:user_id;index:idx_us_user_id" json:"user_id"`
	AuthGuard        string `gorm:"column:auth_guard;size:16" json:"auth_guard"`
	TokenHash        string `gorm:"column:token_hash;size:128" json:"-"`
	RefreshTokenHash string `gorm:"column:refresh_token_hash;size:128" json:"-"`
	IP               string `gorm:"column:ip;size:64" json:"ip"`
	UserAgent        string `gorm:"column:user_agent;type:text" json:"user_agent"`
	Device           string `gorm:"column:device;size:128" json:"device"`
	ClientType       string `gorm:"column:client_type;size:16" json:"client_type"`
	BrowserID        string `gorm:"column:browser_id;size:64" json:"browser_id"`
	// BrowserSessionKey 仅 browser_id 非空的 Web 会话写入。NULL 保留给旧会话、原生客户端与代登录，
	// 让三种数据库都允许多个 NULL，同时用唯一索引确保同用户/guard/浏览器最终只保留一条活跃会话。
	BrowserSessionKey *string `gorm:"column:browser_session_key;size:128;uniqueIndex:idx_us_browser_session_key" json:"-"`
	IsActive          bool    `gorm:"column:is_active" json:"is_active"`
	LoginAt           int64   `gorm:"column:login_at" json:"login_at"`
	LastSeenAt        int64   `gorm:"column:last_seen_at" json:"last_seen_at"`
	ExpiresAt         int64   `gorm:"column:expires_at" json:"expires_at"`
	RefreshExpiresAt  int64   `gorm:"column:refresh_expires_at" json:"-"`
	CreatedAt         int64   `gorm:"column:created_at" json:"created_at"`
	// 以下字段仅用于 API 返回，不能持久化。
	IsOnline  bool   `gorm:"-" json:"is_online"`
	IsCurrent bool   `gorm:"-" json:"is_current"`
	Username  string `gorm:"-" json:"username,omitempty"`
	Nickname  string `gorm:"-" json:"nickname,omitempty"`
}

func (UserSession) TableName() string {
	return "user_sessions"
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

// TouchSessionLastSeen 更新心跳时间。返回值表示该会话此刻仍是活跃且未过期的会话。
// Presence 用它在心跳时复核撤销/过期状态，避免 WS 建连后永久信任首次鉴权结果。
func TouchSessionLastSeen(sessionID uint64) (bool, error) {
	result := db.DB.Model(&UserSession{}).
		Where("id = ? AND is_active = ? AND expires_at > ?", sessionID, true, time.Now().Unix()).
		Update("last_seen_at", time.Now().Unix())
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// IsUserSessionActiveByID 判断会话是否仍可使用。用于向已建连的 Presence WS 推送业务消息前复核；
// 查询失败时由调用方按自身可用性策略处理，避免模型层直接依赖 Presence。
func IsUserSessionActiveByID(sessionID uint64) (bool, error) {
	var count int64
	err := db.DB.Model(&UserSession{}).
		Where("id = ? AND is_active = ? AND expires_at > ?", sessionID, true, time.Now().Unix()).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BindSessionBrowserID 给尚无 browser_id 的旧会话补上浏览器实例 ID（兼容升级前已存在的会话）。
func BindSessionBrowserID(sessionID uint64, browserID string) error {
	browserID = clampBrowserID(browserID)
	if browserID == "" {
		return nil
	}
	return db.DB.Model(&UserSession{}).
		Where("id = ? AND is_active = ? AND (browser_id = '' OR browser_id IS NULL)", sessionID, true).
		Update("browser_id", browserID).Error
}

func makeBrowserSessionKey(userID uint64, authGuard, browserID string) *string {
	browserID = clampBrowserID(browserID)
	if browserID == "" {
		return nil
	}
	key := fmt.Sprintf("%d:%s:%s", userID, authGuard, browserID)
	return &key
}

// CreateUserSession 创建用户会话记录。
func CreateUserSession(userID uint64, authGuard, tokenHash, refreshTokenHash, ip, userAgent, device, clientType, browserID string, expiresAt, refreshExpiresAt int64) error {
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	browserID = clampBrowserID(browserID)
	session := UserSession{
		UserID:            userID,
		AuthGuard:         authGuard,
		TokenHash:         tokenHash,
		RefreshTokenHash:  refreshTokenHash,
		IP:                clampStoredIP(ip),
		UserAgent:         userAgent,
		Device:            clampDevice(device),
		ClientType:        NormalizeClientType(clientType),
		BrowserID:         browserID,
		BrowserSessionKey: makeBrowserSessionKey(userID, authGuard, browserID),
		IsActive:          true,
		LoginAt:           now,
		LastSeenAt:        now,
		ExpiresAt:         expiresAt,
		RefreshExpiresAt:  refreshExpiresAt,
		CreatedAt:         now,
	}
	// 创建和收口必须是同一事务：创建失败时不能把旧会话先撤销；
	// 唯一 browser_session_key 则作为并发登录时的最终约束（NULL 会话不参与该约束）。
	for attempt := 0; attempt < 2; attempt++ {
		err := db.WithTx(func(tx *gorm.DB) error {
			if browserID != "" {
				// MySQL/Postgres 锁住已有同浏览器会话；SQLite 由写事务串行化。
				var existing []UserSession
				if err := db.ForUpdate(tx).
					Where("user_id = ? AND auth_guard = ? AND browser_id = ? AND is_active = ?", userID, authGuard, browserID, true).
					Find(&existing).Error; err != nil {
					return err
				}
			}
			if browserID != "" {
				// 先撤销旧会话并释放其唯一键；事务失败会整体回滚，旧会话不会被误踢。
				if _, err := revokeOtherSessionsByBrowserIDTx(tx, userID, authGuard, browserID, ""); err != nil {
					return err
				}
			}
			if err := tx.Create(&session).Error; err != nil {
				return err
			}
			return nil
		})
		if err == nil || !db.IsDuplicateKeyError(err) {
			return err
		}
		// 并发登录时另一事务可能刚插入同一唯一键；重新开启事务后撤销其会话再创建本次会话。
	}
	return fmt.Errorf("创建浏览器会话冲突，请重试")
}

// revokeActiveSessionsWhere 按条件撤销活跃会话，返回被撤销的 ID 列表。
func revokeActiveSessionsWhere(where string, args ...interface{}) ([]uint64, error) {
	return revokeActiveSessionsWhereTx(db.DB, where, args...)
}

func revokeActiveSessionsWhereTx(tx *gorm.DB, where string, args ...interface{}) ([]uint64, error) {
	var ids []uint64
	if err := tx.Model(&UserSession{}).Where(where, args...).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	if err := tx.Model(&UserSession{}).Where(where, args...).Updates(map[string]interface{}{
		"is_active":           false,
		"browser_session_key": nil,
	}).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// RevokeOtherSessionsByBrowserID 撤销同用户/同 guard/同浏览器下、除 keepSessionID 以外的活跃会话。
func RevokeOtherSessionsByBrowserID(userID uint64, authGuard, browserID, keepSessionID string) ([]uint64, error) {
	return revokeOtherSessionsByBrowserIDTx(db.DB, userID, authGuard, browserID, keepSessionID)
}

func revokeOtherSessionsByBrowserIDTx(tx *gorm.DB, userID uint64, authGuard, browserID, keepSessionID string) ([]uint64, error) {
	browserID = clampBrowserID(browserID)
	if browserID == "" {
		return nil, nil
	}
	if authGuard == "" {
		authGuard = "user"
	}
	where := `user_id = ? AND auth_guard = ? AND browser_id = ? AND is_active = ?`
	args := []interface{}{userID, authGuard, browserID, true}
	if keepSessionID = strings.TrimSpace(keepSessionID); keepSessionID != "" {
		where += ` AND id != ?`
		args = append(args, keepSessionID)
	}
	return revokeActiveSessionsWhereTx(tx, where, args...)
}

// RevokeSiblingWebSessionsByUA 兼容升级前无 browser_id 的旧会话。
func RevokeSiblingWebSessionsByUA(userID uint64, authGuard, userAgent, keepSessionID string) ([]uint64, error) {
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return nil, nil
	}
	if authGuard == "" {
		authGuard = "user"
	}
	// 只收口升级前没有 browser_id 的遗留 Web 会话；同 UA 但已有独立 browser_id 的其它设备必须保留。
	where := `user_id = ? AND auth_guard = ? AND client_type = 'web' AND (browser_id = '' OR browser_id IS NULL) AND user_agent = ? AND is_active = ?`
	args := []interface{}{userID, authGuard, userAgent, true}
	if keepSessionID = strings.TrimSpace(keepSessionID); keepSessionID != "" {
		where += ` AND id != ?`
		args = append(args, keepSessionID)
	}
	return revokeActiveSessionsWhere(where, args...)
}

func IsUserSessionActive(userID uint64, authGuard, tokenHash string) (bool, error) {
	var count int64
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ? AND token_hash = ? AND is_active = ? AND expires_at > ?",
			userID, authGuard, tokenHash, true, now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func IsRefreshSessionActive(userID uint64, authGuard, refreshTokenHash string) (bool, error) {
	var count int64
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ? AND refresh_token_hash = ? AND is_active = ? AND refresh_expires_at > ?",
			userID, authGuard, refreshTokenHash, true, now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsImpersonationSession 判断当前活跃 refresh 会话是否为管理员代登录(login-as)会话。
func IsImpersonationSession(userID uint64, authGuard, refreshTokenHash string) (bool, error) {
	var count int64
	now := time.Now().Unix()
	if authGuard == "" {
		authGuard = "user"
	}
	err := db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ? AND refresh_token_hash = ? AND is_active = ? AND refresh_expires_at > ? AND device = ?",
			userID, authGuard, refreshTokenHash, true, now, ImpersonationDeviceLabel).
		Count(&count).Error
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
	result := db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ? AND refresh_token_hash = ? AND is_active = ? AND refresh_expires_at > ?",
			userID, authGuard, currentRefreshTokenHash, true, now).
		Updates(map[string]interface{}{
			"token_hash":         newTokenHash,
			"refresh_token_hash": newRefreshTokenHash,
			"ip":                 clampStoredIP(ip),
			"user_agent":         userAgent,
			"device":             clampDevice(device),
			"expires_at":         expiresAt,
			"refresh_expires_at": refreshExpiresAt,
			"login_at":           now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}
	// 同秒内 JWT 仅含秒级 exp、无 jti 时，新旧 token 可能完全一致，MySQL 报 RowsAffected=0。
	// 若库中活跃会话的 refresh_token_hash 已是目标值（含「本来就相同」的空操作），视为轮换成功。
	var storedHash string
	err := db.DB.Model(&UserSession{}).
		Select("refresh_token_hash").
		Where("user_id = ? AND auth_guard = ? AND is_active = ? AND refresh_expires_at > ?",
			userID, authGuard, true, now).
		Where("refresh_token_hash = ?", newRefreshTokenHash).
		Limit(1).Scan(&storedHash).Error
	if err != nil {
		return false, err
	}
	return storedHash == newRefreshTokenHash, nil
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
	err := db.DB.Model(&UserSession{}).
		Select("id", "user_id", "auth_guard", "token_hash", "ip", "user_agent", "device", "client_type", "is_active", "login_at", "last_seen_at", "expires_at", "created_at").
		Where("user_id = ? AND auth_guard = ? AND is_active = ? AND ((refresh_expires_at > 0 AND refresh_expires_at > ?) OR (refresh_expires_at = 0 AND expires_at > ?))",
			userID, authGuard, true, now, now).
		Order("login_at DESC").
		Limit(50).
		Find(&sessions).Error
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
	return db.DB.Model(&UserSession{}).
		Where("id = ? AND user_id = ? AND auth_guard = ?", sessionID, userID, authGuard).
		Updates(map[string]interface{}{"is_active": false, "browser_session_key": nil}).Error
}

// RevokeSessionByTokenHash 按 token_hash 直接撤销会话，不检查是否仍在有效期内。
func RevokeSessionByTokenHash(userID uint64, authGuard, tokenHash string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	if tokenHash == "" {
		return nil
	}
	return db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ? AND token_hash = ?", userID, authGuard, tokenHash).
		Updates(map[string]interface{}{"is_active": false, "browser_session_key": nil}).Error
}

// RevokeAllUserSessions 撤销用户所有会话（除当前）
func RevokeAllUserSessions(userID uint64, currentTokenHash string) error {
	return RevokeAllUserSessionsWithGuard(userID, "user", currentTokenHash)
}

func RevokeAllUserSessionsWithGuard(userID uint64, authGuard, currentTokenHash string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	q := db.DB.Model(&UserSession{}).Where("user_id = ? AND auth_guard = ?", userID, authGuard)
	if currentTokenHash != "" {
		q = q.Where("token_hash != ?", currentTokenHash)
	}
	return q.Updates(map[string]interface{}{"is_active": false, "browser_session_key": nil}).Error
}

// CleanupExpiredSessions 清理过期会话，返回删除行数
func CleanupExpiredSessions() (int64, error) {
	now := time.Now().Unix()
	var aff int64
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where(
			"is_active = ? OR (refresh_expires_at > 0 AND refresh_expires_at <= ?) OR (refresh_expires_at = 0 AND expires_at <= ?)",
			false, now, now,
		).Delete(&UserSession{})
		if result.Error != nil {
			return result.Error
		}
		aff = result.RowsAffected
		return nil
	})
	return aff, err
}

// GetUserLoginCount 获取用户登录次数
func GetUserLoginCount(userID uint64) (int64, error) {
	var count int64
	err := db.DB.Model(&UserSession{}).Where("user_id = ?", userID).Count(&count).Error
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
// 无效/过期 token 查无是常态，用 FindOne 避免 record not found 刷屏。
func GetActiveSessionByTokenHash(tokenHash string) (*UserSession, error) {
	var session UserSession
	now := time.Now().Unix()
	err := db.FindOne(db.DB.Model(&UserSession{}).
		Select("id", "user_id", "auth_guard", "token_hash", "ip", "user_agent", "device", "client_type", "browser_id", "is_active", "login_at", "last_seen_at", "expires_at", "refresh_expires_at", "created_at").
		Where("token_hash = ? AND is_active = ? AND expires_at > ?", tokenHash, true, now), &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserSessionByID 按 ID 获取单个会话（含已离线但仍有效的会话）。
func GetUserSessionByID(sessionID uint64) (*UserSession, error) {
	var session UserSession
	err := db.DB.Model(&UserSession{}).
		Select("id", "user_id", "auth_guard", "token_hash", "ip", "user_agent", "device", "client_type", "browser_id", "is_active", "login_at", "last_seen_at", "expires_at", "refresh_expires_at", "created_at").
		Where("id = ?", sessionID).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// AdminRevokeSessionByID 管理端按会话 ID 强制撤销。
func AdminRevokeSessionByID(sessionID uint64) error {
	return db.DB.Model(&UserSession{}).Where("id = ?", sessionID).
		Updates(map[string]interface{}{"is_active": false, "browser_session_key": nil}).Error
}

// AdminRevokeAllUserSessions 管理端撤销某用户指定认证上下文的全部会话。
func AdminRevokeAllUserSessions(userID uint64, authGuard string) error {
	if authGuard == "" {
		authGuard = "user"
	}
	return db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ?", userID, authGuard).
		Updates(map[string]interface{}{"is_active": false, "browser_session_key": nil}).Error
}

// ListActiveSessionIDsExceptCurrent 返回被“踢出其它设备”操作影响的会话 ID。
func ListActiveSessionIDsExceptCurrent(userID uint64, authGuard, currentTokenHash string) ([]uint64, error) {
	if authGuard == "" {
		authGuard = "user"
	}
	q := db.DB.Model(&UserSession{}).
		Where("user_id = ? AND auth_guard = ? AND is_active = ?", userID, authGuard, true)
	if currentTokenHash != "" {
		q = q.Where("token_hash != ?", currentTokenHash)
	}
	var ids []uint64
	if err := q.Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// GetOnlineHeartbeatGraceSeconds 返回当前生效的在线判定容忍窗口。
var GetOnlineHeartbeatGraceSeconds = func() int64 { return OnlineHeartbeatGraceSeconds }

// GetLastSeenAtByUserIDs 批量查询用户最近一次会话心跳时间。
func GetLastSeenAtByUserIDs(userIDs []uint64) (map[uint64]int64, error) {
	result := make(map[uint64]int64, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		UserID     uint64 `gorm:"column:user_id"`
		LastSeenAt int64  `gorm:"column:last_seen_at"`
	}
	if err := db.DB.Model(&UserSession{}).
		Select("user_id, MAX(last_seen_at) AS last_seen_at").
		Where("user_id IN ? AND is_active = ?", userIDs, true).
		Group("user_id").
		Scan(&rows).Error; err != nil {
		return result, err
	}
	for _, r := range rows {
		result[r.UserID] = r.LastSeenAt
	}
	return result, nil
}

// CountOnlineSessions 统计心跳窗口内在线会话数。
func CountOnlineSessions() (int64, error) {
	var count int64
	err := db.DB.Model(&UserSession{}).
		Where("is_active = ? AND last_seen_at >= ?", true, time.Now().Unix()-GetOnlineHeartbeatGraceSeconds()).
		Count(&count).Error
	return count, err
}

// CountOnlineUsers 统计至少有一个在线会话的用户数。
func CountOnlineUsers() (int64, error) {
	var count int64
	err := db.DB.Model(&UserSession{}).
		Where("is_active = ? AND last_seen_at >= ?", true, time.Now().Unix()-GetOnlineHeartbeatGraceSeconds()).
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// OnlineUserRow 管理端在线列表：同一用户+登录端归并为一行，多设备挂在 Devices。
type OnlineUserRow struct {
	UserID      uint64        `json:"user_id"`
	Username    string        `json:"username,omitempty"`
	Nickname    string        `json:"nickname,omitempty"`
	AuthGuard   string        `json:"auth_guard"`
	DeviceCount int           `json:"device_count"`
	LastSeenAt  int64         `json:"last_seen_at"`
	IsOnline    bool          `json:"is_online"`
	Devices     []UserSession `json:"devices"`
}

// onlineSessionListSelect 在线列表 JOIN users 的公共 SELECT 片段。
const onlineSessionListSelect = `s.id, s.user_id, s.auth_guard, s.ip, s.user_agent, s.device, s.client_type, s.browser_id, s.is_active, s.login_at, s.last_seen_at, s.expires_at, s.created_at,
		COALESCE(u.username, '') AS username, COALESCE(u.nickname, '') AS nickname`

// buildOnlineSessionsWhere 构造在线会话筛选条件（keyword/clientType/authGuard）。
func buildOnlineSessionsWhere(keyword, clientType, authGuard string) (string, []interface{}) {
	where := ` WHERE s.is_active = ? AND s.last_seen_at >= ?`
	args := []interface{}{true, time.Now().Unix() - GetOnlineHeartbeatGraceSeconds()}
	if clientType = strings.TrimSpace(clientType); clientType != "" {
		where += ` AND s.client_type = ?`
		args = append(args, NormalizeClientType(clientType))
	}
	if authGuard = strings.TrimSpace(authGuard); authGuard != "" {
		where += ` AND s.auth_guard = ?`
		args = append(args, authGuard)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		where += ` AND (` + db.CastToText("s.user_id") + ` LIKE ? OR s.ip LIKE ? OR s.device LIKE ? OR COALESCE(s.user_agent, '') LIKE ? OR COALESCE(u.username, '') LIKE ? OR COALESCE(u.nickname, '') LIKE ?)`
		like := "%" + keyword + "%"
		args = append(args, like, like, like, like, like, like)
	}
	return where, args
}

// ListOnlineSessions 分页查询在线会话（原始会话行，不做用户归并）。
func ListOnlineSessions(keyword, clientType, authGuard string, page, pageSize int) ([]UserSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where, args := buildOnlineSessionsWhere(keyword, clientType, authGuard)
	var total int64
	if err := db.DB.Raw(`SELECT COUNT(*) FROM user_sessions s INNER JOIN users u ON u.id = s.user_id AND u.delete_time IS NULL`+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	sessions := make([]UserSession, 0)
	listArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	err := db.DB.Raw(`SELECT `+onlineSessionListSelect+`
		FROM user_sessions s
		INNER JOIN users u ON u.id = s.user_id AND u.delete_time IS NULL`+where+
		` ORDER BY s.last_seen_at DESC LIMIT ? OFFSET ?`, listArgs...).Scan(&sessions).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range sessions {
		sessions[i].IsOnline = true
	}
	return sessions, total, nil
}

// ListOnlineUsersGrouped 按 user_id + auth_guard 归并在线用户，一行多设备。
func ListOnlineUsersGrouped(keyword, clientType, authGuard string, page, pageSize int) ([]OnlineUserRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where, args := buildOnlineSessionsWhere(keyword, clientType, authGuard)

	const maxGroupedSessionRows = 5000
	sessions := make([]UserSession, 0)
	err := db.DB.Raw(`SELECT `+onlineSessionListSelect+`
		FROM user_sessions s
		INNER JOIN users u ON u.id = s.user_id AND u.delete_time IS NULL`+where+
		` ORDER BY s.last_seen_at DESC LIMIT ?`, append(append([]interface{}{}, args...), maxGroupedSessionRows)...).Scan(&sessions).Error
	if err != nil {
		return nil, 0, err
	}

	type groupKey struct {
		UserID    uint64
		AuthGuard string
	}
	order := make([]groupKey, 0)
	grouped := make(map[groupKey]*OnlineUserRow)
	for i := range sessions {
		s := sessions[i]
		s.IsOnline = true
		key := groupKey{UserID: s.UserID, AuthGuard: s.AuthGuard}
		row, ok := grouped[key]
		if !ok {
			row = &OnlineUserRow{
				UserID:    s.UserID,
				Username:  s.Username,
				Nickname:  s.Nickname,
				AuthGuard: s.AuthGuard,
				IsOnline:  true,
				Devices:   make([]UserSession, 0, 2),
			}
			grouped[key] = row
			order = append(order, key)
		}
		row.Devices = append(row.Devices, s)
		if s.LastSeenAt > row.LastSeenAt {
			row.LastSeenAt = s.LastSeenAt
		}
		row.DeviceCount = len(row.Devices)
	}

	total := int64(len(order))
	start := (page - 1) * pageSize
	if start >= len(order) {
		return []OnlineUserRow{}, total, nil
	}
	end := start + pageSize
	if end > len(order) {
		end = len(order)
	}
	result := make([]OnlineUserRow, 0, end-start)
	for _, key := range order[start:end] {
		result = append(result, *grouped[key])
	}
	return result, total, nil
}
