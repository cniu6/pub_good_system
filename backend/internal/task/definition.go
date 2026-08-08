package task

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"fst/backend/pkg/db"
	"fst/backend/utils"

	"gorm.io/gorm"
)

// GetDefinition 按 job_code 取一条定义；不存在返回 (nil, nil)
func GetDefinition(jobCode string) (*JobDefinition, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("db not ready")
	}
	var def JobDefinition
	err := db.FindOne(db.DB.Where("job_code = ?", jobCode), &def)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// ListDefinitions 列表（可选 keyword / category / enabled）
func ListDefinitions(keyword, category string, enabled *bool) ([]JobDefinition, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("db not ready")
	}
	q := `SELECT * FROM auto_job_definitions WHERE 1=1`
	args := []interface{}{}
	if keyword != "" {
		like := "%" + keyword + "%"
		q += ` AND (job_code LIKE ? OR name LIKE ? OR description LIKE ?)`
		args = append(args, like, like, like)
	}
	if category != "" {
		q += ` AND category = ?`
		args = append(args, category)
	}
	if enabled != nil {
		v := 0
		if *enabled {
			v = 1
		}
		q += ` AND enabled = ?`
		args = append(args, v)
	}
	q += ` ORDER BY category ASC, job_code ASC`
	var list []JobDefinition
	err := db.DB.Raw(q, args...).Scan(&list).Error
	return list, err
}

// CountDefinitions 总数与启用数
func CountDefinitions() (total, enabled int64, err error) {
	if db.DB == nil {
		return 0, 0, fmt.Errorf("db not ready")
	}
	if err = db.DB.Raw(`SELECT COUNT(*) FROM auto_job_definitions`).Scan(&total).Error; err != nil {
		return
	}
	err = db.DB.Raw(`SELECT COUNT(*) FROM auto_job_definitions WHERE enabled = 1`).Scan(&enabled).Error
	return
}

func validateJobDefinitionFields(name, cronExpr, timezone, paramsJSON string) error {
	if err := utils.ValidateRuneLen(name, "任务名称", utils.MaxJobNameLength); err != nil {
		return err
	}
	if err := utils.ValidateRuneLen(cronExpr, "Cron表达式", utils.MaxCronExprLength); err != nil {
		return err
	}
	if err := utils.ValidateRuneLen(timezone, "时区", utils.MaxTimezoneLength); err != nil {
		return err
	}
	if err := utils.ValidateRuneLen(paramsJSON, "参数JSON", utils.MaxParamsJSONLength); err != nil {
		return err
	}
	return nil
}

// InsertDefinition 插入定义。MaxConcurrency 强制写 1（当前仅支持单实例互斥）
func InsertDefinition(def *JobDefinition) error {
	if err := validateJobDefinitionFields(def.Name, def.CronExpr, def.Timezone, def.ParamsJSON); err != nil {
		return err
	}
	def.MaxConcurrency = 1
	def.LifetimeRunCount = dec(def.LifetimeRunCount)
	def.LifetimeSuccessCount = dec(def.LifetimeSuccessCount)
	def.LifetimeFailCount = dec(def.LifetimeFailCount)
	return db.DB.Create(def).Error
}

func dec(s string) string {
	if strings.TrimSpace(s) == "" {
		return "0"
	}
	return s
}

// UpdateDefinitionMeta 覆盖元数据（导入 update 模式用）。MaxConcurrency 强制为 1
func UpdateDefinitionMeta(def *JobDefinition) error {
	if err := validateJobDefinitionFields(def.Name, def.CronExpr, def.Timezone, def.ParamsJSON); err != nil {
		return err
	}
	def.MaxConcurrency = 1
	return db.DB.Model(&JobDefinition{}).Where("job_code = ?", def.JobCode).Updates(map[string]interface{}{
		"name":              def.Name,
		"description":       def.Description,
		"category":          def.Category,
		"handler_key":       def.HandlerKey,
		"cron_expr":         def.CronExpr,
		"interval_seconds":  def.IntervalSeconds,
		"timezone":          def.Timezone,
		"timeout_sec":       def.TimeoutSec,
		"max_concurrency":   def.MaxConcurrency,
		"params_json":       def.ParamsJSON,
		"update_time":       def.UpdateTime,
	}).Error
}

// UpdateDefinitionFields 管理端局部更新任务字段
func UpdateDefinitionFields(jobCode string, req UpdateJobRequest) error {
	def, err := GetDefinition(jobCode)
	if err != nil {
		return err
	}
	if def == nil {
		return fmt.Errorf("Task does not exist")
	}
	if req.Name != nil {
		if err := utils.ValidateRuneLen(*req.Name, "任务名称", utils.MaxJobNameLength); err != nil {
			return err
		}
		def.Name = *req.Name
	}
	if req.Description != nil {
		def.Description = *req.Description
	}
	if req.CronExpr != nil {
		if err := utils.ValidateRuneLen(*req.CronExpr, "Cron表达式", utils.MaxCronExprLength); err != nil {
			return err
		}
		def.CronExpr = *req.CronExpr
	}
	if req.IntervalSeconds != nil {
		def.IntervalSeconds = *req.IntervalSeconds
	}
	if req.Timezone != nil {
		if err := utils.ValidateRuneLen(*req.Timezone, "时区", utils.MaxTimezoneLength); err != nil {
			return err
		}
		def.Timezone = *req.Timezone
	}
	if req.Enabled != nil {
		if *req.Enabled {
			def.Enabled = 1
		} else {
			def.Enabled = 0
		}
	}
	if req.TimeoutSec != nil {
		def.TimeoutSec = *req.TimeoutSec
	}
	if req.ParamsJSON != nil {
		if err := utils.ValidateRuneLen(*req.ParamsJSON, "参数JSON", utils.MaxParamsJSONLength); err != nil {
			return err
		}
		def.ParamsJSON = *req.ParamsJSON
	}
	def.UpdateTime = nowUnix()
	err = db.DB.Model(&JobDefinition{}).Where("job_code = ?", jobCode).Updates(map[string]interface{}{
		"name":             def.Name,
		"description":      def.Description,
		"cron_expr":        def.CronExpr,
		"interval_seconds": def.IntervalSeconds,
		"timezone":         def.Timezone,
		"enabled":          def.Enabled,
		"timeout_sec":      def.TimeoutSec,
		"params_json":      def.ParamsJSON,
		"update_time":      def.UpdateTime,
	}).Error
	if err != nil {
		return err
	}
	ReloadCache()
	return nil
}

// SetEnabled 启停任务
func SetEnabled(jobCode string, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	r := db.DB.Model(&JobDefinition{}).Where("job_code = ?", jobCode).Updates(map[string]interface{}{
		"enabled":     v,
		"update_time": nowUnix(),
	})
	if r.Error != nil {
		return r.Error
	}
	n := r.RowsAffected
	if n == 0 {
		return fmt.Errorf("Task does not exist")
	}
	ReloadCache()
	return nil
}

// MarkDefinitionRunning 把定义标为 running（CAS：已是 running 则失败）
func MarkDefinitionRunning(jobCode string, startedAt int64) (bool, error) {
	r := db.DB.Model(&JobDefinition{}).
		Where("job_code = ? AND (last_status IS NULL OR last_status = '' OR last_status <> ?)", jobCode, StatusRunning).
		Updates(map[string]interface{}{
			"last_status":     StatusRunning,
			"last_started_at": startedAt,
			"last_error":      "",
			"update_time":     startedAt,
		})
	if r.Error != nil {
		return false, r.Error
	}
	n := r.RowsAffected
	if n > 0 {
		cacheMu.Lock()
		if d, ok := cacheDefs[jobCode]; ok {
			d.LastStatus = StatusRunning
			d.LastStartedAt = startedAt
			d.LastError = ""
			d.UpdateTime = startedAt
			cacheDefs[jobCode] = d
		}
		cacheMu.Unlock()
	}
	return n > 0, nil
}

// bumpDefinitionAfterRun 执行结束后更新 last_* 与 lifetime 计数，并刷内存缓存
func bumpDefinitionAfterRun(jobCode, status string, startedAt, finished int64, errorText string) error {
	successInc, failInc := 0, 0
	switch status {
	case StatusSuccess:
		successInc = 1
	case StatusFailed, StatusTimeout:
		failInc = 1
	}
	err := db.DB.Model(&JobDefinition{}).Where("job_code = ?", jobCode).Updates(map[string]interface{}{
		"last_status":            status,
		"last_started_at":        startedAt,
		"last_finished_at":       finished,
		"last_error":             errorText,
		"lifetime_run_count":     gorm.Expr("lifetime_run_count + 1"),
		"lifetime_success_count": gorm.Expr("lifetime_success_count + ?", successInc),
		"lifetime_fail_count":    gorm.Expr("lifetime_fail_count + ?", failInc),
		"update_time":            finished,
	}).Error
	if err == nil {
		cacheMu.Lock()
		if d, ok := cacheDefs[jobCode]; ok {
			d.LastStatus = status
			d.LastStartedAt = startedAt
			d.LastFinishedAt = finished
			d.LastError = errorText
			d.UpdateTime = finished
			cacheDefs[jobCode] = d
		}
		cacheMu.Unlock()
	}
	return err
}

// CountRunning 当前运行中任务数（看定义表 last_status；runs 跑完才落库）
func CountRunning() (int64, error) {
	var n int64
	err := db.DB.Raw(`SELECT COUNT(*) FROM auto_job_definitions WHERE last_status=?`, StatusRunning).Scan(&n).Error
	return n, err
}

// ListRunningDefinitions 当前运行中的任务（定义表视角）
func ListRunningDefinitions() ([]JobDefinition, error) {
	var list []JobDefinition
	err := db.DB.Raw(`
		SELECT * FROM auto_job_definitions
		WHERE last_status=?
		ORDER BY last_started_at DESC, job_code ASC`, StatusRunning).Scan(&list).Error
	return list, err
}

// SumLifetimeRuns 所有任务终身执行次数合计
func SumLifetimeRuns() (string, error) {
	var s sql.NullString
	err := db.DB.Raw(`SELECT COALESCE(`+db.CastToText("SUM(lifetime_run_count)")+`, '0') FROM auto_job_definitions`).Scan(&s).Error
	if err != nil {
		return "0", err
	}
	if !s.Valid || s.String == "" {
		return "0", nil
	}
	return s.String, nil
}
