package task

import (
	"fmt"
	"log"

	"fst/backend/pkg/db"
)

// KeepRun 把指定执行记录复制到独立保留表。
// 原记录 keep_forever 仍置 1；保留表会按 source_run_id 去重。
func KeepRun(id uint64) (int64, error) {
	run, err := GetRun(id)
	if err != nil {
		return 0, err
	}
	if run == nil {
		return 0, nil
	}

	// 已保留过同一 source_run_id 则不再重复复制
	var exists int64
	if err := db.DB.Raw(`SELECT COUNT(*) FROM auto_job_runs_keep WHERE source_run_id = ?`, id).Scan(&exists).Error; err != nil {
		return 0, err
	}
	if exists > 0 {
		return 0, nil
	}

	keep := &JobRunKeep{
		RunUID:      run.RunUID,
		JobCode:     run.JobCode,
		Category:    run.Category,
		TriggerType: run.TriggerType,
		Status:      run.Status,
		StartedAt:   run.StartedAt,
		FinishedAt:  run.FinishedAt,
		DurationMs:  run.DurationMs,
		Message:     run.Message,
		DetailJSON:  run.DetailJSON,
		ErrorText:   run.ErrorText,
		Operator:    run.Operator,
		SourceRunID: run.ID,
		KeptAt:      nowUnix(),
	}
	// run_timestamp 用于长期稳定检索，取 started_at；不存在则 finished_at；仍不存在则 kept_at
	if keep.StartedAt > 0 {
		keep.RunTimestamp = keep.StartedAt
	} else if keep.FinishedAt > 0 {
		keep.RunTimestamp = keep.FinishedAt
	} else {
		keep.RunTimestamp = keep.KeptAt
	}

	if err := db.DB.Create(keep).Error; err != nil {
		return 0, err
	}
	return 1, nil
}

// ListKeptRuns 分页查询保留记录。
// keyword 会同时匹配 job_code / run_uid / message / error_text / detail_json。
func ListKeptRuns(page, pageSize int, keyword, status, category, jobCode string, startAt, endAt int64) ([]JobRunKeep, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	where := `WHERE 1=1`
	args := []interface{}{}
	if keyword != "" {
		like := "%" + keyword + "%"
		where += ` AND (job_code LIKE ? OR run_uid LIKE ? OR message LIKE ? OR error_text LIKE ? OR detail_json LIKE ?)`
		args = append(args, like, like, like, like, like)
	}
	if status != "" {
		where += ` AND status=?`
		args = append(args, status)
	}
	if category != "" {
		where += ` AND category=?`
		args = append(args, category)
	}
	if jobCode != "" {
		where += ` AND job_code=?`
		args = append(args, jobCode)
	}
	if startAt > 0 {
		where += ` AND run_timestamp>=?`
		args = append(args, startAt)
	}
	if endAt > 0 {
		where += ` AND run_timestamp<=?`
		args = append(args, endAt)
	}

	var total int64
	if err := db.DB.Raw(`SELECT COUNT(*) FROM auto_job_runs_keep `+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	q := `SELECT * FROM auto_job_runs_keep ` + where + ` ORDER BY kept_at DESC, id DESC LIMIT ? OFFSET ?`
	args2 := append(append([]interface{}{}, args...), pageSize, offset)
	var list []JobRunKeep
	err := db.DB.Raw(q, args2...).Scan(&list).Error
	return list, total, err
}

// CountKeptRuns 保留记录总行数
func CountKeptRuns() (int64, error) {
	var n int64
	err := db.DB.Raw(`SELECT COUNT(*) FROM auto_job_runs_keep`).Scan(&n).Error
	return n, err
}

// DeleteKeptRuns 批量删除保留记录（支持按 scope 与 job_code 筛选）
func DeleteKeptRuns(scope, jobCode string) (int64, error) {
	where := `WHERE 1=1`
	args := []interface{}{}
	switch scope {
	case "success":
		where += ` AND status=?`
		args = append(args, StatusSuccess)
	case "failed":
		where += ` AND status IN (?,?)`
		args = append(args, StatusFailed, StatusTimeout)
	case "all", "":
		// 不限制状态
	default:
		return 0, fmt.Errorf("scope only supports success|failed|all")
	}
	if jobCode != "" {
		where += ` AND job_code=?`
		args = append(args, jobCode)
	}
	r := db.DB.Exec(`DELETE FROM auto_job_runs_keep `+where, args...)
	if r.Error != nil {
		return 0, r.Error
	}
	return r.RowsAffected, nil
}

// CopyKeptByRunID 把若干 run id 复制到保留表，返回实际新增条数。
func CopyKeptByRunID(ids []uint64) (int64, error) {
	var total int64
	for _, id := range ids {
		n, err := KeepRun(id)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// maybeAutoKeep 根据全局配置判断是否需要把该执行记录复制到保留表。
// 命中条件：job_code 在 AutoKeepJobCodes 列表里，或 category 在 AutoKeepCategories 列表里。
func maybeAutoKeep(run *JobRun) {
	if run == nil {
		return
	}
	cfg := LoadGlobalConfig()
	matched := false
	for _, code := range cfg.AutoKeepJobCodes {
		if code == run.JobCode {
			matched = true
			break
		}
	}
	if !matched {
		for _, cat := range cfg.AutoKeepCategories {
			if cat == run.Category {
				matched = true
				break
			}
		}
	}
	if !matched {
		return
	}
	if _, err := KeepRun(run.ID); err != nil {
		log.Printf("[AutoJob] auto keep run id=%d job=%s failed: %v", run.ID, run.JobCode, err)
	}
}
