package controllers

import (
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemController struct{}

// getCleanupIntervalMinutes 读取清理间隔（分钟）
func getCleanupIntervalMinutes() int {
	cfg := config.GlobalConfig
	if cfg == nil {
		return 10
	}
	interval := cfg.CleanupIntervalMinutes
	if interval <= 0 {
		return 10
	}
	return interval
}

// GetCleanupStatus 查询验证码清理任务的运行状态
// @Summary 获取清理任务状态
// @Description 返回验证码清理任务的运行状态、间隔、上次/下次执行时间
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/system/cleanup-status [get]
func (ctrl *SystemController) GetCleanupStatus(c *gin.Context) {
	intervalMinutes := getCleanupIntervalMinutes()
	result := map[string]interface{}{
		"running":          false,
		"interval_minutes": intervalMinutes,
	}

	// 正式调度与执行记录由 internal/task（cleanup_sessions_codes）负责
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

	utils.Success(c, result)
}
