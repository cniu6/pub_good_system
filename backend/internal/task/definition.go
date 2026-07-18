package task

import (
	"database/sql"
	"fmt"
	"strings"

	"fst/backend/pkg/db"
)

// GetDefinition 按 job_code 取一条定义；不存在返回 (nil, nil)
func GetDefinition(jobCode string) (*JobDefinition, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("db not ready")
	}
	var def JobDefinition
	err := db.DB.Get(&def, `SELECT * FROM auto_job_definitions WHERE job_code = ?`, jobCode)
	if err == sql.ErrNoRows {
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
	err := db.DB.Select(&list, q, args...)
	return list, err
}

// CountDefinitions 总数与启用数
func CountDefinitions() (total, enabled int64, err error) {
	if db.DB == nil {
		return 0, 0, fmt.Errorf("db not ready")
	}
	if err = db.DB.Get(&total, `SELECT COUNT(*) FROM auto_job_definitions`); err != nil {
		return
	}
	err = db.DB.Get(&enabled, `SELECT COUNT(*) FROM auto_job_definitions WHERE enabled = 1`)
	return
}

// InsertDefinition 插入定义。MaxConcurrency 强制写 1（当前仅支持单实例互斥）
func InsertDefinition(def *JobDefinition) error {
	def.MaxConcurrency = 1
	_, err := db.DB.Exec(`
		INSERT INTO auto_job_definitions (
			job_code, name, description, category, handler_key, cron_expr, interval_seconds, timezone,
			enabled, timeout_sec, max_concurrency, params_json,
			last_status, last_started_at, last_finished_at, last_error,
			lifetime_run_count, lifetime_success_count, lifetime_fail_count,
			create_time, update_time
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		def.JobCode, def.Name, def.Description, def.Category, def.HandlerKey, def.CronExpr, def.IntervalSeconds, def.Timezone,
		def.Enabled, def.TimeoutSec, def.MaxConcurrency, def.ParamsJSON,
		def.LastStatus, def.LastStartedAt, def.LastFinishedAt, def.LastError,
		dec(def.LifetimeRunCount), dec(def.LifetimeSuccessCount), dec(def.LifetimeFailCount),
		def.CreateTime, def.UpdateTime,
	)
	return err
}

func dec(s string) string {
	if strings.TrimSpace(s) == "" {
		return "0"
	}
	return s
}

// UpdateDefinitionMeta 覆盖元数据（导入 update 模式用）。MaxConcurrency 强制为 1
func UpdateDefinitionMeta(def *JobDefinition) error {
	def.MaxConcurrency = 1
	_, err := db.DB.Exec(`
		UPDATE auto_job_definitions SET
			name=?, description=?, category=?, handler_key=?, cron_expr=?, interval_seconds=?, timezone=?,
			timeout_sec=?, max_concurrency=?, params_json=?, update_time=?
		WHERE job_code=?`,
		def.Name, def.Description, def.Category, def.HandlerKey, def.CronExpr, def.IntervalSeconds, def.Timezone,
		def.TimeoutSec, def.MaxConcurrency, def.ParamsJSON, def.UpdateTime, def.JobCode,
	)
	return err
}

// UpdateDefinitionFields 管理端局部更新任务字段
func UpdateDefinitionFields(jobCode string, req UpdateJobRequest) error {
	def, err := GetDefinition(jobCode)
	if err != nil {
		return err
	}
	if def == nil {
		return fmt.Errorf("任务不存在")
	}
	if req.Name != nil {
		def.Name = *req.Name
	}
	if req.Description != nil {
		def.Description = *req.Description
	}
	if req.CronExpr != nil {
		def.CronExpr = *req.CronExpr
	}
	if req.IntervalSeconds != nil {
		def.IntervalSeconds = *req.IntervalSeconds
	}
	if req.Timezone != nil {
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
		def.ParamsJSON = *req.ParamsJSON
	}
	def.UpdateTime = nowUnix()
	_, err = db.DB.Exec(`
		UPDATE auto_job_definitions SET
			name=?, description=?, cron_expr=?, interval_seconds=?, timezone=?,
			enabled=?, timeout_sec=?, params_json=?, update_time=?
		WHERE job_code=?`,
		def.Name, def.Description, def.CronExpr, def.IntervalSeconds, def.Timezone,
		def.Enabled, def.TimeoutSec, def.ParamsJSON, def.UpdateTime, jobCode,
	)
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
	res, err := db.DB.Exec(`UPDATE auto_job_definitions SET enabled=?, update_time=? WHERE job_code=?`, v, nowUnix(), jobCode)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("任务不存在")
	}
	ReloadCache()
	return nil
}

// MarkDefinitionRunning 把定义标为 running（CAS：已是 running 则失败）
func MarkDefinitionRunning(jobCode string, startedAt int64) (bool, error) {
	res, err := db.DB.Exec(`
		UPDATE auto_job_definitions
		SET last_status=?, last_started_at=?, last_error='', update_time=?
		WHERE job_code=? AND (last_status IS NULL OR last_status = '' OR last_status <> ?)`,
		StatusRunning, startedAt, startedAt, jobCode, StatusRunning,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
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
	_, err := db.DB.Exec(`
		UPDATE auto_job_definitions SET
			last_status=?, last_started_at=?, last_finished_at=?, last_error=?,
			lifetime_run_count = lifetime_run_count + 1,
			lifetime_success_count = lifetime_success_count + ?,
			lifetime_fail_count = lifetime_fail_count + ?,
			update_time=?
		WHERE job_code=?`,
		status, startedAt, finished, errorText, successInc, failInc, finished, jobCode,
	)
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
	err := db.DB.Get(&n, `SELECT COUNT(*) FROM auto_job_definitions WHERE last_status=?`, StatusRunning)
	return n, err
}

// ListRunningDefinitions 当前运行中的任务（定义表视角）
func ListRunningDefinitions() ([]JobDefinition, error) {
	var list []JobDefinition
	err := db.DB.Select(&list, `
		SELECT * FROM auto_job_definitions
		WHERE last_status=?
		ORDER BY last_started_at DESC, job_code ASC`, StatusRunning)
	return list, err
}

// SumLifetimeRuns 所有任务终身执行次数合计
func SumLifetimeRuns() (string, error) {
	var s sql.NullString
	err := db.DB.Get(&s, `SELECT COALESCE(CAST(SUM(lifetime_run_count) AS CHAR), '0') FROM auto_job_definitions`)
	if err != nil {
		return "0", err
	}
	if !s.Valid || s.String == "" {
		return "0", nil
	}
	return s.String, nil
}
