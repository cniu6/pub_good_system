package task

import (
	"strconv"
	"strings"

	"fst/backend/app/models"
)

// loadGlobalConfigFromDB 从 system_settings 读全局配置（无缓存时用）
func loadGlobalConfigFromDB() GlobalConfig {
	cfg := GlobalConfig{
		Enabled:      true,
		RunMaxCount:  10000,
		RetainErrors: true,
		AutoPrune:    true,
	}
	m, err := models.GetSettingsMap([]string{
		CfgEnabled, CfgRunMaxCount, CfgRetainErrors, CfgAutoPrune, CfgStuckAfterSec,
	})
	if err != nil {
		return cfg
	}
	if v, ok := m[CfgEnabled]; ok {
		cfg.Enabled = v == "true" || v == "1"
	}
	if v, ok := m[CfgRunMaxCount]; ok {
		if n, e := strconv.Atoi(strings.TrimSpace(v)); e == nil && n > 0 {
			cfg.RunMaxCount = n
		}
	}
	if v, ok := m[CfgRetainErrors]; ok {
		cfg.RetainErrors = v == "true" || v == "1"
	}
	if v, ok := m[CfgAutoPrune]; ok {
		cfg.AutoPrune = v == "true" || v == "1"
	}
	if v, ok := m[CfgStuckAfterSec]; ok {
		if n, e := strconv.Atoi(strings.TrimSpace(v)); e == nil {
			cfg.StuckAfterSec = n
		}
	}
	if cfg.StuckAfterSec <= 0 {
		cfg.StuckAfterSec = 600
	}
	if cfg.StuckAfterSec < 60 {
		cfg.StuckAfterSec = 60
	}
	return cfg
}

// SaveGlobalConfig 写回 system_settings，并刷新内存缓存
func SaveGlobalConfig(cfg GlobalConfig) error {
	if cfg.RunMaxCount <= 0 {
		cfg.RunMaxCount = 10000
	}
	// 卡住阈值：至少 60s；0 表示用默认 600，避免误配成过小导致误杀
	if cfg.StuckAfterSec <= 0 {
		cfg.StuckAfterSec = 600
	}
	if cfg.StuckAfterSec < 60 {
		cfg.StuckAfterSec = 60
	}
	vals := map[string]string{
		CfgEnabled:       boolStr(cfg.Enabled),
		CfgRunMaxCount:   strconv.Itoa(cfg.RunMaxCount),
		CfgRetainErrors:  boolStr(cfg.RetainErrors),
		CfgAutoPrune:     boolStr(cfg.AutoPrune),
		CfgStuckAfterSec: strconv.Itoa(cfg.StuckAfterSec),
	}
	if err := models.BatchUpdateSettings(vals); err != nil {
		return err
	}
	setCachedGlobalConfig(cfg)
	if OnConfigSaved != nil {
		OnConfigSaved()
	}
	return nil
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
