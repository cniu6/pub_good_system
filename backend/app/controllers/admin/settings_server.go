package admin

import (
	"context"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

var serverMonitorStartedAt = time.Now()

// RestartBackend restarts backend process after response is flushed.
// 仅非生产且开启管理端 debug 运维开关时可用；写审计日志后延迟退出。
func (ctrl *SettingsController) RestartBackend(c *gin.Context) {
	if !config.IsAdminDebugOpsEnabled() {
		utils.Fail(c, 403, "当前环境已禁用后端重启能力")
		return
	}

	adminID, _ := c.Get("userID")
	log.Printf("[SECURITY AUDIT] restart-backend | admin_id=%v | ip=%s", adminID, c.ClientIP())

	utils.Success(c, gin.H{"message": "Backend restart requested"})
	go func() {
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()
}

// GetServerMonitoringStatus 返回当前项目服务端运行监控快照。
func (ctrl *SettingsController) GetServerMonitoringStatus(c *gin.Context) {
	now := time.Now()

	var runtimeMem runtime.MemStats
	runtime.ReadMemStats(&runtimeMem)

	cpuUsage := 0.0
	if values, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(values) > 0 {
		cpuUsage = values[0]
	}

	vmTotalMB := 0.0
	vmUsedMB := 0.0
	vmUsedPercent := 0.0
	swapTotalMB := 0.0
	swapUsedMB := 0.0
	swapPercent := 0.0
	if vm, err := mem.VirtualMemory(); err == nil {
		vmTotalMB = bytesToMB(vm.Total)
		vmUsedMB = bytesToMB(vm.Used)
		vmUsedPercent = vm.UsedPercent
	}
	if swap, err := mem.SwapMemory(); err == nil {
		swapTotalMB = bytesToMB(swap.Total)
		swapUsedMB = bytesToMB(swap.Used)
		swapPercent = swap.UsedPercent
	}

	diskPath := "."
	diskTotalGB := 0.0
	diskUsedGB := 0.0
	diskUsedPercent := 0.0
	if du, err := disk.Usage(diskPath); err == nil {
		diskTotalGB = bytesToGB(du.Total)
		diskUsedGB = bytesToGB(du.Used)
		diskUsedPercent = du.UsedPercent
	}

	netBytesSent := uint64(0)
	netBytesRecv := uint64(0)
	netPacketsSent := uint64(0)
	netPacketsRecv := uint64(0)
	if counters, err := gnet.IOCounters(false); err == nil && len(counters) > 0 {
		netBytesSent = counters[0].BytesSent
		netBytesRecv = counters[0].BytesRecv
		netPacketsSent = counters[0].PacketsSent
		netPacketsRecv = counters[0].PacketsRecv
	}

	pid := int32(os.Getpid())
	procCPUPercent := 0.0
	procRSSMB := 0.0
	if p, err := process.NewProcess(pid); err == nil {
		if cpuPercent, err := p.CPUPercent(); err == nil {
			procCPUPercent = cpuPercent
		}
		if memInfo, err := p.MemoryInfo(); err == nil {
			procRSSMB = bytesToMB(memInfo.RSS)
		}
	}

	dbStatus := buildDatabaseStatus()
	smtpStatus := ctrl.buildSMTPStatus()

	utils.Success(c, gin.H{
		"generated_at":   now.Format(time.RFC3339),
		"uptime_seconds": int64(now.Sub(serverMonitorStartedAt).Seconds()),
		"app": gin.H{
			"name":       currentGlobalConfig().AppName,
			"mode":       currentGlobalConfig().AppMode,
			"port":       currentGlobalConfig().Port,
			"go_version": runtime.Version(),
		},
		"metrics": gin.H{
			"cpu": gin.H{
				"usage_percent": cpuUsage,
				"core_count":    runtime.NumCPU(),
			},
			"memory": gin.H{
				"total_mb":     vmTotalMB,
				"used_mb":      vmUsedMB,
				"used_percent": vmUsedPercent,
			},
			"swap": gin.H{
				"total_mb":     swapTotalMB,
				"used_mb":      swapUsedMB,
				"used_percent": swapPercent,
			},
			"disk": gin.H{
				"path":         diskPath,
				"total_gb":     diskTotalGB,
				"used_gb":      diskUsedGB,
				"used_percent": diskUsedPercent,
			},
			"network": gin.H{
				"bytes_sent":   netBytesSent,
				"bytes_recv":   netBytesRecv,
				"packets_sent": netPacketsSent,
				"packets_recv": netPacketsRecv,
			},
		},
		"process": gin.H{
			"pid":             pid,
			"goroutines":      runtime.NumGoroutine(),
			"process_cpu":     procCPUPercent,
			"process_rss_mb":  procRSSMB,
			"memory_alloc_mb": bytesToMB(runtimeMem.Alloc),
			"memory_sys_mb":   bytesToMB(runtimeMem.Sys),
			"heap_alloc_mb":   bytesToMB(runtimeMem.HeapAlloc),
			"heap_inuse_mb":   bytesToMB(runtimeMem.HeapInuse),
			"heap_idle_mb":    bytesToMB(runtimeMem.HeapIdle),
			"stack_inuse_mb":  bytesToMB(runtimeMem.StackInuse),
			"gc_count":        runtimeMem.NumGC,
			"gc_cpu_fraction": runtimeMem.GCCPUFraction,
		},
		"services": []gin.H{dbStatus, smtpStatus},
	})
}

func (ctrl *SettingsController) GetServerOperationsStatus(c *gin.Context) {
	apiLogConfig := services.GetGlobalAPILogRuntimeConfig()
	defs, _ := task.ListDefinitions("", "", nil)
	taskItems := make([]gin.H, 0, len(defs))
	for _, d := range defs {
		taskItems = append(taskItems, gin.H{
			"key":           d.JobCode,
			"label":         d.Name,
			"running":       d.LastStatus == task.StatusRunning,
			"interval_secs": d.IntervalSeconds,
			"last_status":   d.LastStatus,
			"last_message":  d.LastError,
			"scheduled":     d.Enabled == 1,
		})
	}
	utils.Success(c, gin.H{
		"tasks":       taskItems,
		"rate_limits": middleware.GetDynamicRateLimitSnapshots(),
		"api_log": gin.H{
			"enabled":                 apiLogConfig.Enabled,
			"query_days":              apiLogConfig.QueryDays,
			"max_count":               apiLogConfig.MaxCount,
			"per_user_limit_enabled":  apiLogConfig.PerUserLimitEnabled,
			"per_user_max_count":      apiLogConfig.PerUserMaxCount,
		},
	})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func bytesToMB(v uint64) float64 {
	return float64(v) / 1024.0 / 1024.0
}

func bytesToGB(v uint64) float64 {
	return float64(v) / 1024.0 / 1024.0 / 1024.0
}

// dbDriverDisplayName 把归一化驱动名转成管理端展示名（勿写死 MySQL）。
func dbDriverDisplayName() string {
	switch db.DriverName() {
	case "sqlite":
		return "SQLite"
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	default:
		if name := db.DriverName(); name != "" {
			return name
		}
		return "Database"
	}
}

func buildDatabaseStatus() gin.H {
	name := dbDriverDisplayName()
	database, err := db.SQLDB()
	if err != nil || database == nil {
		return gin.H{
			"name":      name,
			"db_driver": db.DriverName(),
			"status":    "down",
			"message":   "数据库连接未初始化",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return gin.H{
			"name":      name,
			"db_driver": db.DriverName(),
			"status":    "down",
			"message":   err.Error(),
		}
	}

	stats := database.Stats()
	return gin.H{
		"name":             name,
		"db_driver":        db.DriverName(),
		"status":           "up",
		"message":          "连接正常",
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
	}
}

func (ctrl *SettingsController) buildSMTPStatus() gin.H {
	settingMap, _ := models.GetSettingsMap([]string{"smtp_host", "smtp_port", "smtp_username", "smtp_password"})
	cfg := currentGlobalConfig()

	host := firstNonEmpty(settingMap["smtp_host"], cfg.SMTPHost)
	port := firstNonEmpty(settingMap["smtp_port"], cfg.SMTPPort)
	username := firstNonEmpty(settingMap["smtp_username"], cfg.SMTPUser)
	password := firstNonEmpty(settingMap["smtp_password"], cfg.SMTPPass)
	sender := strings.TrimSpace(cfg.SystemEmail)
	if sender == "" {
		sender = username
	}

	credentialsOk := (username == "" && password == "") || (username != "" && password != "")
	configured := host != "" && port != "" && sender != "" && credentialsOk
	if !configured {
		return gin.H{
			"name":       "SMTP",
			"status":     "warning",
			"message":    "SMTP 未完成配置",
			"configured": false,
		}
	}

	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return gin.H{
			"name":       "SMTP",
			"status":     "down",
			"message":    err.Error(),
			"configured": true,
			"host":       host,
			"port":       port,
		}
	}
	_ = conn.Close()

	return gin.H{
		"name":       "SMTP",
		"status":     "up",
		"message":    "连接正常",
		"configured": true,
		"host":       host,
		"port":       port,
	}
}
