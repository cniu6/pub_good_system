package epay

import (
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/payment"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config 易支付配置（从 pay_gateways 单条记录构建）
type Config struct {
	ApiURL       string
	PID          string            // 商户ID
	Key          string            // 兼容旧通道：单密钥
	ExtConfig    map[string]string // 扩展配置（V1 MD5 / V2 RSA）
	SignType     string            // 签名算法：MD5 / RSA
	Version      string            // 版本：v1 / v2
	Device       string            // 设备类型：pc / mobile
	PaymentTypes []string
}

// IsV2 是否为 V2 接口（RSA + 新接口地址）
func (c *Config) IsV2() bool {
	return strings.EqualFold(strings.TrimSpace(c.Version), "v2")
}

// ConfigFromGateway 从支付通道模型构建易支付配置
func ConfigFromGateway(gateway *models.PayGateway) *Config {
	if gateway == nil {
		return &Config{}
	}
	extConfig := payment.ParseExtConfig(gateway.ExtConfig)
	// 兼容旧通道：没有 ext_config 时使用模型 getter 兜底密钥
	if len(extConfig) == 0 && gateway.GetKey() != "" {
		extConfig = map[string]string{"key": gateway.GetKey()}
	}
	signType := gateway.GetSignType()
	if signType == "" {
		signType = SignTypeMD5
	}
	device := gateway.Device
	if device == "" {
		device = "pc"
	}
	return &Config{
		ApiURL:       strings.TrimRight(gateway.ApiURL, "/"),
		PID:          gateway.PID,
		Key:          gateway.GetKey(),
		ExtConfig:    extConfig,
		SignType:     signType,
		Version:      gateway.Version,
		Device:       device,
		PaymentTypes: []string{gateway.PayType},
	}
}

// BuildSubmitURL 构造易支付跳转支付 URL（submit.php）
func BuildSubmitURL(config *Config, order *models.PaymentOrder, notifyURL, returnURL string) (string, error) {
	if config.ApiURL == "" || config.PID == "" || (config.Key == "" && len(config.ExtConfig) == 0) {
		return "", fmt.Errorf("Epay configuration incomplete")
	}

	params := map[string]string{
		"pid":          config.PID,
		"type":         order.PaymentType,
		"out_trade_no": order.OrderNo,
		"notify_url":   notifyURL,
		"return_url":   returnURL,
		"name":         order.Subject,
		"money":        fmt.Sprintf("%.2f", order.PayAmount),
	}

	params["sign"] = GenerateSignWithConfig(params, config.SignType, config.ExtConfig)
	params["sign_type"] = config.SignType

	u, err := url.Parse(config.ApiURL + "/submit.php")
	if err != nil {
		return "", fmt.Errorf("Failed to parse gateway address: %w", err)
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// QueryResponse 易支付查单响应
type QueryResponse struct {
	Code        int
	Msg         string
	TradeNo     string
	OutTradeNo  string
	Type        string
	Name        string
	Money       string
	TradeStatus string
}

// RefundResponse 易支付退款响应
type RefundResponse struct {
	Code        int
	Msg         string
	RefundNo    string
	OutRefundNo string
	TradeNo     string
	Money       string
}

type queryResponseRaw struct {
	Code        int             `json:"code"`
	Msg         string          `json:"msg"`
	TradeNo     string          `json:"trade_no"`
	OutTradeNo  string          `json:"out_trade_no"`
	Type        json.RawMessage `json:"type"`
	Name        string          `json:"name"`
	Money       string          `json:"money"`
	TradeStatus string          `json:"trade_status"`
	Status      json.RawMessage `json:"status"`
}

// QueryOrder 向易支付平台查询订单状态
func QueryOrder(config *Config, orderNo, tradeNo string) (*QueryResponse, error) {
	if config.IsV2() {
		return v2QueryOrder(config, orderNo, tradeNo)
	}

	key := GetSignKeyForQuery(config.ExtConfig)
	if key == "" {
		key = config.Key
	}
	if config.ApiURL == "" || config.PID == "" || key == "" {
		return nil, fmt.Errorf("Epay configuration incomplete")
	}

	orderNo = strings.TrimSpace(orderNo)
	tradeNo = models.NormalizeTradeNo(tradeNo)
	if orderNo == "" && tradeNo == "" {
		return nil, fmt.Errorf("Missing query order number")
	}

	queryCandidates := make([]map[string]string, 0, 2)
	if orderNo != "" {
		queryCandidates = append(queryCandidates, map[string]string{
			"act":          "order",
			"pid":          config.PID,
			"key":          key,
			"out_trade_no": orderNo,
		})
	}
	if tradeNo != "" {
		queryCandidates = append(queryCandidates, map[string]string{
			"act":      "order",
			"pid":      config.PID,
			"key":      key,
			"trade_no": tradeNo,
		})
	}

	var lastResult *QueryResponse
	for _, params := range queryCandidates {
		result, err := queryOrderOnce(config, params)
		if err != nil {
			return nil, err
		}
		if result == nil {
			continue
		}
		lastResult = result
		if result.Code == 1 {
			return result, nil
		}
	}

	return lastResult, nil
}

func queryOrderOnce(config *Config, params map[string]string) (*QueryResponse, error) {
	u, err := url.Parse(config.ApiURL + "/api.php")
	if err != nil {
		return nil, fmt.Errorf("Failed to parse gateway address: %w", err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to query Epay order: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %w", err)
	}

	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		return nil, fmt.Errorf("Query interface returned HTML instead of JSON, API URL may be misconfigured")
	}

	var raw queryResponseRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("[Epay] 查询响应解析失败: %s", string(body))
		return nil, fmt.Errorf("Failed to parse query response: %w", err)
	}

	return &QueryResponse{
		Code:        raw.Code,
		Msg:         raw.Msg,
		TradeNo:     raw.TradeNo,
		OutTradeNo:  raw.OutTradeNo,
		Type:        normalizeFlexibleString(raw.Type),
		Name:        raw.Name,
		Money:       raw.Money,
		TradeStatus: normalizeTradeStatus(raw.TradeStatus, raw.Status),
	}, nil
}

func normalizeFlexibleString(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == "false" {
		return ""
	}

	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return strings.TrimSpace(stringValue)
	}

	var boolValue bool
	if err := json.Unmarshal(raw, &boolValue); err == nil {
		if !boolValue {
			return ""
		}
		return "true"
	}

	var numberValue json.Number
	if err := json.Unmarshal(raw, &numberValue); err == nil {
		return strings.TrimSpace(numberValue.String())
	}

	return strings.Trim(trimmed, `"`)
}

func normalizeTradeStatus(tradeStatus string, statusRaw json.RawMessage) string {
	tradeStatus = strings.TrimSpace(tradeStatus)
	if tradeStatus != "" {
		return tradeStatus
	}

	status := strings.ToUpper(strings.TrimSpace(normalizeFlexibleString(statusRaw)))
	if status == "" {
		return ""
	}

	switch status {
	case "1", "SUCCESS", "SUCC", "PAID", "TRADE_SUCCESS", "FINISHED":
		return "TRADE_SUCCESS"
	case "0", "PENDING", "WAIT", "WAIT_BUYER_PAY", "UNPAID", "CREATED":
		return "PENDING"
	default:
		return status
	}
}

// APIPayResponse mapi.php 支付响应
type APIPayResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	PayURL    string `json:"payurl"`
	QRCode    string `json:"qrcode"`
	URLScheme string `json:"urlscheme"`
	TradeNo   string `json:"trade_no"`
}

// APIPay 通过 mapi.php（V1）或 /api/pay/create（V2）发起支付，返回支付链接与交易号
func APIPay(config *Config, order *models.PaymentOrder, notifyURL, returnURL string) (string, string, error) {
	if config.ApiURL == "" || config.PID == "" || (config.Key == "" && len(config.ExtConfig) == 0) {
		return "", "", fmt.Errorf("Epay configuration incomplete")
	}

	if config.IsV2() {
		return v2CreateOrder(config, order, notifyURL, returnURL)
	}

	device := config.Device
	if device == "" {
		device = "pc"
	}

	params := map[string]string{
		"pid":          config.PID,
		"type":         order.PaymentType,
		"out_trade_no": order.OrderNo,
		"notify_url":   notifyURL,
		"return_url":   returnURL,
		"name":         order.Subject,
		"money":        fmt.Sprintf("%.2f", order.PayAmount),
		"clientip":     order.ClientIP,
		"device":       device,
	}

	params["sign"] = GenerateSignWithConfig(params, config.SignType, config.ExtConfig)
	params["sign_type"] = config.SignType

	mapiURL := config.ApiURL + "/mapi.php"
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(mapiURL, formData)
	if err != nil {
		return "", "", fmt.Errorf("Payment interface request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("Payment interface returned error status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("Failed to read response content: %v", err)
	}

	if len(body) == 0 {
		return "", "", fmt.Errorf("Payment interface returned empty response")
	}

	bodyStr := string(body)
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "<") {
		return "", "", fmt.Errorf("Payment interface returned HTML instead of JSON, API URL may be misconfigured")
	}

	log.Printf("[Epay] APIPay 响应: %s", bodyStr)

	var payResp APIPayResponse
	if err := json.Unmarshal(body, &payResp); err != nil {
		return "", "", fmt.Errorf("Failed to parse payment response: %v, response content: %s", err, bodyStr)
	}

	if payResp.Code != 1 {
		return "", "", fmt.Errorf("Payment initiation failed: %s", payResp.Msg)
	}

	normalizedTradeNo := models.NormalizeTradeNo(payResp.TradeNo)

	if payResp.PayURL != "" {
		return payResp.PayURL, normalizedTradeNo, nil
	}
	if payResp.QRCode != "" {
		return payResp.QRCode, normalizedTradeNo, nil
	}
	if payResp.URLScheme != "" {
		return payResp.URLScheme, normalizedTradeNo, nil
	}

	if normalizedTradeNo != "" {
		baseURL := strings.TrimSuffix(mapiURL, "mapi.php") + "cashier.php"
		return fmt.Sprintf("%s?trade_no=%s", baseURL, normalizedTradeNo), normalizedTradeNo, nil
	}

	return "", normalizedTradeNo, fmt.Errorf("Payment interface did not return a usable payment link")
}

// Refund 发起易支付退款（V2 走 /api/pay/refund，V1 不支持）
func Refund(config *Config, orderNo, tradeNo, money, outRefundNo string) (*RefundResponse, error) {
	if config.IsV2() {
		return v2Refund(config, orderNo, tradeNo, money, outRefundNo)
	}
	return nil, fmt.Errorf("Epay V1 refund not supported, please use V2")
}

// ValidatePayType 验证支付方式是否在配置允许列表中
func ValidatePayType(config *Config, paymentType string) bool {
	for _, t := range config.PaymentTypes {
		if t == paymentType {
			return true
		}
	}
	return false
}
