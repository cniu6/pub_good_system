package middleware

import (
	"unicode/utf8"
)

// MySQL 文本类型容量（字节）：
//
//	TEXT       ≈ 65,535
//	MEDIUMTEXT ≈ 16,777,215（约 16MB）
//	LONGTEXT   ≈ 4GB
//
// 本项目 operation_logs / api_access_logs 的 request_body、response_body 均为 MEDIUMTEXT。
// 存库前统一截断到远小于上限的安全值，避免撑爆数据库与拖慢详情接口。
// 注意：按产品要求，访问日志 / 操作日志不做字段级脱敏，仅做长度截断。
const (
	// mysqlMediumTextMaxBytes MySQL MEDIUMTEXT 理论上限（字节）
	mysqlMediumTextMaxBytes = 16_777_215
	// maxLogStoredBodyBytes 实际写入 request_body / response_body 的上限（64KB）
	// 约为 MEDIUMTEXT 的 0.4%，留足余量（含 utf8mb4 与截断标记）
	maxLogStoredBodyBytes = 64 * 1024
	// maxLogReadableBodyBytes 读取请求体时的上限，超过则不读入内存
	maxLogReadableBodyBytes = 64 * 1024
	// maxLogStoredHeadersLength 请求头 JSON 存库上限
	maxLogStoredHeadersLength = 8 * 1024
	// maxLogHeaderValueLength 单个请求头值截断长度
	maxLogHeaderValueLength = 500
	// maxLogIPLength 单 IP 字段上限（与 models 中 size:64 对齐；覆盖 IPv6 + zone id）
	maxLogIPLength = 64
	// maxLogPathLength path / route_path 字段上限（与 models size:255 对齐）
	maxLogPathLength = 255
	// maxLogXForwardedForLength X-Forwarded-For 代理链上限（与 models size:1024 对齐）
	maxLogXForwardedForLength = 1024
	// logTruncateMarker 截断标记（追加在截断内容末尾）
	logTruncateMarker = "...(已截断)"
)

// 避免未使用常量告警：保留与 MEDIUMTEXT 对照的文档常量。
var _ = mysqlMediumTextMaxBytes

// truncateForLog 按字节上限截断，避免切断 UTF-8 多字节字符，并追加「已截断」标记。
// 最终长度严格不超过 limit（标记占用 limit 内的空间）。
func truncateForLog(raw string, limit int) string {
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	markerLen := len(logTruncateMarker)
	if markerLen >= limit {
		// limit 不足以放下完整标记时，尽量返回截断后的标记前缀
		cut := limit
		for cut > 0 && !utf8.RuneStart(logTruncateMarker[cut]) {
			cut--
		}
		if cut <= 0 {
			return ""
		}
		return logTruncateMarker[:cut]
	}
	// 为截断标记预留空间，确保最终长度不超过 limit
	cut := limit - markerLen
	if cut > len(raw) {
		cut = len(raw)
	}
	// 回退到完整 rune 边界，避免乱码
	for cut > 0 && !utf8.RuneStart(raw[cut]) {
		cut--
	}
	if cut <= 0 {
		return logTruncateMarker[:min(markerLen, limit)]
	}
	return raw[:cut] + logTruncateMarker
}
