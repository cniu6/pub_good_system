package models

import (
	"errors"
	"fst/backend/pkg/db"
	"log"
)

// SMSTemplate 短信模板（本地记录/预览；云服务商侧需另行对齐）
type SMSTemplate struct {
	ID          uint64 `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`               // 模板标识，如 register_code
	Lang        string `db:"lang" json:"lang"`               // zh-CN / en-US
	SignName    string `db:"sign_name" json:"sign_name"`     // 短信签名
	Content     string `db:"content" json:"content"`         // 纯文本内容，支持 {code}/{expire}/{app_name}
	Description string `db:"description" json:"description"` // 描述
	Variables   string `db:"variables" json:"variables"`     // 可用变量说明
	Status      uint8  `db:"status" json:"status"`           // 1=启用, 0=禁用
	CreatedAt   string `db:"created_at" json:"created_at"`
	UpdatedAt   string `db:"updated_at" json:"updated_at"`
}

// defaultSMSTemplateSeed 默认短信模板种子（与 plugins/sms/templates 内置一致）
type defaultSMSTemplateSeed struct {
	Name        string
	Lang        string
	SignName    string
	Content     string
	Description string
	Variables   string
}

// GetDefaultSMSTemplateSeeds 返回全部默认短信模板定义（Init / Reset 共用）
func GetDefaultSMSTemplateSeeds() []defaultSMSTemplateSeed {
	return []defaultSMSTemplateSeed{
		{
			Name: "register_code", Lang: "zh-CN", SignName: "",
			Content:     "【{app_name}】您的注册验证码是 {code}，{expire}分钟内有效。如非本人操作，请忽略此短信。",
			Variables:   "code, expire, app_name",
			Description: "用户注册验证码",
		},
		{
			Name: "register_code", Lang: "en-US", SignName: "",
			Content:     "[{app_name}] Your verification code is {code}, valid for {expire} minutes. If not requested by you, please ignore.",
			Variables:   "code, expire, app_name",
			Description: "User registration verification code",
		},
		{
			Name: "login_code", Lang: "zh-CN", SignName: "",
			Content:     "【{app_name}】您正在登录，验证码是 {code}，{expire}分钟内有效。",
			Variables:   "code, expire, app_name",
			Description: "用户登录验证码",
		},
		{
			Name: "login_code", Lang: "en-US", SignName: "",
			Content:     "[{app_name}] Your login code is {code}, valid for {expire} minutes.",
			Variables:   "code, expire, app_name",
			Description: "User login verification code",
		},
		{
			Name: "reset_password", Lang: "zh-CN", SignName: "",
			Content:     "【{app_name}】您正在重置密码，验证码是 {code}，{expire}分钟内有效。如非本人操作，请忽略。",
			Variables:   "code, expire, app_name",
			Description: "密码重置验证码",
		},
		{
			Name: "reset_password", Lang: "en-US", SignName: "",
			Content:     "[{app_name}] Your password reset code is {code}, valid for {expire} minutes. Ignore if not requested.",
			Variables:   "code, expire, app_name",
			Description: "Password reset verification code",
		},
		{
			Name: "bind_phone", Lang: "zh-CN", SignName: "",
			Content:     "【{app_name}】您正在绑定手机，验证码是 {code}，{expire}分钟内有效。",
			Variables:   "code, expire, app_name",
			Description: "手机绑定验证码",
		},
		{
			Name: "bind_phone", Lang: "en-US", SignName: "",
			Content:     "[{app_name}] Your phone binding code is {code}, valid for {expire} minutes.",
			Variables:   "code, expire, app_name",
			Description: "Phone binding verification code",
		},
	}
}

// GetDefaultSMSTemplateByNameLang 按 name+lang 取默认内容（Reset 用）
func GetDefaultSMSTemplateByNameLang(name, lang string) (signName, content, description, variables string, ok bool) {
	for _, s := range GetDefaultSMSTemplateSeeds() {
		if s.Name == name && s.Lang == lang {
			return s.SignName, s.Content, s.Description, s.Variables, true
		}
	}
	return "", "", "", "", false
}

// InitSMSTemplatesTable 创建短信模板表（若不存在）
func InitSMSTemplatesTable() {
	if db.CheckTableExists("sms_templates") {
		return
	}
	schema := `CREATE TABLE IF NOT EXISTS sms_templates (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL COMMENT '模板标识',
		lang VARCHAR(20) NOT NULL DEFAULT 'zh-CN' COMMENT '语言',
		sign_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '短信签名',
		content TEXT NOT NULL COMMENT '短信内容(纯文本)',
		description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
		variables VARCHAR(500) NOT NULL DEFAULT '' COMMENT '可用变量说明',
		status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY idx_sms_tpl_name_lang (name, lang)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	if _, err := db.Exec(schema); err != nil {
		log.Printf("[Init] Failed to create sms_templates table: %v", err)
	} else {
		log.Println("[Init] Created sms_templates table")
	}
}

// CheckSMSTemplateExists 检查模板是否已存在
func CheckSMSTemplateExists(name, lang string) bool {
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM sms_templates WHERE name = ? AND lang = ?", name, lang)
	return err == nil && count > 0
}

// CreateSMSTemplate 创建短信模板
func CreateSMSTemplate(tpl *SMSTemplate) error {
	query := `INSERT INTO sms_templates (name, lang, sign_name, content, description, variables, status)
	          VALUES (:name, :lang, :sign_name, :content, :description, :variables, :status)`
	_, err := db.DB.NamedExec(query, tpl)
	return err
}

// InitSMSTemplates 种子写入默认短信模板（已存在则跳过，不覆盖管理员修改）
func InitSMSTemplates() {
	InitSMSTemplatesTable()
	for _, s := range GetDefaultSMSTemplateSeeds() {
		if CheckSMSTemplateExists(s.Name, s.Lang) {
			continue
		}
		if err := CreateSMSTemplate(&SMSTemplate{
			Name:        s.Name,
			Lang:        s.Lang,
			SignName:    s.SignName,
			Content:     s.Content,
			Description: s.Description,
			Variables:   s.Variables,
			Status:      1,
		}); err != nil {
			log.Printf("[Init] Failed to seed sms template %s/%s: %v", s.Name, s.Lang, err)
		}
	}
}

// ListAllSMSTemplates 列出全部短信模板
func ListAllSMSTemplates() ([]SMSTemplate, error) {
	var list []SMSTemplate
	err := db.DB.Select(&list, "SELECT * FROM sms_templates ORDER BY name, lang")
	return list, err
}

// ListEnabledSMSTemplates 列出已启用的短信模板（供内存 Manager 加载）
func ListEnabledSMSTemplates() ([]SMSTemplate, error) {
	var list []SMSTemplate
	err := db.DB.Select(&list, "SELECT * FROM sms_templates WHERE status = 1 ORDER BY name, lang")
	return list, err
}

// GetSMSTemplateByID 按 ID 获取
func GetSMSTemplateByID(id uint64) (*SMSTemplate, error) {
	var tpl SMSTemplate
	err := db.DB.Get(&tpl, "SELECT * FROM sms_templates WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// GetSMSTemplateByNameLang 按 name+lang 获取（仅启用）
func GetSMSTemplateByNameLang(name, lang string) (*SMSTemplate, error) {
	var tpl SMSTemplate
	err := db.DB.Get(&tpl, "SELECT * FROM sms_templates WHERE name = ? AND lang = ? AND status = 1", name, lang)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateSMSTemplate 更新短信模板可编辑字段
func UpdateSMSTemplate(id uint64, signName, content, description string, status uint8) error {
	_, err := db.Exec(
		`UPDATE sms_templates SET sign_name = ?, content = ?, description = ?, status = ? WHERE id = ?`,
		signName, content, description, status, id,
	)
	return err
}

// ResetSMSTemplateToDefault 将指定模板重置为系统默认内容
func ResetSMSTemplateToDefault(id uint64) error {
	tpl, err := GetSMSTemplateByID(id)
	if err != nil {
		return err
	}
	signName, content, description, variables, ok := GetDefaultSMSTemplateByNameLang(tpl.Name, tpl.Lang)
	if !ok {
		return ErrSMSTemplateNoDefault
	}
	_, err = db.Exec(
		`UPDATE sms_templates SET sign_name = ?, content = ?, description = ?, variables = ?, status = 1 WHERE id = ?`,
		signName, content, description, variables, id,
	)
	return err
}

// ErrSMSTemplateNoDefault 无对应默认模板
var ErrSMSTemplateNoDefault = errors.New("no default template available")
