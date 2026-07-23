package models

import (
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// aggregateDailyRow 按天汇总行（回填聚合表时在应用层计算 day_key，避免 DATE_FORMAT 等方言 SQL）
type aggregateDailyRow struct {
	DayKey     int
	TotalCount int64
}

// scanDailyCountsFromUnixColumn 从 unix 秒时间列统计每日条数（跨库：在 Go 侧算 day_key）
func scanDailyCountsFromUnixColumn(tx *gorm.DB, table, column string, dayKeyFn func(int64) int) ([]aggregateDailyRow, error) {
	type raw struct {
		TS int64 `gorm:"column:ts"`
	}
	var raws []raw
	if err := tx.Table(table).Select(column + " AS ts").Where(column + " > 0").Scan(&raws).Error; err != nil {
		return nil, err
	}
	counts := make(map[int]int64)
	for _, r := range raws {
		if dk := dayKeyFn(r.TS); dk > 0 {
			counts[dk]++
		}
	}
	return sortedAggregateDailyRows(counts), nil
}

// scanDailyCountsFromTimeColumn 从 TIMESTAMP/DATETIME 列统计每日条数
func scanDailyCountsFromTimeColumn(tx *gorm.DB, table, column string, dayKeyFn func(time.Time) int) ([]aggregateDailyRow, error) {
	type raw struct {
		TS time.Time `gorm:"column:ts"`
	}
	var raws []raw
	if err := tx.Table(table).Select(column + " AS ts").Scan(&raws).Error; err != nil {
		return nil, err
	}
	counts := make(map[int]int64)
	for _, r := range raws {
		if r.TS.IsZero() {
			continue
		}
		if dk := dayKeyFn(r.TS); dk > 0 {
			counts[dk]++
		}
	}
	return sortedAggregateDailyRows(counts), nil
}

func sortedAggregateDailyRows(counts map[int]int64) []aggregateDailyRow {
	rows := make([]aggregateDailyRow, 0, len(counts))
	for dk, cnt := range counts {
		rows = append(rows, aggregateDailyRow{DayKey: dk, TotalCount: cnt})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].DayKey < rows[j].DayKey })
	return rows
}

// upsertDailyTotal 按天主键累加 total_count（日志聚合表通用）。
func upsertDailyTotal(tx *gorm.DB, table string, dayKey int, updatedAt int64) error {
	row := map[string]interface{}{
		"day_key":     dayKey,
		"total_count": 1,
		"updated_at":  updatedAt,
	}
	return tx.Table(table).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "day_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_count": gorm.Expr("total_count + 1"),
			"updated_at":  updatedAt,
		}),
	}).Create(row).Error
}

// upsertKeyedTotal 按字符串主键累加 total_count。
func upsertKeyedTotal(tx *gorm.DB, table, keyColumn, keyValue string, updatedAt int64) error {
	row := map[string]interface{}{
		keyColumn:     keyValue,
		"total_count": 1,
		"updated_at":  updatedAt,
	}
	return tx.Table(table).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: keyColumn}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_count": gorm.Expr("total_count + 1"),
			"updated_at":  updatedAt,
		}),
	}).Create(row).Error
}
