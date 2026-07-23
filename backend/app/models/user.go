package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fst/backend/pkg/db"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
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
	ID      uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupId uint64 `gorm:"column:group_id" json:"group_id"`
	// MySQL：string 未指定 size 时 GORM 会迁成 longtext，无法建唯一索引（Error 1170）
	Username      string  `gorm:"column:username;size:64;not null;uniqueIndex:idx_users_username" json:"username"`
	Nickname      string  `gorm:"column:nickname;size:64" json:"nickname"`
	Email         string  `gorm:"column:email;size:255;uniqueIndex:idx_users_email" json:"email"`
	Mobile        string  `gorm:"column:mobile;size:32" json:"mobile"`
	Avatar        string  `gorm:"column:avatar;size:512" json:"avatar"`
	BackGround    string  `gorm:"column:back_ground;size:512" json:"back_ground"`
	Gender        uint8   `gorm:"column:gender" json:"gender"`
	Birthday      *int64  `gorm:"column:birthday" json:"birthday"`
	Money         float64 `gorm:"column:money;type:decimal(10,2)" json:"money"` // 余额（元，DECIMAL；业务加减一律经 utils 按「分」整数计算）
	Score         int64   `gorm:"column:score" json:"score"`
	Level         uint64  `gorm:"column:level" json:"level"`
	Role          string  `gorm:"column:role;size:20" json:"role"` // 'user' or 'admin'
	LastLoginTime *int64  `gorm:"column:last_login_time" json:"last_login_time"`
	LastLoginIp   string  `gorm:"column:last_login_ip;size:64" json:"last_login_ip"`
	LoginFailure  uint8   `gorm:"column:login_failure" json:"login_failure"`
	LockUntil     *int64  `gorm:"column:lock_until" json:"lock_until"` // 账户锁定到期时间（时间戳）
	JoinIp        string  `gorm:"column:join_ip;size:64" json:"join_ip"`
	JoinTime      *int64  `gorm:"column:join_time" json:"join_time"`
	Motto         string  `gorm:"column:motto;size:255" json:"motto"`
	AdminRemark   string  `gorm:"column:admin_remark;size:255" json:"-"`
	Password      string  `gorm:"column:password;size:255" json:"-"`
	Status        uint8   `gorm:"column:status;index:idx_users_status" json:"status"`

	// Apikey 存明文（用户端可随时回显/复制）；json:"-" 禁止随用户列表等通用 JSON 泄漏。
	// 管理端列表/详情走 MaskedApikey()；用户本人 GET /apikey 走 PlainApikeyForOwner()。
	Apikey     *string `gorm:"column:apikey;size:128" json:"-"`
	ApikeyHint *string `gorm:"column:apikey_hint;size:8" json:"-"` // 明文末4位，管理端掩码用
	// API Key 收紧：过期时间、IP 白名单（逗号分隔）、scope（user,admin,*）
	ApikeyExpiresAt *int64  `gorm:"column:apikey_expires_at" json:"-"`
	ApikeyAllowIPs  *string `gorm:"column:apikey_allow_ips;size:1024" json:"-"`
	ApikeyScopes    *string `gorm:"column:apikey_scopes;size:64" json:"-"`
	UpdateTime      *int64  `gorm:"column:update_time" json:"update_time"`
	CreateTime      *int64  `gorm:"column:create_time" json:"create_time"`
	DeleteTime      *int64  `gorm:"column:delete_time" json:"-"`

	Language string `gorm:"column:language;size:20" json:"language"`
	Country  string `gorm:"column:country;size:64" json:"country"`

	// 保留列供 AutoMigrate / 旧库兼容；业务代码不再读写 TOTP
	TotpSecret  *string `gorm:"column:totp_secret;size:64" json:"-"`
	TotpEnabled bool    `gorm:"column:totp_enabled" json:"-"`
}

func (u *User) TableName() string {
	return "users"
}

// userSelectQuery 构造带业务列（含 COALESCE）的用户查询。
func userSelectQuery() *gorm.DB {
	return db.DB.Model(&User{}).Select(BuildUserSelectColumns("users"))
}

// firstUser 按条件查单条用户；未找到时返回 sql.ErrNoRows。
func firstUser(where string, args ...interface{}) (*User, error) {
	var user User
	err := userSelectQuery().Where(where, args...).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// MaskedApikey 管理端展示用掩码（******** + 末4位）；无密钥返回空串。
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
	return "********" + apiKeyHint(strings.TrimSpace(*u.Apikey))
}

// PlainApikeyForOwner 用户本人可读的完整密钥（库内明文）。
func (u *User) PlainApikeyForOwner() string {
	if u == nil || u.Apikey == nil {
		return ""
	}
	return strings.TrimSpace(*u.Apikey)
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
	}
	return strings.Join(columns, ", ")
}

// CreateUser inserts a new user into the database
func CreateUser(user *User) error {
	// API Key：落库明文 + 末4位 hint（用户端可随时回显；管理端用 hint 掩码）
	if user.Apikey != nil && strings.TrimSpace(*user.Apikey) != "" {
		plain := strings.TrimSpace(*user.Apikey)
		hint := apiKeyHint(plain)
		user.Apikey = &plain
		user.ApikeyHint = &hint
	}

	now := time.Now().Unix()
	user.CreateTime = &now
	user.UpdateTime = &now
	if user.JoinTime == nil {
		user.JoinTime = &now
	}

	if user.GroupId == 0 {
		user.GroupId = 1
	}
	if user.Level == 0 {
		user.Level = 1
	}

	if user.Language == "" {
		user.Language = "zh-CN"
	}

	return db.DB.Create(user).Error
}

// GetUserByUsername finds a user by username
func GetUserByUsername(username string) (*User, error) {
	return firstUser("username = ? AND delete_time IS NULL", username)
}

// GetUserByEmail finds a user by email
func GetUserByEmail(email string) (*User, error) {
	return firstUser("email = ? AND delete_time IS NULL", email)
}

// GetUserByMobile finds a user by mobile number（兼容 11 位与 +86 E.164 等价写法）
func GetUserByMobile(mobile string) (*User, error) {
	variants := mobileLookupVariants(mobile)
	if len(variants) == 0 {
		return nil, sql.ErrNoRows
	}
	var lastErr error = sql.ErrNoRows
	for _, v := range variants {
		user, err := firstUser("mobile = ? AND delete_time IS NULL", v)
		if err == nil {
			return user, nil
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
	return firstUser("(username = ? OR email = ?) AND delete_time IS NULL", identifier, identifier)
}

// GetUserByID finds a user by ID
func GetUserByID(id uint64) (*User, error) {
	return firstUser("id = ? AND delete_time IS NULL", id)
}

// UpdatePassword updates the user's password
func UpdatePassword(userID uint64, hashedPassword string) error {
	now := time.Now().Unix()
	if err := db.DB.Model(&User{}).Where("id = ? AND delete_time IS NULL", userID).Updates(map[string]interface{}{
		"password":    hashedPassword,
		"update_time": now,
	}).Error; err != nil {
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
	return db.DB.Model(&User{}).Where("id = ? AND delete_time IS NULL", userID).Updates(map[string]interface{}{
		"last_login_time": now,
		"last_login_ip":   clampStoredIP(loginIP),
		"login_failure":   0,
		"lock_until":      nil,
		"update_time":     now,
	}).Error
}

// GetUserByApiKey 按明文 API Key 查找用户（忽略已软删除；空 key 直接未找到）。
func GetUserByApiKey(apiKey string) (*User, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, sql.ErrNoRows
	}
	return firstUser("apikey = ? AND delete_time IS NULL", apiKey)
}

// GenerateApiKey 生成随机 API 密钥明文
func GenerateApiKey() string {
	return generateApiKey()
}

// ResetUserApiKey 重置用户 API 密钥：落库明文 + 末4位 hint，并返回明文供前端立即展示。
func ResetUserApiKey(userID uint64) (string, error) {
	newKey := generateApiKey()
	hint := apiKeyHint(newKey)
	now := time.Now().Unix()
	if err := db.DB.Model(&User{}).Where("id = ? AND delete_time IS NULL", userID).Updates(map[string]interface{}{
		"apikey":      newKey,
		"apikey_hint": hint,
		"update_time": now,
	}).Error; err != nil {
		return "", err
	}
	return newKey, nil
}

// generateApiKey 生成随机API密钥（使用 crypto/rand）；40 位 hex。
func generateApiKey() string {
	b := make([]byte, 20)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// apiKeyHint 取明文末4位供展示用（不敏感，不足4位原样返回）
func apiKeyHint(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= 4 {
		return raw
	}
	return raw[len(raw)-4:]
}

// looksLikeLegacyHashedApiKey 旧版曾把 SHA256(hex,64) 写进 apikey 列；现行明文为 40 位 hex。
func looksLikeLegacyHashedApiKey(stored string) bool {
	if len(stored) != 64 {
		return false
	}
	for _, c := range stored {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// RepairHashedApiKeys 启动时把库内旧哈希密钥直接重置为新明文（开发期清理，不做双读兼容）。
func RepairHashedApiKeys() {
	if db.DB == nil {
		return
	}
	type row struct {
		ID     uint64
		Apikey string
	}
	var rows []row
	if err := db.DB.Model(&User{}).
		Select("id, apikey").
		Where("delete_time IS NULL AND apikey IS NOT NULL AND apikey <> ''").
		Find(&rows).Error; err != nil {
		log.Printf("[Migrate] 扫描 apikey 失败: %v", err)
		return
	}
	rotated := 0
	for _, r := range rows {
		if !looksLikeLegacyHashedApiKey(strings.TrimSpace(r.Apikey)) {
			continue
		}
		if _, err := ResetUserApiKey(r.ID); err != nil {
			log.Printf("[Migrate] 重置用户 %d 旧哈希 apikey 失败: %v", r.ID, err)
			continue
		}
		rotated++
	}
	if rotated > 0 {
		log.Printf("[Migrate] 已将 %d 个旧哈希 API Key 重置为明文", rotated)
	}
}

// IncrementLoginFailure 增加登录失败次数，如果达到最大失败次数则锁定账户
func IncrementLoginFailure(userID uint64, maxFailureCount int, lockDurationMinutes int) error {
	now := time.Now().Unix()
	if err := db.DB.Model(&User{}).Where("id = ? AND delete_time IS NULL", userID).Updates(map[string]interface{}{
		"login_failure": gorm.Expr("login_failure + 1"),
		"update_time":   now,
	}).Error; err != nil {
		return err
	}

	var user User
	if err := db.DB.Model(&User{}).Select("login_failure").Where("id = ? AND delete_time IS NULL", userID).First(&user).Error; err != nil {
		return err
	}

	if int(user.LoginFailure) >= maxFailureCount {
		lockUntil := now + int64(lockDurationMinutes*60)
		return db.DB.Model(&User{}).Where("id = ? AND delete_time IS NULL", userID).Updates(map[string]interface{}{
			"lock_until":  lockUntil,
			"update_time": now,
		}).Error
	}

	return nil
}
