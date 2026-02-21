package models

import (
	"fst/backend/internal/db"
	"time"
)

// EmailLog 邮件日志
type EmailLog struct {
	ID           uint64    `db:"id" json:"id"`
	ToEmail      string    `db:"to_email" json:"to_email"`
	Subject      string    `db:"subject" json:"subject"`
	Content      string    `db:"content" json:"content"`
	TemplateName string    `db:"template_name" json:"template_name"`
	Status       uint8     `db:"status" json:"status"`
	ErrorMsg     string    `db:"error_msg" json:"error_msg"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID          uint64 `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Lang        string `db:"lang" json:"lang"`
	Title       string `db:"title" json:"title"`
	Subject     string `db:"subject" json:"subject"`
	Content     string `db:"content" json:"content"`
	Description string `db:"description" json:"description"`
	Variables   string `db:"variables" json:"variables"`
	Status      uint8  `db:"status" json:"status"` // 1=启用, 0=禁用
	CreatedAt   string `db:"created_at" json:"created_at"`
	UpdatedAt   string `db:"updated_at" json:"updated_at"`
}

// CreateEmailLog 记录邮件发送日志
func CreateEmailLog(to, subject, content, tplName string, status int, errorMsg string) error {
	query := `INSERT INTO email_logs (to_email, subject, content, template_name, status, error_msg) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, to, subject, content, tplName, status, errorMsg)
	return err
}

// CreateEmailTemplate 创建邮件模板
func CreateEmailTemplate(tpl *EmailTemplate) error {
	query := `INSERT INTO email_templates (name, lang, title, subject, content, description, variables, status) 
	          VALUES (:name, :lang, :title, :subject, :content, :description, :variables, :status)`
	_, err := db.DB.NamedExec(query, tpl)
	return err
}

// CheckTemplateExists 检查模板是否存在
func CheckTemplateExists(name, lang string) bool {
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM email_templates WHERE name = ? AND lang = ?", name, lang)
	return err == nil && count > 0
}

// GetEmailTemplate 获取指定模板
func GetEmailTemplate(name, lang string) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := db.DB.Get(&tpl, "SELECT * FROM email_templates WHERE name = ? AND lang = ? AND status = 1", name, lang)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateEmailTemplateContent 更新模板内容
func UpdateEmailTemplateContent(name, lang, content string) error {
	query := `UPDATE email_templates SET content = ? WHERE name = ? AND lang = ?`
	_, err := db.DB.Exec(query, content, name, lang)
	return err
}

// InitEmailTemplates 初始化默认邮件模板
func InitEmailTemplates() {
	// 注册验证码模板
	if !CheckTemplateExists("register_code", "zh-CN") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "register_code",
			Lang:        "zh-CN",
			Title:       "注册验证码",
			Subject:     "【{app_name}】注册验证码",
			Content:     "<p>您的验证码是：<b>{code}</b></p><p>有效期{expire_minutes}分钟，请勿泄露。</p>",
			Description: "用户注册时发送的验证码",
			Variables:   "code, app_name",
			Status:      1,
		})
	} else {
		// 强制更新已有模板为动态分钟
		_ = UpdateEmailTemplateContent("register_code", "zh-CN", "<p>您的验证码是：<b>{code}</b></p><p>有效期{expire_minutes}分钟，请勿泄露。</p>")
	}
	if !CheckTemplateExists("register_code", "en-US") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "register_code",
			Lang:        "en-US",
			Title:       "Registration Code",
			Subject:     "[{app_name}] Registration Code",
			Content:     "<p>Your verification code is: <b>{code}</b></p><p>Valid for {expire_minutes} minutes. Do not share.</p>",
			Description: "Verification code for user registration",
			Variables:   "code, app_name",
			Status:      1,
		})
	} else {
		// 强制更新已有模板为动态分钟
		_ = UpdateEmailTemplateContent("register_code", "en-US", "<p>Your verification code is: <b>{code}</b></p><p>Valid for {expire_minutes} minutes. Do not share.</p>")
	}

	// 密码重置模板
	if !CheckTemplateExists("reset_password", "zh-CN") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "reset_password",
			Lang:        "zh-CN",
			Title:       "密码重置",
			Subject:     "【{app_name}】密码重置请求",
			Content:     "<p>请点击以下链接重置密码：<a href=\"{link}\">重置密码</a></p><p>或者使用验证码：<b>{code}</b></p><p>有效期15分钟。</p>",
			Description: "用户重置密码时发送的链接和验证码",
			Variables:   "link, code, app_name",
			Status:      1,
		})
	}
	if !CheckTemplateExists("reset_password", "en-US") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "reset_password",
			Lang:        "en-US",
			Title:       "Password Reset",
			Subject:     "[{app_name}] Password Reset Request",
			Content:     "<p>Click here to reset password: <a href=\"{link}\">Reset Password</a></p><p>Or use code: <b>{code}</b></p><p>Valid for 15 minutes.</p>",
			Description: "Link and code for password reset",
			Variables:   "link, code, app_name",
			Status:      1,
		})
	}
}
