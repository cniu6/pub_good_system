package task

import (
	"fmt"

	"fst/backend/pkg/db"
)

// stuckLimitSec 卡住判定秒数：取全局阈值与任务自身 timeout 的较大值，且至少 60s
func stuckLimitSec(globalStuckSec, timeoutSec int) int {
	limit := globalStuckSec
	if limit <= 0 {
		limit = 600
	}
	if timeoutSec > limit {
		limit = timeoutSec
	}
	if limit < 60 {
		limit = 60
	}
	return limit
}

// MarkStuckRuns 清理「孤儿」running：定义表仍 running，但本进程已无执行锁，且超过阈值。
// 本进程还在跑（含超时未结束）的绝不误杀，避免 lifetime_fail 重复 +1。
func MarkStuckRuns(globalStuckSec int) (int64, error) {
	now := nowUnix()
	list, err := ListRunningDefinitions()
	if err != nil {
		return 0, err
	}
	var aff int64
	for _, d := range list {
		if IsJobBusy(d.JobCode) {
			continue
		}
		limit := stuckLimitSec(globalStuckSec, d.TimeoutSec)
		if d.LastStartedAt <= 0 || now-d.LastStartedAt < int64(limit) {
			continue
		}
		errMsg := fmt.Sprintf("definition stuck after %ds", limit)
		res, err := db.DB.Exec(`
			UPDATE auto_job_definitions SET
				last_status=?, last_finished_at=?, last_error=?,
				lifetime_fail_count = lifetime_fail_count + 1, update_time=?
			WHERE job_code=? AND last_status=?`,
			StatusTimeout, now, errMsg, now, d.JobCode, StatusRunning,
		)
		if err != nil {
			return aff, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			aff += n
			cacheMu.Lock()
			if cached, ok := cacheDefs[d.JobCode]; ok {
				cached.LastStatus = StatusTimeout
				cached.LastFinishedAt = now
				cached.LastError = errMsg
				cached.UpdateTime = now
				cacheDefs[d.JobCode] = cached
			}
			cacheMu.Unlock()
		}
	}
	return aff, nil
}
