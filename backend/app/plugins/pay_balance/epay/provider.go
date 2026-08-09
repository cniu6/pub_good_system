package epay

import (
	"context"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/payment"
	"log"
	"strings"
	"time"
)

func init() {
	// 注册易支付通道到全局支付通道注册表
	payment.RegisterChannel(payment.ChannelMeta{
		Type:     ChannelType,
		Name:     "易支付",
		Currency: "CNY",
		PayTypes: []payment.PayTypeMeta{
			{Value: "alipay", Name: "支付宝"},
			{Value: "wxpay", Name: "微信支付"},
			{Value: "qqpay", Name: "QQ钱包"},
		},
		Devices: []payment.DeviceMeta{
			{Value: "pc", Name: "电脑浏览器"},
			{Value: "mobile", Name: "手机浏览器"},
			{Value: "qq", Name: "手机QQ内"},
			{Value: "wechat", Name: "微信内"},
			{Value: "alipay", Name: "支付宝客户端"},
			{Value: "jump", Name: "仅返回跳转URL"},
		},
		SupportCashbox:    true,
		DefaultNotifyPath: "/api/v1/public/payment/notify/epay",
		Versions: []payment.ChannelVersionMeta{
			{
				Version:   "v1",
				Name:      "V1（MD5）",
				SignTypes: []payment.SignTypeMeta{{Value: SignTypeMD5, Name: "MD5"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "merchant_key",
						Label:       "商户密钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "易支付商户密钥（MD5 签名用）",
					},
				},
			},
			{
				Version:   "v2",
				Name:      "V2（RSA）",
				SignTypes: []payment.SignTypeMeta{{Value: SignTypeRSA, Name: "SHA256WithRSA"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "merchant_private_key",
						Label:       "商户RSA私钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "商户后台生成的 RSA 私钥（PEM 格式）",
					},
					{
						Name:        "platform_public_key",
						Label:       "平台RSA公钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "易支付平台的 RSA 公钥（PEM 格式，用于验签回调）",
					},
				},
			},
		},
	}, NewProvider)
}

// Provider 易支付通道适配器
type Provider struct{}

// NewProvider 创建易支付 Provider 实例
func NewProvider() payment.Provider {
	return &Provider{}
}

// Type 返回通道类型标识
func (p *Provider) Type() string {
	return ChannelType
}

// CreatePay 调用易支付网关创建支付订单
func (p *Provider) CreatePay(ctx context.Context, req *payment.CreatePayRequest) (*payment.CreatePayResponse, error) {
	signType := FormatSignType(req.SignType)
	device := req.Device
	if device == "" {
		device = "pc"
	}

	config := &Config{
		ApiURL:    strings.TrimRight(req.ApiURL, "/"),
		PID:       req.PID,
		ExtConfig: req.ExtConfig,
		SignType:  signType,
		Device:    device,
	}

	// 优先使用目标币种/金额
	money := req.Money
	if req.TargetCurrency != "" && req.TargetMoney != "" {
		money = req.TargetMoney
	}

	order := &models.PaymentOrder{
		PaymentType: req.PayType,
		OrderNo:     req.OrderNo,
		PayAmount:   parseMoneyFloat(money),
		Subject:     req.Subject,
		ClientIP:    req.ClientIP,
	}

	// 优先走 mapi.php，失败则回退到 submit.php（保持与旧 Channel 行为一致）
	apiPayURL, tradeNo, apiErr := APIPay(config, order, req.NotifyURL, req.ReturnURL)
	if apiErr != nil {
		log.Printf("[Epay] APIPay failed, fallback to submit: %v", apiErr)
		payURL, err := BuildSubmitURL(config, order, req.NotifyURL, req.ReturnURL)
		if err != nil {
			return nil, err
		}
		return &payment.CreatePayResponse{PayURL: payURL, TradeNo: ""}, nil
	}
	return &payment.CreatePayResponse{PayURL: apiPayURL, TradeNo: tradeNo}, nil
}

// VerifyNotify 校验易支付回调签名
func (p *Provider) VerifyNotify(params map[string]string, signType string, extConfig map[string]string) bool {
	return VerifySignWithConfig(params, FormatSignType(signType), extConfig)
}

// QueryOrder 向易支付网关查询订单状态
func (p *Provider) QueryOrder(ctx context.Context, req *payment.QueryOrderRequest) (*payment.QueryOrderResponse, error) {
	signType := FormatSignType(req.SignType)
	timeout := req.RequestTimeout
	if timeout <= 0 {
		timeout = 10
	}

	config := &Config{
		ApiURL:    strings.TrimRight(req.ApiURL, "/"),
		PID:       req.PID,
		ExtConfig: req.ExtConfig,
		SignType:  signType,
	}

	result, err := QueryOrder(config, req.OrderNo, req.TradeNo)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = timeout
	return &payment.QueryOrderResponse{
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

// Refund 易支付退款
func (p *Provider) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	config := &Config{
		ApiURL:    strings.TrimRight(req.ApiURL, "/"),
		PID:       req.PID,
		ExtConfig: req.ExtConfig,
		SignType:  FormatSignType(req.SignType),
		Version:   req.Version,
	}

	money := req.Money
	if money == "" {
		// RefundRequest 默认按元传，若为空尝试从 ExtConfig 取
		money = req.ExtConfig["refund_money"]
	}
	if money == "" {
		return nil, fmt.Errorf("epay refund money missing")
	}

	outRefundNo := req.ExtConfig["out_refund_no"]
	if outRefundNo == "" {
		outRefundNo = fmt.Sprintf("R%s%d", req.OrderNo, time.Now().Unix())
	}

	result, err := Refund(config, req.OrderNo, req.TradeNo, money, outRefundNo)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("epay refund empty response")
	}

	return &payment.RefundResponse{
		Code:        result.Code,
		Msg:         result.Msg,
		RefundNo:    result.RefundNo,
		OutRefundNo: result.OutRefundNo,
		TradeNo:     result.TradeNo,
		Money:       result.Money,
	}, nil
}

// ValidatePayType 校验支付方式是否被该通道允许
func (p *Provider) ValidatePayType(payType string, extConfig map[string]string) bool {
	return validatePayTypeExt(payType, extConfig)
}

func parseMoneyFloat(money string) float64 {
	v, err := payment.ParseMoneyMinor(money)
	if err != nil {
		return 0
	}
	return float64(v) / 100.0
}

// TestConnection 测试易支付连接
// 用 TEST 订单号查询，只要能拿到非网络/非 HTML 响应即认为连接成功
func (p *Provider) TestConnection(ctx context.Context, extConfig map[string]string) (bool, string) {
	config := &Config{
		ApiURL:    strings.TrimRight(extConfig["api_url"], "/"),
		PID:       extConfig["pid"],
		Key:       extConfig["key"],
		ExtConfig: extConfig,
		SignType:  FormatSignType(extConfig["sign_type"]),
		Version:   extConfig["version"],
	}
	_, err := QueryOrder(config, "TEST", "")
	if err == nil {
		return true, "连接成功"
	}
	msg := err.Error()
	if strings.Contains(msg, "HTML") || strings.Contains(msg, "No such host") || strings.Contains(msg, "connection refused") {
		return false, "无法连接网关：" + msg
	}
	return true, "连接成功，但测试订单返回：" + msg
}

// validatePayTypeExt 校验 payType 是否在允许的支付方式列表中
func validatePayTypeExt(payType string, extConfig map[string]string) bool {
	payType = strings.TrimSpace(payType)
	if payType == "" {
		return true
	}
	allowed := extConfig["pay_types"]
	if allowed == "" {
		// 未配置时允许常见类型
		switch strings.ToLower(payType) {
		case "alipay", "wxpay", "qqpay", "bank", "jdpay", "paypal", "usdt":
			return true
		default:
			return true
		}
	}
	for _, t := range strings.Split(allowed, ",") {
		if strings.EqualFold(strings.TrimSpace(t), payType) {
			return true
		}
	}
	return false
}
