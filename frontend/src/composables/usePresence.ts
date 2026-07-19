import { computed, readonly, ref } from 'vue'
import { getBrowserId } from '@/utils/browserId'

// 默认上报周期（毫秒）；实际值优先取管理端可配置的「在线心跳上报周期」（默认30秒），见 startPresence 的 intervalMs 参数。
const DEFAULT_PING_INTERVAL = 25_000
const MAX_RECONNECT_DELAY = 30_000
const PRESENCE_CHANNEL = 'fst-presence-leader'

const connected = ref(false)
let socket: WebSocket | null = null
let pingTimer: ReturnType<typeof setInterval> | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempts = 0
let shouldReconnect = false
let activeToken = ''
let activePingInterval = DEFAULT_PING_INTERVAL
let forceLogoutHandler: (() => void) | null = null

// ── 多标签页领导选举：同一浏览器只让一个标签页建立 Presence，避免列表出现多行 ──
let isLeader = false
let tabId = ''
let presenceChannel: BroadcastChannel | null = null
let leaderHeartbeatTimer: ReturnType<typeof setInterval> | null = null
let leaderCheckTimer: ReturnType<typeof setInterval> | null = null
let lastLeaderSeenAt = 0

function ensureTabId() {
  if (!tabId)
    tabId = `tab-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
  return tabId
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

function getPresenceUrl(token: string) {
  // Prefer rawPath for direct backend WS (dev proxy often has ws:false).
  const apiBase = __URL_MAP__.url.rawPath || __URL_MAP__.url.path || window.location.origin
  const url = new URL(apiBase, window.location.origin)
  url.protocol = (url.protocol === 'https:' || url.protocol === 'wss:') ? 'wss:' : 'ws:'
  const basePath = url.pathname.replace(/\/$/, '')
  url.pathname = `${basePath}/api/v1/ws/presence`.replace(/\/{2,}/g, '/')
  if (!url.pathname.startsWith('/'))
    url.pathname = `/${url.pathname}`
  url.searchParams.set('token', token)
  // 同浏览器实例 ID，供后端收口多标签重复会话
  url.searchParams.set('browser_id', getBrowserId())
  return url.toString()
}

function scheduleReconnect() {
  if (!shouldReconnect || !activeToken || !isLeader || reconnectTimer)
    return
  const delay = Math.min(1000 * 2 ** reconnectAttempts, MAX_RECONNECT_DELAY)
  reconnectAttempts++
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    connect()
  }, delay)
}

function connect() {
  if (!shouldReconnect || !activeToken || !isLeader || socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING)
    return

  try {
    socket = new WebSocket(getPresenceUrl(activeToken))
  }
  catch {
    scheduleReconnect()
    return
  }

  socket.onopen = () => {
    connected.value = true
    reconnectAttempts = 0
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
      }
    }
    catch {
      // Presence 不处理其他业务消息。
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
  presenceChannel?.postMessage({ type: 'leader', tabId: ensureTabId() })
  if (!leaderHeartbeatTimer) {
    leaderHeartbeatTimer = setInterval(() => {
      if (isLeader)
        presenceChannel?.postMessage({ type: 'leader', tabId: ensureTabId() })
    }, 2000)
  }
  connect()
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
    const data = event.data as { type?: string, tabId?: string } | null
    if (!data?.type)
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
        presenceChannel?.postMessage({ type: 'leader', tabId: ensureTabId() })
    }
  }

  // 先问有没有现成 leader；没有则自己上位
  presenceChannel.postMessage({ type: 'need-leader', tabId })
  window.setTimeout(() => {
    if (!shouldReconnect)
      return
    if (!isLeader && Date.now() - lastLeaderSeenAt > 2500)
      becomeLeader()
  }, 300)

  if (!leaderCheckTimer) {
    leaderCheckTimer = setInterval(() => {
      if (!shouldReconnect)
        return
      if (!isLeader && Date.now() - lastLeaderSeenAt > 5000)
        becomeLeader()
    }, 2000)
  }
}

function teardownPresenceLeaderElection() {
  resignLeader()
  if (leaderCheckTimer) {
    clearInterval(leaderCheckTimer)
    leaderCheckTimer = null
  }
  if (presenceChannel) {
    presenceChannel.close()
    presenceChannel = null
  }
}

/**
 * 建立 Presence 心跳连接。
 * @param intervalMs 心跳上报周期（毫秒），来自管理端可配置的「在线心跳上报周期」设置，默认25秒。
 */
export function startPresence(token: string, onForceLogout: () => void, intervalMs?: number) {
  if (!token)
    return
  forceLogoutHandler = onForceLogout
  activePingInterval = intervalMs && intervalMs > 0 ? intervalMs : DEFAULT_PING_INTERVAL
  shouldReconnect = true
  // Token 刷新时必须换连接，否则服务端仍会验证旧会话。
  if (activeToken && activeToken !== token)
    closeSocketOnly()
  activeToken = token
  shouldReconnect = true
  setupPresenceLeaderElection()
  if (isLeader)
    connect()
}

export function stopPresence() {
  shouldReconnect = false
  activeToken = ''
  reconnectAttempts = 0
  teardownPresenceLeaderElection()
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
