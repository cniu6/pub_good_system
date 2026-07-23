package utils

import (
	"strings"
)

// MaskCertificateNo 对身份证/护照等证件号码做掩码处理，仅保留首尾用于人工核对，
// 其余统一替换为 *，避免用户端/登录响应等场景明文回传完整证件号。
//
// 掩码规则（按 rune 数，兼容护照等含字母场景，各档均保证至少留 1 位掩码）：
//   - 空串：原样返回
//   - 长度 <= 4：仅保留最后 1 位，其余掩码（信息量已很小，避免过度暴露）
//   - 长度 5~10（护照号等较短证件）：保留首尾各 1 位
//   - 长度 > 10（典型 18 位身份证）：保留前 6 位（地区码）+ 后 4 位，中间掩码
func MaskCertificateNo(no string) string {
	no = strings.TrimSpace(no)
	if no == "" {
		return ""
	}
	runes := []rune(no)
	n := len(runes)

	switch {
	case n <= 4:
		return strings.Repeat("*", n-1) + string(runes[n-1])
	case n <= 10:
		return string(runes[0]) + strings.Repeat("*", n-2) + string(runes[n-1])
	default:
		const prefixLen, suffixLen = 6, 4
		maskLen := n - prefixLen - suffixLen
		return string(runes[:prefixLen]) + strings.Repeat("*", maskLen) + string(runes[n-suffixLen:])
	}
}

// MaskAccountNo 对银行卡/收款账号做掩码，默认列表展示用；打款时再按需取明文。
func MaskAccountNo(no string) string {
	no = strings.TrimSpace(no)
	if no == "" {
		return ""
	}
	runes := []rune(no)
	n := len(runes)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	if n <= 8 {
		return string(runes[0]) + strings.Repeat("*", n-2) + string(runes[n-1])
	}
	return string(runes[:4]) + strings.Repeat("*", n-8) + string(runes[n-4:])
}
