package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// 库表 varchar 字段上限（与 models gorm size 对齐；系统写入静默截断、用户输入校验拒绝共用）。
const (
	MaxStoredIPLength      = 64
	MaxPathLength          = 255
	MaxBrowserIDLength     = 64
	MaxDeviceLength        = 128
	MaxMemoLength          = 255
	MaxUsernameLength      = 64
	MaxNicknameLength      = 64
	MaxEmailLength         = 255
	MaxAdminRemarkLength   = 255
	MaxCountryLength       = 64
	MaxLanguageLength      = 20
	MaxSubjectLength       = 255
	MaxDescriptionLength   = 255
	MaxSignNameLength      = 64
	MaxAnnouncementTitle   = 200
	MaxAnnouncementSummary = 255
	MaxCommentLength       = 500
	MaxResolveRemarkLength = 500
	MaxRejectReasonLength  = 255
	MaxJobNameLength       = 128
	MaxCronExprLength      = 128
	MaxTimezoneLength      = 64
	MaxParamsJSONLength    = 64 * 1024
	MaxSettingKeyLength    = 100
	MaxSettingLabelLength  = 100
	MaxCertImageURLLength  = 500
	MaxTradeNoLength       = 64
	MaxOrderNoLength       = 64
	MaxXForwardedForLength = 1024
	MaxModuleLength        = 100
	MaxActionLength        = 100
	MaxMethodLength        = 20
	MaxSceneLength         = 32
)

// ClampBytes 按字节截断到 max，保证 UTF-8 完整，不追加标记（供 IP/path/browser_id 等库字段静默截断）。
func ClampBytes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || s == "" {
		return s
	}
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	if cut <= 0 {
		return ""
	}
	return s[:cut]
}

// ValidateRuneLen 按 rune 计数校验长度；超长返回「字段不能超过N个字符」。
// 调用方在服务层可用 NewClientError(err.Error())，控制器可直接 Fail(400, err.Error())。
func ValidateRuneLen(s, fieldName string, max int) error {
	s = strings.TrimSpace(s)
	if max <= 0 {
		return nil
	}
	if utf8.RuneCountInString(s) > max {
		return fmt.Errorf("%s不能超过%d个字符", fieldName, max)
	}
	return nil
}

// ClampStoredIP 截断到 MaxStoredIPLength，避免写入 varchar 溢出（strict 模式会报错）。
func ClampStoredIP(ip string) string {
	return ClampBytes(ip, MaxStoredIPLength)
}
