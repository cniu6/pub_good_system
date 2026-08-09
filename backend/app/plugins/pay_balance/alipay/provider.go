package alipay

import (
	"context"
	"encoding/json"
	"fmt"
	"fst/backend/pkg/payment"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ChannelType = "alipay"

func init() {
	payment.RegisterChannel(payment.ChannelMeta{
		Type:     ChannelType,
		Name:     "支付宝",
		Currency: "CNY",
		PayTypes: []payment.PayTypeMeta{
			{Value: "alipay", Name: "支付宝"},
			{Value: "pc", Name: "电脑网站支付"},
			{Value: "wap", Name: "手机网站支付"},
		},
		Devices: []payment.DeviceMeta{
			{Value: "pc", Name: "电脑浏览器"},
			{Value: "mobile", Name: "手机浏览器"},
		},
		DefaultNotifyPath: "/api/v1/public/payment/notify/alipay",
		Versions: []payment.ChannelVersionMeta{
			{
				Version:   "v2",
				Name:      "V2（RSA2）",
				SignTypes: []payment.SignTypeMeta{{Value: SignTypeRSA2, Name: "RSA2（SHA256WithRSA）"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "app_id",
						Label:       "应用 ID",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "支付宝开放平台应用 ID",
					},
					{
						Name:        "merchant_private_key",
						Label:       "应用私钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "应用 RSA2 私钥（PEM 格式）",
					},
					{
						Name:        "alipay_public_key",
						Label:       "支付宝公钥",
						Type:        "textarea",
						Required:    true,
						Secret:      true,
						Placeholder: "支付宝 RSA2 公钥（PEM 格式，用于回调验签）",
					},
					{
						Name:        "app_cert_sn",
						Label:       "应用证书 SN",
						Type:        "input",
						Required:    false,
						Secret:      false,
						Placeholder: "可选：应用证书序列号（公钥证书模式）",
					},
					{
						Name:        "alipay_root_cert_sn",
						Label:       "支付宝根证书 SN",
						Type:        "input",
						Required:    false,
						Secret:      false,
						Placeholder: "可选：支付宝根证书序列号（公钥证书模式）",
					},
				},
			},
		},
	}, NewProvider)
}

// Provider 支付宝通道适配器
type Provider struct{}

// NewProvider 创建支付宝 Provider 实例
func NewProvider() payment.Provider {
	return &Provider{}
}

func (p *Provider) Type() string { return ChannelType }

func (p *Provider) ValidatePayType(payType string, extConfig map[string]string) bool {
	return validatePayType(payType, extConfig)
}

// CreatePay 构造支付宝电脑/手机网站支付跳转 URL
func (p *Provider) CreatePay(ctx context.Context, req *payment.CreatePayRequest) (*payment.CreatePayResponse, error) {
	signType := FormatSignType(req.SignType)
	if signType != SignTypeRSA2 {
		return nil, fmt.Errorf("alipay only supports RSA2")
	}

	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}

	appID := strings.TrimSpace(extConfig["app_id"])
	privateKey := strings.TrimSpace(extConfig["merchant_private_key"])
	if appID == "" || privateKey == "" {
		return nil, fmt.Errorf("alipay app_id or private key missing")
	}

	apiURL := strings.TrimRight(req.ApiURL, "/")
	if apiURL == "" {
		apiURL = "https://openapi.alipay.com/gateway.do"
	}

	method := "alipay.trade.page.pay"
	if req.Device == "mobile" || strings.ToLower(req.PayType) == "wap" {
		method = "alipay.trade.wap.pay"
	}

	minor, err := payment.ParseMoneyMinor(req.Money)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}
	totalAmount := fmt.Sprintf("%.2f", float64(minor)/100.0)

	bizContent := map[string]string{
		"out_trade_no": req.OrderNo,
		"total_amount": totalAmount,
		"subject":      req.Subject,
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}
	if method == "alipay.trade.wap.pay" {
		bizContent["product_code"] = "QUICK_WAP_WAY"
	}

	params := map[string]string{
		"app_id":     appID,
		"method":     method,
		"format":     "JSON",
		"charset":    "utf-8",
		"sign_type":  signType,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"notify_url": req.NotifyURL,
		"return_url": req.ReturnURL,
		"biz_content": func() string {
			b, _ := json.Marshal(bizContent)
			return string(b)
		}(),
	}

	if sn := strings.TrimSpace(extConfig["app_cert_sn"]); sn != "" {
		params["app_cert_sn"] = sn
	}
	if rootSN := strings.TrimSpace(extConfig["alipay_root_cert_sn"]); rootSN != "" {
		params["alipay_root_cert_sn"] = rootSN
	}

	sign, err := SignWithRSA2(params, privateKey)
	if err != nil {
		return nil, fmt.Errorf("alipay sign failed: %w", err)
	}
	params["sign"] = sign

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return &payment.CreatePayResponse{PayURL: u.String(), TradeNo: ""}, nil
}

// VerifyNotify 验证支付宝回调签名
func (p *Provider) VerifyNotify(params map[string]string, signType string, extConfig map[string]string) bool {
	if extConfig == nil {
		return false
	}
	publicKey := strings.TrimSpace(extConfig["alipay_public_key"])
	if publicKey == "" {
		return false
	}
	sign := strings.TrimSpace(params["sign"])
	if sign == "" {
		return false
	}
	return VerifyWithRSA2(params, sign, publicKey)
}

// QueryOrder 向支付宝查询订单状态
func (p *Provider) QueryOrder(ctx context.Context, req *payment.QueryOrderRequest) (*payment.QueryOrderResponse, error) {
	signType := FormatSignType(req.SignType)
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}
	appID := strings.TrimSpace(extConfig["app_id"])
	privateKey := strings.TrimSpace(extConfig["merchant_private_key"])
	if appID == "" || privateKey == "" {
		return nil, fmt.Errorf("alipay app_id or private key missing")
	}

	apiURL := strings.TrimRight(req.ApiURL, "/")
	if apiURL == "" {
		apiURL = "https://openapi.alipay.com/gateway.do"
	}

	bizContent := map[string]string{}
	if req.TradeNo != "" {
		bizContent["trade_no"] = req.TradeNo
	}
	if req.OrderNo != "" {
		bizContent["out_trade_no"] = req.OrderNo
	}

	params := map[string]string{
		"app_id":      appID,
		"method":      "alipay.trade.query",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   signType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": JSONContent(bizContent),
	}

	sign, err := SignWithRSA2(params, privateKey)
	if err != nil {
		return nil, fmt.Errorf("alipay query sign failed: %w", err)
	}
	params["sign"] = sign

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	timeout := req.RequestTimeout
	if timeout <= 0 {
		timeout = 10
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("alipay query request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var aliResp alipayQueryResponse
	if err := json.Unmarshal(body, &aliResp); err != nil {
		log.Printf("[Alipay] query response parse failed: %s", string(body))
		return nil, fmt.Errorf("alipay query response parse failed: %w", err)
	}

	var tradeStatus string
	var money string
	if aliResp.AlipayTradeQueryResponse != nil {
		tradeStatus = aliResp.AlipayTradeQueryResponse.TradeStatus
		money = aliResp.AlipayTradeQueryResponse.TotalAmount
	}

	code := 0
	if aliResp.Code == "10000" {
		code = 1
	}
	return &payment.QueryOrderResponse{
		Code:        code,
		Msg:         aliResp.Msg,
		TradeNo:     aliResp.AlipayTradeQueryResponse.TradeNo,
		OutTradeNo:  aliResp.AlipayTradeQueryResponse.OutTradeNo,
		TradeStatus: payment.NormalizeTradeStatus(tradeStatus),
		Money:       money,
	}, nil
}

// Refund 支付宝退款（ stub，需按业务完善）
func (p *Provider) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("alipay refund not implemented")
}

func validatePayType(payType string, extConfig map[string]string) bool {
	_ = extConfig
	payType = strings.TrimSpace(payType)
	if payType == "" {
		return true
	}
	switch strings.ToLower(payType) {
	case "alipay", "pc", "wap", "app", "face_to_face", "scan":
		return true
	default:
		return true
	}
}

type alipayQueryResponse struct {
	Code                     string `json:"code"`
	Msg                      string `json:"msg"`
	AlipayTradeQueryResponse *struct {
		TradeNo     string `json:"trade_no"`
		OutTradeNo  string `json:"out_trade_no"`
		TradeStatus string `json:"trade_status"`
		TotalAmount string `json:"total_amount"`
	} `json:"alipay_trade_query_response"`
}
