package services

import "fst/backend/internal/task"

// StartBackgroundTasks 启动自动任务（核心在 internal/task）。
// 这里只注入「配置保存后刷新 settings 缓存」回调，再调 task.Start。
func StartBackgroundTasks() {
	task.OnConfigSaved = func() {
		if GlobalSettingsService != nil {
			_ = GlobalSettingsService.RefreshCache()
			ApplyGlobalRuntimeConfig()
		}
	}
	task.Start()
}
