// Package panicsafe 提供带 panic 兜底的 goroutine 启动工具。
//
// Go 语义下，任意 goroutine 里未被 recover 的 panic 会直接终止整个进程。
// 对于「发了就不管」的后台任务（异步写日志、发邮件、聚合统计等），必须自带 recover，
// 否则一次偶发 panic 就会拖垮整个服务。本包作为零内部依赖的叶子包，
// 供各层统一复用，不会与业务包产生循环依赖。
package panicsafe

import (
	"log"
	"runtime/debug"
)

// Go 在独立 goroutine 中执行 fn，并兜底 recover，避免后台任务 panic 拖垮整个进程。
// name 仅用于 panic 日志定位；fn 为 nil 时直接返回，不启动 goroutine。
func Go(name string, fn func()) {
	if fn == nil {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[panicsafe] %s panic recovered: %v\n%s", name, r, debug.Stack())
			}
		}()
		fn()
	}()
}
