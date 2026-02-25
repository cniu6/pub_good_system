package services

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/internal/config"
	"fst/backend/utils"
	"strings"
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

	// 发送邮件
	msg := utils.EmailMessage{
		To:      to,
		Subject: subject,
		Body:    content,
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
