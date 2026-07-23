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

const apiAccessLogAggregateGlobalKey = "global"

// APIAccessLogStat API 访问日志全局汇总
type APIAccessLogStat struct {
	StatKey          string `gorm:"column:stat_key;primaryKey;size:32"`
	TotalCount       int64  `gorm:"column:total_count;not null;default:0"`
	SuccessCount     int64  `gorm:"column:success_count;not null;default:0"`
	ClientErrorCount int64  `gorm:"column:client_error_count;not null;default:0"`
	ServerErrorCount int64  `gorm:"column:server_error_count;not null;default:0"`
	TotalDuration    int64  `gorm:"column:total_duration;not null;default:0"`
	UpdatedAt        int64  `gorm:"column:updated_at;not null;default:0"`
}

func (APIAccessLogStat) TableName() string { return "api_access_log_stats" }

// APIAccessLogDailyStat API 访问日志按天汇总
type APIAccessLogDailyStat struct {
	DayKey     int   `gorm:"column:day_key;primaryKey"`
	TotalCount int64 `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64 `gorm:"column:updated_at;not null;default:0"`
}

func (APIAccessLogDailyStat) TableName() string { return "api_access_log_daily_stats" }

// APIAccessLogPathStatRow API 访问日志按路由汇总
type APIAccessLogPathStatRow struct {
	RoutePath     string `gorm:"column:route_path;primaryKey;size:255"`
	TotalCount    int64  `gorm:"column:total_count;not null;default:0"`
	TotalDuration int64  `gorm:"column:total_duration;not null;default:0"`
	UpdatedAt     int64  `gorm:"column:updated_at;not null;default:0"`
}

func (APIAccessLogPathStatRow) TableName() string { return "api_access_log_path_stats" }

// APIAccessLogMethodStatRow API 访问日志按方法汇总
type APIAccessLogMethodStatRow struct {
	Method     string `gorm:"column:method;primaryKey;size:20"`
	TotalCount int64  `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64  `gorm:"column:updated_at;not null;default:0"`
}

func (APIAccessLogMethodStatRow) TableName() string { return "api_access_log_method_stats" }

// APIAccessLogSceneStatRow API 访问日志按场景汇总
type APIAccessLogSceneStatRow struct {
	Scene      string `gorm:"column:scene;primaryKey;size:32"`
	TotalCount int64  `gorm:"column:total_count;not null;default:0"`
	UpdatedAt  int64  `gorm:"column:updated_at;not null;default:0"`
}

func (APIAccessLogSceneStatRow) TableName() string { return "api_access_log_scene_stats" }

// APIAccessLogIPStatRow API 访问日志独立 IP 汇总
type APIAccessLogIPStatRow struct {
	IP          string `gorm:"column:ip;primaryKey;size:45"`
	FirstSeenAt int64  `gorm:"column:first_seen_at;not null;default:0"`
	LastSeenAt  int64  `gorm:"column:last_seen_at;not null;default:0"`
}

func (APIAccessLogIPStatRow) TableName() string { return "api_access_log_ip_stats" }

type apiAccessLogAggregateGlobal struct {
	TotalCount       int64 `gorm:"column:total_count"`
	SuccessCount     int64 `gorm:"column:success_count"`
	ClientErrorCount int64 `gorm:"column:client_error_count"`
	ServerErrorCount int64 `gorm:"column:server_error_count"`
	TotalDuration    int64 `gorm:"column:total_duration"`
}

type apiAccessLogAggregateDailyRow struct {
	DayKey     int   `gorm:"column:day_key"`
	TotalCount int64 `gorm:"column:total_count"`
}

type apiAccessLogAggregatePathRow struct {
	RoutePath     string `gorm:"column:route_path"`
	TotalCount    int64  `gorm:"column:total_count"`
	TotalDuration int64  `gorm:"column:total_duration"`
}

type apiAccessLogAggregateMethodRow struct {
	Method     string `gorm:"column:method"`
	TotalCount int64  `gorm:"column:total_count"`
}

type apiAccessLogAggregateSceneRow struct {
	Scene      string `gorm:"column:scene"`
	TotalCount int64  `gorm:"column:total_count"`
}

type apiAccessLogAggregateIPRow struct {
	IP          string `gorm:"column:ip"`
	FirstSeenAt int64  `gorm:"column:first_seen_at"`
	LastSeenAt  int64  `gorm:"column:last_seen_at"`
}

// BackfillAPIAccessLogAggregateIfNeeded 聚合表由 migrate.RunAutoMigrate 建表，仅回填历史数据
func BackfillAPIAccessLogAggregateIfNeeded() {
	if !db.CheckTableExists("api_access_logs") || !db.CheckTableExists("api_access_log_stats") {
		return
	}

	var existing int64
	if err := db.DB.Model(&APIAccessLogStat{}).Where("stat_key = ?", apiAccessLogAggregateGlobalKey).Count(&existing).Error; err != nil {
		log.Printf("[Init] Failed to check API access log aggregate data: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	if err := db.WithTx(rebuildAPIAccessLogAggregate); err != nil {
		log.Printf("[Init] Failed to backfill API access log aggregate data: %v", err)
	}
}

func rebuildAPIAccessLogAggregate(tx *gorm.DB) error {
	for _, m := range []interface{}{
		&APIAccessLogStat{}, &APIAccessLogDailyStat{},
		&APIAccessLogPathStatRow{}, &APIAccessLogMethodStatRow{},
		&APIAccessLogSceneStatRow{}, &APIAccessLogIPStatRow{},
	} {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Where("1=1").Delete(m).Error; err != nil {
			return err
		}
	}

	now := time.Now().Unix()
	var global apiAccessLogAggregateGlobal
	if err := tx.Raw(`SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 ELSE 0 END), 0) AS client_error_count,
    COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_error_count,
    COALESCE(SUM(duration), 0) AS total_duration
    FROM api_access_logs`).Scan(&global).Error; err != nil {
		return err
	}

	if err := tx.Create(&APIAccessLogStat{
		StatKey: apiAccessLogAggregateGlobalKey,
		TotalCount: global.TotalCount, SuccessCount: global.SuccessCount,
		ClientErrorCount: global.ClientErrorCount, ServerErrorCount: global.ServerErrorCount,
		TotalDuration: global.TotalDuration, UpdatedAt: now,
	}).Error; err != nil {
		return err
	}

	var dailyRows []aggregateDailyRow
	if rows, err := scanDailyCountsFromUnixColumn(tx, "api_access_logs", "create_time", resolveAPIAccessLogAggregateDayKey); err != nil {
		return err
	} else {
		dailyRows = rows
	}
	for _, row := range dailyRows {
		if row.DayKey <= 0 {
			continue
		}
		if err := tx.Create(&APIAccessLogDailyStat{DayKey: row.DayKey, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var pathRows []apiAccessLogAggregatePathRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') AS route_path, COUNT(*) AS total_count, COALESCE(SUM(duration), 0) AS total_duration FROM api_access_logs GROUP BY COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') ORDER BY total_count DESC`).Scan(&pathRows).Error; err != nil {
		return err
	}
	for _, row := range pathRows {
		if err := tx.Create(&APIAccessLogPathStatRow{RoutePath: row.RoutePath, TotalCount: row.TotalCount, TotalDuration: row.TotalDuration, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var methodRows []apiAccessLogAggregateMethodRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS total_count FROM api_access_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY total_count DESC`).Scan(&methodRows).Error; err != nil {
		return err
	}
	for _, row := range methodRows {
		if err := tx.Create(&APIAccessLogMethodStatRow{Method: row.Method, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var sceneRows []apiAccessLogAggregateSceneRow
	if err := tx.Raw(`SELECT COALESCE(NULLIF(scene, ''), 'unknown') AS scene, COUNT(*) AS total_count FROM api_access_logs GROUP BY COALESCE(NULLIF(scene, ''), 'unknown') ORDER BY total_count DESC`).Scan(&sceneRows).Error; err != nil {
		return err
	}
	for _, row := range sceneRows {
		if err := tx.Create(&APIAccessLogSceneStatRow{Scene: row.Scene, TotalCount: row.TotalCount, UpdatedAt: now}).Error; err != nil {
			return err
		}
	}

	var ipRows []apiAccessLogAggregateIPRow
	if err := tx.Raw(`SELECT ip, MIN(create_time) AS first_seen_at, MAX(create_time) AS last_seen_at FROM api_access_logs WHERE ip != '' GROUP BY ip ORDER BY MAX(create_time) DESC`).Scan(&ipRows).Error; err != nil {
		return err
	}
	for _, row := range ipRows {
		if err := tx.Create(&APIAccessLogIPStatRow{IP: row.IP, FirstSeenAt: row.FirstSeenAt, LastSeenAt: row.LastSeenAt}).Error; err != nil {
			return err
		}
	}

	return nil
}

func RecordAPIAccessLogAggregate(item *APIAccessLog) error {
	if item == nil {
		return nil
	}

	return db.WithTx(func(tx *gorm.DB) error {
		createTime := time.Now().Unix()
		if item.CreateTime != nil && *item.CreateTime > 0 {
			createTime = *item.CreateTime
		}
		updatedAt := time.Now().Unix()
		dayKey := resolveAPIAccessLogAggregateDayKey(createTime)
		routePath := resolveAPIAccessLogAggregateRoute(item.RoutePath, item.Path)
		method := resolveAPIAccessLogAggregateMethod(item.Method)
		scene := resolveAPIAccessLogAggregateScene(item.Scene)
		ip := strings.TrimSpace(item.IP)

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
		}).Create(&APIAccessLogStat{
			StatKey: apiAccessLogAggregateGlobalKey, TotalCount: 1,
			SuccessCount: successCount, ClientErrorCount: clientErrorCount,
			ServerErrorCount: serverErrorCount, TotalDuration: int64(item.Duration), UpdatedAt: updatedAt,
		}).Error; err != nil {
			return err
		}

		if err := upsertDailyTotal(tx, "api_access_log_daily_stats", dayKey, updatedAt); err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "route_path"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_count":    gorm.Expr("total_count + 1"),
				"total_duration": gorm.Expr("total_duration + ?", item.Duration),
				"updated_at":     updatedAt,
			}),
		}).Create(&APIAccessLogPathStatRow{
			RoutePath: routePath, TotalCount: 1, TotalDuration: int64(item.Duration), UpdatedAt: updatedAt,
		}).Error; err != nil {
			return err
		}

		if err := upsertKeyedTotal(tx, "api_access_log_method_stats", "method", method, updatedAt); err != nil {
			return err
		}
		if err := upsertKeyedTotal(tx, "api_access_log_scene_stats", "scene", scene, updatedAt); err != nil {
			return err
		}

		if ip != "" {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "ip"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"last_seen_at": gorm.Expr("CASE WHEN last_seen_at > ? THEN last_seen_at ELSE ? END", createTime, createTime),
				}),
			}).Create(&APIAccessLogIPStatRow{IP: ip, FirstSeenAt: createTime, LastSeenAt: createTime}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func getAPIAccessLogStatsFromAggregate() (*APIAccessLogStats, error) {
	if !db.CheckTableExists("api_access_log_stats") || !db.CheckTableExists("api_access_log_daily_stats") || !db.CheckTableExists("api_access_log_path_stats") || !db.CheckTableExists("api_access_log_method_stats") || !db.CheckTableExists("api_access_log_scene_stats") || !db.CheckTableExists("api_access_log_ip_stats") {
		return getAPIAccessLogStatsFromLogsFallback()
	}

	stats := &APIAccessLogStats{
		TopPaths:    []APIAccessPathStat{},
		MethodStats: []APIAccessMethodStat{},
		SceneStats:  []APIAccessSceneStat{},
	}

	var global apiAccessLogAggregateGlobal
	err := db.DB.Model(&APIAccessLogStat{}).
		Select("total_count, success_count, client_error_count, server_error_count, total_duration").
		Where("stat_key = ?", apiAccessLogAggregateGlobalKey).
		First(&global).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return getAPIAccessLogStatsFromLogsFallback()
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

	todayKey := resolveAPIAccessLogAggregateDayKey(time.Now().Unix())
	var daily APIAccessLogDailyStat
	if err := db.DB.Select("total_count").Where("day_key = ?", todayKey).First(&daily).Error; err == nil {
		stats.TodayCount = daily.TotalCount
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := db.DB.Model(&APIAccessLogIPStatRow{}).Count(&stats.DistinctIPCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT route_path, total_count AS count, COALESCE(total_duration / NULLIF(total_count, 0), 0) AS avg_duration FROM api_access_log_path_stats ORDER BY total_count DESC, route_path ASC LIMIT 10`).Scan(&stats.TopPaths).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT method, total_count AS count FROM api_access_log_method_stats ORDER BY total_count DESC, method ASC`).Scan(&stats.MethodStats).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT scene, total_count AS count FROM api_access_log_scene_stats ORDER BY total_count DESC, scene ASC`).Scan(&stats.SceneStats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func getAPIAccessLogStatsFromLogsFallback() (*APIAccessLogStats, error) {
	stats := &APIAccessLogStats{}
	todayStart := resolveAPIAccessLogStartOfLocalDay(time.Now()).Unix()

	if err := db.DB.Raw("SELECT COUNT(*) FROM api_access_logs").Scan(&stats.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM api_access_logs WHERE create_time >= ?", todayStart).Scan(&stats.TodayCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM api_access_logs WHERE status_code >= 200 AND status_code < 400").Scan(&stats.SuccessCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM api_access_logs WHERE status_code >= 400 AND status_code < 500").Scan(&stats.ClientErrorCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM api_access_logs WHERE status_code >= 500").Scan(&stats.ServerErrorCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COUNT(DISTINCT ip) FROM api_access_logs WHERE ip != ''").Scan(&stats.DistinctIPCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COALESCE(AVG(duration), 0) FROM api_access_logs").Scan(&stats.AvgDuration).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw(`SELECT COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') AS route_path, COUNT(*) AS count, COALESCE(AVG(duration), 0) AS avg_duration FROM api_access_logs GROUP BY COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') ORDER BY count DESC LIMIT 10`).Scan(&stats.TopPaths).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS count FROM api_access_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY count DESC").Scan(&stats.MethodStats).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Raw("SELECT COALESCE(NULLIF(scene, ''), 'unknown') AS scene, COUNT(*) AS count FROM api_access_logs GROUP BY COALESCE(NULLIF(scene, ''), 'unknown') ORDER BY count DESC").Scan(&stats.SceneStats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func resolveAPIAccessLogAggregateDayKey(ts int64) int {
	day := resolveAPIAccessLogStartOfLocalDay(time.Unix(ts, 0).In(time.Local))
	return day.Year()*10000 + int(day.Month())*100 + day.Day()
}

func resolveAPIAccessLogStartOfLocalDay(target time.Time) time.Time {
	local := target.In(time.Local)
	year, month, day := local.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, local.Location())
}

func resolveAPIAccessLogAggregateRoute(routePath, path string) string {
	normalized := strings.TrimSpace(routePath)
	if normalized == "" {
		normalized = strings.TrimSpace(path)
	}
	if normalized == "" {
		return "/"
	}
	return normalized
}

func resolveAPIAccessLogAggregateMethod(method string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" {
		return "UNKNOWN"
	}
	return normalized
}

func resolveAPIAccessLogAggregateScene(scene string) string {
	normalized := strings.TrimSpace(scene)
	if normalized == "" {
		return "unknown"
	}
	return normalized
}
