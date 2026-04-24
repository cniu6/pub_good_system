package services

import (
	"fmt"
	"fst/backend/app/models"
	"log"
	"sort"
	"sync"
	"time"
)

type BackgroundTaskInfo struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Running        bool   `json:"running"`
	IntervalSecs   int64  `json:"interval_secs"`
	LastRunTime    string `json:"last_run_time"`
	NextRunTime    string `json:"next_run_time"`
	LastStatus     string `json:"last_status"`
	LastMessage    string `json:"last_message"`
	LastDurationMs int64  `json:"last_duration_ms"`
}

type backgroundTaskState struct {
	key            string
	label          string
	interval       time.Duration
	running        bool
	lastRunTime    time.Time
	lastStatus     string
	lastMessage    string
	lastDurationMs int64
	mu             sync.RWMutex
}

var backgroundTasksMu sync.RWMutex
var backgroundTasks = map[string]*backgroundTaskState{}
var orderTaskStartOnce sync.Once

func ensureBackgroundTask(key, label string, interval time.Duration) *backgroundTaskState {
	backgroundTasksMu.Lock()
	defer backgroundTasksMu.Unlock()
	state, ok := backgroundTasks[key]
	if !ok {
		state = &backgroundTaskState{
			key:        key,
			label:      label,
			interval:   interval,
			lastStatus: "idle",
		}
		backgroundTasks[key] = state
		return state
	}
	state.mu.Lock()
	state.label = label
	state.interval = interval
	state.mu.Unlock()
	return state
}

func runTrackedBackgroundTask(key, label string, interval time.Duration, runner func() (string, error)) (message string, err error) {
	state := ensureBackgroundTask(key, label, interval)
	state.mu.Lock()
	if state.running {
		state.mu.Unlock()
		return "任务正在执行中", fmt.Errorf("task already running")
	}
	state.running = true
	state.mu.Unlock()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		state.mu.Lock()
		state.running = false
		state.lastRunTime = start
		state.lastDurationMs = duration
		if err != nil {
			state.lastStatus = "failed"
			state.lastMessage = err.Error()
		} else {
			state.lastStatus = "success"
			state.lastMessage = message
		}
		state.mu.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("task panic: %v", r)
			message = ""
		}
	}()

	message, err = runner()
	return message, err
}

func snapshotBackgroundTask(state *backgroundTaskState) BackgroundTaskInfo {
	state.mu.RLock()
	defer state.mu.RUnlock()

	item := BackgroundTaskInfo{
		Key:            state.key,
		Label:          state.label,
		Running:        state.running,
		IntervalSecs:   int64(state.interval.Seconds()),
		LastStatus:     state.lastStatus,
		LastMessage:    state.lastMessage,
		LastDurationMs: state.lastDurationMs,
	}
	if !state.lastRunTime.IsZero() {
		item.LastRunTime = state.lastRunTime.Format(time.RFC3339)
		if state.interval > 0 {
			item.NextRunTime = state.lastRunTime.Add(state.interval).Format(time.RFC3339)
		}
	}
	return item
}

func GetBackgroundTaskStatusList() []BackgroundTaskInfo {
	backgroundTasksMu.RLock()
	list := make([]*backgroundTaskState, 0, len(backgroundTasks))
	for _, state := range backgroundTasks {
		list = append(list, state)
	}
	backgroundTasksMu.RUnlock()

	items := make([]BackgroundTaskInfo, 0, len(list))
	for _, state := range list {
		items = append(items, snapshotBackgroundTask(state))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
	return items
}

func RunCleanupNow() (string, error) {
	return runTrackedBackgroundTask("cleanup", "验证码/会话清理", time.Duration(GetCleanupIntervalMinutes())*time.Minute, executeCleanupOnce)
}

func StartExpiredOrderTask() {
	interval := time.Minute
	ensureBackgroundTask("order_maintenance", "过期订单取消/幂等键清理", interval)
	orderTaskStartOnce.Do(func() {
		go func() {
			if _, err := RunExpiredOrderMaintenanceNow(); err != nil {
				log.Printf("[Tasks] 首次执行订单维护任务失败: %v", err)
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for range ticker.C {
				if _, err := RunExpiredOrderMaintenanceNow(); err != nil {
					log.Printf("[Tasks] 周期执行订单维护任务失败: %v", err)
				}
			}
		}()
	})
}

func RunExpiredOrderMaintenanceNow() (string, error) {
	return runTrackedBackgroundTask("order_maintenance", "过期订单取消/幂等键清理", time.Minute, func() (string, error) {
		affected, err := CancelExpiredOrders()
		if err != nil {
			return "", err
		}
		idemAffected, idemErr := models.CleanupExpiredIdempotencyKeys()
		if idemErr != nil {
			return "", idemErr
		}
		return fmt.Sprintf("已取消 %d 个过期订单，已清理 %d 个幂等键", affected, idemAffected), nil
	})
}

func RunBackgroundTaskNow(key string) (string, error) {
	switch key {
	case "cleanup":
		return RunCleanupNow()
	case "order_maintenance":
		return RunExpiredOrderMaintenanceNow()
	default:
		return "", fmt.Errorf("unknown task: %s", key)
	}
}
