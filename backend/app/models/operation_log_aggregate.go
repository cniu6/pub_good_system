package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const operationLogAggregateGlobalKey = "global"

type operationLogAggregateGlobal struct {
	TotalCount       int64 `db:"total_count"`
	SuccessCount     int64 `db:"success_count"`
	ClientErrorCount int64 `db:"client_error_count"`
	ServerErrorCount int64 `db:"server_error_count"`
	TotalDuration    int64 `db:"total_duration"`
}

type operationLogAggregateDailyRow struct {
	DayKey     int   `db:"day_key"`
	TotalCount int64 `db:"total_count"`
}

type operationLogAggregateModuleRow struct {
	Module     string `db:"module"`
	TotalCount int64  `db:"total_count"`
}

type operationLogAggregateActionRow struct {
	Action     string `db:"action"`
	TotalCount int64  `db:"total_count"`
}

type operationLogAggregateMethodRow struct {
	Method     string `db:"method"`
	TotalCount int64  `db:"total_count"`
}

// OperationActionStat 操作日志按动作统计项
type OperationActionStat struct {
	Action string `db:"action" json:"action"`
	Count  int64  `db:"count" json:"count"`
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

// InitOperationLogAggregateTables 初始化操作日志聚合表
func InitOperationLogAggregateTables() {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS operation_log_stats (
      stat_key VARCHAR(32) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      success_count BIGINT NOT NULL DEFAULT 0,
      client_error_count BIGINT NOT NULL DEFAULT 0,
      server_error_count BIGINT NOT NULL DEFAULT 0,
      total_duration BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志累计汇总';`,
		`CREATE TABLE IF NOT EXISTS operation_log_daily_stats (
      day_key INT UNSIGNED NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志按天汇总';`,
		`CREATE TABLE IF NOT EXISTS operation_log_module_stats (
      module VARCHAR(100) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志按模块汇总';`,
		`CREATE TABLE IF NOT EXISTS operation_log_action_stats (
      action VARCHAR(100) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志按动作汇总';`,
		`CREATE TABLE IF NOT EXISTS operation_log_method_stats (
      method VARCHAR(20) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志按方法汇总';`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			log.Printf("[Init] Failed to create operation log aggregate table: %v", err)
		}
	}

	backfillOperationLogAggregateIfNeeded()
}

func backfillOperationLogAggregateIfNeeded() {
	// 日报查询统一经 db.Q 适配；SQLite 也需要回填已有日志，避免切换或重启后统计为空。
	if !db.CheckTableExists("operation_logs") || !db.CheckTableExists("operation_log_stats") {
		return
	}

	var existing int
	if err := db.DB.Get(&existing, "SELECT COUNT(*) FROM operation_log_stats WHERE stat_key = ?", operationLogAggregateGlobalKey); err != nil {
		log.Printf("[Init] Failed to check operation log aggregate data: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Printf("[Init] Failed to begin operation log aggregate backfill: %v", err)
		return
	}

	if err := rebuildOperationLogAggregate(tx); err != nil {
		_ = tx.Rollback()
		log.Printf("[Init] Failed to backfill operation log aggregate data: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[Init] Failed to commit operation log aggregate backfill: %v", err)
	}
}

func rebuildOperationLogAggregate(tx *sqlx.Tx) error {
	tables := []string{
		"operation_log_stats",
		"operation_log_daily_stats",
		"operation_log_module_stats",
		"operation_log_action_stats",
		"operation_log_method_stats",
	}
	for _, tableName := range tables {
		if _, err := tx.Exec("DELETE FROM " + tableName); err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	var global operationLogAggregateGlobal
	if err := tx.Get(&global, `SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 ELSE 0 END), 0) AS client_error_count,
    COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_error_count,
    COALESCE(SUM(duration), 0) AS total_duration
    FROM operation_logs`); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO operation_log_stats (stat_key, total_count, success_count, client_error_count, server_error_count, total_duration, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)`,
		operationLogAggregateGlobalKey,
		global.TotalCount,
		global.SuccessCount,
		global.ClientErrorCount,
		global.ServerErrorCount,
		global.TotalDuration,
		now,
	); err != nil {
		return err
	}

	var dailyRows []operationLogAggregateDailyRow
	if err := tx.Select(&dailyRows, db.Q(`SELECT CAST(DATE_FORMAT(FROM_UNIXTIME(create_time), '%Y%m%d') AS UNSIGNED) AS day_key, COUNT(*) AS total_count FROM operation_logs GROUP BY day_key ORDER BY day_key ASC`)); err != nil {
		return err
	}
	for _, row := range dailyRows {
		if row.DayKey <= 0 {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO operation_log_daily_stats (day_key, total_count, updated_at) VALUES (?, ?, ?)`, row.DayKey, row.TotalCount, now); err != nil {
			return err
		}
	}

	var moduleRows []operationLogAggregateModuleRow
	if err := tx.Select(&moduleRows, `SELECT COALESCE(NULLIF(module, ''), 'unknown') AS module, COUNT(*) AS total_count FROM operation_logs GROUP BY COALESCE(NULLIF(module, ''), 'unknown') ORDER BY total_count DESC`); err != nil {
		return err
	}
	for _, row := range moduleRows {
		if _, err := tx.Exec(`INSERT INTO operation_log_module_stats (module, total_count, updated_at) VALUES (?, ?, ?)`, row.Module, row.TotalCount, now); err != nil {
			return err
		}
	}

	var actionRows []operationLogAggregateActionRow
	if err := tx.Select(&actionRows, `SELECT COALESCE(NULLIF(action, ''), 'unknown') AS action, COUNT(*) AS total_count FROM operation_logs GROUP BY COALESCE(NULLIF(action, ''), 'unknown') ORDER BY total_count DESC`); err != nil {
		return err
	}
	for _, row := range actionRows {
		if _, err := tx.Exec(`INSERT INTO operation_log_action_stats (action, total_count, updated_at) VALUES (?, ?, ?)`, row.Action, row.TotalCount, now); err != nil {
			return err
		}
	}

	var methodRows []operationLogAggregateMethodRow
	if err := tx.Select(&methodRows, `SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS total_count FROM operation_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY total_count DESC`); err != nil {
		return err
	}
	for _, row := range methodRows {
		if _, err := tx.Exec(`INSERT INTO operation_log_method_stats (method, total_count, updated_at) VALUES (?, ?, ?)`, row.Method, row.TotalCount, now); err != nil {
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

	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	createTime := time.Now().Unix()
	if item.CreateTime != nil && *item.CreateTime > 0 {
		createTime = *item.CreateTime
	}
	updatedAt := time.Now().Unix()
	dayKey := resolveOperationLogDayKey(createTime)
	module := resolveOperationLogAggregateModule(item.Module)
	action := resolveOperationLogAggregateAction(item.Action)
	method := resolveOperationLogAggregateMethod(item.Method)

	successCount := 0
	clientErrorCount := 0
	serverErrorCount := 0
	switch {
	case item.StatusCode >= 200 && item.StatusCode < 400:
		successCount = 1
	case item.StatusCode >= 400 && item.StatusCode < 500:
		clientErrorCount = 1
	case item.StatusCode >= 500:
		serverErrorCount = 1
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO operation_log_stats (stat_key, total_count, success_count, client_error_count, server_error_count, total_duration, updated_at)
    VALUES (?, 1, ?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      success_count = success_count + ?,
      client_error_count = client_error_count + ?,
      server_error_count = server_error_count + ?,
      total_duration = total_duration + ?,
      updated_at = ?`),
		operationLogAggregateGlobalKey,
		successCount,
		clientErrorCount,
		serverErrorCount,
		item.Duration,
		updatedAt,
		successCount,
		clientErrorCount,
		serverErrorCount,
		item.Duration,
		updatedAt,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO operation_log_daily_stats (day_key, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`),
		dayKey,
		updatedAt,
		updatedAt,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO operation_log_module_stats (module, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`),
		module,
		updatedAt,
		updatedAt,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO operation_log_action_stats (action, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`),
		action,
		updatedAt,
		updatedAt,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO operation_log_method_stats (method, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`),
		method,
		updatedAt,
		updatedAt,
	); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
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
	if err := db.DB.Get(&global, `SELECT total_count, success_count, client_error_count, server_error_count, total_duration FROM operation_log_stats WHERE stat_key = ?`, operationLogAggregateGlobalKey); err != nil {
		if err == sql.ErrNoRows {
			return getOperationLogStatsFromLogsFallback()
		}
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
	if err := db.DB.Get(&stats.TodayCount, `SELECT total_count FROM operation_log_daily_stats WHERE day_key = ?`, todayKey); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopModules, `SELECT module, total_count AS count FROM operation_log_module_stats ORDER BY total_count DESC, module ASC LIMIT 10`); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopActions, `SELECT action, total_count AS count FROM operation_log_action_stats ORDER BY total_count DESC, action ASC LIMIT 10`); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.MethodStats, `SELECT method, total_count AS count FROM operation_log_method_stats ORDER BY total_count DESC, method ASC`); err != nil {
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

	if err := db.DB.Get(&stats.TotalCount, "SELECT COUNT(*) FROM operation_logs"); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.TodayCount, "SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", todayStart); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.SuccessCount, "SELECT COUNT(*) FROM operation_logs WHERE status_code >= 200 AND status_code < 400"); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.ClientErrorCount, "SELECT COUNT(*) FROM operation_logs WHERE status_code >= 400 AND status_code < 500"); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.ServerErrorCount, "SELECT COUNT(*) FROM operation_logs WHERE status_code >= 500"); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.AvgDuration, "SELECT COALESCE(AVG(duration), 0) FROM operation_logs"); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopModules, `SELECT COALESCE(NULLIF(module, ''), 'unknown') AS module, COUNT(*) AS count FROM operation_logs GROUP BY COALESCE(NULLIF(module, ''), 'unknown') ORDER BY count DESC LIMIT 10`); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopActions, `SELECT COALESCE(NULLIF(action, ''), 'unknown') AS action, COUNT(*) AS count FROM operation_logs GROUP BY COALESCE(NULLIF(action, ''), 'unknown') ORDER BY count DESC LIMIT 10`); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.MethodStats, `SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS count FROM operation_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY count DESC`); err != nil {
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
	return normalized
}

func resolveOperationLogAggregateAction(action string) string {
	normalized := strings.TrimSpace(action)
	if normalized == "" {
		return "unknown"
	}
	return normalized
}

func resolveOperationLogAggregateMethod(method string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" {
		return "UNKNOWN"
	}
	return normalized
}
