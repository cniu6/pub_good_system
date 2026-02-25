package public

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// SettingsController 公共配置控制器（无需登录）
type SettingsController struct{}

// NewSettingsController 创建配置控制器
func NewSettingsController() *SettingsController {
	return &SettingsController{}
}

// AppConfigResponse 应用配置响应结构
type AppConfigResponse struct {
	// 基本配置
	SiteName  string `json:"site_name"`
	SiteDesc  string `json:"site_desc"`
	SiteLogo  string `json:"site_logo"`
	Copyright string `json:"copyright"`
	ICP       string `json:"icp"`
	Version   string `json:"version"`

	// 功能开关
	AllowRegister  bool `json:"allow_register"`
	GeetestEnabled bool `json:"geetest_enabled"`

	// 极验配置
	GeetestCaptchaId string `json:"geetest_captcha_id"`

	// 语言配置
	DefaultLang string `json:"default_lang"`
}

// GetAppConfig 获取应用配置
// @Summary 获取应用配置
// @Description 获取前端应用需要的公开配置信息
// @Tags Public-配置
// @Produce json
// @Success 200 {object} utils.Response{data=AppConfigResponse}
// @Router /api/v1/public/app-config [get]
func (ctrl *SettingsController) GetAppConfig(c *gin.Context) {
	// 尝试从缓存服务获取
	if services.GlobalSettingsService != nil {
		config := services.GlobalSettingsService.GetPublicAppConfig()
		utils.Success(c, config)
		return
	}

	// 回退：直接从数据库获取公开配置
	settings, err := models.GetPublicSettings()
	if err != nil {
		utils.Fail(c, 500, "Failed to load app config")
		return
	}

	// 转换为响应结构
	response := buildAppConfigResponse(settings)
	utils.Success(c, response)
}

// buildAppConfigResponse 从数据库配置构建响应
func buildAppConfigResponse(settings []models.SystemSetting) *AppConfigResponse {
	response := &AppConfigResponse{
		SiteName:         "F.st",
		SiteDesc:         "Full-stack admin template based on Go + Vue 3",
		Copyright:        "(c) 2024 F.st",
		Version:          "1.0.0",
		AllowRegister:    true,
		DefaultLang:      "zhCN",
		GeetestEnabled:   false,
		GeetestCaptchaId: "",
	}

	// 构建配置map
	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	// 填充配置
	if v, ok := configMap["site_name"]; ok {
		response.SiteName = v
	}
	if v, ok := configMap["site_desc"]; ok {
		response.SiteDesc = v
	}
	if v, ok := configMap["site_logo"]; ok {
		response.SiteLogo = v
	}
	if v, ok := configMap["copyright"]; ok {
		response.Copyright = v
	}
	if v, ok := configMap["icp"]; ok {
		response.ICP = v
	}
	if v, ok := configMap["version"]; ok {
		response.Version = v
	}
	if v, ok := configMap["allow_register"]; ok {
		response.AllowRegister = v == "true" || v == "1"
	}
	if v, ok := configMap["default_lang"]; ok {
		response.DefaultLang = v
	}
	enabled := config.GlobalConfig.GeetestEnabled
	if v, ok := configMap["geetest_enabled"]; ok {
		v = strings.TrimSpace(v)
		enabled = v == "true" || v == "1" || strings.EqualFold(v, "true")
	}

	captchaID := strings.TrimSpace(config.GlobalConfig.GeetestID)
	if v, ok := configMap["geetest_captcha_id"]; ok {
		v = strings.TrimSpace(v)
		if v != "" {
			captchaID = v
		}
	}

	captchaKey := strings.TrimSpace(config.GlobalConfig.GeetestKey)
	if v, ok := configMap["geetest_captcha_key"]; ok {
		v = strings.TrimSpace(v)
		if v != "" {
			captchaKey = v
		}
	}

	response.GeetestEnabled = enabled && captchaID != "" && captchaKey != ""
	response.GeetestCaptchaId = captchaID

	return response
}

// RegisterRoutes 注册公共配置路由
func (ctrl *SettingsController) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/app-config", ctrl.GetAppConfig)
}
