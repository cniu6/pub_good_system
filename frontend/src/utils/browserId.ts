import { authStorage } from './storage'

/**
 * 浏览器会话 ID：
 * - 用户端 localStorage 会话：同一浏览器多标签共享，避免重复创建用户会话。
 * - 管理端 / login-as 的 sessionStorage 隔离会话：每个标签独立，避免后登录标签
 *   被后端识别为“同一浏览器的新会话”而撤销此前标签的管理员 Token。
 */
const LOCAL_BROWSER_ID_KEY = 'fst_browser_id'
const SESSION_BROWSER_ID_KEY = 'fst_session_browser_id'

function createBrowserId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function')
    return crypto.randomUUID()
  return `b-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

export function getBrowserId(): string {
  if (typeof window === 'undefined')
    return createBrowserId()

  try {
    const useSessionStorage = authStorage.getActiveScope() === 'session'
    const storage = useSessionStorage ? window.sessionStorage : window.localStorage
    const key = useSessionStorage ? SESSION_BROWSER_ID_KEY : LOCAL_BROWSER_ID_KEY
    const existing = storage.getItem(key)
    if (existing && existing.trim())
      return existing.trim()

    const next = createBrowserId()
    storage.setItem(key, next)
    return next
  }
  catch {
    return createBrowserId()
  }
}
