package models

import (
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 日志聚合回填/兜底统计共用的列表达式（标准 SQL，三库通用）。
const (
	// aggHTTPStatusColumns API/操作日志按状态码分段计数，口径与 Record*Aggregate 增量更新一致
	aggHTTPStatusColumns = "COUNT(*) AS total_count" +
		", COALESCE(SUM(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 ELSE 0 END), 0) AS success_count" +
		", COALESCE(SUM(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 ELSE 0 END), 0) AS client_error_count" +
		", COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_error_count" +
		", COALESCE(SUM(duration), 0) AS total_duration"
	// aggSuccessFailColumns 邮件/短信日志成功、失败计数（status=1 为成功）
	aggSuccessFailColumns = "COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS success_count" +
		", COALESCE(SUM(CASE WHEN status = 1 THEN 0 ELSE 1 END), 0) AS fail_count"

	// 空值归为默认值的分组表达式，与 resolve*Aggregate* 的归一化保持一致
	aggRouteExpr    = "COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/')"
	aggMethodExpr   = "COALESCE(NULLIF(method, ''), 'UNKNOWN')"
	aggSceneExpr    = "COALESCE(NULLIF(scene, ''), 'unknown')"
	aggModuleExpr   = "COALESCE(NULLIF(module, ''), 'unknown')"
	aggActionExpr   = "COALESCE(NULLIF(action, ''), 'unknown')"
	aggTemplateExpr = "COALESCE(NULLIF(template_name, ''), 'unknown')"
	aggProviderExpr = "COALESCE(NULLIF(provider, ''), 'unknown')"
)

// aggCountBy 按表达式分组计数并按次数降序：SELECT <expr> AS <alias>, COUNT(*) AS <countAs><cols>。
func aggCountBy(q *gorm.DB, expr, alias, countAs, cols string) *gorm.DB {
	return q.Select(expr + " AS " + alias + ", COUNT(*) AS " + countAs + cols).Group(expr).Order(countAs + " DESC")
}

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
