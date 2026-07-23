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

const smsLogAggregateGlobalKey = "global"

// SMSLogStat 短信日志全局汇总
type SMSLogStat struct {
	StatKey      string `gorm:"column:stat_key;primaryKey;size:32"`
	TotalCount   int64  `gorm:"column:total_count;not null;default:0"`
	SuccessCount int64  `gorm:"column:success_count;not null;default:0"`
	FailCount    int64  `gorm:"column:fail_count;not null;default:0"`
	UpdatedAt    int64  `gorm:"column:updated_at;not null;default:0"`
}

func (SMSLogStat) TableName() string { return "sms_log_stats" }

// SMSLogDailyStat 短信日志按天汇总
type SMSLogDailyStat struct {
	DayKey     int   `gorm:"column:day_key;primaryKey"`
	TotalCount int64 `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64 `gorm:"column:updated_at;not null;default:0"`
}

func (SMSLogDailyStat) TableName() string { return "sms_log_daily_stats" }

// SMSLogTemplateStatRow 短信日志按模板汇总
type SMSLogTemplateStatRow struct {
	TemplateName string `gorm:"column:template_name;primaryKey;size:100"`
	TotalCount   int64  `gorm:"column:total_count;not null;default:0"`
	SuccessCount int64  `gorm:"column:success_count;not null;default:0"`
	FailCount    int64  `gorm:"column:fail_count;not null;default:0"`
	UpdatedAt    int64  `gorm:"column:updated_at;not null;default:0"`
}

func (SMSLogTemplateStatRow) TableName() string { return "sms_log_template_stats" }

// SMSLogProviderStatRow 短信日志按服务商汇总
type SMSLogProviderStatRow struct {
	Provider   string `gorm:"column:provider;primaryKey;size:64"`
	TotalCount int64  `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64  `gorm:"column:updated_at;not null;default:0"`
}

func (SMSLogProviderStatRow) TableName() string { return "sms_log_provider_stats" }

type smsLogAggregateGlobal struct {
	TotalCount   int64 `gorm:"column:total_count"`
	SuccessCount int64 `gorm:"column:success_count"`
	FailCount    int64 `gorm:"column:fail_count"`
}

type smsLogAggregateDailyRow struct {
	DayKey     int   `gorm:"column:day_key"`
	TotalCount int64 `gorm:"column:total_count"`
}

type smsLogAggregateTemplateRow struct {
	TemplateName string `gorm:"column:template_name"`
	TotalCount   int64  `gorm:"column:total_count"`
	SuccessCount int64  `gorm:"column:success_count"`
	FailCount    int64  `gorm:"column:fail_count"`
}

type smsLogAggregateProviderRow struct {
	Provider   string `gorm:"column:provider"`
	TotalCount int64  `gorm:"column:total_count"`
}

// SMSTemplateStat 短信模板统计项
type SMSTemplateStat struct {
	TemplateName string `gorm:"column:template_name" json:"template_name"`
	Count        int64  `gorm:"column:count" json:"count"`
}

// SMSProviderStat 短信服务商统计项
type SMSProviderStat struct {
	Provider string `gorm:"column:provider" json:"provider"`
	Count    int64  `gorm:"column:count" json:"count"`
}

// SMSLogStatsDetail 短信日志详细统计（读聚合表，清理明细不影响累计）
type SMSLogStatsDetail struct {
	TotalCount    int64             `json:"total_count"`
	TodayCount    int64             `json:"today_count"`
	SuccessCount  int64             `json:"success_count"`
	FailCount     int64             `json:"fail_count"`
	TopTemplates  []SMSTemplateStat `json:"top_templates"`
	ProviderStats []SMSProviderStat `json:"provider_stats"`
}

// BackfillSMSLogAggregateIfNeeded 聚合表由 migrate.RunAutoMigrate 建表，仅回填历史数据
func BackfillSMSLogAggregateIfNeeded() {
	if !db.CheckTableExists("sms_logs") || !db.CheckTableExists("sms_log_stats") {
		return
	}

	var existing int64
	if err := db.DB.Model(&SMSLogStat{}).Where("stat_key = ?", smsLogAggregateGlobalKey).Count(&existing).Error; err != nil {
		log.Printf("[Init] Failed to check SMS log aggregate data: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	if err := db.WithTx(rebuildSMSLogAggregate); err != nil {
		log.Printf("[Init] Failed to backfill SMS log aggregate data: %v", err)
	}
}

func rebuildSMSLogAggregate(tx *gorm.DB) error {
	for _, m := range []interface{}{&SMSLogStat{}, &SMSLogDailyStat{}, &SMSLogTemplateStatRow{}, &SMSLogProviderStatRow{}} {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Where("1=1").Delete(m).Error; err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	var global smsLogAggregateGlobal
	if err := tx.Raw(`SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count
    FROM sms_logs`).Scan(&global).Error; err != nil {
		return err
	}

	if err := tx.Create(&SMSLogStat{
		StatKey: smsLogAggregateGlobalKey, TotalCount: global.TotalCount,
		SuccessCount: global.SuccessCount, FailCount: global.FailCount, UpdatedAt: now,
	}).Error; err != nil {
		return err
	}

	var dailyRows []aggregateDailyRow
	if rows, err := scanDailyCountsFromTimeColumn(tx, "sms_logs", "created_at", resolveSMSLogDayKey); err != nil {
		return err
	} else {
		dailyRows = rows
	}
	for _, row := range dailyRows {
		if row.DayKey <= 0 {
			continue
		}
		if err := tx.Create(&SMSLogDailyStat{DayKey: row.DayKey, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var templateRows []smsLogAggregateTemplateRow
	if err := tx.Raw(`SELECT
    COALESCE(NULLIF(template_name, ''), 'unknown') AS template_name,
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count
    FROM sms_logs
    GROUP BY COALESCE(NULLIF(template_name, ''), 'unknown')
    ORDER BY total_count DESC`).Scan(&templateRows).Error; err != nil {
		return err
	}
	for _, row := range templateRows {
		if err := tx.Create(&SMSLogTemplateStatRow{
			TemplateName: row.TemplateName, TotalCount: row.TotalCount,
			SuccessCount: row.SuccessCount, FailCount: row.FailCount, UpdatedAt: now,
		}).Error; err != nil {
			return err
		}
	}

	var providerRows []smsLogAggregateProviderRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(provider, ''), 'unknown') AS provider, COUNT(*) AS total_count FROM sms_logs GROUP BY COALESCE(NULLIF(provider, ''), 'unknown') ORDER BY total_count DESC`).Scan(&providerRows).Error; err != nil {
		return err
	}
	for _, row := range providerRows {
		if err := tx.Create(&SMSLogProviderStatRow{Provider: row.Provider, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	return nil
}

// RecordSMSLogAggregate 写入一条短信日志后增量更新聚合统计
func RecordSMSLogAggregate(item *SMSLog) error {
	if item == nil {
		return nil
	}

	return db.WithTx(func(tx *gorm.DB) error {
		createTime := item.CreatedAt
		if createTime.IsZero() {
			createTime = time.Now()
		}
		updatedAt := time.Now().Unix()
		dayKey := resolveSMSLogDayKey(createTime)
		templateName := resolveSMSLogAggregateTemplate(item.TemplateName)
		provider := resolveSMSLogAggregateProvider(item.Provider)

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
		}).Create(&SMSLogStat{
			StatKey: smsLogAggregateGlobalKey, TotalCount: 1,
			SuccessCount: successCount, FailCount: failCount, UpdatedAt: updatedAt,
		}).Error; err != nil {
			return err
		}

		if err := upsertDailyTotal(tx, "sms_log_daily_stats", dayKey, updatedAt); err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "template_name"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_count":   gorm.Expr("total_count + 1"),
				"success_count": gorm.Expr("success_count + ?", successCount),
				"fail_count":    gorm.Expr("fail_count + ?", failCount),
				"updated_at":    updatedAt,
			}),
		}).Create(&SMSLogTemplateStatRow{
			TemplateName: templateName, TotalCount: 1,
			SuccessCount: successCount, FailCount: failCount, UpdatedAt: updatedAt,
		}).Error; err != nil {
			return err
		}

		return upsertKeyedTotal(tx, "sms_log_provider_stats", "provider", provider, updatedAt)
	})
}

// GetSMSLogStatsDetail 获取短信日志详细统计（优先聚合表）
func GetSMSLogStatsDetail() (*SMSLogStatsDetail, error) {
	return getSMSLogStatsFromAggregate()
}

func getSMSLogStatsFromAggregate() (*SMSLogStatsDetail, error) {
	if !db.CheckTableExists("sms_log_stats") || !db.CheckTableExists("sms_log_daily_stats") || !db.CheckTableExists("sms_log_template_stats") || !db.CheckTableExists("sms_log_provider_stats") {
		return getSMSLogStatsFromLogsFallback()
	}

	stats := &SMSLogStatsDetail{
		TopTemplates:  []SMSTemplateStat{},
		ProviderStats: []SMSProviderStat{},
	}

	var global smsLogAggregateGlobal
	err := db.DB.Model(&SMSLogStat{}).
		Select("total_count, success_count, fail_count").
		Where("stat_key = ?", smsLogAggregateGlobalKey).
		First(&global).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return getSMSLogStatsFromLogsFallback()
	}
	if err != nil {
		return nil, err
	}

	stats.TotalCount = global.TotalCount
	stats.SuccessCount = global.SuccessCount
	stats.FailCount = global.FailCount

	todayKey := resolveSMSLogDayKey(time.Now())
	var daily SMSLogDailyStat
	if err := db.DB.Select("total_count").Where("day_key = ?", todayKey).First(&daily).Error; err == nil {
		stats.TodayCount = daily.TotalCount
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := db.DB.Raw(`SELECT template_name, total_count AS count FROM sms_log_template_stats ORDER BY total_count DESC, template_name ASC LIMIT 10`).Scan(&stats.TopTemplates).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT provider, total_count AS count FROM sms_log_provider_stats ORDER BY total_count DESC, provider ASC`).Scan(&stats.ProviderStats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func getSMSLogStatsFromLogsFallback() (*SMSLogStatsDetail, error) {
	stats := &SMSLogStatsDetail{
		TopTemplates:  []SMSTemplateStat{},
		ProviderStats: []SMSProviderStat{},
	}
	todayStart := resolveSMSLogStartOfLocalDay(time.Now())

	if err := db.DB.Raw("SELECT COUNT(*) FROM sms_logs").Scan(&stats.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM sms_logs WHERE created_at >= ?", todayStart).Scan(&stats.TodayCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM sms_logs WHERE status = 1").Scan(&stats.SuccessCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM sms_logs WHERE status != 1").Scan(&stats.FailCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(template_name, ''), 'unknown') AS template_name, COUNT(*) AS count FROM sms_logs GROUP BY COALESCE(NULLIF(template_name, ''), 'unknown') ORDER BY count DESC LIMIT 10`).Scan(&stats.TopTemplates).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(provider, ''), 'unknown') AS provider, COUNT(*) AS count FROM sms_logs GROUP BY COALESCE(NULLIF(provider, ''), 'unknown') ORDER BY count DESC`).Scan(&stats.ProviderStats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func resolveSMSLogDayKey(t time.Time) int {
	day := resolveSMSLogStartOfLocalDay(t)
	return day.Year()*10000 + int(day.Month())*100 + day.Day()
}

func resolveSMSLogStartOfLocalDay(target time.Time) time.Time {
	local := target.In(time.Local)
	year, month, day := local.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, local.Location())
}

func resolveSMSLogAggregateTemplate(name string) string {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "unknown"
	}
	return clampBytes(normalized, storedModuleLen) // size:100
}

func resolveSMSLogAggregateProvider(provider string) string {
	normalized := strings.TrimSpace(provider)
	if normalized == "" {
		return "unknown"
	}
	return clampBytes(normalized, storedIPMaxLen) // size:64
}
