package services

import (
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"fmt"
	"log"
	"sync"
	"time"
)

// CleanupStatus 存储验证码清理任务的运行状态（仅内存）
type CleanupStatus struct {
	mu              sync.RWMutex
	lastCleanupTime time.Time
	intervalMinutes int
	running         bool
}

var cleanupStatus = &CleanupStatus{}
var cleanupStartOnce sync.Once

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

// StartCleanupTask 启动验证码定时清理后台任务（幂等）
// 不输出周期性日志，仅在出错时打印，清理时间记录在内存中
func StartCleanupTask() {
	interval := GetCleanupIntervalMinutes()

	// 使用 sync.Once 保证即使被重复调用，也只启动一个后台 goroutine，
	// 避免重复清理在并发下引发的无谓 DB 压力或日志噪音。
	cleanupStartOnce.Do(func() {
		cleanupStatus.mu.Lock()
		cleanupStatus.intervalMinutes = interval
		cleanupStatus.running = true
		cleanupStatus.mu.Unlock()
		ensureBackgroundTask("cleanup", "验证码/会话清理", time.Duration(interval)*time.Minute)

		go func() {
			// 立即执行一次清理
			if _, err := RunCleanupNow(); err != nil {
				log.Printf("[Cleanup] initial run failed: %v", err)
			}

			for {
				currentInterval := GetCleanupIntervalMinutes()
				if currentInterval <= 0 {
					currentInterval = 10
				}

				timer := time.NewTimer(time.Duration(currentInterval) * time.Minute)
				<-timer.C
				timer.Stop()

				if _, err := RunCleanupNow(); err != nil {
					log.Printf("[Cleanup] periodic run failed: %v", err)
				}
			}
		}()
	})
}

// runCleanup 执行一次清理，只在出错时输出日志
func executeCleanupOnce() (string, error) {
	if err := models.SoftDeleteExpiredCodes(); err != nil {
		log.Printf("[Cleanup] Failed to soft delete expired codes: %v", err)
		return "", err
	}
	if err := models.CleanupOldVerificationCodes(); err != nil {
		log.Printf("[Cleanup] Failed to cleanup old codes: %v", err)
		return "", err
	}
	if err := models.CleanupExpiredSessions(); err != nil {
		log.Printf("[Cleanup] Failed to cleanup user sessions: %v", err)
		return "", err
	}

	cleanupTime := time.Now()
	cleanupStatus.mu.Lock()
	cleanupStatus.lastCleanupTime = cleanupTime
	cleanupStatus.mu.Unlock()

	return fmt.Sprintf("清理完成，最近执行时间：%s", cleanupTime.Format("2006-01-02 15:04:05")), nil
}

// GetCleanupStatus 返回清理任务的当前状态
func GetCleanupStatus() map[string]interface{} {
	cleanupStatus.mu.RLock()
	defer cleanupStatus.mu.RUnlock()

	intervalMinutes := GetCleanupIntervalMinutes()

	result := map[string]interface{}{
		"running":          cleanupStatus.running,
		"interval_minutes": intervalMinutes,
	}

	if !cleanupStatus.lastCleanupTime.IsZero() {
		result["last_cleanup_time"] = cleanupStatus.lastCleanupTime.Format("2006-01-02 15:04:05")
		next := cleanupStatus.lastCleanupTime.Add(time.Duration(intervalMinutes) * time.Minute)
		result["next_cleanup_time"] = next.Format("2006-01-02 15:04:05")
	}

	for _, item := range GetBackgroundTaskStatusList() {
		if item.Key == "cleanup" {
			result["running"] = item.Running
			result["last_status"] = item.LastStatus
			result["last_message"] = item.LastMessage
			result["last_duration_ms"] = item.LastDurationMs
			break
		}
	}

	return result
}
