package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	// 中国大陆手机号：1 开头第二位 3-9，共 11 位
	reMobileCN = regexp.MustCompile(`^1[3-9]\d{9}$`)
	// 国际 E.164：+ 开头，国家码不以 0 开头，总位数 8~15（不含 +）
	reMobileE164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
)

// ErrInvalidMobileCN 仅允许中国大陆手机号时的错误
var ErrInvalidMobileCN = errors.New("仅支持中国大陆手机号（+86）")

// ErrInvalidMobile 国际模式下格式非法
var ErrInvalidMobile = errors.New("手机号格式不正确")

// NormalizePhoneInput 去掉空格、横线、括号等常见分隔符，保留开头的 +
func NormalizePhoneInput(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(raw))
	for i, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		if r == '+' && i == 0 && b.Len() == 0 {
			b.WriteRune(r)
		}
		// 忽略空格、横线、括号等
	}
	return b.String()
}

// IsMobileCNOnlyDefault 未配置时的默认策略：仅中国大陆
const IsMobileCNOnlyDefault = true

// NormalizeAndValidateMobile 按策略规范化并校验手机号。
//
// cnOnly=true（默认）：
//   - 接受 138xxxx、86138xxxx、+86138xxxx
//   - 统一存成 11 位国内号（不含国家码）
//
// cnOnly=false：
//   - 接受 E.164（如 +14155552671）
//   - 国内 11 位会规范成 +86xxxxxxxxxxx
//   - 统一存成带 + 的 E.164
func NormalizeAndValidateMobile(raw string, cnOnly bool) (string, error) {
	s := NormalizePhoneInput(raw)
	if s == "" {
		return "", ErrInvalidMobile
	}

	if cnOnly {
		return normalizeCNMobile(s)
	}
	return normalizeInternationalMobile(s)
}

func normalizeCNMobile(s string) (string, error) {
	// +86 / 86 前缀剥掉
	if strings.HasPrefix(s, "+86") {
		s = s[3:]
	} else if strings.HasPrefix(s, "86") && len(s) == 13 {
		s = s[2:]
	}
	// 去掉国内常见的 0086
	if strings.HasPrefix(s, "0086") && len(s) == 15 {
		s = s[4:]
	}
	if !reMobileCN.MatchString(s) {
		return "", ErrInvalidMobileCN
	}
	return s, nil
}

func normalizeInternationalMobile(s string) (string, error) {
	// 已是 E.164
	if strings.HasPrefix(s, "+") {
		if !reMobileE164.MatchString(s) {
			return "", ErrInvalidMobile
		}
		return s, nil
	}
	// 国内 11 位 → +86（关「仅大陆」后仍可便捷输入本地号）
	// 注意：与「1 + 10位北美号」存在歧义，北美号必须写 +1...
	if reMobileCN.MatchString(s) {
		return "+86" + s, nil
	}
	// 86 + 11 位
	if strings.HasPrefix(s, "86") && len(s) == 13 && reMobileCN.MatchString(s[2:]) {
		return "+" + s, nil
	}
	// 0086 + 11 位
	if strings.HasPrefix(s, "0086") && len(s) == 15 && reMobileCN.MatchString(s[4:]) {
		return "+86" + s[4:], nil
	}
	// 其余无 + 的纯数字：不自动猜国家码（避免 1415... 被误加成 +86）
	return "", ErrInvalidMobile
}

// IsValidMobile 仅校验，不返回规范化结果
func IsValidMobile(raw string, cnOnly bool) bool {
	_, err := NormalizeAndValidateMobile(raw, cnOnly)
	return err == nil
}

// MobileLookupVariants 同一号码在库中可能出现的写法（CN 11 位 / +86 E.164），用于唯一性查询。
// 避免开关切换后「138xxxx」与「+86138xxxx」被当成两个号。
func MobileLookupVariants(normalized string) []string {
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return nil
	}
	seen := map[string]struct{}{normalized: {}}
	out := []string{normalized}

	add := func(v string) {
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	if strings.HasPrefix(normalized, "+86") && len(normalized) == 14 && reMobileCN.MatchString(normalized[3:]) {
		add(normalized[3:])
	} else if reMobileCN.MatchString(normalized) {
		add("+86" + normalized)
	}
	return out
}

// FormatPhoneForSMS 短信通道拨号格式：去掉前导 +（阿里云/腾讯等通常要 86138... 或 138...）
func FormatPhoneForSMS(normalized string) string {
	return strings.TrimPrefix(strings.TrimSpace(normalized), "+")
}
