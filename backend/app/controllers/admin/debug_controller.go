package admin

import (
	"bytes"
	"fmt"
	"fst/backend/utils"
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

func NewDebugController() *DebugController {
	return &DebugController{}
}

// GetGoroutineStats 获取协程统计信息
func (ctrl *DebugController) GetGoroutineStats(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	utils.Success(c, gin.H{
		"total_count": runtime.NumGoroutine(),
		"mem_stats": gin.H{
			"heap_alloc":      memStats.HeapAlloc,
			"heap_sys":        memStats.HeapSys,
			"heap_inuse":      memStats.HeapInuse,
			"heap_idle":       memStats.HeapIdle,
			"heap_released":   memStats.HeapReleased,
			"stack_inuse":     memStats.StackInuse,
			"stack_sys":       memStats.StackSys,
			"sys":             memStats.Sys,
			"num_gc":          memStats.NumGC,
			"gc_cpu_fraction": memStats.GCCPUFraction,
		},
	})
}

// ForceGC 强制执行垃圾回收
func (ctrl *DebugController) ForceGC(c *gin.Context) {
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

// GetPprofProfile CPU profile
func (ctrl *DebugController) GetPprofProfile(c *gin.Context) {
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
	debug := c.DefaultQuery("debug", "0")
	debugLevel, _ := strconv.Atoi(debug)

	profile := pprof.Lookup("heap")
	if profile == nil {
		c.String(404, "heap profile not found")
		return
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		c.String(500, "Could not write heap profile: %v", err)
		return
	}

	if debugLevel > 0 {
		c.String(200, buf.String())
	} else {
		c.Data(200, "application/octet-stream", buf.Bytes())
	}
}

// GetPprofGoroutine Goroutine profile
func (ctrl *DebugController) GetPprofGoroutine(c *gin.Context) {
	debug := c.DefaultQuery("debug", "0")
	debugLevel, _ := strconv.Atoi(debug)

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		c.String(404, "goroutine profile not found")
		return
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		c.String(500, "Could not write goroutine profile: %v", err)
		return
	}

	if debugLevel > 0 {
		c.String(200, buf.String())
	} else {
		c.Data(200, "application/octet-stream", buf.Bytes())
	}
}

// GetPprofAllocs Allocs profile
func (ctrl *DebugController) GetPprofAllocs(c *gin.Context) {
	debug := c.DefaultQuery("debug", "0")
	debugLevel, _ := strconv.Atoi(debug)

	profile := pprof.Lookup("allocs")
	if profile == nil {
		c.String(404, "allocs profile not found")
		return
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		c.String(500, "Could not write allocs profile: %v", err)
		return
	}

	if debugLevel > 0 {
		c.String(200, buf.String())
	} else {
		c.Data(200, "application/octet-stream", buf.Bytes())
	}
}

// GetPprofBlock Block profile
func (ctrl *DebugController) GetPprofBlock(c *gin.Context) {
	debug := c.DefaultQuery("debug", "0")
	debugLevel, _ := strconv.Atoi(debug)

	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	profile := pprof.Lookup("block")
	if profile == nil {
		c.String(404, "block profile not found")
		return
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		c.String(500, "Could not write block profile: %v", err)
		return
	}

	if debugLevel > 0 {
		c.String(200, buf.String())
	} else {
		c.Data(200, "application/octet-stream", buf.Bytes())
	}
}

// GetPprofMutex Mutex profile
func (ctrl *DebugController) GetPprofMutex(c *gin.Context) {
	debug := c.DefaultQuery("debug", "0")
	debugLevel, _ := strconv.Atoi(debug)

	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)

	profile := pprof.Lookup("mutex")
	if profile == nil {
		c.String(404, "mutex profile not found")
		return
	}

	var buf bytes.Buffer
	if err := profile.WriteTo(&buf, debugLevel); err != nil {
		c.String(500, "Could not write mutex profile: %v", err)
		return
	}

	if debugLevel > 0 {
		c.String(200, buf.String())
	} else {
		c.Data(200, "application/octet-stream", buf.Bytes())
	}
}

// GetPprofTrace Execution trace
func (ctrl *DebugController) GetPprofTrace(c *gin.Context) {
	secondsStr := c.DefaultQuery("seconds", "5")
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 1 || seconds > 30 {
		seconds = 5
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

// RegisterRoutes 注册调试路由
func (ctrl *DebugController) RegisterRoutes(group *gin.RouterGroup) {
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
			pprof.GET("/trace", ctrl.GetPprofTrace)
		}
	}
}
