package task

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"fst/backend/pkg/db"
)

// InsertRun 写入一条执行记录（与 id 重编号互斥，避免换表瞬间写丢）
func InsertRun(run *JobRun) (uint64, error) {
	renumberMu.Lock()
	defer renumberMu.Unlock()

	res, err := db.Exec(`
		INSERT INTO auto_job_runs (
			run_uid, job_code, category, trigger_type, status, started_at, finished_at, duration_ms,
			message, detail_json, error_text, keep_forever, operator
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		run.RunUID, run.JobCode, run.Category, run.TriggerType, run.Status, run.StartedAt, run.FinishedAt, run.DurationMs,
		run.Message, run.DetailJSON, run.ErrorText, run.KeepForever, run.Operator,
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

// GetRun 按数字 id 取执行记录
func GetRun(id uint64) (*JobRun, error) {
	var run JobRun
	err := db.DB.Get(&run, `SELECT * FROM auto_job_runs WHERE id=?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// ListRuns 分页查询执行记录
func ListRuns(page, pageSize int, keyword, status, category, jobCode string, startAt, endAt int64) ([]JobRun, int64, error) {
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
		where += ` AND (job_code LIKE ? OR message LIKE ? OR error_text LIKE ? OR run_uid LIKE ?)`
		args = append(args, like, like, like, like)
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
		where += ` AND started_at>=?`
		args = append(args, startAt)
	}
	if endAt > 0 {
		where += ` AND started_at<=?`
		args = append(args, endAt)
	}
	var total int64
	if err := db.DB.Get(&total, `SELECT COUNT(*) FROM auto_job_runs `+where, args...); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := `SELECT * FROM auto_job_runs ` + where + ` ORDER BY started_at DESC, id DESC LIMIT ? OFFSET ?`
	args2 := append(append([]interface{}{}, args...), pageSize, offset)
	var list []JobRun
	err := db.DB.Select(&list, q, args2...)
	return list, total, err
}

// CountRuns 执行记录总行数
func CountRuns() (int64, error) {
	var n int64
	err := db.DB.Get(&n, `SELECT COUNT(*) FROM auto_job_runs`)
	return n, err
}

// CountRunsByStatusToday 今日某状态记录数（started_at >= dayStart）
func CountRunsByStatusToday(status string, dayStart int64) (int64, error) {
	var n int64
	err := db.DB.Get(&n, `SELECT COUNT(*) FROM auto_job_runs WHERE status=? AND started_at>=?`, status, dayStart)
	return n, err
}

// PruneSuccessRuns 按上限修剪：先删旧成功；仍超限且未「保留错误」再删旧 failed/timeout
func PruneSuccessRuns(maxCount int) (int64, error) {
	if maxCount <= 0 {
		maxCount = 10000
	}
	n, err := CountRuns()
	if err != nil {
		return 0, err
	}
	if n <= int64(maxCount) {
		return 0, nil
	}
	need := n - int64(maxCount)
	var totalAff int64

	aff, err := deleteOldestRuns(need, []string{StatusSuccess})
	if err != nil {
		return 0, err
	}
	totalAff += aff
	need -= aff
	if need <= 0 {
		return totalAff, nil
	}

	cfg := LoadGlobalConfig()
	if !cfg.RetainErrors {
		aff2, err := deleteOldestRuns(need, []string{StatusFailed, StatusTimeout})
		if err != nil {
			return totalAff, err
		}
		totalAff += aff2
	}
	return totalAff, nil
}

func deleteOldestRuns(limit int64, statuses []string) (int64, error) {
	if limit <= 0 || len(statuses) == 0 {
		return 0, nil
	}
	ph := make([]string, len(statuses))
	args := make([]interface{}, 0, len(statuses)+1)
	for i, s := range statuses {
		ph[i] = "?"
		args = append(args, s)
	}
	args = append(args, limit)
	res, err := db.Exec(`
		DELETE FROM auto_job_runs
		WHERE id IN (
			SELECT id FROM (
				SELECT id FROM auto_job_runs
				WHERE keep_forever=0 AND status IN (`+strings.Join(ph, ",")+`)
				ORDER BY started_at ASC, id ASC
				LIMIT ?
			) t
		)`, args...)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	return aff, nil
}

// runIDNearLimit 自增接近 Go int64 / BIGINT 可用上限前的安全水位（留 100 万余量）
const runIDNearLimit uint64 = uint64(1<<63 - 1 - 1_000_000) // MaxInt64 - 1e6

var renumberMu sync.Mutex

// MaybeRenumberRunIDsIfNearLimit 检测 AUTO_INCREMENT / MAX(id) 是否逼近上限。
// 若逼近：把当前表内记录按时间重写入新表，id 变为 1..N，自增从 N+1 继续。
//
// 注意：不能「保留巨大旧 id 的同时让新行从 1 起」——InnoDB 要求自增 ≥ MAX(id)+1。
// 正常路径跑完才 INSERT runs，故无需检查 runs.status=running。
func MaybeRenumberRunIDsIfNearLimit() (did bool, newCount int64, err error) {
	// SQLite 无 CREATE TABLE LIKE / RENAME TABLE 多表语法；本地临时库跳过重编号（id 也远不到上限）
	if db.IsSQLite() {
		return false, 0, nil
	}
	ai, maxID, err := runsIDWatermark()
	if err != nil {
		return false, 0, err
	}
	if ai < runIDNearLimit && maxID < runIDNearLimit {
		return false, 0, nil
	}

	renumberMu.Lock()
	defer renumberMu.Unlock()

	// 锁内再看一次，避免并发重复重建
	ai, maxID, err = runsIDWatermark()
	if err != nil {
		return false, 0, err
	}
	if ai < runIDNearLimit && maxID < runIDNearLimit {
		return false, 0, nil
	}

	const tmp = "auto_job_runs__renum"
	const old = "auto_job_runs__old"

	_, _ = db.Exec(`DROP TABLE IF EXISTS ` + tmp)
	_, _ = db.Exec(`DROP TABLE IF EXISTS ` + old)

	if _, err := db.Exec(`CREATE TABLE ` + tmp + ` LIKE auto_job_runs`); err != nil {
		return false, 0, fmt.Errorf("创建重编号临时表失败: %w", err)
	}
	res, err := db.Exec(`
		INSERT INTO ` + tmp + ` (
			run_uid, job_code, category, trigger_type, status, started_at, finished_at, duration_ms,
			message, detail_json, error_text, keep_forever, operator
		)
		SELECT
			run_uid, job_code, category, trigger_type, status, started_at, finished_at, duration_ms,
			message, detail_json, error_text, keep_forever, operator
		FROM auto_job_runs
		ORDER BY started_at ASC, id ASC`)
	if err != nil {
		_, _ = db.Exec(`DROP TABLE IF EXISTS ` + tmp)
		return false, 0, fmt.Errorf("重编号拷贝失败: %w", err)
	}
	copied, _ := res.RowsAffected()

	if _, err := db.Exec(`RENAME TABLE auto_job_runs TO ` + old + `, ` + tmp + ` TO auto_job_runs`); err != nil {
		_, _ = db.Exec(`DROP TABLE IF EXISTS ` + tmp)
		return false, 0, fmt.Errorf("重编号换表失败: %w", err)
	}
	_, _ = db.Exec(`DROP TABLE IF EXISTS ` + old)

	log.Printf("[AutoJob] 执行记录 id 已重编号：保留 %d 条，新自增从 %d 起（原水位 ai=%d max_id=%d）",
		copied, copied+1, ai, maxID)
	return true, copied, nil
}

func runsIDWatermark() (autoInc, maxID uint64, err error) {
	if db.DB == nil {
		return 0, 0, fmt.Errorf("db not ready")
	}
	var ai sql.NullInt64
	err = db.DB.Get(&ai, `
		SELECT AUTO_INCREMENT FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'auto_job_runs'`)
	if err != nil {
		return 0, 0, err
	}
	if ai.Valid && ai.Int64 > 0 {
		autoInc = uint64(ai.Int64)
	}
	var mx sql.NullInt64
	_ = db.DB.Get(&mx, `SELECT MAX(id) FROM auto_job_runs`)
	if mx.Valid && mx.Int64 > 0 {
		maxID = uint64(mx.Int64)
	}
	return autoInc, maxID, nil
}

// CleanRuns scope = success|failed|all（只删 keep_forever=0 的终态记录）
func CleanRuns(req CleanRunsRequest) (int64, error) {
	where := `WHERE keep_forever=0`
	args := []interface{}{}
	switch req.Scope {
	case "success", "":
		where += ` AND status=?`
		args = append(args, StatusSuccess)
	case "failed":
		where += ` AND status IN (?,?)`
		args = append(args, StatusFailed, StatusTimeout)
	case "all":
		// 终态：success / failed / timeout（runs 不会有 running）
		where += ` AND status IN (?,?,?)`
		args = append(args, StatusSuccess, StatusFailed, StatusTimeout)
	default:
		return 0, fmt.Errorf("scope 仅支持 success|failed|all")
	}
	if req.JobCode != "" {
		where += ` AND job_code=?`
		args = append(args, req.JobCode)
	}
	res, err := db.Exec(`DELETE FROM auto_job_runs `+where, args...)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	return aff, nil
}

// MarkKeepForever 批量标记/取消永久保留
func MarkKeepForever(ids []uint64, keep bool) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	v := 0
	if keep {
		v = 1
	}
	ph := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, v)
	for i, id := range ids {
		ph[i] = "?"
		args = append(args, id)
	}
	res, err := db.Exec(`UPDATE auto_job_runs SET keep_forever=? WHERE id IN (`+strings.Join(ph, ",")+`)`, args...)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	return aff, nil
}

// RepairBadRunUIDs 修复长度不是 36 的旧 run_uid（曾用 jobCode+nano 超长被截断）
func RepairBadRunUIDs() (int64, error) {
	var ids []uint64
	// db.Q：SQLite 下 CHAR_LENGTH → LENGTH；Select 不会自动适配，必须显式包一层
	if err := db.DB.Select(&ids, db.Q(`
		SELECT id FROM auto_job_runs
		WHERE run_uid = '' OR CHAR_LENGTH(run_uid) <> 36
		ORDER BY id ASC
		LIMIT 5000`)); err != nil {
		return 0, err
	}
	var aff int64
	for _, id := range ids {
		res, err := db.Exec(`UPDATE auto_job_runs SET run_uid=? WHERE id=?`, newRunUID(), id)
		if err != nil {
			return aff, err
		}
		n, _ := res.RowsAffected()
		aff += n
	}
	return aff, nil
}
