package middleware

import (
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type DynamicRateLimitSnapshot struct {
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	Rate              int    `json:"rate"`
	Burst             int    `json:"burst"`
	AllowedCount      int64  `json:"allowed_count"`
	BlockedCount      int64  `json:"blocked_count"`
	TotalCount        int64  `json:"total_count"`
	ActiveVisitors    int    `json:"active_visitors"`
	LastConfigReload  string `json:"last_config_reload"`
	CleanupIntervalMs int64  `json:"cleanup_interval_ms"`
}

type dynamicRateLimiterState struct {
	name            string
	keyFunc         func(*gin.Context) string
	mu              sync.RWMutex
	limiter         *RateLimiter
	enabled         bool
	rate            int
	burst           int
	lastConfigReload time.Time
	allowedCount    int64
	blockedCount    int64
}

var dynamicRateLimitersMu sync.RWMutex
var dynamicRateLimiters = map[string]*dynamicRateLimiterState{}

func ensureDynamicRateLimiterState(name string, keyFunc func(*gin.Context) string) *dynamicRateLimiterState {
	dynamicRateLimitersMu.Lock()
	defer dynamicRateLimitersMu.Unlock()
	state, ok := dynamicRateLimiters[name]
	if ok {
		return state
	}
	state = &dynamicRateLimiterState{name: name, keyFunc: keyFunc}
	dynamicRateLimiters[name] = state
	return state
}

func (s *dynamicRateLimiterState) update(enabled bool, rate, burst int) *RateLimiter {
	var oldLimiter *RateLimiter

	s.mu.Lock()
	if rate <= 0 {
		rate = 1
	}
	if burst <= 0 {
		burst = rate
	}
	if s.enabled != enabled || s.rate != rate || s.burst != burst || (enabled && s.limiter == nil) || (!enabled && s.limiter != nil) {
		oldLimiter = s.limiter
		s.enabled = enabled
		s.rate = rate
		s.burst = burst
		s.lastConfigReload = time.Now()
		if enabled {
			s.limiter = NewRateLimiter(RateLimitConfig{
				Rate:            rate,
				Burst:           burst,
				KeyFunc:         s.keyFunc,
				CleanupInterval: time.Minute,
			})
		} else {
			s.limiter = nil
		}
	}
	limiter := s.limiter
	s.mu.Unlock()

	if oldLimiter != nil {
		oldLimiter.Stop()
	}
	return limiter
}

func (s *dynamicRateLimiterState) snapshot() DynamicRateLimitSnapshot {
	s.mu.RLock()
	limiter := s.limiter
	enabled := s.enabled
	rate := s.rate
	burst := s.burst
	lastReload := s.lastConfigReload
	s.mu.RUnlock()

	activeVisitors := 0
	if limiter != nil {
		limiter.mu.RLock()
		activeVisitors = len(limiter.visitors)
		limiter.mu.RUnlock()
	}

	allowed := atomic.LoadInt64(&s.allowedCount)
	blocked := atomic.LoadInt64(&s.blockedCount)
	item := DynamicRateLimitSnapshot{
		Name:              s.name,
		Enabled:           enabled,
		Rate:              rate,
		Burst:             burst,
		AllowedCount:      allowed,
		BlockedCount:      blocked,
		TotalCount:        allowed + blocked,
		ActiveVisitors:    activeVisitors,
		CleanupIntervalMs: int64(time.Minute / time.Millisecond),
	}
	if !lastReload.IsZero() {
		item.LastConfigReload = lastReload.Format(time.RFC3339)
	}
	return item
}

func GetDynamicRateLimitSnapshots() []DynamicRateLimitSnapshot {
	dynamicRateLimitersMu.RLock()
	states := make([]*dynamicRateLimiterState, 0, len(dynamicRateLimiters))
	for _, state := range dynamicRateLimiters {
		states = append(states, state)
	}
	dynamicRateLimitersMu.RUnlock()

	items := make([]DynamicRateLimitSnapshot, 0, len(states))
	for _, state := range states {
		items = append(items, state.snapshot())
	}
	return items
}

func defaultAdminRateLimitKey(c *gin.Context) string {
	if uid, exists := c.Get("userID"); exists {
		if parsed, ok := uid.(uint64); ok {
			return "admin:user:" + strconv.FormatUint(parsed, 10)
		}
	}
	return "admin:ip:" + c.ClientIP()
}

// isGlobalRateLimitExemptPath 全局限流白名单。
//
// 支付回调（/public/payment/notify、/return）故意不参与全局限流：
// 支付网关的异步通知本来就该进来；若因 IP 令牌桶打满被 429 丢掉，
// 会出现「用户已付款但本系统未入账」。因此这里跳过是刻意设计，
// 不要给回调再加独立限流。验签/订单状态机仍由支付业务层负责防伪造。
//
// pprof 同样豁免，避免 profiling 采样被限流干扰（路径随 ADMIN_API_PATH 变化）。
func isGlobalRateLimitExemptPath(path string) bool {
	if path == "" {
		return false
	}
	if !strings.HasPrefix(path, "/api/") {
		return true
	}
	// 支付异步通知 / 同步回跳：必须放行，防止单次合法回调被限流丢弃
	if strings.HasPrefix(path, "/api/v1/public/payment/notify") ||
		strings.HasPrefix(path, "/api/v1/public/payment/return") {
		return true
	}
	adminPprofPrefix := "/api/v1/admin/debug/pprof"
	if cfg := config.GlobalConfig; cfg != nil {
		adminPprofPrefix = "/api/v1" + config.NormalizeAdminAPIPath(cfg.AdminAPIPath) + "/debug/pprof"
	}
	return strings.HasPrefix(path, adminPprofPrefix)
}

func DynamicGlobalRateLimitMiddleware() gin.HandlerFunc {
	state := ensureDynamicRateLimiterState("global_api", DefaultKeyFunc)
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if isGlobalRateLimitExemptPath(path) {
			c.Next()
			return
		}

		cfg := services.GetGlobalAPIRateLimitRuntimeConfig()
		limiter := state.update(cfg.Enabled, cfg.Rate, cfg.Burst)
		if limiter == nil {
			c.Next()
			return
		}

		if !limiter.Allow(DefaultKeyFunc(c)) {
			atomic.AddInt64(&state.blockedCount, 1)
			utils.Fail(c, 429, "Requests too frequent, please retry later")
			c.Abort()
			return
		}
		atomic.AddInt64(&state.allowedCount, 1)
		c.Next()
	}
}

func DynamicAdminRateLimitMiddleware() gin.HandlerFunc {
	state := ensureDynamicRateLimiterState("admin_api", defaultAdminRateLimitKey)
	return func(c *gin.Context) {
		cfg := services.GetGlobalAdminRateLimitRuntimeConfig()
		limiter := state.update(cfg.Enabled, cfg.Rate, cfg.Burst)
		if limiter == nil {
			c.Next()
			return
		}

		if !limiter.Allow(defaultAdminRateLimitKey(c)) {
			atomic.AddInt64(&state.blockedCount, 1)
			utils.Fail(c, 429, "Admin interface requests too frequent, please retry later")
			c.Abort()
			return
		}
		atomic.AddInt64(&state.allowedCount, 1)
		c.Next()
	}
}
