package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"fst/backend/pkg/panicsafe"
	"log"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// OperationLog 操作日志模型
type OperationLog struct {
	ID           uint64  `gorm:"column:id;primaryKey;autoIncrement;index:idx_op_create_time_id,priority:2" json:"id"`
	UserID       uint64  `gorm:"column:user_id;not null;default:0;index:idx_op_user_create_time,priority:1" json:"user_id"`
	Username     string  `gorm:"column:username;size:100;not null;default:''" json:"username"`
	Module       string  `gorm:"column:module;size:100;not null;default:'';index:idx_op_module_create_time,priority:1" json:"module"`
	Action       string  `gorm:"column:action;size:100;not null;default:'';index:idx_op_action_create_time,priority:1" json:"action"`
	Method       string  `gorm:"column:method;size:20;not null;default:'';index:idx_op_method_create_time,priority:1" json:"method"`
	Path         string  `gorm:"column:path;size:255;not null;default:''" json:"path"`
	IP           string  `gorm:"column:ip;size:45;not null;default:'';index:idx_op_ip_create_time,priority:1" json:"ip"`
	UserAgent    string  `gorm:"column:user_agent;type:text" json:"user_agent"`
	HandlerName  string  `gorm:"column:handler_name;size:255;not null;default:'';index:idx_op_handler_create_time,priority:1" json:"handler_name"`
	RequestBody  *string `gorm:"column:request_body;type:mediumtext" json:"request_body,omitempty"`
	ResponseBody *string `gorm:"column:response_body;type:mediumtext" json:"response_body,omitempty"`
	StatusCode   int     `gorm:"column:status_code;not null;default:0" json:"status_code"`
	Duration     int     `gorm:"column:duration;not null;default:0" json:"duration"`
	CreateTime   *int64  `gorm:"column:create_time;not null;default:0;index:idx_op_create_time_id,priority:1;index:idx_op_user_create_time,priority:2;index:idx_op_module_create_time,priority:2;index:idx_op_action_create_time,priority:2;index:idx_op_method_create_time,priority:2;index:idx_op_ip_create_time,priority:2;index:idx_op_handler_create_time,priority:2" json:"create_time"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

var operationLogCleanupNextAt atomic.Int64

// CreateOperationLog 创建操作日志
func CreateOperationLog(item *OperationLog) error {
	now := time.Now().Unix()
	item.CreateTime = &now
	if err := db.DB.Create(item).Error; err != nil {
		return err
	}

	panicsafe.Go("OperationLog.aggregate", func() {
		if aggErr := RecordOperationLogAggregate(item); aggErr != nil {
			log.Printf("[OperationLog] 汇总更新失败: %v", aggErr)
		}
		scheduleOperationLogRetentionCleanup()
	})
	return nil
}

func scheduleOperationLogRetentionCleanup() {
	scheduleLogRetentionCleanupGeneric(&operationLogCleanupNextAt, "operation_log", "OperationLog",
		CleanExcessOperationLogs, CleanExcessOperationLogsPerUser)
}

// GetOperationLogByID 根据ID获取日志
func GetOperationLogByID(id uint64) (*OperationLog, error) {
	var item OperationLog
	err := db.DB.Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
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

func buildOperationLogQuery(q *OperationLogQuery) *gorm.DB {
	query := db.DB.Model(&OperationLog{})
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if q.Username != "" {
		query = query.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Module != "" {
		query = query.Where("module = ?", q.Module)
	}
	if q.Action != "" {
		query = query.Where("action = ?", q.Action)
	}
	if q.Method != "" {
		query = query.Where("method = ?", q.Method)
	}
	if q.Path != "" {
		query = query.Where("path LIKE ?", "%"+q.Path+"%")
	}
	if q.IP != "" {
		query = query.Where("ip = ?", q.IP)
	}
	if q.StartTime > 0 {
		query = query.Where("create_time >= ?", q.StartTime)
	}
	if q.EndTime > 0 {
		query = query.Where("create_time <= ?", q.EndTime)
	}
	return query
}

// GetOperationLogList 获取日志列表
func GetOperationLogList(query *OperationLogQuery) ([]OperationLog, int64, error) {
	if query == nil {
		query = &OperationLogQuery{}
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	q := buildOperationLogQuery(query)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []OperationLog
	err := q.Order("create_time DESC, id DESC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&logs).Error
	return logs, total, err
}

// DeleteOperationLogsBefore 删除指定时间之前的日志
func DeleteOperationLogsBefore(beforeTime int64) (int64, error) {
	result := db.DB.Where("create_time < ?", beforeTime).Delete(&OperationLog{})
	return result.RowsAffected, result.Error
}

// CleanExcessOperationLogs 清理超出上限的旧日志
func CleanExcessOperationLogs(maxCount int) (int64, error) {
	return cleanExcessRowsGeneric("operation_logs", "create_time", maxCount)
}

// CleanExcessOperationLogsPerUser 按用户清理超出上限的操作日志
func CleanExcessOperationLogsPerUser(maxPerUser int) (int64, error) {
	return cleanExcessRowsPerGroupGeneric[uint64]("operation_logs", "user_id", "create_time", maxPerUser, "")
}

// LogStats 操作日志统计
type LogStats struct {
	TotalCount  int64        `json:"total_count"`
	TodayCount  int64        `json:"today_count"`
	ModuleStats []ModuleStat `json:"module_stats"`
	MethodStats []MethodStat `json:"method_stats"`
}

type ModuleStat struct {
	Module string `gorm:"column:module" json:"module"`
	Count  int64  `gorm:"column:count" json:"count"`
}

type MethodStat struct {
	Method string `gorm:"column:method" json:"method"`
	Count  int64  `gorm:"column:count" json:"count"`
}

// GetOperationLogStats 获取日志统计信息
func GetOperationLogStats() (*LogStats, error) {
	stats := &LogStats{}

	if err := db.DB.Model(&OperationLog{}).Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	if err := db.DB.Model(&OperationLog{}).Where("create_time >= ?", todayStart).Count(&stats.TodayCount).Error; err != nil {
		return nil, err
	}

	if err := db.DB.Model(&OperationLog{}).
		Select("module, COUNT(*) as count").
		Group("module").
		Order("count DESC").
		Limit(10).
		Scan(&stats.ModuleStats).Error; err != nil {
		return nil, err
	}

	if err := db.DB.Model(&OperationLog{}).
		Select("method, COUNT(*) as count").
		Group("method").
		Order("count DESC").
		Scan(&stats.MethodStats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
