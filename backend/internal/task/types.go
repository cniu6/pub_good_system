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
	CfgEnabled       = "auto_job_enabled"
	CfgRunMaxCount   = "auto_job_run_max_count"
	CfgRetainErrors  = "auto_job_retain_errors"
	CfgAutoPrune     = "auto_job_auto_prune"
	CfgStuckAfterSec = "auto_job_stuck_after_sec"
)

// Handler keys
const (
	HandlerPruneAutoJobRuns          = "prune_auto_job_runs"
	HandlerMarkStuckAutoJobs         = "mark_stuck_auto_jobs"
	HandlerCleanupExpiredIdempotency = "cleanup_expired_idempotency"
	HandlerCleanupExpiredOrders      = "cleanup_expired_orders"
	HandlerCleanupSessionsCodes      = "cleanup_sessions_codes"
)

// JobDefinition 任务定义
type JobDefinition struct {
	JobCode              string `db:"job_code" json:"job_code"`
	Name                 string `db:"name" json:"name"`
	Description          string `db:"description" json:"description"`
	Category             string `db:"category" json:"category"`
	HandlerKey           string `db:"handler_key" json:"handler_key"`
	CronExpr             string `db:"cron_expr" json:"cron_expr"`
	IntervalSeconds      int    `db:"interval_seconds" json:"interval_seconds"`
	Timezone             string `db:"timezone" json:"timezone"`
	Enabled         int    `db:"enabled" json:"enabled"`
	TimeoutSec      int    `db:"timeout_sec" json:"timeout_sec"`
	// MaxConcurrency 表字段占位：当前固定为 1，Trigger 用进程内互斥，不按该值做多并发
	MaxConcurrency int `db:"max_concurrency" json:"max_concurrency"`
	// ParamsJSON 预留：管理端可编辑，当前内置 handler 均未读取
	ParamsJSON           string `db:"params_json" json:"params_json"`
	LastStatus           string `db:"last_status" json:"last_status"`
	LastStartedAt        int64  `db:"last_started_at" json:"last_started_at"`
	LastFinishedAt       int64  `db:"last_finished_at" json:"last_finished_at"`
	LastError            string `db:"last_error" json:"last_error"`
	LifetimeRunCount     string `db:"lifetime_run_count" json:"lifetime_run_count"`
	LifetimeSuccessCount string `db:"lifetime_success_count" json:"lifetime_success_count"`
	LifetimeFailCount    string `db:"lifetime_fail_count" json:"lifetime_fail_count"`
	CreateTime           int64  `db:"create_time" json:"create_time"`
	UpdateTime           int64  `db:"update_time" json:"update_time"`
}

// JobRun 执行记录
type JobRun struct {
	ID          uint64 `db:"id" json:"id"`
	RunUID      string `db:"run_uid" json:"run_uid"`
	JobCode     string `db:"job_code" json:"job_code"`
	Category    string `db:"category" json:"category"`
	TriggerType string `db:"trigger_type" json:"trigger"`
	Status      string `db:"status" json:"status"`
	StartedAt   int64  `db:"started_at" json:"started_at"`
	FinishedAt  int64  `db:"finished_at" json:"finished_at"`
	DurationMs  int64  `db:"duration_ms" json:"duration_ms"`
	Message     string `db:"message" json:"message"`
	DetailJSON  string `db:"detail_json" json:"detail_json"`
	ErrorText   string `db:"error_text" json:"error_text"`
	KeepForever int    `db:"keep_forever" json:"keep_forever"`
	Operator    string `db:"operator" json:"operator"`
}

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
	Enabled       bool `json:"auto_job_enabled"`
	RunMaxCount   int  `json:"auto_job_run_max_count"`
	RetainErrors  bool `json:"auto_job_retain_errors"`
	AutoPrune     bool `json:"auto_job_auto_prune"`
	StuckAfterSec int  `json:"auto_job_stuck_after_sec"`
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
