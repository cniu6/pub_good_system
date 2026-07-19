package admin

import (
	"errors"
	"strconv"
	"strings"

	"fst/backend/app/models"
	sms_plugin "fst/backend/app/plugins/sms"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// SMSTemplateController 短信模板管理控制器
type SMSTemplateController struct{}

// NewSMSTemplateController 创建短信模板控制器
func NewSMSTemplateController() *SMSTemplateController {
	return &SMSTemplateController{}
}

// List 获取短信模板列表
// @Summary 获取短信模板列表
// @Tags Admin-短信模板
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/sms-templates [get]
func (ctrl *SMSTemplateController) List(c *gin.Context) {
	templates, err := models.ListAllSMSTemplates()
	if err != nil {
		utils.Fail(c, 500, "Failed to fetch templates")
		return
	}
	utils.Success(c, templates)
}

// Detail 获取短信模板详情
// @Summary 获取短信模板详情
// @Tags Admin-短信模板
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/sms-templates/{id} [get]
func (ctrl *SMSTemplateController) Detail(c *gin.Context) {
	id, err := parseSMSTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	tpl, err := models.GetSMSTemplateByID(id)
	if err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}
	utils.Success(c, tpl)
}

// SMSTemplateUpdateRequest 更新短信模板请求
type SMSTemplateUpdateRequest struct {
	SignName    string `json:"sign_name"`
	Content     string `json:"content" binding:"required"`
	Description string `json:"description"`
	Status      *uint8 `json:"status"`
}

// Update 更新短信模板
// @Summary 更新短信模板
// @Tags Admin-短信模板
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body SMSTemplateUpdateRequest true "更新信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/sms-templates/{id} [put]
func (ctrl *SMSTemplateController) Update(c *gin.Context) {
	id, err := parseSMSTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	var req SMSTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.SignName = utils.Clean_XSS(req.SignName)
	req.Description = utils.Clean_XSS(req.Description)
	// Content 为纯文本短信内容，保留变量占位符，仅做 XSS 清理
	req.Content = utils.Clean_XSS(req.Content)

	existing, err := models.GetSMSTemplateByID(id)
	if err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	if err := models.UpdateSMSTemplate(id, req.SignName, req.Content, req.Description, status); err != nil {
		utils.Fail(c, 500, "Failed to update template")
		return
	}

	// 热加载内存模板，使后续发送立即生效
	sms_plugin.ReloadTemplates()
	utils.Success(c, gin.H{"message": "Template updated successfully"})
}

// SMSPreviewRequest 预览请求
type SMSPreviewRequest struct {
	Content string                 `json:"content" binding:"required"`
	Vars    map[string]interface{} `json:"vars"`
}

// Preview 预览短信模板（纯文本，替换示例变量）
// @Summary 预览短信模板
// @Tags Admin-短信模板
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body SMSPreviewRequest true "预览参数"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/sms-templates/{id}/preview [post]
func (ctrl *SMSTemplateController) Preview(c *gin.Context) {
	id, err := parseSMSTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	var req SMSPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	tpl, err := models.GetSMSTemplateByID(id)
	if err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	content := tpl.Content
	if req.Content != "" {
		content = req.Content
	}

	// 默认示例变量（可被请求覆盖）
	sample := map[string]string{
		"code":     "888888",
		"expire":   "5",
		"app_name": "TestApp",
	}
	for k, v := range req.Vars {
		val := smsPreviewToString(v)
		if strings.TrimSpace(val) != "" {
			sample[k] = val
		}
	}
	for k, v := range sample {
		placeholder := "{" + k + "}"
		content = strings.ReplaceAll(content, placeholder, v)
	}

	utils.Success(c, gin.H{
		"content":   content,
		"sign_name": tpl.SignName,
	})
}

// Reset 重置短信模板为默认
// @Summary 重置短信模板
// @Tags Admin-短信模板
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/sms-templates/{id}/reset [post]
func (ctrl *SMSTemplateController) Reset(c *gin.Context) {
	id, err := parseSMSTemplateID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid template ID")
		return
	}

	if _, err := models.GetSMSTemplateByID(id); err != nil {
		utils.Fail(c, 404, "Template not found")
		return
	}

	if err := models.ResetSMSTemplateToDefault(id); err != nil {
		if errors.Is(err, models.ErrSMSTemplateNoDefault) {
			utils.Fail(c, 400, "No default template available for this template")
			return
		}
		utils.Fail(c, 500, "Failed to reset template")
		return
	}

	sms_plugin.ReloadTemplates()
	utils.Success(c, gin.H{"message": "Template reset successfully"})
}

func parseSMSTemplateID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func smsPreviewToString(v interface{}) string {
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
