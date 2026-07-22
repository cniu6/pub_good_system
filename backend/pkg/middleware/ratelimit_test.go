package middleware

import (
	"fmt"
	"testing"
	"time"
)

// TestRateLimiter_RecoversAfterBeingRejected 验证令牌桶修复：
// 高频（间隔小于 1/Rate 秒）连续请求耗尽令牌后，只要按真实速率等待，
// 桶必须能恢复，而不是像 int() 截断版本那样永久卡死在拒绝状态。
func TestRateLimiter_RecoversAfterBeingRejected(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		Rate:            5, // 每秒 5 个
		Burst:           2, // 突发上限 2，便于快速把桶打空
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	})
	defer rl.Stop()

	key := "recover-test"

	// 先把突发额度打空
	if !rl.Allow(key) {
		t.Fatal("第 1 次应放行（消耗 1 个突发令牌）")
	}
	if !rl.Allow(key) {
		t.Fatal("第 2 次应放行（消耗剩余突发令牌）")
	}

	// 桶已空，紧接着高频重试多次（每次间隔远小于 1/Rate=200ms）。
	// 修复前的 bug：int() 截断丢弃小数进度，且每次调用（无论放行/拒绝）都会刷新 last_seen，
	// 导致小数进度永远清零、桶永远恢复不了。
	rejectedDuringBurst := 0
	for i := 0; i < 20; i++ {
		if !rl.Allow(key) {
			rejectedDuringBurst++
		}
		time.Sleep(2 * time.Millisecond) // 20 次 * 2ms = 40ms，远小于恢复 1 个令牌所需的 200ms
	}
	if rejectedDuringBurst != 20 {
		t.Fatalf("桶应仍处于耗尽状态，拒绝次数=%d, want 20", rejectedDuringBurst)
	}

	// 停止请求，等待超过恢复 1 个令牌所需时间（1/Rate = 200ms），桶必须能恢复
	time.Sleep(250 * time.Millisecond)
	if !rl.Allow(key) {
		t.Fatal("等待恢复窗口后应重新放行（token bucket 应已恢复至少 1 个令牌）")
	}
}

// TestRateLimiter_FractionalTokensAccumulateAcrossCalls 验证小数令牌不会被逐次调用截断丢失：
// 即使每次调用间隔很短（每次只应补充 0 点几个令牌），累积多次后总量应等于按总耗时计算的结果，
// 而不是像 int() 截断版本那样永远停在 0。
func TestRateLimiter_FractionalTokensAccumulateAcrossCalls(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		Rate:            10, // 每秒 10 个，每次补充理论值 = elapsed*10
		Burst:           1,
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	})
	defer rl.Stop()

	key := "fraction-test"
	if !rl.Allow(key) {
		t.Fatal("第 1 次应放行")
	}

	// 桶已空。以 5ms 间隔调用 40 次（总耗时 ~200ms，理论应恢复 ~2 个令牌）。
	// 修复前：int(0.005*10)=int(0.05)=0，每次都截断为 0，永远无法放行。
	allowedAgain := false
	for i := 0; i < 40; i++ {
		time.Sleep(5 * time.Millisecond)
		if rl.Allow(key) {
			allowedAgain = true
			break
		}
	}
	if !allowedAgain {
		t.Fatal("小数令牌应能跨多次调用累积，最终应有一次放行；修复前会永远拒绝")
	}
}

// TestRateLimiter_VisitorHardCapEvictsWhenFull 验证访问者表达到硬上限时会自动腾位置，
// 不会无限增长（防止大量不同 key 把内存刷爆）。
func TestRateLimiter_VisitorHardCapEvictsWhenFull(t *testing.T) {
	savedCap := visitorHardCap
	visitorHardCap = 5
	defer func() { visitorHardCap = savedCap }()

	rl := NewRateLimiter(RateLimitConfig{
		Rate:            1,
		Burst:           1,
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	})
	defer rl.Stop()

	// 灌入远超上限数量的不同 key
	for i := 0; i < 50; i++ {
		rl.Allow(fmt.Sprintf("visitor-%d", i))
	}

	rl.mu.RLock()
	count := len(rl.visitors)
	rl.mu.RUnlock()

	if count > visitorHardCap {
		t.Fatalf("访问者表大小=%d，不应超过硬上限=%d", count, visitorHardCap)
	}
}

// TestRateLimiter_BasicBurstAndBlock 基础用例：突发额度内放行，超出后拒绝。
func TestRateLimiter_BasicBurstAndBlock(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		Rate:            1,
		Burst:           3,
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	})
	defer rl.Stop()

	key := "burst-test"
	for i := 0; i < 3; i++ {
		if !rl.Allow(key) {
			t.Fatalf("第 %d 次应在突发额度内放行", i+1)
		}
	}
	if rl.Allow(key) {
		t.Fatal("超出突发额度应被拒绝")
	}
}

// TestRateLimiter_DifferentKeysIndependent 不同 key 的令牌桶应互相独立，不共享额度。
func TestRateLimiter_DifferentKeysIndependent(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		Rate:            1,
		Burst:           1,
		KeyFunc:         DefaultKeyFunc,
		CleanupInterval: time.Minute,
	})
	defer rl.Stop()

	if !rl.Allow("key-a") {
		t.Fatal("key-a 第 1 次应放行")
	}
	if rl.Allow("key-a") {
		t.Fatal("key-a 第 2 次应被拒绝（突发已用完）")
	}
	if !rl.Allow("key-b") {
		t.Fatal("key-b 是独立的桶，应放行")
	}
}
