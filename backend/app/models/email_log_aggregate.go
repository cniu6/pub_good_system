package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const emailLogAggregateGlobalKey = "global"

type emailLogAggregateGlobal struct {
	TotalCount   int64 `db:"total_count"`
	SuccessCount int64 `db:"success_count"`
	FailCount    int64 `db:"fail_count"`
}

type emailLogAggregateDailyRow struct {
	DayKey     int   `db:"day_key"`
	TotalCount int64 `db:"total_count"`
}

type emailLogAggregateTemplateRow struct {
	TemplateName string `db:"template_name"`
	TotalCount   int64  `db:"total_count"`
	SuccessCount int64  `db:"success_count"`
	FailCount    int64  `db:"fail_count"`
}

// EmailTemplateStat 邮件模板统计项
type EmailTemplateStat struct {
	TemplateName string `db:"template_name" json:"template_name"`
	Count        int64  `db:"count" json:"count"`
}

// EmailLogStatsDetail 邮件日志详细统计（读聚合表，清理明细不影响累计）
type EmailLogStatsDetail struct {
	TotalCount   int64               `json:"total_count"`
	TodayCount   int64               `json:"today_count"`
	SuccessCount int64               `json:"success_count"`
	FailCount    int64               `json:"fail_count"`
	TopTemplates []EmailTemplateStat `json:"top_templates"`
}

// InitEmailLogAggregateTables 初始化邮件日志聚合表
func InitEmailLogAggregateTables() {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS email_log_stats (
      stat_key VARCHAR(32) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      success_count BIGINT NOT NULL DEFAULT 0,
      fail_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件日志累计汇总';`,
		`CREATE TABLE IF NOT EXISTS email_log_daily_stats (
      day_key INT UNSIGNED NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件日志按天汇总';`,
		`CREATE TABLE IF NOT EXISTS email_log_template_stats (
      template_name VARCHAR(64) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      success_count BIGINT NOT NULL DEFAULT 0,
      fail_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件日志按模板汇总';`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			log.Printf("[Init] Failed to create email log aggregate table: %v", err)
		}
	}

	backfillEmailLogAggregateIfNeeded()
}

func backfillEmailLogAggregateIfNeeded() {
	if !db.CheckTableExists("email_logs") || !db.CheckTableExists("email_log_stats") {
		return
	}

	var existing int
	if err := db.DB.Get(&existing, "SELECT COUNT(*) FROM email_log_stats WHERE stat_key = ?", emailLogAggregateGlobalKey); err != nil {
		log.Printf("[Init] Failed to check email log aggregate data: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Printf("[Init] Failed to begin email log aggregate backfill: %v", err)
		return
	}

	if err := rebuildEmailLogAggregate(tx); err != nil {
		_ = tx.Rollback()
		log.Printf("[Init] Failed to backfill email log aggregate data: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[Init] Failed to commit email log aggregate backfill: %v", err)
	}
}

func rebuildEmailLogAggregate(tx *sqlx.Tx) error {
	tables := []string{
		"email_log_stats",
		"email_log_daily_stats",
		"email_log_template_stats",
	}
	for _, tableName := range tables {
		if _, err := tx.Exec("DELETE FROM " + tableName); err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	var global emailLogAggregateGlobal
	if err := tx.Get(&global, `SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count
    FROM email_logs`); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO email_log_stats (stat_key, total_count, success_count, fail_count, updated_at)
    VALUES (?, ?, ?, ?, ?)`,
		emailLogAggregateGlobalKey,
		global.TotalCount,
		global.SuccessCount,
		global.FailCount,
		now,
	); err != nil {
		return err
	}

	var dailyRows []emailLogAggregateDailyRow
	if err := tx.Select(&dailyRows, db.Q(`SELECT CAST(DATE_FORMAT(FROM_UNIXTIME(UNIX_TIMESTAMP(created_at)), '%Y%m%d') AS UNSIGNED) AS day_key, COUNT(*) AS total_count FROM email_logs GROUP BY day_key ORDER BY day_key ASC`)); err != nil {
		return err
	}
	for _, row := range dailyRows {
		if row.DayKey <= 0 {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO email_log_daily_stats (day_key, total_count, updated_at) VALUES (?, ?, ?)`, row.DayKey, row.TotalCount, now); err != nil {
			return err
		}
	}

	var templateRows []emailLogAggregateTemplateRow
	if err := tx.Select(&templateRows, `SELECT
    COALESCE(NULLIF(template_name, ''), 'unknown') AS template_name,
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count
    FROM email_logs
    GROUP BY COALESCE(NULLIF(template_name, ''), 'unknown')
    ORDER BY total_count DESC`); err != nil {
		return err
	}
	for _, row := range templateRows {
		if _, err := tx.Exec(`INSERT INTO email_log_template_stats (template_name, total_count, success_count, fail_count, updated_at) VALUES (?, ?, ?, ?, ?)`,
			row.TemplateName, row.TotalCount, row.SuccessCount, row.FailCount, now); err != nil {
			return err
		}
	}

	return nil
}

// RecordEmailLogAggregate 写入一封邮件日志后增量更新聚合统计
func RecordEmailLogAggregate(item *EmailLog) error {
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

	createTime := item.CreatedAt
	if createTime.IsZero() {
		createTime = time.Now()
	}
	updatedAt := time.Now().Unix()
	dayKey := resolveEmailLogDayKey(createTime)
	templateName := resolveEmailLogAggregateTemplate(item.TemplateName)

	successCount := 0
	failCount := 0
	if item.Status == 1 {
		successCount = 1
	} else {
		failCount = 1
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO email_log_stats (stat_key, total_count, success_count, fail_count, updated_at)
    VALUES (?, 1, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      success_count = success_count + ?,
      fail_count = fail_count + ?,
      updated_at = ?`),
		emailLogAggregateGlobalKey,
		successCount,
		failCount,
		updatedAt,
		successCount,
		failCount,
		updatedAt,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		db.Q(`INSERT INTO email_log_daily_stats (day_key, total_count, updated_at)
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
		db.Q(`INSERT INTO email_log_template_stats (template_name, total_count, success_count, fail_count, updated_at)
    VALUES (?, 1, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      success_count = success_count + ?,
      fail_count = fail_count + ?,
      updated_at = ?`),
		templateName,
		successCount,
		failCount,
		updatedAt,
		successCount,
		failCount,
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

// GetEmailLogStatsDetail 获取邮件日志详细统计（优先聚合表）
func GetEmailLogStatsDetail() (*EmailLogStatsDetail, error) {
	return getEmailLogStatsFromAggregate()
}

func getEmailLogStatsFromAggregate() (*EmailLogStatsDetail, error) {
	if !db.CheckTableExists("email_log_stats") || !db.CheckTableExists("email_log_daily_stats") || !db.CheckTableExists("email_log_template_stats") {
		return getEmailLogStatsFromLogsFallback()
	}

	stats := &EmailLogStatsDetail{
		TopTemplates: []EmailTemplateStat{},
	}

	var global emailLogAggregateGlobal
	if err := db.DB.Get(&global, `SELECT total_count, success_count, fail_count FROM email_log_stats WHERE stat_key = ?`, emailLogAggregateGlobalKey); err != nil {
		if err == sql.ErrNoRows {
			return getEmailLogStatsFromLogsFallback()
		}
		return nil, err
	}

	stats.TotalCount = global.TotalCount
	stats.SuccessCount = global.SuccessCount
	stats.FailCount = global.FailCount

	todayKey := resolveEmailLogDayKey(time.Now())
	if err := db.DB.Get(&stats.TodayCount, `SELECT total_count FROM email_log_daily_stats WHERE day_key = ?`, todayKey); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopTemplates, `SELECT template_name, total_count AS count FROM email_log_template_stats ORDER BY total_count DESC, template_name ASC LIMIT 10`); err != nil {
		return nil, err
	}

	return stats, nil
}

func getEmailLogStatsFromLogsFallback() (*EmailLogStatsDetail, error) {
	stats := &EmailLogStatsDetail{
		TopTemplates: []EmailTemplateStat{},
	}
	todayStart := resolveEmailLogStartOfLocalDay(time.Now())

	if err := db.DB.Get(&stats.TotalCount, "SELECT COUNT(*) FROM email_logs"); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.TodayCount, "SELECT COUNT(*) FROM email_logs WHERE created_at >= ?", todayStart); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.SuccessCount, "SELECT COUNT(*) FROM email_logs WHERE status = 1"); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.FailCount, "SELECT COUNT(*) FROM email_logs WHERE status != 1"); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopTemplates, `SELECT COALESCE(NULLIF(template_name, ''), 'unknown') AS template_name, COUNT(*) AS count FROM email_logs GROUP BY COALESCE(NULLIF(template_name, ''), 'unknown') ORDER BY count DESC LIMIT 10`); err != nil {
		return nil, err
	}
	return stats, nil
}

func resolveEmailLogDayKey(t time.Time) int {
	day := resolveEmailLogStartOfLocalDay(t)
	return day.Year()*10000 + int(day.Month())*100 + day.Day()
}

func resolveEmailLogStartOfLocalDay(target time.Time) time.Time {
	local := target.In(time.Local)
	year, month, day := local.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, local.Location())
}

func resolveEmailLogAggregateTemplate(name string) string {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "unknown"
	}
	return normalized
}
