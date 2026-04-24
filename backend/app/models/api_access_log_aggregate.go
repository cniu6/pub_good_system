package models

import (
  "database/sql"
  "fst/backend/pkg/db"
  "log"
  "strings"
  "time"

  "github.com/jmoiron/sqlx"
)

const apiAccessLogAggregateGlobalKey = "global"

type apiAccessLogAggregateGlobal struct {
  TotalCount       int64 `db:"total_count"`
  SuccessCount     int64 `db:"success_count"`
  ClientErrorCount int64 `db:"client_error_count"`
  ServerErrorCount int64 `db:"server_error_count"`
  TotalDuration    int64 `db:"total_duration"`
}

type apiAccessLogAggregateDailyRow struct {
  DayKey     int   `db:"day_key"`
  TotalCount int64 `db:"total_count"`
}

type apiAccessLogAggregatePathRow struct {
  RoutePath     string `db:"route_path"`
  TotalCount    int64  `db:"total_count"`
  TotalDuration int64  `db:"total_duration"`
}

type apiAccessLogAggregateMethodRow struct {
  Method     string `db:"method"`
  TotalCount int64  `db:"total_count"`
}

type apiAccessLogAggregateSceneRow struct {
  Scene      string `db:"scene"`
  TotalCount int64  `db:"total_count"`
}

type apiAccessLogAggregateIPRow struct {
  IP          string `db:"ip"`
  FirstSeenAt int64  `db:"first_seen_at"`
  LastSeenAt  int64  `db:"last_seen_at"`
}

// InitAPIAccessLogAggregateTables 初始化API访问日志聚合表
func InitAPIAccessLogAggregateTables() {
  schemas := []string{
    `CREATE TABLE IF NOT EXISTS api_access_log_stats (
      stat_key VARCHAR(32) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      success_count BIGINT NOT NULL DEFAULT 0,
      client_error_count BIGINT NOT NULL DEFAULT 0,
      server_error_count BIGINT NOT NULL DEFAULT 0,
      total_duration BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口日志累计汇总';`,
    `CREATE TABLE IF NOT EXISTS api_access_log_daily_stats (
      day_key INT UNSIGNED NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口日志按天汇总';`,
    `CREATE TABLE IF NOT EXISTS api_access_log_path_stats (
      route_path VARCHAR(255) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      total_duration BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口日志路由汇总';`,
    `CREATE TABLE IF NOT EXISTS api_access_log_method_stats (
      method VARCHAR(20) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口日志方法汇总';`,
    `CREATE TABLE IF NOT EXISTS api_access_log_scene_stats (
      scene VARCHAR(32) NOT NULL PRIMARY KEY,
      total_count BIGINT NOT NULL DEFAULT 0,
      updated_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口日志场景汇总';`,
    `CREATE TABLE IF NOT EXISTS api_access_log_ip_stats (
      ip VARCHAR(45) NOT NULL PRIMARY KEY,
      first_seen_at BIGINT NOT NULL DEFAULT 0,
      last_seen_at BIGINT NOT NULL DEFAULT 0
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口日志独立IP汇总';`,
  }

  for _, schema := range schemas {
    if _, err := db.DB.Exec(schema); err != nil {
      log.Printf("[Init] Failed to create API access log aggregate table: %v", err)
    }
  }

  backfillAPIAccessLogAggregateIfNeeded()
}

func backfillAPIAccessLogAggregateIfNeeded() {
  if !db.CheckTableExists("api_access_logs") || !db.CheckTableExists("api_access_log_stats") {
    return
  }

  var existing int
  if err := db.DB.Get(&existing, "SELECT COUNT(*) FROM api_access_log_stats WHERE stat_key = ?", apiAccessLogAggregateGlobalKey); err != nil {
    log.Printf("[Init] Failed to check API access log aggregate data: %v", err)
    return
  }
  if existing > 0 {
    return
  }

  tx, err := db.DB.Beginx()
  if err != nil {
    log.Printf("[Init] Failed to begin API access log aggregate backfill: %v", err)
    return
  }

  if err := rebuildAPIAccessLogAggregate(tx); err != nil {
    _ = tx.Rollback()
    log.Printf("[Init] Failed to backfill API access log aggregate data: %v", err)
    return
  }

  if err := tx.Commit(); err != nil {
    log.Printf("[Init] Failed to commit API access log aggregate backfill: %v", err)
  }
}

func rebuildAPIAccessLogAggregate(tx *sqlx.Tx) error {
  tables := []string{
    "api_access_log_stats",
    "api_access_log_daily_stats",
    "api_access_log_path_stats",
    "api_access_log_method_stats",
    "api_access_log_scene_stats",
    "api_access_log_ip_stats",
  }
  for _, tableName := range tables {
    if _, err := tx.Exec("DELETE FROM " + tableName); err != nil {
      return err
    }
  }

  now := time.Now().Unix()
  var global apiAccessLogAggregateGlobal
  if err := tx.Get(&global, `SELECT
    COUNT(*) AS total_count,
    COALESCE(SUM(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 ELSE 0 END), 0) AS success_count,
    COALESCE(SUM(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 ELSE 0 END), 0) AS client_error_count,
    COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_error_count,
    COALESCE(SUM(duration), 0) AS total_duration
    FROM api_access_logs`); err != nil {
    return err
  }

  if _, err := tx.Exec(
    `INSERT INTO api_access_log_stats (stat_key, total_count, success_count, client_error_count, server_error_count, total_duration, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)`,
    apiAccessLogAggregateGlobalKey,
    global.TotalCount,
    global.SuccessCount,
    global.ClientErrorCount,
    global.ServerErrorCount,
    global.TotalDuration,
    now,
  ); err != nil {
    return err
  }

  var dailyRows []apiAccessLogAggregateDailyRow
  if err := tx.Select(&dailyRows, `SELECT CAST(DATE_FORMAT(FROM_UNIXTIME(create_time), '%Y%m%d') AS UNSIGNED) AS day_key, COUNT(*) AS total_count FROM api_access_logs GROUP BY day_key ORDER BY day_key ASC`); err != nil {
    return err
  }
  for _, row := range dailyRows {
    if row.DayKey <= 0 {
      continue
    }
    if _, err := tx.Exec(`INSERT INTO api_access_log_daily_stats (day_key, total_count, updated_at) VALUES (?, ?, ?)`, row.DayKey, row.TotalCount, now); err != nil {
      return err
    }
  }

  var pathRows []apiAccessLogAggregatePathRow
  if err := tx.Select(&pathRows, `SELECT COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') AS route_path, COUNT(*) AS total_count, COALESCE(SUM(duration), 0) AS total_duration FROM api_access_logs GROUP BY COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') ORDER BY total_count DESC`); err != nil {
    return err
  }
  for _, row := range pathRows {
    if _, err := tx.Exec(`INSERT INTO api_access_log_path_stats (route_path, total_count, total_duration, updated_at) VALUES (?, ?, ?, ?)`, row.RoutePath, row.TotalCount, row.TotalDuration, now); err != nil {
      return err
    }
  }

  var methodRows []apiAccessLogAggregateMethodRow
  if err := tx.Select(&methodRows, `SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS total_count FROM api_access_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY total_count DESC`); err != nil {
    return err
  }
  for _, row := range methodRows {
    if _, err := tx.Exec(`INSERT INTO api_access_log_method_stats (method, total_count, updated_at) VALUES (?, ?, ?)`, row.Method, row.TotalCount, now); err != nil {
      return err
    }
  }

  var sceneRows []apiAccessLogAggregateSceneRow
  if err := tx.Select(&sceneRows, `SELECT COALESCE(NULLIF(scene, ''), 'unknown') AS scene, COUNT(*) AS total_count FROM api_access_logs GROUP BY COALESCE(NULLIF(scene, ''), 'unknown') ORDER BY total_count DESC`); err != nil {
    return err
  }
  for _, row := range sceneRows {
    if _, err := tx.Exec(`INSERT INTO api_access_log_scene_stats (scene, total_count, updated_at) VALUES (?, ?, ?)`, row.Scene, row.TotalCount, now); err != nil {
      return err
    }
  }

  var ipRows []apiAccessLogAggregateIPRow
  if err := tx.Select(&ipRows, `SELECT ip, MIN(create_time) AS first_seen_at, MAX(create_time) AS last_seen_at FROM api_access_logs WHERE ip != '' GROUP BY ip ORDER BY MAX(create_time) DESC`); err != nil {
    return err
  }
  for _, row := range ipRows {
    if _, err := tx.Exec(`INSERT INTO api_access_log_ip_stats (ip, first_seen_at, last_seen_at) VALUES (?, ?, ?)`, row.IP, row.FirstSeenAt, row.LastSeenAt); err != nil {
      return err
    }
  }

  return nil
}

func RecordAPIAccessLogAggregate(item *APIAccessLog) error {
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
  dayKey := resolveAPIAccessLogAggregateDayKey(createTime)
  routePath := resolveAPIAccessLogAggregateRoute(item.RoutePath, item.Path)
  method := resolveAPIAccessLogAggregateMethod(item.Method)
  scene := resolveAPIAccessLogAggregateScene(item.Scene)
  ip := strings.TrimSpace(item.IP)

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
    `INSERT INTO api_access_log_stats (stat_key, total_count, success_count, client_error_count, server_error_count, total_duration, updated_at)
    VALUES (?, 1, ?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      success_count = success_count + ?,
      client_error_count = client_error_count + ?,
      server_error_count = server_error_count + ?,
      total_duration = total_duration + ?,
      updated_at = ?`,
    apiAccessLogAggregateGlobalKey,
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
    `INSERT INTO api_access_log_daily_stats (day_key, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`,
    dayKey,
    updatedAt,
    updatedAt,
  ); err != nil {
    return err
  }

  if _, err := tx.Exec(
    `INSERT INTO api_access_log_path_stats (route_path, total_count, total_duration, updated_at)
    VALUES (?, 1, ?, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      total_duration = total_duration + ?,
      updated_at = ?`,
    routePath,
    item.Duration,
    updatedAt,
    item.Duration,
    updatedAt,
  ); err != nil {
    return err
  }

  if _, err := tx.Exec(
    `INSERT INTO api_access_log_method_stats (method, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`,
    method,
    updatedAt,
    updatedAt,
  ); err != nil {
    return err
  }

  if _, err := tx.Exec(
    `INSERT INTO api_access_log_scene_stats (scene, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`,
    scene,
    updatedAt,
    updatedAt,
  ); err != nil {
    return err
  }

  if ip != "" {
    if _, err := tx.Exec(
      `INSERT INTO api_access_log_ip_stats (ip, first_seen_at, last_seen_at)
      VALUES (?, ?, ?)
      ON DUPLICATE KEY UPDATE
        last_seen_at = CASE WHEN last_seen_at > ? THEN last_seen_at ELSE ? END`,
      ip,
      createTime,
      createTime,
      createTime,
      createTime,
    ); err != nil {
      return err
    }
  }

  if err := tx.Commit(); err != nil {
    return err
  }
  committed = true
  return nil
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
  if err := db.DB.Get(&global, `SELECT total_count, success_count, client_error_count, server_error_count, total_duration FROM api_access_log_stats WHERE stat_key = ?`, apiAccessLogAggregateGlobalKey); err != nil {
    if err == sql.ErrNoRows {
      return getAPIAccessLogStatsFromLogsFallback()
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

  todayKey := resolveAPIAccessLogAggregateDayKey(time.Now().Unix())
  if err := db.DB.Get(&stats.TodayCount, `SELECT total_count FROM api_access_log_daily_stats WHERE day_key = ?`, todayKey); err != nil && err != sql.ErrNoRows {
    return nil, err
  }
  if err := db.DB.Get(&stats.DistinctIPCount, `SELECT COUNT(*) FROM api_access_log_ip_stats`); err != nil {
    return nil, err
  }
  if err := db.DB.Select(&stats.TopPaths, `SELECT route_path, total_count AS count, COALESCE(total_duration / NULLIF(total_count, 0), 0) AS avg_duration FROM api_access_log_path_stats ORDER BY total_count DESC, route_path ASC LIMIT 10`); err != nil {
    return nil, err
  }
  if err := db.DB.Select(&stats.MethodStats, `SELECT method, total_count AS count FROM api_access_log_method_stats ORDER BY total_count DESC, method ASC`); err != nil {
    return nil, err
  }
  if err := db.DB.Select(&stats.SceneStats, `SELECT scene, total_count AS count FROM api_access_log_scene_stats ORDER BY total_count DESC, scene ASC`); err != nil {
    return nil, err
  }

  return stats, nil
}

func getAPIAccessLogStatsFromLogsFallback() (*APIAccessLogStats, error) {
  stats := &APIAccessLogStats{}
  todayStart := resolveAPIAccessLogStartOfLocalDay(time.Now()).Unix()

  if err := db.DB.Get(&stats.TotalCount, "SELECT COUNT(*) FROM api_access_logs"); err != nil {
    return nil, err
  }
  if err := db.DB.Get(&stats.TodayCount, "SELECT COUNT(*) FROM api_access_logs WHERE create_time >= ?", todayStart); err != nil {
    return nil, err
  }
  if err := db.DB.Get(&stats.SuccessCount, "SELECT COUNT(*) FROM api_access_logs WHERE status_code >= 200 AND status_code < 400"); err != nil {
    return nil, err
  }
  if err := db.DB.Get(&stats.ClientErrorCount, "SELECT COUNT(*) FROM api_access_logs WHERE status_code >= 400 AND status_code < 500"); err != nil {
    return nil, err
  }
  if err := db.DB.Get(&stats.ServerErrorCount, "SELECT COUNT(*) FROM api_access_logs WHERE status_code >= 500"); err != nil {
    return nil, err
  }
  if err := db.DB.Get(&stats.DistinctIPCount, "SELECT COUNT(DISTINCT ip) FROM api_access_logs WHERE ip != ''"); err != nil {
    return nil, err
  }
  if err := db.DB.Get(&stats.AvgDuration, "SELECT COALESCE(AVG(duration), 0) FROM api_access_logs"); err != nil {
    return nil, err
  }
  if err := db.DB.Select(&stats.TopPaths, `SELECT COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') AS route_path, COUNT(*) AS count, COALESCE(AVG(duration), 0) AS avg_duration FROM api_access_logs GROUP BY COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') ORDER BY count DESC LIMIT 10`); err != nil {
    return nil, err
  }
  if err := db.DB.Select(&stats.MethodStats, "SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS count FROM api_access_logs GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY count DESC"); err != nil {
    return nil, err
  }
  if err := db.DB.Select(&stats.SceneStats, "SELECT COALESCE(NULLIF(scene, ''), 'unknown') AS scene, COUNT(*) AS count FROM api_access_logs GROUP BY COALESCE(NULLIF(scene, ''), 'unknown') ORDER BY count DESC"); err != nil {
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
