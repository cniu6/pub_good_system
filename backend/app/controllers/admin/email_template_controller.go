package admin

import (
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"log"
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
		email_svc: services.GetEmailService(),
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
	templates, err := models.ListAllEmailTemplates()
	if err != nil {
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
	id, err := parseEmailTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	template, err := models.GetEmailTemplateByID(id)
	if err != nil {
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
	id, err := parseEmailTemplateID(c)
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
	if err := utils.ValidateRuneLen(req.Subject, "邮件主题", utils.MaxSubjectLength); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	if err := utils.ValidateRuneLen(req.Description, "描述", utils.MaxDescriptionLength); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// Content 不需要过滤，因为是HTML邮件内容

	existing, err := models.GetEmailTemplateByID(id)
	if err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	if err := models.UpdateEmailTemplate(id, req.Subject, req.Content, req.Description, status); err != nil {
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
	id, err := parseEmailTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	var req EmailPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	template, err := models.GetEmailTemplateByID(id)
	if err != nil {
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
		content = strings.ReplaceAll(content, placeholder, utils.InterfaceToString(v))
		subject = strings.ReplaceAll(subject, placeholder, utils.InterfaceToString(v))
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
	id, err := parseEmailTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	if _, err := models.GetEmailTemplateByID(id); err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	if err := models.ResetEmailTemplateToDefault(id); err != nil {
		if errors.Is(err, models.ErrEmailTemplateNoDefault) {
			utils.Fail(c, 400, "No default template available for this template")
			return
		}
		utils.Fail(c, 500, "Failed to reset template")
		return
	}

	utils.Success(c, gin.H{"message": "Template reset successfully"})
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
		tpl, err := models.GetEmailTemplateByID(req.TemplateID)
		if err != nil {
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
		log.Printf("[ADMIN][EMAIL] send test email failed to=%s: %v", req.To, err)
		// 管理端返回具体原因，便于排查端口/SSL 误配等问题
		utils.Fail(c, 500, fmt.Sprintf("发送测试邮件失败: %v", err))
		return
	}

	utils.Success(c, gin.H{"message": "测试邮件已发送"})
}

func parseEmailTemplateID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

