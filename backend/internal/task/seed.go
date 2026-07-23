package task

import (
	"fmt"
)

// PresetJob 默认任务预设
type PresetJob struct {
	JobCode         string
	Name            string
	Description     string
	Category        string
	HandlerKey      string
	CronExpr        string
	IntervalSeconds int
	Timezone        string
	Enabled         bool
	TimeoutSec      int
	ParamsJSON      string
}

// DefaultPresets 库空或导入时用的默认任务（管理系统）
func DefaultPresets() []PresetJob {
	return []PresetJob{
		{
			JobCode:         "prune_auto_job_runs",
			Name:            "执行记录自动修剪",
			Description:     "按记录上限删除最旧成功记录，默认保留错误/keep_forever",
			Category:        "maintenance",
			HandlerKey:      HandlerPruneAutoJobRuns,
			IntervalSeconds: 3600,
			Timezone:        "Asia/Shanghai",
			Enabled:         true,
			TimeoutSec:      120,
			ParamsJSON:      `{}`,
		},
		{
			JobCode:         "mark_stuck_auto_jobs",
			Name:            "卡住任务巡检",
			Description:     "定义表 last_status=running 超时标为 timeout",
			Category:        "maintenance",
			HandlerKey:      HandlerMarkStuckAutoJobs,
			IntervalSeconds: 180,
			Timezone:        "Asia/Shanghai",
			Enabled:         true,
			TimeoutSec:      60,
			ParamsJSON:      `{}`,
		},
		{
			JobCode:         "cleanup_expired_idempotency",
			Name:            "幂等键清理",
			Description:     "清理过期接口幂等键",
			Category:        "cleanup",
			HandlerKey:      HandlerCleanupExpiredIdempotency,
			IntervalSeconds: 3600,
			Timezone:        "Asia/Shanghai",
			Enabled:         true,
			TimeoutSec:      120,
			ParamsJSON:      `{}`,
		},
		{
			JobCode:         "cleanup_expired_orders",
			Name:            "过期支付单取消",
			Description:     "取消超时未支付订单",
			Category:        "cleanup",
			HandlerKey:      HandlerCleanupExpiredOrders,
			IntervalSeconds: 120,
			Timezone:        "Asia/Shanghai",
			Enabled:         true,
			TimeoutSec:      120,
			ParamsJSON:      `{}`,
		},
		{
			JobCode:         "reconcile_payment_orders",
			Name:            "支付订单主动对账",
			Description:     "扫描待支付与近期取消/失败订单，向网关查单并补账，异常写入 payment_exceptions",
			Category:        "payment",
			HandlerKey:      HandlerReconcilePaymentOrders,
			IntervalSeconds: 180,
			Timezone:        "Asia/Shanghai",
			Enabled:         true,
			TimeoutSec:      180,
			ParamsJSON:      `{}`,
		},
		{
			JobCode:         "cleanup_sessions_codes",
			Name:            "验证码/会话清理",
			Description:     "软删过期验证码并清理过期会话",
			Category:        "cleanup",
			HandlerKey:      HandlerCleanupSessionsCodes,
			IntervalSeconds: 600,
			Timezone:        "Asia/Shanghai",
			Enabled:         true,
			TimeoutSec:      180,
			ParamsJSON:      `{}`,
		},
	}
}

// ImportPresets 导入默认任务。mode=skip 已存在跳过；update 覆盖元数据（不重置 lifetime/last）
func ImportPresets(mode string) (map[string]interface{}, error) {
	if mode == "" {
		mode = "skip"
	}
	if mode != "skip" && mode != "update" {
		return nil, fmt.Errorf("mode 仅支持 skip/update")
	}
	presets := DefaultPresets()
	now := nowUnix()
	inserted, updated, skipped := 0, 0, 0
	for _, p := range presets {
		existing, err := GetDefinition(p.JobCode)
		if err != nil {
			return nil, err
		}
		enabled := 0
		if p.Enabled {
			enabled = 1
		}
		tz := p.Timezone
		if tz == "" {
			tz = "Asia/Shanghai"
		}
		timeout := p.TimeoutSec
		if timeout <= 0 {
			timeout = 300
		}
		params := p.ParamsJSON
		if params == "" {
			params = "{}"
		}
		if existing == nil {
			def := &JobDefinition{
				JobCode:              p.JobCode,
				Name:                 p.Name,
				Description:          p.Description,
				Category:             p.Category,
				HandlerKey:           p.HandlerKey,
				CronExpr:             p.CronExpr,
				IntervalSeconds:      p.IntervalSeconds,
				Timezone:             tz,
				Enabled:              enabled,
				TimeoutSec:           timeout,
				MaxConcurrency:       1, // 固定单实例；InsertDefinition 写入时也会强制为 1
				ParamsJSON:           params,
				LifetimeRunCount:     "0",
				LifetimeSuccessCount: "0",
				LifetimeFailCount:    "0",
				CreateTime:           now,
				UpdateTime:           now,
			}
			if err := InsertDefinition(def); err != nil {
				return nil, err
			}
			inserted++
			continue
		}
		if mode == "skip" {
			skipped++
			continue
		}
		existing.Name = p.Name
		existing.Description = p.Description
		existing.Category = p.Category
		existing.HandlerKey = p.HandlerKey
		existing.CronExpr = p.CronExpr
		existing.IntervalSeconds = p.IntervalSeconds
		existing.Timezone = tz
		existing.TimeoutSec = timeout
		existing.ParamsJSON = params
		existing.UpdateTime = now
		if err := UpdateDefinitionMeta(existing); err != nil {
			return nil, err
		}
		updated++
	}
	ReloadCache()
	return map[string]interface{}{
		"inserted": inserted,
		"updated":  updated,
		"skipped":  skipped,
		"total":    len(presets),
	}, nil
}

// EnsurePresetsIfEmpty 确保默认任务存在：库空则全量导入；已有库则按 skip 补齐缺失 job_code（如新增对账任务）
func EnsurePresetsIfEmpty() error {
	_, err := ImportPresets("skip")
	return err
}

// softenHighFrequencyIntervals 已有库过密间隔温和上调
func softenHighFrequencyIntervals() {
	for _, b := range []struct {
		code string
		min  int
	}{
		{"mark_stuck_auto_jobs", 180},
		{"cleanup_expired_orders", 120},
		{"reconcile_payment_orders", 180},
	} {
		def, err := GetDefinition(b.code)
		if err != nil || def == nil {
			continue
		}
		if def.IntervalSeconds > 0 && def.IntervalSeconds < b.min {
			def.IntervalSeconds = b.min
			def.UpdateTime = nowUnix()
			_ = UpdateDefinitionMeta(def)
		}
	}
}
