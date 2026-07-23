package services

import (
	"context"

	"fst/backend/internal/task"
)

// StartBackgroundTasks 启动自动任务（核心在 internal/task）。
// 这里只注入「配置保存后刷新 settings 缓存」回调，再调 task.Start。
func StartBackgroundTasks() {
	task.OnConfigSaved = func() {
		if GlobalSettingsService != nil {
			_ = GlobalSettingsService.RefreshCache()
			ApplyGlobalRuntimeConfig()
		}
	}
	// 注入支付对账批处理，避免 task 包直接依赖 services 造成循环引用
	task.PaymentReconcileBatchFn = func(ctx context.Context, limit int) (map[string]interface{}, error) {
		result, err := ReconcilePaymentOrdersBatch(ctx, limit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"scanned":    result.Scanned,
			"recovered":  result.Recovered,
			"exceptions": result.Exceptions,
			"skipped":    result.Skipped,
			"errors":     result.Errors,
		}, nil
	}
	task.Start()
}
