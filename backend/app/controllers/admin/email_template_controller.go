package admin

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/db"
	"fst/backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// EmailTemplateController 邮件模板管理控制器
type EmailTemplateController struct {
	email_svc *services.EmailService
}

// NewEmailTemplateController 创建邮件模板控制器
func NewEmailTemplateController() *EmailTemplateController {
	return &EmailTemplateController{
		email_svc: services.NewEmailService(),
	}
}

// List 获取邮件模板列表
// @Summary 获取邮件模板列表
// @Description 获取所有邮件模板列表
// @Tags Admin-邮件模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-templates [get]
func (ctrl *EmailTemplateController) List(c *gin.Context) {
	// 查询所有模板
	var templates []models.EmailTemplate
	query := "SELECT * FROM email_templates ORDER BY name, lang"

	if err := db.DB.Select(&templates, query); err != nil {
		utils.Fail(c, 500, "Failed to fetch templates")
		return
	}

	utils.Success(c, templates)
}

// Detail 获取邮件模板详情
// @Summary 获取邮件模板详情
// @Description 根据ID获取邮件模板详情
// @Tags Admin-邮件模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-templates/{id} [get]
func (ctrl *EmailTemplateController) Detail(c *gin.Context) {
	id_str := c.Param("id")
	id, err := strconv.ParseUint(id_str, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	var template models.EmailTemplate
	query := "SELECT * FROM email_templates WHERE id = ?"
	if err := db.DB.Get(&template, query, id); err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	utils.Success(c, template)
}

// UpdateRequest 更新模板请求
type EmailTemplateUpdateRequest struct {
	Subject     string `json:"subject"`
	Content     string `json:"content" binding:"required"`
	Description string `json:"description"`
	Status      *uint8 `json:"status"`
}

// Update 更新邮件模板
// @Summary 更新邮件模板
// @Description 更新邮件模板内容
// @Tags Admin-邮件模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body EmailTemplateUpdateRequest true "更新信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-templates/{id} [put]
func (ctrl *EmailTemplateController) Update(c *gin.Context) {
	id_str := c.Param("id")
	id, err := strconv.ParseUint(id_str, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	var req EmailTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Subject = utils.Clean_XSS(req.Subject)
	req.Description = utils.Clean_XSS(req.Description)
	// Content 不需要过滤，因为是HTML邮件内容

	// 检查模板是否存在
	var existing models.EmailTemplate
	check_query := "SELECT * FROM email_templates WHERE id = ?"
	if err := db.DB.Get(&existing, check_query, id); err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	// 更新模板
	update_query := `UPDATE email_templates SET subject = ?, content = ?, description = ?, status = ? WHERE id = ?`
	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	if _, err := db.DB.Exec(update_query, req.Subject, req.Content, req.Description, status, id); err != nil {
		utils.Fail(c, 500, "Failed to update template")
		return
	}

	utils.Success(c, gin.H{"message": "Template updated successfully"})
}

// PreviewRequest 预览请求
type EmailPreviewRequest struct {
	Content string                 `json:"content" binding:"required"`
	Vars    map[string]interface{} `json:"vars"`
}

// Preview 预览邮件模板
// @Summary 预览邮件模板
// @Description 预览邮件模板渲染效果
// @Tags Admin-邮件模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body EmailPreviewRequest true "预览参数"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-templates/{id}/preview [post]
func (ctrl *EmailTemplateController) Preview(c *gin.Context) {
	id_str := c.Param("id")
	id, err := strconv.ParseUint(id_str, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	var req EmailPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 获取模板
	var template models.EmailTemplate
	query := "SELECT * FROM email_templates WHERE id = ?"
	if err := db.DB.Get(&template, query, id); err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	// 渲染模板内容
	content := template.Content
	if req.Content != "" {
		content = req.Content // 使用传入的内容进行预览
	}

	// 渲染主题
	subject := template.Subject

	// 替换变量
	for k, v := range req.Vars {
		placeholder := "{" + k + "}"
		content = strings.ReplaceAll(content, placeholder, toString(v))
		subject = strings.ReplaceAll(subject, placeholder, toString(v))
	}

	// 使用 HTML 布局包装预览内容
	wrapped := ctrl.email_svc.WrapHTMLLayout(subject, content)

	utils.Success(c, gin.H{
		"subject": subject,
		"content": content,
		"wrapped": wrapped,
	})
}

// Reset 重置邮件模板为默认
// @Summary 重置邮件模板
// @Description 重置邮件模板为系统默认模板
// @Tags Admin-邮件模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-templates/{id}/reset [post]
func (ctrl *EmailTemplateController) Reset(c *gin.Context) {
	id_str := c.Param("id")
	id, err := strconv.ParseUint(id_str, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	// 获取模板名称和语言
	var template models.EmailTemplate
	query := "SELECT * FROM email_templates WHERE id = ?"
	if err := db.DB.Get(&template, query, id); err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	// 默认模板内容
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

	resetPasswordZH := `<p style="margin:0 0 16px 0;">您好，我们收到了您的密码重置请求。请点击下方按钮重置密码：</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">重置密码</a>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">如果按钮无法点击，您也可以使用以下验证码：</p>` +
		`<div style="text-align:center;margin:20px 0;">` +
		`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ 有效期为 <strong>15 分钟</strong>，请尽快操作。</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">如果这不是您本人的操作，请忽略此邮件，您的密码不会被更改。</p>`

	resetPasswordEN := `<p style="margin:0 0 16px 0;">Hello, we received a request to reset your password. Click the button below to proceed:</p>` +
		`<div style="text-align:center;margin:28px 0;">` +
		`<a href="{link}" style="display:inline-block;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;padding:14px 48px;border-radius:10px;">Reset Password</a>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">If the button doesn't work, you can also use this verification code:</p>` +
		`<div style="text-align:center;margin:20px 0;">` +
		`<div style="display:inline-block;background:#f0f2f5;font-size:28px;font-weight:700;letter-spacing:6px;padding:14px 36px;border-radius:10px;color:#1a1a2e;border:2px dashed #667eea;">{code}</div>` +
		`</div>` +
		`<p style="margin:0 0 8px 0;">⏱ Valid for <strong>15 minutes</strong>.</p>` +
		`<p style="margin:0;color:#a0a0b8;font-size:13px;">If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>`

	default_templates := map[string]map[string]struct {
		Subject string
		Content string
	}{
		"register_code": {
			"zh-CN": {Subject: "【{app_name}】注册验证码", Content: registerCodeZH},
			"en-US": {Subject: "[{app_name}] Registration Code", Content: registerCodeEN},
		},
		"reset_password": {
			"zh-CN": {Subject: "【{app_name}】密码重置请求", Content: resetPasswordZH},
			"en-US": {Subject: "[{app_name}] Password Reset Request", Content: resetPasswordEN},
		},
	}

	// 获取默认模板
	if templates, ok := default_templates[template.Name]; ok {
		if default_tpl, ok := templates[template.Lang]; ok {
			update_query := "UPDATE email_templates SET subject = ?, content = ? WHERE id = ?"
			if _, err := db.DB.Exec(update_query, default_tpl.Subject, default_tpl.Content, id); err != nil {
				utils.Fail(c, 500, "Failed to reset template")
				return
			}
			utils.Success(c, gin.H{"message": "Template reset successfully"})
			return
		}
	}

	utils.Fail(c, 400, "No default template available for this template")
}

// EmailSendTestRequest 发件测试请求
type EmailSendTestRequest struct {
	To         string `json:"to" binding:"required"`
	Subject    string `json:"subject"`
	Content    string `json:"content"`
	TemplateID uint64 `json:"template_id"` // 可选：使用模板发送
}

// SendTest 发件测试
// @Summary 发件测试
// @Description 发送测试邮件，验证SMTP配置是否正常，支持选择模板发送
// @Tags Admin-邮件模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body EmailSendTestRequest true "测试邮件参数"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-send-test [post]
func (ctrl *EmailTemplateController) SendTest(c *gin.Context) {
	var req EmailSendTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	var subject, content string

	if req.TemplateID > 0 {
		// 使用模板发送
		var tpl models.EmailTemplate
		query := "SELECT * FROM email_templates WHERE id = ?"
		if err := db.DB.Get(&tpl, query, req.TemplateID); err != nil {
			utils.Fail(c, 404, "模板不存在")
			return
		}

		// 使用示例变量渲染
		subject = tpl.Subject
		content = tpl.Content

		// 替换常见变量为示例值
		example_vars := map[string]string{
			"{app_name}":       "TestApp",
			"{code}":           "888888",
			"{expire_minutes}": "15",
			"{link}":           "https://example.com/reset?token=test123",
		}
		for k, v := range example_vars {
			subject = strings.ReplaceAll(subject, k, v)
			content = strings.ReplaceAll(content, k, v)
		}

		// 包装 HTML 布局
		content = ctrl.email_svc.WrapHTMLLayout(subject, content)
	} else {
		// 自定义内容发送
		subject = req.Subject
		if subject == "" {
			subject = "发件测试 / Email Test"
		}
		content = req.Content
		if content == "" {
			content = "<p>这是一封测试邮件，用于验证您的邮件发送配置是否正常。</p><p>This is a test email to verify your email configuration.</p>"
		}
	}

	err := ctrl.email_svc.SendEmail(req.To, subject, content)
	if err != nil {
		utils.Fail(c, 500, "发送失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "测试邮件已发送"})
}

// 辅助函数
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}
