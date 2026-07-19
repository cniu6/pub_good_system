/**
 * 浏览器实例 ID：同一浏览器（含多标签页）共享，存在 localStorage。
 * 用途：登录/在线心跳时告诉后端「这是同一台浏览器」，避免多标签各自登录产生多条在线会话。
 */
const BROWSER_ID_KEY = 'fst_browser_id'

function createBrowserId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function')
    return crypto.randomUUID()
  return `b-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

export function getBrowserId(): string {
  if (typeof localStorage === 'undefined')
    return createBrowserId()
  try {
    const existing = localStorage.getItem(BROWSER_ID_KEY)
    if (existing && existing.trim())
      return existing.trim()
    const next = createBrowserId()
    localStorage.setItem(BROWSER_ID_KEY, next)
    return next
  }
  catch {
    return createBrowserId()
  }
}
