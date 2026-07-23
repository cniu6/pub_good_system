package presence

import (
	"strconv"
	"sync"
	"time"
)

type fixedWindowLimiter struct {
	mu        sync.Mutex
	entries   map[string]*rateEntry
	limit     int
	window    time.Duration
	lastSweep time.Time
}

type rateEntry struct {
	start time.Time
	count int
}

func newFixedWindowLimiter(limit int, window time.Duration) *fixedWindowLimiter {
	return &fixedWindowLimiter{entries: make(map[string]*rateEntry), limit: limit, window: window}
}

func (l *fixedWindowLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastSweep.IsZero() || now.Sub(l.lastSweep) >= l.window {
		for key, item := range l.entries {
			if now.Sub(item.start) >= l.window {
				delete(l.entries, key)
			}
		}
		l.lastSweep = now
	}
	entry := l.entries[key]
	if entry == nil || now.Sub(entry.start) >= l.window {
		l.entries[key] = &rateEntry{start: now, count: 1}
		return true
	}
	if entry.count >= l.limit {
		return false
	}
	entry.count++
	return true
}

// Presence 使用独立限流器，避免影响 global_api 的配额与统计。
var (
	handshakeLimiter = newFixedWindowLimiter(10, time.Minute)
	ticketLimiter    = newFixedWindowLimiter(12, time.Minute)
)

// AllowWSTicketIssue 限制单个用户在短时间内签发 WS ticket 的次数。
// ticket 只应在建连/重连时申请；该限流用于兜底前端异常重试，防止短时未消费票据堆积在内存。
func AllowWSTicketIssue(userID uint64, authGuard string) bool {
	return ticketLimiter.allow(strconv.FormatUint(userID, 10) + ":" + authGuard)
}
