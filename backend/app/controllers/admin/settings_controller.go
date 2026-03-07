package admin

import (
	"context"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/internal/db"
	"fst/backend/utils"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
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
	IsPublic    bool   `json:"is_public"`
	IsEditable  bool   `json:"is_editable"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateSettingMetaRequest 更新配置元数据请求
type UpdateSettingMetaRequest struct {
	Value       string `json:"value"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Label       string `json:"label"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	IsEditable  bool   `json:"is_editable"`
	SortOrder   int    `json:"sort_order"`
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
	"sms":      "短信设置",
	"custom":   "自定义配置",
}

var serverMonitorStartedAt = time.Now()

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
	if !ctrl.validateSettingValue(req.Value, setting.Type) {
		utils.Fail(c, 400, "Invalid value type for "+setting.Type)
		return
	}

	// 更新配置
	if err := models.UpdateSetting(key, req.Value); err != nil {
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
	_, err := models.GetSettingByKey(key)
	if err != nil {
		utils.Fail(c, 404, "Setting not found")
		return
	}

	var req UpdateSettingMetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 验证类型
	if req.Type != "" && !ctrl.isValidType(req.Type) {
		utils.Fail(c, 400, "Invalid type. Must be one of: string, number, boolean, json")
		return
	}

	// 类型校验
	if !ctrl.validateSettingValue(req.Value, req.Type) {
		utils.Fail(c, 400, "Invalid value type for "+req.Type)
		return
	}

	// 构建更新对象
	setting := &models.SystemSetting{
		Key:         key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Label:       req.Label,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		IsEditable:  req.IsEditable,
		SortOrder:   req.SortOrder,
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
		if !ctrl.validateSettingValue(req.Settings[key], setting.Type) {
			utils.Fail(c, 400, "Invalid value type for "+key)
			return
		}
	}

	// 批量更新
	if err := models.BatchUpdateSettings(req.Settings); err != nil {
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
		utils.Fail(c, 400, "Invalid category. Must be one of: basic, security, email, custom")
		return
	}

	// 创建配置
	setting := &models.SystemSetting{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Label:       req.Label,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		IsEditable:  req.IsEditable,
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

// ========================================
// 辅助方法
// ========================================

// isValidType 验证配置类型是否有效
func (ctrl *SettingsController) isValidType(t string) bool {
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"json":    true,
	}
	return validTypes[t]
}

// isValidCategory 验证分类是否有效
func (ctrl *SettingsController) isValidCategory(cat string) bool {
	validCategories := map[string]bool{
		"basic":    true,
		"security": true,
		"email":    true,
		"sms":      true,
		"custom":   true,
	}
	return validCategories[cat]
}

// validateSettingValue 根据类型验证值
func (ctrl *SettingsController) validateSettingValue(value, typ string) bool {
	switch typ {
	case "number":
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case "boolean":
		lower := strings.ToLower(value)
		return lower == "true" || lower == "false" || lower == "1" || lower == "0"
	case "json":
		// JSON 类型允许任意字符串，前端负责解析
		return true
	case "string":
		return true
	default:
		return true
	}
}

func (ctrl *SettingsController) resolveSettingValueForAdmin(setting models.SystemSetting) interface{} {
	switch setting.Key {
	case "geetest_enabled":
		return services.GetGlobalGeetestRuntimeConfig().Enabled
	case "geetest_captcha_id":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(config.GlobalConfig.GeetestID)
		}
		return val
	case "geetest_captcha_key":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(config.GlobalConfig.GeetestKey)
		}
		return val
	case "smtp_host":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(config.GlobalConfig.SMTPHost)
		}
		return val
	case "smtp_port":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			if port, err := strconv.Atoi(strings.TrimSpace(config.GlobalConfig.SMTPPort)); err == nil && port > 0 {
				return port
			}
			return setting.GetTypedValue()
		}
		if port, err := strconv.Atoi(val); err == nil && port > 0 {
			return port
		}
		return setting.GetTypedValue()
	case "smtp_username":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(config.GlobalConfig.SMTPUser)
		}
		return val
	case "smtp_password":
		val := setting.Value
		if strings.TrimSpace(val) == "" {
			val = config.GlobalConfig.SMTPPass
		}
		return val
	case "smtp_ssl":
		if strings.TrimSpace(setting.Value) == "" {
			return config.GlobalConfig.SMTPSSL
		}
		return setting.GetTypedValue()
	case "system_email_name":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(config.GlobalConfig.SystemEmailName)
		}
		return val
	case "jwt_access_expire":
		if strings.TrimSpace(setting.Value) == "" {
			if config.GlobalConfig.JWTAccessExpire > 0 {
				return config.GlobalConfig.JWTAccessExpire
			}
		}
		return setting.GetTypedValue()
	case "jwt_refresh_expire":
		if strings.TrimSpace(setting.Value) == "" {
			if config.GlobalConfig.JWTRefreshExpire > 0 {
				return config.GlobalConfig.JWTRefreshExpire
			}
		}
		return setting.GetTypedValue()
	case "login_max_failure":
		if strings.TrimSpace(setting.Value) == "" {
			if config.GlobalConfig.LoginMaxFailureCount > 0 {
				return config.GlobalConfig.LoginMaxFailureCount
			}
		}
		return setting.GetTypedValue()
	case "login_lock_duration":
		if strings.TrimSpace(setting.Value) == "" {
			if config.GlobalConfig.LoginLockDurationMinutes > 0 {
				return config.GlobalConfig.LoginLockDurationMinutes
			}
		}
		return setting.GetTypedValue()
	case "email_verify_enabled":
		return services.GetGlobalVerifyConfig().EmailEnabled
	case "sms_verify_enabled":
		return services.GetGlobalVerifyConfig().SMSEnabled
	case "sms_provider":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = config.GlobalConfig.SMSProvider
		}
		if val == "" {
			val = "console"
		}
		return val
	case "sms_access_key":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = config.GlobalConfig.SMSAccessKey
		}
		return val
	case "sms_secret_key":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = config.GlobalConfig.SMSSecretKey
		}
		return val
	case "sms_sign_name":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = config.GlobalConfig.SMSSignName
		}
		return val
	case "sms_template_code":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = config.GlobalConfig.SMSTemplateCode
		}
		return val
	case "sms_region":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = config.GlobalConfig.SMSRegion
		}
		return val
	default:
		return setting.GetTypedValue()
	}
}

func parseBoolSettingValue(v string, fallback bool) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return fallback
	}
	if v == "1" || v == "true" {
		return true
	}
	if v == "0" || v == "false" {
		return false
	}
	return fallback
}

func parsePositiveIntSetting(v string, fallback int) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	num, err := strconv.Atoi(v)
	if err != nil || num <= 0 {
		return fallback
	}
	return num
}

func (ctrl *SettingsController) refreshRuntimeConfig() {
	if services.GlobalSettingsService != nil {
		services.GlobalSettingsService.InvalidateCache()
	}

	geetest := services.GetGlobalGeetestRuntimeConfig()
	config.GlobalConfig.GeetestEnabled = geetest.Enabled
	config.GlobalConfig.GeetestID = geetest.CaptchaID
	config.GlobalConfig.GeetestKey = geetest.CaptchaKey

	keys := []string{
		"smtp_host",
		"smtp_port",
		"smtp_username",
		"smtp_password",
		"smtp_ssl",
		"system_email_name",
		"jwt_access_expire",
		"jwt_refresh_expire",
		"login_max_failure",
		"login_lock_duration",
	}
	settingMap, err := models.GetSettingsMap(keys)
	if err != nil {
		return
	}

	if v, ok := settingMap["smtp_host"]; ok {
		config.GlobalConfig.SMTPHost = strings.TrimSpace(v)
	}
	if v, ok := settingMap["smtp_port"]; ok {
		config.GlobalConfig.SMTPPort = strings.TrimSpace(v)
	}
	if v, ok := settingMap["smtp_username"]; ok {
		config.GlobalConfig.SMTPUser = strings.TrimSpace(v)
	}
	if v, ok := settingMap["smtp_password"]; ok {
		config.GlobalConfig.SMTPPass = v
	}
	if v, ok := settingMap["smtp_ssl"]; ok {
		config.GlobalConfig.SMTPSSL = parseBoolSettingValue(v, config.GlobalConfig.SMTPSSL)
	}
	if v, ok := settingMap["system_email_name"]; ok {
		config.GlobalConfig.SystemEmailName = strings.TrimSpace(v)
	}
	if v, ok := settingMap["jwt_access_expire"]; ok {
		config.GlobalConfig.JWTAccessExpire = parsePositiveIntSetting(v, config.GlobalConfig.JWTAccessExpire)
	}
	if v, ok := settingMap["jwt_refresh_expire"]; ok {
		config.GlobalConfig.JWTRefreshExpire = parsePositiveIntSetting(v, config.GlobalConfig.JWTRefreshExpire)
	}
	if v, ok := settingMap["login_max_failure"]; ok {
		config.GlobalConfig.LoginMaxFailureCount = parsePositiveIntSetting(v, config.GlobalConfig.LoginMaxFailureCount)
	}
	if v, ok := settingMap["login_lock_duration"]; ok {
		config.GlobalConfig.LoginLockDurationMinutes = parsePositiveIntSetting(v, config.GlobalConfig.LoginLockDurationMinutes)
	}

	// 同步邮箱/短信验证开关
	verifyConfig := services.GetGlobalVerifyConfig()
	config.GlobalConfig.EmailVerifyEnabled = verifyConfig.EmailEnabled
	config.GlobalConfig.SMSVerifyEnabled = verifyConfig.SMSEnabled

	// 同步短信服务配置并更新 SMS Provider
	smsConfig := services.GetGlobalSMSRuntimeConfig()
	config.GlobalConfig.SMSProvider = smsConfig.Provider
	config.GlobalConfig.SMSAccessKey = smsConfig.AccessKey
	config.GlobalConfig.SMSSecretKey = smsConfig.SecretKey
	config.GlobalConfig.SMSSignName = smsConfig.SignName
	config.GlobalConfig.SMSTemplateCode = smsConfig.TemplateCode
	config.GlobalConfig.SMSRegion = smsConfig.Region

	if services.GlobalSMSService != nil {
		services.GlobalSMSService.SetConfig(services.SMSConfig{
			Provider:     smsConfig.Provider,
			AccessKey:    smsConfig.AccessKey,
			SecretKey:    smsConfig.SecretKey,
			SignName:     smsConfig.SignName,
			TemplateCode: smsConfig.TemplateCode,
			Region:       smsConfig.Region,
		})
	}
}

// RestartBackend restarts backend process after response is flushed.
func (ctrl *SettingsController) RestartBackend(c *gin.Context) {
	utils.Success(c, gin.H{"message": "Backend restart requested"})
	go func() {
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()
}

// GetServerMonitoringStatus 返回当前项目服务端运行监控快照。
func (ctrl *SettingsController) GetServerMonitoringStatus(c *gin.Context) {
	now := time.Now()

	var runtimeMem runtime.MemStats
	runtime.ReadMemStats(&runtimeMem)

	cpuUsage := 0.0
	if values, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(values) > 0 {
		cpuUsage = values[0]
	}

	vmTotalMB := 0.0
	vmUsedMB := 0.0
	vmUsedPercent := 0.0
	swapTotalMB := 0.0
	swapUsedMB := 0.0
	swapPercent := 0.0
	if vm, err := mem.VirtualMemory(); err == nil {
		vmTotalMB = bytesToMB(vm.Total)
		vmUsedMB = bytesToMB(vm.Used)
		vmUsedPercent = vm.UsedPercent
	}
	if swap, err := mem.SwapMemory(); err == nil {
		swapTotalMB = bytesToMB(swap.Total)
		swapUsedMB = bytesToMB(swap.Used)
		swapPercent = swap.UsedPercent
	}

	diskPath := "."
	diskTotalGB := 0.0
	diskUsedGB := 0.0
	diskUsedPercent := 0.0
	if du, err := disk.Usage(diskPath); err == nil {
		diskTotalGB = bytesToGB(du.Total)
		diskUsedGB = bytesToGB(du.Used)
		diskUsedPercent = du.UsedPercent
	}

	netBytesSent := uint64(0)
	netBytesRecv := uint64(0)
	netPacketsSent := uint64(0)
	netPacketsRecv := uint64(0)
	if counters, err := gnet.IOCounters(false); err == nil && len(counters) > 0 {
		netBytesSent = counters[0].BytesSent
		netBytesRecv = counters[0].BytesRecv
		netPacketsSent = counters[0].PacketsSent
		netPacketsRecv = counters[0].PacketsRecv
	}

	pid := int32(os.Getpid())
	procCPUPercent := 0.0
	procRSSMB := 0.0
	if p, err := process.NewProcess(pid); err == nil {
		if cpuPercent, err := p.CPUPercent(); err == nil {
			procCPUPercent = cpuPercent
		}
		if memInfo, err := p.MemoryInfo(); err == nil {
			procRSSMB = bytesToMB(memInfo.RSS)
		}
	}

	dbStatus := buildDatabaseStatus()
	smtpStatus := ctrl.buildSMTPStatus()

	utils.Success(c, gin.H{
		"generated_at":   now.Format(time.RFC3339),
		"uptime_seconds": int64(now.Sub(serverMonitorStartedAt).Seconds()),
		"app": gin.H{
			"name":       config.GlobalConfig.AppName,
			"mode":       config.GlobalConfig.AppMode,
			"port":       config.GlobalConfig.Port,
			"go_version": runtime.Version(),
		},
		"metrics": gin.H{
			"cpu": gin.H{
				"usage_percent": cpuUsage,
				"core_count":    runtime.NumCPU(),
			},
			"memory": gin.H{
				"total_mb":     vmTotalMB,
				"used_mb":      vmUsedMB,
				"used_percent": vmUsedPercent,
			},
			"swap": gin.H{
				"total_mb":     swapTotalMB,
				"used_mb":      swapUsedMB,
				"used_percent": swapPercent,
			},
			"disk": gin.H{
				"path":         diskPath,
				"total_gb":     diskTotalGB,
				"used_gb":      diskUsedGB,
				"used_percent": diskUsedPercent,
			},
			"network": gin.H{
				"bytes_sent":   netBytesSent,
				"bytes_recv":   netBytesRecv,
				"packets_sent": netPacketsSent,
				"packets_recv": netPacketsRecv,
			},
		},
		"process": gin.H{
			"pid":             pid,
			"goroutines":      runtime.NumGoroutine(),
			"process_cpu":     procCPUPercent,
			"process_rss_mb":  procRSSMB,
			"memory_alloc_mb": bytesToMB(runtimeMem.Alloc),
			"memory_sys_mb":   bytesToMB(runtimeMem.Sys),
			"heap_alloc_mb":   bytesToMB(runtimeMem.HeapAlloc),
			"heap_inuse_mb":   bytesToMB(runtimeMem.HeapInuse),
			"heap_idle_mb":    bytesToMB(runtimeMem.HeapIdle),
			"stack_inuse_mb":  bytesToMB(runtimeMem.StackInuse),
			"gc_count":        runtimeMem.NumGC,
			"gc_cpu_fraction": runtimeMem.GCCPUFraction,
		},
		"services": []gin.H{dbStatus, smtpStatus},
	})
}

func bytesToMB(v uint64) float64 {
	return float64(v) / 1024.0 / 1024.0
}

func bytesToGB(v uint64) float64 {
	return float64(v) / 1024.0 / 1024.0 / 1024.0
}

func buildDatabaseStatus() gin.H {
	database := db.GetDB()
	if database == nil {
		return gin.H{
			"name":    "MySQL",
			"status":  "down",
			"message": "数据库连接未初始化",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return gin.H{
			"name":    "MySQL",
			"status":  "down",
			"message": err.Error(),
		}
	}

	stats := database.Stats()
	return gin.H{
		"name":             "MySQL",
		"status":           "up",
		"message":          "连接正常",
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (ctrl *SettingsController) buildSMTPStatus() gin.H {
	settingMap, _ := models.GetSettingsMap([]string{"smtp_host", "smtp_port", "smtp_username", "smtp_password"})

	host := firstNonEmpty(settingMap["smtp_host"], config.GlobalConfig.SMTPHost)
	port := firstNonEmpty(settingMap["smtp_port"], config.GlobalConfig.SMTPPort)
	username := firstNonEmpty(settingMap["smtp_username"], config.GlobalConfig.SMTPUser)
	password := firstNonEmpty(settingMap["smtp_password"], config.GlobalConfig.SMTPPass)

	configured := host != "" && port != "" && username != "" && password != ""
	if !configured {
		return gin.H{
			"name":       "SMTP",
			"status":     "warning",
			"message":    "SMTP 未完成配置",
			"configured": false,
		}
	}

	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return gin.H{
			"name":       "SMTP",
			"status":     "down",
			"message":    err.Error(),
			"configured": true,
			"host":       host,
			"port":       port,
		}
	}
	_ = conn.Close()

	return gin.H{
		"name":       "SMTP",
		"status":     "up",
		"message":    "连接正常",
		"configured": true,
		"host":       host,
		"port":       port,
	}
}

// RegisterRoutes 注册管理端配置路由
func (ctrl *SettingsController) RegisterRoutes(group *gin.RouterGroup) {
	settings := group.Group("/settings")
	{
		settings.GET("", ctrl.List)
		settings.GET("/category/:category", ctrl.GetByCategory)
		settings.GET("/server-monitoring", ctrl.GetServerMonitoringStatus)
		settings.POST("", ctrl.Create)
		settings.POST("/restart-backend", ctrl.RestartBackend)
		settings.PUT("/batch", ctrl.BatchUpdate)
		settings.GET("/:key", ctrl.Get)
		settings.PUT("/:key", ctrl.Update)
		settings.PUT("/:key/meta", ctrl.UpdateMeta)
		settings.DELETE("/:key", ctrl.Delete)
	}
}
