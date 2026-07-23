package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"fst/backend/pkg/panicsafe"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// SMSLog 短信日志
type SMSLog struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"column:user_id;not null;default:0;index:idx_sms_user_id" json:"user_id"`
	Phone        string    `gorm:"column:phone;size:32;not null;index:idx_sms_phone" json:"phone"`
	Provider     string    `gorm:"column:provider;size:32;not null;index:idx_sms_provider" json:"provider"`
	TemplateCode string    `gorm:"column:template_code;size:64;not null;default:''" json:"template_code"`
	TemplateName string    `gorm:"column:template_name;size:64;not null;default:'';index:idx_sms_template_name" json:"template_name"`
	Lang         string    `gorm:"column:lang;size:16;not null;default:'zh-CN'" json:"lang"`
	Content      string    `gorm:"column:content;size:512;not null;default:''" json:"content"`
	Status       uint8     `gorm:"column:status;not null;default:0;index:idx_sms_status" json:"status"`
	ErrorMsg     string    `gorm:"column:error_msg;size:512;not null;default:''" json:"error_msg"`
	RequestID    string    `gorm:"column:request_id;size:128;not null;default:''" json:"request_id"`
	Response     string    `gorm:"column:response;type:text" json:"response"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index:idx_sms_created_at" json:"created_at"`
}

// TableName 表名
func (SMSLog) TableName() string { return "sms_logs" }

var smsLogCleanupNextAt atomic.Int64

// CreateSMSLog 记录短信发送日志
func CreateSMSLog(logEntry *SMSLog) error {
	logEntry.CreatedAt = time.Now()
	if err := db.DB.Create(logEntry).Error; err != nil {
		return err
	}

	panicsafe.Go("SMSLog.aggregate", func() {
		if aggErr := RecordSMSLogAggregate(logEntry); aggErr != nil {
			log.Printf("[SMSLog] 汇总更新失败: %v", aggErr)
		}
		scheduleSMSLogRetentionCleanup()
	})
	return nil
}

func scheduleSMSLogRetentionCleanup() {
	scheduleLogRetentionCleanupGeneric(&smsLogCleanupNextAt, "sms_log", "SMSLog",
		CleanExcessSMSLogs, CleanExcessSMSLogsPerRecipient)
}

// SMSLogQuery 短信日志查询参数
type SMSLogQuery struct {
	Page         int    `form:"page" json:"page"`
	PageSize     int    `form:"page_size" json:"page_size"`
	UserID       uint64 `form:"user_id" json:"user_id"`
	Phone        string `form:"phone" json:"phone"`
	Provider     string `form:"provider" json:"provider"`
	TemplateName string `form:"template_name" json:"template_name"`
	Lang         string `form:"lang" json:"lang"`
	Status       int    `form:"status" json:"status"`
	StartTime    string `form:"start_time" json:"start_time"`
	EndTime      string `form:"end_time" json:"end_time"`
}

func buildSMSLogQuery(q *SMSLogQuery) *gorm.DB {
	query := db.DB.Model(&SMSLog{})
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if q.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+q.Phone+"%")
	}
	if q.Provider != "" {
		query = query.Where("provider = ?", q.Provider)
	}
	if q.TemplateName != "" {
		query = query.Where("template_name = ?", q.TemplateName)
	}
	if q.Lang != "" {
		query = query.Where("lang = ?", q.Lang)
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	if q.StartTime != "" {
		query = query.Where("created_at >= ?", q.StartTime)
	}
	if q.EndTime != "" {
		query = query.Where("created_at <= ?", q.EndTime)
	}
	return query
}

// GetSMSLogList 分页查询短信日志
func GetSMSLogList(q *SMSLogQuery) ([]SMSLog, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}

	base := buildSMSLogQuery(q)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []SMSLog
	err := base.Select("id, user_id, phone, provider, template_code, template_name, lang, content, status, error_msg, request_id, created_at").
		Order("created_at DESC, id DESC").
		Limit(q.PageSize).
		Offset((q.Page - 1) * q.PageSize).
		Find(&logs).Error
	return logs, total, err
}

// GetSMSLogByID 根据ID获取短信日志详情
func GetSMSLogByID(id uint64) (*SMSLog, error) {
	var logEntry SMSLog
	err := db.DB.Where("id = ?", id).First(&logEntry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

// DeleteSMSLogsBefore 删除指定时间之前的短信日志
func DeleteSMSLogsBefore(before string) (int64, error) {
	result := db.DB.Where("created_at < ?", before).Delete(&SMSLog{})
	return result.RowsAffected, result.Error
}

// CleanExcessSMSLogs 清理超出全局上限的旧短信日志
func CleanExcessSMSLogs(maxCount int) (int64, error) {
	return cleanExcessRowsGeneric("sms_logs", "created_at", maxCount)
}

// CleanExcessSMSLogsPerRecipient 按手机号清理超出上限的短信日志
func CleanExcessSMSLogsPerRecipient(maxPerRecipient int) (int64, error) {
	return cleanExcessRowsPerGroupGeneric[string]("sms_logs", "phone", "created_at", maxPerRecipient, "phone != ''")
}

// GetSMSLogStats 短信日志统计（兼容旧接口，优先读聚合表）
func GetSMSLogStats() (total int64, success int64, fail int64, err error) {
	stats, statsErr := GetSMSLogStatsDetail()
	if statsErr != nil {
		return 0, 0, 0, statsErr
	}
	return stats.TotalCount, stats.SuccessCount, stats.FailCount, nil
}

// GetSMSTemplateNames 获取短信日志中所有模板名（去重）
func GetSMSTemplateNames() ([]string, error) {
	var names []string
	err := db.DB.Model(&SMSLog{}).
		Distinct("template_name").
		Where("template_name != ''").
		Order("template_name").
		Pluck("template_name", &names).Error
	return names, err
}

// MaskPhone 脱敏手机号（兼容国内 11 位与国际 E.164）
func MaskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	prefix := 3
	if strings.HasPrefix(phone, "+") && len(phone) >= 12 {
		prefix = 4
	}
	return phone[:prefix] + "****" + phone[len(phone)-4:]
}
