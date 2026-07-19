package presence

import (
	"sync"
	"time"
)

type fixedWindowLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	limit   int
	window  time.Duration
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
)
