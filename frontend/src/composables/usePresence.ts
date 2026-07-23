import { computed, readonly, ref } from 'vue'
import { getBrowserId } from '@/utils/browserId'

// 默认上报周期（毫秒）；实际值优先取管理端可配置的「在线心跳上报周期」（默认30秒），见 startPresence 的 intervalMs 参数。
const DEFAULT_PING_INTERVAL = 30_000
const MAX_RECONNECT_DELAY = 30_000
const PRESENCE_CHANNEL = 'fst-presence-leader'

const connected = ref(false)
let socket: WebSocket | null = null
let pingTimer: ReturnType<typeof setInterval> | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempts = 0
// 下一次允许真正发起连接（含请求 ticket）的时间戳（绝对时间，不随选主切换重置）。
// 多标签选主抖动时，resignLeader 会清掉 reconnectTimer，但退避“配额”必须保留，
// 否则某个标签重新当选 leader 会立刻绕过退避直连，造成疯狂请求。
let nextAllowedConnectAt = 0
let shouldReconnect = false
let activeToken = ''
let handlingAuthFailure = false
// Presence 选主按「账号 + 登录端」分组，而非按 token 分组。管理端标签的 token 会按
// sessionStorage 隔离；若按 token 分组，同一管理员的每个标签都会独立建立 WS 并反复重连。
// 不同账号或 user/admin 两套登录态仍各自维护独立的 Presence 连接。
let activePresenceGroupKey = ''
let activePingInterval = DEFAULT_PING_INTERVAL
let forceLogoutHandler: (() => void) | null = null

// ── 多标签页领导选举：同一浏览器只让一个标签页建立 Presence，避免列表出现多行 ──
let isLeader = false
let tabId = ''
let presenceChannel: BroadcastChannel | null = null
let leaderHeartbeatTimer: ReturnType<typeof setInterval> | null = null
let leaderCheckTimer: ReturnType<typeof setInterval> | null = null
let becomeLeaderTimer: ReturnType<typeof setTimeout> | null = null
let lastLeaderSeenAt = 0

function ensureTabId() {
  if (!tabId)
    tabId = `tab-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
  return tabId
}

function buildPresenceGroupKey(userID: number, guard: 'admin' | 'user'): string {
  return `${guard}:${userID}`
}

function clearTimers() {
  if (pingTimer) {
    clearInterval(pingTimer)
    pingTimer = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

function getPresenceUrl(ticket: string) {
  // Prefer rawPath for direct backend WS (dev proxy often has ws:false).
  const apiBase = __URL_MAP__.url.rawPath || __URL_MAP__.url.path || window.location.origin
  const url = new URL(apiBase, window.location.origin)
  url.protocol = (url.protocol === 'https:' || url.protocol === 'wss:') ? 'wss:' : 'ws:'
  const basePath = url.pathname.replace(/\/$/, '')
  url.pathname = `${basePath}/api/v1/ws/presence`.replace(/\/{2,}/g, '/')
  if (!url.pathname.startsWith('/'))
    url.pathname = `/${url.pathname}`
  // 使用短时一次性 ticket，避免 JWT 出现在 URL / 代理日志
  url.searchParams.set('ticket', ticket)
  // 同浏览器实例 ID，供后端收口多标签重复会话
  url.searchParams.set('browser_id', getBrowserId())
  return url.toString()
}

interface PresenceTicketResult {
  ticket: string | null
  unauthorized: boolean
}

async function fetchPresenceTicket(): Promise<PresenceTicketResult> {
  try {
    const { request } = await import('@/service/http')
    const res = await request.Post<Service.ResponseResult<{ ticket: string }>>('/api/v1/system/ws-ticket')
    if (res?.isSuccess && res.data?.ticket)
      return { ticket: res.data.ticket, unauthorized: false }
    return { ticket: null, unauthorized: Number((res as { code?: unknown })?.code) === 401 }
  }
  catch (error: unknown) {
    // 网络/5xx 才走退避；401 表示本地会话已无效，继续刷 ticket 只会制造日志和流量。
    const code = error && typeof error === 'object' && 'code' in error
      ? Number((error as { code?: unknown }).code)
      : 0
    return { ticket: null, unauthorized: code === 401 }
  }
}

function handlePresenceUnauthorized() {
  if (handlingAuthFailure)
    return
  handlingAuthFailure = true
  shouldReconnect = false
  closeSocketOnly()
  forceLogoutHandler?.()
}

function scheduleReconnect() {
  if (!shouldReconnect || !activeToken || !isLeader || reconnectTimer)
    return
  const delay = Math.min(1000 * 2 ** reconnectAttempts, MAX_RECONNECT_DELAY)
  reconnectAttempts++
  nextAllowedConnectAt = Date.now() + delay
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    void connect()
  }, delay)
}

async function connect() {
  if (!shouldReconnect || !activeToken || !isLeader || socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING)
    return

  // 退避配额未到：常见于多标签选主刚好在退避等待期间发生切换——
  // 新 leader 不能立刻发请求，改为在剩余配额到期后再试，避免绕过退避疯狂重连。
  const waitMs = nextAllowedConnectAt - Date.now()
  if (waitMs > 0) {
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        void connect()
      }, waitMs)
    }
    return
  }

  const ticketResult = await fetchPresenceTicket()
  if (ticketResult.unauthorized) {
    handlePresenceUnauthorized()
    return
  }
  if (!ticketResult.ticket) {
    scheduleReconnect()
    return
  }

  try {
    socket = new WebSocket(getPresenceUrl(ticketResult.ticket))
  }
  catch {
    scheduleReconnect()
    return
  }

  socket.onopen = () => {
    connected.value = true
    reconnectAttempts = 0
    nextAllowedConnectAt = 0
    // 服务端把任意文本/二进制消息都视为心跳；使用文本 ping 便于排查。
    pingTimer = setInterval(() => {
      if (socket?.readyState === WebSocket.OPEN)
        socket.send('ping')
    }, activePingInterval)
  }

  socket.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data)
      if (message?.type === 'force_logout') {
        shouldReconnect = false
        clearTimers()
        socket?.close()
        forceLogoutHandler?.()
        return
      }
      // 公告发布实时事件：通知铃铛/其它标签刷新（仅同会话标签，不同账号标签的公告内容不通用）
      if (message?.type === 'announcement' || message?.type === 'unread_count') {
        window.dispatchEvent(new CustomEvent('fst:announcement', { detail: message }))
        try {
          presenceChannel?.postMessage({ type: 'announcement', payload: message, sessionKey: activePresenceGroupKey })
        }
        catch { /* ignore */ }
      }
    }
    catch {
      // 忽略非 JSON
    }
  }

  socket.onclose = () => {
    connected.value = false
    if (pingTimer) {
      clearInterval(pingTimer)
      pingTimer = null
    }
    socket = null
    scheduleReconnect()
  }

  socket.onerror = () => {
    // close 事件统一负责退避重连，避免 error/close 重复安排。
  }
}

function closeSocketOnly() {
  clearTimers()
  const currentSocket = socket
  socket = null
  connected.value = false
  currentSocket?.close()
}

function becomeLeader() {
  if (isLeader)
    return
  isLeader = true
  lastLeaderSeenAt = Date.now()
  presenceChannel?.postMessage({ type: 'leader', tabId: ensureTabId(), sessionKey: activePresenceGroupKey })
  if (!leaderHeartbeatTimer) {
    leaderHeartbeatTimer = setInterval(() => {
      if (isLeader)
        presenceChannel?.postMessage({ type: 'leader', tabId: ensureTabId(), sessionKey: activePresenceGroupKey })
    }, 2000)
  }
  void connect()
}

// 抢主前先随机抖动错开：同时开 N 个标签（或原 leader 标签被关掉）时，
// 每个标签判定"无主"的时刻几乎同一时刻到达，若立刻各自 becomeLeader()，
// 会瞬间打出 N 份 ws-ticket 请求 + N 个 WebSocket 连接。抖动等待若干毫秒后
// 重新确认一次"仍然无主"，谁的抖动最短谁先广播 leader，其它标签收到后就不会再抢。
function attemptBecomeLeader() {
  if (isLeader || becomeLeaderTimer)
    return
  const jitter = 50 + Math.random() * 250
  becomeLeaderTimer = setTimeout(() => {
    becomeLeaderTimer = null
    if (!shouldReconnect || isLeader || Date.now() - lastLeaderSeenAt <= 800)
      return
    becomeLeader()
  }, jitter)
}

// 主动让位时先广播一声：其它标签不用等 5s 心跳超时才发现无主，
// 收到即可直接抢主（仍走 attemptBecomeLeader 的抖动+复核，不会立刻一堆标签同时连）。
function notifyLeaderDeparture() {
  if (isLeader)
    presenceChannel?.postMessage({ type: 'leader-left', tabId, sessionKey: activePresenceGroupKey })
}

function resignLeader() {
  if (!isLeader)
    return
  isLeader = false
  if (leaderHeartbeatTimer) {
    clearInterval(leaderHeartbeatTimer)
    leaderHeartbeatTimer = null
  }
  closeSocketOnly()
}

function handlePageHide() {
  notifyLeaderDeparture()
}

function setupPresenceLeaderElection() {
  ensureTabId()
  if (typeof BroadcastChannel === 'undefined') {
    // 不支持 BroadcastChannel 时退化为单页直连（仍靠后端 browser_id 收口）
    becomeLeader()
    return
  }
  if (presenceChannel)
    return

  presenceChannel = new BroadcastChannel(PRESENCE_CHANNEL)
  presenceChannel.onmessage = (event) => {
    const data = event.data as { type?: string, tabId?: string, sessionKey?: string, payload?: unknown } | null
    if (!data?.type)
      return
    // 不同账号或 user/admin 登录端互不干扰，不能被对方的 leader 状态误判为"已有人接管"。
    if (data.sessionKey !== activePresenceGroupKey)
      return
    if (data.type === 'leader' && data.tabId && data.tabId !== tabId) {
      lastLeaderSeenAt = Date.now()
      if (isLeader && data.tabId < tabId) {
        // 字典序更小的 tab 优先当 leader，避免双 leader
        resignLeader()
      }
    }
    else if (data.type === 'need-leader') {
      if (isLeader)
        presenceChannel?.postMessage({ type: 'leader', tabId: ensureTabId(), sessionKey: activePresenceGroupKey })
    }
    else if (data.type === 'announcement') {
      // 非 leader 标签也刷新铃铛
      window.dispatchEvent(new CustomEvent('fst:announcement', { detail: data.payload }))
    }
    else if (data.type === 'leader-left') {
      // 旧 leader 主动让位（关闭标签/登出），不必等心跳超时，立即尝试抢主
      if (!isLeader)
        attemptBecomeLeader()
    }
  }
  window.addEventListener('pagehide', handlePageHide)

  // 先问有没有现成 leader；没有则自己上位（经 attemptBecomeLeader 抖动+复核，避免多标签同时抢主）
  presenceChannel.postMessage({ type: 'need-leader', tabId, sessionKey: activePresenceGroupKey })
  window.setTimeout(() => {
    if (!shouldReconnect)
      return
    if (!isLeader && Date.now() - lastLeaderSeenAt > 2500)
      attemptBecomeLeader()
  }, 300)

  if (!leaderCheckTimer) {
    leaderCheckTimer = setInterval(() => {
      if (!shouldReconnect)
        return
      if (!isLeader && Date.now() - lastLeaderSeenAt > 5000)
        attemptBecomeLeader()
    }, 2000)
  }
}

function teardownPresenceLeaderElection() {
  if (becomeLeaderTimer) {
    clearTimeout(becomeLeaderTimer)
    becomeLeaderTimer = null
  }
  // 主动登出/停用时，若自己是 leader，先广播让位再关闭连接，其它标签能立刻接棒。
  notifyLeaderDeparture()
  resignLeader()
  if (leaderCheckTimer) {
    clearInterval(leaderCheckTimer)
    leaderCheckTimer = null
  }
  window.removeEventListener('pagehide', handlePageHide)
  if (presenceChannel) {
    presenceChannel.close()
    presenceChannel = null
  }
}

/**
 * 建立 Presence 心跳连接。
 * @param token 当前会话 JWT，用于 Presence WebSocket 鉴权
 * @param onForceLogout 服务端强退时的本地登出回调
 * @param userID 当前登录账号 ID，用于跨标签页 Presence 选主。
 * @param guard 当前登录端类型，admin 与 user 会话互不复用连接。
 * @param intervalMs 心跳上报周期（毫秒），来自管理端可配置的「在线心跳上报周期」设置，默认 30 秒。
 */
export function startPresence(
  token: string,
  onForceLogout: () => void,
  userID: number,
  guard: 'admin' | 'user',
  intervalMs?: number,
) {
  if (!token || !Number.isSafeInteger(userID) || userID <= 0)
    return
  forceLogoutHandler = onForceLogout
  activePingInterval = intervalMs && intervalMs > 0 ? intervalMs : DEFAULT_PING_INTERVAL
  shouldReconnect = true
  // Token 刷新时必须换连接，否则服务端仍会验证旧会话；换新 token 视为全新开始，清空旧退避配额。
  if (activeToken && activeToken !== token) {
    closeSocketOnly()
    reconnectAttempts = 0
    nextAllowedConnectAt = 0
    handlingAuthFailure = false
  }
  activeToken = token
  activePresenceGroupKey = buildPresenceGroupKey(userID, guard)
  shouldReconnect = true
  setupPresenceLeaderElection()
  if (isLeader)
    connect()
}

export function stopPresence() {
  shouldReconnect = false
  // 必须先让位广播（带着仍然有效的分组 key），再清空它，
  // 否则同组标签会因为 key 不匹配而过滤掉这条"让位"消息，白白多等 5s 心跳超时。
  teardownPresenceLeaderElection()
  activeToken = ''
  activePresenceGroupKey = ''
  handlingAuthFailure = false
  reconnectAttempts = 0
  nextAllowedConnectAt = 0
  closeSocketOnly()
}

export function usePresence() {
  return {
    connected: readonly(connected),
    isConnected: computed(() => connected.value),
    start: startPresence,
    stop: stopPresence,
  }
}
