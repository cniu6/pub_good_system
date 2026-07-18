package services

import (
	"fst/backend/app/models"
	"sync"
)

// PaymentQueryResult 统一查单结果（各通道 SDK 映射到此结构）
type PaymentQueryResult struct {
	Code        int
	Msg         string
	TradeNo     string
	OutTradeNo  string
	Type        string
	Name        string
	Money       string
	TradeStatus string
}

// PaymentChannel 通用支付通道接口（类似短信 SMSProvider）
// 具体实现放在 plugins/pay_balance/<channel>/，按 gateway.Type 分发
type PaymentChannel interface {
	// Type 通道类型标识，与 pay_gateways.type 一致，如 "epay"
	Type() string

	// CreatePay 向远端创建支付订单，返回支付链接与平台交易号
	CreatePay(gateway *models.PayGateway, order *models.PaymentOrder, notifyURL, returnURL string) (payURL, tradeNo string, err error)

	// VerifyNotify 校验异步/同步回调签名
	VerifyNotify(params map[string]string, key string) bool

	// QueryOrder 向远端查询订单状态
	QueryOrder(gateway *models.PayGateway, orderNo, tradeNo string) (*PaymentQueryResult, error)

	// ValidatePayType 校验支付方式是否被该通道允许
	ValidatePayType(gateway *models.PayGateway, payType string) bool
}

var (
	paymentChannelMu sync.RWMutex
	paymentChannels  = make(map[string]PaymentChannel)
)

// RegisterPaymentChannel 注册支付通道实现（由 pay_balance 插件 Init 调用）
func RegisterPaymentChannel(channel PaymentChannel) {
	if channel == nil {
		return
	}
	paymentChannelMu.Lock()
	defer paymentChannelMu.Unlock()
	paymentChannels[channel.Type()] = channel
}

// GetPaymentChannel 按通道类型获取已注册实现
func GetPaymentChannel(channelType string) (PaymentChannel, bool) {
	paymentChannelMu.RLock()
	defer paymentChannelMu.RUnlock()
	ch, ok := paymentChannels[channelType]
	return ch, ok
}

// ListPaymentChannelTypes 返回已注册通道类型列表（测试/调试用）
func ListPaymentChannelTypes() []string {
	paymentChannelMu.RLock()
	defer paymentChannelMu.RUnlock()
	types := make([]string, 0, len(paymentChannels))
	for t := range paymentChannels {
		types = append(types, t)
	}
	return types
}

// ClearPaymentChannels 清空已注册支付通道（仅测试用）
func ClearPaymentChannels() {
	paymentChannelMu.Lock()
	defer paymentChannelMu.Unlock()
	paymentChannels = make(map[string]PaymentChannel)
}
