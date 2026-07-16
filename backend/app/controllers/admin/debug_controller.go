package admin

import (
	"bytes"
	"fmt"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"regexp"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gpprof "github.com/google/pprof/profile"
)

type DebugController struct{}

const suspectedLeakWaitThreshold = 5 * time.Minute

func NewDebugController() *DebugController {
	return &DebugController{}
}

var (
	goroutineHeaderPattern = regexp.MustCompile(`^goroutine\s+(\d+)\s+\[(.+)\]:$`)
	waitDurationPattern    = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(hour|hours|minute|minutes|second|seconds|millisecond|milliseconds|ms)\b`)
	knownSystemGoroutineFunctionPatterns = []string{
		"fst/backend/utils.ServeHTTPServer",
		"database/sql.",
		"github.com/go-sql-driver/mysql.",
		"os/signal.",
		"fst/backend/app/services.StartCleanupTask.",
		"fst/backend/app/services.StartExpiredOrderTask.",
		"fst/backend/pkg/middleware.(*RateLimiter).cleanupRoutine",
	}
	knownSystemGoroutineCreatorPatterns = []string{
		"database/sql.OpenDB",
		"os/signal.Notify",
		"github.com/go-sql-driver/mysql.(*mysqlConn).startWatcher",
		"fst/backend/app/services.StartCleanupTask.",
		"fst/backend/app/services.StartExpiredOrderTask.",
		"fst/backend/pkg/middleware.NewRateLimiter",
	}
)

type RuntimeGoroutineInfo struct {
	ID             int64  `json:"id"`
	State          string `json:"state"`
	WaitTime       string `json:"wait_time,omitempty"`
	Stack          string `json:"stack"`
	Function       string `json:"function"`
	LockedToThread bool   `json:"locked_to_thread,omitempty"`
	CreatedBy      string `json:"created_by,omitempty"`
	StackLines     int    `json:"stack_lines"`
}

type parsedRuntimeGoroutine struct {
	info         RuntimeGoroutineInfo
	waitDuration time.Duration
	rawBlock     string
}

func captureGoroutineProfileText(debugLevel int) (string, error) {
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return "", fmt.Errorf("goroutine profile not found")
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func splitGoroutineBlocks(profileText string) []string {
	normalized := strings.ReplaceAll(profileText, "\r\n", "\n")
	rawBlocks := strings.Split(normalized, "\n\n")
	blocks := make([]string, 0, len(rawBlocks))

	for _, block := range rawBlocks {
		block = strings.TrimSpace(block)
		if block == "" || !strings.HasPrefix(block, "goroutine ") {
			continue
		}
		blocks = append(blocks, block)
	}

	return blocks
}

func parseGoroutineState(state string) (string, string, bool) {
	parts := strings.Split(state, ",")
	parsedState := ""
	waitTime := ""
	lockedToThread := false

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		switch {
		case strings.EqualFold(part, "locked to thread"):
			lockedToThread = true
		case waitTime == "":
			if _, ok := extractWaitDuration(part); ok {
				waitTime = part
				continue
			}
			fallthrough
		default:
			if parsedState == "" {
				parsedState = part
			}
		}
	}

	if parsedState == "" {
		parsedState = strings.TrimSpace(state)
	}

	return parsedState, waitTime, lockedToThread
}

func parseRuntimeGoroutineBlocks(profileText string) []parsedRuntimeGoroutine {
	blocks := splitGoroutineBlocks(profileText)
	parsed := make([]parsedRuntimeGoroutine, 0, len(blocks))

	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		if len(lines) == 0 {
			continue
		}

		header := strings.TrimSpace(lines[0])
		matches := goroutineHeaderPattern.FindStringSubmatch(header)
		if len(matches) != 3 {
			continue
		}

		id, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			continue
		}

		state, waitTime, lockedToThread := parseGoroutineState(matches[2])
		waitDuration, _ := extractWaitDuration(waitTime)

		functionName := "-"
		if len(lines) > 1 {
			functionLine := strings.TrimSpace(lines[1])
			if idx := strings.Index(functionLine, "("); idx > 0 {
				functionName = strings.TrimSpace(functionLine[:idx])
			} else if functionLine != "" {
				functionName = functionLine
			}
		}

		createdBy := ""
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "created by ") {
				createdBy = strings.TrimSpace(strings.TrimPrefix(line, "created by "))
				break
			}
		}

		parsed = append(parsed, parsedRuntimeGoroutine{
			info: RuntimeGoroutineInfo{
				ID:             id,
				State:          state,
				WaitTime:       waitTime,
				Stack:          block,
				Function:       functionName,
				LockedToThread: lockedToThread,
				CreatedBy:      createdBy,
				StackLines:     len(lines),
			},
			waitDuration: waitDuration,
			rawBlock:     block,
		})
	}

	return parsed
}

func summarizeRuntimeState(state string) string {
	normalized := strings.ToLower(state)
	switch {
	case normalized == "running" || normalized == "runnable":
		return "running"
	case strings.Contains(normalized, "wait") || strings.Contains(normalized, "select") || strings.Contains(normalized, "sleep"):
		return "waiting"
	case strings.Contains(normalized, "chan"):
		return "channel"
	case strings.Contains(normalized, "syscall") || strings.Contains(normalized, "io wait") || strings.Contains(normalized, "poll"):
		return "syscall"
	case strings.Contains(normalized, "semacquire") || strings.Contains(normalized, "mutex"):
		return "mutex"
	default:
		return "other"
	}
}

func filterRuntimeGoroutinesByWait(stacks []parsedRuntimeGoroutine, minWait time.Duration) []parsedRuntimeGoroutine {
	if minWait <= 0 {
		return stacks
	}

	filtered := make([]parsedRuntimeGoroutine, 0, len(stacks))
	for _, stack := range stacks {
		if stack.waitDuration >= minWait {
			filtered = append(filtered, stack)
		}
	}

	return filtered
}

func sortRuntimeGoroutinesByWait(stacks []RuntimeGoroutineInfo) {
	sort.Slice(stacks, func(i, j int) bool {
		leftWait, _ := extractWaitDuration(stacks[i].WaitTime)
		rightWait, _ := extractWaitDuration(stacks[j].WaitTime)
		if leftWait == rightWait {
			return stacks[i].ID < stacks[j].ID
		}
		return leftWait > rightWait
	})
}

func matchRuntimeGoroutinePatterns(value string, patterns []string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	for _, pattern := range patterns {
		if pattern != "" && strings.Contains(trimmed, pattern) {
			return true
		}
	}

	return false
}

func isKnownSystemGoroutine(stack parsedRuntimeGoroutine) bool {
	if matchRuntimeGoroutinePatterns(stack.info.Function, knownSystemGoroutineFunctionPatterns) {
		return true
	}

	if matchRuntimeGoroutinePatterns(stack.info.CreatedBy, knownSystemGoroutineCreatorPatterns) {
		return true
	}

	return false
}

func isPotentialLeakRuntimeGoroutine(stack parsedRuntimeGoroutine) bool {
	if !isPotentialLeakState(stack.info.WaitTime) {
		return false
	}

	return !isKnownSystemGoroutine(stack)
}

func collectGoroutineTrackingStats() (int, int) {
	profileText, err := captureGoroutineProfileText(2)
	if err != nil {
		return runtime.NumGoroutine(), 0
	}

	stacks := parseRuntimeGoroutineBlocks(profileText)
	trackedCount := len(stacks)
	potentialLeaks := 0
	for _, stack := range stacks {
		if isPotentialLeakRuntimeGoroutine(stack) {
			potentialLeaks++
		}
	}

	if trackedCount == 0 {
		trackedCount = runtime.NumGoroutine()
	}

	return trackedCount, potentialLeaks
}

func isPotentialLeakState(state string) bool {
	waitDuration, ok := extractWaitDuration(state)
	if !ok {
		return false
	}
	return waitDuration >= suspectedLeakWaitThreshold
}

func extractWaitDuration(state string) (time.Duration, bool) {
	matches := waitDurationPattern.FindStringSubmatch(state)
	if len(matches) != 3 {
		return 0, false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, false
	}

	switch strings.ToLower(matches[2]) {
	case "hour", "hours":
		return time.Duration(value * float64(time.Hour)), true
	case "minute", "minutes":
		return time.Duration(value * float64(time.Minute)), true
	case "second", "seconds":
		return time.Duration(value * float64(time.Second)), true
	case "millisecond", "milliseconds", "ms":
		return time.Duration(value * float64(time.Millisecond)), true
	default:
		return 0, false
	}
}

// ensureAdminDebugAllowed 校验 debug/pprof 是否允许访问：生产永久关闭，非生产看 ENABLE_ADMIN_DEBUG。
func ensureAdminDebugAllowed(c *gin.Context) bool {
	if !config.IsAdminDebugOpsEnabled() {
		utils.Fail(c, 403, "调试接口已禁用（生产或 ENABLE_ADMIN_DEBUG=false）")
		return false
	}
	return true
}

// GetGoroutineStats 获取协程统计信息
func (ctrl *DebugController) GetGoroutineStats(c *gin.Context) {
	if !ensureAdminDebugAllowed(c) {
		return
	}
	includeStacks := c.Query("stacks") == "true"
	minWaitMinutes, _ := strconv.Atoi(c.DefaultQuery("min_wait_minutes", "0"))
	if minWaitMinutes < 0 {
		minWaitMinutes = 0
	}
	minWaitDuration := time.Duration(minWaitMinutes) * time.Minute

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	profileText, err := captureGoroutineProfileText(2)
	parsedStacks := make([]parsedRuntimeGoroutine, 0)
	if err == nil {
		parsedStacks = parseRuntimeGoroutineBlocks(profileText)
	}

	totalCount := runtime.NumGoroutine()
	if len(parsedStacks) > 0 {
		totalCount = len(parsedStacks)
	}

	trackedCount := len(parsedStacks)
	if trackedCount == 0 {
		trackedCount = totalCount
	}

	longRunning := make([]RuntimeGoroutineInfo, 0)
	potentialLeakStacks := make([]RuntimeGoroutineInfo, 0)
	runtimeStacks := make([]RuntimeGoroutineInfo, 0)
	runtimeStateSummary := make(map[string]int)
	byState := make(map[string]int)

	for _, stack := range parsedStacks {
		byState[stack.info.State]++

		if stack.waitDuration >= time.Minute {
			longRunning = append(longRunning, stack.info)
		}

		if isPotentialLeakRuntimeGoroutine(stack) {
			potentialLeakStacks = append(potentialLeakStacks, stack.info)
		}
	}

	sortRuntimeGoroutinesByWait(longRunning)
	sortRuntimeGoroutinesByWait(potentialLeakStacks)

	if includeStacks {
		filteredStacks := filterRuntimeGoroutinesByWait(parsedStacks, minWaitDuration)
		for _, stack := range filteredStacks {
			runtimeStacks = append(runtimeStacks, stack.info)
			runtimeStateSummary[summarizeRuntimeState(stack.info.State)]++
		}
		sortRuntimeGoroutinesByWait(runtimeStacks)
	}

	utils.Success(c, gin.H{
		"total_count":          totalCount,
		"tracked_count":        trackedCount,
		"potential_leaks":      len(potentialLeakStacks),
		"potential_leak_stacks": potentialLeakStacks,
		"long_running":         longRunning,
		"runtime_stacks":       runtimeStacks,
		"runtime_state_summary": runtimeStateSummary,
		"by_state":             byState,
		"num_cpu":              runtime.NumCPU(),
		"gomaxprocs":           runtime.GOMAXPROCS(0),
		"num_cgo_call":         runtime.NumCgoCall(),
		"mem_stats": gin.H{
			"heap_alloc":      memStats.HeapAlloc,
			"total_alloc":     memStats.TotalAlloc,
			"heap_sys":        memStats.HeapSys,
			"heap_inuse":      memStats.HeapInuse,
			"heap_idle":       memStats.HeapIdle,
			"heap_released":   memStats.HeapReleased,
			"heap_objects":    memStats.HeapObjects,
			"stack_inuse":     memStats.StackInuse,
			"stack_sys":       memStats.StackSys,
			"sys":             memStats.Sys,
			"mallocs":         memStats.Mallocs,
			"frees":           memStats.Frees,
			"next_gc":         memStats.NextGC,
			"last_gc":         memStats.LastGC,
			"pause_total_ns":  memStats.PauseTotalNs,
			"num_gc":          memStats.NumGC,
			"num_forced_gc":   memStats.NumForcedGC,
			"gc_cpu_fraction": memStats.GCCPUFraction,
		},
	})
}

// ForceGC 强制执行垃圾回收
func (ctrl *DebugController) ForceGC(c *gin.Context) {
	if !ensureAdminDebugAllowed(c) {
		return
	}
	beforeGoroutines := runtime.NumGoroutine()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	afterGoroutines := runtime.NumGoroutine()

	utils.Success(c, gin.H{
		"goroutines_before": beforeGoroutines,
		"goroutines_after":  afterGoroutines,
		"message":           "GC completed",
	})
}

// buildCPUProfileTextSummary captures CPU profile and returns text summary similar to mcbeproxy.
func buildCPUProfileTextSummary(seconds int) (string, error) {
	var buf bytes.Buffer
	if err := pprof.StartCPUProfile(&buf); err != nil {
		return "", err
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	pprof.StopCPUProfile()

	parsed, err := gpprof.Parse(&buf)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("CPU Profile (%d seconds)\n", seconds))
	result.WriteString(fmt.Sprintf("Duration: %.2fs\n", float64(parsed.DurationNanos)/1e9))
	result.WriteString(fmt.Sprintf("Samples: %d\n", len(parsed.Sample)))
	result.WriteString("=" + strings.Repeat("=", 79) + "\n\n")

	var total int64
	for _, sample := range parsed.Sample {
		if len(sample.Value) > 0 {
			total += sample.Value[0]
		}
	}

	funcSamples := make(map[string]int64)
	for _, sample := range parsed.Sample {
		if len(sample.Value) == 0 || len(sample.Location) == 0 {
			continue
		}
		loc := sample.Location[0]
		if len(loc.Line) == 0 || loc.Line[0].Function == nil {
			continue
		}
		name := loc.Line[0].Function.Name
		funcSamples[name] += sample.Value[0]
	}

	type item struct {
		name  string
		count int64
	}
	list := make([]item, 0, len(funcSamples))
	for name, count := range funcSamples {
		list = append(list, item{name: name, count: count})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].count > list[j].count })

	result.WriteString("Top functions by samples:\n")
	limit := 30
	if len(list) < limit {
		limit = len(list)
	}
	for i := 0; i < limit; i++ {
		percent := 0.0
		if total > 0 {
			percent = float64(list[i].count) * 100 / float64(total)
		}
		result.WriteString(fmt.Sprintf("%2d. %-70s %8d (%6.2f%%)\n", i+1, list[i].name, list[i].count, percent))
	}

	result.WriteString("\n")
	result.WriteString("Tip: use go tool pprof on binary profile for deeper analysis.\n")
	return result.String(), nil
}

func buildTraceTextSummary(seconds int) (string, error) {
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	beforeGoroutines := runtime.NumGoroutine()

	var buf bytes.Buffer
	if err := trace.Start(&buf); err != nil {
		return "", err
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	trace.Stop()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	afterGoroutines := runtime.NumGoroutine()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Execution Trace (%d seconds)\n", seconds))
	result.WriteString(fmt.Sprintf("Trace Size: %d bytes\n", buf.Len()))
	result.WriteString(fmt.Sprintf("NumCPU: %d\n", runtime.NumCPU()))
	result.WriteString(fmt.Sprintf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0)))
	result.WriteString(strings.Repeat("=", 80) + "\n\n")

	result.WriteString("Runtime Snapshot:\n")
	result.WriteString(fmt.Sprintf("- Goroutines: %d -> %d (delta %+d)\n", beforeGoroutines, afterGoroutines, afterGoroutines-beforeGoroutines))
	result.WriteString(fmt.Sprintf("- GC Count: %d -> %d (delta %+d)\n", before.NumGC, after.NumGC, int64(after.NumGC)-int64(before.NumGC)))
	result.WriteString(fmt.Sprintf("- HeapAlloc: %d -> %d bytes (delta %+d)\n", before.HeapAlloc, after.HeapAlloc, int64(after.HeapAlloc)-int64(before.HeapAlloc)))
	result.WriteString(fmt.Sprintf("- HeapInuse: %d -> %d bytes (delta %+d)\n", before.HeapInuse, after.HeapInuse, int64(after.HeapInuse)-int64(before.HeapInuse)))
	result.WriteString(fmt.Sprintf("- TotalAlloc: %d -> %d bytes (delta %+d)\n", before.TotalAlloc, after.TotalAlloc, int64(after.TotalAlloc)-int64(before.TotalAlloc)))
	result.WriteString(fmt.Sprintf("- PauseTotalNs: %d -> %d (delta %+d)\n", before.PauseTotalNs, after.PauseTotalNs, int64(after.PauseTotalNs)-int64(before.PauseTotalNs)))
	result.WriteString(fmt.Sprintf("- NumForcedGC: %d -> %d (delta %+d)\n", before.NumForcedGC, after.NumForcedGC, int64(after.NumForcedGC)-int64(before.NumForcedGC)))
	result.WriteString("\nTip: click download to save the binary trace and inspect it with Go trace tooling.\n")
	return result.String(), nil
}

func writeNamedProfile(c *gin.Context, profileName string) {
	if !ensureAdminDebugAllowed(c) {
		return
	}
	debug := c.DefaultQuery("debug", "0")
	debugLevel, _ := strconv.Atoi(debug)

	profile := pprof.Lookup(profileName)
	if profile == nil {
		c.String(404, "%s profile not found", profileName)
		return
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		c.String(500, "Could not write %s profile: %v", profileName, err)
		return
	}

	if debugLevel > 0 {
		c.String(200, buf.String())
	} else {
		c.Data(200, "application/octet-stream", buf.Bytes())
	}
}

// GetPprofProfile CPU profile
func (ctrl *DebugController) GetPprofProfile(c *gin.Context) {
	if !ensureAdminDebugAllowed(c) {
		return
	}
	secondsStr := c.DefaultQuery("seconds", "30")
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 1 || seconds > 120 {
		seconds = 30
	}

	if c.Query("binary") != "1" {
		text, textErr := buildCPUProfileTextSummary(seconds)
		if textErr != nil {
			c.String(500, "Could not build CPU profile text: %v", textErr)
			return
		}
		c.String(200, text)
		return
	}

	var buf bytes.Buffer
	if err := pprof.StartCPUProfile(&buf); err != nil {
		c.String(500, "Could not enable CPU profiling: %v", err)
		return
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	pprof.StopCPUProfile()

	c.Data(200, "application/octet-stream", buf.Bytes())
}

// GetPprofHeap Heap profile
func (ctrl *DebugController) GetPprofHeap(c *gin.Context) {
	writeNamedProfile(c, "heap")
}

// GetPprofGoroutine Goroutine profile
func (ctrl *DebugController) GetPprofGoroutine(c *gin.Context) {
	if !ensureAdminDebugAllowed(c) {
		return
	}
	debugLevel, _ := strconv.Atoi(c.DefaultQuery("debug", "0"))
	minWaitMinutes, _ := strconv.Atoi(c.DefaultQuery("min_wait_minutes", "0"))
	if minWaitMinutes < 0 {
		minWaitMinutes = 0
	}

	if debugLevel >= 2 && minWaitMinutes > 0 {
		profileText, err := captureGoroutineProfileText(debugLevel)
		if err != nil {
			c.String(500, "Could not write goroutine profile: %v", err)
			return
		}

		filtered := filterRuntimeGoroutinesByWait(parseRuntimeGoroutineBlocks(profileText), time.Duration(minWaitMinutes)*time.Minute)
		blocks := make([]string, 0, len(filtered))
		for _, stack := range filtered {
			blocks = append(blocks, stack.rawBlock)
		}

		c.String(200, strings.Join(blocks, "\n\n"))
		return
	}

	writeNamedProfile(c, "goroutine")
}

// GetPprofAllocs Allocs profile
func (ctrl *DebugController) GetPprofAllocs(c *gin.Context) {
	writeNamedProfile(c, "allocs")
}

// GetPprofBlock Block profile
func (ctrl *DebugController) GetPprofBlock(c *gin.Context) {
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)
	writeNamedProfile(c, "block")
}

// GetPprofMutex Mutex profile
func (ctrl *DebugController) GetPprofMutex(c *gin.Context) {
	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)
	writeNamedProfile(c, "mutex")
}

// GetPprofThreadCreate Thread creation profile
func (ctrl *DebugController) GetPprofThreadCreate(c *gin.Context) {
	writeNamedProfile(c, "threadcreate")
}

// GetPprofTrace Execution trace
func (ctrl *DebugController) GetPprofTrace(c *gin.Context) {
	if !ensureAdminDebugAllowed(c) {
		return
	}
	secondsStr := c.DefaultQuery("seconds", "5")
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 1 || seconds > 30 {
		seconds = 5
	}

	if c.Query("binary") != "1" {
		text, textErr := buildTraceTextSummary(seconds)
		if textErr != nil {
			c.String(500, "Could not build trace text: %v", textErr)
			return
		}
		c.String(200, text)
		return
	}

	var buf bytes.Buffer
	if err := trace.Start(&buf); err != nil {
		c.String(500, "Could not enable tracing: %v", err)
		return
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	trace.Stop()

	c.Data(200, "application/octet-stream", buf.Bytes())
}

// RegisterRoutes 注册调试路由。
// 生产环境不注册；非生产由 ENABLE_ADMIN_DEBUG / EnableAdminDebugOps 控制，默认开发可开。
func (ctrl *DebugController) RegisterRoutes(group *gin.RouterGroup) {
	if !config.IsAdminDebugOpsEnabled() {
		return
	}

	debug := group.Group("/debug")
	{
		debug.GET("/goroutines/stats", ctrl.GetGoroutineStats)
		debug.POST("/gc", ctrl.ForceGC)

		// pprof endpoints
		pprof := debug.Group("/pprof")
		{
			pprof.GET("/profile", ctrl.GetPprofProfile)
			pprof.GET("/heap", ctrl.GetPprofHeap)
			pprof.GET("/goroutine", ctrl.GetPprofGoroutine)
			pprof.GET("/allocs", ctrl.GetPprofAllocs)
			pprof.GET("/block", ctrl.GetPprofBlock)
			pprof.GET("/mutex", ctrl.GetPprofMutex)
			pprof.GET("/threadcreate", ctrl.GetPprofThreadCreate)
			pprof.GET("/trace", ctrl.GetPprofTrace)
		}
	}
}

