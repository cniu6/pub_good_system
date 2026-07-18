package services

import (
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"time"
)

// GetCleanupIntervalMinutes 读取清理间隔（分钟），供兼容接口与配置展示。
func GetCleanupIntervalMinutes() int {
	cfg := config.GlobalConfig
	if cfg == nil {
		return 10
	}
	interval := cfg.CleanupIntervalMinutes
	if interval <= 0 {
		interval = 10
	}
	return interval
}

// GetCleanupStatus 返回验证码/会话清理任务状态（系统接口用）。
// 正式调度与执行记录由 internal/task（cleanup_sessions_codes）负责。
func GetCleanupStatus() map[string]interface{} {
	intervalMinutes := GetCleanupIntervalMinutes()
	result := map[string]interface{}{
		"running":          false,
		"interval_minutes": intervalMinutes,
	}

	def, err := task.GetDefinition("cleanup_sessions_codes")
	if err == nil && def != nil {
		result["running"] = def.LastStatus == task.StatusRunning
		result["last_status"] = def.LastStatus
		result["last_message"] = def.LastError
		if def.LastFinishedAt > 0 {
			t := time.Unix(def.LastFinishedAt, 0)
			result["last_cleanup_time"] = t.Format("2006-01-02 15:04:05")
			next := t.Add(time.Duration(def.IntervalSeconds) * time.Second)
			if def.IntervalSeconds <= 0 {
				next = t.Add(time.Duration(intervalMinutes) * time.Minute)
			}
			result["next_cleanup_time"] = next.Format("2006-01-02 15:04:05")
		}
		if def.IntervalSeconds > 0 {
			result["interval_minutes"] = def.IntervalSeconds / 60
		}
	}

	return result
}
