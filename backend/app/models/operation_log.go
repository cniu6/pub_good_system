package models

import (
	"fmt"
	"fst/backend/internal/db"
	"log"
	"time"
)

// OperationLog 操作日志模型
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
	RequestBody  *string `db:"request_body" json:"request_body,omitempty"`
	ResponseBody *string `db:"response_body" json:"response_body,omitempty"`
	StatusCode   int     `db:"status_code" json:"status_code"`
	Duration     int     `db:"duration" json:"duration"` // 耗时(ms)
	CreateTime   *int64  `db:"create_time" json:"create_time"`
}

func (o *OperationLog) TableName() string {
	return "operation_logs"
}

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
			request_body MEDIUMTEXT COMMENT '请求体',
			response_body MEDIUMTEXT COMMENT '响应体',
			status_code INT NOT NULL DEFAULT 0 COMMENT '状态码',
			duration INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			INDEX idx_create_time_id (create_time, id),
			INDEX idx_user_create_time (user_id, create_time),
			INDEX idx_module_create_time (module, create_time),
			INDEX idx_action_create_time (action, create_time),
			INDEX idx_method_create_time (method, create_time),
			INDEX idx_ip_create_time (ip, create_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

		_, err := db.DB.Exec(schema)
		if err != nil {
			log.Printf("[Init] Failed to create operation_logs table: %v", err)
		} else {
			log.Println("[Init] Created operation_logs table")
		}
		return
	}

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
		{"request_body", "ALTER TABLE operation_logs ADD COLUMN request_body MEDIUMTEXT COMMENT '请求体' AFTER user_agent"},
		{"response_body", "ALTER TABLE operation_logs ADD COLUMN response_body MEDIUMTEXT COMMENT '响应体' AFTER request_body"},
		{"status_code", "ALTER TABLE operation_logs ADD COLUMN status_code INT NOT NULL DEFAULT 0 COMMENT '状态码' AFTER response_body"},
		{"duration", "ALTER TABLE operation_logs ADD COLUMN duration INT NOT NULL DEFAULT 0 COMMENT '耗时(ms)' AFTER status_code"},
		{"create_time", "ALTER TABLE operation_logs ADD COLUMN create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间' AFTER duration"},
	}

	for _, repair := range columnRepairs {
		if !db.CheckColumnExists("operation_logs", repair.column) {
			if _, err := db.DB.Exec(repair.alterSQL); err != nil {
				log.Printf("[Init] Failed to add operation_logs.%s: %v", repair.column, err)
			} else {
				log.Printf("[Init] Added operation_logs.%s", repair.column)
			}
		}
	}

	if db.CheckColumnExists("operation_logs", "created_at") && db.CheckColumnExists("operation_logs", "create_time") {
		_, _ = db.DB.Exec("UPDATE operation_logs SET create_time = UNIX_TIMESTAMP(created_at) WHERE create_time = 0 AND created_at IS NOT NULL")
	}

	indexRepairs := map[string]string{
		"idx_create_time_id":     "ALTER TABLE operation_logs ADD INDEX idx_create_time_id (create_time, id)",
		"idx_user_create_time":   "ALTER TABLE operation_logs ADD INDEX idx_user_create_time (user_id, create_time)",
		"idx_module_create_time": "ALTER TABLE operation_logs ADD INDEX idx_module_create_time (module, create_time)",
		"idx_action_create_time": "ALTER TABLE operation_logs ADD INDEX idx_action_create_time (action, create_time)",
		"idx_method_create_time": "ALTER TABLE operation_logs ADD INDEX idx_method_create_time (method, create_time)",
		"idx_ip_create_time":     "ALTER TABLE operation_logs ADD INDEX idx_ip_create_time (ip, create_time)",
	}

	for indexName, alterSQL := range indexRepairs {
		db.EnsureIndex("operation_logs", indexName, alterSQL)
	}
}

// ========== CRUD 操作 ==========

// CreateOperationLog 创建操作日志
func CreateOperationLog(log *OperationLog) error {
	query := `INSERT INTO operation_logs (user_id, username, module, action, method, path, ip,
			  user_agent, request_body, response_body, status_code, duration, create_time)
			  VALUES (:user_id, :username, :module, :action, :method, :path, :ip,
			  :user_agent, :request_body, :response_body, :status_code, :duration, :create_time)`

	now := time.Now().Unix()
	log.CreateTime = &now

	result, err := db.DB.NamedExec(query, log)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = uint64(id)
	return nil
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
	result, err := db.DB.Exec("DELETE FROM operation_logs WHERE create_time < ?", before_time)
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
	// 删除最旧的记录，只保留最新 maxCount 条
	// 注意：MySQL 子查询中 LIMIT 不支持参数化占位符，必须直接拼接
	query := fmt.Sprintf("DELETE FROM operation_logs WHERE id NOT IN (SELECT id FROM (SELECT id FROM operation_logs ORDER BY create_time DESC, id DESC LIMIT %d) AS t)", maxCount)
	result, err := db.DB.Exec(query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
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
	today_start := time.Now().Truncate(24 * time.Hour).Unix()
	err = db.DB.Get(&stats.TodayCount, "SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", today_start)
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
