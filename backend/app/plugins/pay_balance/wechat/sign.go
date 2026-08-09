package wechat

import (
	"crypto/md5"
	"fmt"
	"fst/backend/pkg/payment"
	"sort"
	"strings"
)

// SignTypeMD5 微信支付 V2 MD5 签名
const SignTypeMD5 = "MD5"

// SignTypeHMACSHA256 微信支付 V2 HMAC-SHA256 签名
const SignTypeHMACSHA256 = "HMAC-SHA256"

// BuildParamString 按 key ASCII 升序拼接 key=value&...，空值与 sign 不参与
func BuildParamString(params map[string]string) string {
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || strings.TrimSpace(v) == "" {
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
		sb.WriteString(filtered[k])
	}
	return sb.String()
}

// SignWithV2 微信支付 V2 签名：MD5 或 HMAC-SHA256
func SignWithV2(params map[string]string, signType, apiKey string) string {
	data := BuildParamString(params)
	data += "&key=" + strings.TrimSpace(apiKey)
	switch strings.ToUpper(strings.TrimSpace(signType)) {
	case SignTypeHMACSHA256:
		return payment.HMACWithSHA256(data, apiKey)
	default:
		hash := md5.Sum([]byte(data))
		return fmt.Sprintf("%x", hash)
	}
}

// VerifyWithV2 验证微信支付 V2 回调签名
func VerifyWithV2(params map[string]string, signType, apiKey string) bool {
	sign := strings.TrimSpace(params["sign"])
	if sign == "" {
		return false
	}
	expected := SignWithV2(params, signType, apiKey)
	return strings.EqualFold(sign, expected)
}

// FormatSignType 规范化微信支付签名算法
func FormatSignType(signType string) string {
	s := strings.ToUpper(strings.TrimSpace(signType))
	switch s {
	case SignTypeHMACSHA256:
		return SignTypeHMACSHA256
	default:
		return SignTypeMD5
	}
}
