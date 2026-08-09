package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"fst/backend/pkg/payment"
	"io"
	"net/http"
	"strings"
	"time"
)

const ChannelType = "paypal"

func init() {
	payment.RegisterChannel(payment.ChannelMeta{
		Type:     ChannelType,
		Name:     "PayPal",
		Currency: "USD",
		PayTypes: []payment.PayTypeMeta{
			{Value: "paypal", Name: "PayPal Checkout"},
		},
		Devices: []payment.DeviceMeta{
			{Value: "pc", Name: "PC"},
			{Value: "mobile", Name: "Mobile"},
		},
		DefaultNotifyPath: "/api/v1/public/payment/notify/paypal",
		Versions: []payment.ChannelVersionMeta{
			{
				Version:   "v2",
				Name:      "V2（REST Orders）",
				SignTypes: []payment.SignTypeMeta{{Value: "OAuth2", Name: "OAuth2 Client Credentials"}},
				ConfigFields: []payment.ConfigField{
					{
						Name:        "client_id",
						Label:       "Client ID",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "PayPal App Client ID",
					},
					{
						Name:        "client_secret",
						Label:       "Client Secret",
						Type:        "input",
						Required:    true,
						Secret:      true,
						Placeholder: "PayPal App Client Secret",
					},
					{
						Name:        "webhook_id",
						Label:       "Webhook ID",
						Type:        "input",
						Required:    true,
						Secret:      false,
						Placeholder: "PayPal Webhook ID，用于回调验签",
					},
					{
						Name:        "mode",
						Label:       "模式",
						Type:        "select",
						Required:    false,
						Secret:      false,
						Placeholder: "sandbox 或 live",
						Options: []payment.ConfigFieldOption{
							{Value: "sandbox", Label: "Sandbox"},
							{Value: "live", Label: "Live"},
						},
					},
				},
			},
		},
	}, NewProvider)
}

// Provider PayPal 通道适配器
type Provider struct{}

// NewProvider 创建 PayPal Provider 实例
func NewProvider() payment.Provider {
	return &Provider{}
}

func (p *Provider) Type() string { return ChannelType }

func (p *Provider) ValidatePayType(payType string, extConfig map[string]string) bool {
	_ = payType
	_ = extConfig
	return true
}

func (p *Provider) apiURL(extConfig map[string]string) string {
	mode := strings.ToLower(strings.TrimSpace(extConfig["mode"]))
	if mode == "live" {
		return "https://api-m.paypal.com"
	}
	return "https://api-m.sandbox.paypal.com"
}

func (p *Provider) CreatePay(ctx context.Context, req *payment.CreatePayRequest) (*payment.CreatePayResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}
	clientID := strings.TrimSpace(extConfig["client_id"])
	clientSecret := strings.TrimSpace(extConfig["client_secret"])
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("paypal client_id / client_secret missing")
	}

	apiURL := strings.TrimRight(req.ApiURL, "/")
	if apiURL == "" {
		apiURL = p.apiURL(extConfig)
	}

	token, err := getAccessToken(clientID, clientSecret, apiURL, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("paypal get access token failed: %w", err)
	}

	minor, err := payment.ParseMoneyMinor(req.Money)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}
	amount := fmt.Sprintf("%.2f", float64(minor)/100.0)

	payload := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": req.OrderNo,
				"custom_id":    req.OrderNo,
				"amount": map[string]string{
					"currency_code": req.Currency,
					"value":         amount,
				},
			},
		},
		"application_context": map[string]string{
			"return_url": req.ReturnURL,
			"cancel_url": req.ReturnURL,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL+"/v2/checkout/orders", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("PayPal-Request-Id", req.OrderNo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("paypal create order failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paypal create order status %d: %s", resp.StatusCode, string(respBody))
	}

	var orderResp struct {
		ID    string `json:"id"`
		Links []struct {
			Href   string `json:"href"`
			Rel    string `json:"rel"`
			Method string `json:"method"`
		} `json:"links"`
	}
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, err
	}

	var approvalURL string
	for _, link := range orderResp.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}
	if approvalURL == "" && orderResp.ID != "" {
		approvalURL = fmt.Sprintf("%s/checkoutnow?token=%s", strings.Replace(apiURL, "api-m.", "www.", 1), orderResp.ID)
	}

	return &payment.CreatePayResponse{
		PayURL:  approvalURL,
		TradeNo: orderResp.ID,
	}, nil
}

// VerifyNotify 旧接口回调，PayPal 不走 form 参数，直接返回 false
func (p *Provider) VerifyNotify(params map[string]string, signType string, extConfig map[string]string) bool {
	return false
}

// VerifyNotifyWithPayload 验证 PayPal webhook 回调
func (p *Provider) VerifyNotifyWithPayload(ctx context.Context, body []byte, headers map[string]string, signType string, extConfig map[string]string) (bool, *payment.NotifyPayload, error) {
	if extConfig == nil {
		return false, nil, fmt.Errorf("paypal ext_config missing")
	}
	clientID := strings.TrimSpace(extConfig["client_id"])
	clientSecret := strings.TrimSpace(extConfig["client_secret"])
	webhookID := strings.TrimSpace(extConfig["webhook_id"])
	if clientID == "" || clientSecret == "" || webhookID == "" {
		return false, nil, fmt.Errorf("paypal client_id / client_secret / webhook_id missing")
	}

	apiURL := strings.TrimRight(extConfig["api_url"], "/")
	if apiURL == "" {
		apiURL = p.apiURL(extConfig)
	}

	token, err := getAccessToken(clientID, clientSecret, apiURL, 10*time.Second)
	if err != nil {
		return false, nil, fmt.Errorf("paypal get access token failed: %w", err)
	}

	payload := verifyWebhookPayload{
		AuthAlgo:         headers["paypal-auth-algo"],
		CertURL:          headers["paypal-cert-url"],
		TransmissionID:   headers["paypal-transmission-id"],
		TransmissionSig:  headers["paypal-transmission-sig"],
		TransmissionTime: headers["paypal-transmission-time"],
		WebhookID:        webhookID,
		WebhookEvent:     body,
	}

	ok, err := VerifyWebhookSignature(apiURL, token, payload, 10*time.Second)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}

	// 解析 webhook 事件体，提取订单信息
	var event struct {
		Resource struct {
			ID            string `json:"id"`
			Status        string `json:"status"`
			PurchaseUnits []struct {
				ReferenceID string `json:"reference_id"`
				Amount      struct {
					Value string `json:"value"`
				} `json:"amount"`
			} `json:"purchase_units"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return false, nil, err
	}

	outTradeNo := ""
	money := ""
	if len(event.Resource.PurchaseUnits) > 0 {
		outTradeNo = event.Resource.PurchaseUnits[0].ReferenceID
		money = event.Resource.PurchaseUnits[0].Amount.Value
	}
	if outTradeNo == "" {
		// 尝试从 resource.id 取 PayPal Order ID 作为 trade_no，但无法直接映射到系统订单
		// 这里需要 webhook 中的 reference_id 与系统 order_no 一致
		return false, nil, fmt.Errorf("paypal webhook missing reference_id")
	}

	status := "PENDING"
	if strings.EqualFold(event.Resource.Status, "COMPLETED") || strings.EqualFold(event.Resource.Status, "CAPTURED") {
		status = "TRADE_SUCCESS"
	}

	return true, &payment.NotifyPayload{
		OutTradeNo:  outTradeNo,
		TradeNo:     event.Resource.ID,
		TradeStatus: payment.NormalizeTradeStatus(status),
		Money:       money,
		PayType:     "paypal",
	}, nil
}

// QueryOrder 查询 PayPal 订单
func (p *Provider) QueryOrder(ctx context.Context, req *payment.QueryOrderRequest) (*payment.QueryOrderResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}
	clientID := strings.TrimSpace(extConfig["client_id"])
	clientSecret := strings.TrimSpace(extConfig["client_secret"])
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("paypal client_id / client_secret missing")
	}

	apiURL := strings.TrimRight(req.ApiURL, "/")
	if apiURL == "" {
		apiURL = p.apiURL(extConfig)
	}

	token, err := getAccessToken(clientID, clientSecret, apiURL, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("paypal get access token failed: %w", err)
	}

	orderID := req.TradeNo
	if orderID == "" {
		return nil, fmt.Errorf("paypal trade_no missing")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v2/checkout/orders/%s", apiURL, orderID), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: time.Duration(req.RequestTimeout) * time.Second}
	if client.Timeout == 0 {
		client.Timeout = 10 * time.Second
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("paypal query order failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orderResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, err
	}

	return &payment.QueryOrderResponse{
		Code:        1,
		TradeNo:     orderResp.ID,
		TradeStatus: payment.NormalizeTradeStatus(orderResp.Status),
		Money:       "",
	}, nil
}

// TestConnection 测试 PayPal 配置：获取 access_token
func (p *Provider) TestConnection(ctx context.Context, extConfig map[string]string) (bool, string) {
	clientID := strings.TrimSpace(extConfig["client_id"])
	clientSecret := strings.TrimSpace(extConfig["client_secret"])
	if clientID == "" || clientSecret == "" {
		return false, "缺少 client_id 或 client_secret"
	}
	apiURL := p.apiURL(extConfig)
	_, err := getAccessToken(clientID, clientSecret, apiURL, 10*time.Second)
	if err != nil {
		return false, err.Error()
	}
	return true, "连接成功"
}

// Refund PayPal 退款
// 先查询订单获取 capture_id，再调用 /v2/payments/captures/{capture_id}/refund
func (p *Provider) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	extConfig := req.ExtConfig
	if extConfig == nil {
		extConfig = make(map[string]string)
	}
	clientID := strings.TrimSpace(extConfig["client_id"])
	clientSecret := strings.TrimSpace(extConfig["client_secret"])
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("paypal client_id / client_secret missing")
	}

	apiURL := strings.TrimRight(req.ApiURL, "/")
	if apiURL == "" {
		apiURL = p.apiURL(extConfig)
	}

	token, err := getAccessToken(clientID, clientSecret, apiURL, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("paypal get access token failed: %w", err)
	}

	orderID := req.TradeNo
	if orderID == "" {
		// 如果 trade_no 为空，尝试从 out_trade_no 查 order
		return nil, fmt.Errorf("paypal refund requires trade_no (paypal order id)")
	}

	// 查询订单拿到 capture_id
	captureID, err := p.getCaptureID(ctx, apiURL, token, orderID)
	if err != nil {
		return nil, err
	}

	refundURL := fmt.Sprintf("%s/v2/payments/captures/%s/refund", apiURL, captureID)
	minor, err := payment.ParseMoneyMinor(req.Money)
	if err != nil {
		return nil, fmt.Errorf("invalid refund money: %w", err)
	}
	refundBody, _ := json.Marshal(map[string]interface{}{
		"amount": map[string]string{
			"value":         fmt.Sprintf("%.2f", float64(minor)/100.0),
			"currency_code": strings.ToUpper(req.ExtConfig["currency"]),
		},
	})
	if req.ExtConfig["currency"] == "" {
		refundBody, _ = json.Marshal(map[string]interface{}{
			"amount": map[string]string{
				"value":         fmt.Sprintf("%.2f", float64(minor)/100.0),
				"currency_code": "USD",
			},
		})
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", refundURL, bytes.NewReader(refundBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("PayPal-Request-Id", "refund-"+req.OrderNo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("paypal refund failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var refundResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Value string `json:"value"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		return nil, fmt.Errorf("paypal refund parse failed: %w", err)
	}

	code := 1
	if !strings.EqualFold(refundResp.Status, "COMPLETED") && !strings.EqualFold(refundResp.Status, "PENDING") {
		code = 0
	}

	return &payment.RefundResponse{
		Code:        code,
		Msg:         refundResp.Status,
		RefundNo:    refundResp.ID,
		OutRefundNo: "refund-" + req.OrderNo,
		TradeNo:     orderID,
		Money:       refundResp.Amount.Value,
	}, nil
}

// getCaptureID 从 PayPal order 中拿到第一个 capture_id
func (p *Provider) getCaptureID(ctx context.Context, apiURL, token, orderID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v2/checkout/orders/%s", apiURL, orderID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("paypal query order for refund failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var orderResp struct {
		PurchaseUnits []struct {
			Payments struct {
				Captures []struct {
					ID string `json:"id"`
				} `json:"captures"`
			} `json:"payments"`
		} `json:"purchase_units"`
	}
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return "", fmt.Errorf("paypal parse order for refund failed: %w", err)
	}
	for _, unit := range orderResp.PurchaseUnits {
		for _, capture := range unit.Payments.Captures {
			if capture.ID != "" {
				return capture.ID, nil
			}
		}
	}
	return "", fmt.Errorf("paypal order has no capture")
}
