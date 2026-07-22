package models

import (
	"errors"
	"fst/backend/pkg/db"
	"fst/backend/pkg/panicsafe"
	"log"
	"sync/atomic"
	"time"
)

// EmailLog 邮件日志
type EmailLog struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"user_id"` // 关联用户（匿名发送为 0）
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

var emailLogCleanupNextAt atomic.Int64

// EnsureEmailLogsUserIDColumn 兼容旧表补齐 user_id
func EnsureEmailLogsUserIDColumn() {
	if !db.CheckTableExists("email_logs") {
		return
	}
	if !db.CheckColumnExists("email_logs", "user_id") {
		if _, err := db.Exec("ALTER TABLE email_logs ADD COLUMN user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联用户ID（匿名发送为0）' AFTER id"); err != nil {
			log.Printf("[Init] Failed to add email_logs.user_id: %v", err)
		} else {
			log.Println("[Init] Added email_logs.user_id")
		}
	}
	db.EnsureIndex("email_logs", "idx_email_logs_user_id", "ALTER TABLE email_logs ADD INDEX idx_email_logs_user_id (user_id)")
	InitEmailLogAggregateTables()
}

// CreateEmailLog 记录邮件发送日志（userID 可选，匿名发送传 0）
func CreateEmailLog(to, subject, content, tplName string, status int, errorMsg string) error {
	return CreateEmailLogWithUser(0, to, subject, content, tplName, status, errorMsg)
}

// CreateEmailLogWithUser 记录带用户关联的邮件日志
func CreateEmailLogWithUser(userID uint64, to, subject, content, tplName string, status int, errorMsg string) error {
	query := `INSERT INTO email_logs (user_id, to_email, subject, content, template_name, status, error_msg) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, userID, to, subject, content, tplName, status, errorMsg)
	if err != nil {
		return err
	}

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
	if id, idErr := result.LastInsertId(); idErr == nil {
		entry.ID = uint64(id)
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

// GetEmailLogList 分页查询邮件日志
func GetEmailLogList(q *EmailLogQuery) ([]EmailLog, int64, error) {
	var logs []EmailLog
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if q.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, q.UserID)
	}
	if q.ToEmail != "" {
		where += " AND to_email LIKE ?"
		args = append(args, "%"+q.ToEmail+"%")
	}
	if q.TemplateName != "" {
		where += " AND template_name = ?"
		args = append(args, q.TemplateName)
	}
	if q.Status >= 0 {
		where += " AND status = ?"
		args = append(args, q.Status)
	}
	if q.StartTime != "" {
		where += " AND created_at >= ?"
		args = append(args, q.StartTime)
	}
	if q.EndTime != "" {
		where += " AND created_at <= ?"
		args = append(args, q.EndTime)
	}

	err := db.DB.Get(&total, "SELECT COUNT(*) FROM email_logs "+where, args...)
	if err != nil {
		return nil, 0, err
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	offset := (q.Page - 1) * q.PageSize

	list_sql := "SELECT id, user_id, to_email, subject, template_name, status, error_msg, created_at FROM email_logs " +
		where + " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, q.PageSize, offset)

	err = db.DB.Select(&logs, list_sql, args...)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetEmailLogByID 根据 ID 获取邮件日志详情（含 content）
func GetEmailLogByID(id uint64) (*EmailLog, error) {
	var logEntry EmailLog
	err := db.DB.Get(&logEntry, "SELECT * FROM email_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

// DeleteEmailLogsBefore 删除指定时间之前的邮件日志
func DeleteEmailLogsBefore(before string) (int64, error) {
	result, err := db.Exec("DELETE FROM email_logs WHERE created_at < ?", before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
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
	err := db.DB.Select(&names, "SELECT DISTINCT template_name FROM email_logs WHERE template_name != '' ORDER BY template_name")
	if err != nil {
		return nil, err
	}
	return names, nil
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
	_, err := db.Exec(query, content, name, lang)
	return err
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

// InitEmailTemplates 种子写入默认邮件模板（已存在则跳过，不覆盖管理员在后台改过的内容）。
// 之前的实现对已存在的模板会无条件 UpdateEmailTemplateContent 覆盖回默认文案，导致管理员每次重启服务
// 后台改的邮件模板内容都会被冲掉，是真实 bug，这里改成和 InitSMSTemplates 一致的「仅缺失时插入」。
func InitEmailTemplates() {
	EnsureEmailLogsUserIDColumn()

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
	err := db.DB.Select(&list, "SELECT * FROM email_templates ORDER BY name, lang")
	return list, err
}

// GetEmailTemplateByID 按 ID 获取
func GetEmailTemplateByID(id uint64) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := db.DB.Get(&tpl, "SELECT * FROM email_templates WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateEmailTemplate 更新邮件模板可编辑字段
func UpdateEmailTemplate(id uint64, subject, content, description string, status uint8) error {
	_, err := db.Exec(
		`UPDATE email_templates SET subject = ?, content = ?, description = ?, status = ? WHERE id = ?`,
		subject, content, description, status, id,
	)
	return err
}

// ResetEmailTemplateToDefault 将指定模板重置为系统默认内容（与 InitEmailTemplates 共用同一份种子数据）
func ResetEmailTemplateToDefault(id uint64) error {
	tpl, err := GetEmailTemplateByID(id)
	if err != nil {
		return err
	}
	seed, ok := GetDefaultEmailTemplateByNameLang(tpl.Name, tpl.Lang)
	if !ok {
		return ErrEmailTemplateNoDefault
	}
	_, err = db.Exec(
		`UPDATE email_templates SET subject = ?, content = ?, description = ?, variables = ?, status = 1 WHERE id = ?`,
		seed.Subject, seed.Content, seed.Description, seed.Variables, id,
	)
	return err
}

// ErrEmailTemplateNoDefault 无对应默认模板
var ErrEmailTemplateNoDefault = errors.New("no default template available")
