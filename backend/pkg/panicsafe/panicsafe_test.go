package panicsafe

import (
	"sync"
	"testing"
	"time"
)

// TestGoRecoversPanic 验证 panicsafe.Go 内部 panic 不会外泄（测试进程不崩溃即通过）。
func TestGoRecoversPanic(t *testing.T) {
	done := make(chan struct{})
	Go("test-panic", func() {
		defer close(done)
		panic("boom")
	})
	select {
	case <-done:
		// panic 被 recover，goroutine 正常收尾
	case <-time.After(2 * time.Second):
		t.Fatal("panicsafe.Go 未在预期时间内完成")
	}
}

// TestGoRunsNormalFunc 验证正常函数会被执行。
func TestGoRunsNormalFunc(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	ran := false
	Go("test-normal", func() {
		defer wg.Done()
		ran = true
	})
	wg.Wait()
	if !ran {
		t.Fatal("panicsafe.Go 未执行传入函数")
	}
}

// TestGoNilFuncNoPanic 传入 nil 时应安全返回，不启动 goroutine。
func TestGoNilFuncNoPanic(t *testing.T) {
	Go("test-nil", nil)
}
