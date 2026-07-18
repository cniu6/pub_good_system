package epay

import (
	"crypto/md5"
	"fmt"
	"log"
	"sort"
	"strings"
)

// GenerateSign 生成易支付 MD5 签名（创建支付订单，包含 type）
// 规则：参数按 key ASCII 升序，拼接 key=value&...，末尾拼密钥，取 MD5
func GenerateSign(params map[string]string, key string) string {
	return generateSignInternal(params, key, false)
}

// GenerateNotifySign 生成易支付回调验签用 MD5（排除 type）
// 回调签名规则：sign、sign_type、type 和空值不参与签名
func GenerateNotifySign(params map[string]string, key string) string {
	return generateSignInternal(params, key, true)
}

// generateSignInternal 内部签名
// excludeType=true 时排除 type（回调验签），false 时包含（发起支付）
func generateSignInternal(params map[string]string, key string, excludeType bool) string {
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

// VerifySign 验证易支付回调签名（兼容包含/不包含 type 的不同实现）
func VerifySign(params map[string]string, key string) bool {
	sign, ok := params["sign"]
	if !ok || sign == "" {
		return false
	}
	expectedNotify := GenerateNotifySign(params, key)
	if strings.EqualFold(sign, expectedNotify) {
		return true
	}
	expectedGeneric := GenerateSign(params, key)
	if strings.EqualFold(sign, expectedGeneric) {
		return true
	}
	log.Printf("[Epay] 签名不匹配: received=%s, expected_notify=%s, expected_generic=%s, params=%v",
		sign, expectedNotify, expectedGeneric, params)
	return false
}
