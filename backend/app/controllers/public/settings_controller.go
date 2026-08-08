package public

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
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
	AllowRegister       bool `json:"allow_register"`
	AllowUserLogin      bool `json:"allow_user_login"`
	AnnouncementEnabled bool `json:"announcement_enabled"`
	AllowDeleteAccount  bool `json:"allow_delete_account"`
	GeetestEnabled      bool `json:"geetest_enabled"`

	// 极验配置
	GeetestCaptchaId string `json:"geetest_captcha_id"`

	// 语言配置
	DefaultLang string `json:"default_lang"`

	// 验证码开关
	EmailVerifyEnabled bool `json:"email_verify_enabled"`
	SMSVerifyEnabled   bool `json:"sms_verify_enabled"`
	MobileCNOnly       bool `json:"mobile_cn_only"`
	// MobileIPCountryDetect 仅在关闭「仅大陆号」时生效：按 IP/CDN 预选区号
	MobileIPCountryDetect bool `json:"mobile_ip_country_detect"`

	// 实名认证配置
	RealnameEnabled    bool   `json:"realname_enabled"`
	RealnameNotifyText string `json:"realname_notify_text"`

	// 提现配置
	WithdrawEnabled      bool     `json:"withdraw_enabled"`
	WithdrawMinAmount    float64  `json:"withdraw_min_amount"`
	WithdrawNotifyText   string   `json:"withdraw_notify_text"`
	WithdrawAccountTypes []string `json:"withdraw_account_types"`

	// 管理端 REST API 在 /api/v1 下的前缀（来自 env ADMIN_API_PATH，默认 /admin）
	AdminAPIPath string `json:"admin_api_path"`

	// Presence / 在线心跳总开关（默认 false）
	PresenceEnabled bool `json:"presence_enabled"`

	// 在线心跳上报周期（秒），前端 Presence 心跳按此间隔发送
	OnlineReportIntervalSeconds int `json:"online_report_interval_seconds"`

	UserAPILogVisible       bool `json:"user_api_log_visible"`
	UserOperationLogVisible bool `json:"user_operation_log_visible"`
}
//@name 应用配置响应

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
		SiteName:           "F.st",
		SiteDesc:           "Full-stack admin template based on Go + Vue 3",
		Copyright:          "(c) 2024 F.st",
		Version:            "1.0.0",
		AllowRegister:       true,
		AllowUserLogin:      true,
		AnnouncementEnabled: false,
		AllowDeleteAccount:  false,
		DefaultLang:        "zhCN",
		GeetestEnabled:     false,
		GeetestCaptchaId:   "",
		EmailVerifyEnabled: true,
		SMSVerifyEnabled:   false,
		MobileCNOnly:          true,
		MobileIPCountryDetect: false,
		RealnameEnabled:       true,
		RealnameNotifyText: "完成实名认证后可享受更多服务",
		WithdrawEnabled:    true,
		WithdrawMinAmount:  10,
		WithdrawNotifyText: "提现申请提交后需管理员审核，通过后人工打款。",
		WithdrawAccountTypes: []string{"bank", "alipay", "wechat", "usdt"},
		AdminAPIPath:       "/admin",
		PresenceEnabled:             false,
		OnlineReportIntervalSeconds: 30,
		UserAPILogVisible:       true,
		UserOperationLogVisible: true,
	}
	// 从全局配置快照读取 ADMIN_API_PATH，供前端注入管理端请求前缀
	if cfg := config.CloneGlobalConfig(); cfg != nil {
		response.AdminAPIPath = config.NormalizeAdminAPIPath(cfg.AdminAPIPath)
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
	if v, ok := configMap["allow_user_login"]; ok {
		response.AllowUserLogin = v == "true" || v == "1"
	} else {
		response.AllowUserLogin = true
	}
	if v, ok := configMap["announcement_enabled"]; ok {
		response.AnnouncementEnabled = v == "true" || v == "1"
	}
	if v, ok := configMap["allow_delete_account"]; ok {
		response.AllowDeleteAccount = v == "true" || v == "1"
	}
	if v, ok := configMap["default_lang"]; ok {
		response.DefaultLang = v
	}
	if v, ok := configMap["email_verify_enabled"]; ok {
		response.EmailVerifyEnabled = v == "true" || v == "1"
	}
	if v, ok := configMap["sms_verify_enabled"]; ok {
		response.SMSVerifyEnabled = v == "true" || v == "1"
	}
	if v, ok := configMap["mobile_cn_only"]; ok {
		response.MobileCNOnly = v == "true" || v == "1"
	} else {
		response.MobileCNOnly = true
	}
	if v, ok := configMap["mobile_ip_country_detect"]; ok {
		response.MobileIPCountryDetect = v == "true" || v == "1"
	} else {
		response.MobileIPCountryDetect = false
	}
	if v, ok := configMap["realname_enabled"]; ok {
		response.RealnameEnabled = v == "true" || v == "1"
	}
	if v, ok := configMap["realname_notify_text"]; ok && strings.TrimSpace(v) != "" {
		response.RealnameNotifyText = v
	}
	if v, ok := configMap["withdraw_enabled"]; ok {
		response.WithdrawEnabled = v == "true" || v == "1"
	}
	if v, ok := configMap["withdraw_min_amount"]; ok && strings.TrimSpace(v) != "" {
		response.WithdrawMinAmount = services.ParseJSONFloatForPublic(v, 10)
	}
	if v, ok := configMap["withdraw_notify_text"]; ok && strings.TrimSpace(v) != "" {
		response.WithdrawNotifyText = v
	}
	if v, ok := configMap["withdraw_account_types"]; ok && strings.TrimSpace(v) != "" {
		response.WithdrawAccountTypes = services.ParseJSONStringArrayForPublic(v, []string{"bank", "alipay", "wechat", "usdt"})
	}
	geetestConfig := services.GetGlobalGeetestRuntimeConfig()
	if v, ok := configMap["geetest_enabled"]; ok {
		v = strings.TrimSpace(v)
		geetestConfig.Enabled = v == "true" || v == "1" || strings.EqualFold(v, "true")
	}

	if v, ok := configMap["geetest_captcha_id"]; ok {
		v = strings.TrimSpace(v)
		if v != "" {
			geetestConfig.CaptchaID = v
		}
	}

	if v, ok := configMap["geetest_captcha_key"]; ok {
		v = strings.TrimSpace(v)
		if v != "" {
			geetestConfig.CaptchaKey = v
		}
	}

	response.GeetestEnabled = geetestConfig.Enabled && geetestConfig.CaptchaID != "" && geetestConfig.CaptchaKey != ""
	response.GeetestCaptchaId = geetestConfig.CaptchaID
	if v, ok := configMap["presence_enabled"]; ok {
		response.PresenceEnabled = v == "true" || v == "1"
	} else {
		response.PresenceEnabled = false
	}
	response.OnlineReportIntervalSeconds = services.GetGlobalOnlinePresenceRuntimeConfig().ReportIntervalSeconds
	if v, ok := configMap["user_api_log_visible"]; ok {
		response.UserAPILogVisible = v == "true" || v == "1"
	} else {
		response.UserAPILogVisible = true
	}
	if v, ok := configMap["user_operation_log_visible"]; ok {
		response.UserOperationLogVisible = v == "true" || v == "1"
	} else {
		response.UserOperationLogVisible = true
	}

	return response
}

// RegisterRoutes 注册公共配置路由
func (ctrl *SettingsController) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/app-config", ctrl.GetAppConfig)
}

