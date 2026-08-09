package task

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// Trigger 触发来源
const (
	TriggerSchedule = "schedule"
	TriggerManual   = "manual"
)

// Run status
const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusTimeout = "timeout"
)

// Config keys（system_settings）
const (
	CfgEnabled            = "auto_job_enabled"
	CfgRunMaxCount        = "auto_job_run_max_count"
	CfgRetainErrors       = "auto_job_retain_errors"
	CfgAutoPrune          = "auto_job_auto_prune"
	CfgStuckAfterSec      = "auto_job_stuck_after_sec"
	CfgAutoKeepJobCodes   = "auto_job_auto_keep_job_codes"
	CfgAutoKeepCategories = "auto_job_auto_keep_categories"
)

// Handler keys
const (
	HandlerPruneAutoJobRuns          = "prune_auto_job_runs"
	HandlerMarkStuckAutoJobs         = "mark_stuck_auto_jobs"
	HandlerCleanupExpiredIdempotency = "cleanup_expired_idempotency"
	HandlerCleanupExpiredOrders      = "cleanup_expired_orders"
	HandlerCleanupSessionsCodes      = "cleanup_sessions_codes"
	HandlerReconcilePaymentOrders    = "reconcile_payment_orders"
	HandlerRefreshExchangeRates      = "refresh_exchange_rates"
)

// CurrencyRefreshDynamicRatesFn 由 services 注入，避免 task 与 services 循环依赖
var CurrencyRefreshDynamicRatesFn func(ctx context.Context) (map[string]float64, error)

// JobDefinition 任务定义
// 注意：MySQL 上 string 缺 size 会迁成 longtext；主键/索引列必须带 size，否则 Error 1170。
type JobDefinition struct {
	JobCode         string `gorm:"column:job_code;primaryKey;size:64" json:"job_code"`
	Name            string `gorm:"column:name;size:128" json:"name"`
	Description     string `gorm:"column:description;type:text" json:"description"`
	Category        string `gorm:"column:category;size:64" json:"category"`
	HandlerKey      string `gorm:"column:handler_key;size:64" json:"handler_key"`
	CronExpr        string `gorm:"column:cron_expr;size:128" json:"cron_expr"`
	IntervalSeconds int    `gorm:"column:interval_seconds" json:"interval_seconds"`
	Timezone        string `gorm:"column:timezone;size:64" json:"timezone"`
	Enabled         int    `gorm:"column:enabled;index:idx_auto_job_def_enabled" json:"enabled"`
	TimeoutSec      int    `gorm:"column:timeout_sec" json:"timeout_sec"`
	// MaxConcurrency 表字段占位：当前固定为 1，Trigger 用进程内互斥，不按该值做多并发
	MaxConcurrency int `gorm:"column:max_concurrency" json:"max_concurrency"`
	// ParamsJSON 预留：管理端可编辑，当前内置 handler 均未读取
	ParamsJSON           string `gorm:"column:params_json;type:text" json:"params_json"`
	LastStatus           string `gorm:"column:last_status;size:32" json:"last_status"`
	LastStartedAt        int64  `gorm:"column:last_started_at" json:"last_started_at"`
	LastFinishedAt       int64  `gorm:"column:last_finished_at" json:"last_finished_at"`
	LastError            string `gorm:"column:last_error;type:text" json:"last_error"`
	LifetimeRunCount     string `gorm:"column:lifetime_run_count;size:64" json:"lifetime_run_count"`
	LifetimeSuccessCount string `gorm:"column:lifetime_success_count;size:64" json:"lifetime_success_count"`
	LifetimeFailCount    string `gorm:"column:lifetime_fail_count;size:64" json:"lifetime_fail_count"`
	CreateTime           int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime           int64  `gorm:"column:update_time" json:"update_time"`
}

// TableName GORM 表名
func (JobDefinition) TableName() string { return "auto_job_definitions" }

// JobRun 执行记录
type JobRun struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RunUID      string `gorm:"column:run_uid;size:36;not null;default:''" json:"run_uid"`
	JobCode     string `gorm:"column:job_code;size:64;index:idx_auto_job_runs_job_started,priority:1" json:"job_code"`
	Category    string `gorm:"column:category;size:64" json:"category"`
	TriggerType string `gorm:"column:trigger_type;size:32" json:"trigger"`
	Status      string `gorm:"column:status;size:32" json:"status"`
	StartedAt   int64  `gorm:"column:started_at;index:idx_auto_job_runs_job_started,priority:2" json:"started_at"`
	FinishedAt  int64  `gorm:"column:finished_at" json:"finished_at"`
	DurationMs  int64  `gorm:"column:duration_ms" json:"duration_ms"`
	Message     string `gorm:"column:message;type:text" json:"message"`
	DetailJSON  string `gorm:"column:detail_json;type:text" json:"detail_json"`
	ErrorText   string `gorm:"column:error_text;type:text" json:"error_text"`
	KeepForever int    `gorm:"column:keep_forever" json:"keep_forever"`
	Operator    string `gorm:"column:operator;size:64" json:"operator"`
}

// TableName GORM 表名
func (JobRun) TableName() string { return "auto_job_runs" }

// JobRunKeep 被标记保留的执行记录副本（独立表，方便检索/长期保存）
type JobRunKeep struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RunUID       string `gorm:"column:run_uid;size:36;not null;default:''" json:"run_uid"`
	JobCode      string `gorm:"column:job_code;size:64;index:idx_auto_job_keeps_job_started,priority:1" json:"job_code"`
	Category     string `gorm:"column:category;size:64" json:"category"`
	TriggerType  string `gorm:"column:trigger_type;size:32" json:"trigger"`
	Status       string `gorm:"column:status;size:32" json:"status"`
	StartedAt    int64  `gorm:"column:started_at;index:idx_auto_job_keeps_job_started,priority:2" json:"started_at"`
	FinishedAt   int64  `gorm:"column:finished_at" json:"finished_at"`
	DurationMs   int64  `gorm:"column:duration_ms" json:"duration_ms"`
	Message      string `gorm:"column:message;type:text" json:"message"`
	DetailJSON   string `gorm:"column:detail_json;type:text" json:"detail_json"`
	ErrorText    string `gorm:"column:error_text;type:text" json:"error_text"`
	Operator     string `gorm:"column:operator;size:64" json:"operator"`
	SourceRunID  uint64 `gorm:"column:source_run_id" json:"source_run_id"`
	KeptAt       int64  `gorm:"column:kept_at" json:"kept_at"`
	RunTimestamp int64  `gorm:"column:run_timestamp" json:"run_timestamp"`
}

// TableName GORM 表名
func (JobRunKeep) TableName() string { return "auto_job_runs_keep" }

// HandlerResult handler 返回
type HandlerResult struct {
	Message string
	Detail  map[string]interface{}
	// Quiet：调度空跑成功不落执行记录（仍累计 lifetime）
	Quiet bool
}

// JobHandler 业务 handler
type JobHandler func(ctx context.Context, job *JobDefinition) (*HandlerResult, error)

// RunOptions 触发选项
type RunOptions struct {
	Trigger  string
	Operator string
	Force    bool // 手动忽略全局开关
}

// GlobalConfig 运行时配置
type GlobalConfig struct {
	Enabled            bool     `json:"auto_job_enabled"`
	RunMaxCount        int      `json:"auto_job_run_max_count"`
	RetainErrors       bool     `json:"auto_job_retain_errors"`
	AutoPrune          bool     `json:"auto_job_auto_prune"`
	StuckAfterSec      int      `json:"auto_job_stuck_after_sec"`
	AutoKeepJobCodes   []string `json:"auto_job_auto_keep_job_codes"`
	AutoKeepCategories []string `json:"auto_job_auto_keep_categories"`
}

// Overview 总览卡片
type Overview struct {
	EnabledJobs        int64  `json:"enabled_jobs"`
	TotalJobs          int64  `json:"total_jobs"`
	RunningCount       int64  `json:"running_count"`
	TodaySuccess       int64  `json:"today_success"`
	TodayFailed        int64  `json:"today_failed"`
	LifetimeRunTotal   string `json:"lifetime_run_total"`
	RunRowCount        int64  `json:"run_row_count"`
	RunMaxCount        int    `json:"run_max_count"`
	SchedulerRunning   bool   `json:"scheduler_running"`
	SchedulerUptimeSec int64  `json:"scheduler_uptime_sec"`
	LastTickAt         int64  `json:"last_tick_at"`
	GlobalEnabled      bool   `json:"global_enabled"`
}

// CleanRunsRequest 清理请求（仅 success/failed/all）
type CleanRunsRequest struct {
	Scope   string `json:"scope"` // success|failed|all
	JobCode string `json:"job_code"`
}

// ImportPresetsRequest 导入预设
type ImportPresetsRequest struct {
	Mode string `json:"mode"` // skip|update
}

// UpdateJobRequest 更新任务
type UpdateJobRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	CronExpr        *string `json:"cron_expr"`
	IntervalSeconds *int    `json:"interval_seconds"`
	Timezone        *string `json:"timezone"`
	Enabled         *bool   `json:"enabled"`
	TimeoutSec      *int    `json:"timeout_sec"`
	ParamsJSON      *string `json:"params_json"`
}

// MarkKeepRequest 标记保留
type MarkKeepRequest struct {
	IDs         []uint64 `json:"ids"`
	KeepForever bool     `json:"keep_forever"`
}

func nowUnix() int64 {
	return time.Now().Unix()
}

// newRunUID 生成符合 CHAR(36) 的 UUID 形态对外键（勿再用 jobCode+nano，会超长截断）
func newRunUID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		// 极罕见：退化为时间戳填充，仍保持 36 字符
		n := uint64(time.Now().UnixNano())
		for i := 0; i < 8; i++ {
			raw[i] = byte(n >> (56 - 8*i))
			raw[8+i] = byte(n >> (8 * i))
		}
	}
	raw[6] = (raw[6] & 0x0f) | 0x40 // version 4
	raw[8] = (raw[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:16])
}

func marshalDetail(detail map[string]interface{}) string {
	if detail == nil {
		return ""
	}
	b, err := json.Marshal(detail)
	if err != nil {
		return ""
	}
	return string(b)
}
