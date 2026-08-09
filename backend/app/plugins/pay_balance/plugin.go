package pay_balance

import (
	"fst/backend/app/plugins"
	_ "fst/backend/app/plugins/pay_balance/alipay"
	"fst/backend/app/plugins/pay_balance/epay"
	_ "fst/backend/app/plugins/pay_balance/paypal"
	_ "fst/backend/app/plugins/pay_balance/stripe"
	_ "fst/backend/app/plugins/pay_balance/wechat"
	"fst/backend/app/services"
	"fst/backend/pkg/pluginregistry"
	"log"

	"github.com/gin-gonic/gin"
)

func init() {
	pluginregistry.Register(NewPlugin())
}

// Plugin 余额充值支付通道插件
// 负责注册各支付 SDK 通道（epay 及后续 alipay/wxpay/paypal 等）
type Plugin struct {
	plugins.BasePlugin
}

// NewPlugin 创建 pay_balance 插件实例
func NewPlugin() plugins.Plugin {
	p := &Plugin{
		BasePlugin: plugins.NewBasePlugin(
			"pay_balance",
			"1.0.0",
			"余额充值支付通道插件，支持易支付等第三方/官方支付 SDK",
		),
	}
	p.BasePlugin.SetPriority(95)
	return p
}

func (p *Plugin) Configure(cfg map[string]interface{}) error {
	return nil
}

func (p *Plugin) Init() error {
	// 注册易支付通道；后续同级目录 alipay/wxpay/paypal 在此一并 Register
	services.RegisterPaymentChannel(epay.NewChannel())
	log.Printf("[PayBalance] 已注册支付通道: %s", epay.ChannelType)
	return nil
}

func (p *Plugin) Migrate() error {
	return nil
}

func (p *Plugin) RegisterRoutes(router *gin.RouterGroup) {
	// 暂不新增对外路由，沿用现有支付 API
}

func (p *Plugin) Shutdown() error {
	log.Println("[PayBalance] 已关闭")
	return nil
}
