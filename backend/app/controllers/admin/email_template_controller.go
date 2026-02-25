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
	update_query := `UPDATE email_templates SET subject = ?, content = ?, description = ?, status = ?, updated_at = NOW() WHERE id = ?`
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

	// 替换变量
	for k, v := range req.Vars {
		placeholder := "{" + k + "}"
		content = strings.ReplaceAll(content, placeholder, toString(v))
	}

	utils.Success(c, gin.H{
		"subject": template.Subject,
		"content": content,
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
	default_templates := map[string]map[string]struct {
		Subject string
		Content string
	}{
		"register_code": {
			"zh-CN": {
				Subject: "【{app_name}】注册验证码",
				Content: "<p>您的验证码是：<b>{code}</b></p><p>有效期{expire_minutes}分钟，请勿泄露。</p>",
			},
			"en-US": {
				Subject: "[{app_name}] Registration Code",
				Content: "<p>Your verification code is: <b>{code}</b></p><p>Valid for {expire_minutes} minutes. Do not share.</p>",
			},
		},
		"reset_password": {
			"zh-CN": {
				Subject: "【{app_name}】密码重置请求",
				Content: "<p>请点击以下链接重置密码：<a href=\"{link}\">重置密码</a></p><p>或者使用验证码：<b>{code}</b></p><p>有效期15分钟。</p>",
			},
			"en-US": {
				Subject: "[{app_name}] Password Reset Request",
				Content: "<p>Click here to reset password: <a href=\"{link}\">Reset Password</a></p><p>Or use code: <b>{code}</b></p><p>Valid for 15 minutes.</p>",
			},
		},
	}

	// 获取默认模板
	if templates, ok := default_templates[template.Name]; ok {
		if default_tpl, ok := templates[template.Lang]; ok {
			update_query := "UPDATE email_templates SET subject = ?, content = ?, updated_at = NOW() WHERE id = ?"
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
