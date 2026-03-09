package middleware

import (
	"fst/backend/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// Rate 每秒允许的请求数
	Rate int
	// Burst 突发流量上限
	Burst int
	// KeyFunc 用于生成限流键的函数
	KeyFunc func(*gin.Context) string
	// CleanupInterval 清理过期记录的间隔
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = RateLimitConfig{
	Rate:            100, // 每秒100个请求
	Burst:           200, // 突发上限200
	KeyFunc:         DefaultKeyFunc,
	CleanupInterval: time.Minute,
}

// DefaultKeyFunc 默认的限流键生成函数（基于IP）
func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// visitor 访问者记录
type visitor struct {
	last_seen time.Time
	tokens    int
	mu        sync.Mutex
}

// RateLimiter 限流器
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	config   RateLimitConfig
	stop_ch  chan struct{}
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	limiter := &RateLimiter{
		visitors: make(map[string]*visitor),
		config:   config,
		stop_ch:  make(chan struct{}),
	}

	// 启动清理协程
	go limiter.cleanupRoutine()

	return limiter
}

// cleanupRoutine 定期清理过期的访问者记录
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stop_ch:
			return
		}
	}
}

// cleanup 清理过期记录
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-time.Minute * 5) // 5分钟未访问则清理

	for key, v := range rl.visitors {
		v.mu.Lock()
		if v.last_seen.Before(threshold) {
			delete(rl.visitors, key)
		}
		v.mu.Unlock()
	}
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	close(rl.stop_ch)
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	v, exists := rl.visitors[key]
	if !exists {
		v = &visitor{
			last_seen: time.Now(),
			tokens:    rl.config.Burst,
		}
		rl.visitors[key] = v
	}
	rl.mu.Unlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.last_seen)
	v.last_seen = now

	// 令牌桶算法：根据时间间隔补充令牌
	v.tokens += int(elapsed.Seconds() * float64(rl.config.Rate))
	if v.tokens > rl.config.Burst {
		v.tokens = rl.config.Burst
	}

	if v.tokens <= 0 {
		return false
	}

	v.tokens--
	return true
}

// RateLimitMiddleware 限流中间件（使用默认配置）
func RateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddlewareWithConfig(DefaultRateLimitConfig)
}

// RateLimitMiddlewareWithConfig 使用自定义配置的限流中间件
func RateLimitMiddlewareWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		if !limiter.Allow(key) {
			utils.Fail(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimitMiddleware 严格限流中间件
// 用于登录、注册等敏感接口
func StrictRateLimitMiddleware() gin.HandlerFunc {
	config := RateLimitConfig{
		Rate:            5,   // 每秒5个请求
		Burst:           10,  // 突发上限10
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	}
	return RateLimitMiddlewareWithConfig(config)
}

// IPRateLimitMiddleware 基于IP的限流中间件
func IPRateLimitMiddleware(rate, burst int) gin.HandlerFunc {
	config := RateLimitConfig{
		Rate:            rate,
		Burst:           burst,
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	}
	return RateLimitMiddlewareWithConfig(config)
}

// UserRateLimitMiddleware 基于用户ID的限流中间件
func UserRateLimitMiddleware(rate, burst int) gin.HandlerFunc {
	config := RateLimitConfig{
		Rate: rate,
		Burst: burst,
		KeyFunc: func(c *gin.Context) string {
			if uid, exists := c.Get("userID"); exists {
				return "user:" + strconv.FormatUint(uid.(uint64), 10)
			}
			return "ip:" + c.ClientIP()
		},
		CleanupInterval: time.Minute,
	}
	return RateLimitMiddlewareWithConfig(config)
}

// PathRateLimitMiddleware 基于路径的限流中间件
func PathRateLimitMiddleware(rate, burst int) gin.HandlerFunc {
	config := RateLimitConfig{
		Rate: rate,
		Burst: burst,
		KeyFunc: func(c *gin.Context) string {
			return "path:" + c.Request.URL.Path + ":" + c.ClientIP()
		},
		CleanupInterval: time.Minute,
	}
	return RateLimitMiddlewareWithConfig(config)
}
