<script setup lang="ts">
  import { computed, h, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
  import { useTableColumnVisibility } from '@/hooks'
  import {
    NAlert,
    NButton,
    NCard,
    NCollapse,
    NCollapseItem,
    NCode,
    NDataTable,
    NDescriptions,
    NDescriptionsItem,
    NDivider,
    NEmpty,
    NGrid,
    NGi,
    NInputNumber,
    NProgress,
    NSpace,
    NStatistic,
    NSwitch,
    NTabPane,
    NTabs,
    NTag,
    NText,
    NTooltip,
    useMessage,
  } from 'naive-ui'
  import type { DataTableColumns } from 'naive-ui'
  import { adminApi } from '@/service/api/admin'
  import type { GoroutineStatsResponse, RuntimeGoroutineInfo } from '@/service/api/admin/debug'
  import type { ServerMonitoringStatusResponse } from '@/service/api/admin/settings'
  import type { BackgroundTaskInfo, DynamicRateLimitSnapshot, ServerOperationsStatusResponse } from '@/service/api/admin/server'
  import { authStorage, parseBooleanSetting, parseNumberSetting } from '@/utils'

  const route = useRoute()
  const router = useRouter()
  const message = useMessage()
  const { t } = useI18n()

  const tabOptions = ['monitor', 'ops', 'debug'] as const
  const activeTab = ref('monitor')
  const monitoringLoading = ref(false)
  const operationsLoading = ref(false)
  const savingRuntime = ref(false)
  const forcingGC = ref(false)
  const runningTaskKey = ref('')
  const autoRefresh = ref(false)
  const restartLoading = ref(false)
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

  type DebugProfileType = 'cpu' | 'heap' | 'goroutine' | 'allocs' | 'block' | 'mutex' | 'threadcreate' | 'trace'
  type RuntimeStateCategory = 'running' | 'waiting' | 'channel' | 'syscall' | 'mutex' | 'other'
  type ServiceHealthRow = NonNullable<ServerMonitoringStatusResponse['services']>[number]

  interface PprofResultPanel {
    key: DebugProfileType
    title: string
    text: string
    maxHeight: number
    tags: Array<{ label: string, type?: 'default' | 'error' | 'info' | 'primary' | 'success' | 'warning' }>
  }

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

  const runtimeForm = reactive({
    api_access_log_enabled: true,
    api_log_query_days: 7,
    api_log_max_count: 1000,
    api_rate_limit_enabled: false,
    api_rate_limit_rate: 120,
    api_rate_limit_burst: 240,
    admin_rate_limit_enabled: false,
    admin_rate_limit_rate: 60,
    admin_rate_limit_burst: 120,
  })

  let refreshTimer: number | null = null
  let debugRefreshTimer: number | null = null

  const taskColumns: DataTableColumns<BackgroundTaskInfo> = [
    { title: t('adminServer.tasks.name'), key: 'label', minWidth: 160 },
    {
      title: t('adminServer.tasks.status'),
      key: 'last_status',
      width: 110,
      render(row) {
        if (row.running) return h(NTag, { type: 'warning', size: 'small' }, () => t('adminServer.running'))
        const tagType = row.last_status === 'success' ? 'success' : row.last_status === 'failed' ? 'error' : 'default'
        const label = row.last_status === 'success' ? t('adminServer.success') : row.last_status === 'failed' ? t('adminServer.failed') : t('adminServer.idle')
        return h(NTag, { type: tagType as any, size: 'small' }, () => label)
      },
    },
    { title: t('adminServer.tasks.interval'), key: 'interval_secs', width: 110, render: row => `${row.interval_secs}s` },
    { title: t('adminServer.tasks.lastRun'), key: 'last_run_time', minWidth: 170, render: row => row.last_run_time || '-' },
    { title: t('adminServer.tasks.nextRun'), key: 'next_run_time', minWidth: 170, render: row => row.next_run_time || '-' },
    { title: t('adminServer.tasks.duration'), key: 'last_duration_ms', width: 110, render: row => `${row.last_duration_ms || 0}ms` },
    { title: t('adminServer.tasks.message'), key: 'last_message', minWidth: 220, ellipsis: { tooltip: true } },
    {
      title: t('adminServer.tasks.actions'),
      key: 'actions',
      width: 110,
      render(row) {
        return h(NButton, {
          size: 'small',
          type: 'primary',
          loading: runningTaskKey.value === row.key,
          disabled: row.running,
          onClick: () => handleRunTask(row.key),
        }, () => t('adminServer.tasks.runNow'))
      },
    },
  ]

 const rateLimitColumns: DataTableColumns<DynamicRateLimitSnapshot> = [
   { title: t('adminServer.rateLimit.name'), key: 'name', minWidth: 130 },
  {
    title: t('adminServer.rateLimit.enabled'),
    key: 'enabled',
    width: 90,
    render(row) {
      return h(NTag, { type: row.enabled ? 'success' : 'default', size: 'small' }, () => row.enabled ? t('common.enable') : t('common.disable'))
    },
  },
  { title: t('adminServer.rateLimit.rate'), key: 'rate', width: 90 },
  { title: t('adminServer.rateLimit.burst'), key: 'burst', width: 90 },
  { title: t('adminServer.rateLimit.allowed'), key: 'allowed_count', width: 100 },
  { title: t('adminServer.rateLimit.blocked'), key: 'blocked_count', width: 100 },
  { title: t('adminServer.rateLimit.visitors'), key: 'active_visitors', width: 110 },
  { title: t('adminServer.rateLimit.lastReload'), key: 'last_config_reload', minWidth: 180, render: row => row.last_config_reload || '-' },
]

const serviceHealthColumns: DataTableColumns<ServiceHealthRow> = [
  { title: t('adminSettings.columnService'), key: 'name', minWidth: 140 },
  {
    title: t('adminSettings.columnStatus'),
    key: 'status',
    width: 110,
    render: (row: ServiceHealthRow) => h(NTag, { type: row.status === 'up' ? 'success' : row.status === 'warning' ? 'warning' : 'error', size: 'small' }, () => row.status === 'up' ? t('adminSettings.statusNormal') : row.status === 'warning' ? t('adminSettings.statusWarning') : t('adminSettings.statusError')),
  },
  { title: t('adminSettings.columnMessage'), key: 'message', minWidth: 220 },
]

const serviceHealthSelectableColumnOptions = computed(() => [
  { key: 'name', label: t('adminSettings.columnService') },
  { key: 'status', label: t('adminSettings.columnStatus') },
  { key: 'message', label: t('adminSettings.columnMessage') },
])

const taskSelectableColumnOptions = computed(() => [
  { key: 'label', label: t('adminServer.tasks.name') },
  { key: 'last_status', label: t('adminServer.tasks.status') },
  { key: 'interval_secs', label: t('adminServer.tasks.interval') },
  { key: 'last_run_time', label: t('adminServer.tasks.lastRun') },
  { key: 'next_run_time', label: t('adminServer.tasks.nextRun') },
  { key: 'last_duration_ms', label: t('adminServer.tasks.duration') },
  { key: 'last_message', label: t('adminServer.tasks.message') },
])

const rateLimitSelectableColumnOptions = computed(() => [
  { key: 'name', label: t('adminServer.rateLimit.name') },
  { key: 'enabled', label: t('adminServer.rateLimit.enabled') },
  { key: 'rate', label: t('adminServer.rateLimit.rate') },
  { key: 'burst', label: t('adminServer.rateLimit.burst') },
  { key: 'allowed_count', label: t('adminServer.rateLimit.allowed') },
  { key: 'blocked_count', label: t('adminServer.rateLimit.blocked') },
  { key: 'active_visitors', label: t('adminServer.rateLimit.visitors') },
  { key: 'last_config_reload', label: t('adminServer.rateLimit.lastReload') },
])

const runtimeStackSelectableColumnOptions = computed(() => [
  { key: 'id', label: '#' },
  { key: 'function', label: t('adminSettings.stackFunction') },
  { key: 'state', label: t('adminSettings.columnStatus') },
  { key: 'wait_time', label: t('adminSettings.waitTime') },
  { key: 'created_by', label: t('adminSettings.createdBy') },
])

const {
  columnOptions: serviceHealthColumnOptions,
  selectedColumnKeys: serviceHealthSelectedColumnKeys,
  visibleColumns: serviceHealthVisibleColumns,
  visibleColumnCount: serviceHealthVisibleColumnCount,
  totalColumnCount: serviceHealthTotalColumnCount,
  tableScrollX: serviceHealthTableScrollX,
  resetSelectedColumns: resetServiceHealthSelectedColumns,
} = useTableColumnVisibility<ServiceHealthRow>({
  storageKey: 'admin-server-service-health',
  columns: serviceHealthColumns,
  options: serviceHealthSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 620,
})

const {
  columnOptions: taskColumnOptions,
  selectedColumnKeys: taskSelectedColumnKeys,
  visibleColumns: taskVisibleColumns,
  visibleColumnCount: taskVisibleColumnCount,
  totalColumnCount: taskTotalColumnCount,
  tableScrollX: taskTableScrollX,
  resetSelectedColumns: resetTaskSelectedColumns,
} = useTableColumnVisibility<BackgroundTaskInfo>({
  storageKey: 'admin-server-tasks',
  columns: taskColumns,
  options: taskSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 960,
})

const {
  columnOptions: rateLimitColumnOptions,
  selectedColumnKeys: rateLimitSelectedColumnKeys,
  visibleColumns: rateLimitVisibleColumns,
  visibleColumnCount: rateLimitVisibleColumnCount,
  totalColumnCount: rateLimitTotalColumnCount,
  tableScrollX: rateLimitTableScrollX,
  resetSelectedColumns: resetRateLimitSelectedColumns,
} = useTableColumnVisibility<DynamicRateLimitSnapshot>({
  storageKey: 'admin-server-rate-limits',
  columns: rateLimitColumns,
  options: rateLimitSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 920,
})

const {
  columnOptions: runtimeStackColumnOptions,
  selectedColumnKeys: runtimeStackSelectedColumnKeys,
  visibleColumns: runtimeStackVisibleColumns,
  visibleColumnCount: runtimeStackVisibleColumnCount,
  totalColumnCount: runtimeStackTotalColumnCount,
  tableScrollX: runtimeStackTableScrollX,
  resetSelectedColumns: resetRuntimeStackSelectedColumns,
} = useTableColumnVisibility<RuntimeGoroutineInfo>({
  storageKey: 'admin-server-runtime-stack-preview',
  columns: computed(() => runtimeStackColumns),
  options: runtimeStackSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})

const processInfo = computed(() => monitoring.value?.process)
const metricInfo = computed(() => monitoring.value?.metrics)
const serviceList = computed(() => monitoring.value?.services || [])
const operationTasks = computed(() => operations.value?.tasks || [])
const rateLimits = computed(() => operations.value?.rate_limits || [])
const hasAnyPprofResult = computed(() => Object.values(debugResults).some(value => Boolean(value)))
const runtimeStateSummary = computed(() => {
  const orderedKeys: RuntimeStateCategory[] = ['running', 'waiting', 'channel', 'syscall', 'mutex', 'other']
  return orderedKeys
    .map((key) => ({
      key,
      count: Number(runtimeStateSummaryMap.value[key] || 0),
      label: getRuntimeStateSummaryLabel(key),
      type: getRuntimeStateSummaryTagType(key),
    }))
    .filter(item => item.count > 0)
})
const potentialLeakIdSet = computed(() => new Set(potentialLeakStacks.value.map(stack => stack.id)))
const heapProfileStats = computed(() => {
  if (!debugResults.heapText) {
    return null
  }

  const alloc = extractProfileMetric(debugResults.heapText, [
    /#\s*Alloc\s*=\s*(\d+)/i,
    /heap profile:\s*\d+:\s*(\d+)/i,
  ])
  const objects = extractProfileMetric(debugResults.heapText, [
    /#\s*HeapObjects\s*=\s*(\d+)/i,
    /#\s*objects\s*=\s*(\d+)/i,
  ])

  if (alloc == null && objects == null) {
    return null
  }

  return {
    alloc: alloc || 0,
    objects: objects || 0,
  }
})
const goroutineProfileCount = computed(() => {
  return (debugResults.goroutineText.match(/^goroutine\s+\d+\s+\[/gm) || []).length
})
const pprofResultPanels = computed<PprofResultPanel[]>(() => {
  const panels: PprofResultPanel[] = []

  if (debugResults.cpuText) {
    panels.push({
      key: 'cpu',
      title: t('adminSettings.cpuProfileResult'),
      text: debugResults.cpuText,
      maxHeight: 500,
      tags: [{ label: `${cpuSeconds.value}s`, type: 'success' }],
    })
  }

  if (debugResults.heapText) {
    const tags: PprofResultPanel['tags'] = []
    if (heapProfileStats.value) {
      if (heapProfileStats.value.alloc > 0) {
        tags.push({ label: formatBytes(heapProfileStats.value.alloc), type: 'info' })
      }
      if (heapProfileStats.value.objects > 0) {
        tags.push({ label: `${formatInteger(heapProfileStats.value.objects)} ${t('adminSettings.heapObjects')}` })
      }
    }

    panels.push({
      key: 'heap',
      title: t('adminSettings.heapProfileResult'),
      text: debugResults.heapText,
      maxHeight: 420,
      tags,
    })
  }

  if (debugResults.goroutineText) {
    panels.push({
      key: 'goroutine',
      title: t('adminSettings.goroutineProfileResult'),
      text: debugResults.goroutineText,
      maxHeight: 500,
      tags: [{ label: `${formatInteger(goroutineProfileCount.value)} ${t('adminSettings.goroutines')}`, type: 'info' }],
    })
  }

  if (debugResults.allocsText) {
    panels.push({
      key: 'allocs',
      title: t('adminSettings.allocsProfileResult'),
      text: debugResults.allocsText,
      maxHeight: 420,
      tags: [],
    })
  }

  if (debugResults.blockText) {
    panels.push({
      key: 'block',
      title: t('adminSettings.blockProfileResult'),
      text: debugResults.blockText,
      maxHeight: 420,
      tags: [],
    })
  }

  if (debugResults.mutexText) {
    panels.push({
      key: 'mutex',
      title: t('adminSettings.mutexProfileResult'),
      text: debugResults.mutexText,
      maxHeight: 420,
      tags: [],
    })
  }

  if (debugResults.threadcreateText) {
    panels.push({
      key: 'threadcreate',
      title: t('adminSettings.threadCreateProfileResult'),
      text: debugResults.threadcreateText,
      maxHeight: 420,
      tags: [],
    })
  }

  if (debugResults.traceText) {
    panels.push({
      key: 'trace',
      title: t('adminSettings.traceProfileResult'),
      text: debugResults.traceText,
      maxHeight: 360,
      tags: [],
    })
  }

  return panels
})

const runtimeStackColumns: DataTableColumns<RuntimeGoroutineInfo> = [
  {
    title: '#',
    key: 'id',
    width: 72,
    render(row) {
      return h(NText, { code: true }, () => `#${row.id}`)
    },
  },
  {
    title: t('adminSettings.stackFunction'),
    key: 'function',
    minWidth: 240,
    ellipsis: { tooltip: true },
  },
  {
    title: t('adminSettings.columnStatus'),
    key: 'state',
    width: 150,
    render(row) {
      return h(NTag, { type: getRuntimeStateTagType(row.state) as any, size: 'small' }, () => row.state)
    },
  },
  {
    title: t('adminSettings.waitTime'),
    key: 'wait_time',
    width: 140,
    render(row) {
      return row.wait_time || '-'
    },
  },
  {
    title: t('adminSettings.createdBy'),
    key: 'created_by',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render(row) {
      return row.created_by || '-'
    },
  },
]

function getRuntimeStateCategory(state: string): RuntimeStateCategory {
  const normalized = state.toLowerCase()
  if (normalized === 'running' || normalized === 'runnable') {
    return 'running'
  }
  if (normalized.includes('semacquire') || normalized.includes('mutex')) {
    return 'mutex'
  }
  if (normalized.includes('syscall') || normalized.includes('io wait') || normalized.includes('poll')) {
    return 'syscall'
  }
  if (normalized.includes('chan')) {
    return 'channel'
  }
  if (normalized.includes('wait') || normalized.includes('select')) {
    return 'waiting'
  }
  return 'other'
}

function getRuntimeStateTagType(state: string) {
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

function extractProfileMetric(text: string, patterns: RegExp[]) {
  for (const pattern of patterns) {
    const match = text.match(pattern)
    if (match) {
      const value = Number.parseInt(match[1], 10)
      if (Number.isFinite(value)) {
        return value
      }
    }
  }
  return null
}

function normalizePercent(value: number) {
  if (!Number.isFinite(value) || value < 0) {
    return 0
  }
  if (value > 100) {
    return 100
  }
  return Number(value.toFixed(2))
}

function formatInteger(value: number): string {
  if (!Number.isFinite(value)) {
    return '-'
  }
  return Math.round(value).toLocaleString()
}

function formatPercent(value: number): string {
  return `${normalizePercent(value).toFixed(2)}%`
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let idx = 0
  while (size >= 1024 && idx < units.length - 1) {
    size /= 1024
    idx++
  }
  return `${size.toFixed(2)} ${units[idx]}`
}

function formatStorageFromMB(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  const gb = value / 1024
  if (gb >= 1024) {
    return `${(gb / 1024).toFixed(2)} TB`
  }
  if (gb >= 1) {
    return `${gb.toFixed(2)} GB`
  }
  return `${value.toFixed(2)} MB`
}

function formatStorageFromGB(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(2)} TB`
  }
  return `${value.toFixed(2)} GB`
}

function formatRuntimeTimestamp(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '-'
  }
  return new Date(Math.floor(value / 1e6)).toLocaleString()
}

function formatNsDuration(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  if (value >= 1e9) {
    return `${(value / 1e9).toFixed(2)}s`
  }
  if (value >= 1e6) {
    return `${(value / 1e6).toFixed(2)}ms`
  }
  if (value >= 1e3) {
    return `${(value / 1e3).toFixed(2)}µs`
  }
  return `${value.toFixed(0)}ns`
}

function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return '-'
  }
  const day = Math.floor(seconds / 86400)
  const hour = Math.floor((seconds % 86400) / 3600)
  const minute = Math.floor((seconds % 3600) / 60)
  const second = Math.floor(seconds % 60)
  return t('adminSettings.uptimePreciseFormat', { day, hour, minute, second })
}

async function loadRuntimeConfig() {
  try {
    const res = await adminApi.settings.list()
    const categories = res.data?.categories || []
    for (const category of categories) {
      for (const item of category.items) {
        if (item.key in runtimeForm) {
          ;(runtimeForm as any)[item.key] = item.type === 'boolean'
            ? parseBooleanSetting(item.value)
            : parseNumberSetting(item.value, (runtimeForm as any)[item.key])
        }
      }
    }
  }
  catch {
    message.error(t('adminServer.loadRuntimeFailed'))
  }
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

async function saveRuntimeConfig() {
  savingRuntime.value = true
  try {
    const payload: Record<string, string> = {
      api_access_log_enabled: String(runtimeForm.api_access_log_enabled),
      api_log_query_days: String(Math.max(1, Math.floor(runtimeForm.api_log_query_days || 7))),
      api_log_max_count: String(Math.max(100, Math.floor(runtimeForm.api_log_max_count || 1000))),
      api_rate_limit_enabled: String(runtimeForm.api_rate_limit_enabled),
      api_rate_limit_rate: String(Math.max(1, Math.floor(runtimeForm.api_rate_limit_rate || 120))),
      api_rate_limit_burst: String(Math.max(1, Math.floor(runtimeForm.api_rate_limit_burst || 240))),
      admin_rate_limit_enabled: String(runtimeForm.admin_rate_limit_enabled),
      admin_rate_limit_rate: String(Math.max(1, Math.floor(runtimeForm.admin_rate_limit_rate || 60))),
      admin_rate_limit_burst: String(Math.max(1, Math.floor(runtimeForm.admin_rate_limit_burst || 120))),
    }
    const res = await adminApi.settings.batchUpdate(payload)
    if (!res.isSuccess) throw new Error(res.message || t('adminServer.saveRuntimeFailed'))
    message.success(t('adminServer.saveRuntimeSuccess'))
    await loadOperations()
  }
  catch (error: any) {
    message.error(error?.message || t('adminServer.saveRuntimeFailed'))
  }
  finally {
    savingRuntime.value = false
  }
}

async function handleRunTask(key: string) {
  runningTaskKey.value = key
  try {
    const res = await adminApi.server.runTask(key)
    message.success(res.data?.message || t('adminServer.tasks.runSuccess'))
    await loadOperations()
  }
  catch {
    message.error(t('adminServer.tasks.runFailed'))
  }
  finally {
    runningTaskKey.value = ''
  }
}

async function handleRestartBackend() {
  restartLoading.value = true
  try {
    await adminApi.settings.restartBackend()
    message.success(t('adminSettings.restartBackendRequested'))
  }
  catch {
    message.error(t('adminSettings.restartBackendFailed'))
  }
  finally {
    restartLoading.value = false
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
    if (payload && typeof payload === 'object' && 'code' in payload && Number((payload as any).code) !== 200) {
      throw new Error(String((payload as any).message || t('adminSettings.loadFailed')))
    }
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
  if (!res.ok) {
    throw new Error(await res.text())
  }
  const contentType = res.headers.get('content-type') || ''
  if (contentType.includes('application/json') || contentType.includes('text/plain')) {
    const text = await res.text()
    try {
      const payload = JSON.parse(text)
      if (payload && typeof payload === 'object' && 'code' in payload && Number((payload as any).code) !== 200) {
        throw new Error(String((payload as any).message || t('adminSettings.loadFailed')))
      }
    }
    catch (error) {
      if (error instanceof SyntaxError) {
        throw new Error(text)
      }
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
    if (runtimeStacksLoaded.value) {
      await Promise.all([loadMonitoring(), loadRuntimeStacks()])
    }
    else {
      await Promise.all([loadMonitoring(), loadGoroutineStats()])
    }
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

function normalizeActiveTab(value: unknown) {
  return typeof value === 'string' && tabOptions.includes(value as any) ? value : 'monitor'
}

watch(
  () => route.query.tab,
  (value) => {
    activeTab.value = normalizeActiveTab(value)
  },
  { immediate: true },
)

watch(activeTab, (value) => {
  const nextTab = normalizeActiveTab(value)
  if (route.query.tab === nextTab)
    return
  router.replace({ query: { ...route.query, tab: nextTab } })
})

onMounted(async () => {
  await Promise.all([loadRuntimeConfig(), refreshAll(), loadGoroutineStats()])
})

onUnmounted(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
  }
  if (debugRefreshTimer) {
    window.clearInterval(debugRefreshTimer)
  }
})
</script>

<template>
  <NTabs v-model:value="activeTab" type="line" animated>
    <NTabPane name="monitor" :tab="t('adminSettings.systemMonitor')">
      <NSpace vertical :size="16">
        <NCard>
          <template #header>
            <NSpace align="center" justify="space-between" style="width: 100%;">
              <NText strong>{{ t('adminSettings.realtimeMonitor') }}</NText>
              <NSpace align="center">
                <NText depth="3">{{ t('adminSettings.autoRefresh') }}</NText>
                <NSwitch :value="autoRefresh" @update:value="toggleAutoRefresh" />
                <NButton size="small" type="primary" :loading="monitoringLoading || operationsLoading" @click="refreshAll">{{ t('adminSettings.refresh') }}</NButton>
              </NSpace>
            </NSpace>
          </template>

          <NGrid :x-gap="12" :y-gap="12" cols="2 s:2 m:4 l:4" responsive="screen">
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.cpu')"><template #default>{{ formatPercent(metricInfo?.cpu.usage_percent || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ t('adminSettings.cpuCores', { count: metricInfo?.cpu.core_count || 0 }) }}</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.systemMemory')"><template #default>{{ formatPercent(metricInfo?.memory.used_percent || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ formatStorageFromMB(metricInfo?.memory.used_mb || 0) }}/{{ formatStorageFromMB(metricInfo?.memory.total_mb || 0) }}</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.swap')"><template #default>{{ formatPercent(metricInfo?.swap.used_percent || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ formatStorageFromMB(metricInfo?.swap.used_mb || 0) }}/{{ formatStorageFromMB(metricInfo?.swap.total_mb || 0) }}</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.diskUsage')"><template #default>{{ formatPercent(metricInfo?.disk.used_percent || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ formatStorageFromGB(metricInfo?.disk.used_gb || 0) }}/{{ formatStorageFromGB(metricInfo?.disk.total_gb || 0) }}</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.processMemory')"><template #default>{{ formatStorageFromMB(processInfo?.process_rss_mb || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">CPU {{ Number((processInfo?.process_cpu || 0).toFixed(2)) }}%</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.goHeap')"><template #default>{{ formatStorageFromMB(processInfo?.heap_alloc_mb || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">sys {{ formatStorageFromMB(processInfo?.memory_sys_mb || 0) }}</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.goroutines')" :value="formatInteger(processInfo?.goroutines || 0)"><template #suffix><NText depth="3" style="font-size: 10px">GC {{ formatInteger(processInfo?.gc_count || 0) }}</NText></template></NStatistic></NCard></NGi>
            <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.uptime')"><template #default>{{ formatUptime(monitoring?.uptime_seconds || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ monitoring?.generated_at || '-' }}</NText></template></NStatistic></NCard></NGi>
          </NGrid>

          <NDivider />

          <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 m:2 l:2" responsive="screen">
            <NGi>
              <NCard size="small" :title="t('adminSettings.systemInfo')">
                <NDescriptions bordered :column="2" label-placement="left">
                  <NDescriptionsItem :label="t('adminSettings.appName')">{{ monitoring?.app?.name || '-' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="t('adminSettings.systemVersion')">{{ monitoring?.app?.go_version || '-' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="t('adminSettings.appMode')">{{ monitoring?.app?.mode || '-' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="t('adminSettings.port')">{{ monitoring?.app?.port || '-' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="t('adminSettings.pid')">{{ processInfo?.pid || '-' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="t('adminSettings.lastRefreshed')">{{ monitoring?.generated_at || '-' }}</NDescriptionsItem>
                </NDescriptions>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.network')">
                <NSpace vertical size="small">
                  <NStatistic :label="t('adminSettings.network')"><template #default>{{ formatBytes((metricInfo?.network.bytes_sent || 0) + (metricInfo?.network.bytes_recv || 0)) }}</template></NStatistic>
                  <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.upload') }}</NText><NText>{{ formatBytes(metricInfo?.network.bytes_sent || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.download') }}</NText><NText>{{ formatBytes(metricInfo?.network.bytes_recv || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.uploadPackets') }}</NText><NText>{{ formatInteger(metricInfo?.network.packets_sent || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.downloadPackets') }}</NText><NText>{{ formatInteger(metricInfo?.network.packets_recv || 0) }}</NText></NSpace>
                </NSpace>
              </NCard>
            </NGi>
          </NGrid>

          <NCard size="small" :title="t('adminSettings.memoryDetails')" style="margin-top: 12px;">
            <NGrid :x-gap="10" :y-gap="10" cols="1 s:2 m:4 l:4" responsive="screen">
              <NGi><NStatistic :label="t('adminSettings.goMemoryAlloc')"><template #default>{{ formatStorageFromMB(processInfo?.memory_alloc_mb || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.goMemorySys')"><template #default>{{ formatStorageFromMB(processInfo?.memory_sys_mb || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.heapAlloc')"><template #default>{{ formatStorageFromMB(processInfo?.heap_alloc_mb || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.heapInUse')"><template #default>{{ formatStorageFromMB(processInfo?.heap_inuse_mb || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.heapIdle')"><template #default>{{ formatStorageFromMB(processInfo?.heap_idle_mb || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.stackInUse')"><template #default>{{ formatStorageFromMB(processInfo?.stack_inuse_mb || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.gcCount')"><template #default>{{ formatInteger(processInfo?.gc_count || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.gcCPU')"><template #default>{{ Number(((processInfo?.gc_cpu_fraction || 0) * 100).toFixed(4)) }}%</template></NStatistic></NGi>
            </NGrid>
          </NCard>
        </NCard>

        <NCard :title="t('adminSettings.serviceHealthSnapshot')">
          <template #header-extra>
            <TableColumnSelector
              v-model="serviceHealthSelectedColumnKeys"
              :options="serviceHealthColumnOptions"
              :visible-count="serviceHealthVisibleColumnCount"
              :total-count="serviceHealthTotalColumnCount"
              :button-label="t('common.showFields')"
              :title="t('common.visibleFields')"
              :hint="t('common.columnVisibilityHint')"
              :reset-label="t('common.restoreDefaultFields')"
              @reset="resetServiceHealthSelectedColumns"
            />
          </template>
          <NDataTable :columns="serviceHealthVisibleColumns" :data="serviceList" :pagination="false" :scroll-x="serviceHealthTableScrollX" />
        </NCard>
      </NSpace>
    </NTabPane>

    <NTabPane name="ops" :tab="t('adminServer.operationsTab')">
      <NSpace vertical :size="16">
        <NCard :title="t('adminServer.tasks.title')">
          <template #header-extra>
            <TableColumnSelector
              v-model="taskSelectedColumnKeys"
              :options="taskColumnOptions"
              :visible-count="taskVisibleColumnCount"
              :total-count="taskTotalColumnCount"
              :button-label="t('common.showFields')"
              :title="t('common.visibleFields')"
              :hint="t('common.columnVisibilityHint')"
              :reset-label="t('common.restoreDefaultFields')"
              @reset="resetTaskSelectedColumns"
            />
          </template>
          <NDataTable :columns="taskVisibleColumns" :data="operationTasks" :loading="operationsLoading" :pagination="false" :scroll-x="taskTableScrollX" />
        </NCard>

        <NCard :title="t('adminServer.rateLimit.title')">
          <template #header-extra>
            <TableColumnSelector
              v-model="rateLimitSelectedColumnKeys"
              :options="rateLimitColumnOptions"
              :visible-count="rateLimitVisibleColumnCount"
              :total-count="rateLimitTotalColumnCount"
              :button-label="t('common.showFields')"
              :title="t('common.visibleFields')"
              :hint="t('common.columnVisibilityHint')"
              :reset-label="t('common.restoreDefaultFields')"
              @reset="resetRateLimitSelectedColumns"
            />
          </template>
          <NDataTable :columns="rateLimitVisibleColumns" :data="rateLimits" :loading="operationsLoading" :pagination="false" :scroll-x="rateLimitTableScrollX" />
        </NCard>

        <NCard :title="t('adminServer.runtimeConfig.title')">
          <NGrid cols="2" :x-gap="24" :y-gap="18">
            <NGi>
              <NSpace vertical>
                <NSpace align="center" justify="space-between"><NText strong>{{ t('adminServer.runtimeConfig.apiLog') }}</NText><NSwitch v-model:value="runtimeForm.api_access_log_enabled" /></NSpace>
                <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.runtimeConfig.queryDays') }}</NText><NInputNumber v-model:value="runtimeForm.api_log_query_days" :min="1" :max="365" /></NSpace>
                <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.runtimeConfig.maxCount') }}</NText><NInputNumber v-model:value="runtimeForm.api_log_max_count" :min="100" :max="200000" /></NSpace>
              </NSpace>
            </NGi>
            <NGi>
              <NSpace vertical>
                <NSpace align="center" justify="space-between"><NText strong>{{ t('adminServer.runtimeConfig.globalRateLimit') }}</NText><NSwitch v-model:value="runtimeForm.api_rate_limit_enabled" /></NSpace>
                <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.rate') }}</NText><NInputNumber v-model:value="runtimeForm.api_rate_limit_rate" :min="1" :max="10000" /></NSpace>
                <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.burst') }}</NText><NInputNumber v-model:value="runtimeForm.api_rate_limit_burst" :min="1" :max="20000" /></NSpace>
              </NSpace>
            </NGi>
            <NGi>
              <NSpace vertical>
                <NSpace align="center" justify="space-between"><NText strong>{{ t('adminServer.runtimeConfig.adminRateLimit') }}</NText><NSwitch v-model:value="runtimeForm.admin_rate_limit_enabled" /></NSpace>
                <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.rate') }}</NText><NInputNumber v-model:value="runtimeForm.admin_rate_limit_rate" :min="1" :max="10000" /></NSpace>
                <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.burst') }}</NText><NInputNumber v-model:value="runtimeForm.admin_rate_limit_burst" :min="1" :max="20000" /></NSpace>
              </NSpace>
            </NGi>
            <NGi>
              <NSpace vertical>
                <NText depth="3">{{ t('adminServer.runtimeConfig.hint') }}</NText>
                <NSpace>
                  <NButton type="primary" :loading="savingRuntime" @click="saveRuntimeConfig">{{ t('adminServer.runtimeConfig.save') }}</NButton>
                  <NButton :loading="restartLoading" @click="handleRestartBackend">{{ t('adminServer.runtimeConfig.restartBackend') }}</NButton>
                </NSpace>
              </NSpace>
            </NGi>
          </NGrid>
        </NCard>
      </NSpace>
    </NTabPane>

    <NTabPane name="debug" :tab="t('adminSettings.debugTools')">
      <NSpace vertical :size="16">
        <NCard :title="t('adminSettings.systemOverview')" size="small">
          <template #header-extra>
            <NSpace>
              <NButton size="small" :type="debugAutoRefresh ? 'primary' : 'default'" @click="toggleDebugAutoRefresh(!debugAutoRefresh)">
                {{ debugAutoRefresh ? t('adminSettings.stopRefresh') : t('adminSettings.autoRefresh') }}
              </NButton>
              <NButton size="small" :loading="debugLoading.goroutineStats" @click="loadGoroutineStats">{{ t('adminSettings.refresh') }}</NButton>
              <NButton size="small" type="warning" :loading="forcingGC" @click="handleForceGC">{{ t('adminSettings.forceGC') }}</NButton>
            </NSpace>
          </template>
          <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 m:2 l:2" responsive="screen">
            <NGi>
              <NCard size="small" :title="t('adminSettings.processResources')">
                <NSpace vertical size="small">
                  <div>
                    <NSpace justify="space-between">
                      <NText>{{ t('adminSettings.cpu') }}</NText>
                      <NText>{{ Number((processInfo?.process_cpu || 0).toFixed(1)) }}%</NText>
                    </NSpace>
                    <NProgress
                      type="line"
                      :percentage="normalizePercent(processInfo?.process_cpu || 0)"
                      :status="(processInfo?.process_cpu ?? 0) > 80 ? 'error' : 'success'"
                      :show-indicator="false"
                      style="margin-top: 4px; transform: scaleY(0.7); transform-origin: center;"
                    />
                  </div>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.memory') }}</NText><NText>{{ formatStorageFromMB(processInfo?.process_rss_mb || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.goroutines') }}</NText><NText>{{ formatInteger(processInfo?.goroutines || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.pid') }}</NText><NText>{{ processInfo?.pid || '-' }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.gomaxprocs') }}</NText><NText>{{ formatInteger(goroutineStats?.gomaxprocs || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.numCPU') }}</NText><NText>{{ formatInteger(goroutineStats?.num_cpu || 0) }}</NText></NSpace>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.goroutineStats')">
                <NSpace vertical size="small">
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.runtimeTotal') }}</NText><NText>{{ formatInteger(goroutineStats?.total_count || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.tracked') }}</NText><NText>{{ formatInteger(goroutineStats?.tracked_count || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.heapAlloc') }}</NText><NText>{{ formatBytes(goroutineStats?.mem_stats?.heap_alloc || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.heapInUse') }}</NText><NText>{{ formatBytes(goroutineStats?.mem_stats?.heap_inuse || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.gcCount') }}</NText><NText>{{ formatInteger(goroutineStats?.mem_stats?.num_gc || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.totalAlloc') }}</NText><NText>{{ formatBytes(goroutineStats?.mem_stats?.total_alloc || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.heapObjects') }}</NText><NText>{{ formatInteger(goroutineStats?.mem_stats?.heap_objects || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.numCgoCall') }}</NText><NText>{{ formatInteger(goroutineStats?.num_cgo_call || 0) }}</NText></NSpace>
                  <NSpace justify="space-between"><NText>{{ t('adminSettings.potentialLeaks') }}</NText><NText type="error">{{ formatInteger(goroutineStats?.potential_leaks || 0) }}</NText></NSpace>
                </NSpace>
              </NCard>
            </NGi>
          </NGrid>

          <NCard size="small" :title="t('adminSettings.memoryDetails')" style="margin-top: 12px;">
            <NGrid :x-gap="10" :y-gap="10" cols="1 s:2 m:4 l:4" responsive="screen">
              <NGi><NStatistic :label="t('adminSettings.heapSys')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.heap_sys || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.heapIdle')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.heap_idle || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.heapReleased')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.heap_released || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.stackInUse')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.stack_inuse || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.stackSys')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.stack_sys || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.runtimeSys')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.sys || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.mallocs')"><template #default>{{ formatInteger(goroutineStats?.mem_stats?.mallocs || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.frees')"><template #default>{{ formatInteger(goroutineStats?.mem_stats?.frees || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.nextGC')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.next_gc || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.lastGC')"><template #default>{{ formatRuntimeTimestamp(goroutineStats?.mem_stats?.last_gc || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.pauseTotal')"><template #default>{{ formatNsDuration(goroutineStats?.mem_stats?.pause_total_ns || 0) }}</template></NStatistic></NGi>
              <NGi><NStatistic :label="t('adminSettings.forcedGC')"><template #default>{{ formatInteger(goroutineStats?.mem_stats?.num_forced_gc || 0) }}</template></NStatistic></NGi>
            </NGrid>
          </NCard>
        </NCard>

        <NCard :title="t('adminSettings.pprofTitle')" size="small">
          <template #header-extra>
            <NButton size="small" @click="clearAllPprofResults">{{ t('adminSettings.clearResults') }}</NButton>
          </template>

          <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 m:3 l:3" responsive="screen">
            <NGi>
              <NCard size="small" :title="t('adminSettings.cpuProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.cpuProfileDesc') }}</NText>
                  <NSpace>
                    <NInputNumber v-model:value="cpuSeconds" :min="5" :max="120" size="small" style="width: 90px" />
                    <NButton size="small" type="primary" :loading="debugLoading.cpu" @click="captureProfile('cpu')">{{ t('adminSettings.capture') }}</NButton>
                  </NSpace>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.heapProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.heapProfileDesc') }}</NText>
                  <NButton size="small" type="primary" :loading="debugLoading.heap" @click="captureProfile('heap')">{{ t('adminSettings.capture') }}</NButton>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.goroutineProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.goroutineProfileDesc') }}</NText>
                  <NButton size="small" type="primary" :loading="debugLoading.goroutine" @click="captureProfile('goroutine')">{{ t('adminSettings.capture') }}</NButton>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.allocsProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.allocsProfileDesc') }}</NText>
                  <NButton size="small" type="primary" :loading="debugLoading.allocs" @click="captureProfile('allocs')">{{ t('adminSettings.capture') }}</NButton>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.blockProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.blockProfileDesc') }}</NText>
                  <NButton size="small" type="primary" :loading="debugLoading.block" @click="captureProfile('block')">{{ t('adminSettings.capture') }}</NButton>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.mutexProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.mutexProfileDesc') }}</NText>
                  <NButton size="small" type="primary" :loading="debugLoading.mutex" @click="captureProfile('mutex')">{{ t('adminSettings.capture') }}</NButton>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.threadCreateProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.threadCreateProfileDesc') }}</NText>
                  <NButton size="small" type="primary" :loading="debugLoading.threadcreate" @click="captureProfile('threadcreate')">{{ t('adminSettings.capture') }}</NButton>
                </NSpace>
              </NCard>
            </NGi>
            <NGi>
              <NCard size="small" :title="t('adminSettings.traceProfile')">
                <NSpace vertical size="small">
                  <NText depth="3">{{ t('adminSettings.traceProfileDesc') }}</NText>
                  <NSpace>
                    <NInputNumber v-model:value="traceSeconds" :min="1" :max="30" size="small" style="width: 90px" />
                    <NButton size="small" type="primary" :loading="debugLoading.trace" @click="previewTraceProfile">{{ t('common.preview') }}</NButton>
                    <NButton size="small" :loading="debugLoading.trace" @click="downloadTraceProfile">{{ t('adminSettings.downloadBinary') }}</NButton>
                  </NSpace>
                </NSpace>
              </NCard>
            </NGi>
          </NGrid>

          <NEmpty v-if="!hasAnyPprofResult" :description="t('adminSettings.clickToCapture')" style="margin-top: 16px" />
          <NCollapse v-else style="margin-top: 16px">
            <NCollapseItem
              v-for="panel in pprofResultPanels"
              :key="panel.key"
              :title="panel.title"
              :name="panel.key"
            >
              <template #header-extra>
                <NSpace v-if="panel.tags.length > 0" size="small">
                  <NTag v-for="tag in panel.tags" :key="`${panel.key}-${tag.label}`" :type="tag.type as any" size="small">
                    {{ tag.label }}
                  </NTag>
                </NSpace>
              </template>
              <NCode
                :code="panel.text"
                language="text"
                word-wrap
                :style="{ maxHeight: `${panel.maxHeight}px`, overflow: 'auto' }"
              />
            </NCollapseItem>
          </NCollapse>
        </NCard>

        <NCard :title="t('adminSettings.runtimeStacks')" size="small">
          <template #header-extra>
            <NSpace>
              <TableColumnSelector
                v-model="runtimeStackSelectedColumnKeys"
                :options="runtimeStackColumnOptions"
                :visible-count="runtimeStackVisibleColumnCount"
                :total-count="runtimeStackTotalColumnCount"
                :button-label="t('common.showFields')"
                :title="t('common.visibleFields')"
                :hint="t('common.columnVisibilityHint')"
                :reset-label="t('common.restoreDefaultFields')"
                @reset="resetRuntimeStackSelectedColumns"
              />
              <NTooltip trigger="hover">
                <template #trigger>
                  <NInputNumber v-model:value="stackFilterMinWaitMinutes" :min="0" size="small" style="width: 140px" :placeholder="t('adminSettings.minWaitMinutes')" />
                </template>
                {{ t('adminSettings.filterTooltip') }}
              </NTooltip>
              <NButton size="small" :loading="debugLoading.stacks" @click="loadRuntimeStacks">{{ t('adminSettings.loadStacks') }}</NButton>
              <NButton size="small" @click="clearRuntimeStacks">{{ t('adminSettings.clearStacks') }}</NButton>
            </NSpace>
          </template>
          <NAlert v-if="potentialLeakStacks.length > 0" type="error" :title="t('adminSettings.suspectedLeakTitle')" style="margin-bottom: 12px">
            {{ t('adminSettings.suspectedLeakHint', { count: potentialLeakStacks.length }) }}
            <NDataTable
              :columns="runtimeStackVisibleColumns"
              :data="potentialLeakStacks"
              size="small"
              :bordered="false"
              :pagination="{ pageSize: 5 }"
              :scroll-x="runtimeStackTableScrollX"
              style="margin-top: 12px"
            />
          </NAlert>

          <NCollapse v-if="longRunningStacks.length > 0" style="margin-bottom: 12px">
            <NCollapseItem :title="t('adminSettings.longRunningTitle', { count: longRunningStacks.length })" name="long-running-preview">
              <NDataTable
                :columns="runtimeStackVisibleColumns"
                :data="longRunningStacks"
                size="small"
                :bordered="false"
                :pagination="{ pageSize: 10 }"
                :scroll-x="runtimeStackTableScrollX"
              />
            </NCollapseItem>
          </NCollapse>

          <NEmpty v-if="!runtimeStacksLoaded" :description="t('adminSettings.clickToLoadStacks')" />
          <template v-else>
            <NSpace wrap size="small" style="margin-bottom: 12px">
              <NTag type="info">{{ t('adminSettings.runtimeTotal') }}: {{ formatInteger(runtimeStacks.length) }}</NTag>
              <NTag
                v-for="summary in runtimeStateSummary"
                :key="summary.key"
                :type="summary.type as any"
                size="small"
              >
                {{ summary.label }}: {{ formatInteger(summary.count) }}
              </NTag>
            </NSpace>

            <NCollapse v-if="runtimeStacks.length > 0" accordion>
              <NCollapseItem v-for="stack in runtimeStacks" :key="stack.id" :name="stack.id">
                <template #header>
                  <NSpace align="center">
                    <NText code>#{{ stack.id }}</NText>
                    <NText>{{ stack.function }}</NText>
                    <NTag v-if="potentialLeakIdSet.has(stack.id)" type="error" size="small">{{ t('adminSettings.suspectedLeakTag') }}</NTag>
                    <NTag v-if="stack.locked_to_thread" type="warning" size="small">{{ t('adminSettings.lockedToThread') }}</NTag>
                  </NSpace>
                </template>
                <template #header-extra>
                  <NSpace size="small">
                    <NTag :type="getRuntimeStateTagType(stack.state) as any" size="small">{{ stack.state }}</NTag>
                    <NText v-if="stack.wait_time" depth="3" style="font-size: 12px">{{ stack.wait_time }}</NText>
                    <NText depth="3" style="font-size: 11px">{{ t('adminSettings.stackLines', { count: stack.stack_lines }) }}</NText>
                  </NSpace>
                </template>
                <NSpace vertical>
                  <NText v-if="stack.created_by" depth="3" style="font-size: 12px">
                    {{ t('adminSettings.createdBy') }}: {{ stack.created_by }}
                  </NText>
                  <NCode :code="stack.stack" language="text" word-wrap style="max-height: 420px; overflow: auto; font-size: 12px;" />
                </NSpace>
              </NCollapseItem>
            </NCollapse>
            <NEmpty v-else :description="t('adminSettings.noStacksMatchFilter')" />
          </template>
        </NCard>
      </NSpace>
    </NTabPane>
  </NTabs>
</template>
