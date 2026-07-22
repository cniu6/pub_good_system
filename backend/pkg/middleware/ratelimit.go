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

// DefaultKeyFunc 默认的限流键生成函数（基于IP）
func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// visitor 访问者记录
//
// 【bug 修复说明】tokens 原为 int：每次补充令牌用 int(elapsed*rate) 会把不足 1 个的小数部分
// 直接截断丢弃。当 Rate 较小（如严格限流 5/s）而客户端又高频重试（间隔 < 1/Rate 秒）时，
// 每次补充都算出 0，且不管本次放行与否都会刷新 last_seen——小数进度永远无法累积，
// 桶会一直卡在耗尽状态，直到客户端完全停止请求满 1/Rate 秒才能恢复。
// 改成 float64 后，小数进度逐次累加，无论调用多频繁都能按真实时间比例正确恢复令牌。
type visitor struct {
	last_seen time.Time
	tokens    float64
	mu        sync.Mutex
}

// visitorHardCap 单个限流器最多同时跟踪的访问者（一般是不同 IP/用户）数量。
// 【问题】限流器用进程内存 map 记录每个 key 的状态，仅靠"5 分钟不活跃才清理"兜底；
// 如果攻击者能从大量不同 IP（代理池/僵尸网络）持续打入不同 key，5 分钟窗口内 map 会
// 无限膨胀——防滥用组件自己反而成了内存消耗的攻击面。
// 加一个硬上限：达到上限时先尝试即时清理过期项，仍满则随机淘汰一条腾位置。
// 用 var（非 const）是为了方便单测缩小上限验证淘汰逻辑，生产代码不应修改此值。
var visitorHardCap = 50000

// RateLimiter 限流器
//
// 【架构说明 / 已知限制】本限流器基于进程内存 map，计数仅在单实例内有效。
// 当前为单包可执行文件 + nginx 代理的单机部署，够用。
// 若将来横向扩容（nginx 后挂多台应用服务器，共用同一数据库），本内存限流会“各算各的”，
// 等效阈值被放大 N 倍。推荐处置优先级：
//   1) 首选在 nginx 入口层做 limit_req/limit_conn（所有流量都经 nginx，天然全局，零代码）；
//   2) 若需应用层精细限流，再换 Redis 等集中式计数器。
// 需注意：真正的暴力破解防护（账号级锁定 login_failure/lock_until、验证码失败次数上限）
// 已落在共享数据库里，多实例下依然生效；本 IP 限流只是前置粗粒度防滥用，不是安全边界。
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	config   RateLimitConfig
	stop_ch  chan struct{}
	done_ch  chan struct{}
	stopOnce sync.Once
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	limiter := &RateLimiter{
		visitors: make(map[string]*visitor),
		config:   config,
		stop_ch:  make(chan struct{}),
		done_ch:  make(chan struct{}),
	}

	// 启动清理协程
	go limiter.cleanupRoutine()

	return limiter
}

// cleanupRoutine 定期清理过期的访问者记录
func (rl *RateLimiter) cleanupRoutine() {
	defer close(rl.done_ch)
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
	rl.stopOnce.Do(func() {
		close(rl.stop_ch)
	})
	<-rl.done_ch
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	v, exists := rl.visitors[key]
	if !exists {
		if len(rl.visitors) >= visitorHardCap {
			rl.evictLocked()
		}
		v = &visitor{
			last_seen: time.Now(),
			tokens:    float64(rl.config.Burst),
		}
		rl.visitors[key] = v
	}
	rl.mu.Unlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.last_seen)
	v.last_seen = now

	// 令牌桶算法：按时间间隔补充令牌（float64 保留小数进度，见 visitor 定义处说明）
	v.tokens += elapsed.Seconds() * float64(rl.config.Rate)
	if v.tokens > float64(rl.config.Burst) {
		v.tokens = float64(rl.config.Burst)
	}

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

// evictLocked 在已持有 rl.mu 写锁的前提下腾出访问者表空间：
// 先做一次即时清理（复用 cleanup 的过期判定逻辑），仍达上限则随机淘汰一条。
// 调用方必须已持有 rl.mu.Lock()。
func (rl *RateLimiter) evictLocked() {
	threshold := time.Now().Add(-time.Minute * 5)
	for key, v := range rl.visitors {
		v.mu.Lock()
		idle := v.last_seen.Before(threshold)
		v.mu.Unlock()
		if idle {
			delete(rl.visitors, key)
		}
	}
	if len(rl.visitors) >= visitorHardCap {
		// map 迭代顺序本身是随机的，取第一个即等效随机淘汰，无需额外维护"最久未访问"顺序
		for key := range rl.visitors {
			delete(rl.visitors, key)
			break
		}
	}
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

// UserRateLimitMiddleware 基于用户ID的限流中间件
func UserRateLimitMiddleware(rate, burst int) gin.HandlerFunc {
	config := RateLimitConfig{
		Rate:  rate,
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

