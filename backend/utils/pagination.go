package utils

import (
	"errors"
	"time"
)

// NormalizePagination 统一分页参数默认值/上限处理：page<=0 → 1；pageSize<=0 → 默认值；
// pageSize 超过上限 → 裁剪到上限。项目里 10+ 个 admin/user controller 各自手写了同一套
// 「page 默认 1、page_size 默认 20、上限 100」逻辑，这里统一收敛，避免有的地方漏加上限。
func NormalizePagination(page, pageSize int) (int, int) {
	return NormalizePaginationWithLimits(page, pageSize, 20, 100)
}

// NormalizePaginationWithLimits 同 NormalizePagination，但允许自定义默认每页数量与上限
// （maxSize <= 0 表示不限制上限）。
func NormalizePaginationWithLimits(page, pageSize, defaultSize, maxSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultSize
	}
	if maxSize > 0 && pageSize > maxSize {
		pageSize = maxSize
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
