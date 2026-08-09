package wechat

import (
	"context"
	"fmt"
	"fst/backend/pkg/payment"
	"io"
	"net/http"
	"strings"
	"time"
)

const ChannelType = "wechat"

func init() {
	payment.RegisterChannel(payment.ChannelMeta{
		Type:     ChannelType,
		Name:     "微信支付",
		Currency: "CNY",
		PayTypes: []payment.PayTypeMeta{
			{Value: "native", Name: "Native 扫码支付"},
			{Value: "jsapi", Name: "JSAPI 公众号/小程序"},
			{Value: "h5", Name: "H5 支付"},
		},
		Devices: []payment.DeviceMeta{
			{Value: "pc", Name: "电脑浏览器"},
			{Value: "mobile", Name: "手机浏览器"},
			{Value: "mp", Name: "微信小程序"},
			{Value: "pub", Name: "微信公众号"},
		},
		DefaultNotifyPath: "/api/v1/public/payment/notify/wechat",
		Versions: []payment.ChannelVersionMeta{
			{
				Version:   "v2",
				Name:      "V2（MD5 / HMAC-SHA256）",
				SignTypes: []payment.SignTypeMeta{{Value: SignTypeMD5, Name: "MD5"}, {Value: SignTypeHMACSHA256, Name: "HMAC-SHA256"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "app_id",
						Label:       "公众号/小程序 AppID",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "微信分配的公众账号 ID",
					},
					{
						Name:        "mch_id",
						Label:       "商户号",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "微信支付商户号",
					},
					{
						Name:        "api_key",
						Label:       "API 密钥",
						Type:        "input",
						Required:    true,
						Secret:      true,
						Placeholder: "微信商户平台 API 密钥（用于 V2 签名）",
					},
					{
						Name:        "app_secret",
						Label:       "AppSecret",
						Type:        "input",
						Required:    false,
						Secret:      true,
						Placeholder: "公众号/小程序 AppSecret（JSAPI 需要）",
					},
				},
			},
			{
				Version:   "v3",
				Name:      "V3（RSA）",
				SignTypes: []payment.SignTypeMeta{{Value: "RSA", Name: "SHA256WithRSA"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "app_id",
						Label:       "AppID",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "微信应用 ID",
					},
					{
						Name:        "mch_id",
						Label:       "商户号",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "微信支付商户号",
					},
					{
						Name:        "api_v3_key",
						Label:       "APIv3 密钥",
						Type:        "input",
						Required:    true,
						Secret:      true,
						Placeholder: "微信商户平台 APIv3 密钥",
					},
					{
						Name:        "serial_no",
						Label:       "证书序列号",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "商户 API 证书序列号",
					},
					{
						Name:        "private_key",
						Label:       "商户 API 证书私钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "PEM 格式私钥",
					},
					{
						Name:        "platform_public_key",
						Label:       "微信支付平台公钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "PEM 格式公钥，用于回调验签",
					},
				},
			},
		},
	}, NewProvider)
}

// Provider 微信支付通道适配器
type Provider struct{}

// NewProvider 创建微信支付 Provider 实例
func NewProvider() payment.Provider {
	return &Provider{}
}

func (p *Provider) Type() string { return ChannelType }

func (p *Provider) ValidatePayType(payType string, extConfig map[string]string) bool {
	_ = extConfig
	payType = strings.TrimSpace(payType)
	if payType == "" {
		return true
	}
	switch strings.ToLower(payType) {
	case "native", "jsapi", "h5", "app", "miniprogram", "pub", "mweb":
		return true
	default:
		return true
	}
}

// CreatePay 构造微信支付请求（当前为模板实现）
// V2 需调用 unifiedorder XML 接口；V3 需 POST /v3/pay/transactions/* 并做请求签名
func (p *Provider) CreatePay(ctx context.Context, req *payment.CreatePayRequest) (*payment.CreatePayResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}

	version := strings.ToLower(strings.TrimSpace(req.Version))
	if version == "" {
		version = "v2"
	}

	apiURL := strings.TrimRight(req.ApiURL, "/")
	if apiURL == "" {
		if version == "v3" {
			apiURL = "https://api.mch.weixin.qq.com"
		} else {
			apiURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
		}
	}

	if version == "v2" {
		// V2 需要 mch_id / api_key 做 XML 请求签名
		if strings.TrimSpace(extConfig["mch_id"]) == "" || strings.TrimSpace(extConfig["api_key"]) == "" {
			return nil, fmt.Errorf("wechat v2 mch_id / api_key missing")
		}
		// 这里返回一个占位 pay_url；实际接入需调用 unifiedorder 获取 code_url / prepay_id
		return &payment.CreatePayResponse{
			PayURL:  "",
			QRCode:  "",
			TradeNo: "",
		}, nil
	}

	// V3 需要更多密钥与请求签名
	if strings.TrimSpace(extConfig["mch_id"]) == "" || strings.TrimSpace(extConfig["serial_no"]) == "" || strings.TrimSpace(extConfig["private_key"]) == "" {
		return nil, fmt.Errorf("wechat v3 mch_id / serial_no / private_key missing")
	}
	return &payment.CreatePayResponse{
		PayURL:  "",
		QRCode:  "",
		TradeNo: "",
	}, nil
}

// VerifyNotify 验证微信支付回调签名
// V2：从 form/XML 参数中取 sign 用 api_key 校验
// V3：需要 headers + body，当前接口只支持从参数中传 signature，生产环境建议扩展回调控制器透传 headers
func (p *Provider) VerifyNotify(params map[string]string, signType string, extConfig map[string]string) bool {
	if extConfig == nil {
		return false
	}
	version := strings.ToLower(strings.TrimSpace(params["version"]))
	if version == "" {
		version = "v2"
	}

	switch version {
	case "v3":
		// V3 回调验签依赖 Wechatpay-Signature / Timestamp / Nonce / Body，无法仅凭 params 完成
		// 若网关把签名放入 params["signature"] 且 body 在 params["body"] 中，可临时校验
		signature := strings.TrimSpace(params["signature"])
		body := strings.TrimSpace(params["body"])
		publicKey := strings.TrimSpace(extConfig["platform_public_key"])
		if signature == "" || body == "" || publicKey == "" {
			return false
		}
		return payment.RSAVerify(body, signature, publicKey)
	default:
		// V2
		apiKey := strings.TrimSpace(extConfig["api_key"])
		if apiKey == "" {
			return false
		}
		return VerifyWithV2(params, signType, apiKey)
	}
}

// QueryOrder 查询微信支付订单
func (p *Provider) QueryOrder(ctx context.Context, req *payment.QueryOrderRequest) (*payment.QueryOrderResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}

	version := strings.ToLower(strings.TrimSpace(req.SignType))
	if version == "" {
		version = "v2"
	}

	// 占位：实际接入需按 V2 orderquery 或 V3 GET /v3/pay/transactions/out-trade-no/{order_no}
	if strings.TrimSpace(extConfig["mch_id"]) == "" {
		return nil, fmt.Errorf("wechat mch_id missing")
	}

	return &payment.QueryOrderResponse{
		Code:        -1,
		Msg:         "wechat query not fully implemented",
		OutTradeNo:  req.OrderNo,
		TradeNo:     req.TradeNo,
		TradeStatus: "PENDING",
	}, nil
}

// TestConnection 测试微信支付配置
// V3：调用 /v3/certificates 拉取平台证书；V2：检查 mch_id / api_key 是否齐全
func (p *Provider) TestConnection(ctx context.Context, extConfig map[string]string) (bool, string) {
	version := strings.ToLower(strings.TrimSpace(extConfig["version"]))
	if version == "" {
		version = "v2"
	}

	if version == "v3" {
		mchID := strings.TrimSpace(extConfig["mch_id"])
		serialNo := strings.TrimSpace(extConfig["serial_no"])
		privateKey := strings.TrimSpace(extConfig["private_key"])
		if mchID == "" || serialNo == "" || privateKey == "" {
			return false, "缺少 mch_id / serial_no / private_key"
		}

		httpReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.mch.weixin.qq.com/v3/certificates", nil)
		if err != nil {
			return false, err.Error()
		}
		headers, err := buildV3AuthHeader(http.MethodGet, "/v3/certificates", nil, mchID, serialNo, privateKey)
		if err != nil {
			return false, err.Error()
		}
		delete(headers, "Wechatpay-Serial")
		for k, v := range headers {
			httpReq.Header.Set(k, v)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true, "连接成功"
		}
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	mchID := strings.TrimSpace(extConfig["mch_id"])
	apiKey := strings.TrimSpace(extConfig["api_key"])
	if mchID == "" || apiKey == "" {
		return false, "缺少 mch_id / api_key"
	}
	return true, "配置齐全（V2 需真实网络环境进一步验证）"
}

// Refund 微信退款
// V3 走 /v3/refund/domestic/refunds；V2 需要商户证书，暂不实现
func (p *Provider) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}

	version := strings.ToLower(strings.TrimSpace(req.Version))
	if version == "" {
		version = "v2"
	}

	if version == "v3" {
		return v3Refund(ctx, extConfig, req)
	}

	return nil, fmt.Errorf("wechat v2 refund not implemented: requires merchant client certificate")
}
