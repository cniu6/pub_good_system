package models

import (
	"fst/backend/internal/db"
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

	list_query := "SELECT * FROM operation_logs " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
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

// GetOperationLogStats 获取操作日志统计
type LogStats struct {
	TotalCount   int64 `db:"total_count" json:"total_count"`
	TodayCount   int64 `db:"today_count" json:"today_count"`
	ModuleStats  []ModuleStat  `json:"module_stats"`
	MethodStats  []MethodStat  `json:"method_stats"`
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
