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

const emailLogAggregateGlobalKey = "global"

// EmailLogStat 邮件日志全局汇总
type EmailLogStat struct {
	StatKey      string `gorm:"column:stat_key;primaryKey;size:32"`
	TotalCount   int64  `gorm:"column:total_count;not null;default:0"`
	SuccessCount int64  `gorm:"column:success_count;not null;default:0"`
	FailCount    int64  `gorm:"column:fail_count;not null;default:0"`
	UpdatedAt    int64  `gorm:"column:updated_at;not null;default:0"`
}

func (EmailLogStat) TableName() string { return "email_log_stats" }

// EmailLogDailyStat 邮件日志按天汇总
type EmailLogDailyStat struct {
	DayKey     int   `gorm:"column:day_key;primaryKey"`
	TotalCount int64 `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64 `gorm:"column:updated_at;not null;default:0"`
}

func (EmailLogDailyStat) TableName() string { return "email_log_daily_stats" }

// EmailLogTemplateStatRow 邮件日志按模板汇总
type EmailLogTemplateStatRow struct {
	TemplateName string `gorm:"column:template_name;primaryKey;size:100"`
	TotalCount   int64  `gorm:"column:total_count;not null;default:0"`
	SuccessCount int64  `gorm:"column:success_count;not null;default:0"`
	FailCount    int64  `gorm:"column:fail_count;not null;default:0"`
	UpdatedAt    int64  `gorm:"column:updated_at;not null;default:0"`
}

func (EmailLogTemplateStatRow) TableName() string { return "email_log_template_stats" }

type emailLogAggregateGlobal struct {
	TotalCount   int64 `gorm:"column:total_count"`
	SuccessCount int64 `gorm:"column:success_count"`
	FailCount    int64 `gorm:"column:fail_count"`
}

type emailLogAggregateDailyRow struct {
	DayKey     int   `gorm:"column:day_key"`
	TotalCount int64 `gorm:"column:total_count"`
}

type emailLogAggregateTemplateRow struct {
	TemplateName string `gorm:"column:template_name"`
	TotalCount   int64  `gorm:"column:total_count"`
	SuccessCount int64  `gorm:"column:success_count"`
	FailCount    int64  `gorm:"column:fail_count"`
}

// EmailTemplateStat 邮件模板统计项
type EmailTemplateStat struct {
	TemplateName string `gorm:"column:template_name" json:"template_name"`
	Count        int64  `gorm:"column:count" json:"count"`
}

// EmailLogStatsDetail 邮件日志详细统计（读聚合表，清理明细不影响累计）
type EmailLogStatsDetail struct {
	TotalCount   int64               `json:"total_count"`
	TodayCount   int64               `json:"today_count"`
	SuccessCount int64               `json:"success_count"`
	FailCount    int64               `json:"fail_count"`
	TopTemplates []EmailTemplateStat `json:"top_templates"`
}

// BackfillEmailLogAggregateIfNeeded 聚合表由 migrate.RunAutoMigrate 建表，仅回填历史数据
func BackfillEmailLogAggregateIfNeeded() {
	if !db.CheckTableExists("email_logs") || !db.CheckTableExists("email_log_stats") {
		return
	}

	var existing int64
	if err := db.DB.Model(&EmailLogStat{}).Where("stat_key = ?", emailLogAggregateGlobalKey).Count(&existing).Error; err != nil {
		log.Printf("[Init] Failed to check email log aggregate data: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	if err := db.WithTx(rebuildEmailLogAggregate); err != nil {
		log.Printf("[Init] Failed to backfill email log aggregate data: %v", err)
	}
}

func rebuildEmailLogAggregate(tx *gorm.DB) error {
	for _, m := range []interface{}{&EmailLogStat{}, &EmailLogDailyStat{}, &EmailLogTemplateStatRow{}} {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Where("1=1").Delete(m).Error; err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	var global emailLogAggregateGlobal
	if err := tx.Raw(`SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count
    FROM email_logs`).Scan(&global).Error; err != nil {
		return err
	}

	if err := tx.Create(&EmailLogStat{
		StatKey: emailLogAggregateGlobalKey, TotalCount: global.TotalCount,
		SuccessCount: global.SuccessCount, FailCount: global.FailCount, UpdatedAt: now,
	}).Error; err != nil {
		return err
	}

	var dailyRows []aggregateDailyRow
	if rows, err := scanDailyCountsFromTimeColumn(tx, "email_logs", "created_at", resolveEmailLogDayKey); err != nil {
		return err
	} else {
		dailyRows = rows
	}
	for _, row := range dailyRows {
		if row.DayKey <= 0 {
			continue
		}
		if err := tx.Create(&EmailLogDailyStat{DayKey: row.DayKey, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var templateRows []emailLogAggregateTemplateRow
	if err := tx.Raw(`SELECT
    COALESCE(NULLIF(template_name, ''), 'unknown') AS template_name,
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count
    FROM email_logs
    GROUP BY COALESCE(NULLIF(template_name, ''), 'unknown')
    ORDER BY total_count DESC`).Scan(&templateRows).Error; err != nil {
		return err
	}
	for _, row := range templateRows {
		if err := tx.Create(&EmailLogTemplateStatRow{
			TemplateName: row.TemplateName, TotalCount: row.TotalCount,
			SuccessCount: row.SuccessCount, FailCount: row.FailCount, UpdatedAt: now,
		}).Error; err != nil {
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

	return db.WithTx(func(tx *gorm.DB) error {
		createTime := item.CreatedAt
		if createTime.IsZero() {
			createTime = time.Now()
		}
		updatedAt := time.Now().Unix()
		dayKey := resolveEmailLogDayKey(createTime)
		templateName := resolveEmailLogAggregateTemplate(item.TemplateName)

		successCount := int64(0)
		failCount := int64(0)
		if item.Status == 1 {
			successCount = 1
		} else {
			failCount = 1
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "stat_key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_count":   gorm.Expr("total_count + 1"),
				"success_count": gorm.Expr("success_count + ?", successCount),
				"fail_count":    gorm.Expr("fail_count + ?", failCount),
				"updated_at":    updatedAt,
			}),
		}).Create(&EmailLogStat{
			StatKey: emailLogAggregateGlobalKey, TotalCount: 1,
			SuccessCount: successCount, FailCount: failCount, UpdatedAt: updatedAt,
		}).Error; err != nil {
			return err
		}

		if err := upsertDailyTotal(tx, "email_log_daily_stats", dayKey, updatedAt); err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "template_name"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_count":   gorm.Expr("total_count + 1"),
				"success_count": gorm.Expr("success_count + ?", successCount),
				"fail_count":    gorm.Expr("fail_count + ?", failCount),
				"updated_at":    updatedAt,
			}),
		}).Create(&EmailLogTemplateStatRow{
			TemplateName: templateName, TotalCount: 1,
			SuccessCount: successCount, FailCount: failCount, UpdatedAt: updatedAt,
		}).Error
	})
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
	err := db.DB.Model(&EmailLogStat{}).
		Select("total_count, success_count, fail_count").
		Where("stat_key = ?", emailLogAggregateGlobalKey).
		First(&global).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return getEmailLogStatsFromLogsFallback()
	}
	if err != nil {
		return nil, err
	}

	stats.TotalCount = global.TotalCount
	stats.SuccessCount = global.SuccessCount
	stats.FailCount = global.FailCount

	todayKey := resolveEmailLogDayKey(time.Now())
	var daily EmailLogDailyStat
	if err := db.DB.Select("total_count").Where("day_key = ?", todayKey).First(&daily).Error; err == nil {
		stats.TodayCount = daily.TotalCount
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := db.DB.Raw(`SELECT template_name, total_count AS count FROM email_log_template_stats ORDER BY total_count DESC, template_name ASC LIMIT 10`).Scan(&stats.TopTemplates).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func getEmailLogStatsFromLogsFallback() (*EmailLogStatsDetail, error) {
	stats := &EmailLogStatsDetail{
		TopTemplates: []EmailTemplateStat{},
	}
	todayStart := resolveEmailLogStartOfLocalDay(time.Now())

	if err := db.DB.Raw("SELECT COUNT(*) FROM email_logs").Scan(&stats.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM email_logs WHERE created_at >= ?", todayStart).Scan(&stats.TodayCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM email_logs WHERE status = 1").Scan(&stats.SuccessCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM email_logs WHERE status != 1").Scan(&stats.FailCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(template_name, ''), 'unknown') AS template_name, COUNT(*) AS count FROM email_logs GROUP BY COALESCE(NULLIF(template_name, ''), 'unknown') ORDER BY count DESC LIMIT 10`).Scan(&stats.TopTemplates).Error; err != nil {
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
	return clampBytes(normalized, storedModuleLen) // size:100
}
