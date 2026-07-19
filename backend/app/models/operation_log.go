package models

import (
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// OperationLog 操作日志模型
// request_body / response_body 列为 MEDIUMTEXT（约 16MB），写入前由中间件截断到 64KB。
type OperationLog struct {
	ID           uint64  `db:"id" json:"id"`
	UserID       uint64  `db:"user_id" json:"user_id"`
	Username     string  `db:"username" json:"username"`
	Module       string  `db:"module" json:"module"`
	Action       string  `db:"action" json:"action"`
	Method       string  `db:"method" json:"method"`
	Path         string  `db:"path" json:"path"`
	IP           string  `db:"ip" json:"ip"`
	UserAgent    string  `db:"user_agent" json:"user_agent"`
	HandlerName  string  `db:"handler_name" json:"handler_name"` // Gin handler / controller 方法名
	RequestBody  *string `db:"request_body" json:"request_body,omitempty"`
	ResponseBody *string `db:"response_body" json:"response_body,omitempty"`
	StatusCode   int     `db:"status_code" json:"status_code"`
	Duration     int     `db:"duration" json:"duration"` // 耗时(ms)
	CreateTime   *int64  `db:"create_time" json:"create_time"`
}

func (o *OperationLog) TableName() string {
	return "operation_logs"
}

var operationLogCleanupNextAt atomic.Int64

func InitOperationLogsTable() {
	if !db.CheckTableExists("operation_logs") {
		schema := `CREATE TABLE IF NOT EXISTS operation_logs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
			username VARCHAR(100) NOT NULL DEFAULT '' COMMENT '用户名',
			module VARCHAR(100) NOT NULL DEFAULT '' COMMENT '模块',
			action VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作',
			method VARCHAR(20) NOT NULL DEFAULT '' COMMENT '请求方法',
			path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
			ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
			user_agent TEXT COMMENT '浏览器UA',
			handler_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处理函数/Handler名',
			request_body MEDIUMTEXT COMMENT '请求体(写入前截断)',
			response_body MEDIUMTEXT COMMENT '响应体(写入前截断)',
			status_code INT NOT NULL DEFAULT 0 COMMENT '状态码',
			duration INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			INDEX idx_create_time_id (create_time, id),
			INDEX idx_user_create_time (user_id, create_time),
			INDEX idx_module_create_time (module, create_time),
			INDEX idx_action_create_time (action, create_time),
			INDEX idx_method_create_time (method, create_time),
			INDEX idx_ip_create_time (ip, create_time),
			INDEX idx_handler_create_time (handler_name, create_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

		_, err := db.Exec(schema)
		if err != nil {
			log.Printf("[Init] Failed to create operation_logs table: %v", err)
		} else {
			log.Println("[Init] Created operation_logs table")
		}
	} else {
		columnRepairs := []struct {
			column   string
			alterSQL string
		}{
			{"user_id", "ALTER TABLE operation_logs ADD COLUMN user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID' AFTER id"},
			{"username", "ALTER TABLE operation_logs ADD COLUMN username VARCHAR(100) NOT NULL DEFAULT '' COMMENT '用户名' AFTER user_id"},
			{"module", "ALTER TABLE operation_logs ADD COLUMN module VARCHAR(100) NOT NULL DEFAULT '' COMMENT '模块' AFTER username"},
			{"action", "ALTER TABLE operation_logs ADD COLUMN action VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作' AFTER module"},
			{"method", "ALTER TABLE operation_logs ADD COLUMN method VARCHAR(20) NOT NULL DEFAULT '' COMMENT '请求方法' AFTER action"},
			{"path", "ALTER TABLE operation_logs ADD COLUMN path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径' AFTER method"},
			{"ip", "ALTER TABLE operation_logs ADD COLUMN ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址' AFTER path"},
			{"user_agent", "ALTER TABLE operation_logs ADD COLUMN user_agent TEXT COMMENT '浏览器UA' AFTER ip"},
			{"handler_name", "ALTER TABLE operation_logs ADD COLUMN handler_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处理函数/Handler名' AFTER user_agent"},
			{"request_body", "ALTER TABLE operation_logs ADD COLUMN request_body MEDIUMTEXT COMMENT '请求体(写入前截断)' AFTER handler_name"},
			{"response_body", "ALTER TABLE operation_logs ADD COLUMN response_body MEDIUMTEXT COMMENT '响应体(写入前截断)' AFTER request_body"},
			{"status_code", "ALTER TABLE operation_logs ADD COLUMN status_code INT NOT NULL DEFAULT 0 COMMENT '状态码' AFTER response_body"},
			{"duration", "ALTER TABLE operation_logs ADD COLUMN duration INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)' AFTER status_code"},
			{"create_time", "ALTER TABLE operation_logs ADD COLUMN create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间' AFTER duration"},
		}

		for _, repair := range columnRepairs {
			if !db.CheckColumnExists("operation_logs", repair.column) {
				if _, err := db.Exec(repair.alterSQL); err != nil {
					log.Printf("[Init] Failed to add operation_logs.%s: %v", repair.column, err)
				} else {
					log.Printf("[Init] Added operation_logs.%s", repair.column)
				}
			}
		}

		if db.CheckColumnExists("operation_logs", "created_at") && db.CheckColumnExists("operation_logs", "create_time") {
			_, _ = db.Exec("UPDATE operation_logs SET create_time = UNIX_TIMESTAMP(created_at) WHERE create_time = 0 AND created_at IS NOT NULL")
		}

		indexRepairs := map[string]string{
			"idx_create_time_id":      "ALTER TABLE operation_logs ADD INDEX idx_create_time_id (create_time, id)",
			"idx_user_create_time":    "ALTER TABLE operation_logs ADD INDEX idx_user_create_time (user_id, create_time)",
			"idx_module_create_time":  "ALTER TABLE operation_logs ADD INDEX idx_module_create_time (module, create_time)",
			"idx_action_create_time":  "ALTER TABLE operation_logs ADD INDEX idx_action_create_time (action, create_time)",
			"idx_method_create_time":  "ALTER TABLE operation_logs ADD INDEX idx_method_create_time (method, create_time)",
			"idx_ip_create_time":      "ALTER TABLE operation_logs ADD INDEX idx_ip_create_time (ip, create_time)",
			"idx_handler_create_time": "ALTER TABLE operation_logs ADD INDEX idx_handler_create_time (handler_name, create_time)",
		}

		for indexName, alterSQL := range indexRepairs {
			db.EnsureIndex("operation_logs", indexName, alterSQL)
		}
	}

	InitOperationLogAggregateTables()
}

// ========== CRUD 操作 ==========

// CreateOperationLog 创建操作日志
func CreateOperationLog(item *OperationLog) error {
	query := `INSERT INTO operation_logs (user_id, username, module, action, method, path, ip,
			  user_agent, handler_name, request_body, response_body, status_code, duration, create_time)
			  VALUES (:user_id, :username, :module, :action, :method, :path, :ip,
			  :user_agent, :handler_name, :request_body, :response_body, :status_code, :duration, :create_time)`

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

	// 异步更新聚合统计 + 触发保留清理（节流）；明细清理不会回减累计统计
	go func(op *OperationLog) {
		if aggErr := RecordOperationLogAggregate(op); aggErr != nil {
			log.Printf("[OperationLog] 汇总更新失败: %v", aggErr)
		}
		scheduleOperationLogRetentionCleanup()
	}(item)

	return nil
}

func scheduleOperationLogRetentionCleanup() {
	now := time.Now().UnixNano()
	nextAt := operationLogCleanupNextAt.Load()
	if nextAt > now {
		return
	}
	if !operationLogCleanupNextAt.CompareAndSwap(nextAt, now+int64(30*time.Second)) {
		return
	}

	cfg := loadOperationLogRetentionConfig()
	if cfg.MaxCount > 0 {
		if _, err := CleanExcessOperationLogs(cfg.MaxCount); err != nil {
			log.Printf("[OperationLog] 自动清理超限日志失败: %v", err)
		}
	}
	if cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0 {
		if _, err := CleanExcessOperationLogsPerUser(cfg.PerUserMaxCount); err != nil {
			log.Printf("[OperationLog] 按用户清理超限日志失败: %v", err)
		}
	}
}

type operationLogRetentionConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

func loadOperationLogRetentionConfig() operationLogRetentionConfig {
	cfg := operationLogRetentionConfig{MaxCount: 1000, PerUserMaxCount: 1000}
	settingsMap, err := GetSettingsMap([]string{
		"operation_log_max_count",
		"operation_log_per_user_limit_enabled",
		"operation_log_per_user_max_count",
	})
	if err != nil {
		return cfg
	}
	if v, ok := settingsMap["operation_log_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.MaxCount = n
		}
	}
	if v, ok := settingsMap["operation_log_per_user_limit_enabled"]; ok {
		lower := strings.ToLower(strings.TrimSpace(v))
		cfg.PerUserLimitEnabled = lower == "true" || lower == "1"
	}
	if v, ok := settingsMap["operation_log_per_user_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.PerUserMaxCount = n
		}
	}
	return cfg
}

// GetOperationLogByID 根据ID获取日志
func GetOperationLogByID(id uint64) (*OperationLog, error) {
	var log OperationLog
	err := db.DB.Get(&log, "SELECT * FROM operation_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// OperationLogQuery 日志查询参数
type OperationLogQuery struct {
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
	UserID    uint64 `form:"user_id" json:"user_id"`
	Username  string `form:"username" json:"username"`
	Module    string `form:"module" json:"module"`
	Action    string `form:"action" json:"action"`
	Method    string `form:"method" json:"method"`
	Path      string `form:"path" json:"path"`
	IP        string `form:"ip" json:"ip"`
	StartTime int64  `form:"start_time" json:"start_time"`
	EndTime   int64  `form:"end_time" json:"end_time"`
}

// GetOperationLogList 获取日志列表
func GetOperationLogList(query *OperationLogQuery) ([]OperationLog, int64, error) {
	if query == nil {
		query = &OperationLogQuery{}
	}

	var logs []OperationLog
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if query.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, query.UserID)
	}
	if query.Username != "" {
		where += " AND username LIKE ?"
		args = append(args, "%"+query.Username+"%")
	}
	if query.Module != "" {
		where += " AND module = ?"
		args = append(args, query.Module)
	}
	if query.Action != "" {
		where += " AND action = ?"
		args = append(args, query.Action)
	}
	if query.Method != "" {
		where += " AND method = ?"
		args = append(args, query.Method)
	}
	if query.Path != "" {
		where += " AND path LIKE ?"
		args = append(args, "%"+query.Path+"%")
	}
	if query.IP != "" {
		where += " AND ip = ?"
		args = append(args, query.IP)
	}
	if query.StartTime > 0 {
		where += " AND create_time >= ?"
		args = append(args, query.StartTime)
	}
	if query.EndTime > 0 {
		where += " AND create_time <= ?"
		args = append(args, query.EndTime)
	}

	// 查询总数
	count_query := "SELECT COUNT(*) FROM operation_logs " + where
	err := db.DB.Get(&total, count_query, args...)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	list_query := "SELECT * FROM operation_logs " + where + " ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, query.PageSize, offset)

	err = db.DB.Select(&logs, list_query, args...)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteOperationLogsBefore 删除指定时间之前的日志
func DeleteOperationLogsBefore(before_time int64) (int64, error) {
	result, err := db.Exec("DELETE FROM operation_logs WHERE create_time < ?", before_time)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanExcessOperationLogs 清理超出上限的旧日志，只保留最新的 maxCount 条
func CleanExcessOperationLogs(maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	// 先查总数
	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM operation_logs"); err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	var cutoff struct {
		ID         uint64 `db:"id"`
		CreateTime int64  `db:"create_time"`
	}
	if err := db.DB.Get(&cutoff,
		"SELECT id, create_time FROM operation_logs ORDER BY create_time DESC, id DESC LIMIT 1 OFFSET ?",
		maxCount-1,
	); err != nil {
		return 0, err
	}

	result, err := db.Exec(
		"DELETE FROM operation_logs WHERE create_time < ? OR (create_time = ? AND id < ?)",
		cutoff.CreateTime,
		cutoff.CreateTime,
		cutoff.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanExcessOperationLogsPerUser 按用户清理超出上限的操作日志（每个用户最多保留 maxPerUser 条）
func CleanExcessOperationLogsPerUser(maxPerUser int) (int64, error) {
	if maxPerUser <= 0 {
		return 0, nil
	}
	var groups []struct {
		UserID uint64 `db:"user_id"`
		Cnt    int64  `db:"cnt"`
	}
	if err := db.DB.Select(&groups,
		"SELECT user_id, COUNT(*) AS cnt FROM operation_logs GROUP BY user_id HAVING COUNT(*) > ?",
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
			"SELECT id, create_time FROM operation_logs WHERE user_id = ? ORDER BY create_time DESC, id DESC LIMIT 1 OFFSET ?",
			g.UserID, maxPerUser-1,
		); err != nil {
			continue
		}
		result, err := db.Exec(
			"DELETE FROM operation_logs WHERE user_id = ? AND (create_time < ? OR (create_time = ? AND id < ?))",
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

// GetOperationLogStats 获取操作日志统计
type LogStats struct {
	TotalCount  int64        `db:"total_count" json:"total_count"`
	TodayCount  int64        `db:"today_count" json:"today_count"`
	ModuleStats []ModuleStat `json:"module_stats"`
	MethodStats []MethodStat `json:"method_stats"`
}

type ModuleStat struct {
	Module string `db:"module" json:"module"`
	Count  int64  `db:"count" json:"count"`
}

type MethodStat struct {
	Method string `db:"method" json:"method"`
	Count  int64  `db:"count" json:"count"`
}

// GetOperationLogStats 获取日志统计信息
func GetOperationLogStats() (*LogStats, error) {
	stats := &LogStats{}

	// 总数
	err := db.DB.Get(&stats.TotalCount, "SELECT COUNT(*) FROM operation_logs")
	if err != nil {
		return nil, err
	}

	// 今日数量
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err = db.DB.Get(&stats.TodayCount, "SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", todayStart)
	if err != nil {
		return nil, err
	}

	// 按模块统计
	err = db.DB.Select(&stats.ModuleStats, "SELECT module, COUNT(*) as count FROM operation_logs GROUP BY module ORDER BY count DESC LIMIT 10")
	if err != nil {
		return nil, err
	}

	// 按方法统计
	err = db.DB.Select(&stats.MethodStats, "SELECT method, COUNT(*) as count FROM operation_logs GROUP BY method ORDER BY count DESC")
	if err != nil {
		return nil, err
	}

	return stats, nil
}

