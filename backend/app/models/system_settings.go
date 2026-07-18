package models

import (
	"database/sql"
	"encoding/json"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"
)

// SystemSetting 系统配置项
type SystemSetting struct {
	ID          uint64    `db:"id" json:"id"`
	Key         string    `db:"setting_key" json:"key"`
	Value       string    `db:"setting_value" json:"value"`
	Type        string    `db:"setting_type" json:"type"` // string, number, boolean, json
	Category    string    `db:"category" json:"category"` // basic, security, email, custom
	Label       string    `db:"label" json:"label"`       // 显示名称
	Description string    `db:"description" json:"description"`
	IsPublic    bool      `db:"is_public" json:"is_public"` // 是否公开给前端
	IsEditable  bool      `db:"is_editable" json:"is_editable"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// TableName 返回表名
func (s *SystemSetting) TableName() string {
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
		var num float64
		if _, err := json.Marshal(s.Value); err == nil {
			json.Unmarshal([]byte(s.Value), &num)
			return num
		}
		return 0
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

// InitSystemSettingsTable 初始化系统设置表
func InitSystemSettingsTable() {
	// 检查表是否存在
	if !db.CheckTableExists("system_settings") {
		schema := `CREATE TABLE IF NOT EXISTS system_settings (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			setting_key VARCHAR(100) NOT NULL COMMENT '配置键名',
			setting_value TEXT NOT NULL COMMENT '配置值',
			setting_type VARCHAR(20) NOT NULL DEFAULT 'string' COMMENT '值类型:string,number,boolean,json',
			category VARCHAR(50) NOT NULL DEFAULT 'basic' COMMENT '分类:basic,security,email,custom',
			label VARCHAR(100) NOT NULL COMMENT '显示名称',
			description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述说明',
			is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否公开给前端:0=否,1=是',
			is_editable TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否可编辑:0=否,1=是',
			sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY idx_setting_key (setting_key),
			INDEX idx_category (category),
			INDEX idx_is_public (is_public)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';`

		_, err := db.DB.Exec(schema)
		if err != nil {
			log.Fatalf("Error creating system_settings table: %v", err)
		}
		log.Println("[Init] system_settings table created")
	}

	// 初始化默认配置
	initDefaultSettings()
}

// 默认配置项定义
var defaultSettings = []SystemSetting{
	// ===== 基本设置 =====
	{Key: "site_name", Value: "F.st", Type: "string", Category: "basic", Label: "系统名称", Description: "显示在浏览器标签和页面的系统名称", IsPublic: true, IsEditable: true, SortOrder: 1},
	{Key: "site_desc", Value: "基于 Go + Vue 3 的全栈管理系统模板", Type: "string", Category: "basic", Label: "系统描述", Description: "系统简介描述", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "site_logo", Value: "", Type: "string", Category: "basic", Label: "站点Logo", Description: "站点Logo图片URL", IsPublic: true, IsEditable: true, SortOrder: 3},
	{Key: "copyright", Value: "© 2024 F.st", Type: "string", Category: "basic", Label: "版权信息", Description: "页脚版权声明", IsPublic: true, IsEditable: true, SortOrder: 4},
	{Key: "icp", Value: "", Type: "string", Category: "basic", Label: "ICP备案号", Description: "网站ICP备案号", IsPublic: true, IsEditable: true, SortOrder: 5},
	{Key: "allow_register", Value: "true", Type: "boolean", Category: "basic", Label: "允许注册", Description: "是否允许新用户注册", IsPublic: true, IsEditable: true, SortOrder: 6},
	{Key: "default_lang", Value: "zhCN", Type: "string", Category: "basic", Label: "默认语言", Description: "系统默认语言", IsPublic: true, IsEditable: true, SortOrder: 7},
	{Key: "version", Value: "1.0.0", Type: "string", Category: "basic", Label: "系统版本", Description: "当前系统版本号", IsPublic: true, IsEditable: true, SortOrder: 8},
	{Key: "allow_delete_account", Value: "false", Type: "boolean", Category: "basic", Label: "允许注销账号", Description: "是否允许用户自助注销账号", IsPublic: true, IsEditable: true, SortOrder: 7},
	{Key: "frontend_url", Value: "", Type: "string", Category: "basic", Label: "前端地址", Description: "前端访问地址（如 http://example.com），结尾不要加 /", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "backend_api_url", Value: "", Type: "string", Category: "basic", Label: "后端API地址", Description: "后端API外网地址（如 http://api.example.com），结尾不要加 /", IsPublic: false, IsEditable: true, SortOrder: 10},

	// ===== 安全设置 =====
	{Key: "geetest_enabled", Value: "false", Type: "boolean", Category: "security", Label: "极验验证码", Description: "是否启用极验行为验证", IsPublic: true, IsEditable: true, SortOrder: 1},
	{Key: "geetest_captcha_id", Value: "", Type: "string", Category: "security", Label: "极验 Captcha ID", Description: "极验验证码 ID", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "geetest_captcha_key", Value: "", Type: "string", Category: "security", Label: "极验 Captcha Key", Description: "极验验证码 Key", IsPublic: false, IsEditable: true, SortOrder: 3},
	{Key: "jwt_access_expire", Value: "7200", Type: "number", Category: "security", Label: "Token有效期", Description: "Access Token 有效期（秒）", IsPublic: false, IsEditable: true, SortOrder: 4},
	{Key: "jwt_refresh_expire", Value: "604800", Type: "number", Category: "security", Label: "Refresh Token有效期", Description: "Refresh Token 有效期（秒）", IsPublic: false, IsEditable: true, SortOrder: 5},
	{Key: "register_code_expire_minutes", Value: "60", Type: "number", Category: "security", Label: "注册验证码有效期", Description: "注册验证码有效期（分钟）", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "login_max_failure", Value: "5", Type: "number", Category: "security", Label: "登录失败锁定次数", Description: "连续登录失败多少次后锁定账户", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "login_lock_duration", Value: "10", Type: "number", Category: "security", Label: "账户锁定时长", Description: "账户锁定时长（分钟）", IsPublic: false, IsEditable: true, SortOrder: 7},
	{Key: "operation_log_query_days", Value: "30", Type: "number", Category: "security", Label: "操作日志查询天数", Description: "操作日志默认查询范围（天）", IsPublic: false, IsEditable: true, SortOrder: 8},
	{Key: "operation_log_max_count", Value: "20", Type: "number", Category: "security", Label: "操作日志最大数量", Description: "操作日志单页最大查询数量", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "api_access_log_enabled", Value: "true", Type: "boolean", Category: "security", Label: "启用API接口日志", Description: "是否记录API接口访问日志（请求/响应内容会按规则脱敏）", IsPublic: false, IsEditable: true, SortOrder: 10},
	{Key: "api_log_query_days", Value: "7", Type: "number", Category: "security", Label: "API日志查询天数", Description: "API接口日志默认查询范围（天）", IsPublic: false, IsEditable: true, SortOrder: 11},
	{Key: "api_log_max_count", Value: "1000", Type: "number", Category: "security", Label: "API日志保留上限", Description: "API接口日志自动保留的最大条数", IsPublic: false, IsEditable: true, SortOrder: 12},
	{Key: "api_rate_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "启用全局API限流", Description: "是否对全部 /api 请求启用基础限流（开发环境默认关闭）", IsPublic: false, IsEditable: true, SortOrder: 13},
	{Key: "api_rate_limit_rate", Value: "120", Type: "number", Category: "security", Label: "全局API每秒速率", Description: "全局API限流每秒允许请求数", IsPublic: false, IsEditable: true, SortOrder: 14},
	{Key: "api_rate_limit_burst", Value: "240", Type: "number", Category: "security", Label: "全局API突发上限", Description: "全局API限流突发流量上限", IsPublic: false, IsEditable: true, SortOrder: 15},
	{Key: "admin_rate_limit_enabled", Value: "false", Type: "boolean", Category: "security", Label: "启用管理端限流", Description: "是否对管理员后台接口额外启用更严格的限流", IsPublic: false, IsEditable: true, SortOrder: 16},
	{Key: "admin_rate_limit_rate", Value: "60", Type: "number", Category: "security", Label: "管理端每秒速率", Description: "管理员后台接口每秒允许请求数", IsPublic: false, IsEditable: true, SortOrder: 17},
	{Key: "admin_rate_limit_burst", Value: "120", Type: "number", Category: "security", Label: "管理端突发上限", Description: "管理员后台接口限流突发流量上限", IsPublic: false, IsEditable: true, SortOrder: 18},

	// ===== 邮件设置 =====
	{Key: "email_verify_enabled", Value: "true", Type: "boolean", Category: "email", Label: "邮箱验证码", Description: "是否启用邮箱验证码功能（关闭后修改邮箱无需验证）", IsPublic: true, IsEditable: true, SortOrder: 0},
	{Key: "smtp_host", Value: "", Type: "string", Category: "email", Label: "SMTP服务器", Description: "SMTP邮件服务器地址", IsPublic: false, IsEditable: true, SortOrder: 1},
	{Key: "smtp_port", Value: "587", Type: "number", Category: "email", Label: "SMTP端口", Description: "SMTP服务器端口", IsPublic: false, IsEditable: true, SortOrder: 2},
	{Key: "smtp_username", Value: "", Type: "string", Category: "email", Label: "发件人邮箱", Description: "SMTP登录用户名/邮箱", IsPublic: false, IsEditable: true, SortOrder: 3},
	{Key: "smtp_password", Value: "", Type: "string", Category: "email", Label: "邮箱密码", Description: "SMTP登录密码或应用密钥", IsPublic: false, IsEditable: true, SortOrder: 4},
	{Key: "smtp_ssl", Value: "true", Type: "boolean", Category: "email", Label: "SSL加密", Description: "是否启用SSL加密", IsPublic: false, IsEditable: true, SortOrder: 5},
	{Key: "system_email_address", Value: "", Type: "string", Category: "email", Label: "系统发件邮箱", Description: "邮件头中显示的发件邮箱地址；留空时回退 SMTP 登录邮箱", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "system_email_name", Value: "F.st", Type: "string", Category: "email", Label: "发件人名称", Description: "邮件中显示的发件人名称", IsPublic: false, IsEditable: true, SortOrder: 7},

	// ===== 短信设置 =====
	{Key: "sms_verify_enabled", Value: "false", Type: "boolean", Category: "sms", Label: "短信验证码", Description: "是否启用短信验证码功能（关闭后修改手机号无需验证）", IsPublic: true, IsEditable: true, SortOrder: 0},
	{Key: "sms_provider", Value: "console", Type: "string", Category: "sms", Label: "短信服务商", Description: "短信服务商标识：console(仅控制台日志)、aliyun(阿里云 SendSms)、tencent(腾讯云 SendSms)、custom(自定义HTTP)", IsPublic: false, IsEditable: true, SortOrder: 1},
	{Key: "sms_access_key", Value: "", Type: "string", Category: "sms", Label: "AccessKey", Description: "阿里云填 AccessKeyId，腾讯云填 SecretId，自定义网关填 X-Api-Key", IsPublic: false, IsEditable: true, SortOrder: 2},
	{Key: "sms_secret_key", Value: "", Type: "string", Category: "sms", Label: "SecretKey", Description: "阿里云填 AccessKeySecret，腾讯云填 SecretKey，自定义网关作为 Authorization Bearer", IsPublic: false, IsEditable: true, SortOrder: 3},
	{Key: "sms_sign_name", Value: "", Type: "string", Category: "sms", Label: "短信签名", Description: "必须填写服务商控制台已审核通过的签名内容（腾讯云填签名内容而不是签名ID）", IsPublic: false, IsEditable: true, SortOrder: 4},
	{Key: "sms_template_code", Value: "", Type: "string", Category: "sms", Label: "默认模板ID", Description: "默认短信模板 ID / Code：阿里云对应 TemplateCode，腾讯云对应 TemplateId", IsPublic: false, IsEditable: true, SortOrder: 5},
	{Key: "sms_template_code_en", Value: "", Type: "string", Category: "sms", Label: "英文模板ID", Description: "英文短信模板 ID / Code，可选；en-US 时优先使用，未填写时回退默认模板", IsPublic: false, IsEditable: true, SortOrder: 6},
	{Key: "sms_region", Value: "", Type: "string", Category: "sms", Label: "服务区域", Description: "服务区域；留空时阿里云默认 cn-hangzhou，腾讯云默认 ap-guangzhou", IsPublic: false, IsEditable: true, SortOrder: 7},
	{Key: "sms_sdk_app_id", Value: "", Type: "string", Category: "sms", Label: "腾讯云 AppID", Description: "腾讯云短信 SmsSdkAppId，仅 tencent Provider 必填", IsPublic: false, IsEditable: true, SortOrder: 8},
	{Key: "sms_endpoint", Value: "", Type: "string", Category: "sms", Label: "HTTP Endpoint", Description: "自定义短信网关请求地址，仅 custom Provider 使用", IsPublic: false, IsEditable: true, SortOrder: 9},
	{Key: "sms_body_format", Value: "json", Type: "string", Category: "sms", Label: "请求体格式", Description: "自定义短信网关请求体格式：json 或 form，仅 custom Provider 使用", IsPublic: false, IsEditable: true, SortOrder: 10},

	// ===== 支付设置 =====
	{Key: "payment_enabled", Value: "false", Type: "boolean", Category: "payment", Label: "支付功能", Description: "是否启用在线支付充值功能", IsPublic: true, IsEditable: true, SortOrder: 0},
	{Key: "payment_order_expire_minutes", Value: "30", Type: "number", Category: "payment", Label: "订单有效期", Description: "订单有效期（分钟），超时自动取消", IsPublic: false, IsEditable: true, SortOrder: 1},
	{Key: "withdraw_enabled", Value: "true", Type: "boolean", Category: "payment", Label: "提现功能", Description: "是否启用用户提现申请功能", IsPublic: true, IsEditable: true, SortOrder: 2},
	{Key: "withdraw_min_amount", Value: "10", Type: "number", Category: "payment", Label: "最低提现金额", Description: "用户单次提现的最低金额", IsPublic: true, IsEditable: true, SortOrder: 3},
	{Key: "withdraw_notify_text", Value: "", Type: "string", Category: "payment", Label: "提现提示语", Description: "显示在用户提现页面的说明文案（留空则使用系统默认多语言文案）", IsPublic: true, IsEditable: true, SortOrder: 4},
	{Key: "withdraw_account_types", Value: "[\"bank\",\"alipay\",\"wechat\",\"usdt\"]", Type: "json", Category: "payment", Label: "支持收款方式", Description: "用户可选择的提现收款方式列表(JSON数组)", IsPublic: true, IsEditable: true, SortOrder: 5},

	// ===== 实名认证设置 =====
	{Key: "realname_enabled", Value: "true", Type: "boolean", Category: "security", Label: "实名认证功能", Description: "是否启用实名认证功能入口；仅控制站内实名入口，不代表已接第三方实名服务", IsPublic: true, IsEditable: true, SortOrder: 19},
	{Key: "realname_review_required", Value: "true", Type: "boolean", Category: "security", Label: "实名认证审核", Description: "是否需要管理员审核实名认证申请；当前仓库默认仍是站内人工审核流", IsPublic: false, IsEditable: true, SortOrder: 20},
	{Key: "realname_notify_text", Value: "", Type: "string", Category: "security", Label: "实名认证提示语", Description: "显示在用户实名认证页面的说明文案（可用于提示人工审核或第三方核验说明）", IsPublic: true, IsEditable: true, SortOrder: 21},
}

// initDefaultSettings 初始化默认配置
func initDefaultSettings() {
	for _, setting := range defaultSettings {
		// 检查是否已存在
		var existing SystemSetting
		err := db.DB.Get(&existing, "SELECT * FROM system_settings WHERE setting_key = ?", setting.Key)

		if err == sql.ErrNoRows {
			// 不存在，插入默认值
			_, err := db.DB.Exec(`
				INSERT INTO system_settings (setting_key, setting_value, setting_type, category, label, description, is_public, is_editable, sort_order)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				setting.Key, setting.Value, setting.Type, setting.Category, setting.Label, setting.Description, setting.IsPublic, setting.IsEditable, setting.SortOrder)
			if err != nil {
				log.Printf("[Init] Failed to insert default setting %s: %v", setting.Key, err)
			} else {
				log.Printf("[Init] Inserted default setting: %s", setting.Key)
			}
		} else if err != nil {
			log.Printf("[Init] Error checking setting %s: %v", setting.Key, err)
		} else {
			if existing.Type != setting.Type || existing.Category != setting.Category || existing.Label != setting.Label || existing.Description != setting.Description || existing.IsPublic != setting.IsPublic || existing.IsEditable != setting.IsEditable || existing.SortOrder != setting.SortOrder {
				_, err := db.DB.Exec(`
					UPDATE system_settings
					SET setting_type = ?, category = ?, label = ?, description = ?, is_public = ?, is_editable = ?, sort_order = ?, updated_at = NOW()
					WHERE setting_key = ?`,
					setting.Type, setting.Category, setting.Label, setting.Description, setting.IsPublic, setting.IsEditable, setting.SortOrder, setting.Key)
				if err != nil {
					log.Printf("[Init] Failed to sync default setting meta %s: %v", setting.Key, err)
				}
			}
		}
		// 已存在则跳过
	}
}

// GetSettingByKey 根据键名获取配置
func GetSettingByKey(key string) (*SystemSetting, error) {
	var setting SystemSetting
	err := db.DB.Get(&setting, "SELECT * FROM system_settings WHERE setting_key = ?", key)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetSettingsByCategory 根据分类获取配置列表
func GetSettingsByCategory(category string) ([]SystemSetting, error) {
	var settings []SystemSetting
	err := db.DB.Select(&settings, "SELECT * FROM system_settings WHERE category = ? ORDER BY sort_order", category)
	return settings, err
}

// GetAllSettings 获取所有配置
func GetAllSettings() ([]SystemSetting, error) {
	var settings []SystemSetting
	err := db.DB.Select(&settings, "SELECT * FROM system_settings ORDER BY category, sort_order")
	return settings, err
}

// GetPublicSettings 获取所有公开配置（前端可访问）
// 二次过滤：即使 DB 中敏感 key 被误标 is_public=1，也不返回其值
func GetPublicSettings() ([]SystemSetting, error) {
	var settings []SystemSetting
	err := db.DB.Select(&settings, "SELECT * FROM system_settings WHERE is_public = 1 ORDER BY category, sort_order")
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
	_, err := db.DB.Exec("UPDATE system_settings SET setting_value = ?, updated_at = NOW() WHERE setting_key = ?", value, key)
	return err
}

// UpdateSettingWithMeta 更新配置值和元数据
func UpdateSettingWithMeta(setting *SystemSetting) error {
	_, err := db.DB.Exec(`
		UPDATE system_settings 
		SET setting_value = ?, setting_type = ?, category = ?, label = ?, description = ?, is_public = ?, is_editable = ?, sort_order = ?, updated_at = NOW()
		WHERE setting_key = ?`,
		setting.Value, setting.Type, setting.Category, setting.Label, setting.Description, setting.IsPublic, setting.IsEditable, setting.SortOrder, setting.Key)
	return err
}

// CreateSetting 创建新配置
func CreateSetting(setting *SystemSetting) error {
	_, err := db.DB.Exec(`
		INSERT INTO system_settings (setting_key, setting_value, setting_type, category, label, description, is_public, is_editable, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		setting.Key, setting.Value, setting.Type, setting.Category, setting.Label, setting.Description, setting.IsPublic, setting.IsEditable, setting.SortOrder)
	return err
}

// DeleteSetting 删除配置
func DeleteSetting(key string) error {
	_, err := db.DB.Exec("DELETE FROM system_settings WHERE setting_key = ?", key)
	return err
}

// BatchUpdateSettings 批量更新配置
func BatchUpdateSettings(settings map[string]string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for key, value := range settings {
		_, err := tx.Exec("UPDATE system_settings SET setting_value = ?, updated_at = NOW() WHERE setting_key = ?", value, key)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSettingsMap 获取配置的键值对map
func GetSettingsMap(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result, nil
	}

	// 构建 IN 查询占位符
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, k := range keys {
		placeholders[i] = "?"
		args[i] = k
	}
	query := "SELECT setting_key, setting_value FROM system_settings WHERE setting_key IN (" + strings.Join(placeholders, ",") + ")"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, nil
}

