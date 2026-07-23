package models

import (
	"strings"
	"unicode/utf8"
)

// 库表字段上限（与 utils / gorm size 对齐；放在 models 内避免与 utils 互相 import）。
const (
	storedIPMaxLen     = 64
	storedBrowserIDLen = 64
	storedDeviceLen    = 128
	storedPathLen      = 255
	storedModuleLen    = 100
	storedActionLen    = 100
	storedMethodLen    = 20
	storedSceneLen     = 32
)

// clampBytes 按字节截断，保证 UTF-8 完整，不追加标记。
func clampBytes(s string, max int) string {
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

// clampStoredIP 截断到 storedIPMaxLen，避免 varchar 溢出。
func clampStoredIP(ip string) string {
	return clampBytes(ip, storedIPMaxLen)
}

func clampBrowserID(id string) string {
	return clampBytes(id, storedBrowserIDLen)
}

func clampDevice(device string) string {
	return clampBytes(device, storedDeviceLen)
}

func clampPath(path string) string {
	return clampBytes(path, storedPathLen)
}
