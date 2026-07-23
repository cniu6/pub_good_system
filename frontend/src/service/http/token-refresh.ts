import { fetchUpdateToken } from '../api/user/login'
import { getRuntimeRouteMode } from '@/router/runtime-mode'
import { authStorage } from '@/utils'

export type LoginTokenPayload = Api.Login.Info & { expiresAt?: number }

const TOKEN_REFRESH_CHANNEL = 'fst-token-refresh'
const TOKEN_REFRESH_LOCK_PREFIX = 'fst-token-refresh-lock'
const TOKEN_REFRESH_LEASE_PREFIX = 'fst-token-refresh-lease'
const REFRESH_LEASE_TTL_MS = 15_000

let refreshPromise: Promise<LoginTokenPayload | null> | null = null
let tokenRefreshChannel: BroadcastChannel | null = null
let channelListenerAttached = false
const recentRefreshResults = new Map<string, LoginTokenPayload>()
const remoteRefreshWaiters = new Map<string, Promise<LoginTokenPayload | null>>()
const resolveRemoteRefreshWaiters = new Map<string, (payload: LoginTokenPayload | null) => void>()

function getCurrentAuthGuard(): 'user' | 'admin' {
  return getRuntimeRouteMode() === 'admin' ? 'admin' : 'user'
}

function getRefreshLockName(sessionKey: string): string {
  const scope = authStorage.getActiveScope()
  return `${TOKEN_REFRESH_LOCK_PREFIX}:${scope}:${getCurrentAuthGuard()}:${sessionKey}`
}

function getRefreshAheadSeconds(): number {
  const ahead = Number(import.meta.env.VITE_TOKEN_REFRESH_AHEAD || 60)
  return Number.isFinite(ahead) && ahead > 0 ? ahead : 60
}

/**
 * 刷新协调必须按「具体 refresh token」分组，不能只按 session/local + guard。
 * 否则多个 login-as 窗口都是 session + user/admin，会互相接收并覆盖 token。
 *
 * 这里只传递不可逆摘要，不把 refresh token 原文放进 BroadcastChannel 或 Web Locks 名称。
 */
async function getRefreshSessionKey(): Promise<string | null> {
  const refreshToken = authStorage.get('refreshToken')
  if (!refreshToken)
    return null

  const input = new TextEncoder().encode(refreshToken)
  if (globalThis.crypto?.subtle) {
    const digest = await globalThis.crypto.subtle.digest('SHA-256', input)
    return Array.from(new Uint8Array(digest), value => value.toString(16).padStart(2, '0')).join('')
  }

  // 极旧浏览器没有 Web Crypto 时仍需隔离不同会话；此分支只做兼容分组，不承担密码学用途。
  let first = 0x811C9DC5
  let second = 0x9E3779B9
  for (const value of input) {
    first = Math.imul(first ^ value, 0x01000193)
    second = Math.imul(second ^ value, 0x85EBCA6B)
  }
  return `${(first >>> 0).toString(36)}-${(second >>> 0).toString(36)}`
}

function ensureTokenRefreshChannel(): BroadcastChannel | null {
  if (typeof BroadcastChannel === 'undefined')
    return null
  if (!tokenRefreshChannel)
    tokenRefreshChannel = new BroadcastChannel(TOKEN_REFRESH_CHANNEL)
  return tokenRefreshChannel
}

function buildLoginInfoFromStorage(): LoginTokenPayload | null {
  const accessToken = authStorage.get('accessToken')
  const refreshToken = authStorage.get('refreshToken')
  if (!accessToken || !refreshToken)
    return null

  const userInfo = authStorage.get('userInfo')
  const expiresAt = authStorage.get('accessTokenExpiresAt')
  return {
    ...(userInfo || {}),
    accessToken,
    refreshToken,
    expiresAt: expiresAt ?? undefined,
  } as LoginTokenPayload
}

function isAccessTokenStillFresh(): boolean {
  const expiresAt = authStorage.get('accessTokenExpiresAt')
  if (!expiresAt)
    return false

  const now = Math.floor(Date.now() / 1000)
  return expiresAt - now > getRefreshAheadSeconds()
}

function buildLoginInfoFromPayload(payload: LoginTokenPayload): LoginTokenPayload {
  const storedRoles = authStorage.get('role')
  const fallbackRoles = Array.isArray(storedRoles)
    ? storedRoles
    : (storedRoles ? [storedRoles] : ['user'])
  const rawRoles = Array.isArray(payload.role) && payload.role.length
    ? payload.role
    : fallbackRoles
  const role = rawRoles.map(item => (item === 'admin' || item === 'super' ? 'admin' : 'user')) as Entity.RoleType[]
  return {
    ...(authStorage.get('userInfo') || {}),
    ...payload,
    role,
  } as LoginTokenPayload
}

// 锁内先落盘再释放锁：等待锁的同会话标签重新检查 storage 后会直接复用新 token，
// 不会拿已轮换的旧 refresh token 再请求一次、触发后端重放防护。
function persistRefreshedLoginInfo(payload: LoginTokenPayload) {
  const loginInfo = buildLoginInfoFromPayload(payload)
  authStorage.setActive('accessToken', loginInfo.accessToken)
  authStorage.setActive('refreshToken', loginInfo.refreshToken)
  authStorage.setActive('role', loginInfo.role)
  authStorage.setActive('userInfo', loginInfo)
  if (loginInfo.expiresAt)
    authStorage.setActive('accessTokenExpiresAt', loginInfo.expiresAt)
}

function ensureRemoteRefreshWaiter(sessionKey: string) {
  const existing = remoteRefreshWaiters.get(sessionKey)
  if (existing)
    return existing

  const promise = new Promise<LoginTokenPayload | null>((resolve) => {
    const timer = window.setTimeout(() => {
      if (resolveRemoteRefreshWaiters.get(sessionKey) === resolve) {
        resolveRemoteRefreshWaiters.delete(sessionKey)
        remoteRefreshWaiters.delete(sessionKey)
      }
      resolve(null)
    }, REFRESH_LEASE_TTL_MS)
    resolveRemoteRefreshWaiters.set(sessionKey, (payload) => {
      window.clearTimeout(timer)
      resolveRemoteRefreshWaiters.delete(sessionKey)
      remoteRefreshWaiters.delete(sessionKey)
      resolve(payload)
    })
  })
  remoteRefreshWaiters.set(sessionKey, promise)
  return promise
}

function resolveRemoteRefreshWaiter(sessionKey: string, payload: LoginTokenPayload | null) {
  const resolve = resolveRemoteRefreshWaiters.get(sessionKey)
  resolve?.(payload)
}

function rememberRefreshResult(sessionKey: string, payload: LoginTokenPayload) {
  recentRefreshResults.set(sessionKey, payload)
  window.setTimeout(() => {
    if (recentRefreshResults.get(sessionKey) === payload)
      recentRefreshResults.delete(sessionKey)
  }, REFRESH_LEASE_TTL_MS)
}

function broadcastRefreshStarted(sessionKey: string) {
  ensureTokenRefreshChannel()?.postMessage({
    type: 'refresh-started',
    scope: authStorage.getActiveScope(),
    guard: getCurrentAuthGuard(),
    sessionKey,
  })
}

function broadcastRefreshFinished(sessionKey: string) {
  ensureTokenRefreshChannel()?.postMessage({
    type: 'refresh-failed',
    scope: authStorage.getActiveScope(),
    guard: getCurrentAuthGuard(),
    sessionKey,
  })
}

async function broadcastTokenRefreshed(sessionKey: string, payload: LoginTokenPayload) {
  const nextSessionKey = await getRefreshSessionKey()
  const channel = ensureTokenRefreshChannel()
  channel?.postMessage({
    type: 'token-refreshed',
    scope: authStorage.getActiveScope(),
    guard: getCurrentAuthGuard(),
    sessionKey,
    nextSessionKey,
    payload,
  })
}

/**
 * 监听其他标签页的刷新结果，避免本页仍持有旧 refresh token 再次轮换触发重放吊销。
 */
export function initTokenRefreshSync(onRefreshed: (payload: LoginTokenPayload) => void) {
  const channel = ensureTokenRefreshChannel()
  if (!channel || channelListenerAttached)
    return

  channelListenerAttached = true
  channel.onmessage = async (event: MessageEvent) => {
    const data = event.data as {
      type?: string
      scope?: string
      guard?: string
      sessionKey?: string
      nextSessionKey?: string
      payload?: LoginTokenPayload
    } | null

    if (!data?.type || !data.sessionKey)
      return
    if (data.scope !== authStorage.getActiveScope())
      return
    if (data.guard !== getCurrentAuthGuard())
      return

    const currentSessionKey = await getRefreshSessionKey()
    // localStorage 会被刷新发起标签在锁内先更新；其它同会话标签此时读到的是新 token。
    // 因而同时允许「刷新前」和「刷新后」指纹匹配；其它独立 login-as 会话两者都不可能匹配。
    const isSameSession = !!currentSessionKey && (currentSessionKey === data.sessionKey
      || (data.type === 'token-refreshed' && currentSessionKey === data.nextSessionKey))
    if (!isSameSession)
      return

    if (data.type === 'refresh-started') {
      ensureRemoteRefreshWaiter(data.sessionKey)
      return
    }
    if (data.type === 'refresh-failed') {
      resolveRemoteRefreshWaiter(data.sessionKey, null)
      return
    }
    if (data.type !== 'token-refreshed' || !data.payload)
      return

    rememberRefreshResult(data.sessionKey, data.payload)
    resolveRemoteRefreshWaiter(data.sessionKey, data.payload)
    onRefreshed(data.payload)
  }
}

async function withStorageLease<T>(lockName: string, task: () => Promise<T>): Promise<T> {
  if (typeof localStorage === 'undefined')
    return task()

  const key = `${TOKEN_REFRESH_LEASE_PREFIX}:${lockName}`
  const owner = `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
  const deadline = Date.now() + REFRESH_LEASE_TTL_MS

  while (Date.now() < deadline) {
    try {
      const raw = localStorage.getItem(key)
      const current = raw ? JSON.parse(raw) as { owner?: string, expiresAt?: number } : null
      if (!current?.owner || !current.expiresAt || current.expiresAt <= Date.now()) {
        localStorage.setItem(key, JSON.stringify({ owner, expiresAt: Date.now() + REFRESH_LEASE_TTL_MS }))
        const acquired = localStorage.getItem(key)
        if (acquired && (JSON.parse(acquired) as { owner?: string }).owner === owner) {
          try {
            return await task()
          }
          finally {
            const latest = localStorage.getItem(key)
            if (latest && (JSON.parse(latest) as { owner?: string }).owner === owner)
              localStorage.removeItem(key)
          }
        }
      }
    }
    catch {
      // localStorage 被隐私模式禁用时，无法提供降级互斥，只能退化到单页 refreshPromise。
      return task()
    }
    await new Promise(resolve => window.setTimeout(resolve, 40 + Math.random() * 80))
  }

  return task()
}

async function withRefreshLock<T>(sessionKey: string, task: () => Promise<T>): Promise<T> {
  const lockName = getRefreshLockName(sessionKey)
  if (typeof navigator !== 'undefined' && navigator.locks?.request) {
    return navigator.locks.request(lockName, { mode: 'exclusive' }, task)
  }
  return withStorageLease(lockName, task)
}

async function executeRefreshWithLock(sessionKey: string): Promise<LoginTokenPayload | null> {
  return withRefreshLock(sessionKey, async () => {
    // 等待锁期间，其他标签页可能已完成刷新并写回 storage。
    if (isAccessTokenStillFresh())
      return buildLoginInfoFromStorage()

    // 独立 sessionStorage 标签不共享 storage，但会收到同会话的刷新结果。
    // 若已知其它标签正在刷新，先等它广播结果，避免轮换同一枚旧 refresh token。
    const recent = recentRefreshResults.get(sessionKey)
    if (recent)
      return recent
    const remoteWaiter = remoteRefreshWaiters.get(sessionKey)
    if (remoteWaiter) {
      const remoteResult = await remoteWaiter
      if (remoteResult)
        return remoteResult
    }

    const refreshToken = authStorage.get('refreshToken')
    if (!refreshToken)
      return null

    try {
      broadcastRefreshStarted(sessionKey)
      const result = await fetchUpdateToken({ refreshToken, authGuard: getCurrentAuthGuard() })
      const data = result.data as LoginTokenPayload | null
      if (result.isSuccess && data) {
        persistRefreshedLoginInfo(data)
        rememberRefreshResult(sessionKey, data)
        await broadcastTokenRefreshed(sessionKey, data)
        return data
      }
      resolveRemoteRefreshWaiter(sessionKey, null)
      broadcastRefreshFinished(sessionKey)
      return null
    }
    catch {
      resolveRemoteRefreshWaiter(sessionKey, null)
      broadcastRefreshFinished(sessionKey)
      return null
    }
  })
}

// 统一刷新请求：同页去重 + 跨标签页互斥锁，避免并发轮换 refresh token 触发重放吊销。
export async function refreshAuthToken(): Promise<LoginTokenPayload | null> {
  if (refreshPromise)
    return refreshPromise

  refreshPromise = getRefreshSessionKey().then((sessionKey) => {
    if (!sessionKey)
      return null
    return executeRefreshWithLock(sessionKey)
  }).finally(() => {
    refreshPromise = null
  })

  return refreshPromise
}
