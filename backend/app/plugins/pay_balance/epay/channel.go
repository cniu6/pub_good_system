package epay

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"log"
)

const ChannelType = "epay"

// Channel 易支付通道实现（实现 services.PaymentChannel）
type Channel struct{}

// NewChannel 创建易支付通道实例
func NewChannel() services.PaymentChannel {
	return &Channel{}
}

func (c *Channel) Type() string {
	return ChannelType
}

// CreatePay 先尝试 mapi API，失败则回退到 submit.php 跳转支付
func (c *Channel) CreatePay(gateway *models.PayGateway, order *models.PaymentOrder, notifyURL, returnURL string) (string, string, error) {
	config := ConfigFromGateway(gateway)

	apiPayURL, tradeNo, apiErr := APIPay(config, order, notifyURL, returnURL)
	if apiErr != nil {
		log.Printf("[Epay] API支付失败，回退到跳转支付: %v", apiErr)
		payURL, err := BuildSubmitURL(config, order, notifyURL, returnURL)
		if err != nil {
			return "", "", err
		}
		return payURL, "", nil
	}
	return apiPayURL, tradeNo, nil
}

func (c *Channel) VerifyNotify(params map[string]string, key string) bool {
	return VerifySign(params, key)
}

func (c *Channel) QueryOrder(gateway *models.PayGateway, orderNo, tradeNo string) (*services.PaymentQueryResult, error) {
	config := ConfigFromGateway(gateway)
	result, err := QueryOrder(config, orderNo, tradeNo)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &services.PaymentQueryResult{
		Code:        result.Code,
		Msg:         result.Msg,
		TradeNo:     result.TradeNo,
		OutTradeNo:  result.OutTradeNo,
		Type:        result.Type,
		Name:        result.Name,
		Money:       result.Money,
		TradeStatus: result.TradeStatus,
	}, nil
}

func (c *Channel) ValidatePayType(gateway *models.PayGateway, payType string) bool {
	return ValidatePayType(ConfigFromGateway(gateway), payType)
}
