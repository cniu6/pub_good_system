package task

import (
	"time"
)

// BuildOverview 管理端总览卡片数据
func BuildOverview() (*Overview, error) {
	cfg := LoadGlobalConfig()
	total, enabled, err := CountDefinitions()
	if err != nil {
		return nil, err
	}
	running, _ := CountRunning()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	todayOK, _ := CountRunsByStatusToday(StatusSuccess, dayStart)
	todayFail, _ := CountRunsByStatusToday(StatusFailed, dayStart)
	todayTimeout, _ := CountRunsByStatusToday(StatusTimeout, dayStart)
	lifetime, _ := SumLifetimeRuns()
	rowCount, _ := CountRuns()

	return &Overview{
		EnabledJobs:        enabled,
		TotalJobs:          total,
		RunningCount:       running,
		TodaySuccess:       todayOK,
		TodayFailed:        todayFail + todayTimeout,
		LifetimeRunTotal:   lifetime,
		RunRowCount:        rowCount,
		RunMaxCount:        cfg.RunMaxCount,
		SchedulerRunning:   IsSchedulerRunning(),
		SchedulerUptimeSec: SchedulerUptimeSec(),
		LastTickAt:         LastTickAt(),
		GlobalEnabled:      cfg.Enabled,
	}, nil
}
