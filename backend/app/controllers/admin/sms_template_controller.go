package admin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fst/backend/app/models"
	sms_plugin "fst/backend/app/plugins/sms"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
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
		val := utils.InterfaceToString(v)
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

// SMSSendTestRequest 短信发送测试请求
type SMSSendTestRequest struct {
	Phone string `json:"phone" binding:"required"`
	Lang  string `json:"lang"`
}

// SendTest 短信发送测试（对齐邮件 email-send-test）
// 无真实云配置时可走 console/已配置的 sms 插件；失败返回清晰错误，便于后续接真云。
// @Summary 短信发送测试
// @Tags Admin-短信模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SMSSendTestRequest true "测试手机号"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/sms-send-test [post]
func (ctrl *SMSTemplateController) SendTest(c *gin.Context) {
	var req SMSSendTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		utils.Fail(c, 400, "手机号不能为空")
		return
	}
	normalized, err := utils.NormalizeAndValidateMobile(phone, services.GetGlobalMobileCNOnly())
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	if services.GlobalSMSService == nil {
		utils.Fail(c, 500, "短信服务未初始化")
		return
	}
	providerName := services.GlobalSMSService.GetProviderName()
	if providerName == "none" {
		utils.Fail(c, 500, "未配置短信服务商")
		return
	}
	// 生产环境不允许仅靠 console；本地开发允许 console 打日志验证链路
	if providerName != "console" && !services.GlobalSMSService.IsConfigured() {
		utils.Fail(c, 500, "短信服务商未完成配置（请检查 AccessKey / 签名 / 模板 Code 等）")
		return
	}
	if providerName == "console" && !config.IsProductionMode() {
		// 本地 console 可测
	} else if providerName == "console" {
		utils.Fail(c, 500, "生产环境不可使用 console 短信通道，请配置云厂商或 custom")
		return
	}

	lang := strings.TrimSpace(req.Lang)
	if lang == "" {
		lang = "zh-CN"
	}
	testCode := "888888"
	params := map[string]string{
		"__template_name":  "bind_phone",
		"__template_order": "code,expire",
	}
	if err := services.GlobalSMSService.SendCode(normalized, testCode, 10, params, lang); err != nil {
		utils.Fail(c, 500, fmt.Sprintf("发送测试短信失败: %v", err))
		return
	}

	utils.Success(c, gin.H{
		"message":  "测试短信已发送（或已写入 console 日志）",
		"provider": providerName,
		"phone":    models.MaskPhone(normalized),
	})
}

func parseSMSTemplateID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}
