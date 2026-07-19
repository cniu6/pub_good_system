package sms_plugin

import (
	"fst/backend/app/models"
	"fst/backend/app/plugins"
	smstemplates "fst/backend/app/plugins/sms/templates"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/pkg/pluginregistry"
	"log"

	"github.com/gin-gonic/gin"
)

func init() {
	pluginregistry.Register(NewPlugin())
}

type Plugin struct {
	plugins.BasePlugin
	templateMgr *smstemplates.Manager
}

var defaultPlugin *Plugin

func NewPlugin() plugins.Plugin {
	p := &Plugin{
		BasePlugin: plugins.NewBasePlugin(
			"sms",
			"1.0.0",
			"短信插件，支持阿里云/腾讯云/自定义HTTP多种服务商，多语言模板",
		),
	}
	p.BasePlugin.SetPriority(90)
	defaultPlugin = p
	return p
}

func (p *Plugin) Configure(cfg map[string]interface{}) error {
	log.Printf("[SMSPlugin] 配置已加载")
	return nil
}

func (p *Plugin) Init() error {
	p.templateMgr = smstemplates.NewManager()
	// 先装载内置默认，再尝试用数据库覆盖（库为空则保留默认）
	p.templateMgr.InitDefaultTemplates()
	models.InitSMSTemplatesTable()
	models.InitSMSTemplates()
	p.templateMgr.LoadFromDB()

	cfg := config.GlobalConfig
	if cfg != nil && services.GlobalSMSService != nil {
		ApplyRuntimeProvider(buildSMSConfig(cfg))
	}

	log.Printf("[SMSPlugin] 初始化完成，Provider: %s, Configured: %v",
		services.GlobalSMSService.GetProviderName(),
		services.GlobalSMSService.IsConfigured())
	return nil
}

// ReloadTemplates 供管理端更新/重置后热加载内存模板
func ReloadTemplates() {
	getTemplateManager().Reload()
}

func buildSMSConfig(cfg *config.Config) services.SMSConfig {
	if cfg == nil {
		return services.SMSConfig{}
	}
	return services.SMSConfig{
		Provider:       cfg.SMSProvider,
		AccessKey:      cfg.SMSAccessKey,
		SecretKey:      cfg.SMSSecretKey,
		SignName:       cfg.SMSSignName,
		TemplateCode:   cfg.SMSTemplateCode,
		TemplateCodeEN: cfg.SMSTemplateCodeEN,
		Region:         cfg.SMSRegion,
		Endpoint:       cfg.SMSEndpoint,
		SdkAppID:       cfg.SMSSdkAppID,
		BodyFormat:     cfg.SMSBodyFormat,
	}
}

func newProvider(cfg services.SMSConfig) services.SMSProvider {
	switch cfg.Provider {
	case "aliyun":
		return NewAliyunProvider(cfg)
	case "tencent":
		return NewTencentProvider(cfg)
	case "custom":
		return NewCustomProvider(cfg)
	default:
		return nil
	}
}

// ApplyRuntimeProvider 统一负责根据最新配置重建并切换短信 Provider。
func ApplyRuntimeProvider(cfg services.SMSConfig) {
	if services.GlobalSMSService == nil {
		return
	}
	services.GlobalSMSService.SetConfig(cfg)
	services.GlobalSMSService.UseProvider(cfg.Provider, newProvider(cfg))
}

func getTemplateManager() *smstemplates.Manager {
	if defaultPlugin == nil {
		defaultPlugin = &Plugin{}
	}
	if defaultPlugin.templateMgr == nil {
		defaultPlugin.templateMgr = smstemplates.NewManager()
		defaultPlugin.templateMgr.InitDefaultTemplates()
	}
	return defaultPlugin.templateMgr
}

func (p *Plugin) Migrate() error {
	models.InitSMSTable()
	models.InitSMSTemplatesTable()
	models.InitSMSTemplates()
	log.Println("[SMSPlugin] 数据库迁移完成")
	return nil
}

func (p *Plugin) RegisterRoutes(router *gin.RouterGroup) {
	log.Printf("[SMSPlugin] 路由注册完成")
}

func (p *Plugin) Shutdown() error {
	log.Println("[SMSPlugin] 已关闭")
	return nil
}

