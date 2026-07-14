package models

import "gorm.io/gorm"

// Gender 性别常量
type Gender uint8

const (
	GenderUnknown Gender = 0 // 未知
	GenderMale    Gender = 1 // 男
	GenderFemale  Gender = 2 // 女
)

// UserStatus 用户状态常量
type UserStatus uint8

const (
	UserStatusInactive UserStatus = 0 // 禁用
	UserStatusActive   UserStatus = 1 // 启用
)

// UserRole 用户角色常量
type UserRole string

const (
	RoleUser  UserRole = "user"  // 普通用户
	RoleAdmin UserRole = "admin" // 管理员
)

// User 用户模型
// 包含增强字段: 国家、地区、语言
type User struct {
	ID         uint    `json:"id" gorm:"primaryKey;type:bigint(20) unsigned;autoIncrement;comment:用户ID"`
	GroupID    uint    `json:"group_id" gorm:"index;default:1;not null;comment:分组ID"`
	Username   string  `json:"username" gorm:"uniqueIndex;size:100;not null;comment:用户名"`
	Nickname   string  `json:"nickname" gorm:"size:100;not null;default:'';comment:昵称"`
	Email      string  `json:"email" gorm:"uniqueIndex;size:150;not null;default:'';comment:邮箱"`
	Mobile     string  `json:"mobile" gorm:"index;size:50;not null;default:'';comment:手机"`
	Avatar     string  `json:"avatar" gorm:"size:255;not null;default:'';comment:头像URL"`
	Background string  `json:"background" gorm:"size:255;not null;default:'';comment:背景图URL"`
	Gender     Gender  `json:"gender" gorm:"default:0;not null;comment:性别:0=未知,1=男,2=女"`
	Birthday   *uint64 `json:"birthday" gorm:"type:bigint(16) unsigned;comment:生日时间戳"`

	// 地区信息字段 (新增)
	Country  string `json:"country" gorm:"size:100;not null;default:'';comment:国家"`
	Region   string `json:"region" gorm:"size:100;not null;default:'';comment:地区/省份"`
	Language string `json:"language" gorm:"size:20;not null;default:'zh-CN';comment:语言偏好"`

	Money float64  `json:"money" gorm:"type:decimal(12,2);index;default:0.00;not null;comment:账户余额"`
	Score int64    `json:"score" gorm:"index;default:0;not null;comment:积分"`
	Level uint     `json:"level" gorm:"index;default:1;not null;comment:用户等级"`
	Role  UserRole `json:"role" gorm:"index;size:20;not null;default:'user';comment:角色:user=普通用户,admin=管理员"`

	LastLoginTime *uint64 `json:"last_login_time" gorm:"index;type:bigint(16) unsigned;comment:上次登录时间戳"`
	LastLoginIP   string  `json:"last_login_ip" gorm:"size:50;not null;default:'';comment:上次登录IP"`
	LoginFailure  uint8   `json:"login_failure" gorm:"default:0;not null;comment:连续登录失败次数"`
	JoinIP        string  `json:"join_ip" gorm:"size:50;not null;default:'';comment:注册IP"`
	JoinTime      *uint64 `json:"join_time" gorm:"index;type:bigint(16) unsigned;comment:注册时间戳"`

	Motto    string     `json:"motto" gorm:"size:255;not null;default:'';comment:个性签名"`
	Password string     `json:"-" gorm:"size:255;not null;default:'';comment:密码(不返回给前端)"`
	Status   UserStatus `json:"status" gorm:"index;default:1;not null;comment:状态:1=启用,0=禁用"`
	APIKey   string     `json:"apikey" gorm:"uniqueIndex;size:255;comment:API密钥"`

	CreateTime *uint64        `json:"create_time" gorm:"type:bigint(16) unsigned;comment:创建时间戳"`
	UpdateTime *uint64        `json:"update_time" gorm:"type:bigint(16) unsigned;comment:更新时间戳"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`  // 用户名
	Nickname string `json:"nickname" binding:"omitempty,max=50"`       // 昵称
	Email    string `json:"email" binding:"required,email"`            // 邮箱
	Mobile   string `json:"mobile" binding:"omitempty,max=50"`         // 手机
	Password string `json:"password" binding:"required,min=6,max=100"` // 密码
	Country  string `json:"country" binding:"omitempty,max=100"`       // 国家
	Region   string `json:"region" binding:"omitempty,max=100"`        // 地区
	Language string `json:"language" binding:"omitempty,max=20"`       // 语言
	IP       string `json:"-"`                                         // 注册IP(服务器自动获取)
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"` // 密码
	IP       string `json:"-"`                           // 登录IP(服务器自动获取)
}

// UserUpdateRequestForUser 用户更新个人信息请求
type UserUpdateRequestForUser struct {
	Nickname   string  `json:"nickname" binding:"omitempty,max=50"`    // 昵称
	Email      string  `json:"email" binding:"omitempty,email"`        // 邮箱
	Mobile     string  `json:"mobile" binding:"omitempty,max=50"`      // 手机
	Avatar     string  `json:"avatar" binding:"omitempty"`             // 头像
	Background string  `json:"background" binding:"omitempty"`         // 背景图
	Gender     *Gender `json:"gender" binding:"omitempty,min=0,max=2"` // 性别
	Birthday   *uint64 `json:"birthday" binding:"omitempty"`           // 生日
	Motto      string  `json:"motto" binding:"omitempty,max=255"`      // 签名
	Country    string  `json:"country" binding:"omitempty,max=100"`    // 国家
	Region     string  `json:"region" binding:"omitempty,max=100"`     // 地区
	Language   string  `json:"language" binding:"omitempty,max=20"`    // 语言
}

// UserUpdateRequestForAdmin 管理员更新用户信息请求
type UserUpdateRequestForAdmin struct {
	GroupID    *uint       `json:"group_id"`                               // 分组ID
	Nickname   string      `json:"nickname" binding:"omitempty,max=50"`    // 昵称
	Email      string      `json:"email" binding:"omitempty,email"`        // 邮箱
	Mobile     string      `json:"mobile" binding:"omitempty,max=50"`      // 手机
	Avatar     string      `json:"avatar" binding:"omitempty"`             // 头像
	Background string      `json:"background" binding:"omitempty"`         // 背景图
	Gender     *Gender     `json:"gender" binding:"omitempty,min=0,max=2"` // 性别
	Birthday   *uint64     `json:"birthday" binding:"omitempty"`           // 生日
	Motto      string      `json:"motto" binding:"omitempty,max=255"`      // 签名
	Country    string      `json:"country" binding:"omitempty,max=100"`    // 国家
	Region     string      `json:"region" binding:"omitempty,max=100"`     // 地区
	Language   string      `json:"language" binding:"omitempty,max=20"`    // 语言
	Money      *float64    `json:"money" binding:"omitempty"`              // 余额
	Score      *int64      `json:"score" binding:"omitempty"`              // 积分
	Level      *uint       `json:"level" binding:"omitempty"`              // 等级
	Role       UserRole    `json:"role" binding:"omitempty,max=20"`        // 角色
	Status     *UserStatus `json:"status" binding:"omitempty"`             // 状态
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`               // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"` // 新密码
}
