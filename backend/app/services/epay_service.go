package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// EpayConfig 易支付配置（从 pay_gateways 表的单条记录构建）
type EpayConfig struct {
	Enabled      bool     // 易支付通道是否启用
	ApiURL       string   // 易支付网关地址
	PID          string   // 商户ID
	Key          string   // 商户密钥
	PaymentTypes []string // 支持的支付方式
}

// GenerateEpaySign 生成易支付MD5签名（用于创建支付订单，包含 type 参数）
// 规则：将参数按 key 的 ASCII 升序排列，拼接 key=value&...，末尾拼接密钥，取 MD5
func GenerateEpaySign(params map[string]string, key string) string {
	return generateEpaySignInternal(params, key, false)
}

// GenerateEpayNotifySign 生成易支付回调验签用MD5签名（排除 type 参数）
// 易支付回调签名规则：sign、sign_type、type 和空值不参与签名
func GenerateEpayNotifySign(params map[string]string, key string) string {
	return generateEpaySignInternal(params, key, true)
}

// generateEpaySignInternal 内部签名生成函数
// excludeType=true 时排除 type 参数（用于回调验签），false 时包含（用于发起支付）
func generateEpaySignInternal(params map[string]string, key string, excludeType bool) string {
	// 过滤空值和签名相关字段
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		if excludeType && k == "type" {
			continue
		}
		filtered[k] = v
	}

	// 按 key ASCII 升序排列
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(filtered[k])
	}
	buf.WriteString(strings.TrimSpace(key))

	// MD5
	hash := md5.Sum([]byte(buf.String()))
	return fmt.Sprintf("%x", hash)
}

// VerifyEpaySign 验证易支付回调签名（兼容包含/不包含 type 的不同实现）
func VerifyEpaySign(params map[string]string, key string) bool {
	sign, ok := params["sign"]
	if !ok || sign == "" {
		return false
	}
	expectedNotify := GenerateEpayNotifySign(params, key)
	if strings.EqualFold(sign, expectedNotify) {
		return true
	}
	expectedGeneric := GenerateEpaySign(params, key)
	if strings.EqualFold(sign, expectedGeneric) {
		return true
	}
	maskedKey := key
	if len(maskedKey) > 4 {
		maskedKey = maskedKey[:2] + "***" + maskedKey[len(maskedKey)-2:]
	}
	log.Printf("[Epay] 签名不匹配: received=%s, expected_notify=%s, expected_generic=%s, key=%s, params=%v", sign, expectedNotify, expectedGeneric, maskedKey, params)
	return false
}

// EpaySubmitParams 易支付提交参数
type EpaySubmitParams struct {
	PID        string // 商户ID
	Type       string // 支付方式
	OutTradeNo string // 商户订单号
	NotifyURL  string // 异步通知地址
	ReturnURL  string // 同步跳转地址
	Name       string // 商品名称
	Money      string // 金额
	Sign       string // 签名
	SignType   string // 签名类型
}

// BuildEpaySubmitURL 构造易支付跳转支付URL
func BuildEpaySubmitURL(config *EpayConfig, order *models.PaymentOrder, notifyURL, returnURL string) (string, error) {
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

	sign := GenerateEpaySign(params, config.Key)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	// 构造完整URL
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

// EpayQueryResponse 易支付查询响应
type EpayQueryResponse struct {
	Code        int    `json:"code"`
	TradeNo     string `json:"trade_no"`
	OutTradeNo  string `json:"out_trade_no"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Money       string `json:"money"`
	TradeStatus string `json:"trade_status"`
}

// QueryEpayOrder 向易支付平台查询订单状态
func QueryEpayOrder(config *EpayConfig, tradeNo string) (*EpayQueryResponse, error) {
	if config.ApiURL == "" || config.PID == "" || config.Key == "" {
		return nil, fmt.Errorf("易支付配置不完整")
	}

	params := map[string]string{
		"act":      "order",
		"pid":      config.PID,
		"trade_no": tradeNo,
	}
	sign := GenerateEpaySign(params, config.Key)
	params["sign"] = sign
	params["sign_type"] = "MD5"

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

	var result EpayQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[Epay] 查询响应解析失败: %s", string(body))
		return nil, fmt.Errorf("解析查询响应失败: %w", err)
	}

	return &result, nil
}

// ValidatePaymentType 验证支付方式是否被允许
func ValidatePaymentType(config *EpayConfig, paymentType string) bool {
	for _, t := range config.PaymentTypes {
		if t == paymentType {
			return true
		}
	}
	return false
}

// EpayAPIPayResponse API支付响应
type EpayAPIPayResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	PayURL    string `json:"payurl"`
	QRCode    string `json:"qrcode"`
	URLScheme string `json:"urlscheme"`
	TradeNo   string `json:"trade_no"`
}

// EpayAPIPay 通过API接口发起支付（mapi.php），返回支付链接
func EpayAPIPay(config *EpayConfig, order *models.PaymentOrder, notifyURL, returnURL string) (string, error) {
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
		"clientip":     order.ClientIP,
	}

	sign := GenerateEpaySign(params, config.Key)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	mapiURL := config.ApiURL + "/mapi.php"
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(mapiURL, formData)
	if err != nil {
		return "", fmt.Errorf("请求支付接口失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("支付接口返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应内容失败: %v", err)
	}

	if len(body) == 0 {
		return "", fmt.Errorf("支付接口返回空响应")
	}

	bodyStr := string(body)
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "<") {
		return "", fmt.Errorf("支付接口返回HTML页面而非JSON，可能是API地址配置错误")
	}

	log.Printf("[Epay] APIPay 响应: %s", bodyStr)

	var payResp EpayAPIPayResponse
	if err := json.Unmarshal(body, &payResp); err != nil {
		return "", fmt.Errorf("解析支付响应失败: %v, 响应内容: %s", err, bodyStr)
	}

	if payResp.Code != 1 {
		return "", fmt.Errorf("发起支付失败: %s", payResp.Msg)
	}

	normalizedTradeNo := models.NormalizeTradeNo(payResp.TradeNo)

	// 更新订单交易号（过滤掉远程API可能返回的占位符值）
	if normalizedTradeNo != "" {
		models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusPending, normalizedTradeNo)
		order.TradeNo = normalizedTradeNo
	}

	// 优先返回支付链接
	if payResp.PayURL != "" {
		return payResp.PayURL, nil
	}
	if payResp.QRCode != "" {
		return payResp.QRCode, nil
	}
	if payResp.URLScheme != "" {
		return payResp.URLScheme, nil
	}

	// 构建 cashier 链接
	if normalizedTradeNo != "" {
		baseURL := strings.TrimSuffix(mapiURL, "mapi.php") + "cashier.php"
		return fmt.Sprintf("%s?trade_no=%s", baseURL, normalizedTradeNo), nil
	}

	return "", fmt.Errorf("支付接口未返回可用的支付链接")
}
