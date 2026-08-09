package alipay

import (
	"encoding/json"
	"fst/backend/pkg/payment"
	"log"
	"net/url"
	"sort"
	"strings"
)

// SignTypeRSA2 支付宝 RSA2 签名算法（即 SHA256WithRSA）
const SignTypeRSA2 = "RSA2"

// BuildParamString 按 key ASCII 升序拼接 key=value&...，并对 value 做 URL 编码
// 空值、sign、sign_type 不参与签名
func BuildParamString(params map[string]string) string {
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

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(url.QueryEscape(filtered[k]))
	}
	return sb.String()
}

// SignWithRSA2 使用商户 RSA 私钥对参数签名字符串做 RSA2 签名，返回 Base64
func SignWithRSA2(params map[string]string, privateKeyPEM string) (string, error) {
	data := BuildParamString(params)
	return payment.RSASign(data, privateKeyPEM)
}

// VerifyWithRSA2 使用支付宝公钥验证 RSA2 签名
func VerifyWithRSA2(params map[string]string, sign string, publicKeyPEM string) bool {
	sign = strings.TrimSpace(sign)
	if sign == "" {
		return false
	}
	data := BuildParamString(params)
	return payment.RSAVerify(data, sign, publicKeyPEM)
}

// trimSignPrefix 日志里只保留签名前几位，降低泄露风险
func trimSignPrefix(sign string) string {
	if len(sign) <= 6 {
		return "***"
	}
	return sign[:6]
}

// LogVerifyFailure 打印验签失败日志（不包含完整签名）
func LogVerifyFailure(sign, expected string) {
	log.Printf("[Alipay] sign mismatch received=%s... expected=%s...", trimSignPrefix(sign), trimSignPrefix(expected))
}

// FormatSignType 规范化签名算法标识，未知值兜底 RSA2
func FormatSignType(signType string) string {
	s := strings.ToUpper(strings.TrimSpace(signType))
	switch s {
	case SignTypeRSA2:
		return SignTypeRSA2
	default:
		return SignTypeRSA2
	}
}

// JSONContent 把对象序列化为 JSON 字符串；失败返回空字符串
func JSONContent(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
