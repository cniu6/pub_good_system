package task

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OnConfigSaved 配置写入后回调（由 services 注入刷新 settings 缓存）
var OnConfigSaved func()

var (
	schedOnce    sync.Once
	schedRunning bool
	schedStartAt time.Time
	lastTickAt   int64

	cacheMu    sync.RWMutex
	cacheDefs  = map[string]JobDefinition{}
	cacheCfg   GlobalConfig
	cacheReady bool

	memLocks   sync.Map // job_code -> *sync.Mutex
	lastFireMu sync.Mutex
	lastFire   = map[string]int64{}
)

// Start 启动进程内调度（幂等）
func Start() {
	schedOnce.Do(func() {
		if err := EnsurePresetsIfEmpty(); err != nil {
			log.Printf("[AutoJob] 导入默认任务失败: %v", err)
		}
		softenHighFrequencyIntervals()
		ReloadCache()
		schedRunning = true
		schedStartAt = time.Now()
		log.Printf("[AutoJob] scheduler started")
		go loop()
	})
}

func IsSchedulerRunning() bool { return schedRunning }

func SchedulerUptimeSec() int64 {
	if !schedRunning || schedStartAt.IsZero() {
		return 0
	}
	return int64(time.Since(schedStartAt).Seconds())
}

func LastTickAt() int64 { return lastTickAt }

// ReloadCache 从库加载定义 + 配置到内存
func ReloadCache() {
	cfg := loadGlobalConfigFromDB()
	list, err := ListDefinitions("", "", nil)
	if err != nil {
		log.Printf("[AutoJob] ReloadCache 失败: %v", err)
		return
	}
	m := make(map[string]JobDefinition, len(list))
	for i := range list {
		m[list[i].JobCode] = list[i]
	}
	cacheMu.Lock()
	cacheCfg = cfg
	cacheDefs = m
	cacheReady = true
	cacheMu.Unlock()
	log.Printf("[AutoJob] 缓存已加载：任务 %d，开关=%v", len(m), cfg.Enabled)
}

// LoadGlobalConfig 优先读内存
func LoadGlobalConfig() GlobalConfig {
	cacheMu.RLock()
	ready, cfg := cacheReady, cacheCfg
	cacheMu.RUnlock()
	if ready {
		return cfg
	}
	return loadGlobalConfigFromDB()
}

func setCachedGlobalConfig(cfg GlobalConfig) {
	cacheMu.Lock()
	cacheCfg = cfg
	cacheReady = true
	cacheMu.Unlock()
}

func cachedEnabledJobs() []JobDefinition {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	out := make([]JobDefinition, 0)
	for _, d := range cacheDefs {
		if d.Enabled == 1 {
			out = append(out, d)
		}
	}
	return out
}

func loop() {
	time.Sleep(2 * time.Second)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	tickOnce()
	for range ticker.C {
		tickOnce()
	}
}

func tickOnce() {
	lastTickAt = nowUnix()
	cfg := LoadGlobalConfig()
	if !cfg.Enabled {
		return
	}
	now := time.Now()
	for _, job := range cachedEnabledJobs() {
		j := job
		if !shouldFire(&j, now) {
			continue
		}
		go func(code string) {
			// 调度派发层兜底：防止 Trigger 内部未捕获的异常拖垮进程
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[AutoJob] schedule %s panic: %v", code, r)
				}
			}()
			if _, err := Trigger(code, RunOptions{Trigger: TriggerSchedule}); err != nil {
				if !strings.Contains(err.Error(), "正在执行") && !strings.Contains(err.Error(), "已关闭") {
					log.Printf("[AutoJob] schedule %s: %v", code, err)
				}
			}
		}(j.JobCode)
	}
}

func shouldFire(job *JobDefinition, now time.Time) bool {
	// 缓存里已是 running：不要每 5 秒再 Trigger 一次（只会打「正在执行中」）
	if job.LastStatus == StatusRunning {
		return false
	}
	loc := loadTZ(job.Timezone)
	local := now.In(loc)

	if job.IntervalSeconds > 0 {
		last := job.LastFinishedAt
		if last <= 0 {
			last = job.LastStartedAt
		}
		if last <= 0 || local.Unix()-last >= int64(job.IntervalSeconds) {
			return markFire(job.JobCode, local.Unix())
		}
		return false
	}
	if strings.TrimSpace(job.CronExpr) == "" {
		return false
	}
	if !matchCron(job.CronExpr, local) {
		return false
	}
	bucket := local.Unix() / 60
	lastFireMu.Lock()
	defer lastFireMu.Unlock()
	if lastFire[job.JobCode] == bucket {
		return false
	}
	lastFire[job.JobCode] = bucket
	return true
}

func markFire(jobCode string, ts int64) bool {
	lastFireMu.Lock()
	defer lastFireMu.Unlock()
	if prev, ok := lastFire[jobCode]; ok && ts-prev < 5 {
		return false
	}
	lastFire[jobCode] = ts
	return true
}

func loadTZ(name string) *time.Location {
	if name == "" {
		name = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

// matchCron 5 段：分 时 日 月 周（* / 数字 / 列表 / 步长）
func matchCron(expr string, t time.Time) bool {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return false
	}
	return matchField(parts[0], t.Minute(), 0, 59) &&
		matchField(parts[1], t.Hour(), 0, 23) &&
		matchField(parts[2], t.Day(), 1, 31) &&
		matchField(parts[3], int(t.Month()), 1, 12) &&
		matchField(parts[4], int(t.Weekday()), 0, 6)
}

func matchField(field string, value, min, max int) bool {
	field = strings.TrimSpace(field)
	if field == "*" {
		return true
	}
	if strings.HasPrefix(field, "*/") {
		n, err := strconv.Atoi(field[2:])
		return err == nil && n > 0 && (value-min)%n == 0
	}
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			ab := strings.SplitN(part, "-", 2)
			if len(ab) != 2 {
				continue
			}
			a, e1 := strconv.Atoi(ab[0])
			b, e2 := strconv.Atoi(ab[1])
			if e1 == nil && e2 == nil && value >= a && value <= b {
				return true
			}
			continue
		}
		n, err := strconv.Atoi(part)
		if err == nil && n == value {
			return true
		}
	}
	return false
}

func jobMutex(jobCode string) *sync.Mutex {
	v, _ := memLocks.LoadOrStore(jobCode, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// IsJobBusy 本进程是否仍持有该任务执行锁（含「已判超时但 handler 还在跑」）
func IsJobBusy(jobCode string) bool {
	mu := jobMutex(jobCode)
	if !mu.TryLock() {
		return true
	}
	mu.Unlock()
	return false
}

// Trigger 执行一次任务
func Trigger(jobCode string, opts RunOptions) (*JobRun, error) {
	if opts.Trigger == "" {
		opts.Trigger = TriggerManual
	}
	cfg := LoadGlobalConfig()
	if !opts.Force && opts.Trigger == TriggerSchedule && !cfg.Enabled {
		return nil, fmt.Errorf("全局自动任务已关闭")
	}

	def, err := GetDefinition(jobCode)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, fmt.Errorf("任务不存在: %s", jobCode)
	}
	if opts.Trigger == TriggerSchedule && def.Enabled != 1 {
		return nil, fmt.Errorf("任务未启用")
	}

	mu := jobMutex(jobCode)
	if !mu.TryLock() {
		return nil, fmt.Errorf("任务正在执行中")
	}
	// 正常路径由 defer 解锁；超时路径把锁移交给后台（等 handler 真跑完再 Unlock）
	unlockByDefer := true
	defer func() {
		if unlockByDefer {
			mu.Unlock()
		}
	}()

	startedAt := nowUnix()
	ok, err := MarkDefinitionRunning(jobCode, startedAt)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("任务正在执行中")
	}

	handler, ok := GetHandler(def.HandlerKey)
	if !ok {
		run, _ := persistFailed(def, opts, startedAt, "handler 未注册", "handler not found: "+def.HandlerKey)
		maybePrune(cfg)
		return run, fmt.Errorf("handler 未注册: %s", def.HandlerKey)
	}

	timeout := def.TimeoutSec
	if timeout <= 0 {
		timeout = 300
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	type outT struct {
		res *HandlerResult
		err error
	}
	ch := make(chan outT, 1)
	go func() {
		var res *HandlerResult
		var err error
		// handler panic 转成错误返回，避免拖垮整个进程
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("handler panic: %v", r)
				res = nil
			}
			ch <- outT{res, err}
		}()
		res, err = handler(ctx, def)
	}()

	var status, message, detailJSON, errorText string
	keep, quiet := 0, false
	applyOut := func(out outT) {
		if out.err != nil {
			status, message, errorText, keep = StatusFailed, "执行失败", out.err.Error(), 1
			if out.res != nil {
				if out.res.Message != "" {
					message = out.res.Message
				}
				detailJSON = marshalDetail(out.res.Detail)
			}
			return
		}
		status, message = StatusSuccess, "OK"
		if out.res != nil {
			if out.res.Message != "" {
				message = out.res.Message
			}
			detailJSON = marshalDetail(out.res.Detail)
			quiet = out.res.Quiet
		}
	}

	timedOut := false
	select {
	case out := <-ch:
		applyOut(out)
	case <-ctx.Done():
		// 超时与刚好完成可能同时就绪：优先取结果，避免误标 timeout
		select {
		case out := <-ch:
			applyOut(out)
		default:
			timedOut = true
			status, message, errorText, keep = StatusTimeout, "执行超时", ctx.Err().Error(), 1
		}
	}

	if timedOut {
		// 锁留给后台：同一 job_code 在旧 handler 未结束前不能再跑
		unlockByDefer = false
		go func() {
			<-ch
			mu.Unlock()
		}()
	}

	finished := nowUnix()
	durationMs := time.Since(time.Unix(startedAt, 0)).Milliseconds()
	if durationMs < 0 {
		durationMs = 0
	}
	run := &JobRun{
		RunUID:      newRunUID(),
		JobCode:     jobCode,
		Category:    def.Category,
		TriggerType: opts.Trigger,
		StartedAt:   startedAt,
		FinishedAt:  finished,
		DurationMs:  durationMs,
		Operator:    opts.Operator,
		Status:      status,
		Message:     message,
		DetailJSON:  detailJSON,
		ErrorText:   errorText,
		KeepForever: keep,
	}

	// 调度空跑成功：不落库
	if opts.Trigger == TriggerSchedule && status == StatusSuccess && quiet && !opts.Force {
		_ = bumpDefinitionAfterRun(jobCode, status, startedAt, finished, "")
		return run, nil
	}

	id, err := InsertRun(run)
	if err != nil {
		_ = bumpDefinitionAfterRun(jobCode, status, startedAt, finished, "写入执行记录失败: "+err.Error())
		return nil, err
	}
	run.ID = id
	_ = bumpDefinitionAfterRun(jobCode, status, startedAt, finished, errorText)
	maybePrune(cfg)

	if status != StatusSuccess {
		return run, fmt.Errorf("%s", firstNonEmpty(errorText, message))
	}
	return run, nil
}

func persistFailed(def *JobDefinition, opts RunOptions, startedAt int64, message, errText string) (*JobRun, error) {
	finished := nowUnix()
	durationMs := time.Since(time.Unix(startedAt, 0)).Milliseconds()
	if durationMs < 0 {
		durationMs = 0
	}
	run := &JobRun{
		RunUID:      newRunUID(),
		JobCode:     def.JobCode,
		Category:    def.Category,
		TriggerType: opts.Trigger,
		Status:      StatusFailed,
		StartedAt:   startedAt,
		FinishedAt:  finished,
		DurationMs:  durationMs,
		Message:     message,
		ErrorText:   errText,
		KeepForever: 1,
		Operator:    opts.Operator,
	}
	id, err := InsertRun(run)
	if err != nil {
		_ = bumpDefinitionAfterRun(def.JobCode, StatusFailed, startedAt, finished, errText)
		return nil, err
	}
	run.ID = id
	_ = bumpDefinitionAfterRun(def.JobCode, StatusFailed, startedAt, finished, errText)
	return run, nil
}

func maybePrune(cfg GlobalConfig) {
	if !cfg.AutoPrune {
		return
	}
	if _, err := PruneSuccessRuns(cfg.RunMaxCount); err != nil {
		log.Printf("[AutoJob] prune: %v", err)
	}
	if _, _, err := MaybeRenumberRunIDsIfNearLimit(); err != nil {
		log.Printf("[AutoJob] id renumber: %v", err)
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
