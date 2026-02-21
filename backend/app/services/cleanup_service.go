package services

import (
	"fst/backend/app/models"
	"fst/backend/internal/config"
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

// StartCleanupTask 启动验证码定时清理后台任务
// 不输出周期性日志，仅在出错时打印，清理时间记录在内存中
func StartCleanupTask() {
	interval := config.GlobalConfig.CleanupIntervalMinutes
	if interval <= 0 {
		interval = 10
	}

	cleanupStatus.mu.Lock()
	cleanupStatus.intervalMinutes = interval
	cleanupStatus.running = true
	cleanupStatus.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()

		// 立即执行一次清理
		runCleanup()

		for range ticker.C {
			runCleanup()
		}
	}()
}

// runCleanup 执行一次清理，只在出错时输出日志
func runCleanup() {
	if err := models.SoftDeleteExpiredCodes(); err != nil {
		log.Printf("[Cleanup] Failed to soft delete expired codes: %v", err)
	}
	if err := models.CleanupOldVerificationCodes(); err != nil {
		log.Printf("[Cleanup] Failed to cleanup old codes: %v", err)
	}

	cleanupStatus.mu.Lock()
	cleanupStatus.lastCleanupTime = time.Now()
	cleanupStatus.mu.Unlock()
}

// GetCleanupStatus 返回清理任务的当前状态
func GetCleanupStatus() map[string]interface{} {
	cleanupStatus.mu.RLock()
	defer cleanupStatus.mu.RUnlock()

	result := map[string]interface{}{
		"running":          cleanupStatus.running,
		"interval_minutes": cleanupStatus.intervalMinutes,
	}

	if !cleanupStatus.lastCleanupTime.IsZero() {
		result["last_cleanup_time"] = cleanupStatus.lastCleanupTime.Format("2006-01-02 15:04:05")
		next := cleanupStatus.lastCleanupTime.Add(time.Duration(cleanupStatus.intervalMinutes) * time.Minute)
		result["next_cleanup_time"] = next.Format("2006-01-02 15:04:05")
	}

	return result
}
