package services

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/internal/config"
	"fst/backend/utils"
	"strings"
	"time"
)

// EmailService 邮件服务
type EmailService struct{}

// NewEmailService 创建邮件服务实例
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendResult 发送结果
type SendResult struct {
	Success bool
	Error   error
}

// SendEmail 发送简单邮件
func (s *EmailService) SendEmail(to, subject, body string) error {
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

	go models.CreateEmailLog(to, subject, body, "", status, error_msg)

	return err
}

// SendTemplateEmail 发送模板邮件
func (s *EmailService) SendTemplateEmail(to, template_name, lang string, vars map[string]string) error {
	// 获取模板
	tpl, err := models.GetEmailTemplate(template_name, lang)
	if err != nil {
		return fmt.Errorf("模板不存在: %s (%s)", template_name, lang)
	}

	// 合并默认变量
	all_vars := s.buildDefaultVars(vars)

	// 渲染主题
	subject := s.renderTemplate(tpl.Subject, all_vars)

	// 渲染内容
	content := s.renderTemplate(tpl.Content, all_vars)

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

	go models.CreateEmailLog(to, subject, content, template_name, status, error_msg)

	return send_err
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(to, code, lang string, expire_minutes int) error {
	// 默认中文
	if lang == "" {
		lang = "zh-CN"
	}

	vars := map[string]string{
		"code":            code,
		"expire_minutes":  fmt.Sprintf("%d", expire_minutes),
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
	go func() {
		err := s.SendEmail(to, subject, body)
		if callback != nil {
			callback(SendResult{
				Success: err == nil,
				Error:   err,
			})
		}
	}()
}

// SendTemplateEmailAsync 异步发送模板邮件
func (s *EmailService) SendTemplateEmailAsync(to, template_name, lang string, vars map[string]string, callback func(SendResult)) {
	go func() {
		err := s.SendTemplateEmail(to, template_name, lang, vars)
		if callback != nil {
			callback(SendResult{
				Success: err == nil,
				Error:   err,
			})
		}
	}()
}

// buildDefaultVars 构建默认变量
func (s *EmailService) buildDefaultVars(vars map[string]string) map[string]string {
	cfg := config.GlobalConfig

	result := map[string]string{
		"app_name": cfg.AppName,
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

// WrapHTMLLayout 将邮件内容包装在精美的 HTML 布局中
func (s *EmailService) WrapHTMLLayout(subject, content string) string {
	cfg := config.GlobalConfig
	appName := cfg.AppName
	if appName == "" {
		appName = "System"
	}

	year := fmt.Sprintf("%d", time.Now().Year())

	return `<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + subject + `</title>
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
            <h1 style="margin:0;font-size:26px;font-weight:700;color:#ffffff;letter-spacing:1px;">` + appName + `</h1>
          </td>
        </tr>
        <!-- Subject -->
        <tr>
          <td style="padding:32px 40px 0 40px;">
            <h2 style="margin:0 0 8px 0;font-size:20px;font-weight:600;color:#1a1a2e;">` + subject + `</h2>
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
            <p style="margin:0;font-size:12px;color:#a0a0b8;">&copy; ` + year + ` ` + appName + ` · All rights reserved</p>
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>`
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

	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP主机未配置")
	}
	if cfg.SMTPUser == "" {
		return fmt.Errorf("SMTP用户名未配置")
	}
	if cfg.SMTPPass == "" {
		return fmt.Errorf("SMTP密码未配置")
	}

	return nil
}

// IsEmailConfigured 检查邮件是否已配置
func (s *EmailService) IsEmailConfigured() bool {
	return s.ValidateEmailConfig() == nil
}
