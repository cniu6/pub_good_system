package task

import (
	"fmt"
	"log"

	"fst/backend/pkg/db"

	"gorm.io/gorm"
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
	if err := db.DB.Model(&JobRunKeep{}).Where("source_run_id = ?", id).Count(&exists).Error; err != nil {
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

	q := db.DB.Model(&JobRunKeep{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("job_code LIKE ? OR run_uid LIKE ? OR message LIKE ? OR error_text LIKE ? OR detail_json LIKE ?", like, like, like, like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if jobCode != "" {
		q = q.Where("job_code = ?", jobCode)
	}
	if startAt > 0 {
		q = q.Where("run_timestamp >= ?", startAt)
	}
	if endAt > 0 {
		q = q.Where("run_timestamp <= ?", endAt)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []JobRunKeep
	err := q.Order("kept_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	return list, total, err
}

// DeleteKeptRuns 批量删除保留记录（支持按 scope 与 job_code 筛选）
func DeleteKeptRuns(scope, jobCode string) (int64, error) {
	// scope=all 且不限 job_code 时是有意的全表删除，需显式允许（GORM 默认拦截无条件 DELETE）。
	q := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true})
	switch scope {
	case "success":
		q = q.Where("status = ?", StatusSuccess)
	case "failed":
		q = q.Where("status IN ?", []string{StatusFailed, StatusTimeout})
	case "all", "":
		// 不限制状态
	default:
		return 0, fmt.Errorf("scope only supports success|failed|all")
	}
	if jobCode != "" {
		q = q.Where("job_code = ?", jobCode)
	}
	r := q.Delete(&JobRunKeep{})
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
