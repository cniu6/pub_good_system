package utils

import (
	"sync"
	"time"
)

// CodeSendLimiter IP+联系方式 验证码发送配额（内存级，单机足够）
type CodeSendLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*codeBucket
	window   time.Duration
	maxPerIP int
	maxPerContact int
}

type codeBucket struct {
	count   int
	resetAt time.Time
}

var DefaultCodeSendLimiter = NewCodeSendLimiter(time.Hour, 30, 10)

func NewCodeSendLimiter(window time.Duration, maxPerIP, maxPerContact int) *CodeSendLimiter {
	return &CodeSendLimiter{
		buckets:       make(map[string]*codeBucket),
		window:        window,
		maxPerIP:      maxPerIP,
		maxPerContact: maxPerContact,
	}
}

func (l *CodeSendLimiter) Allow(ip, contact string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// 偶尔清理过期
	if len(l.buckets) > 5000 {
		for k, b := range l.buckets {
			if now.After(b.resetAt) {
				delete(l.buckets, k)
			}
		}
	}
	return l.allowKey("ip:"+ip, l.maxPerIP, now) && l.allowKey("c:"+contact, l.maxPerContact, now)
}

func (l *CodeSendLimiter) allowKey(key string, max int, now time.Time) bool {
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &codeBucket{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if b.count >= max {
		return false
	}
	b.count++
	return true
}
