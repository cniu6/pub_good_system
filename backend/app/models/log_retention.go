package models

import (
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

const logRetentionDeleteBatchSize = 200

// log_retention.go：email_logs / sms_logs / operation_logs / api_access_logs 四类日志的
// 「保留清理」逻辑高度重复，这里抽成通用实现。

// cleanExcessRowsGeneric 全局保留最新 maxCount 条（按 timeCol 排序），小批删除更旧的记录。
// 先取 id 再按主键删除，不用 DELETE ... NOT IN (SELECT ...)，缩短锁持有时间，避免与高频写日志竞争。
func cleanExcessRowsGeneric(model any, timeCol string, maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var total int64
	if err := db.DB.Model(model).Count(&total).Error; err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	var ids []uint64
	deleteCount := min(int64(logRetentionDeleteBatchSize), total-int64(maxCount))
	if err := db.DB.Model(model).Order(timeCol+" ASC, id ASC").Limit(int(deleteCount)).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	return deleteRowsByID(model, ids)
}

// cleanExcessRowsPerGroupGeneric 按 groupCol 分组，每组最多保留 maxPerGroup 条，多的按时间清理。
func cleanExcessRowsPerGroupGeneric[G any](model any, groupCol, timeCol string, maxPerGroup int, extraWhere string) (int64, error) {
	if maxPerGroup <= 0 {
		return 0, nil
	}
	scoped := func() *gorm.DB {
		q := db.DB.Model(model)
		if strings.TrimSpace(extraWhere) != "" {
			q = q.Where(extraWhere)
		}
		return q
	}

	var groups []struct {
		Grp G     `gorm:"column:grp"`
		Cnt int64 `gorm:"column:cnt"`
	}
	if err := scoped().Select(groupCol+" AS grp, COUNT(*) AS cnt").Group(groupCol).
		Having("COUNT(*) > ?", maxPerGroup).Scan(&groups).Error; err != nil {
		return 0, err
	}

	var totalAffected int64
	for _, g := range groups {
		var ids []uint64
		deleteCount := min(int64(logRetentionDeleteBatchSize), g.Cnt-int64(maxPerGroup))
		if err := scoped().Where(groupCol+" = ?", g.Grp).Order(timeCol+" ASC, id ASC").
			Limit(int(deleteCount)).Pluck("id", &ids).Error; err != nil {
			log.Printf("[LogRetention] 查询待清理记录失败 model=%T group_col=%s grp=%v: %v", model, groupCol, g.Grp, err)
			continue
		}
		n, err := deleteRowsByID(model, ids)
		if err != nil {
			log.Printf("[LogRetention] 按分组清理超限记录失败 model=%T group_col=%s grp=%v: %v", model, groupCol, g.Grp, err)
			continue
		}
		totalAffected += n
	}
	return totalAffected, nil
}

// deleteRowsByID 按主键物理删除（日志清理不走软删）。
func deleteRowsByID(model any, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := db.DB.Unscoped().Where("id IN ?", ids).Delete(model)
	return result.RowsAffected, result.Error
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
