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
	"strconv"
	"strings"
	"time"
)

// v2CreateOrder V2 创建订单：POST /api/pay/create
// 统一走 RSA 签名，需带 timestamp
func v2CreateOrder(config *Config, order *models.PaymentOrder, notifyURL, returnURL string) (string, string, error) {
	if config.ApiURL == "" || config.PID == "" {
		return "", "", fmt.Errorf("Epay V2 configuration incomplete")
	}

	privateKey := config.ExtConfig["merchant_private_key"]
	if privateKey == "" {
		privateKey = config.ExtConfig["private_key"]
	}
	if privateKey == "" {
		return "", "", fmt.Errorf("Epay V2 merchant_private_key missing")
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
		"timestamp":    strconv.FormatInt(time.Now().Unix(), 10),
		"sign_type":    SignTypeRSA,
	}
	params["sign"] = generateV2RSASign(params, config.ExtConfig)

	createURL := strings.TrimRight(config.ApiURL, "/") + "/api/pay/create"
	return postFormAndParsePayURL(createURL, params)
}

// v2QueryOrder V2 查单：GET /api/pay/query
func v2QueryOrder(config *Config, orderNo, tradeNo string) (*QueryResponse, error) {
	if config.ApiURL == "" || config.PID == "" {
		return nil, fmt.Errorf("Epay V2 configuration incomplete")
	}

	privateKey := config.ExtConfig["merchant_private_key"]
	if privateKey == "" {
		privateKey = config.ExtConfig["private_key"]
	}
	if privateKey == "" {
		return nil, fmt.Errorf("Epay V2 merchant_private_key missing")
	}

	params := map[string]string{
		"pid":       config.PID,
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		"sign_type": SignTypeRSA,
	}
	if orderNo != "" {
		params["out_trade_no"] = orderNo
	}
	if tradeNo != "" {
		params["trade_no"] = tradeNo
	}
	if len(params) == 3 {
		return nil, fmt.Errorf("Missing query order number")
	}
	params["sign"] = generateV2RSASign(params, config.ExtConfig)

	queryURL := strings.TrimRight(config.ApiURL, "/") + "/api/pay/query"
	u, err := url.Parse(queryURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("Epay V2 query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		return nil, fmt.Errorf("Epay V2 query returned HTML")
	}

	var raw queryResponseRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("[Epay] V2 query parse failed: %s", bodyStr)
		return nil, fmt.Errorf("Failed to parse V2 query response: %w", err)
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

// v2Refund V2 退款：POST /api/pay/refund
func v2Refund(config *Config, orderNo, tradeNo, money, outRefundNo string) (*RefundResponse, error) {
	if config.ApiURL == "" || config.PID == "" {
		return nil, fmt.Errorf("Epay V2 configuration incomplete")
	}

	privateKey := config.ExtConfig["merchant_private_key"]
	if privateKey == "" {
		privateKey = config.ExtConfig["private_key"]
	}
	if privateKey == "" {
		return nil, fmt.Errorf("Epay V2 merchant_private_key missing")
	}

	params := map[string]string{
		"pid":          config.PID,
		"money":        money,
		"timestamp":    strconv.FormatInt(time.Now().Unix(), 10),
		"sign_type":    SignTypeRSA,
		"out_refund_no": outRefundNo,
	}
	if orderNo != "" {
		params["out_trade_no"] = orderNo
	}
	if tradeNo != "" {
		params["trade_no"] = tradeNo
	}
	if orderNo == "" && tradeNo == "" {
		return nil, fmt.Errorf("Epay V2 refund requires order number")
	}

	params["sign"] = generateV2RSASign(params, config.ExtConfig)

	refundURL := strings.TrimRight(config.ApiURL, "/") + "/api/pay/refund"
	body, err := postForm(refundURL, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code        int    `json:"code"`
		Msg         string `json:"msg"`
		RefundNo    string `json:"refund_no"`
		OutRefundNo string `json:"out_refund_no"`
		TradeNo     string `json:"trade_no"`
		Money       string `json:"money"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Failed to parse V2 refund response: %w", err)
	}
	if resp.Code != 0 && resp.Code != 1 {
		return nil, fmt.Errorf("Epay V2 refund failed: %s", resp.Msg)
	}

	return &RefundResponse{
		Code:        resp.Code,
		Msg:         resp.Msg,
		RefundNo:    resp.RefundNo,
		OutRefundNo: resp.OutRefundNo,
		TradeNo:     resp.TradeNo,
		Money:       resp.Money,
	}, nil
}

// postForm 发送 application/x-www-form-urlencoded POST 请求并返回响应 body
func postForm(apiURL string, params map[string]string) ([]byte, error) {
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(apiURL, formData)
	if err != nil {
		return nil, fmt.Errorf("Epay V2 request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		return nil, fmt.Errorf("Epay V2 returned HTML instead of JSON, URL may be misconfigured")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Epay V2 returned status %d: %s", resp.StatusCode, bodyStr)
	}
	return body, nil
}

// postFormAndParsePayURL 发送 POST 表单并解析支付链接
func postFormAndParsePayURL(apiURL string, params map[string]string) (string, string, error) {
	body, err := postForm(apiURL, params)
	if err != nil {
		return "", "", err
	}

	var payResp APIPayResponse
	if err := json.Unmarshal(body, &payResp); err != nil {
		return "", "", fmt.Errorf("Failed to parse V2 payment response: %w", err)
	}
	if payResp.Code != 1 && payResp.Code != 0 {
		return "", "", fmt.Errorf("Epay V2 payment failed: %s", payResp.Msg)
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
	return "", normalizedTradeNo, fmt.Errorf("Epay V2 did not return a usable payment link")
}

// parseMoneyMinorToYuan 把分整数格式化为元（两位小数），退款用
func parseMoneyMinorToYuan(minor int64) string {
	return payment.FormatMoneyYuan(minor)
}
