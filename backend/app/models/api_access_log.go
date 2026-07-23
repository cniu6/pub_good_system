package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"strings"
	"time"

	"gorm.io/gorm"
)

// APIAccessLog API接口访问日志
type APIAccessLog struct {
	ID                  uint64  `gorm:"column:id;primaryKey;autoIncrement;index:idx_aal_create_time_id,priority:2" json:"id"`
	RequestID           string  `gorm:"column:request_id;size:36;not null;uniqueIndex:uk_request_id" json:"request_id"`
	UserID              uint64  `gorm:"column:user_id;not null;default:0;index:idx_aal_user_create_time,priority:1" json:"user_id"`
	Username            string  `gorm:"column:username;size:100;not null;default:''" json:"username"`
	Role                string  `gorm:"column:role;size:32;not null;default:''" json:"role"`
	AuthMethod          string  `gorm:"column:auth_method;size:16;not null;default:'';index:idx_aal_auth_method_create_time,priority:1" json:"auth_method"`
	Scene               string  `gorm:"column:scene;size:32;not null;default:'';index:idx_aal_scene_create_time,priority:1" json:"scene"`
	Method              string  `gorm:"column:method;size:20;not null;default:'';index:idx_aal_method_create_time,priority:1" json:"method"`
	Transport           string  `gorm:"column:transport;size:32;not null;default:'http';index:idx_aal_transport_create_time,priority:1" json:"transport"`
	Protocol            string  `gorm:"column:protocol;size:32;not null;default:''" json:"protocol"`
	Path                string  `gorm:"column:path;size:255;not null;default:''" json:"path"`
	RoutePath           string  `gorm:"column:route_path;size:255;not null;default:''" json:"route_path"`
	HandlerName         string  `gorm:"column:handler_name;size:255;not null;default:'';index:idx_aal_handler_create_time,priority:1" json:"handler_name"`
	RequestContentType  string  `gorm:"column:request_content_type;size:255;not null;default:''" json:"request_content_type"`
	ResponseContentType string  `gorm:"column:response_content_type;size:255;not null;default:''" json:"response_content_type"`
	QueryString         string  `gorm:"column:query_string;type:text" json:"query_string"`
	PathParams          *string `gorm:"column:path_params;type:text" json:"path_params,omitempty"`
	IP                  string  `gorm:"column:ip;size:64;not null;default:'';index:idx_aal_ip_create_time,priority:1" json:"ip"`
	SourceIP            string  `gorm:"column:source_ip;size:64;not null;default:''" json:"source_ip"`
	XIP                 string  `gorm:"column:x_ip;size:64;not null;default:''" json:"x_ip"`
	XForwardedFor       string  `gorm:"column:x_forwarded_for;size:1024;not null;default:''" json:"x_forwarded_for"`
	XRealIP             string  `gorm:"column:x_real_ip;size:64;not null;default:''" json:"x_real_ip"`
	UserAgent           string  `gorm:"column:user_agent;type:text" json:"user_agent"`
	Referer             string  `gorm:"column:referer;size:500;not null;default:''" json:"referer"`
	RequestHeaders      *string `gorm:"column:request_headers;type:mediumtext" json:"request_headers,omitempty"`
	RequestBody         *string `gorm:"column:request_body;type:mediumtext" json:"request_body,omitempty"`
	ResponseBody        *string `gorm:"column:response_body;type:mediumtext" json:"response_body,omitempty"`
	StatusCode          int     `gorm:"column:status_code;not null;default:0;index:idx_aal_status_create_time,priority:1" json:"status_code"`
	Duration            int     `gorm:"column:duration;not null;default:0" json:"duration"`
	RequestSize         int64   `gorm:"column:request_size;not null;default:0" json:"request_size"`
	ResponseSize        int64   `gorm:"column:response_size;not null;default:0" json:"response_size"`
	CreateTime          *int64  `gorm:"column:create_time;not null;default:0;index:idx_aal_create_time_id,priority:1;index:idx_aal_scene_create_time,priority:2;index:idx_aal_auth_method_create_time,priority:2;index:idx_aal_user_create_time,priority:2;index:idx_aal_method_create_time,priority:2;index:idx_aal_transport_create_time,priority:2;index:idx_aal_status_create_time,priority:2;index:idx_aal_ip_create_time,priority:2;index:idx_aal_handler_create_time,priority:2" json:"create_time"`
}

func (APIAccessLog) TableName() string {
	return "api_access_logs"
}

// CreateAPIAccessLog 创建 API 访问日志
func CreateAPIAccessLog(item *APIAccessLog) error {
	now := time.Now().Unix()
	item.CreateTime = &now
	return db.DB.Create(item).Error
}

// GetAPIAccessLogByID 按 ID 获取
func GetAPIAccessLogByID(id uint64) (*APIAccessLog, error) {
	var item APIAccessLog
	err := db.DB.Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetAPIAccessLogByRequestID 按 request_id 获取
func GetAPIAccessLogByRequestID(requestID string) (*APIAccessLog, error) {
	var item APIAccessLog
	err := db.DB.Where("request_id = ?", strings.TrimSpace(requestID)).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// APIAccessLogQuery 查询参数
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

func buildAPIAccessLogQuery(query *APIAccessLogQuery) *gorm.DB {
	q := db.DB.Model(&APIAccessLog{})
	if query.RequestID != "" {
		q = q.Where("request_id = ?", strings.TrimSpace(query.RequestID))
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		q = q.Where(
			"request_id LIKE ? OR path LIKE ? OR route_path LIKE ? OR username LIKE ? OR ip LIKE ? OR source_ip LIKE ? OR x_ip LIKE ? OR x_forwarded_for LIKE ? OR x_real_ip LIKE ? OR handler_name LIKE ? OR query_string LIKE ? OR transport LIKE ? OR protocol LIKE ? OR request_content_type LIKE ? OR response_content_type LIKE ?",
			keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword,
		)
	}
	if query.Scene != "" {
		q = q.Where("scene = ?", query.Scene)
	}
	if query.AuthMethod != "" {
		q = q.Where("auth_method = ?", strings.TrimSpace(query.AuthMethod))
	}
	if query.Transport != "" {
		q = q.Where("transport = ?", strings.TrimSpace(query.Transport))
	}
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Username != "" {
		q = q.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Method != "" {
		q = q.Where("method = ?", query.Method)
	}
	if query.Path != "" {
		keyword := "%" + query.Path + "%"
		q = q.Where("path LIKE ? OR route_path LIKE ?", keyword, keyword)
	}
	if query.IP != "" {
		q = q.Where("ip = ? OR source_ip = ? OR x_ip = ? OR x_real_ip = ? OR x_forwarded_for LIKE ?",
			query.IP, query.IP, query.IP, query.IP, "%"+query.IP+"%")
	}
	if query.StatusCode > 0 {
		q = q.Where("status_code = ?", query.StatusCode)
	}
	if query.StartTime > 0 {
		q = q.Where("create_time >= ?", query.StartTime)
	}
	if query.EndTime > 0 {
		q = q.Where("create_time <= ?", query.EndTime)
	}
	return q
}

// GetAPIAccessLogList 分页列表
func GetAPIAccessLogList(query *APIAccessLogQuery) ([]APIAccessLog, int64, error) {
	if query == nil {
		query = &APIAccessLogQuery{}
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	q := buildAPIAccessLogQuery(query)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []APIAccessLog
	err := q.Select(`id, request_id, user_id, username, role, auth_method, scene, method, transport, protocol, path, route_path, handler_name, request_content_type, response_content_type, query_string, ip, source_ip, x_ip, x_forwarded_for, x_real_ip, user_agent, referer, status_code, duration, request_size, response_size, create_time`).
		Order("create_time DESC, id DESC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&list).Error
	return list, total, err
}

// DeleteAPIAccessLogsBefore 删除指定时间之前的日志
func DeleteAPIAccessLogsBefore(beforeTime int64) (int64, error) {
	result := db.DB.Where("create_time < ?", beforeTime).Delete(&APIAccessLog{})
	return result.RowsAffected, result.Error
}

// CleanExcessAPIAccessLogs 清理超出全局上限的旧 API 访问日志
func CleanExcessAPIAccessLogs(maxCount int) (int64, error) {
	return cleanExcessRowsGeneric("api_access_logs", "create_time", maxCount)
}

// CleanExcessAPIAccessLogsPerUser 按用户清理超出上限的 API 访问日志
func CleanExcessAPIAccessLogsPerUser(maxPerUser int) (int64, error) {
	return cleanExcessRowsPerGroupGeneric[uint64]("api_access_logs", "user_id", "create_time", maxPerUser, "user_id > 0")
}

// APIAccessLogStats API 访问日志统计
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
	RoutePath   string  `gorm:"column:route_path" json:"route_path"`
	Count       int64   `gorm:"column:count" json:"count"`
	AvgDuration float64 `gorm:"column:avg_duration" json:"avg_duration"`
}

type APIAccessMethodStat struct {
	Method string `gorm:"column:method" json:"method"`
	Count  int64  `gorm:"column:count" json:"count"`
}

type APIAccessSceneStat struct {
	Scene string `gorm:"column:scene" json:"scene"`
	Count int64  `gorm:"column:count" json:"count"`
}

// GetAPIAccessLogStats 获取统计（优先读聚合表）
func GetAPIAccessLogStats() (*APIAccessLogStats, error) {
	return getAPIAccessLogStatsFromAggregate()
}

// GetAPIAccessLogStatsByUserID 当前用户自己的 API 访问日志统计
func GetAPIAccessLogStatsByUserID(userID uint64) (*APIAccessLogStats, error) {
	stats := &APIAccessLogStats{
		TopPaths:    []APIAccessPathStat{},
		MethodStats: []APIAccessMethodStat{},
	}
	todayStart := resolveAPIAccessLogStartOfLocalDay(time.Now()).Unix()
	// 用户中心统计仅统计 API Key 调用（不含 JWT 网页请求）
	base := db.DB.Model(&APIAccessLog{}).Where("user_id = ? AND auth_method = ?", userID, "apikey")

	if err := base.Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).Where("user_id = ? AND auth_method = ? AND create_time >= ?", userID, "apikey", todayStart).Count(&stats.TodayCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).Where("user_id = ? AND auth_method = ? AND status_code >= 200 AND status_code < 400", userID, "apikey").Count(&stats.SuccessCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).Where("user_id = ? AND auth_method = ? AND status_code >= 400 AND status_code < 500", userID, "apikey").Count(&stats.ClientErrorCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).Where("user_id = ? AND auth_method = ? AND status_code >= 500", userID, "apikey").Count(&stats.ServerErrorCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).Where("user_id = ? AND auth_method = ?", userID, "apikey").Select("COALESCE(AVG(duration), 0)").Scan(&stats.AvgDuration).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).
		Select("COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/') AS route_path, COUNT(*) AS count, COALESCE(AVG(duration), 0) AS avg_duration").
		Where("user_id = ? AND auth_method = ?", userID, "apikey").
		Group("COALESCE(NULLIF(COALESCE(NULLIF(route_path, ''), path), ''), '/')").
		Order("count DESC").
		Limit(10).
		Scan(&stats.TopPaths).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&APIAccessLog{}).
		Select("COALESCE(NULLIF(method, ''), 'UNKNOWN') AS method, COUNT(*) AS count").
		Where("user_id = ? AND auth_method = ?", userID, "apikey").
		Group("COALESCE(NULLIF(method, ''), 'UNKNOWN')").
		Order("count DESC").
		Scan(&stats.MethodStats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}
