package models

import (
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"
)

// APIAccessLog API接口访问日志
// 统一记录管理端 / 用户端 / 公共接口的请求与响应概要，便于审计、排障与统计。
type APIAccessLog struct {
	ID                  uint64  `db:"id" json:"id"`
	RequestID           string  `db:"request_id" json:"request_id"`
	UserID              uint64  `db:"user_id" json:"user_id"`
	Username            string  `db:"username" json:"username"`
	Role                string  `db:"role" json:"role"`
	AuthMethod          string  `db:"auth_method" json:"auth_method"`
	Scene               string  `db:"scene" json:"scene"`
	Method              string  `db:"method" json:"method"`
	Transport           string  `db:"transport" json:"transport"`
	Protocol            string  `db:"protocol" json:"protocol"`
	Path                string  `db:"path" json:"path"`
	RoutePath           string  `db:"route_path" json:"route_path"`
	HandlerName         string  `db:"handler_name" json:"handler_name"`
	RequestContentType  string  `db:"request_content_type" json:"request_content_type"`
	ResponseContentType string  `db:"response_content_type" json:"response_content_type"`
	QueryString         string  `db:"query_string" json:"query_string"`
	PathParams          *string `db:"path_params" json:"path_params,omitempty"`
	IP                  string  `db:"ip" json:"ip"`
	SourceIP            string  `db:"source_ip" json:"source_ip"`
	XIP                 string  `db:"x_ip" json:"x_ip"`
	XForwardedFor       string  `db:"x_forwarded_for" json:"x_forwarded_for"`
	XRealIP             string  `db:"x_real_ip" json:"x_real_ip"`
	UserAgent           string  `db:"user_agent" json:"user_agent"`
	Referer             string  `db:"referer" json:"referer"`
	RequestHeaders      *string `db:"request_headers" json:"request_headers,omitempty"`
	RequestBody         *string `db:"request_body" json:"request_body,omitempty"`
	ResponseBody        *string `db:"response_body" json:"response_body,omitempty"`
	StatusCode          int     `db:"status_code" json:"status_code"`
	Duration            int     `db:"duration" json:"duration"`
	RequestSize         int64   `db:"request_size" json:"request_size"`
	ResponseSize        int64   `db:"response_size" json:"response_size"`
	CreateTime          *int64  `db:"create_time" json:"create_time"`
}

func (l *APIAccessLog) TableName() string {
	return "api_access_logs"
}

func InitAPIAccessLogsTable() {
	if !db.CheckTableExists("api_access_logs") {
		schema := `CREATE TABLE IF NOT EXISTS api_access_logs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			request_id CHAR(36) NOT NULL COMMENT '请求唯一标识UUID',
			user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
			username VARCHAR(100) NOT NULL DEFAULT '' COMMENT '用户名',
			role VARCHAR(32) NOT NULL DEFAULT '' COMMENT '角色',
			auth_method VARCHAR(16) NOT NULL DEFAULT '' COMMENT '鉴权方式：jwt/apikey/none',
			scene VARCHAR(32) NOT NULL DEFAULT '' COMMENT '接口场景：admin/user/public/system/plugin',
			method VARCHAR(20) NOT NULL DEFAULT '' COMMENT '请求方法',
			transport VARCHAR(32) NOT NULL DEFAULT 'http' COMMENT '连接类型：http/websocket/sse/stream',
			protocol VARCHAR(32) NOT NULL DEFAULT '' COMMENT '请求协议版本',
			path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原始请求路径',
			route_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Gin 路由模板路径',
			handler_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处理函数',
			request_content_type VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求内容类型',
			response_content_type VARCHAR(255) NOT NULL DEFAULT '' COMMENT '响应内容类型',
			query_string TEXT COMMENT '查询参数',
			path_params TEXT COMMENT '路径参数(JSON)',
			ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '客户端IP',
			source_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '来源IP(RemoteAddr)',
			x_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'X-IP头',
			x_forwarded_for VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'X-Forwarded-For头',
			x_real_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'X-Real-IP头',
			user_agent TEXT COMMENT 'User-Agent',
			referer VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'Referer',
			request_headers MEDIUMTEXT COMMENT '请求头(JSON，凭证字段已脱敏)',
			request_body MEDIUMTEXT COMMENT '请求体(写入前截断)',
			response_body MEDIUMTEXT COMMENT '响应体(写入前截断)',
			status_code INT NOT NULL DEFAULT 0 COMMENT 'HTTP状态码',
			duration INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
			request_size BIGINT NOT NULL DEFAULT 0 COMMENT '请求体大小(bytes)',
			response_size BIGINT NOT NULL DEFAULT 0 COMMENT '响应体大小(bytes)',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			UNIQUE INDEX uk_request_id (request_id),
			INDEX idx_create_time_id (create_time, id),
			INDEX idx_scene_create_time (scene, create_time),
			INDEX idx_user_create_time (user_id, create_time),
			INDEX idx_method_create_time (method, create_time),
			INDEX idx_transport_create_time (transport, create_time),
			INDEX idx_status_create_time (status_code, create_time),
			INDEX idx_ip_create_time (ip, create_time),
			INDEX idx_handler_create_time (handler_name, create_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口访问日志表';`

		if _, err := db.Exec(schema); err != nil {
			log.Printf("[Init] Failed to create api_access_logs table: %v", err)
		} else {
			log.Println("[Init] Created api_access_logs table")
		}
		InitAPIAccessLogAggregateTables()
		return
	}

	columnRepairs := []struct {
		column   string
		alterSQL string
	}{
		{"request_id", "ALTER TABLE api_access_logs ADD COLUMN request_id CHAR(36) NULL COMMENT '请求唯一标识UUID' AFTER id"},
		{"user_id", "ALTER TABLE api_access_logs ADD COLUMN user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID' AFTER id"},
		{"username", "ALTER TABLE api_access_logs ADD COLUMN username VARCHAR(100) NOT NULL DEFAULT '' COMMENT '用户名' AFTER user_id"},
		{"role", "ALTER TABLE api_access_logs ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT '' COMMENT '角色' AFTER username"},
		{"auth_method", "ALTER TABLE api_access_logs ADD COLUMN auth_method VARCHAR(16) NOT NULL DEFAULT '' COMMENT '鉴权方式：jwt/apikey/none' AFTER role"},
		{"scene", "ALTER TABLE api_access_logs ADD COLUMN scene VARCHAR(32) NOT NULL DEFAULT '' COMMENT '接口场景：admin/user/public/system/plugin' AFTER role"},
		{"method", "ALTER TABLE api_access_logs ADD COLUMN method VARCHAR(20) NOT NULL DEFAULT '' COMMENT '请求方法' AFTER scene"},
		{"transport", "ALTER TABLE api_access_logs ADD COLUMN transport VARCHAR(32) NOT NULL DEFAULT 'http' COMMENT '连接类型：http/websocket/sse/stream' AFTER method"},
		{"protocol", "ALTER TABLE api_access_logs ADD COLUMN protocol VARCHAR(32) NOT NULL DEFAULT '' COMMENT '请求协议版本' AFTER transport"},
		{"path", "ALTER TABLE api_access_logs ADD COLUMN path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原始请求路径' AFTER method"},
		{"route_path", "ALTER TABLE api_access_logs ADD COLUMN route_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Gin 路由模板路径' AFTER path"},
		{"handler_name", "ALTER TABLE api_access_logs ADD COLUMN handler_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处理函数' AFTER route_path"},
		{"request_content_type", "ALTER TABLE api_access_logs ADD COLUMN request_content_type VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求内容类型' AFTER handler_name"},
		{"response_content_type", "ALTER TABLE api_access_logs ADD COLUMN response_content_type VARCHAR(255) NOT NULL DEFAULT '' COMMENT '响应内容类型' AFTER request_content_type"},
		{"query_string", "ALTER TABLE api_access_logs ADD COLUMN query_string TEXT COMMENT '查询参数' AFTER route_path"},
		{"path_params", "ALTER TABLE api_access_logs ADD COLUMN path_params TEXT COMMENT '路径参数(JSON)' AFTER query_string"},
		{"ip", "ALTER TABLE api_access_logs ADD COLUMN ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '客户端IP' AFTER query_string"},
		{"source_ip", "ALTER TABLE api_access_logs ADD COLUMN source_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '来源IP(RemoteAddr)' AFTER ip"},
		{"x_ip", "ALTER TABLE api_access_logs ADD COLUMN x_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'X-IP头' AFTER source_ip"},
		{"x_forwarded_for", "ALTER TABLE api_access_logs ADD COLUMN x_forwarded_for VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'X-Forwarded-For头' AFTER x_ip"},
		{"x_real_ip", "ALTER TABLE api_access_logs ADD COLUMN x_real_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'X-Real-IP头' AFTER x_forwarded_for"},
		{"user_agent", "ALTER TABLE api_access_logs ADD COLUMN user_agent TEXT COMMENT 'User-Agent' AFTER ip"},
		{"referer", "ALTER TABLE api_access_logs ADD COLUMN referer VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'Referer' AFTER user_agent"},
		{"request_headers", "ALTER TABLE api_access_logs ADD COLUMN request_headers MEDIUMTEXT COMMENT '请求头(JSON，凭证字段已脱敏)' AFTER referer"},
		{"request_body", "ALTER TABLE api_access_logs ADD COLUMN request_body MEDIUMTEXT COMMENT '请求体(写入前截断)' AFTER referer"},
		{"response_body", "ALTER TABLE api_access_logs ADD COLUMN response_body MEDIUMTEXT COMMENT '响应体(写入前截断)' AFTER request_body"},
		{"status_code", "ALTER TABLE api_access_logs ADD COLUMN status_code INT NOT NULL DEFAULT 0 COMMENT 'HTTP状态码' AFTER response_body"},
		{"duration", "ALTER TABLE api_access_logs ADD COLUMN duration INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)' AFTER status_code"},
		{"request_size", "ALTER TABLE api_access_logs ADD COLUMN request_size BIGINT NOT NULL DEFAULT 0 COMMENT '请求体大小(bytes)' AFTER duration"},
		{"response_size", "ALTER TABLE api_access_logs ADD COLUMN response_size BIGINT NOT NULL DEFAULT 0 COMMENT '响应体大小(bytes)' AFTER request_size"},
		{"create_time", "ALTER TABLE api_access_logs ADD COLUMN create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间' AFTER response_size"},
	}

	for _, repair := range columnRepairs {
		if !db.CheckColumnExists("api_access_logs", repair.column) {
			if _, err := db.Exec(repair.alterSQL); err != nil {
				log.Printf("[Init] Failed to add api_access_logs.%s: %v", repair.column, err)
			} else {
				log.Printf("[Init] Added api_access_logs.%s", repair.column)
			}
		}
	}

	if db.CheckColumnExists("api_access_logs", "request_id") {
		if _, err := db.Exec("UPDATE api_access_logs SET request_id = LOWER(UUID()) WHERE request_id IS NULL OR request_id = ''"); err != nil {
			log.Printf("[Init] Failed to backfill api_access_logs.request_id: %v", err)
		}
	}

	indexRepairs := map[string]string{
		"uk_request_id":               "ALTER TABLE api_access_logs ADD UNIQUE INDEX uk_request_id (request_id)",
		"idx_create_time_id":          "ALTER TABLE api_access_logs ADD INDEX idx_create_time_id (create_time, id)",
		"idx_scene_create_time":       "ALTER TABLE api_access_logs ADD INDEX idx_scene_create_time (scene, create_time)",
		"idx_auth_method_create_time": "ALTER TABLE api_access_logs ADD INDEX idx_auth_method_create_time (auth_method, create_time)",
		"idx_user_create_time":        "ALTER TABLE api_access_logs ADD INDEX idx_user_create_time (user_id, create_time)",
		"idx_method_create_time":      "ALTER TABLE api_access_logs ADD INDEX idx_method_create_time (method, create_time)",
		"idx_transport_create_time":   "ALTER TABLE api_access_logs ADD INDEX idx_transport_create_time (transport, create_time)",
		"idx_status_create_time":      "ALTER TABLE api_access_logs ADD INDEX idx_status_create_time (status_code, create_time)",
		"idx_ip_create_time":          "ALTER TABLE api_access_logs ADD INDEX idx_ip_create_time (ip, create_time)",
		"idx_handler_create_time":     "ALTER TABLE api_access_logs ADD INDEX idx_handler_create_time (handler_name, create_time)",
	}
	for indexName, alterSQL := range indexRepairs {
		db.EnsureIndex("api_access_logs", indexName, alterSQL)
	}

	InitAPIAccessLogAggregateTables()
}

func CreateAPIAccessLog(item *APIAccessLog) error {
	query := `INSERT INTO api_access_logs (
		request_id, user_id, username, role, auth_method, scene, method, transport, protocol, path, route_path, handler_name, request_content_type, response_content_type, query_string,
		path_params, ip, source_ip, x_ip, x_forwarded_for, x_real_ip, user_agent, referer, request_headers,
		request_body, response_body, status_code, duration, request_size, response_size, create_time
	) VALUES (
		:request_id, :user_id, :username, :role, :auth_method, :scene, :method, :transport, :protocol, :path, :route_path, :handler_name, :request_content_type, :response_content_type, :query_string,
		:path_params, :ip, :source_ip, :x_ip, :x_forwarded_for, :x_real_ip, :user_agent, :referer, :request_headers,
		:request_body, :response_body, :status_code, :duration, :request_size, :response_size, :create_time
	)`
	now := time.Now().Unix()
	item.CreateTime = &now
	result, err := db.DB.NamedExec(query, item)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = uint64(id)
	return nil
}

func GetAPIAccessLogByID(id uint64) (*APIAccessLog, error) {
	var item APIAccessLog
	err := db.DB.Get(&item, "SELECT * FROM api_access_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func GetAPIAccessLogByRequestID(requestID string) (*APIAccessLog, error) {
	var item APIAccessLog
	err := db.DB.Get(&item, "SELECT * FROM api_access_logs WHERE request_id = ?", strings.TrimSpace(requestID))
	if err != nil {
		return nil, err
	}
	return &item, nil
}

type APIAccessLogQuery struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	Keyword    string `form:"keyword" json:"keyword"`
	RequestID  string `form:"request_id" json:"request_id"`
	Scene      string `form:"scene" json:"scene"`
	AuthMethod string `form:"auth_method" json:"auth_method"`
	Transport  string `form:"transport" json:"transport"`
	UserID     uint64 `form:"user_id" json:"user_id"`
	Username   string `form:"username" json:"username"`
	Method     string `form:"method" json:"method"`
	Path       string `form:"path" json:"path"`
	IP         string `form:"ip" json:"ip"`
	StatusCode int    `form:"status_code" json:"status_code"`
	StartTime  int64  `form:"start_time" json:"start_time"`
	EndTime    int64  `form:"end_time" json:"end_time"`
}

func GetAPIAccessLogList(query *APIAccessLogQuery) ([]APIAccessLog, int64, error) {
	if query == nil {
		query = &APIAccessLogQuery{}
	}

	var list []APIAccessLog
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if query.RequestID != "" {
		where += " AND request_id = ?"
		args = append(args, strings.TrimSpace(query.RequestID))
	}
	if query.Keyword != "" {
		where += " AND (request_id LIKE ? OR path LIKE ? OR route_path LIKE ? OR username LIKE ? OR ip LIKE ? OR source_ip LIKE ? OR x_ip LIKE ? OR x_forwarded_for LIKE ? OR x_real_ip LIKE ? OR handler_name LIKE ? OR query_string LIKE ? OR transport LIKE ? OR protocol LIKE ? OR request_content_type LIKE ? OR response_content_type LIKE ?)"
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword)
	}
	if query.Scene != "" {
		where += " AND scene = ?"
		args = append(args, query.Scene)
	}
	if query.AuthMethod != "" {
		where += " AND auth_method = ?"
		args = append(args, strings.TrimSpace(query.AuthMethod))
	}
	if query.Transport != "" {
		where += " AND transport = ?"
		args = append(args, strings.TrimSpace(query.Transport))
	}
	if query.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, query.UserID)
	}
	if query.Username != "" {
		where += " AND username LIKE ?"
		args = append(args, "%"+query.Username+"%")
	}
	if query.Method != "" {
		where += " AND method = ?"
		args = append(args, query.Method)
	}
	if query.Path != "" {
		where += " AND (path LIKE ? OR route_path LIKE ?)"
		keyword := "%" + query.Path + "%"
		args = append(args, keyword, keyword)
	}
	if query.IP != "" {
		where += " AND (ip = ? OR source_ip = ? OR x_ip = ? OR x_real_ip = ? OR x_forwarded_for LIKE ?)"
		args = append(args, query.IP, query.IP, query.IP, query.IP, "%"+query.IP+"%")
	}
	if query.StatusCode > 0 {
		where += " AND status_code = ?"
		args = append(args, query.StatusCode)
	}
	if query.StartTime > 0 {
		where += " AND create_time >= ?"
		args = append(args, query.StartTime)
	}
	if query.EndTime > 0 {
		where += " AND create_time <= ?"
		args = append(args, query.EndTime)
	}

	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM api_access_logs "+where, args...); err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	listQuery := "SELECT id, request_id, user_id, username, role, auth_method, scene, method, transport, protocol, path, route_path, handler_name, request_content_type, response_content_type, query_string, ip, source_ip, x_ip, x_forwarded_for, x_real_ip, user_agent, referer, status_code, duration, request_size, response_size, create_time FROM api_access_logs " + where + " ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, query.PageSize, offset)
	if err := db.DB.Select(&list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func DeleteAPIAccessLogsBefore(beforeTime int64) (int64, error) {
	result, err := db.Exec("DELETE FROM api_access_logs WHERE create_time < ?", beforeTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func CleanExcessAPIAccessLogs(maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM api_access_logs"); err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	var cutoff struct {
		ID         uint64 `db:"id"`
		CreateTime int64  `db:"create_time"`
	}
	if err := db.DB.Get(&cutoff, "SELECT id, create_time FROM api_access_logs ORDER BY create_time DESC, id DESC LIMIT 1 OFFSET ?", maxCount-1); err != nil {
		return 0, err
	}

	result, err := db.Exec(
		"DELETE FROM api_access_logs WHERE create_time < ? OR (create_time = ? AND id < ?)",
		cutoff.CreateTime,
		cutoff.CreateTime,
		cutoff.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanExcessAPIAccessLogsPerUser 按用户清理超出上限的 API 访问日志
func CleanExcessAPIAccessLogsPerUser(maxPerUser int) (int64, error) {
	if maxPerUser <= 0 {
		return 0, nil
	}
	var groups []struct {
		UserID uint64 `db:"user_id"`
		Cnt    int64  `db:"cnt"`
	}
	// 仅清理已登录用户的日志；user_id=0（未鉴权）不按用户限制
	if err := db.DB.Select(&groups,
		"SELECT user_id, COUNT(*) AS cnt FROM api_access_logs WHERE user_id > 0 GROUP BY user_id HAVING COUNT(*) > ?",
		maxPerUser,
	); err != nil {
		return 0, err
	}

	var totalAffected int64
	for _, g := range groups {
		var cutoff struct {
			ID         uint64 `db:"id"`
			CreateTime int64  `db:"create_time"`
		}
		if err := db.DB.Get(&cutoff,
			"SELECT id, create_time FROM api_access_logs WHERE user_id = ? ORDER BY create_time DESC, id DESC LIMIT 1 OFFSET ?",
			g.UserID, maxPerUser-1,
		); err != nil {
			continue
		}
		result, err := db.Exec(
			"DELETE FROM api_access_logs WHERE user_id = ? AND (create_time < ? OR (create_time = ? AND id < ?))",
			g.UserID, cutoff.CreateTime, cutoff.CreateTime, cutoff.ID,
		)
		if err != nil {
			continue
		}
		if n, _ := result.RowsAffected(); n > 0 {
			totalAffected += n
		}
	}
	return totalAffected, nil
}

type APIAccessLogStats struct {
	TotalCount       int64                 `json:"total_count"`
	TodayCount       int64                 `json:"today_count"`
	SuccessCount     int64                 `json:"success_count"`
	ClientErrorCount int64                 `json:"client_error_count"`
	ServerErrorCount int64                 `json:"server_error_count"`
	DistinctIPCount  int64                 `json:"distinct_ip_count"`
	AvgDuration      float64               `json:"avg_duration"`
	TopPaths         []APIAccessPathStat   `json:"top_paths"`
	MethodStats      []APIAccessMethodStat `json:"method_stats"`
	SceneStats       []APIAccessSceneStat  `json:"scene_stats"`
}

type APIAccessPathStat struct {
	RoutePath   string  `db:"route_path" json:"route_path"`
	Count       int64   `db:"count" json:"count"`
	AvgDuration float64 `db:"avg_duration" json:"avg_duration"`
}

type APIAccessMethodStat struct {
	Method string `db:"method" json:"method"`
	Count  int64  `db:"count" json:"count"`
}

type APIAccessSceneStat struct {
	Scene string `db:"scene" json:"scene"`
	Count int64  `db:"count" json:"count"`
}

func GetAPIAccessLogStats() (*APIAccessLogStats, error) {
	return getAPIAccessLogStatsFromAggregate()
}

// GetAPIAccessLogStatsByUserID 当前用户自己的 API 访问日志统计（仅本人数据）
// 全局聚合表（api_access_log_stats 等）不区分 user_id，无法直接复用；
// 这里直接查询 api_access_logs 原始表，SQL 结构与 admin 端聚合失败时的
// 兜底查询 getAPIAccessLogStatsFromLogsFallback 保持一致，仅额外附加 user_id 过滤条件。
func GetAPIAccessLogStatsByUserID(userID uint64) (*APIAccessLogStats, error) {
	stats := &APIAccessLogStats{
		TopPaths:    []APIAccessPathStat{},
		MethodStats: []APIAccessMethodStat{},
	}
	todayStart := resolveAPIAccessLogStartOfLocalDay(time.Now()).Unix()

	if err := db.DB.Get(&stats.TotalCount, "SELECT COUNT(*) FROM api_access_logs WHERE user_id = ?", userID); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.TodayCount, "SELECT COUNT(*) FROM api_access_logs WHERE user_id = ? AND create_time >= ?", userID, todayStart); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.SuccessCount, "SELECT COUNT(*) FROM api_access_logs WHERE user_id = ? AND status_code >= 200 AND status_code < 400", userID); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.ClientErrorCount, "SELECT COUNT(*) FROM api_access_logs WHERE user_id = ? AND status_code >= 400 AND status_code < 500", userID); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.ServerErrorCount, "SELECT COUNT(*) FROM api_access_logs WHERE user_id = ? AND status_code >= 500", userID); err != nil {
		return nil, err
	}
	if err := db.DB.Get(&stats.AvgDuration, "SELECT COALESCE(AVG(duration), 0) FROM api_access_logs WHERE user_id = ?", userID); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.TopPaths, `SELECT COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') AS route_path, COUNT(*) AS count, COALESCE(AVG(duration), 0) AS avg_duration FROM api_access_logs WHERE user_id = ? GROUP BY COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') ORDER BY count DESC LIMIT 10`, userID); err != nil {
		return nil, err
	}
	if err := db.DB.Select(&stats.MethodStats, "SELECT COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS count FROM api_access_logs WHERE user_id = ? GROUP BY COALESCE(NULLIF(method, ''), 'UNKNOWN') ORDER BY count DESC", userID); err != nil {
		return nil, err
	}
	// 单用户维度不统计独立 IP 数与场景分布（场景对本人固定为 user，无区分意义）
	return stats, nil
}
