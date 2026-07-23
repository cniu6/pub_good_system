package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SystemSetting 系统配置项
type SystemSetting struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Key         string    `gorm:"column:setting_key;size:100;not null;uniqueIndex:idx_setting_key" json:"key"`
	Value       string    `gorm:"column:setting_value;type:text;not null" json:"value"`
	Type        string    `gorm:"column:setting_type;size:20;not null;default:'string'" json:"type"` // string, number, boolean, json
	Category    string    `gorm:"column:category;size:50;not null;default:'basic';index:idx_category" json:"category"`
	Label       string    `gorm:"column:label;size:100;not null" json:"label"`
	Description string    `gorm:"column:description;size:255;not null;default:''" json:"description"`
	IsPublic    bool      `gorm:"column:is_public;not null;default:false;index:idx_is_public" json:"is_public"`
	IsEditable  bool      `gorm:"column:is_editable;not null;default:true" json:"is_editable"`
	SortOrder   int       `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 返回表名
func (SystemSetting) TableName() string {
	return "system_settings"
}

// SettingDTO 设置项传输对象（用于API返回）
type SettingDTO struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Category    string      `json:"category"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	IsPublic    bool        `json:"is_public"`
	IsEditable  bool        `json:"is_editable"`
}

// SettingsGroup 按分类分组的设置
type SettingsGroup struct {
	Category string       `json:"category"`
	Label    string       `json:"label"`
	Items    []SettingDTO `json:"items"`
}

// GetTypedValue 根据类型返回正确的值类型
func (s *SystemSetting) GetTypedValue() interface{} {
	switch s.Type {
	case "number":
		num, err := strconv.ParseFloat(strings.TrimSpace(s.Value), 64)
		if err != nil {
			return float64(0)
		}
		return num
	case "boolean":
		return s.Value == "true" || s.Value == "1"
	case "json":
		var data interface{}
		if err := json.Unmarshal([]byte(s.Value), &data); err == nil {
			return data
		}
		return nil
	default:
		return s.Value
	}
}

// SeedSystemSettings 写入系统设置默认值（建表由 GORM AutoMigrate 负责）
func SeedSystemSettings() {
	initDefaultSettings()
}

// 默认配置项定义（内容与原文件一致，此处省略注释以控制篇幅）
var defaultSettings = []SystemSetting{
	{Key: "site_name", Value: "F.st", Type: "string", Category: "basic", Label: "系统名称", Description: "显示在浏览器标签和页面的系统名称", IsPublic: true, IsEditable: true, SortOrder: 1},
	{Key: "site_desc", Value: "基于 Go + Vue 3 的全栈管理系统模板", Type: "string", Category: "basic", Label: "系统描述", Description: "系统简介描述", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "site_logo", Value: "", Type: "string", Category: "basic", Label: "站点Logo", Description: "站点Logo图片URL", IsPublic: true, IsEditable: true, SortOrder: 3},
	{Key: "copyright", Value: "© 2024 F.st", Type: "string", Category: "basic", Label: "版权信息", Description: "页脚版权声明", IsPublic: true, IsEditable: true, SortOrder: 4},
	{Key: "icp", Value: "", Type: "string", Category: "basic", Label: "ICP备案号", Description: "网站ICP备案号", IsPublic: true, IsEditable: true, SortOrder: 5},
	{Key: "allow_register", Value: "true", Type: "boolean", Category: "basic", Label: "允许注册", Description: "是否允许新用户注册", IsPublic: true, IsEditable: true, SortOrder: 6},
	{Key: "announcement_enabled", Value: "true", Type: "boolean", Category: "basic", Label: "站内公告", Description: "关闭后前台不展示公告入口与内容", IsPublic: true, IsEditable: true, SortOrder: 6},
	{Key: "default_lang", Value: "zhCN", Type: "string", Category: "basic", Label: "默认语言", Description: "系统默认语言", IsPublic: true, IsEditable: true, SortOrder: 7},
	{Key: "version", Value: "1.0.0", Type: "string", Category: "basic", Label: "系统版本", Description: "当前系统版本号", IsPublic: true, IsEditable: true, SortOrder: 8},
	{Key: "allow_delete_account", Value: "false", Type: "boolean", Category: "basic", Label: "允许注销账号", Description: "是否允许用户自助注销账号", IsPublic: true, IsEditable: true, SortOrder: 7},
	{Key: "frontend_url", Value: "", Type: "string", Category: "basic", Label: "前端地址", Description: "前端访问地址（如 http://example.com），结尾不要加 /", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "backend_api_url", Value: "", Type: "string", Category: "basic", Label: "后端API地址", Description: "后端API外网地址（如 http://api.example.com），结尾不要加 /", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "geetest_enabled", Value: "false", Type: "boolean", Category: "security", Label: "极验验证码", Description: "是否启用极验行为验证", IsPublic: true, IsEditable: true, SortOrder: 1},
	{Key: "geetest_captcha_id", Value: "", Type: "string", Category: "security", Label: "极验 Captcha ID", Description: "极验验证码 ID", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "geetest_captcha_key", Value: "", Type: "string", Category: "security", Label: "极验 Captcha Key", Description: "极验验证码 Key", IsPublic: false, IsEditable: true, SortOrder: 3},
	{Key: "jwt_access_expire", Value: "7200", Type: "number", Category: "security", Label: "Token有效期", Description: "Access Token 有效期（秒）", IsPublic: false, IsEditable: true, SortOrder: 4},
	{Key: "jwt_refresh_expire", Value: "604800", Type: "number", Category: "security", Label: "Refresh Token有效期", Description: "Refresh Token 有效期（秒）", IsPublic: false, IsEditable: true, SortOrder: 5},
	{Key: "register_code_expire_minutes", Value: "60", Type: "number", Category: "security", Label: "注册验证码有效期", Description: "注册验证码有效期（分钟）", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "login_max_failure", Value: "5", Type: "number", Category: "security", Label: "登录失败锁定次数", Description: "连续登录失败多少次后锁定账户", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "login_lock_duration", Value: "10", Type: "number", Category: "security", Label: "账户锁定时长", Description: "账户锁定时长（分钟）", IsPublic: false, IsEditable: true, SortOrder: 7},
	{Key: "disable_web_login", Value: "false", Type: "boolean", Category: "security", Label: "禁止网页端登录", Description: "开启后，普通用户（非管理员）无法通过网页/浏览器直接登录；登录请求需带 client_type=app（如小程序/App）才能通过，管理员登录不受影响。适用于仅通过小程序/App 对外提供服务的场景。注意：client_type 由请求自报，不做客户端可信校验，这是一个引导前端 UX 的软限制，不构成强制安全边界", IsPublic: true, IsEditable: true, SortOrder: 8},
	{Key: "operation_log_query_days", Value: "30", Type: "number", Category: "security", Label: "操作日志查询天数", Description: "操作日志默认查询范围（天）", IsPublic: false, IsEditable: true, SortOrder: 8},
	{Key: "operation_log_max_count", Value: "1000", Type: "number", Category: "security", Label: "操作日志保留上限", Description: "操作日志自动保留的最大总条数", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "operation_log_per_user_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "操作日志每用户上限开关", Description: "开启后额外限制每个用户保留的操作日志条数", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "operation_log_per_user_max_count", Value: "1000", Type: "number", Category: "security", Label: "操作日志每用户上限", Description: "每个用户最多保留的操作日志条数（需开启开关）", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "api_access_log_enabled", Value: "true", Type: "boolean", Category: "security", Label: "启用API接口日志", Description: "是否记录API接口访问日志（请求/响应体截断入库；请求头凭证字段脱敏）", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "api_log_query_days", Value: "7", Type: "number", Category: "security", Label: "API日志查询天数", Description: "API接口日志默认查询范围（天）", IsPublic: false, IsEditable: true, SortOrder: 11},
	{Key: "api_log_max_count", Value: "1000", Type: "number", Category: "security", Label: "API日志保留上限", Description: "API接口日志自动保留的最大条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "api_log_per_user_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "API日志每用户上限开关", Description: "开启后额外限制每个用户保留的API日志条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "api_log_per_user_max_count", Value: "1000", Type: "number", Category: "security", Label: "API日志每用户上限", Description: "每个用户最多保留的API日志条数（需开启开关）", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "sms_log_max_count", Value: "1000", Type: "number", Category: "security", Label: "短信日志保留上限", Description: "短信日志自动保留的最大总条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "sms_log_per_user_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "短信日志每收件人上限开关", Description: "开启后额外限制每个手机号保留的短信日志条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "sms_log_per_user_max_count", Value: "1000", Type: "number", Category: "security", Label: "短信日志每收件人上限", Description: "每个手机号最多保留的短信日志条数（需开启开关）", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "email_log_max_count", Value: "1000", Type: "number", Category: "security", Label: "邮件日志保留上限", Description: "邮件日志自动保留的最大总条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "email_log_per_user_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "邮件日志每收件人上限开关", Description: "开启后额外限制每个邮箱保留的邮件日志条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "email_log_per_user_max_count", Value: "1000", Type: "number", Category: "security", Label: "邮件日志每收件人上限", Description: "每个邮箱最多保留的邮件日志条数（需开启开关）", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "api_rate_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "启用全局API限流", Description: "是否对全部 /api 请求启用基础限流（开发环境默认关闭）", IsPublic: false, IsEditable: true, SortOrder: 13},
	{Key: "api_rate_limit_rate", Value: "120", Type: "number", Category: "security", Label: "全局API每秒速率", Description: "全局API限流每秒允许请求数", IsPublic: false, IsEditable: true, SortOrder: 14},
	{Key: "api_rate_limit_burst", Value: "240", Type: "number", Category: "security", Label: "全局API突发上限", Description: "全局API限流突发流量上限", IsPublic: false, IsEditable: true, SortOrder: 15},
	{Key: "admin_rate_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "启用管理端限流", Description: "是否对管理员后台接口额外启用更严格的限流", IsPublic: false, IsEditable: true, SortOrder: 16},
	{Key: "admin_rate_limit_rate", Value: "60", Type: "number", Category: "security", Label: "管理端每秒速率", Description: "管理员后台接口每秒允许请求数", IsPublic: false, IsEditable: true, SortOrder: 17},
	{Key: "admin_rate_limit_burst", Value: "120", Type: "number", Category: "security", Label: "管理端突发上限", Description: "管理员后台接口限流突发流量上限", IsPublic: false, IsEditable: true, SortOrder: 18},
	{Key: "api_key_auth_enabled", Value: "false", Type: "boolean", Category: "security", Label: "允许APIKey鉴权", Description: "关闭后所有 X-Api-Key 请求直接拒绝（仅允许 Authorization: Bearer 登录），默认关闭，需管理员主动开启", IsPublic: false, IsEditable: true, SortOrder: 19},
	{Key: "finance_dual_approval", Value: "false", Type: "boolean", Category: "payment", Label: "财务双人复核", Description: "开启后强制补单等高危财务操作需另一管理员审批后才生效；默认关闭", IsPublic: false, IsEditable: true, SortOrder: 30},
	{Key: "email_verify_enabled", Value: "true", Type: "boolean", Category: "email", Label: "邮箱验证码", Description: "是否启用邮箱验证码功能（关闭后修改邮箱无需验证）", IsPublic: true, IsEditable: true, SortOrder: 0},
	{Key: "smtp_host", Value: "", Type: "string", Category: "email", Label: "SMTP服务器", Description: "SMTP邮件服务器地址", IsPublic: false, IsEditable: true, SortOrder: 1},
	{Key: "smtp_port", Value: "587", Type: "number", Category: "email", Label: "SMTP端口", Description: "SMTP服务器端口", IsPublic: false, IsEditable: true, SortOrder: 2},
	{Key: "smtp_username", Value: "", Type: "string", Category: "email", Label: "发件人邮箱", Description: "SMTP登录用户名/邮箱", IsPublic: false, IsEditable: true, SortOrder: 3},
	{Key: "smtp_password", Value: "", Type: "string", Category: "email", Label: "邮箱密码", Description: "SMTP登录密码或应用密钥", IsPublic: false, IsEditable: true, SortOrder: 4},
	{Key: "smtp_ssl", Value: "false", Type: "boolean", Category: "email", Label: "SSL加密", Description: "按端口选择：465请开启（SSL）；587/25请关闭（STARTTLS）。两种都是加密，配错会握手失败", IsPublic: false, IsEditable: true, SortOrder: 5},
	{Key: "smtp_proxy_enabled", Value: "false", Type: "boolean", Category: "email", Label: "SMTP出站代理", Description: "开启后邮件经代理发送（国内访问 Yandex/Gmail 等）；关闭则直连，避免误配", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "smtp_proxy_type", Value: "socks5", Type: "string", Category: "email", Label: "代理类型", Description: "http / https / socks5 / socks5h（推荐 socks5 或 http）", IsPublic: false, IsEditable: true, SortOrder: 7},
	{Key: "smtp_proxy_host", Value: "", Type: "string", Category: "email", Label: "代理地址", Description: "代理服务器主机名或 IP", IsPublic: false, IsEditable: true, SortOrder: 8},
	{Key: "smtp_proxy_port", Value: "1080", Type: "number", Category: "email", Label: "代理端口", Description: "代理端口，常见 1080/7890/10808", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "smtp_proxy_username", Value: "", Type: "string", Category: "email", Label: "代理用户名", Description: "代理认证用户名（可选）", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "smtp_proxy_password", Value: "", Type: "string", Category: "email", Label: "代理密码", Description: "代理认证密码（可选）", IsPublic: false, IsEditable: true, SortOrder: 11},
	{Key: "system_email_address", Value: "", Type: "string", Category: "email", Label: "系统发件邮箱", Description: "邮件头中显示的发件邮箱地址；留空时回退 SMTP 登录邮箱", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "system_email_name", Value: "F.st", Type: "string", Category: "email", Label: "发件人名称", Description: "邮件中显示的发件人名称", IsPublic: false, IsEditable: true, SortOrder: 13},
	{Key: "sms_verify_enabled", Value: "false", Type: "boolean", Category: "sms", Label: "短信验证码", Description: "是否启用短信验证码功能（关闭后修改手机号无需验证）", IsPublic: true, IsEditable: true, SortOrder: 0},
	{Key: "mobile_cn_only", Value: "true", Type: "boolean", Category: "sms", Label: "仅中国大陆手机号", Description: "开启后仅允许中国大陆手机号（+86 / 11位）；关闭后允许国际号（国家区号选择 + 本地号码）", IsPublic: true, IsEditable: true, SortOrder: 1},
	{Key: "mobile_ip_country_detect", Value: "false", Type: "boolean", Category: "sms", Label: "按IP自动匹配国家", Description: "仅在关闭「仅中国大陆手机号」时生效：根据客户端IP/CDN国家头自动预选区号；关闭则按界面语言，再保底美国(+1)", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "sms_provider", Value: "console", Type: "string", Category: "sms", Label: "短信服务商", Description: "短信服务商标识：console(仅控制台日志)、aliyun(阿里云 SendSms)、tencent(腾讯云 SendSms)、custom(自定义HTTP)", IsPublic: false, IsEditable: true, SortOrder: 3},
	{Key: "sms_access_key", Value: "", Type: "string", Category: "sms", Label: "AccessKey", Description: "阿里云填 AccessKeyId，腾讯云填 SecretId，自定义网关填 X-Api-Key", IsPublic: false, IsEditable: true, SortOrder: 4},
	{Key: "sms_secret_key", Value: "", Type: "string", Category: "sms", Label: "SecretKey", Description: "阿里云填 AccessKeySecret，腾讯云填 SecretKey，自定义网关作为 Authorization Bearer", IsPublic: false, IsEditable: true, SortOrder: 5},
	{Key: "sms_sign_name", Value: "", Type: "string", Category: "sms", Label: "短信签名", Description: "必须填写服务商控制台已审核通过的签名内容（腾讯云填签名内容而不是签名ID）", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "sms_template_code", Value: "", Type: "string", Category: "sms", Label: "默认模板ID", Description: "默认短信模板 ID / Code：阿里云对应 TemplateCode，腾讯云对应 TemplateId", IsPublic: false, IsEditable: true, SortOrder: 7},
	{Key: "sms_template_code_en", Value: "", Type: "string", Category: "sms", Label: "英文模板ID", Description: "英文短信模板 ID / Code，en-US 时优先使用，未填写时回退默认模板", IsPublic: false, IsEditable: true, SortOrder: 8},
	{Key: "sms_region", Value: "", Type: "string", Category: "sms", Label: "服务区域", Description: "服务区域；留空时阿里云默认 cn-hangzhou，腾讯云默认 ap-guangzhou", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "sms_sdk_app_id", Value: "", Type: "string", Category: "sms", Label: "腾讯云 AppID", Description: "腾讯云短信 SmsSdkAppId，仅 tencent Provider 必填", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "sms_endpoint", Value: "", Type: "string", Category: "sms", Label: "HTTP Endpoint", Description: "自定义短信网关请求地址，仅 custom Provider 使用", IsPublic: false, IsEditable: true, SortOrder: 11},
	{Key: "sms_body_format", Value: "json", Type: "string", Category: "sms", Label: "请求体格式", Description: "自定义短信网关请求体格式：json 或 form，仅 custom Provider 使用", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "payment_enabled", Value: "false", Type: "boolean", Category: "payment", Label: "支付功能", Description: "是否启用在线支付充值功能", IsPublic: true, IsEditable: true, SortOrder: 0},
	{Key: "payment_order_expire_minutes", Value: "30", Type: "number", Category: "payment", Label: "订单有效期", Description: "订单有效期（分钟），超时自动取消", IsPublic: false, IsEditable: true, SortOrder: 1},
	{Key: "withdraw_enabled", Value: "true", Type: "boolean", Category: "payment", Label: "提现功能", Description: "是否启用用户提现申请功能", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "withdraw_min_amount", Value: "10", Type: "number", Category: "payment", Label: "最低提现金额", Description: "用户单次提现的最低金额", IsPublic: true, IsEditable: true, SortOrder: 3},
	{Key: "withdraw_notify_text", Value: "", Type: "string", Category: "payment", Label: "提现提示语", Description: "显示在用户提现页面的说明文案（留空则使用系统默认多语言文案）", IsPublic: true, IsEditable: true, SortOrder: 4},
	{Key: "withdraw_account_types", Value: "[\"bank\",\"alipay\",\"wechat\",\"usdt\"]", Type: "json", Category: "payment", Label: "支持收款方式", Description: "用户可选择的提现收款方式列表(JSON数组)", IsPublic: true, IsEditable: true, SortOrder: 5},
	{Key: "withdraw_require_realname", Value: "false", Type: "boolean", Category: "payment", Label: "提现需要实名认证", Description: "开启后，用户提现前必须已完成实名认证并通过审核，否则拒绝提现申请；默认关闭，不影响未接入实名认证的现网", IsPublic: true, IsEditable: true, SortOrder: 6},
	{Key: "realname_enabled", Value: "true", Type: "boolean", Category: "security", Label: "实名认证功能", Description: "是否启用实名认证功能入口；仅控制站内实名入口，不代表已接第三方实名服务", IsPublic: true, IsEditable: true, SortOrder: 19},
	{Key: "realname_review_required", Value: "true", Type: "boolean", Category: "security", Label: "实名认证审核", Description: "是否需要管理员审核实名认证申请；当前仓库默认仍是站内人工审核流", IsPublic: false, IsEditable: true, SortOrder: 20},
	{Key: "realname_notify_text", Value: "", Type: "string", Category: "security", Label: "实名认证提示语", Description: "显示在用户实名认证页面的说明文案（可用于提示人工审核或第三方核验说明）", IsPublic: true, IsEditable: true, SortOrder: 21},
	{Key: "realname_api_enabled", Value: "false", Type: "boolean", Category: "security", Label: "第三方实名API", Description: "是否启用第三方实名核验接口（需配置密钥后才真正生效）", IsPublic: false, IsEditable: true, SortOrder: 22},
	{Key: "realname_api_provider", Value: "aliyun", Type: "string", Category: "security", Label: "实名API服务商", Description: "aliyun / tencent / baidu / custom", IsPublic: false, IsEditable: true, SortOrder: 23},
	{Key: "realname_api_app_key", Value: "", Type: "string", Category: "security", Label: "实名API AppKey", Description: "第三方实名服务 AccessKey / AppKey", IsPublic: false, IsEditable: true, SortOrder: 24},
	{Key: "realname_api_app_secret", Value: "", Type: "string", Category: "security", Label: "实名API AppSecret", Description: "第三方实名服务 Secret", IsPublic: false, IsEditable: true, SortOrder: 25},
	{Key: "realname_api_endpoint", Value: "", Type: "string", Category: "security", Label: "实名API Endpoint", Description: "自定义实名接口地址；官方服务商可留空", IsPublic: false, IsEditable: true, SortOrder: 26},
	{Key: "online_report_interval_seconds", Value: "30", Type: "number", Category: "security", Label: "在线心跳上报周期", Description: "客户端每隔多少秒上报一次在线心跳，默认30秒；判定离线的容忍窗口按此值的3倍换算", IsPublic: true, IsEditable: true, SortOrder: 27},
}

// initDefaultSettings 初始化默认配置
func initDefaultSettings() {
	for _, setting := range defaultSettings {
		var existing SystemSetting
		err := db.DB.Where("setting_key = ?", setting.Key).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row := setting
			if err := db.DB.Create(&row).Error; err != nil {
				log.Printf("[Init] Failed to insert default setting %s: %v", setting.Key, err)
			} else {
				log.Printf("[Init] Inserted default setting: %s", setting.Key)
			}
			continue
		}
		if err != nil {
			log.Printf("[Init] Error checking setting %s: %v", setting.Key, err)
			continue
		}
		if existing.Type != setting.Type || existing.Category != setting.Category || existing.Label != setting.Label ||
			existing.Description != setting.Description || existing.IsPublic != setting.IsPublic ||
			existing.IsEditable != setting.IsEditable || existing.SortOrder != setting.SortOrder {
			if err := db.DB.Model(&SystemSetting{}).Where("setting_key = ?", setting.Key).Updates(map[string]any{
				"setting_type": setting.Type,
				"category":     setting.Category,
				"label":        setting.Label,
				"description":  setting.Description,
				"is_public":    setting.IsPublic,
				"is_editable":  setting.IsEditable,
				"sort_order":   setting.SortOrder,
			}).Error; err != nil {
				log.Printf("[Init] Failed to sync default setting meta %s: %v", setting.Key, err)
			}
		}
	}
}

// GetSettingByKey 根据键名获取配置
func GetSettingByKey(key string) (*SystemSetting, error) {
	var setting SystemSetting
	err := db.DB.Where("setting_key = ?", key).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetSettingsByCategory 根据分类获取配置列表
func GetSettingsByCategory(category string) ([]SystemSetting, error) {
	var settings []SystemSetting
	err := db.DB.Where("category = ?", category).Order("sort_order").Find(&settings).Error
	return settings, err
}

// GetAllSettings 获取所有配置
func GetAllSettings() ([]SystemSetting, error) {
	var settings []SystemSetting
	err := db.DB.Order("category, sort_order").Find(&settings).Error
	return settings, err
}

// GetPublicSettings 获取所有公开配置（前端可访问）
func GetPublicSettings() ([]SystemSetting, error) {
	var settings []SystemSetting
	err := db.DB.Where("is_public = ?", true).Order("category, sort_order").Find(&settings).Error
	if err != nil {
		return nil, err
	}
	filtered := make([]SystemSetting, 0, len(settings))
	for _, s := range settings {
		if isPublicSettingKeyForbidden(s.Key) {
			continue
		}
		filtered = append(filtered, s)
	}
	return filtered, nil
}

// isPublicSettingKeyForbidden 公开接口层禁止下发的配置 key（关键词匹配）
func isPublicSettingKeyForbidden(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return true
	}
	forbiddenExact := map[string]struct{}{
		"geetest_captcha_key": {},
		"smtp_password":       {},
		"smtp_proxy_password": {},
		"sms_access_key":      {},
		"sms_secret_key":      {},
	}
	if _, ok := forbiddenExact[key]; ok {
		return true
	}
	for _, frag := range []string{"password", "passwd", "secret", "private_key", "api_key", "apikey", "access_key", "token", "credential"} {
		if strings.Contains(key, frag) {
			return true
		}
	}
	return false
}

// UpdateSetting 更新配置值
func UpdateSetting(key string, value string) error {
	return db.DB.Model(&SystemSetting{}).Where("setting_key = ?", key).Update("setting_value", value).Error
}

// UpdateSettingWithMeta 更新配置值和元数据
func UpdateSettingWithMeta(setting *SystemSetting) error {
	return db.DB.Model(&SystemSetting{}).Where("setting_key = ?", setting.Key).Updates(map[string]any{
		"setting_value": setting.Value,
		"setting_type":  setting.Type,
		"category":      setting.Category,
		"label":         setting.Label,
		"description":   setting.Description,
		"is_public":     setting.IsPublic,
		"is_editable":   setting.IsEditable,
		"sort_order":    setting.SortOrder,
	}).Error
}

// CreateSetting 创建新配置
func CreateSetting(setting *SystemSetting) error {
	return db.DB.Create(setting).Error
}

// DeleteSetting 删除配置
func DeleteSetting(key string) error {
	return db.DB.Where("setting_key = ?", key).Delete(&SystemSetting{}).Error
}

// BatchUpdateSettings 批量更新配置
func BatchUpdateSettings(settings map[string]string) error {
	return db.WithTx(func(tx *gorm.DB) error {
		for key, value := range settings {
			if err := tx.Model(&SystemSetting{}).Where("setting_key = ?", key).Update("setting_value", value).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetSettingsMap 获取配置的键值对map
func GetSettingsMap(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result, nil
	}

	var rows []SystemSetting
	if err := db.DB.Select("setting_key", "setting_value").Where("setting_key IN ?", keys).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Key] = row.Value
	}
	return result, nil
}
