package models

import (
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
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

	go func(item *EmailLog) {
		if aggErr := RecordEmailLogAggregate(item); aggErr != nil {
			log.Printf("[EmailLog] 汇总更新失败: %v", aggErr)
		}
		scheduleEmailLogRetentionCleanup()
	}(entry)

	return nil
}

func scheduleEmailLogRetentionCleanup() {
	now := time.Now().UnixNano()
	nextAt := emailLogCleanupNextAt.Load()
	if nextAt > now {
		return
	}
	if !emailLogCleanupNextAt.CompareAndSwap(nextAt, now+int64(30*time.Second)) {
		return
	}

	cfg := loadEmailLogRetentionConfig()
	if cfg.MaxCount > 0 {
		if _, err := CleanExcessEmailLogs(cfg.MaxCount); err != nil {
			log.Printf("[EmailLog] 自动清理超限日志失败: %v", err)
		}
	}
	if cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0 {
		if _, err := CleanExcessEmailLogsPerRecipient(cfg.PerUserMaxCount); err != nil {
			log.Printf("[EmailLog] 按收件人清理超限日志失败: %v", err)
		}
	}
}

type emailLogRetentionConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

func loadEmailLogRetentionConfig() emailLogRetentionConfig {
	cfg := emailLogRetentionConfig{MaxCount: 1000, PerUserMaxCount: 1000}
	settingsMap, err := GetSettingsMap([]string{
		"email_log_max_count",
		"email_log_per_user_limit_enabled",
		"email_log_per_user_max_count",
	})
	if err != nil {
		return cfg
	}
	if v, ok := settingsMap["email_log_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.MaxCount = n
		}
	}
	if v, ok := settingsMap["email_log_per_user_limit_enabled"]; ok {
		lower := strings.ToLower(strings.TrimSpace(v))
		cfg.PerUserLimitEnabled = lower == "true" || lower == "1"
	}
	if v, ok := settingsMap["email_log_per_user_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.PerUserMaxCount = n
		}
	}
	return cfg
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

// CleanExcessEmailLogs 清理超出全局上限的旧邮件日志
func CleanExcessEmailLogs(maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM email_logs"); err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	var cutoff struct {
		ID        uint64    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}
	if err := db.DB.Get(&cutoff,
		"SELECT id, created_at FROM email_logs ORDER BY created_at DESC, id DESC LIMIT 1 OFFSET ?",
		maxCount-1,
	); err != nil {
		return 0, err
	}

	result, err := db.Exec(
		"DELETE FROM email_logs WHERE created_at < ? OR (created_at = ? AND id < ?)",
		cutoff.CreatedAt, cutoff.CreatedAt, cutoff.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanExcessEmailLogsPerRecipient 按收件邮箱清理超出上限的邮件日志
func CleanExcessEmailLogsPerRecipient(maxPerRecipient int) (int64, error) {
	if maxPerRecipient <= 0 {
		return 0, nil
	}
	var groups []struct {
		ToEmail string `db:"to_email"`
		Cnt     int64  `db:"cnt"`
	}
	if err := db.DB.Select(&groups,
		"SELECT to_email, COUNT(*) AS cnt FROM email_logs WHERE to_email != '' GROUP BY to_email HAVING COUNT(*) > ?",
		maxPerRecipient,
	); err != nil {
		return 0, err
	}

	var totalAffected int64
	for _, g := range groups {
		var cutoff struct {
			ID        uint64    `db:"id"`
			CreatedAt time.Time `db:"created_at"`
		}
		if err := db.DB.Get(&cutoff,
			"SELECT id, created_at FROM email_logs WHERE to_email = ? ORDER BY created_at DESC, id DESC LIMIT 1 OFFSET ?",
			g.ToEmail, maxPerRecipient-1,
		); err != nil {
			continue
		}
		result, err := db.Exec(
			"DELETE FROM email_logs WHERE to_email = ? AND (created_at < ? OR (created_at = ? AND id < ?))",
			g.ToEmail, cutoff.CreatedAt, cutoff.CreatedAt, cutoff.ID,
		)
		if err != nil {
			continue
		}
		if n, _ := result.RowsAffected(); n > 0 {
			totalAffected += n
		}
	}
	return totalAffected, nil
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

// InitEmailTemplates 初始化默认邮件模板
func InitEmailTemplates() {
	EnsureEmailLogsUserIDColumn()

	// 注册验证码模板
	registerCodeZH := `<p style="margin:0 0 16px 0;">您好，感谢您的注册！请使用以下验证码完成验证：</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<div style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:32px;font-weight:700;letter-spacing:8px;padding:16px 40px;border-radius:12px;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ 验证码有效期为 <strong>{expire_minutes} 分钟</strong>，请尽快使用。</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件。请勿将验证码透露给任何人。</p>`

	registerCodeEN := `<p style="margin:0 0 16px 0;">Hello! Thank you for signing up. Please use the following code to verify your account:</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<div style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:32px;font-weight:700;letter-spacing:8px;padding:16px 40px;border-radius:12px;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ This code is valid for <strong>{expire_minutes} minutes</strong>.</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request this, please ignore this email. Never share your code with anyone.</p>`

	if !CheckTemplateExists("register_code", "zh-CN") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "register_code",
			Lang:        "zh-CN",
			Title:       "注册验证码",
			Subject:     "【{app_name}】注册验证码",
			Content:     registerCodeZH,
			Description: "用户注册时发送的验证码",
			Variables:   "code, app_name, expire_minutes",
			Status:      1,
		})
	} else {
		_ = UpdateEmailTemplateContent("register_code", "zh-CN", registerCodeZH)
	}
	if !CheckTemplateExists("register_code", "en-US") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "register_code",
			Lang:        "en-US",
			Title:       "Registration Code",
			Subject:     "[{app_name}] Registration Code",
			Content:     registerCodeEN,
			Description: "Verification code for user registration",
			Variables:   "code, app_name, expire_minutes",
			Status:      1,
		})
	} else {
		_ = UpdateEmailTemplateContent("register_code", "en-US", registerCodeEN)
	}

	// 密码重置模板：验证码不放入链接，用户需在重置页面手动输入邮件正文中的验证码
	resetPasswordZH := `<p style="margin:0 0 16px 0;">您好，我们收到了您的密码重置请求。请点击下方按钮打开重置页面，并输入下方验证码完成重置：</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">打开重置密码页面</a>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">请在重置页面手动输入以下验证码（出于安全考虑，验证码不会包含在链接中）：</p>` +
		`<div style="text-align:center;margin:20px 0;">` +
		`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ 有效期为 <strong>15 分钟</strong>，请尽快操作。</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件，您的密码不会被更改。</p>`

	resetPasswordEN := `<p style="margin:0 0 16px 0;">Hello, we received a request to reset your password. Click the button below to open the reset page, then enter the verification code below to complete the reset:</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">Open Reset Page</a>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">Please enter this verification code on the reset page (for security, it is never included in the link):</p>` +
		`<div style="text-align:center;margin:20px 0;">` +
		`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ Valid for <strong>15 minutes</strong>.</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>`

	if !CheckTemplateExists("reset_password", "zh-CN") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "reset_password",
			Lang:        "zh-CN",
			Title:       "密码重置",
			Subject:     "【{app_name}】密码重置请求",
			Content:     resetPasswordZH,
			Description: "用户重置密码时发送的链接和验证码",
			Variables:   "link, code, app_name",
			Status:      1,
		})
	} else {
		_ = UpdateEmailTemplateContent("reset_password", "zh-CN", resetPasswordZH)
	}
	if !CheckTemplateExists("reset_password", "en-US") {
		CreateEmailTemplate(&EmailTemplate{
			Name:        "reset_password",
			Lang:        "en-US",
			Title:       "Password Reset",
			Subject:     "[{app_name}] Password Reset Request",
			Content:     resetPasswordEN,
			Description: "Link and code for password reset",
			Variables:   "link, code, app_name",
			Status:      1,
		})
	} else {
		_ = UpdateEmailTemplateContent("reset_password", "en-US", resetPasswordEN)
	}
}
