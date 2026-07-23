package models

import (
	"errors"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const operationLogAggregateGlobalKey = "global"

// OperationLogStat 操作日志全局汇总
type OperationLogStat struct {
	StatKey          string `gorm:"column:stat_key;primaryKey;size:32"`
	TotalCount       int64  `gorm:"column:total_count;not null;default:0"`
	SuccessCount     int64  `gorm:"column:success_count;not null;default:0"`
	ClientErrorCount int64  `gorm:"column:client_error_count;not null;default:0"`
	ServerErrorCount int64  `gorm:"column:server_error_count;not null;default:0"`
	TotalDuration    int64  `gorm:"column:total_duration;not null;default:0"`
	UpdatedAt        int64  `gorm:"column:updated_at;not null;default:0"`
}

func (OperationLogStat) TableName() string { return "operation_log_stats" }

// OperationLogDailyStat 操作日志按天汇总
type OperationLogDailyStat struct {
	DayKey     int   `gorm:"column:day_key;primaryKey"`
	TotalCount int64 `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64 `gorm:"column:updated_at;not null;default:0"`
}

func (OperationLogDailyStat) TableName() string { return "operation_log_daily_stats" }

// OperationLogModuleStatRow 操作日志按模块汇总
type OperationLogModuleStatRow struct {
	Module     string `gorm:"column:module;primaryKey;size:100"`
	TotalCount int64  `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64  `gorm:"column:updated_at;not null;default:0"`
}

func (OperationLogModuleStatRow) TableName() string { return "operation_log_module_stats" }

// OperationLogActionStatRow 操作日志按动作汇总
type OperationLogActionStatRow struct {
	Action     string `gorm:"column:action;primaryKey;size:100"`
	TotalCount int64  `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64  `gorm:"column:updated_at;not null;default:0"`
}

func (OperationLogActionStatRow) TableName() string { return "operation_log_action_stats" }

// OperationLogMethodStatRow 操作日志按 HTTP 方法汇总
type OperationLogMethodStatRow struct {
	Method     string `gorm:"column:method;primaryKey;size:20"`
	TotalCount int64  `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64  `gorm:"column:updated_at;not null;default:0"`
}

func (OperationLogMethodStatRow) TableName() string { return "operation_log_method_stats" }

type operationLogAggregateGlobal struct {
	TotalCount       int64 `gorm:"column:total_count"`
	SuccessCount     int64 `gorm:"column:success_count"`
	ClientErrorCount int64 `gorm:"column:client_error_count"`
	ServerErrorCount int64 `gorm:"column:server_error_count"`
	TotalDuration    int64 `gorm:"column:total_duration"`
}

type operationLogAggregateDailyRow struct {
	DayKey     int   `gorm:"column:day_key"`
	TotalCount int64 `gorm:"column:total_count"`
}

type operationLogAggregateModuleRow struct {
	Module     string `gorm:"column:module"`
	TotalCount int64  `gorm:"column:total_count"`
}

type operationLogAggregateActionRow struct {
	Action     string `gorm:"column:action"`
	TotalCount int64  `gorm:"column:total_count"`
}

type operationLogAggregateMethodRow struct {
	Method     string `gorm:"column:method"`
	TotalCount int64  `gorm:"column:total_count"`
}

// OperationActionStat 操作日志按动作统计项
type OperationActionStat struct {
	Action string `gorm:"column:action" json:"action"`
	Count  int64  `gorm:"column:count" json:"count"`
}

// OperationLogStatsDetail 操作日志详细统计（读聚合表，清理明细不影响累计）
type OperationLogStatsDetail struct {
	TotalCount       int64                 `json:"total_count"`
	TodayCount       int64                 `json:"today_count"`
	SuccessCount     int64                 `json:"success_count"`
	ClientErrorCount int64                 `json:"client_error_count"`
	ServerErrorCount int64                 `json:"server_error_count"`
	AvgDuration      float64               `json:"avg_duration"`
	TopModules       []ModuleStat          `json:"top_modules"`
	TopActions       []OperationActionStat `json:"top_actions"`
	MethodStats      []MethodStat          `json:"method_stats"`
}

// BackfillOperationLogAggregateIfNeeded 聚合表由 migrate.RunAutoMigrate 建表，仅回填历史数据
func BackfillOperationLogAggregateIfNeeded() {
	if !db.CheckTableExists("operation_logs") || !db.CheckTableExists("operation_log_stats") {
		return
	}

	var existing int64
	if err := db.DB.Model(&OperationLogStat{}).Where("stat_key = ?", operationLogAggregateGlobalKey).Count(&existing).Error; err != nil {
		log.Printf("[Init] Failed to check operation log aggregate data: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	if err := db.WithTx(rebuildOperationLogAggregate); err != nil {
		log.Printf("[Init] Failed to backfill operation log aggregate data: %v", err)
	}
}

func rebuildOperationLogAggregate(tx *gorm.DB) error {
	for _, m := range []interface{}{
		&OperationLogStat{}, &OperationLogDailyStat{},
		&OperationLogModuleStatRow{}, &OperationLogActionStatRow{}, &OperationLogMethodStatRow{},
	} {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Where("1=1").Delete(m).Error; err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	var global operationLogAggregateGlobal
	if err := tx.Raw(`SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 ELSE 0 END), 0) AS client_error_count,
    COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_error_count,
    COALESCE(SUM(duration), 0) AS total_duration
    FROM operation_logs`).Scan(&global).Error; err != nil {
		return err
	}

	if err := tx.Create(&OperationLogStat{
		StatKey: operationLogAggregateGlobalKey,
		TotalCount: global.TotalCount, SuccessCount: global.SuccessCount,
		ClientErrorCount: global.ClientErrorCount, ServerErrorCount: global.ServerErrorCount,
		TotalDuration: global.TotalDuration, UpdatedAt: now,
	}).Error; err != nil {
		return err
	}

	var dailyRows []aggregateDailyRow
	if rows, err := scanDailyCountsFromUnixColumn(tx, "operation_logs", "create_time", resolveOperationLogDayKey); err != nil {
		return err
	} else {
		dailyRows = rows
	}
	for _, row := range dailyRows {
		if row.DayKey <= 0 {
			continue
		}
		if err := tx.Create(&OperationLogDailyStat{DayKey: row.DayKey, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var moduleRows []operationLogAggregateModuleRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(module, ''), 'unknown') AS module, COUNT(*) AS total_count FROM operation_logs GROUP BY COALESCE(NULLIF(module, ''), 'unknown') ORDER BY total_count DESC`).Scan(&moduleRows).Error; err != nil {
		return err
	}
	for _, row := range moduleRows {
		if err := tx.Create(&OperationLogModuleStatRow{Module: row.Module, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var actionRows []operationLogAggregateActionRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(action, ''), 'unknown') AS action, COUNT(*) AS total_count FROM operation_logs GROUP BY COALESCE(NULLIF(action, ''), 'unknown') ORDER BY total_count DESC`).Scan(&actionRows).Error; err != nil {
		return err
	}
	for _, row := range actionRows {
		if err := tx.Create(&OperationLogActionStatRow{Action: row.Action, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var methodRows []operationLogAggregateMethodRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS total_count FROM operation_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY total_count DESC`).Scan(&methodRows).Error; err != nil {
		return err
	}
	for _, row := range methodRows {
		if err := tx.Create(&OperationLogMethodStatRow{Method: row.Method, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	return nil
}

// RecordOperationLogAggregate 写入一条操作日志后增量更新聚合统计（明细清理不会回减）
func RecordOperationLogAggregate(item *OperationLog) error {
	if item == nil {
		return nil
	}

	return db.WithTx(func(tx *gorm.DB) error {
		createTime := time.Now().Unix()
		if item.CreateTime != nil && *item.CreateTime > 0 {
			createTime = *item.CreateTime
		}
		updatedAt := time.Now().Unix()
		dayKey := resolveOperationLogDayKey(createTime)
		module := resolveOperationLogAggregateModule(item.Module)
		action := resolveOperationLogAggregateAction(item.Action)
		method := resolveOperationLogAggregateMethod(item.Method)

		successCount := int64(0)
		clientErrorCount := int64(0)
		serverErrorCount := int64(0)
		switch {
		case item.StatusCode >= 200 && item.StatusCode < 400:
			successCount = 1
		case item.StatusCode >= 400 && item.StatusCode < 500:
			clientErrorCount = 1
		case item.StatusCode >= 500:
			serverErrorCount = 1
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "stat_key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_count":        gorm.Expr("total_count + 1"),
				"success_count":      gorm.Expr("success_count + ?", successCount),
				"client_error_count": gorm.Expr("client_error_count + ?", clientErrorCount),
				"server_error_count": gorm.Expr("server_error_count + ?", serverErrorCount),
				"total_duration":     gorm.Expr("total_duration + ?", item.Duration),
				"updated_at":         updatedAt,
			}),
		}).Create(&OperationLogStat{
			StatKey: operationLogAggregateGlobalKey, TotalCount: 1,
			SuccessCount: successCount, ClientErrorCount: clientErrorCount,
			ServerErrorCount: serverErrorCount, TotalDuration: int64(item.Duration), UpdatedAt: updatedAt,
		}).Error; err != nil {
			return err
		}

		if err := upsertDailyTotal(tx, "operation_log_daily_stats", dayKey, updatedAt); err != nil {
			return err
		}
		if err := upsertKeyedTotal(tx, "operation_log_module_stats", "module", module, updatedAt); err != nil {
			return err
		}
		if err := upsertKeyedTotal(tx, "operation_log_action_stats", "action", action, updatedAt); err != nil {
			return err
		}
		return upsertKeyedTotal(tx, "operation_log_method_stats", "method", method, updatedAt)
	})
}

// GetOperationLogStatsDetail 获取操作日志详细统计（优先聚合表，缺失时回退扫明细）
func GetOperationLogStatsDetail() (*OperationLogStatsDetail, error) {
	return getOperationLogStatsFromAggregate()
}

func getOperationLogStatsFromAggregate() (*OperationLogStatsDetail, error) {
	if !db.CheckTableExists("operation_log_stats") || !db.CheckTableExists("operation_log_daily_stats") || !db.CheckTableExists("operation_log_module_stats") || !db.CheckTableExists("operation_log_action_stats") || !db.CheckTableExists("operation_log_method_stats") {
		return getOperationLogStatsFromLogsFallback()
	}

	stats := &OperationLogStatsDetail{
		TopModules:  []ModuleStat{},
		TopActions:  []OperationActionStat{},
		MethodStats: []MethodStat{},
	}

	var global operationLogAggregateGlobal
	err := db.DB.Model(&OperationLogStat{}).
		Select("total_count, success_count, client_error_count, server_error_count, total_duration").
		Where("stat_key = ?", operationLogAggregateGlobalKey).
		First(&global).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return getOperationLogStatsFromLogsFallback()
	}
	if err != nil {
		return nil, err
	}

	stats.TotalCount = global.TotalCount
	stats.SuccessCount = global.SuccessCount
	stats.ClientErrorCount = global.ClientErrorCount
	stats.ServerErrorCount = global.ServerErrorCount
	if global.TotalCount > 0 {
		stats.AvgDuration = float64(global.TotalDuration) / float64(global.TotalCount)
	}

	todayKey := resolveOperationLogDayKey(time.Now().Unix())
	var daily OperationLogDailyStat
	if err := db.DB.Select("total_count").Where("day_key = ?", todayKey).First(&daily).Error; err == nil {
		stats.TodayCount = daily.TotalCount
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := db.DB.Raw(`SELECT module, total_count AS count FROM operation_log_module_stats ORDER BY total_count DESC, module ASC LIMIT 10`).Scan(&stats.TopModules).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT action, total_count AS count FROM operation_log_action_stats ORDER BY total_count DESC, action ASC LIMIT 10`).Scan(&stats.TopActions).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT method, total_count AS count FROM operation_log_method_stats ORDER BY total_count DESC, method ASC`).Scan(&stats.MethodStats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func getOperationLogStatsFromLogsFallback() (*OperationLogStatsDetail, error) {
	stats := &OperationLogStatsDetail{
		TopModules:  []ModuleStat{},
		TopActions:  []OperationActionStat{},
		MethodStats: []MethodStat{},
	}
	todayStart := resolveOperationLogStartOfLocalDay(time.Now()).Unix()

	if err := db.DB.Raw("SELECT COUNT(*) FROM operation_logs").Scan(&stats.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", todayStart).Scan(&stats.TodayCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM operation_logs WHERE status_code >= 200 AND status_code < 400").Scan(&stats.SuccessCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM operation_logs WHERE status_code >= 400 AND status_code < 500").Scan(&stats.ClientErrorCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM operation_logs WHERE status_code >= 500").Scan(&stats.ServerErrorCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COALESCE(AVG(duration), 0) FROM operation_logs").Scan(&stats.AvgDuration).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(module, ''), 'unknown') AS module, COUNT(*) AS count FROM operation_logs GROUP BY COALESCE(NULLIF(module, ''), 'unknown') ORDER BY count DESC LIMIT 10`).Scan(&stats.TopModules).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(action, ''), 'unknown') AS action, COUNT(*) AS count FROM operation_logs GROUP BY COALESCE(NULLIF(action, ''), 'unknown') ORDER BY count DESC LIMIT 10`).Scan(&stats.TopActions).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS count FROM operation_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY count DESC`).Scan(&stats.MethodStats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func resolveOperationLogDayKey(ts int64) int {
	day := resolveOperationLogStartOfLocalDay(time.Unix(ts, 0).In(time.Local))
	return day.Year()*10000 + int(day.Month())*100 + day.Day()
}

func resolveOperationLogStartOfLocalDay(target time.Time) time.Time {
	local := target.In(time.Local)
	year, month, day := local.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, local.Location())
}

func resolveOperationLogAggregateModule(module string) string {
	normalized := strings.TrimSpace(module)
	if normalized == "" {
		return "unknown"
	}
	return clampBytes(normalized, storedModuleLen)
}

func resolveOperationLogAggregateAction(action string) string {
	normalized := strings.TrimSpace(action)
	if normalized == "" {
		return "unknown"
	}
	return clampBytes(normalized, storedActionLen)
}

func resolveOperationLogAggregateMethod(method string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" {
		return "UNKNOWN"
	}
	return clampBytes(normalized, storedMethodLen)
}
