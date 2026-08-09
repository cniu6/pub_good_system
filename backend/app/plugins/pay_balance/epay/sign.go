package epay

import (
	"fst/backend/pkg/payment"
	"log"
	"sort"
	"strings"
)

// SignType 签名算法标识
const (
	SignTypeMD5 = "MD5"
	SignTypeRSA = "RSA"
)

// GenerateSign 生成易支付 MD5 签名（兼容旧接口：单 key 字符串）
// 用于测试和旧调用点；新代码请使用 GenerateSignWithConfig
func GenerateSign(params map[string]string, key string) string {
	return GenerateSignWithConfig(params, SignTypeMD5, map[string]string{"key": key})
}

// GenerateSignWithConfig 生成易支付签名（创建支付订单，包含 type）
// 按 signType 分发 MD5 / RSA：
//   - MD5: 参数排序拼接 key=value&...，末尾拼 merchant_key，取 MD5
//   - RSA: 参数排序拼接（不含末尾密钥），用 merchant_private_key 做 SHA256WithRSA，Base64
func GenerateSignWithConfig(params map[string]string, signType string, extConfig map[string]string) string {
	return generateSignInternal(params, signType, extConfig, false)
}

// GenerateNotifySign 生成易支付回调验签用 MD5 签名（兼容旧接口：单 key）
func GenerateNotifySign(params map[string]string, key string) string {
	return GenerateNotifySignWithConfig(params, SignTypeMD5, map[string]string{"key": key})
}

// GenerateNotifySignWithConfig 生成易支付回调验签用签名（排除 type）
// 回调签名规则：sign、sign_type、type 和空值不参与签名
func GenerateNotifySignWithConfig(params map[string]string, signType string, extConfig map[string]string) string {
	return generateSignInternal(params, signType, extConfig, true)
}

// generateSignInternal 内部签名
// excludeType=true 时排除 type（回调验签），false 时包含（发起支付）
func generateSignInternal(params map[string]string, signType string, extConfig map[string]string, excludeType bool) string {
	if extConfig == nil {
		extConfig = make(map[string]string)
	}

	switch strings.ToUpper(strings.TrimSpace(signType)) {
	case SignTypeRSA:
		return generateRSASign(params, extConfig, excludeType)
	default:
		return generateMD5Sign(params, extConfig, excludeType)
	}
}

// generateMD5Sign MD5 签名
func generateMD5Sign(params map[string]string, extConfig map[string]string, excludeType bool) string {
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" || strings.TrimSpace(v) == "" {
			continue
		}
		if excludeType && k == "type" {
			continue
		}
		filtered[k] = v
	}

	key := extConfig["merchant_key"]
	if key == "" {
		key = extConfig["key"]
	}
	return payment.SignWithMD5(filtered, key)
}

// generateRSASign RSA 签名：参数排序拼接（不含末尾密钥），用商户私钥签名
func generateRSASign(params map[string]string, extConfig map[string]string, excludeType bool) string {
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" || strings.TrimSpace(v) == "" {
			continue
		}
		if excludeType && k == "type" {
			continue
		}
		filtered[k] = v
	}

	privateKey := extConfig["merchant_private_key"]
	if privateKey == "" {
		privateKey = extConfig["private_key"]
	}

	sig, err := payment.RSASign(buildParamString(filtered), privateKey)
	if err != nil {
		log.Printf("[Epay] RSA sign failed: %v", err)
		return ""
	}
	return sig
}

// buildParamString 按 key ASCII 升序拼接 key=value&...
func buildParamString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
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
		sb.WriteString(params[k])
	}
	return sb.String()
}

// VerifySign 验证易支付回调签名（兼容旧接口：单 key 字符串）
func VerifySign(params map[string]string, key string) bool {
	return VerifySignWithConfig(params, SignTypeMD5, map[string]string{"key": key})
}

// VerifySignWithConfig 验证易支付回调签名（兼容包含/不包含 type 的不同实现）
// MD5: 用 merchant_key 重新签名比对
// RSA: 用 platform_public_key 验证 SHA256WithRSA 签名
func VerifySignWithConfig(params map[string]string, signType string, extConfig map[string]string) bool {
	sign, ok := params["sign"]
	if !ok || sign == "" {
		return false
	}

	expectedNotify := GenerateNotifySignWithConfig(params, signType, extConfig)
	if strings.EqualFold(sign, expectedNotify) {
		return true
	}
	expectedGeneric := GenerateSignWithConfig(params, signType, extConfig)
	if strings.EqualFold(sign, expectedGeneric) {
		return true
	}

	safeOrder := params["out_trade_no"]
	log.Printf("[Epay] sign mismatch: order_no=%s, sign_type=%s, received_sign_len=%d, expected_notify_prefix=%s..., expected_generic_prefix=%s...",
		safeOrder, signType, len(sign), trimSignPrefix(expectedNotify), trimSignPrefix(expectedGeneric))
	return false
}

// trimSignPrefix 日志里只保留签名前几位，降低泄露风险
func trimSignPrefix(sign string) string {
	if len(sign) <= 6 {
		return "***"
	}
	return sign[:6]
}

// FormatSignType 规范化签名算法标识，未知值兜底 MD5
func FormatSignType(signType string) string {
	s := strings.ToUpper(strings.TrimSpace(signType))
	switch s {
	case SignTypeRSA:
		return SignTypeRSA
	default:
		return SignTypeMD5
	}
}

// GetSignKeyForQuery 取查单用的 key：MD5 用 merchant_key，RSA 用 merchant_private_key
func GetSignKeyForQuery(extConfig map[string]string) string {
	if extConfig == nil {
		return ""
	}
	key := extConfig["merchant_key"]
	if key == "" {
		key = extConfig["merchant_private_key"]
	}
	if key == "" {
		key = extConfig["key"]
	}
	return key
}
