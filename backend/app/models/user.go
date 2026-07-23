package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"regexp"
	"strings"
	"time"
)

// 与 utils/phone.go 对齐：用于手机号等价写法查询（避免 import cycle）
var reMobileCNLookup = regexp.MustCompile(`^1[3-9]\d{9}$`)

// mobileLookupVariants 同一号码可能存成 11 位或 +86 E.164
func mobileLookupVariants(normalized string) []string {
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return nil
	}
	out := []string{normalized}
	if strings.HasPrefix(normalized, "+86") && len(normalized) == 14 && reMobileCNLookup.MatchString(normalized[3:]) {
		out = append(out, normalized[3:])
	} else if reMobileCNLookup.MatchString(normalized) {
		out = append(out, "+86"+normalized)
	}
	return out
}

type User struct {
	ID            uint64  `db:"id" json:"id"`
	GroupId       uint64  `db:"group_id" json:"group_id"`
	Username      string  `db:"username" json:"username"`
	Nickname      string  `db:"nickname" json:"nickname"`
	Email         string  `db:"email" json:"email"`
	Mobile        string  `db:"mobile" json:"mobile"`
	Avatar        string  `db:"avatar" json:"avatar"`
	BackGround    string  `db:"back_ground" json:"back_ground"`
	Gender        uint8   `db:"gender" json:"gender"`
	Birthday      *int64  `db:"birthday" json:"birthday"`
	Money         float64 `db:"money" json:"money"` // 余额（元，DECIMAL；业务加减一律经 utils 按「分」整数计算）
	Score         int64   `db:"score" json:"score"`
	Level         uint64  `db:"level" json:"level"`
	Role          string  `db:"role" json:"role"` // 'user' or 'admin'
	LastLoginTime *int64  `db:"last_login_time" json:"last_login_time"`
	LastLoginIp   string  `db:"last_login_ip" json:"last_login_ip"`
	LoginFailure  uint8   `db:"login_failure" json:"login_failure"`
	LockUntil     *int64  `db:"lock_until" json:"lock_until"` // 账户锁定到期时间（时间戳）
	JoinIp        string  `db:"join_ip" json:"join_ip"`
	JoinTime      *int64  `db:"join_time" json:"join_time"`
	Motto         string  `db:"motto" json:"motto"`
	AdminRemark   string  `db:"admin_remark" json:"-"`
	Password      string  `db:"password" json:"-"`
	Status        uint8   `db:"status" json:"status"`

	// 兼容旧表里的 is_active 字段（不在业务中使用，仅为避免扫描报错）
	IsActive *uint8 `db:"is_active" json:"-"`

	// 兼容旧表里的 created_at / updated_at / deleted_at 字段（不在业务中使用，仅为避免扫描报错）
	CreatedAtRaw *time.Time `db:"created_at" json:"-"`
	UpdatedAtRaw *time.Time `db:"updated_at" json:"-"`
	DeletedAtRaw *time.Time `db:"deleted_at" json:"-"`

	// Apikey 数据库中存储 SHA256 哈希，绝不是明文；因此禁止直接 JSON 下发（json:"-"）。
	// 展示统一走 MaskedApikey()（仅末4位），明文只在生成/重置的响应中一次性返回。
	Apikey     *string `db:"apikey" json:"-"`
	ApikeyHint *string `db:"apikey_hint" json:"-"` // API Key 明文末4位，仅用于拼出展示用的掩码，不是敏感信息
	// API Key 收紧：过期时间、IP 白名单（逗号分隔）、scope（user,admin,*）
	ApikeyExpiresAt *int64  `db:"apikey_expires_at" json:"-"`
	ApikeyAllowIPs  *string `db:"apikey_allow_ips" json:"-"`
	ApikeyScopes    *string `db:"apikey_scopes" json:"-"`
	UpdateTime      *int64  `db:"update_time" json:"update_time"`
	CreateTime      *int64  `db:"create_time" json:"create_time"`
	DeleteTime      *int64  `db:"delete_time" json:"-"`

	// Requested additions
	Language string `db:"language" json:"language"`
	Country  string `db:"country" json:"country"`
	// Token 为历史兼容字段，可能含敏感值，禁止随 JSON 响应下发
	Token string `db:"token" json:"-"`

	// 管理端 TOTP 二次验证（secret 禁止下发；enabled 可展示）
	TotpSecret  *string `db:"totp_secret" json:"-"`
	TotpEnabled bool    `db:"totp_enabled" json:"totp_enabled"`
}

func (u *User) TableName() string {
	return "users"
}

// MaskedApikey 返回用于展示的掩码 API Key（固定 ******** + 末4位）；
// 未设置时返回空串。真实哈希/明文都不经此方法或任何 JSON 字段下发，
// 明文仅在 ResetUserApiKey 生成时一次性返回给调用方。
func (u *User) MaskedApikey() string {
	if u == nil || u.Apikey == nil || strings.TrimSpace(*u.Apikey) == "" {
		return ""
	}
	if u.ApikeyHint != nil {
		hint := strings.TrimSpace(*u.ApikeyHint)
		if hint != "" {
			return "********" + hint
		}
	}
	return "********"
}

func BuildUserSelectColumns(tableAlias string) string {
	alias := strings.TrimSpace(tableAlias)
	if alias == "" {
		alias = "users"
	}
	qualified := func(column string) string {
		return alias + "." + column
	}
	columns := []string{
		qualified("id"),
		qualified("group_id"),
		qualified("username"),
		qualified("nickname"),
		qualified("email"),
		qualified("mobile"),
		qualified("avatar"),
		qualified("back_ground"),
		qualified("gender"),
		qualified("birthday"),
		qualified("money"),
		qualified("score"),
		qualified("level"),
		qualified("role"),
		qualified("last_login_time"),
		qualified("last_login_ip"),
		qualified("login_failure"),
		qualified("lock_until"),
		qualified("join_ip"),
		qualified("join_time"),
		qualified("motto"),
		"COALESCE(" + qualified("admin_remark") + ", '') AS admin_remark",
		qualified("password"),
		qualified("status"),
		qualified("apikey"),
		qualified("apikey_hint"),
		qualified("apikey_expires_at"),
		qualified("apikey_allow_ips"),
		qualified("apikey_scopes"),
		qualified("update_time"),
		qualified("create_time"),
		qualified("delete_time"),
		qualified("language"),
		qualified("country"),
		qualified("token"),
		qualified("totp_secret"),
		"COALESCE(" + qualified("totp_enabled") + ", 0) AS totp_enabled",
	}
	return strings.Join(columns, ", ")
}

// CreateUser inserts a new user into the database
func CreateUser(user *User) error {
	query := `INSERT INTO users (
		group_id, username, nickname, email, mobile, avatar, back_ground, gender, birthday, 
		money, score, level, role, last_login_time, last_login_ip, login_failure, 
		join_ip, join_time, motto, admin_remark, password, status, apikey, apikey_hint, update_time, create_time, 
		language, country, token
	) VALUES (
		:group_id, :username, :nickname, :email, :mobile, :avatar, :back_ground, :gender, :birthday, 
		:money, :score, :level, :role, :last_login_time, :last_login_ip, :login_failure, 
		:join_ip, :join_time, :motto, :admin_remark, :password, :status, :apikey, :apikey_hint, :update_time, :create_time, 
		:language, :country, :token
	)`

	// 调用方（如注册自动发放）可能在 user.Apikey 里传入明文，这里统一落库前转成 SHA256 哈希 + 末4位提示，
	// 数据库任何时候都不落地明文 API Key。
	if user.Apikey != nil && strings.TrimSpace(*user.Apikey) != "" {
		plain := strings.TrimSpace(*user.Apikey)
		hashed := hashAPIKey(plain)
		hint := apiKeyHint(plain)
		user.Apikey = &hashed
		user.ApikeyHint = &hint
	}

	now := time.Now().Unix()
	user.CreateTime = &now
	user.UpdateTime = &now
	if user.JoinTime == nil {
		user.JoinTime = &now
	}

	// Set default values if not set (though DB has defaults, struct zero values might overwrite)
	// Usually zero values are fine if we want DB defaults, BUT `NamedExec` will insert zero values.
	// We should probably rely on DB defaults or set them here.
	// Since we pass all fields, we should set defaults in Go or omit fields.
	// For simplicity, let's assume the user struct is populated with defaults or zero values are acceptable (e.g. 0, empty string).
	// But note: DB Default '1' for GroupId. Go zero value is 0.
	if user.GroupId == 0 {
		user.GroupId = 1
	}
	if user.Level == 0 {
		user.Level = 1
	}
	// 管理端允许显式创建“禁用”用户，这里不要把 0 强行覆盖回 1。

	// Ensure language default
	if user.Language == "" {
		user.Language = "zh-CN"
	}

	// Postgres 不支持 LastInsertId：直接 INSERT ... RETURNING id（只执行一次）
	if db.IsPostgres() {
		rows, qErr := db.DB.NamedQuery(query+" RETURNING id", user)
		if qErr != nil {
			return qErr
		}
		defer rows.Close()
		if !rows.Next() {
			return fmt.Errorf("postgres insert user: no returning id")
		}
		var id uint64
		if scanErr := rows.Scan(&id); scanErr != nil {
			return scanErr
		}
		user.ID = id
		return nil
	}

	result, err := db.DB.NamedExec(query, user)
	if err != nil {
		return err
	}
	id, err := db.LastInsertID(result)
	if err != nil {
		return err
	}
	user.ID = uint64(id)
	return nil
}

// GetUserByUsername finds a user by username
func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE username = ? AND delete_time IS NULL", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail finds a user by email
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE email = ? AND delete_time IS NULL", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByMobile finds a user by mobile number（兼容 11 位与 +86 E.164 等价写法）
func GetUserByMobile(mobile string) (*User, error) {
	variants := mobileLookupVariants(mobile)
	if len(variants) == 0 {
		return nil, sql.ErrNoRows
	}
	var lastErr error = sql.ErrNoRows
	for _, v := range variants {
		var user User
		err := db.DB.Get(&user, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE mobile = ? AND delete_time IS NULL", v)
		if err == nil {
			return &user, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

// GetUserByUsernameOrEmail finds a user by username or email
func GetUserByUsernameOrEmail(identifier string) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE (username = ? OR email = ?) AND delete_time IS NULL", identifier, identifier)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by ID
func GetUserByID(id uint64) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE id = ? AND delete_time IS NULL", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates the user's password
func UpdatePassword(userID uint64, hashedPassword string) error {
	now := time.Now().Unix()
	_, err := db.Exec("UPDATE users SET password = ?, update_time = ? WHERE id = ?", hashedPassword, now, userID)
	if err != nil {
		return err
	}
	if err := RevokeAllUserSessionsWithGuard(userID, "user", ""); err != nil {
		return err
	}
	return RevokeAllUserSessionsWithGuard(userID, "admin", "")
}

// UpdateLoginInfo 更新用户登录信息（成功登录后调用）
func UpdateLoginInfo(userID uint64, loginIP string) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"UPDATE users SET last_login_time = ?, last_login_ip = ?, login_failure = 0, lock_until = NULL, update_time = ? WHERE id = ?",
		now, loginIP, now, userID,
	)
	return err
}

// GetUserByApiKey 按 API Key 查找用户（apikey 列存储 SHA256 哈希；忽略已软删除；空 key 直接未找到）。
// 兼容历史遗留的明文 API Key：哈希未命中时回退按明文查一次，命中后立即回写为哈希+末4位，
// 不影响该 Key 本次及后续继续可用，逐步淘汰库中明文残留。
func GetUserByApiKey(apiKey string) (*User, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, sql.ErrNoRows
	}

	hashed := hashAPIKey(apiKey)
	var user User
	err := db.DB.Get(&user, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE apikey = ? AND delete_time IS NULL", hashed)
	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 哈希未命中，回退尝试历史遗留明文（老数据升级前发放的 Key）
	var legacyUser User
	legacyErr := db.DB.Get(&legacyUser, "SELECT "+BuildUserSelectColumns("users")+" FROM users WHERE apikey = ? AND delete_time IS NULL", apiKey)
	if legacyErr != nil {
		return nil, sql.ErrNoRows
	}
	hint := apiKeyHint(apiKey)
	if _, upErr := db.Exec("UPDATE users SET apikey = ?, apikey_hint = ? WHERE id = ? AND apikey = ?", hashed, hint, legacyUser.ID, apiKey); upErr != nil {
		log.Printf("[User] 回写历史明文 API Key 为哈希失败 user_id=%d: %v", legacyUser.ID, upErr)
	}
	legacyUser.Apikey = &hashed
	legacyUser.ApikeyHint = &hint
	return &legacyUser, nil
}

// GenerateApiKey 生成随机 API 密钥明文（供注册自动发放等调用；落库前会被 CreateUser 统一哈希）
func GenerateApiKey() string {
	return generateApiKey()
}

// ResetUserApiKey 重置用户 API 密钥：落库为 SHA256 哈希 + 末4位提示，返回值为明文，且仅此一次可见，
// 调用方（用户自服务/管理员操作）需自行妥善展示给用户后即不再可查完整明文。
func ResetUserApiKey(userID uint64) (string, error) {
	newKey := generateApiKey()
	hashed := hashAPIKey(newKey)
	hint := apiKeyHint(newKey)
	now := time.Now().Unix()
	_, err := db.Exec("UPDATE users SET apikey = ?, apikey_hint = ?, update_time = ? WHERE id = ?", hashed, hint, now, userID)
	if err != nil {
		return "", err
	}
	return newKey, nil
}

// generateApiKey 生成随机API密钥（使用 crypto/rand）
func generateApiKey() string {
	b := make([]byte, 20)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// hashAPIKey 对 API Key 明文做 SHA256 哈希（十六进制），用于落库/查询比对，数据库不保存明文。
func hashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// apiKeyHint 取明文末4位供展示用（不敏感，不足4位原样返回）
func apiKeyHint(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= 4 {
		return raw
	}
	return raw[len(raw)-4:]
}

// IncrementLoginFailure 增加登录失败次数，如果达到最大失败次数则锁定账户
func IncrementLoginFailure(userID uint64, maxFailureCount int, lockDurationMinutes int) error {
	now := time.Now().Unix()
	// 先增加失败次数
	_, err := db.Exec("UPDATE users SET login_failure = login_failure + 1, update_time = ? WHERE id = ?", now, userID)
	if err != nil {
		return err
	}

	// 检查是否需要锁定（需要先查询当前失败次数）
	var user User
	err = db.DB.Get(&user, "SELECT login_failure FROM users WHERE id = ?", userID)
	if err != nil {
		return err
	}

	// 如果达到最大失败次数，设置锁定时间
	if int(user.LoginFailure) >= maxFailureCount {
		lockUntil := now + int64(lockDurationMinutes*60)
		_, err = db.Exec("UPDATE users SET lock_until = ?, update_time = ? WHERE id = ?", lockUntil, now, userID)
		return err
	}

	return nil
}

// UpdateUserTOTPSecret 写入 TOTP 密钥；enabled=true 时同时开启 2FA
func UpdateUserTOTPSecret(userID uint64, secret string, enabled bool) error {
	now := time.Now().Unix()
	en := 0
	if enabled {
		en = 1
	}
	_, err := db.Exec(
		`UPDATE users SET totp_secret = ?, totp_enabled = ?, update_time = ? WHERE id = ? AND delete_time IS NULL`,
		secret, en, now, userID,
	)
	return err
}

// ClearUserTOTP 清空并禁用 TOTP
func ClearUserTOTP(userID uint64) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		`UPDATE users SET totp_secret = NULL, totp_enabled = 0, update_time = ? WHERE id = ? AND delete_time IS NULL`,
		now, userID,
	)
	return err
}

