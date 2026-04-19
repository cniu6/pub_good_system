package sms_plugin

import "strings"

func normalizeSMSLang(lang string) string {
	lang = strings.TrimSpace(lang)
	switch strings.ToLower(lang) {
	case "en", "en-us":
		return "en-US"
	case "zh", "zh-cn", "":
		return "zh-CN"
	default:
		return lang
	}
}

