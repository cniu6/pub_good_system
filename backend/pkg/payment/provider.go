// Package payment 提供多通道支付能力的基础接口与通用工具。
//
// 设计参考 go-wind-admin：每个支付通道作为独立 Provider 注册到全局注册表，
// 通过 channel_type 分发。通道密钥、版本、签名算法等运营配置统一存放在
// pay_gateways.ext_config（JSON）中，明文存储，不做脱敏/掩码。
package payment

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Provider 支付通道适配器接口。
// 具体实现放在各插件的 pay/<channel>/ 目录下，按 channel_type 注册。
type Provider interface {
	// Type 返回通道类型标识，如 "epay"
	Type() string

	// CreatePay 向远端创建支付订单，返回支付链接与平台交易号
	CreatePay(ctx context.Context, req *CreatePayRequest) (*CreatePayResponse, error)

	// VerifyNotify 校验异步回调签名
	// signType: 签名算法（如 MD5 / RSA），从通道配置取
	// extConfig: 通道扩展配置（含验签所需密钥）
	VerifyNotify(params map[string]string, signType string, extConfig map[string]string) bool

	// QueryOrder 向远端查询订单状态
	QueryOrder(ctx context.Context, req *QueryOrderRequest) (*QueryOrderResponse, error)

	// Refund 向远端发起退款（可选能力）
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// ValidatePayType 校验支付方式是否被该通道允许
	ValidatePayType(payType string, extConfig map[string]string) bool

	// TestConnection 测试通道配置是否可用
	TestConnection(ctx context.Context, extConfig map[string]string) (bool, string)
}

// NotifyPayload 回调验签后提取的归一化回调数据
type NotifyPayload struct {
	OutTradeNo  string // 系统订单号
	TradeNo     string // 第三方交易号
	TradeStatus string // 归一化交易状态
	Money       string // 金额（元）
	PayType     string // 支付方式
}

// PayloadVerifier 可选接口：针对需要 raw body + headers 验签的通道（WeChat V3 / Stripe / PayPal 等）。
// Provider 若实现此接口，回调控制器优先调用；否则回退到旧 VerifyNotify。
type PayloadVerifier interface {
	VerifyNotifyWithPayload(ctx context.Context, body []byte, headers map[string]string, signType string, extConfig map[string]string) (bool, *NotifyPayload, error)
}

// CreatePayRequest 创建支付请求参数
type CreatePayRequest struct {
	PID       string            // 商户ID
	ExtConfig map[string]string // 扩展配置（从 PayGateway.ext_config 解析）
	ApiURL    string            // 网关地址
	PayType   string            // 支付方式
	SignType  string            // 签名算法：MD5 / RSA
	Version   string            // 通道版本，如 v1 / v2 / v3
	Device    string            // 设备类型，如 pc / mobile
	OrderNo   string            // 系统订单号
	Subject   string            // 订单标题
	Money     string            // 金额（元，保留两位小数）
	Currency  string            // 币种，如 CNY / USD
	NotifyURL string            // 异步通知地址
	ReturnURL string            // 同步跳转地址
	ClientIP  string            // 客户端IP
	Param     string            // 业务扩展参数（防爆破，可选）
}

// CreatePayResponse 创建支付响应
type CreatePayResponse struct {
	PayURL    string // 支付链接
	TradeNo   string // 第三方交易号
	QRCode    string // 二维码内容
	URLScheme string // URL Scheme
}

// QueryOrderRequest 查询订单请求
type QueryOrderRequest struct {
	PID            string            // 商户ID
	ExtConfig      map[string]string // 扩展配置
	SignType       string            // 签名算法
	Version        string            // 通道版本
	ApiURL         string            // 网关地址
	OrderNo        string            // 系统订单号
	TradeNo        string            // 第三方交易号
	RequestTimeout int64             // 单次网关请求超时（秒），<=0 使用默认值
}

// QueryOrderResponse 查询订单响应
type QueryOrderResponse struct {
	Code        int
	Msg         string
	TradeNo     string
	OutTradeNo  string
	Type        string
	Name        string
	Money       string
	TradeStatus string
}

// RefundRequest 退款请求参数
type RefundRequest struct {
	PID       string            // 商户ID
	ExtConfig map[string]string // 扩展配置
	SignType  string            // 签名算法
	Version   string            // 通道版本
	ApiURL    string            // 网关地址
	OrderNo   string            // 系统订单号（二选一）
	TradeNo   string            // 第三方交易号（二选一，优先）
	Money     string            // 退款金额（元，保留两位小数）
}

// RefundResponse 退款响应
type RefundResponse struct {
	Code        int
	Msg         string
	RefundNo    string // 平台退款单号
	OutRefundNo string // 商户退款单号
	TradeNo     string // 平台订单号
	Money       string // 退款金额
}

// ParseExtConfig 解析 ext_config JSON 字符串为 map。
// 空字符串或解析失败均返回空 map（不返回 nil），方便调用方统一处理。
func ParseExtConfig(extConfigJSON string) map[string]string {
	result := make(map[string]string)
	s := strings.TrimSpace(extConfigJSON)
	if s == "" {
		return result
	}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return make(map[string]string)
	}
	return result
}

// MarshalExtConfig 将 map 序列化为 ext_config JSON 字符串。空 map 返回空字符串。
func MarshalExtConfig(extConfig map[string]string) string {
	if len(extConfig) == 0 {
		return ""
	}
	data, err := json.Marshal(extConfig)
	if err != nil {
		return ""
	}
	return string(data)
}

// NormalizeTradeStatus 归一化网关返回的交易状态为可比较值
func NormalizeTradeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "1", "SUCCESS", "SUCC", "PAID", "TRADE_SUCCESS", "FINISHED":
		return "TRADE_SUCCESS"
	case "0", "PENDING", "WAIT", "WAIT_BUYER_PAY", "UNPAID", "CREATED":
		return "PENDING"
	default:
		return strings.ToUpper(strings.TrimSpace(status))
	}
}

// ParseMoneyMinor 把 "元" 字符串解析为"分"整数，避免字符串精确比较导致金额误拒
func ParseMoneyMinor(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("money is empty")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("invalid money %q", value)
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, fmt.Errorf("invalid money %q", value)
	}
	fraction := "00"
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 {
		return 0, errors.New("money has more than two decimals")
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	cents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money %q", value)
	}
	if whole > (math.MaxInt64-cents)/100 {
		return 0, errors.New("money overflow")
	}
	return whole*100 + cents, nil
}

// FormatMoneyYuan 把"分"整数格式化为保留两位小数的"元"字符串
func FormatMoneyYuan(minor int64) string {
	return fmt.Sprintf("%.2f", float64(minor)/100.0)
}

// NormalizeTradeNo 规范化第三方交易号，过滤无意义占位符并截断
func NormalizeTradeNo(tradeNo string) string {
	trimmed := strings.TrimSpace(tradeNo)
	if trimmed == "" {
		return ""
	}
	normalized := strings.ToUpper(trimmed)
	replacer := strings.NewReplacer("_", "", "-", "", "/", "", "\\", "", " ", "", ":", "", ";", "", ".", "", "#", "")
	compact := replacer.Replace(normalized)
	switch compact {
	case "TRADENO", "OUTTRADENO", "NULL", "UNDEFINED", "NONE", "NIL", "NA":
		return ""
	default:
		if len(trimmed) > 64 {
			return trimmed[:64]
		}
		return trimmed
	}
}

// SignWithMD5 使用 MD5 对参数签名，末尾拼接 key
func SignWithMD5(params map[string]string, key string) string {
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" || strings.TrimSpace(v) == "" {
			continue
		}
		filtered[k] = v
	}
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

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
	hash := md5.Sum([]byte(buf.String()))
	return fmt.Sprintf("%x", hash)
}

// RSASign 使用商户 RSA 私钥对数据做 SHA256WithRSA 签名，返回 Base64
func RSASign(data, privateKeyPEM string) (string, error) {
	privateKey, err := parseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}
	hashed := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// HMACWithSHA256 使用 HMAC-SHA256 签名
func HMACWithSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// VerifyHMACWithSHA256 验证 HMAC-SHA256 签名
func VerifyHMACWithSHA256(data, sign, key string) bool {
	return hmac.Equal([]byte(sign), []byte(HMACWithSHA256(data, key)))
}

// RSAVerify 使用平台 RSA 公钥验证 SHA256WithRSA 签名
func RSAVerify(data, signatureBase64, publicKeyPEM string) bool {
	publicKey, err := parseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return false
	}
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}
	hashed := sha256.Sum256([]byte(data))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature); err != nil {
		return false
	}
	return true
}

func parseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = strings.TrimSpace(pemStr)
	if !strings.Contains(pemStr, "-----BEGIN") {
		pemStr = "-----BEGIN RSA PRIVATE KEY-----\n" + pemStr + "\n-----END RSA PRIVATE KEY-----"
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("PEM decode failed")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return rsaKey, nil
}

func parseRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	pemStr = strings.TrimSpace(pemStr)
	if !strings.Contains(pemStr, "-----BEGIN") {
		pemStr = "-----BEGIN PUBLIC KEY-----\n" + pemStr + "\n-----END PUBLIC KEY-----"
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("PEM decode failed")
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("failed to parse RSA public key")
}
