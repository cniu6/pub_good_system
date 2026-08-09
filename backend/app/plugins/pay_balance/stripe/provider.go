package stripe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"fst/backend/pkg/payment"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const ChannelType = "stripe"

func init() {
	payment.RegisterChannel(payment.ChannelMeta{
		Type:     ChannelType,
		Name:     "Stripe",
		Currency: "USD",
		PayTypes: []payment.PayTypeMeta{
			{Value: "card", Name: "Card"},
			{Value: "alipay", Name: "Alipay"},
			{Value: "wechat_pay", Name: "WeChat Pay"},
		},
		Devices: []payment.DeviceMeta{
			{Value: "pc", Name: "PC"},
			{Value: "mobile", Name: "Mobile"},
		},
		DefaultNotifyPath: "/api/v1/public/payment/notify/stripe",
		Versions: []payment.ChannelVersionMeta{
			{
				Version:   "v1",
				Name:      "V1（PaymentIntents）",
				SignTypes: []payment.SignTypeMeta{{Value: "HMAC", Name: "HMAC-SHA256 Webhook"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "secret_key",
						Label:       "Secret Key",
						Type:        "input",
						Required:    true,
						Secret:      true,
						Placeholder: "sk_...",
					},
					{
						Name:        "publishable_key",
						Label:       "Publishable Key",
						Type:        "input",
						Required:    false,
						Secret:      false,
						Placeholder: "pk_...",
					},
					{
						Name:        "webhook_secret",
						Label:       "Webhook Secret",
						Type:        "input",
						Required:    true,
						Secret:      true,
						Placeholder: "whsec_...",
					},
				},
			},
		},
	}, NewProvider)
}

// Provider Stripe 通道适配器
type Provider struct{}

// NewProvider 创建 Stripe Provider 实例
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
	case "card", "alipay", "wechat_pay", "bancontact", "ideal", "eps", "giropay", "p24", "sofort":
		return true
	default:
		return true
	}
}

func (p *Provider) apiURL() string {
	return "https://api.stripe.com"
}

func (p *Provider) CreatePay(ctx context.Context, req *payment.CreatePayRequest) (*payment.CreatePayResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}
	secretKey := strings.TrimSpace(extConfig["secret_key"])
	if secretKey == "" {
		return nil, fmt.Errorf("stripe secret_key missing")
	}

	minor, err := payment.ParseMoneyMinor(req.Money)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}

	apiURL := p.apiURL()
	data := []string{
		"amount=" + fmt.Sprintf("%d", minor),
		"currency=" + strings.ToLower(req.Currency),
		"metadata[order_no]=" + req.OrderNo,
		"automatic_payment_methods[enabled]=true",
	}
	if req.Subject != "" {
		data = append(data, "description="+req.Subject)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL+"/v1/payment_intents", bytes.NewReader([]byte(strings.Join(data, "&"))))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Authorization", "Bearer "+secretKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe create payment intent failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stripe create payment intent status %d: %s", resp.StatusCode, string(respBody))
	}

	var pi struct {
		ID           string `json:"id"`
		ClientSecret string `json:"client_secret"`
		Status       string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &pi); err != nil {
		return nil, err
	}

	return &payment.CreatePayResponse{
		PayURL:  "",
		TradeNo: pi.ID,
	}, nil
}

// VerifyNotify 旧接口回调，Stripe 不走 form 参数，直接返回 false
func (p *Provider) VerifyNotify(params map[string]string, signType string, extConfig map[string]string) bool {
	return false
}

// VerifyNotifyWithPayload 验证 Stripe webhook 回调
func (p *Provider) VerifyNotifyWithPayload(ctx context.Context, body []byte, headers map[string]string, signType string, extConfig map[string]string) (bool, *payment.NotifyPayload, error) {
	if extConfig == nil {
		return false, nil, fmt.Errorf("stripe ext_config missing")
	}
	webhookSecret := strings.TrimSpace(extConfig["webhook_secret"])
	if webhookSecret == "" {
		return false, nil, fmt.Errorf("stripe webhook_secret missing")
	}

	sigHeader := headers["stripe-signature"]
	if sigHeader == "" {
		return false, nil, fmt.Errorf("stripe-signature header missing")
	}

	ok, err := verifyWebhookSignature(body, sigHeader, webhookSecret, 5*time.Minute)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}

	var event struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Object string `json:"object"`
		Data   struct {
			Object struct {
				ID            string `json:"id"`
				Status        string `json:"status"`
				Amount        int64  `json:"amount"`
				Currency      string `json:"currency"`
				Metadata      map[string]string `json:"metadata"`
				PaymentMethod string `json:"payment_method"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return false, nil, err
	}

	outTradeNo := ""
	if event.Data.Object.Metadata != nil {
		outTradeNo = event.Data.Object.Metadata["order_no"]
	}
	if outTradeNo == "" {
		return false, nil, fmt.Errorf("stripe webhook missing order_no in metadata")
	}

	status := "PENDING"
	switch event.Type {
	case "payment_intent.succeeded", "charge.succeeded":
		status = "succeeded"
	case "payment_intent.payment_failed", "charge.failed":
		status = "failed"
	case "payment_intent.canceled":
		status = "canceled"
	}

	money := payment.FormatMoneyYuan(event.Data.Object.Amount)
	return true, &payment.NotifyPayload{
		OutTradeNo:  outTradeNo,
		TradeNo:     event.Data.Object.ID,
		TradeStatus: payment.NormalizeTradeStatus(status),
		Money:       money,
		PayType:     "card",
	}, nil
}

// QueryOrder 查询 Stripe PaymentIntent
func (p *Provider) QueryOrder(ctx context.Context, req *payment.QueryOrderRequest) (*payment.QueryOrderResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}
	secretKey := strings.TrimSpace(extConfig["secret_key"])
	if secretKey == "" {
		return nil, fmt.Errorf("stripe secret_key missing")
	}

	orderID := req.TradeNo
	if orderID == "" {
		return nil, fmt.Errorf("stripe trade_no missing")
	}

	apiURL := p.apiURL()
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/payment_intents/%s", apiURL, orderID), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+secretKey)

	client := &http.Client{Timeout: time.Duration(req.RequestTimeout) * time.Second}
	if client.Timeout == 0 {
		client.Timeout = 10 * time.Second
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe query order failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pi struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount int64  `json:"amount"`
	}
	if err := json.Unmarshal(respBody, &pi); err != nil {
		return nil, err
	}

	return &payment.QueryOrderResponse{
		Code:        1,
		TradeNo:     pi.ID,
		TradeStatus: payment.NormalizeTradeStatus(pi.Status),
		Money:       payment.FormatMoneyYuan(pi.Amount),
	}, nil
}

// Refund Stripe 退款
func (p *Provider) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	log.Printf("[Stripe] refund not fully implemented: order_no=%s", req.OrderNo)
	return nil, fmt.Errorf("stripe refund not implemented")
}
