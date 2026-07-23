package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"log"

	"gorm.io/gorm"
)

// SMSTemplate 短信模板（本地记录/预览；云服务商侧需另行对齐）
type SMSTemplate struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;size:100;not null;uniqueIndex:idx_sms_tpl_name_lang,priority:1" json:"name"`
	Lang        string `gorm:"column:lang;size:20;not null;default:'zh-CN';uniqueIndex:idx_sms_tpl_name_lang,priority:2" json:"lang"`
	SignName    string `gorm:"column:sign_name;size:64;not null;default:''" json:"sign_name"`
	Content     string `gorm:"column:content;type:text;not null" json:"content"`
	Description string `gorm:"column:description;size:255;not null;default:''" json:"description"`
	Variables   string `gorm:"column:variables;size:500;not null;default:''" json:"variables"`
	Status      uint8  `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   string `gorm:"column:created_at;size:32" json:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at;size:32" json:"updated_at"`
}

// TableName 表名
func (SMSTemplate) TableName() string { return "sms_templates" }

// defaultSMSTemplateSeed 默认短信模板种子（与 plugins/sms/templates 内置一致）
type defaultSMSTemplateSeed struct {
	Name        string
	Lang        string
	SignName    string
	Content     string
	Description string
	Variables   string
}

// GetDefaultSMSTemplateSeeds 返回全部默认短信模板定义（Seed / Reset 共用）
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

// CheckSMSTemplateExists 检查模板是否已存在
func CheckSMSTemplateExists(name, lang string) bool {
	var count int64
	err := db.DB.Model(&SMSTemplate{}).Where("name = ? AND lang = ?", name, lang).Count(&count).Error
	return err == nil && count > 0
}

// CreateSMSTemplate 创建短信模板
func CreateSMSTemplate(tpl *SMSTemplate) error {
	return db.DB.Create(tpl).Error
}

// SeedSMSTemplates 种子写入默认短信模板（已存在则跳过，不覆盖管理员修改）
func SeedSMSTemplates() {
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
	err := db.DB.Order("name, lang").Find(&list).Error
	return list, err
}

// ListEnabledSMSTemplates 列出已启用的短信模板（供内存 Manager 加载）
func ListEnabledSMSTemplates() ([]SMSTemplate, error) {
	var list []SMSTemplate
	err := db.DB.Where("status = 1").Order("name, lang").Find(&list).Error
	return list, err
}

// GetSMSTemplateByID 按 ID 获取
func GetSMSTemplateByID(id uint64) (*SMSTemplate, error) {
	var tpl SMSTemplate
	err := db.DB.Where("id = ?", id).First(&tpl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// GetSMSTemplateByNameLang 按 name+lang 获取（仅启用；缺模板用 FindOne）
func GetSMSTemplateByNameLang(name, lang string) (*SMSTemplate, error) {
	var tpl SMSTemplate
	if err := db.FindOne(db.DB.Where("name = ? AND lang = ? AND status = 1", name, lang), &tpl); err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateSMSTemplate 更新短信模板可编辑字段
func UpdateSMSTemplate(id uint64, signName, content, description string, status uint8) error {
	return db.DB.Model(&SMSTemplate{}).Where("id = ?", id).Updates(map[string]any{
		"sign_name": signName, "content": content, "description": description, "status": status,
	}).Error
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
	return db.DB.Model(&SMSTemplate{}).Where("id = ?", id).Updates(map[string]any{
		"sign_name": signName, "content": content, "description": description,
		"variables": variables, "status": 1,
	}).Error
}

// ErrSMSTemplateNoDefault 无对应默认模板
var ErrSMSTemplateNoDefault = errors.New("no default template available")
