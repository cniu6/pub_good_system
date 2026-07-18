package admin

import (
	"fst/backend/app/models"
	"fst/backend/utils"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SettingsController 管理端配置控制器
type SettingsController struct{}

// NewSettingsController 创建配置控制器
func NewSettingsController() *SettingsController {
	return &SettingsController{}
}

// ========================================
// 请求结构体
// ========================================

// UpdateSettingRequest 更新单个配置请求
type UpdateSettingRequest struct {
	Value string `json:"value"`
}

// BatchUpdateSettingsRequest 批量更新配置请求
type BatchUpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// CreateSettingRequest 创建新配置请求
type CreateSettingRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	Type        string `json:"type"`     // string, number, boolean, json
	Category    string `json:"category"` // basic, security, email, custom
	Label       string `json:"label" binding:"required"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"`
	IsEditable  *bool  `json:"is_editable"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateSettingMetaRequest 更新配置元数据请求
type UpdateSettingMetaRequest struct {
	Value       *string `json:"value"`
	Type        *string `json:"type"`
	Category    *string `json:"category"`
	Label       *string `json:"label"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
	IsEditable  *bool   `json:"is_editable"`
	SortOrder   *int    `json:"sort_order"`
}

// SettingsListResponse 配置列表响应
type SettingsListResponse struct {
	Categories []models.SettingsGroup `json:"categories"`
}

// CategoryLabelMap 分类名称映射
var CategoryLabelMap = map[string]string{
	"basic":    "基本设置",
	"security": "安全设置",
	"email":    "邮件设置",
	"payment":  "支付设置",
	"sms":      "短信设置",
	"custom":   "自定义配置",
}

// ========================================
// 控制器方法
// ========================================

// List 获取所有配置
// @Summary 获取所有系统配置
// @Description 获取所有系统配置，按分类分组
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=SettingsListResponse}
// @Router /api/v1/admin/settings [get]
func (ctrl *SettingsController) List(c *gin.Context) {
	settings, err := models.GetAllSettings()
	if err != nil {
		utils.Fail(c, 500, "Failed to load settings")
		return
	}

	// 按分类分组
	categoryMap := make(map[string][]models.SettingDTO)
	for _, s := range settings {
		dto := models.SettingDTO{
			Key:         s.Key,
			Value:       ctrl.resolveSettingValueForAdmin(s),
			Type:        s.Type,
			Category:    s.Category,
			Label:       s.Label,
			Description: s.Description,
			IsPublic:    s.IsPublic,
			IsEditable:  s.IsEditable,
		}
		categoryMap[s.Category] = append(categoryMap[s.Category], dto)
	}

	// 构建响应
	var categories []models.SettingsGroup
	for cat, items := range categoryMap {
		label, ok := CategoryLabelMap[cat]
		if !ok {
			label = cat
		}
		categories = append(categories, models.SettingsGroup{
			Category: cat,
			Label:    label,
			Items:    items,
		})
	}

	utils.Success(c, SettingsListResponse{Categories: categories})
}

// GetByCategory 获取指定分类的配置
// @Summary 获取指定分类的配置
// @Description 获取指定分类下的所有配置项
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category path string true "分类名称"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings/category/{category} [get]
func (ctrl *SettingsController) GetByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		utils.Fail(c, 400, "Category is required")
		return
	}

	settings, err := models.GetSettingsByCategory(category)
	if err != nil {
		utils.Fail(c, 500, "Failed to load settings")
		return
	}

	var items []models.SettingDTO
	for _, s := range settings {
		items = append(items, models.SettingDTO{
			Key:         s.Key,
			Value:       ctrl.resolveSettingValueForAdmin(s),
			Type:        s.Type,
			Category:    s.Category,
			Label:       s.Label,
			Description: s.Description,
			IsPublic:    s.IsPublic,
			IsEditable:  s.IsEditable,
		})
	}

	utils.Success(c, items)
}

// Get 获取单个配置
// @Summary 获取单个配置
// @Description 根据键名获取配置详情
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键名"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings/{key} [get]
func (ctrl *SettingsController) Get(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.Fail(c, 400, "Key is required")
		return
	}

	setting, err := models.GetSettingByKey(key)
	if err != nil {
		utils.Fail(c, 404, "Setting not found")
		return
	}

	dto := models.SettingDTO{
		Key:         setting.Key,
		Value:       ctrl.resolveSettingValueForAdmin(*setting),
		Type:        setting.Type,
		Category:    setting.Category,
		Label:       setting.Label,
		Description: setting.Description,
		IsPublic:    setting.IsPublic,
		IsEditable:  setting.IsEditable,
	}

	utils.Success(c, dto)
}

// Update 更新单个配置值
// @Summary 更新单个配置值
// @Description 更新指定配置项的值
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键名"
// @Param request body UpdateSettingRequest true "配置值"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings/{key} [put]
func (ctrl *SettingsController) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.Fail(c, 400, "Key is required")
		return
	}

	// 检查配置是否存在
	setting, err := models.GetSettingByKey(key)
	if err != nil {
		utils.Fail(c, 404, "Setting not found")
		return
	}

	// 检查是否可编辑
	if !setting.IsEditable {
		utils.Fail(c, 403, "This setting is not editable")
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 类型校验
	resolvedValue := ctrl.normalizeSettingValueForWrite(*setting, req.Value)
	if !ctrl.validateSettingValue(resolvedValue, setting.Type) {
		utils.Fail(c, 400, "Invalid value type for "+setting.Type)
		return
	}

	// 更新配置
	if err := models.UpdateSetting(key, resolvedValue); err != nil {
		utils.Fail(c, 500, "Failed to update setting")
		return
	}

	ctrl.refreshRuntimeConfig()

	utils.Success(c, gin.H{"message": "Setting updated successfully"})
}

// UpdateMeta 更新配置元数据
// @Summary 更新配置元数据
// @Description 更新配置项的完整信息（包括值和元数据）
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键名"
// @Param request body UpdateSettingMetaRequest true "配置信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings/{key}/meta [put]
func (ctrl *SettingsController) UpdateMeta(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.Fail(c, 400, "Key is required")
		return
	}

	// 检查配置是否存在
	existingSetting, err := models.GetSettingByKey(key)
	if err != nil {
		utils.Fail(c, 404, "Setting not found")
		return
	}

	var req UpdateSettingMetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	effectiveType := existingSetting.Type
	if req.Type != nil {
		candidateType := strings.TrimSpace(*req.Type)
		if candidateType == "" || !ctrl.isValidType(candidateType) {
			utils.Fail(c, 400, "Invalid type. Must be one of: string, number, boolean, json")
			return
		}
		effectiveType = candidateType
	}

	effectiveCategory := existingSetting.Category
	if req.Category != nil {
		candidateCategory := strings.TrimSpace(*req.Category)
		if candidateCategory == "" || !ctrl.isValidCategory(candidateCategory) {
			utils.Fail(c, 400, "Invalid category. Must be one of: basic, security, email, payment, sms, custom")
			return
		}
		effectiveCategory = candidateCategory
	}

	effectiveValue := existingSetting.Value
	if req.Value != nil {
		effectiveValue = ctrl.normalizeSettingValueForWrite(*existingSetting, *req.Value)
	}
	if !ctrl.validateSettingValue(effectiveValue, effectiveType) {
		utils.Fail(c, 400, "Invalid value type for "+effectiveType)
		return
	}

	effectiveLabel := existingSetting.Label
	if req.Label != nil {
		effectiveLabel = utils.Clean_XSS(*req.Label)
		if strings.TrimSpace(effectiveLabel) == "" {
			utils.Fail(c, 400, "Label is required")
			return
		}
	}

	effectiveDescription := existingSetting.Description
	if req.Description != nil {
		effectiveDescription = utils.Clean_XSS(*req.Description)
	}

	effectiveIsPublic := existingSetting.IsPublic
	if req.IsPublic != nil {
		effectiveIsPublic = *req.IsPublic
	}
	// 敏感配置禁止公开：避免 smtp_password / sms_secret_key 等经 /public/settings 外泄
	if effectiveIsPublic && isSensitiveSettingKey(key) {
		utils.Fail(c, 400, "敏感配置不允许设为公开")
		return
	}

	effectiveIsEditable := existingSetting.IsEditable
	if req.IsEditable != nil {
		effectiveIsEditable = *req.IsEditable
	}

	effectiveSortOrder := existingSetting.SortOrder
	if req.SortOrder != nil {
		effectiveSortOrder = *req.SortOrder
	}

	setting := &models.SystemSetting{
		Key:         key,
		Value:       effectiveValue,
		Type:        effectiveType,
		Category:    effectiveCategory,
		Label:       effectiveLabel,
		Description: effectiveDescription,
		IsPublic:    effectiveIsPublic,
		IsEditable:  effectiveIsEditable,
		SortOrder:   effectiveSortOrder,
	}

	if err := models.UpdateSettingWithMeta(setting); err != nil {
		utils.Fail(c, 500, "Failed to update setting")
		return
	}

	ctrl.refreshRuntimeConfig()

	utils.Success(c, gin.H{"message": "Setting updated successfully"})
}

// BatchUpdate 批量更新配置
// @Summary 批量更新配置
// @Description 批量更新多个配置项的值
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchUpdateSettingsRequest true "配置键值对"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings/batch [put]
func (ctrl *SettingsController) BatchUpdate(c *gin.Context) {
	var req BatchUpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	resolvedSettings := make(map[string]string, len(req.Settings))

	// 验证每个配置项是否可编辑
	for key := range req.Settings {
		setting, err := models.GetSettingByKey(key)
		if err != nil {
			utils.Fail(c, 404, "Setting not found: "+key)
			return
		}
		if !setting.IsEditable {
			utils.Fail(c, 403, "Setting is not editable: "+key)
			return
		}
		resolvedValue := ctrl.normalizeSettingValueForWrite(*setting, req.Settings[key])
		if !ctrl.validateSettingValue(resolvedValue, setting.Type) {
			utils.Fail(c, 400, "Invalid value type for "+key)
			return
		}
		resolvedSettings[key] = resolvedValue
	}

	// 批量更新
	if err := models.BatchUpdateSettings(resolvedSettings); err != nil {
		utils.Fail(c, 500, "Failed to update settings")
		return
	}

	ctrl.refreshRuntimeConfig()

	utils.Success(c, gin.H{"message": "Settings updated successfully"})
}

// Create 创建新配置
// @Summary 创建新配置
// @Description 创建一个新的自定义配置项
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateSettingRequest true "配置信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings [post]
func (ctrl *SettingsController) Create(c *gin.Context) {
	var req CreateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Key = utils.Clean_XSS(req.Key)
	req.Label = utils.Clean_XSS(req.Label)
	req.Description = utils.Clean_XSS(req.Description)

	// 验证key格式（只允许字母、数字、下划线）
	keyRegex := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !keyRegex.MatchString(req.Key) {
		utils.Fail(c, 400, "Key must start with lowercase letter and contain only lowercase letters, numbers, and underscores")
		return
	}

	// 检查key是否已存在
	if _, err := models.GetSettingByKey(req.Key); err == nil {
		utils.Fail(c, 400, "Setting key already exists")
		return
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "string"
	}
	if req.Category == "" {
		req.Category = "custom"
	}

	// 验证类型
	if !ctrl.isValidType(req.Type) {
		utils.Fail(c, 400, "Invalid type. Must be one of: string, number, boolean, json")
		return
	}

	// 类型校验
	if !ctrl.validateSettingValue(req.Value, req.Type) {
		utils.Fail(c, 400, "Invalid value type for "+req.Type)
		return
	}

	// 验证分类
	if !ctrl.isValidCategory(req.Category) {
		utils.Fail(c, 400, "Invalid category. Must be one of: basic, security, email, payment, sms, custom")
		return
	}

	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	// 新建时同样禁止敏感 key 公开；自定义 key 名命中敏感名单也拦截
	if isPublic && isSensitiveSettingKey(req.Key) {
		utils.Fail(c, 400, "敏感配置不允许设为公开")
		return
	}

	isEditable := true
	if req.IsEditable != nil {
		isEditable = *req.IsEditable
	}

	// 创建配置
	setting := &models.SystemSetting{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Label:       req.Label,
		Description: req.Description,
		IsPublic:    isPublic,
		IsEditable:  isEditable,
		SortOrder:   req.SortOrder,
	}

	if err := models.CreateSetting(setting); err != nil {
		utils.Fail(c, 500, "Failed to create setting")
		return
	}

	ctrl.refreshRuntimeConfig()

	utils.Success(c, gin.H{
		"message": "Setting created successfully",
		"key":     req.Key,
	})
}

// Delete 删除配置
// @Summary 删除配置
// @Description 删除指定的自定义配置项
// @Tags Admin-系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键名"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/settings/{key} [delete]
func (ctrl *SettingsController) Delete(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.Fail(c, 400, "Key is required")
		return
	}

	// 检查配置是否存在
	setting, err := models.GetSettingByKey(key)
	if err != nil {
		utils.Fail(c, 404, "Setting not found")
		return
	}

	// 只允许删除自定义配置
	if setting.Category != "custom" {
		utils.Fail(c, 403, "Only custom settings can be deleted")
		return
	}

	if err := models.DeleteSetting(key); err != nil {
		utils.Fail(c, 500, "Failed to delete setting")
		return
	}

	ctrl.refreshRuntimeConfig()

	utils.Success(c, gin.H{"message": "Setting deleted successfully"})
}

// RegisterRoutes 注册管理端配置路由
func (ctrl *SettingsController) RegisterRoutes(group *gin.RouterGroup) {
	settings := group.Group("/settings")
	{
		settings.GET("", ctrl.List)
		settings.GET("/category/:category", ctrl.GetByCategory)
		settings.GET("/server-monitoring", ctrl.GetServerMonitoringStatus)
		settings.GET("/server-ops", ctrl.GetServerOperationsStatus)
		settings.POST("", ctrl.Create)
		settings.POST("/restart-backend", ctrl.RestartBackend)
		settings.PUT("/batch", ctrl.BatchUpdate)
		settings.GET("/:key", ctrl.Get)
		settings.PUT("/:key", ctrl.Update)
		settings.PUT("/:key/meta", ctrl.UpdateMeta)
		settings.DELETE("/:key", ctrl.Delete)
	}
}
