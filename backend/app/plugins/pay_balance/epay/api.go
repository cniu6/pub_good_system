package epay

import (
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
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
	PID          string
	Key          string
	PaymentTypes []string
}

// ConfigFromGateway 从支付通道模型构建易支付配置
func ConfigFromGateway(gateway *models.PayGateway) *Config {
	if gateway == nil {
		return &Config{}
	}
	return &Config{
		ApiURL:       strings.TrimRight(gateway.ApiURL, "/"),
		PID:          gateway.PID,
		Key:          gateway.Key,
		PaymentTypes: []string{gateway.PayType},
	}
}

// BuildSubmitURL 构造易支付跳转支付 URL（submit.php）
func BuildSubmitURL(config *Config, order *models.PaymentOrder, notifyURL, returnURL string) (string, error) {
	if config.ApiURL == "" || config.PID == "" || config.Key == "" {
		return "", fmt.Errorf("易支付配置不完整")
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

	params["sign"] = GenerateSign(params, config.Key)
	params["sign_type"] = "MD5"

	u, err := url.Parse(config.ApiURL + "/submit.php")
	if err != nil {
		return "", fmt.Errorf("解析网关地址失败: %w", err)
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
	if config.ApiURL == "" || config.PID == "" || config.Key == "" {
		return nil, fmt.Errorf("易支付配置不完整")
	}

	orderNo = strings.TrimSpace(orderNo)
	tradeNo = models.NormalizeTradeNo(tradeNo)
	if orderNo == "" && tradeNo == "" {
		return nil, fmt.Errorf("缺少查询订单号")
	}

	queryCandidates := make([]map[string]string, 0, 2)
	if orderNo != "" {
		queryCandidates = append(queryCandidates, map[string]string{
			"act":          "order",
			"pid":          config.PID,
			"key":          config.Key,
			"out_trade_no": orderNo,
		})
	}
	if tradeNo != "" {
		queryCandidates = append(queryCandidates, map[string]string{
			"act":      "order",
			"pid":      config.PID,
			"key":      config.Key,
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
		return nil, fmt.Errorf("解析网关地址失败: %w", err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("查询易支付订单失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		return nil, fmt.Errorf("查询接口返回HTML页面而非JSON，可能是API地址配置错误")
	}

	var raw queryResponseRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("[Epay] 查询响应解析失败: %s", string(body))
		return nil, fmt.Errorf("解析查询响应失败: %w", err)
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

// APIPay 通过 mapi.php 发起支付，返回支付链接与交易号
func APIPay(config *Config, order *models.PaymentOrder, notifyURL, returnURL string) (string, string, error) {
	if config.ApiURL == "" || config.PID == "" || config.Key == "" {
		return "", "", fmt.Errorf("易支付配置不完整")
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
	}

	params["sign"] = GenerateSign(params, config.Key)
	params["sign_type"] = "MD5"

	mapiURL := config.ApiURL + "/mapi.php"
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(mapiURL, formData)
	if err != nil {
		return "", "", fmt.Errorf("请求支付接口失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("支付接口返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应内容失败: %v", err)
	}

	if len(body) == 0 {
		return "", "", fmt.Errorf("支付接口返回空响应")
	}

	bodyStr := string(body)
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "<") {
		return "", "", fmt.Errorf("支付接口返回HTML页面而非JSON，可能是API地址配置错误")
	}

	log.Printf("[Epay] APIPay 响应: %s", bodyStr)

	var payResp APIPayResponse
	if err := json.Unmarshal(body, &payResp); err != nil {
		return "", "", fmt.Errorf("解析支付响应失败: %v, 响应内容: %s", err, bodyStr)
	}

	if payResp.Code != 1 {
		return "", "", fmt.Errorf("发起支付失败: %s", payResp.Msg)
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

	return "", normalizedTradeNo, fmt.Errorf("支付接口未返回可用的支付链接")
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
