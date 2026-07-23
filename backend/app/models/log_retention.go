package models

import (
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// log_retention.go：email_logs / sms_logs / operation_logs / api_access_logs 四类日志的
// 「保留清理」逻辑高度重复，这里抽成通用实现。

// cleanExcessRowsGeneric 全局保留最新 maxCount 条（按 timeCol 排序），删除更旧的记录。
func cleanExcessRowsGeneric(table, timeCol string, maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var total int64
	if err := db.DB.Table(table).Count(&total).Error; err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	deleteSQL := fmt.Sprintf(
		"DELETE FROM %s WHERE id NOT IN (SELECT id FROM (SELECT id FROM %s ORDER BY %s DESC, id DESC LIMIT ?) AS keep_rows)",
		table, table, timeCol,
	)
	result := db.DB.Exec(deleteSQL, maxCount)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// cleanExcessRowsPerGroupGeneric 按 groupCol 分组，每组最多保留 maxPerGroup 条，多的按时间清理。
func cleanExcessRowsPerGroupGeneric[G any](table, groupCol, timeCol string, maxPerGroup int, extraWhere string) (int64, error) {
	if maxPerGroup <= 0 {
		return 0, nil
	}

	groupWhere := ""
	if strings.TrimSpace(extraWhere) != "" {
		groupWhere = " WHERE " + extraWhere
	}

	var groups []struct {
		Grp G     `gorm:"column:grp"`
		Cnt int64 `gorm:"column:cnt"`
	}
	groupSQL := fmt.Sprintf("SELECT %s AS grp, COUNT(*) AS cnt FROM %s%s GROUP BY %s HAVING COUNT(*) > ?", groupCol, table, groupWhere, groupCol)
	if err := db.DB.Raw(groupSQL, maxPerGroup).Scan(&groups).Error; err != nil {
		return 0, err
	}

	deleteSQL := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ? AND id NOT IN (SELECT id FROM (SELECT id FROM %s WHERE %s = ? ORDER BY %s DESC, id DESC LIMIT ?) AS keep_rows)",
		table, groupCol, table, groupCol, timeCol,
	)

	var totalAffected int64
	for _, g := range groups {
		result := db.DB.Exec(deleteSQL, g.Grp, g.Grp, maxPerGroup)
		if result.Error != nil {
			log.Printf("[LogRetention] 按分组清理超限记录失败 table=%s group_col=%s grp=%v: %v", table, groupCol, g.Grp, result.Error)
			continue
		}
		if result.RowsAffected > 0 {
			totalAffected += result.RowsAffected
		}
	}
	return totalAffected, nil
}

// logRetentionConfig 三个通用配置项：全局上限 / 是否启用按用户（收件人）上限 / 按用户上限。
type logRetentionConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

// loadLogRetentionConfigGeneric 按 key 前缀读取日志保留配置
func loadLogRetentionConfigGeneric(keyPrefix string) logRetentionConfig {
	cfg := logRetentionConfig{MaxCount: 1000, PerUserMaxCount: 1000}
	settingsMap, err := GetSettingsMap([]string{
		keyPrefix + "_max_count",
		keyPrefix + "_per_user_limit_enabled",
		keyPrefix + "_per_user_max_count",
	})
	if err != nil {
		return cfg
	}
	if v, ok := settingsMap[keyPrefix+"_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.MaxCount = n
		}
	}
	if v, ok := settingsMap[keyPrefix+"_per_user_limit_enabled"]; ok {
		lower := strings.ToLower(strings.TrimSpace(v))
		cfg.PerUserLimitEnabled = lower == "true" || lower == "1"
	}
	if v, ok := settingsMap[keyPrefix+"_per_user_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.PerUserMaxCount = n
		}
	}
	return cfg
}

// scheduleLogRetentionCleanupGeneric 通用节流调度：30 秒内只真正触发一次清理
func scheduleLogRetentionCleanupGeneric(
	nextAt *atomic.Int64,
	keyPrefix, logTag string,
	cleanGlobal func(maxCount int) (int64, error),
	cleanPerUser func(maxPerUser int) (int64, error),
) {
	now := time.Now().UnixNano()
	prev := nextAt.Load()
	if prev > now {
		return
	}
	if !nextAt.CompareAndSwap(prev, now+int64(30*time.Second)) {
		return
	}

	cfg := loadLogRetentionConfigGeneric(keyPrefix)
	if cfg.MaxCount > 0 {
		if _, err := cleanGlobal(cfg.MaxCount); err != nil {
			log.Printf("[%s] 自动清理超限日志失败: %v", logTag, err)
		}
	}
	if cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0 {
		if _, err := cleanPerUser(cfg.PerUserMaxCount); err != nil {
			log.Printf("[%s] 按用户/收件人清理超限日志失败: %v", logTag, err)
		}
	}
}
