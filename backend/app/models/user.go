package models

import (
	"fst/backend/internal/db"
	"time"
)

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
	Money         float64 `db:"money" json:"money"`
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
	Password      string  `db:"password" json:"-"`
	Status        uint8   `db:"status" json:"status"`

	// 兼容旧表里的 is_active 字段（不在业务中使用，仅为避免扫描报错）
	IsActive *uint8 `db:"is_active" json:"-"`

	// 兼容旧表里的 created_at / updated_at / deleted_at 字段（不在业务中使用，仅为避免扫描报错）
	CreatedAtRaw *time.Time `db:"created_at" json:"-"`
	UpdatedAtRaw *time.Time `db:"updated_at" json:"-"`
	DeletedAtRaw *time.Time `db:"deleted_at" json:"-"`

	Apikey     *string `db:"apikey" json:"apikey"`
	UpdateTime *int64  `db:"update_time" json:"update_time"`
	CreateTime *int64  `db:"create_time" json:"create_time"`
	DeleteTime *int64  `db:"delete_time" json:"-"`

	// Requested additions
	Language string `db:"language" json:"language"`
	Country  string `db:"country" json:"country"`
	Token    string `db:"token" json:"token"`
}

func (u *User) TableName() string {
	return "users"
}

// CreateUser inserts a new user into the database
func CreateUser(user *User) error {
	query := `INSERT INTO users (
		group_id, username, nickname, email, mobile, avatar, back_ground, gender, birthday, 
		money, score, level, role, last_login_time, last_login_ip, login_failure, 
		join_ip, join_time, motto, password, status, apikey, update_time, create_time, 
		language, country, token
	) VALUES (
		:group_id, :username, :nickname, :email, :mobile, :avatar, :back_ground, :gender, :birthday, 
		:money, :score, :level, :role, :last_login_time, :last_login_ip, :login_failure, 
		:join_ip, :join_time, :motto, :password, :status, :apikey, :update_time, :create_time, 
		:language, :country, :token
	)`

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
	if user.Status == 0 {
		// Caution: if user wants to create inactive user (0), this overrides.
		// But usually creation implies active or pending.
		// If 0 is valid "inactive", we shouldn't force 1.
		// However, Go int zero value is 0. If I don't set it, it inserts 0.
		// DB default is 1.
		// If I want DB default, I should not include it in INSERT or use NULL?
		// NamedExec doesn't skip zero values easily.
		// I will set it to 1 if it is meant to be active by default.
		// Let's assume standard creation is active.
		user.Status = 1
	}

	// Ensure language default
	if user.Language == "" {
		user.Language = "zh-CN"
	}

	result, err := db.DB.NamedExec(query, user)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint64(id)
	return nil
}

// GetUserByUsername finds a user by username
func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT * FROM users WHERE username = ? AND delete_time IS NULL", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail finds a user by email
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT * FROM users WHERE email = ? AND delete_time IS NULL", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsernameOrEmail finds a user by username or email
func GetUserByUsernameOrEmail(identifier string) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT * FROM users WHERE (username = ? OR email = ?) AND delete_time IS NULL", identifier, identifier)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by ID
func GetUserByID(id uint64) (*User, error) {
	var user User
	err := db.DB.Get(&user, "SELECT * FROM users WHERE id = ? AND delete_time IS NULL", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates the user's password
func UpdatePassword(userID uint64, hashedPassword string) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET password = ?, update_time = ? WHERE id = ?", hashedPassword, now, userID)
	return err
}

// UpdateLoginInfo 更新用户登录信息（成功登录后调用）
func UpdateLoginInfo(userID uint64, loginIP string) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec(
		"UPDATE users SET last_login_time = ?, last_login_ip = ?, login_failure = 0, lock_until = NULL, update_time = ? WHERE id = ?",
		now, loginIP, now, userID,
	)
	return err
}

// IncrementLoginFailure 增加登录失败次数，如果达到最大失败次数则锁定账户
func IncrementLoginFailure(userID uint64, maxFailureCount int, lockDurationMinutes int) error {
	now := time.Now().Unix()
	// 先增加失败次数
	_, err := db.DB.Exec("UPDATE users SET login_failure = login_failure + 1, update_time = ? WHERE id = ?", now, userID)
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
		_, err = db.DB.Exec("UPDATE users SET lock_until = ?, update_time = ? WHERE id = ?", lockUntil, now, userID)
		return err
	}

	return nil
}
