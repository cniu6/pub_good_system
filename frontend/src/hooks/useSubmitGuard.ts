import type { Ref } from 'vue'
import { ref } from 'vue'

/**
 * 写操作防重复提交（连点 / 确认框连点）。
 * 以 UI loading + 锁为主，不依赖额外请求头。
 *
 * 用法 A — 页面自带 submitting：
 *   const { submitting, run } = useSubmitGuard()
 *   await run(async () => { await api.save(...) })
 *
 * 用法 B — 复用已有 Ref（对话框 onPositiveClick 等）：
 *   onPositiveClick: () => withSubmitLock(submitting, async () => { ... })
 *
 * 用法 C — 开关回调（n-switch onUpdateValue）：
 *   onUpdateValue: (v) => { void withSubmitLock(saving, async () => { await api.update(v) }) }
 *
 * 列表翻页/搜索用 useRequestGuard，不要用本锁。
 */

/** 用已有 Ref 加锁执行；进行中直接忽略二次调用 */
export async function withSubmitLock<T>(
  lock: Ref<boolean>,
  fn: () => Promise<T>,
): Promise<T | undefined> {
  if (lock.value)
    return undefined
  lock.value = true
  try {
    return await fn()
  }
  finally {
    lock.value = false
  }
}

/** 页面级提交中守卫，配合按钮 :loading / :disabled */
export function useSubmitGuard() {
  const submitting = ref(false)

  async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    return withSubmitLock(submitting, fn)
  }

  return { submitting, run }
}
