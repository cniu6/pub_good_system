package utils

import (
	"errors"
	"time"
)

// NormalizePagination 统一分页参数默认值/上限处理：page<=0 → 1；pageSize<=0 → 20；
// pageSize 超过 100 → 裁剪到 100。
func NormalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// ErrInvalidTimeRange 开始时间晚于结束时间
var ErrInvalidTimeRange = errors.New("Start time cannot be later than end time")

// NormalizeTimeRange 统一「结束时间默认当前、开始时间默认结束时间往前 N 天」的时间窗口解析，
// 附带 defaultDays 上限裁剪。多个日志类 controller（operation/api_access 等）各自复制了一份
// 几乎一样的逻辑，这里统一。start > end 时返回 ErrInvalidTimeRange，调用方按需转 utils.Fail。
func NormalizeTimeRange(startTime, endTime int64, defaultDays, maxDays int) (int64, int64, error) {
	if defaultDays <= 0 {
		defaultDays = 30
	}
	if maxDays > 0 && defaultDays > maxDays {
		defaultDays = maxDays
	}

	now := time.Now().Unix()
	if endTime <= 0 {
		endTime = now
	}
	if startTime <= 0 {
		startTime = endTime - int64(defaultDays*24*60*60)
	}
	if startTime > endTime {
		return startTime, endTime, ErrInvalidTimeRange
	}
	return startTime, endTime, nil
}
