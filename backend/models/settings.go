package models

import "gorm.io/gorm"

// Settings 系统设置模型
// 用于存储可开关的系统功能配置
type Settings struct {
	ID          uint   `json:"id" gorm:"primaryKey;type:bigint(20) unsigned;autoIncrement;comment:设置ID"`
	Key         string `json:"key" gorm:"uniqueIndex;size:100;not null;comment:设置键名"`
	Value       string `json:"value" gorm:"type:text;comment:设置值"`
	Type        string `json:"type" gorm:"size:20;default:'string';comment:值类型:string/int/bool/json"`
	Description string `json:"description" gorm:"size:255;default:'';comment:设置描述"`
	Category    string `json:"category" gorm:"index;size:50;default:'general';comment:设置分类"`
	IsPublic    bool   `json:"is_public" gorm:"default:false;comment:是否公开(前端可读取)"`

	CreateTime *uint64        `json:"create_time" gorm:"type:bigint(16) unsigned;comment:创建时间戳"`
	UpdateTime *uint64        `json:"update_time" gorm:"type:bigint(16) unsigned;comment:更新时间戳"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (Settings) TableName() string {
	return "settings"
}

// 系统设置键名常量
const (
	// 基础设置
	SettingSiteName        = "site_name"        // 站点名称
	SettingSiteLogo        = "site_logo"        // 站点Logo
	SettingSiteDescription = "site_description" // 站点描述

	// 功能开关
	SettingEnableC2C      = "enable_c2c"      // 启用C2C功能
	SettingEnableB2C      = "enable_b2c"      // 启用B2C功能
	SettingEnableSaaS     = "enable_saas"     // 启用SaaS功能
	SettingEnableRegister = "enable_register" // 启用用户注册
	SettingEnableAPI      = "enable_api"      // 启用API访问

	// 交易设置
	SettingCurrencyDefault = "currency_default" // 默认货币
	SettingTaxRate         = "tax_rate"         // 税率
	SettingMinOrderAmount  = "min_order_amount" // 最小订单金额

	// 支付设置
	SettingPaymentAlipay = "payment_alipay" // 支付宝支付开关
	SettingPaymentWechat = "payment_wechat" // 微信支付开关
	SettingPaymentStripe = "payment_stripe" // Stripe支付开关

	// 邮件设置
	SettingEmailSMTPHost = "email_smtp_host" // SMTP主机
	SettingEmailSMTPPort = "email_smtp_port" // SMTP端口
	SettingEmailUsername = "email_username"  // 邮箱用户名
	SettingEmailPassword = "email_password"  // 邮箱密码
	SettingEmailFrom     = "email_from"      // 发件人地址
	SettingEmailFromName = "email_from_name" // 发件人名称
)

// SettingsCreateRequest 创建设置请求
type SettingsCreateRequest struct {
	Key         string `json:"key" binding:"required,max=100"`                      // 设置键名
	Value       string `json:"value" binding:"required"`                            // 设置值
	Type        string `json:"type" binding:"omitempty,oneof=string int bool json"` // 值类型
	Description string `json:"description" binding:"omitempty,max=255"`             // 设置描述
	Category    string `json:"category" binding:"omitempty,max=50"`                 // 设置分类
	IsPublic    bool   `json:"is_public"`                                           // 是否公开
}

// SettingsUpdateRequest 更新设置请求
type SettingsUpdateRequest struct {
	Value       string `json:"value" binding:"omitempty"`                           // 设置值
	Type        string `json:"type" binding:"omitempty,oneof=string int bool json"` // 值类型
	Description string `json:"description" binding:"omitempty,max=255"`             // 设置描述
	Category    string `json:"category" binding:"omitempty,max=50"`                 // 设置分类
	IsPublic    *bool  `json:"is_public" binding:"omitempty"`                       // 是否公开
}
