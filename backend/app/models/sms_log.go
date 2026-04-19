package models

import (
	"fst/backend/pkg/db"
	"log"
	"time"
)

// SMSLog 短信日志
type SMSLog struct {
	ID           uint64    `db:"id" json:"id"`
	Phone        string    `db:"phone" json:"phone"`
	Provider     string    `db:"provider" json:"provider"` // aliyun, tencent, custom
	TemplateCode string    `db:"template_code" json:"template_code"`
	TemplateName string    `db:"template_name" json:"template_name"`
	Lang         string    `db:"lang" json:"lang"` // zh-CN, en-US, etc.
	Content      string    `db:"content" json:"content"`      // 实际发送的短信内容（脱敏）
	Status       uint8     `db:"status" json:"status"`       // 1=成功, 0=失败
	ErrorMsg     string    `db:"error_msg" json:"error_msg"` // 错误信息
	RequestID    string    `db:"request_id" json:"request_id"` // 服务商返回的请求ID
	Response     string    `db:"response" json:"response"`     // 完整响应（JSON）
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// InitSMSTable 初始化短信日志表
func InitSMSTable() {
	if !db.CheckTableExists("sms_logs") {
		schema := `CREATE TABLE IF NOT EXISTS sms_logs (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
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
			INDEX idx_provider (provider),
			INDEX idx_template_name (template_name),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
		_, err := db.DB.Exec(schema)
		if err != nil {
			log.Printf("[Init] Failed to create sms_logs table: %v", err)
		} else {
			log.Println("[Init] Created sms_logs table")
		}
	}
}

// CreateSMSLog 记录短信发送日志
func CreateSMSLog(log *SMSLog) error {
	query := `INSERT INTO sms_logs (phone, provider, template_code, template_name, lang, content, status, error_msg, request_id, response)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query,
		log.Phone, log.Provider, log.TemplateCode, log.TemplateName,
		log.Lang, log.Content, log.Status, log.ErrorMsg,
		log.RequestID, log.Response,
	)
	return err
}

// SMSLogQuery 短信日志查询参数
type SMSLogQuery struct {
	Page         int    `form:"page" json:"page"`
	PageSize     int    `form:"page_size" json:"page_size"`
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

	listSQL := "SELECT id, phone, provider, template_code, template_name, lang, content, status, error_msg, request_id, created_at FROM sms_logs " +
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
	var log SMSLog
	err := db.DB.Get(&log, "SELECT * FROM sms_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// DeleteSMSLogsBefore 删除指定时间之前的短信日志
func DeleteSMSLogsBefore(before string) (int64, error) {
	result, err := db.DB.Exec("DELETE FROM sms_logs WHERE created_at < ?", before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetSMSLogStats 短信日志统计
func GetSMSLogStats() (total int64, success int64, fail int64, err error) {
	err = db.DB.Get(&total, "SELECT COUNT(*) FROM sms_logs")
	if err != nil {
		return
	}
	err = db.DB.Get(&success, "SELECT COUNT(*) FROM sms_logs WHERE status = 1")
	if err != nil {
		return
	}
	fail = total - success
	return
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

// MaskPhone 脱敏手机号
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

