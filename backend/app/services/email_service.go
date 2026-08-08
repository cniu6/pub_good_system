package services

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"fst/backend/pkg/panicsafe"
	"fst/backend/utils"
	"log"
	"strings"
	"sync"
	"time"
)

// EmailService 邮件服务
type EmailService struct{}

var (
	emailServiceOnce sync.Once
	emailServiceInst *EmailService
)

// NewEmailService 创建邮件服务实例（测试或显式新建可用）
func NewEmailService() *EmailService {
	return &EmailService{}
}

// GetEmailService 返回包级单例，避免控制器重复 new
func GetEmailService() *EmailService {
	emailServiceOnce.Do(func() {
		emailServiceInst = NewEmailService()
	})
	return emailServiceInst
}

// SendResult 发送结果
type SendResult struct {
	Success bool
	Error   error
}

// SendEmail 发送简单邮件（匿名，user_id=0）
func (s *EmailService) SendEmail(to, subject, body string) error {
	return s.SendEmailWithUser(0, to, subject, body)
}

// SendEmailWithUser 发送简单邮件并写入关联 user_id
func (s *EmailService) SendEmailWithUser(userID uint64, to, subject, body string) error {
	msg := utils.EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	err := utils.SendEmail(msg)

	// 记录日志
	status := 1
	error_msg := ""
	if err != nil {
		status = 0
		error_msg = err.Error()
	}

	if logErr := models.CreateEmailLogWithUser(userID, to, subject, body, "", status, error_msg); logErr != nil {
		log.Printf("[Email] 记录邮件日志失败: %v", logErr)
	}

	return err
}

// sensitiveEmailTemplates 列出承载验证码/重置口令等敏感数据的模板名。
// 这些模板在写入 email_logs 时需要对 content 做脱敏，避免验证码明文落库。
var sensitiveEmailTemplates = map[string]struct{}{
	"register_code":  {},
	"reset_password": {},
	"change_email":   {},
	"change_phone":   {},
	"bind_email":     {},
	"bind_phone":     {},
}

// buildEmailLogContent 针对敏感模板返回替代文本，避免验证码等敏感信息明文落库。
func buildEmailLogContent(templateName, content string) string {
	if _, ok := sensitiveEmailTemplates[templateName]; ok {
		return "[REDACTED: sensitive template content masked]"
	}
	return content
}

// SendTemplateEmail 发送模板邮件（匿名，user_id=0）
func (s *EmailService) SendTemplateEmail(to, template_name, lang string, vars map[string]string) error {
	return s.SendTemplateEmailWithUser(0, to, template_name, lang, vars)
}

// SendTemplateEmailWithUser 发送模板邮件并写入关联 user_id（登录用户绑邮箱/改密等应传真实 ID）
func (s *EmailService) SendTemplateEmailWithUser(userID uint64, to, template_name, lang string, vars map[string]string) error {
	subject, content, err := s.RenderTemplateMail(template_name, lang, vars)
	if err != nil {
		return err
	}

	// 包装 HTML 布局
	htmlBody := s.WrapHTMLLayout(subject, content)

	// 发送邮件
	msg := utils.EmailMessage{
		To:      to,
		Subject: subject,
		Body:    htmlBody,
	}

	send_err := utils.SendEmail(msg)

	// 记录日志
	status := 1
	error_msg := ""
	if send_err != nil {
		status = 0
		error_msg = send_err.Error()
	}

	loggedContent := buildEmailLogContent(template_name, content)
	if logErr := models.CreateEmailLogWithUser(userID, to, subject, loggedContent, template_name, status, error_msg); logErr != nil {
		log.Printf("[Email] 记录模板邮件日志失败: %v", logErr)
	}

	return send_err
}

// RenderTemplateMail 仅渲染模板，不直接发送，用于老控制器复用统一模板链路。
func (s *EmailService) RenderTemplateMail(templateName, lang string, vars map[string]string) (subject string, content string, err error) {
	tpl, err := models.GetEmailTemplate(templateName, lang)
	if err != nil {
		return "", "", fmt.Errorf("Template does not exist: %s (%s)", templateName, lang)
	}

	allVars := s.buildDefaultVars(vars)
	subject = s.renderTemplate(tpl.Subject, allVars)
	content = s.renderTemplate(tpl.Content, allVars)
	return subject, content, nil
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(to, code, lang string, expire_minutes int) error {
	// 默认中文
	if lang == "" {
		lang = "zh-CN"
	}

	vars := map[string]string{
		"code":           code,
		"expire_minutes": fmt.Sprintf("%d", expire_minutes),
	}

	return s.SendTemplateEmail(to, "register_code", lang, vars)
}

// SendPasswordReset 发送密码重置邮件
func (s *EmailService) SendPasswordReset(to, link, code, lang string) error {
	// 默认中文
	if lang == "" {
		lang = "zh-CN"
	}

	vars := map[string]string{
		"link": link,
		"code": code,
	}

	return s.SendTemplateEmail(to, "reset_password", lang, vars)
}

// SendEmailAsync 异步发送邮件
func (s *EmailService) SendEmailAsync(to, subject, body string, callback func(SendResult)) {
	panicsafe.Go("EmailService.SendEmailAsync", func() {
		err := s.SendEmail(to, subject, body)
		if callback != nil {
			callback(SendResult{
				Success: err == nil,
				Error:   err,
			})
		}
	})
}

// SendTemplateEmailAsync 异步发送模板邮件
func (s *EmailService) SendTemplateEmailAsync(to, template_name, lang string, vars map[string]string, callback func(SendResult)) {
	panicsafe.Go("EmailService.SendTemplateEmailAsync", func() {
		err := s.SendTemplateEmail(to, template_name, lang, vars)
		if callback != nil {
			callback(SendResult{
				Success: err == nil,
				Error:   err,
			})
		}
	})
}

// buildDefaultVars 构建默认变量
func (s *EmailService) buildDefaultVars(vars map[string]string) map[string]string {
	cfg := config.GlobalConfig
	appName := "System"
	if cfg != nil && cfg.AppName != "" {
		appName = cfg.AppName
	}

	result := map[string]string{
		"app_name": appName,
		"app_url":  "", // 可扩展
	}

	// 合并传入的变量
	for k, v := range vars {
		result[k] = v
	}

	return result
}

// renderTemplate 渲染模板
func (s *EmailService) renderTemplate(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		placeholder := fmt.Sprintf("{%s}", k)
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}

// WrapHTMLLayout 将邮件内容包装在精美的 HTML 布局中。
// 注意：content 视为已信任的 HTML（由后台模板提供），但 subject / appName
// 会进入纯文本节点，必须做 HTML 转义以防止渲染后被解析为标签。
func (s *EmailService) WrapHTMLLayout(subject, content string) string {
	cfg := config.GlobalConfig
	appName := "System"
	if cfg != nil && cfg.AppName != "" {
		appName = cfg.AppName
	}
	if appName == "" {
		appName = "System"
	}

	year := fmt.Sprintf("%d", time.Now().Year())
	safeSubject := htmlEscapeStr(subject)
	safeAppName := htmlEscapeStr(appName)

	return `<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + safeSubject + `</title>
</head>
<body style="margin:0;padding:0;background-color:#f0f2f5;font-family:'Segoe UI','PingFang SC','Microsoft YaHei',sans-serif;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f0f2f5;padding:40px 0;">
  <tr>
    <td align="center">
      <!-- Main Card -->
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%;background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">
        <!-- Header -->
        <tr>
          <td style="background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);padding:36px 40px;text-align:center;">
            <h1 style="margin:0;font-size:26px;font-weight:700;color:#ffffff;letter-spacing:1px;">` + safeAppName + `</h1>
          </td>
        </tr>
        <!-- Subject -->
        <tr>
          <td style="padding:32px 40px 0 40px;">
            <h2 style="margin:0 0 8px 0;font-size:20px;font-weight:600;color:#1a1a2e;">` + safeSubject + `</h2>
            <div style="width:48px;height:3px;background:linear-gradient(90deg,#667eea,#764ba2);border-radius:2px;"></div>
          </td>
        </tr>
        <!-- Content -->
        <tr>
          <td style="padding:24px 40px 36px 40px;">
            <div style="font-size:15px;line-height:1.8;color:#4a4a68;">` + content + `</div>
          </td>
        </tr>
        <!-- Divider -->
        <tr>
          <td style="padding:0 40px;">
            <div style="border-top:1px solid #e8e8f0;"></div>
          </td>
        </tr>
        <!-- Footer -->
        <tr>
          <td style="padding:24px 40px 32px 40px;text-align:center;">
            <p style="margin:0 0 4px 0;font-size:12px;color:#a0a0b8;">此邮件由系统自动发送，请勿直接回复</p>
            <p style="margin:0;font-size:12px;color:#a0a0b8;">&copy; ` + year + ` ` + safeAppName + ` · All rights reserved</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>`
}

// htmlEscapeStr 对邮件布局中非 HTML 片段的纯文本做基础转义。
func htmlEscapeStr(raw string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(raw)
}

// BatchSendEmail 批量发送邮件
func (s *EmailService) BatchSendEmail(recipients []string, subject, body string) map[string]error {
	results := make(map[string]error)

	for _, to := range recipients {
		err := s.SendEmail(to, subject, body)
		results[to] = err
	}

	return results
}

// BatchSendTemplateEmail 批量发送模板邮件
func (s *EmailService) BatchSendTemplateEmail(recipients []string, template_name, lang string, vars map[string]string) map[string]error {
	results := make(map[string]error)

	for _, to := range recipients {
		err := s.SendTemplateEmail(to, template_name, lang, vars)
		results[to] = err
	}

	return results
}

// CheckTemplateExists 检查模板是否存在
func (s *EmailService) CheckTemplateExists(name, lang string) bool {
	return models.CheckTemplateExists(name, lang)
}

// GetTemplate 获取模板
func (s *EmailService) GetTemplate(name, lang string) (*models.EmailTemplate, error) {
	return models.GetEmailTemplate(name, lang)
}

// CreateTemplate 创建模板
func (s *EmailService) CreateTemplate(tpl *models.EmailTemplate) error {
	return models.CreateEmailTemplate(tpl)
}

// UpdateTemplateContent 更新模板内容
func (s *EmailService) UpdateTemplateContent(name, lang, content string) error {
	return models.UpdateEmailTemplateContent(name, lang, content)
}

// ValidateEmailConfig 验证邮件配置
func (s *EmailService) ValidateEmailConfig() error {
	cfg := config.GlobalConfig
	if cfg == nil {
		return fmt.Errorf("Email configuration not initialized")
	}

	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	if cfg.SMTPPort == "" {
		return fmt.Errorf("SMTP port not configured")
	}
	if (cfg.SMTPUser == "") != (cfg.SMTPPass == "") {
		return fmt.Errorf("SMTP username and password must be configured together or left empty together")
	}
	if cfg.SystemEmail == "" && cfg.SMTPUser == "" {
		return fmt.Errorf("Sender email not configured")
	}

	return nil
}

// IsEmailConfigured 检查邮件是否已配置
func (s *EmailService) IsEmailConfigured() bool {
	return s.ValidateEmailConfig() == nil
}
