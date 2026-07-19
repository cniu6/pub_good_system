package models

import (
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// SMSLog 短信日志
type SMSLog struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"user_id"` // 关联用户（匿名发送为 0）
	Phone        string    `db:"phone" json:"phone"`
	Provider     string    `db:"provider" json:"provider"` // aliyun, tencent, custom
	TemplateCode string    `db:"template_code" json:"template_code"`
	TemplateName string    `db:"template_name" json:"template_name"`
	Lang         string    `db:"lang" json:"lang"`                 // zh-CN, en-US, etc.
	Content      string    `db:"content" json:"content"`           // 实际发送的短信内容（脱敏）
	Status       uint8     `db:"status" json:"status"`             // 1=成功, 0=失败
	ErrorMsg     string    `db:"error_msg" json:"error_msg"`       // 错误信息
	RequestID    string    `db:"request_id" json:"request_id"`     // 服务商返回的请求ID
	Response     string    `db:"response" json:"response"`         // 完整响应（JSON）
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

var smsLogCleanupNextAt atomic.Int64

// InitSMSTable 初始化短信日志表
func InitSMSTable() {
	if !db.CheckTableExists("sms_logs") {
		schema := `CREATE TABLE IF NOT EXISTS sms_logs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联用户ID（匿名发送为0）',
			phone VARCHAR(32) NOT NULL COMMENT '手机号（脱敏）',
			provider VARCHAR(32) NOT NULL COMMENT '服务商: aliyun, tencent, custom',
			template_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板ID',
			template_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模板名称',
			lang VARCHAR(16) NOT NULL DEFAULT 'zh-CN' COMMENT '语言',
			content VARCHAR(512) NOT NULL DEFAULT '' COMMENT '发送内容（脱敏）',
			status TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态: 1=成功, 0=失败',
			error_msg VARCHAR(512) NOT NULL DEFAULT '' COMMENT '错误信息',
			request_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT '服务商请求ID',
			response TEXT COMMENT '完整响应',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			INDEX idx_phone (phone),
			INDEX idx_user_id (user_id),
			INDEX idx_provider (provider),
			INDEX idx_template_name (template_name),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
		_, err := db.Exec(schema)
		if err != nil {
			log.Printf("[Init] Failed to create sms_logs table: %v", err)
		} else {
			log.Println("[Init] Created sms_logs table")
		}
	} else {
		// 兼容旧表：补齐 user_id 字段
		if !db.CheckColumnExists("sms_logs", "user_id") {
			if _, err := db.Exec("ALTER TABLE sms_logs ADD COLUMN user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联用户ID（匿名发送为0）' AFTER id"); err != nil {
				log.Printf("[Init] Failed to add sms_logs.user_id: %v", err)
			} else {
				log.Println("[Init] Added sms_logs.user_id")
			}
		}
		db.EnsureIndex("sms_logs", "idx_user_id", "ALTER TABLE sms_logs ADD INDEX idx_user_id (user_id)")
	}

	InitSMSLogAggregateTables()
}

// CreateSMSLog 记录短信发送日志
func CreateSMSLog(logEntry *SMSLog) error {
	query := `INSERT INTO sms_logs (user_id, phone, provider, template_code, template_name, lang, content, status, error_msg, request_id, response)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query,
		logEntry.UserID, logEntry.Phone, logEntry.Provider, logEntry.TemplateCode, logEntry.TemplateName,
		logEntry.Lang, logEntry.Content, logEntry.Status, logEntry.ErrorMsg,
		logEntry.RequestID, logEntry.Response,
	)
	if err != nil {
		return err
	}
	if id, idErr := result.LastInsertId(); idErr == nil {
		logEntry.ID = uint64(id)
	}
	logEntry.CreatedAt = time.Now()

	// 异步更新聚合统计 + 触发保留清理（节流）
	go func(item *SMSLog) {
		if aggErr := RecordSMSLogAggregate(item); aggErr != nil {
			log.Printf("[SMSLog] 汇总更新失败: %v", aggErr)
		}
		scheduleSMSLogRetentionCleanup()
	}(logEntry)

	return nil
}

func scheduleSMSLogRetentionCleanup() {
	now := time.Now().UnixNano()
	nextAt := smsLogCleanupNextAt.Load()
	if nextAt > now {
		return
	}
	if !smsLogCleanupNextAt.CompareAndSwap(nextAt, now+int64(30*time.Second)) {
		return
	}

	cfg := loadSMSLogRetentionConfig()
	if cfg.MaxCount > 0 {
		if _, err := CleanExcessSMSLogs(cfg.MaxCount); err != nil {
			log.Printf("[SMSLog] 自动清理超限日志失败: %v", err)
		}
	}
	if cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0 {
		if _, err := CleanExcessSMSLogsPerRecipient(cfg.PerUserMaxCount); err != nil {
			log.Printf("[SMSLog] 按收件人清理超限日志失败: %v", err)
		}
	}
}

type smsLogRetentionConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

func loadSMSLogRetentionConfig() smsLogRetentionConfig {
	cfg := smsLogRetentionConfig{MaxCount: 1000, PerUserMaxCount: 1000}
	settingsMap, err := GetSettingsMap([]string{
		"sms_log_max_count",
		"sms_log_per_user_limit_enabled",
		"sms_log_per_user_max_count",
	})
	if err != nil {
		return cfg
	}
	if v, ok := settingsMap["sms_log_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.MaxCount = n
		}
	}
	if v, ok := settingsMap["sms_log_per_user_limit_enabled"]; ok {
		lower := strings.ToLower(strings.TrimSpace(v))
		cfg.PerUserLimitEnabled = lower == "true" || lower == "1"
	}
	if v, ok := settingsMap["sms_log_per_user_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.PerUserMaxCount = n
		}
	}
	return cfg
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
	Status       int    `form:"status" json:"status"` // -1=全部, 0=失败, 1=成功
	StartTime    string `form:"start_time" json:"start_time"`
	EndTime      string `form:"end_time" json:"end_time"`
}

// GetSMSLogList 分页查询短信日志
func GetSMSLogList(q *SMSLogQuery) ([]SMSLog, int64, error) {
	var logs []SMSLog
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if q.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, q.UserID)
	}
	if q.Phone != "" {
		where += " AND phone LIKE ?"
		args = append(args, "%"+q.Phone+"%")
	}
	if q.Provider != "" {
		where += " AND provider = ?"
		args = append(args, q.Provider)
	}
	if q.TemplateName != "" {
		where += " AND template_name = ?"
		args = append(args, q.TemplateName)
	}
	if q.Lang != "" {
		where += " AND lang = ?"
		args = append(args, q.Lang)
	}
	if q.Status >= 0 {
		where += " AND status = ?"
		args = append(args, q.Status)
	}
	if q.StartTime != "" {
		where += " AND created_at >= ?"
		args = append(args, q.StartTime)
	}
	if q.EndTime != "" {
		where += " AND created_at <= ?"
		args = append(args, q.EndTime)
	}

	err := db.DB.Get(&total, "SELECT COUNT(*) FROM sms_logs "+where, args...)
	if err != nil {
		return nil, 0, err
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	offset := (q.Page - 1) * q.PageSize

	listSQL := "SELECT id, user_id, phone, provider, template_code, template_name, lang, content, status, error_msg, request_id, created_at FROM sms_logs " +
		where + " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, q.PageSize, offset)

	err = db.DB.Select(&logs, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetSMSLogByID 根据ID获取短信日志详情
func GetSMSLogByID(id uint64) (*SMSLog, error) {
	var logEntry SMSLog
	err := db.DB.Get(&logEntry, "SELECT * FROM sms_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

// DeleteSMSLogsBefore 删除指定时间之前的短信日志
func DeleteSMSLogsBefore(before string) (int64, error) {
	result, err := db.Exec("DELETE FROM sms_logs WHERE created_at < ?", before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanExcessSMSLogs 清理超出全局上限的旧短信日志
func CleanExcessSMSLogs(maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM sms_logs"); err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	var cutoff struct {
		ID        uint64    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}
	if err := db.DB.Get(&cutoff,
		"SELECT id, created_at FROM sms_logs ORDER BY created_at DESC, id DESC LIMIT 1 OFFSET ?",
		maxCount-1,
	); err != nil {
		return 0, err
	}

	result, err := db.Exec(
		"DELETE FROM sms_logs WHERE created_at < ? OR (created_at = ? AND id < ?)",
		cutoff.CreatedAt, cutoff.CreatedAt, cutoff.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanExcessSMSLogsPerRecipient 按手机号清理超出上限的短信日志
func CleanExcessSMSLogsPerRecipient(maxPerRecipient int) (int64, error) {
	if maxPerRecipient <= 0 {
		return 0, nil
	}
	var groups []struct {
		Phone string `db:"phone"`
		Cnt   int64  `db:"cnt"`
	}
	if err := db.DB.Select(&groups,
		"SELECT phone, COUNT(*) AS cnt FROM sms_logs WHERE phone != '' GROUP BY phone HAVING COUNT(*) > ?",
		maxPerRecipient,
	); err != nil {
		return 0, err
	}

	var totalAffected int64
	for _, g := range groups {
		var cutoff struct {
			ID        uint64    `db:"id"`
			CreatedAt time.Time `db:"created_at"`
		}
		if err := db.DB.Get(&cutoff,
			"SELECT id, created_at FROM sms_logs WHERE phone = ? ORDER BY created_at DESC, id DESC LIMIT 1 OFFSET ?",
			g.Phone, maxPerRecipient-1,
		); err != nil {
			continue
		}
		result, err := db.Exec(
			"DELETE FROM sms_logs WHERE phone = ? AND (created_at < ? OR (created_at = ? AND id < ?))",
			g.Phone, cutoff.CreatedAt, cutoff.CreatedAt, cutoff.ID,
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
	err := db.DB.Select(&names, "SELECT DISTINCT template_name FROM sms_logs WHERE template_name != '' ORDER BY template_name")
	if err != nil {
		return nil, err
	}
	return names, nil
}

// MaskPhone 脱敏手机号（兼容国内 11 位与国际 E.164）
func MaskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	// 保留前 3 后 4，中间打码；国际号过长时前缀多留一点国家码可读性
	prefix := 3
	if strings.HasPrefix(phone, "+") && len(phone) >= 12 {
		prefix = 4
	}
	return phone[:prefix] + "****" + phone[len(phone)-4:]
}
