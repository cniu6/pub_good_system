/**
 * 服务器管理：共享状态与 API
 * 模块级单例，各 Tab 共用同一份状态。
 */
import { onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import type { GoroutineStatsResponse, RuntimeGoroutineInfo } from '@/service/api/admin/debug'
import type { ServerMonitoringStatusResponse } from '@/service/api/admin/settings'
import type { ServerOperationsStatusResponse } from '@/service/api/admin/server'
import { authStorage, parseBooleanSetting, parseNumberSetting } from '@/utils'

export const tabOptions = ['monitor', 'ops', 'debug'] as const

export type DebugProfileType = 'cpu' | 'heap' | 'goroutine' | 'allocs' | 'block' | 'mutex' | 'threadcreate' | 'trace'
export type RuntimeStateCategory = 'running' | 'waiting' | 'channel' | 'syscall' | 'mutex' | 'other'
export type ServiceHealthRow = NonNullable<ServerMonitoringStatusResponse['services']>[number]

export interface PprofResultPanel {
  key: DebugProfileType
  title: string
  text: string
  maxHeight: number
  tags: Array<{ label: string, type?: 'default' | 'error' | 'info' | 'primary' | 'success' | 'warning' }>
}

const activeTab = ref('monitor')
const monitoringLoading = ref(false)
const operationsLoading = ref(false)
const savingRateLimit = ref(false)
const forcingGC = ref(false)
const autoRefresh = ref(false)
const debugAutoRefresh = ref(false)
const debugLoading = reactive<Record<string, boolean>>({
  goroutineStats: false,
  cpu: false,
  heap: false,
  goroutine: false,
  allocs: false,
  block: false,
  mutex: false,
  threadcreate: false,
  trace: false,
  stacks: false,
})

const monitoring = ref<ServerMonitoringStatusResponse | null>(null)
const operations = ref<ServerOperationsStatusResponse | null>(null)
const goroutineStats = ref<GoroutineStatsResponse | null>(null)
const runtimeStacks = ref<RuntimeGoroutineInfo[]>([])
const longRunningStacks = ref<RuntimeGoroutineInfo[]>([])
const potentialLeakStacks = ref<RuntimeGoroutineInfo[]>([])
const runtimeStateSummaryMap = ref<Record<string, number>>({})
const runtimeStacksLoaded = ref(false)
const stackFilterMinWaitMinutes = ref(0)
const cpuSeconds = ref(30)
const traceSeconds = ref(5)
const debugResults = reactive({
  cpuText: '',
  heapText: '',
  goroutineText: '',
  allocsText: '',
  blockText: '',
  mutexText: '',
  threadcreateText: '',
  traceText: '',
})

/** 运维与限流页：仅全局/管理端限流配置（不含日志等其它设置） */
const rateLimitForm = reactive({
  api_rate_limit_enabled: false,
  api_rate_limit_rate: 120,
  api_rate_limit_burst: 240,
  admin_rate_limit_enabled: false,
  admin_rate_limit_rate: 60,
  admin_rate_limit_burst: 120,
})

let refreshTimer: number | null = null
let debugRefreshTimer: number | null = null
let lifecycleBound = false

/** 仅 server-management 使用的整数格式化 */
export function formatInteger(value?: number | null): string {
  const n = Number(value ?? 0)
  if (!Number.isFinite(n))
    return '-'
  return Math.round(n).toLocaleString()
}

/** 纳秒时间戳 → 本地时间 */
export function formatRuntimeTimestamp(value?: number | null): string {
  const n = Number(value ?? 0)
  if (!Number.isFinite(n) || n <= 0)
    return '-'
  return new Date(Math.floor(n / 1e6)).toLocaleString()
}

/** 纳秒时长 */
export function formatNsDuration(value?: number | null): string {
  const n = Number(value ?? 0)
  if (!Number.isFinite(n) || n < 0)
    return '-'
  if (n >= 1e9)
    return `${(n / 1e9).toFixed(2)}s`
  if (n >= 1e6)
    return `${(n / 1e6).toFixed(2)}ms`
  if (n >= 1e3)
    return `${(n / 1e3).toFixed(2)}µs`
  return `${n.toFixed(0)}ns`
}

export function getRuntimeStateCategory(state: string): RuntimeStateCategory {
  const normalized = state.toLowerCase()
  if (normalized === 'running' || normalized === 'runnable')
    return 'running'
  if (normalized.includes('semacquire') || normalized.includes('mutex'))
    return 'mutex'
  if (normalized.includes('syscall') || normalized.includes('io wait') || normalized.includes('poll'))
    return 'syscall'
  if (normalized.includes('chan'))
    return 'channel'
  if (normalized.includes('wait') || normalized.includes('select'))
    return 'waiting'
  return 'other'
}

export function getRuntimeStateTagType(state: string) {
  switch (getRuntimeStateCategory(state)) {
    case 'running':
      return 'success'
    case 'waiting':
      return 'warning'
    case 'channel':
      return 'default'
    case 'syscall':
      return 'info'
    case 'mutex':
      return 'error'
    default:
      return 'default'
  }
}

export function extractProfileMetric(text: string, patterns: RegExp[]) {
  for (const pattern of patterns) {
    const match = text.match(pattern)
    if (match) {
      const value = Number.parseInt(match[1], 10)
      if (Number.isFinite(value))
        return value
    }
  }
  return null
}

export function normalizeActiveTab(value: unknown) {
  return typeof value === 'string' && (tabOptions as readonly string[]).includes(value) ? value : 'monitor'
}

export function useServerManagement() {
  const message = useMessage()
  const { t } = useI18n()

  function getRuntimeStateSummaryLabel(category: RuntimeStateCategory) {
    switch (category) {
      case 'running':
        return t('adminSettings.runtimeStateRunning')
      case 'waiting':
        return t('adminSettings.runtimeStateWaiting')
      case 'channel':
        return t('adminSettings.runtimeStateChannel')
      case 'syscall':
        return t('adminSettings.runtimeStateSyscall')
      case 'mutex':
        return t('adminSettings.runtimeStateMutex')
      default:
        return t('adminSettings.runtimeStateOther')
    }
  }

  function getRuntimeStateSummaryTagType(category: RuntimeStateCategory) {
    switch (category) {
      case 'running':
        return 'success'
      case 'waiting':
        return 'warning'
      case 'channel':
        return 'default'
      case 'syscall':
        return 'info'
      case 'mutex':
        return 'error'
      default:
        return 'default'
    }
  }

  async function loadRateLimitConfig() {
    try {
      const res = await adminApi.settings.list()
      if (!res.isSuccess) {
        message.error(res.message || t('adminServer.loadRateLimitFailed'))
        return
      }
      const categories = res.data?.categories || []
      for (const category of categories) {
        for (const item of category.items) {
          if (item.key in rateLimitForm) {
            ;(rateLimitForm as any)[item.key] = item.type === 'boolean'
              ? parseBooleanSetting(item.value)
              : parseNumberSetting(item.value, (rateLimitForm as any)[item.key])
          }
        }
      }
    }
    catch {
      message.error(t('adminServer.loadRateLimitFailed'))
    }
  }

  function ensureDebugAuthHeaders() {
    const token = authStorage.get('accessToken')
    const headers: Record<string, string> = {}
    if (token)
      headers.Authorization = `Bearer ${token}`
    return headers
  }

  async function fetchDebugText(url: string) {
    const res = await fetch(url, { headers: ensureDebugAuthHeaders() })
    const text = await res.text()
    if (!res.ok)
      throw new Error(text)
    try {
      const payload = JSON.parse(text)
      if (payload && typeof payload === 'object' && 'code' in payload && Number((payload as any).code) !== 200)
        throw new Error(String((payload as any).message || t('adminSettings.loadFailed')))
    }
    catch (error) {
      if (error instanceof SyntaxError)
        return text
      throw error
    }
    return text
  }

  async function downloadDebugBlob(url: string, filename: string) {
    const res = await fetch(url, { headers: ensureDebugAuthHeaders() })
    if (!res.ok)
      throw new Error(await res.text())
    const contentType = res.headers.get('content-type') || ''
    if (contentType.includes('application/json') || contentType.includes('text/plain')) {
      const text = await res.text()
      try {
        const payload = JSON.parse(text)
        if (payload && typeof payload === 'object' && 'code' in payload && Number((payload as any).code) !== 200)
          throw new Error(String((payload as any).message || t('adminSettings.loadFailed')))
      }
      catch (error) {
        if (error instanceof SyntaxError)
          throw new Error(text)
        throw error
      }
      throw new Error(text)
    }
    const blob = await res.blob()
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(objectUrl)
  }

  async function downloadTraceProfile() {
    debugLoading.trace = true
    try {
      const url = await adminApi.debug.traceProfile(traceSeconds.value, true)
      const filename = `pprof-trace-${Date.now()}.out`
      await downloadDebugBlob(url, filename)
      debugResults.traceText = `${t('adminSettings.downloadBinary')}: ${filename}`
      message.success(t('adminServer.debug.captureSuccess'))
    }
    catch (error: any) {
      message.error(`${t('adminSettings.captureFailed')}${error?.message || ''}`)
    }
    finally {
      debugLoading.trace = false
    }
  }

  async function previewTraceProfile() {
    debugLoading.trace = true
    debugResults.traceText = ''
    try {
      const url = await adminApi.debug.traceProfile(traceSeconds.value)
      debugResults.traceText = await fetchDebugText(url)
      message.success(t('adminServer.debug.captureSuccess'))
    }
    catch (error: any) {
      message.error(`${t('adminSettings.captureFailed')}${error?.message || ''}`)
    }
    finally {
      debugLoading.trace = false
    }
  }

  async function loadMonitoring() {
    monitoringLoading.value = true
    try {
      const res = await adminApi.server.monitoring()
      if (!res.isSuccess) {
        message.error(res.message || t('adminSettings.loadMonitoringFailed'))
        return
      }
      monitoring.value = res.data || null
    }
    catch {
      message.error(t('adminSettings.loadMonitoringFailed'))
    }
    finally {
      monitoringLoading.value = false
    }
  }

  async function loadOperations() {
    operationsLoading.value = true
    try {
      const res = await adminApi.server.operations()
      if (!res.isSuccess) {
        message.error(res.message || t('adminServer.loadOperationsFailed'))
        return
      }
      operations.value = res.data || null
    }
    catch {
      message.error(t('adminServer.loadOperationsFailed'))
    }
    finally {
      operationsLoading.value = false
    }
  }

  async function refreshAll() {
    await Promise.all([loadMonitoring(), loadOperations()])
  }

  async function saveRateLimitConfig() {
    savingRateLimit.value = true
    try {
      const payload: Record<string, string> = {
        api_rate_limit_enabled: String(rateLimitForm.api_rate_limit_enabled),
        api_rate_limit_rate: String(Math.max(1, Math.floor(rateLimitForm.api_rate_limit_rate || 120))),
        api_rate_limit_burst: String(Math.max(1, Math.floor(rateLimitForm.api_rate_limit_burst || 240))),
        admin_rate_limit_enabled: String(rateLimitForm.admin_rate_limit_enabled),
        admin_rate_limit_rate: String(Math.max(1, Math.floor(rateLimitForm.admin_rate_limit_rate || 60))),
        admin_rate_limit_burst: String(Math.max(1, Math.floor(rateLimitForm.admin_rate_limit_burst || 120))),
      }
      const res = await adminApi.settings.batchUpdate(payload)
      if (!res.isSuccess)
        throw new Error(res.message || t('adminServer.saveRateLimitFailed'))
      message.success(t('adminServer.saveRateLimitSuccess'))
      await loadOperations()
    }
    catch (error: any) {
      message.error(error?.message || t('adminServer.saveRateLimitFailed'))
    }
    finally {
      savingRateLimit.value = false
    }
  }

  async function loadGoroutineStats() {
    debugLoading.goroutineStats = true
    try {
      const res = await adminApi.debug.goroutineStats({ stacks: false })
      const data = res.data || null
      goroutineStats.value = data
      longRunningStacks.value = data?.long_running || []
      potentialLeakStacks.value = data?.potential_leak_stacks || []
    }
    catch (error: any) {
      message.error(`${t('adminSettings.loadDebugStatsFailed')}${error?.message || ''}`)
    }
    finally {
      debugLoading.goroutineStats = false
    }
  }

  async function handleForceGC() {
    forcingGC.value = true
    try {
      const res = await adminApi.debug.forceGC()
      message.success(res.data?.message || t('adminSettings.gcCompleted', { before: res.data?.goroutines_before || 0, after: res.data?.goroutines_after || 0 }))
      if (runtimeStacksLoaded.value)
        await Promise.all([loadMonitoring(), loadRuntimeStacks()])
      else
        await Promise.all([loadMonitoring(), loadGoroutineStats()])
    }
    catch (error: any) {
      message.error(`${t('adminSettings.operationFailed')}${error?.message || ''}`)
    }
    finally {
      forcingGC.value = false
    }
  }

  async function captureProfile(type: 'cpu' | 'heap' | 'goroutine' | 'allocs' | 'block' | 'mutex' | 'threadcreate') {
    debugLoading[type] = true
    ;(debugResults as any)[`${type}Text`] = ''
    try {
      let url = ''
      switch (type) {
        case 'cpu':
          url = await adminApi.debug.cpuProfile(cpuSeconds.value)
          break
        case 'heap':
          url = await adminApi.debug.heapProfile()
          break
        case 'goroutine':
          url = await adminApi.debug.goroutineProfile(0)
          break
        case 'allocs':
          url = await adminApi.debug.allocsProfile()
          break
        case 'block':
          url = await adminApi.debug.blockProfile()
          break
        case 'mutex':
          url = await adminApi.debug.mutexProfile()
          break
        case 'threadcreate':
          url = await adminApi.debug.threadcreateProfile()
          break
      }
      const text = await fetchDebugText(url)
      ;(debugResults as any)[`${type}Text`] = text
      message.success(t('adminServer.debug.captureSuccess'))
    }
    catch (error: any) {
      message.error(`${t('adminSettings.captureFailed')}${error?.message || ''}`)
    }
    finally {
      debugLoading[type] = false
    }
  }

  async function loadRuntimeStacks() {
    debugLoading.stacks = true
    runtimeStacks.value = []
    longRunningStacks.value = []
    potentialLeakStacks.value = []
    runtimeStateSummaryMap.value = {}
    try {
      const res = await adminApi.debug.goroutineStats({
        stacks: true,
        min_wait_minutes: Math.max(0, Number(stackFilterMinWaitMinutes.value || 0)),
      })
      const data = res.data || null
      goroutineStats.value = data
      runtimeStacks.value = data?.runtime_stacks || []
      longRunningStacks.value = data?.long_running || []
      potentialLeakStacks.value = data?.potential_leak_stacks || []
      runtimeStateSummaryMap.value = data?.runtime_state_summary || {}
      runtimeStacksLoaded.value = true
      const filterMsg = stackFilterMinWaitMinutes.value > 0 ? t('adminSettings.filtered', { minutes: stackFilterMinWaitMinutes.value }) : ''
      message.success(t('adminSettings.stacksLoaded') + filterMsg)
    }
    catch (error: any) {
      message.error(`${t('adminSettings.loadFailed')}${error?.message || ''}`)
    }
    finally {
      debugLoading.stacks = false
    }
  }

  function clearAllPprofResults() {
    debugResults.cpuText = ''
    debugResults.heapText = ''
    debugResults.goroutineText = ''
    debugResults.allocsText = ''
    debugResults.blockText = ''
    debugResults.mutexText = ''
    debugResults.threadcreateText = ''
    debugResults.traceText = ''
    message.success(t('adminSettings.resultsCleared'))
  }

  function clearRuntimeStacks() {
    runtimeStacks.value = []
    longRunningStacks.value = []
    potentialLeakStacks.value = []
    runtimeStateSummaryMap.value = {}
    runtimeStacksLoaded.value = false
    message.success(t('adminSettings.stacksCleared'))
  }

  function toggleDebugAutoRefresh(enabled: boolean) {
    debugAutoRefresh.value = enabled
    if (debugRefreshTimer) {
      window.clearInterval(debugRefreshTimer)
      debugRefreshTimer = null
    }
    if (enabled) {
      debugRefreshTimer = window.setInterval(() => {
        loadGoroutineStats()
      }, 3000)
    }
  }

  function toggleAutoRefresh(enabled: boolean) {
    autoRefresh.value = enabled
    if (refreshTimer) {
      window.clearInterval(refreshTimer)
      refreshTimer = null
    }
    if (enabled) {
      refreshTimer = window.setInterval(() => {
        refreshAll()
      }, 15000)
    }
  }

  async function initPage() {
    await Promise.all([loadRateLimitConfig(), refreshAll(), loadGoroutineStats()])
  }

  /** 仅在外壳挂载一次，避免 Tab 重复注册卸载清理 */
  function bindLifecycleOnce() {
    if (lifecycleBound)
      return
    lifecycleBound = true
    onUnmounted(() => {
      if (refreshTimer)
        window.clearInterval(refreshTimer)
      if (debugRefreshTimer)
        window.clearInterval(debugRefreshTimer)
      lifecycleBound = false
    })
  }

  return {
    activeTab,
    monitoringLoading,
    operationsLoading,
    savingRateLimit,
    forcingGC,
    autoRefresh,
    debugAutoRefresh,
    debugLoading,
    monitoring,
    operations,
    goroutineStats,
    runtimeStacks,
    longRunningStacks,
    potentialLeakStacks,
    runtimeStateSummaryMap,
    runtimeStacksLoaded,
    stackFilterMinWaitMinutes,
    cpuSeconds,
    traceSeconds,
    debugResults,
    rateLimitForm,
    getRuntimeStateSummaryLabel,
    getRuntimeStateSummaryTagType,
    loadRateLimitConfig,
    downloadTraceProfile,
    previewTraceProfile,
    loadMonitoring,
    loadOperations,
    refreshAll,
    saveRateLimitConfig,
    loadGoroutineStats,
    handleForceGC,
    captureProfile,
    loadRuntimeStacks,
    clearAllPprofResults,
    clearRuntimeStacks,
    toggleDebugAutoRefresh,
    toggleAutoRefresh,
    initPage,
    bindLifecycleOnce,
  }
}
