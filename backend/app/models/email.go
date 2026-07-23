package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"fst/backend/pkg/panicsafe"
	"log"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// EmailLog 邮件日志
type EmailLog struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"column:user_id;not null;default:0;index:idx_email_logs_user_id" json:"user_id"`
	ToEmail      string    `gorm:"column:to_email;size:150;not null;index:idx_email_logs_to" json:"to_email"`
	Subject      string    `gorm:"column:subject;size:255;not null" json:"subject"`
	Content      string    `gorm:"column:content;type:text;not null" json:"content"`
	TemplateName string    `gorm:"column:template_name;size:100;not null;default:'';index:idx_email_logs_template_name" json:"template_name"`
	Status       uint8     `gorm:"column:status;not null;default:0;index:idx_email_logs_status_created,priority:1" json:"status"`
	ErrorMsg     string    `gorm:"column:error_msg;type:text" json:"error_msg"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index:idx_email_logs_created_at;index:idx_email_logs_status_created,priority:2" json:"created_at"`
}

// TableName 表名
func (EmailLog) TableName() string { return "email_logs" }

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;size:100;not null;uniqueIndex:idx_tpl_name_lang,priority:1" json:"name"`
	Lang        string `gorm:"column:lang;size:20;not null;default:'zh-CN';uniqueIndex:idx_tpl_name_lang,priority:2" json:"lang"`
	Title       string `gorm:"column:title;size:100;not null" json:"title"`
	Subject     string `gorm:"column:subject;size:255;not null" json:"subject"`
	Content     string `gorm:"column:content;type:text;not null" json:"content"`
	Description string `gorm:"column:description;size:255;not null;default:''" json:"description"`
	Variables   string `gorm:"column:variables;size:500;not null;default:''" json:"variables"`
	Status      uint8  `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   string `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 表名
func (EmailTemplate) TableName() string { return "email_templates" }

var emailLogCleanupNextAt atomic.Int64

// CreateEmailLog 记录邮件发送日志（userID 可选，匿名发送传 0）
func CreateEmailLog(to, subject, content, tplName string, status int, errorMsg string) error {
	return CreateEmailLogWithUser(0, to, subject, content, tplName, status, errorMsg)
}

// CreateEmailLogWithUser 记录带用户关联的邮件日志
func CreateEmailLogWithUser(userID uint64, to, subject, content, tplName string, status int, errorMsg string) error {
	entry := &EmailLog{
		UserID:       userID,
		ToEmail:      to,
		Subject:      subject,
		Content:      content,
		TemplateName: tplName,
		Status:       uint8(status),
		ErrorMsg:     errorMsg,
		CreatedAt:    time.Now(),
	}
	if err := db.DB.Create(entry).Error; err != nil {
		return err
	}

	panicsafe.Go("EmailLog.aggregate", func() {
		if aggErr := RecordEmailLogAggregate(entry); aggErr != nil {
			log.Printf("[EmailLog] 汇总更新失败: %v", aggErr)
		}
		scheduleEmailLogRetentionCleanup()
	})

	return nil
}

func scheduleEmailLogRetentionCleanup() {
	scheduleLogRetentionCleanupGeneric(&emailLogCleanupNextAt, "email_log", "EmailLog",
		CleanExcessEmailLogs, CleanExcessEmailLogsPerRecipient)
}

// EmailLogQuery 邮件日志查询参数
type EmailLogQuery struct {
	Page         int    `form:"page" json:"page"`
	PageSize     int    `form:"page_size" json:"page_size"`
	UserID       uint64 `form:"user_id" json:"user_id"`
	ToEmail      string `form:"to_email" json:"to_email"`
	TemplateName string `form:"template_name" json:"template_name"`
	Status       int    `form:"status" json:"status"` // -1=全部, 0=失败, 1=成功
	StartTime    string `form:"start_time" json:"start_time"`
	EndTime      string `form:"end_time" json:"end_time"`
}

func buildEmailLogQuery(q *EmailLogQuery) *gorm.DB {
	query := db.DB.Model(&EmailLog{})
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if q.ToEmail != "" {
		query = query.Where("to_email LIKE ?", "%"+q.ToEmail+"%")
	}
	if q.TemplateName != "" {
		query = query.Where("template_name = ?", q.TemplateName)
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	if q.StartTime != "" {
		query = query.Where("created_at >= ?", q.StartTime)
	}
	if q.EndTime != "" {
		query = query.Where("created_at <= ?", q.EndTime)
	}
	return query
}

// GetEmailLogList 分页查询邮件日志
func GetEmailLogList(q *EmailLogQuery) ([]EmailLog, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}

	base := buildEmailLogQuery(q)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []EmailLog
	err := base.Select("id, user_id, to_email, subject, template_name, status, error_msg, created_at").
		Order("created_at DESC, id DESC").
		Limit(q.PageSize).
		Offset((q.Page - 1) * q.PageSize).
		Find(&logs).Error
	return logs, total, err
}

// GetEmailLogByID 根据 ID 获取邮件日志详情（含 content）
func GetEmailLogByID(id uint64) (*EmailLog, error) {
	var logEntry EmailLog
	err := db.DB.Where("id = ?", id).First(&logEntry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

// DeleteEmailLogsBefore 删除指定时间之前的邮件日志
func DeleteEmailLogsBefore(before string) (int64, error) {
	result := db.DB.Where("created_at < ?", before).Delete(&EmailLog{})
	return result.RowsAffected, result.Error
}

// CleanExcessEmailLogs 清理超出全局上限的旧邮件日志（只保留最新 maxCount 条）
func CleanExcessEmailLogs(maxCount int) (int64, error) {
	return cleanExcessRowsGeneric("email_logs", "created_at", maxCount)
}

// CleanExcessEmailLogsPerRecipient 按收件邮箱清理超出上限的邮件日志（每个收件邮箱最多保留 maxPerRecipient 条）
func CleanExcessEmailLogsPerRecipient(maxPerRecipient int) (int64, error) {
	return cleanExcessRowsPerGroupGeneric[string]("email_logs", "to_email", "created_at", maxPerRecipient, "to_email != ''")
}

// GetEmailLogStats 邮件日志统计（兼容旧接口）
func GetEmailLogStats() (total int64, success int64, fail int64, err error) {
	stats, statsErr := GetEmailLogStatsDetail()
	if statsErr != nil {
		return 0, 0, 0, statsErr
	}
	return stats.TotalCount, stats.SuccessCount, stats.FailCount, nil
}

// GetEmailTemplateNames 获取所有模板名（去重），用于前端筛选
func GetEmailTemplateNames() ([]string, error) {
	var names []string
	err := db.DB.Model(&EmailLog{}).
		Distinct("template_name").
		Where("template_name != ''").
		Order("template_name").
		Pluck("template_name", &names).Error
	return names, err
}

// CreateEmailTemplate 创建邮件模板
func CreateEmailTemplate(tpl *EmailTemplate) error {
	return db.DB.Create(tpl).Error
}

// CheckTemplateExists 检查模板是否存在
func CheckTemplateExists(name, lang string) bool {
	var count int64
	err := db.DB.Model(&EmailTemplate{}).Where("name = ? AND lang = ?", name, lang).Count(&count).Error
	return err == nil && count > 0
}

// GetEmailTemplate 获取指定模板
func GetEmailTemplate(name, lang string) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := db.DB.Where("name = ? AND lang = ? AND status = 1", name, lang).First(&tpl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateEmailTemplateContent 更新模板内容
func UpdateEmailTemplateContent(name, lang, content string) error {
	return db.DB.Model(&EmailTemplate{}).Where("name = ? AND lang = ?", name, lang).Update("content", content).Error
}

// defaultEmailTemplateSeed 默认邮件模板种子（Init / Reset 共用同一份内容，不再各写各的）
type defaultEmailTemplateSeed struct {
	Name        string
	Lang        string
	Title       string
	Subject     string
	Content     string
	Description string
	Variables   string
}

// GetDefaultEmailTemplateSeeds 返回全部默认邮件模板定义。
// 注意：Controller 的 Reset 接口和这里的 Init 种子必须共用同一份数据，否则「系统默认」会有两套不一致的内容
// （历史上 Controller.Reset 曾内嵌一份旧文案，和这里的种子已经不一样，属于真实 bug，见 backend/留档.md）。
func GetDefaultEmailTemplateSeeds() []defaultEmailTemplateSeed {
	return []defaultEmailTemplateSeed{
		{
			Name: "register_code", Lang: "zh-CN",
			Title:   "注册验证码",
			Subject: "【{app_name}】注册验证码",
			Content: `<p style="margin:0 0 16px 0;">您好，感谢您的注册！请使用以下验证码完成验证：</p>` +
				`<div style="text-align:center;margin:28px 0;">` +
				`<div style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:32px;font-weight:700;letter-spacing:8px;padding:16px 40px;border-radius:12px;">{code}</div>` +
				`</div>` +
				`<p style="margin:0 0 8px 0;">⏱ 验证码有效期为 <strong>{expire_minutes} 分钟</strong>，请尽快使用。</p>` +
				`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件。请勿将验证码透露给任何人。</p>`,
			Description: "用户注册时发送的验证码",
			Variables:   "code, app_name, expire_minutes",
		},
		{
			Name: "register_code", Lang: "en-US",
			Title:   "Registration Code",
			Subject: "[{app_name}] Registration Code",
			Content: `<p style="margin:0 0 16px 0;">Hello! Thank you for signing up. Please use the following code to verify your account:</p>` +
				`<div style="text-align:center;margin:28px 0;">` +
				`<div style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:32px;font-weight:700;letter-spacing:8px;padding:16px 40px;border-radius:12px;">{code}</div>` +
				`</div>` +
				`<p style="margin:0 0 8px 0;">⏱ This code is valid for <strong>{expire_minutes} minutes</strong>.</p>` +
				`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request this, please ignore this email. Never share your code with anyone.</p>`,
			Description: "Verification code for user registration",
			Variables:   "code, app_name, expire_minutes",
		},
		{
			// 密码重置模板：验证码不放入链接，用户需在重置页面手动输入邮件正文中的验证码
			Name: "reset_password", Lang: "zh-CN",
			Title:   "密码重置",
			Subject: "【{app_name}】密码重置请求",
			Content: `<p style="margin:0 0 16px 0;">您好，我们收到了您的密码重置请求。请点击下方按钮打开重置页面，并输入下方验证码完成重置：</p>` +
				`<div style="text-align:center;margin:28px 0;">` +
				`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">打开重置密码页面</a>` +
				`</div>` +
				`<p style="margin:0 0 8px 0;">请在重置页面手动输入以下验证码（出于安全考虑，验证码不会包含在链接中）：</p>` +
				`<div style="text-align:center;margin:20px 0;">` +
				`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
				`</div>` +
				`<p style="margin:0 0 8px 0;">⏱ 有效期为 <strong>15 分钟</strong>，请尽快操作。</p>` +
				`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件，您的密码不会被更改。</p>`,
			Description: "用户重置密码时发送的链接和验证码",
			Variables:   "link, code, app_name",
		},
		{
			Name: "reset_password", Lang: "en-US",
			Title:   "Password Reset",
			Subject: "[{app_name}] Password Reset Request",
			Content: `<p style="margin:0 0 16px 0;">Hello, we received a request to reset your password. Click the button below to open the reset page, then enter the verification code below to complete the reset:</p>` +
				`<div style="text-align:center;margin:28px 0;">` +
				`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">Open Reset Page</a>` +
				`</div>` +
				`<p style="margin:0 0 8px 0;">Please enter this verification code on the reset page (for security, it is never included in the link):</p>` +
				`<div style="text-align:center;margin:20px 0;">` +
				`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
				`</div>` +
				`<p style="margin:0 0 8px 0;">⏱ Valid for <strong>15 minutes</strong>.</p>` +
				`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>`,
			Description: "Link and code for password reset",
			Variables:   "link, code, app_name",
		},
	}
}

// GetDefaultEmailTemplateByNameLang 按 name+lang 取默认内容（Reset 用）
func GetDefaultEmailTemplateByNameLang(name, lang string) (seed defaultEmailTemplateSeed, ok bool) {
	for _, s := range GetDefaultEmailTemplateSeeds() {
		if s.Name == name && s.Lang == lang {
			return s, true
		}
	}
	return defaultEmailTemplateSeed{}, false
}

// SeedEmailTemplates 种子写入默认邮件模板（已存在则跳过，不覆盖管理员在后台改过的内容）。
// 之前的实现对已存在的模板会无条件 UpdateEmailTemplateContent 覆盖回默认文案，导致管理员每次重启服务
// 后台改的邮件模板内容都会被冲掉，是真实 bug，这里改成和 SeedSMSTemplates 一致的「仅缺失时插入」。
func SeedEmailTemplates() {
	for _, s := range GetDefaultEmailTemplateSeeds() {
		if CheckTemplateExists(s.Name, s.Lang) {
			continue
		}
		if err := CreateEmailTemplate(&EmailTemplate{
			Name:        s.Name,
			Lang:        s.Lang,
			Title:       s.Title,
			Subject:     s.Subject,
			Content:     s.Content,
			Description: s.Description,
			Variables:   s.Variables,
			Status:      1,
		}); err != nil {
			log.Printf("[Init] Failed to seed email template %s/%s: %v", s.Name, s.Lang, err)
		}
	}
}

// ListAllEmailTemplates 列出全部邮件模板
func ListAllEmailTemplates() ([]EmailTemplate, error) {
	var list []EmailTemplate
	err := db.DB.Order("name, lang").Find(&list).Error
	return list, err
}

// GetEmailTemplateByID 按 ID 获取
func GetEmailTemplateByID(id uint64) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := db.DB.Where("id = ?", id).First(&tpl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateEmailTemplate 更新邮件模板可编辑字段
func UpdateEmailTemplate(id uint64, subject, content, description string, status uint8) error {
	return db.DB.Model(&EmailTemplate{}).Where("id = ?", id).Updates(map[string]any{
		"subject": subject, "content": content, "description": description, "status": status,
	}).Error
}

// ResetEmailTemplateToDefault 将指定模板重置为系统默认内容（与 SeedEmailTemplates 共用同一份种子数据）
func ResetEmailTemplateToDefault(id uint64) error {
	tpl, err := GetEmailTemplateByID(id)
	if err != nil {
		return err
	}
	seed, ok := GetDefaultEmailTemplateByNameLang(tpl.Name, tpl.Lang)
	if !ok {
		return ErrEmailTemplateNoDefault
	}
	return db.DB.Model(&EmailTemplate{}).Where("id = ?", id).Updates(map[string]any{
		"subject": seed.Subject, "content": seed.Content, "description": seed.Description,
		"variables": seed.Variables, "status": 1,
	}).Error
}

// ErrEmailTemplateNoDefault 无对应默认模板
var ErrEmailTemplateNoDefault = errors.New("no default template available")
