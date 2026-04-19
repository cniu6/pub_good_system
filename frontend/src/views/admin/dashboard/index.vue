<script setup lang="ts">
import { computed, h, markRaw, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NEmpty,
  NFlex,
  NGi,
  NGrid,
  NIcon,
  NIconWrapper,
  NNumberAnimation,
  NProgress,
  NSpace,
  NSpin,
  NStatistic,
  NSwitch,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  ApiOutlined,
  BugOutlined,
  CheckCircleOutlined,
  DashboardOutlined,
  DollarOutlined,
  StarOutlined,
  FileTextOutlined,
  ReloadOutlined,
  SettingOutlined,
  UserOutlined,
} from '@vicons/antd'
import { adminApi } from '@/service/api/admin'
import { useEcharts } from '@/hooks'
import type { ECOption } from '@/hooks'
import type {
  AdminDashboardRecentUser,
  AdminDashboardStatistics,
  AdminDashboardTrendPoint,
} from '@/service/api/admin/dashboard'
import type { BackgroundTaskInfo, DynamicRateLimitSnapshot, ServerOperationsStatusResponse } from '@/service/api/admin/server'
import type { ServerMonitoringService, ServerMonitoringStatusResponse } from '@/service/api/admin/settings'

type DashboardAlert = {
  key: string
  type: 'error' | 'warning' | 'info'
  title: string
  detail: string
  priority: number
  path?: string
  query?: Record<string, string>
  actionLabel?: string
}

type DashboardShortcutAction = {
  key: string
  label: string
  description: string
  icon: any
  type: 'primary' | 'info' | 'warning' | 'default' | 'error' | 'success'
  actionLabel: string
  path?: string
  query?: Record<string, string>
  onClick?: () => void | Promise<void>
}

type DashboardSummaryMetric = {
  key: string
  label: string
  value: number
  precision: number
  prefix?: string
  suffix?: string
}

type DashboardSummaryCard = {
  key: string
  title: string
  icon: any
  color: string
  metrics: DashboardSummaryMetric[]
}

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const mode = import.meta.env.MODE
const loading = ref(false)
const runningTaskKey = ref('')
const forceGCLoading = ref(false)
const alertOnlyIssues = ref(false)
const lastRefreshAt = ref<number | null>(null)

// 统计数据（从后端获取）
const statistics = reactive<AdminDashboardStatistics>({
  total_users: 0,
  today_new_users: 0,
  today_active_users: 0,
  active_users_7d: 0,
  total_money_logs: 0,
  total_score_logs: 0,
  total_operation_logs: 0,
  today_operation_logs: 0,
  active_sessions: 0,
  total_payment_orders: 0,
  paid_payment_orders: 0,
  pending_payment_orders: 0,
  total_payment_amount: 0,
  today_payment_orders: 0,
  today_payment_amount: 0,
  month_payment_amount: 0,
  year_payment_amount: 0,
  total_user_balance: 0,
  pending_withdraw_count: 0,
  approved_withdraw_count: 0,
  paid_withdraw_count: 0,
  paid_withdraw_amount: 0,
  total_realname_requests: 0,
  pending_realname_count: 0,
  approved_realname_count: 0,
  rejected_realname_count: 0,
})

// 最近注册用户
const recentUsers = ref<AdminDashboardRecentUser[]>([])
const recentLoginUsers = ref<AdminDashboardRecentUser[]>([])
const trends = ref<AdminDashboardTrendPoint[]>([])
const monitoring = ref<ServerMonitoringStatusResponse | null>(null)
const operations = ref<ServerOperationsStatusResponse | null>(null)
const debugEnabled = computed(() => mode !== 'production')

// 用户表格列
const userColumns = computed<DataTableColumns<AdminDashboardRecentUser>>(() => [
  { title: 'ID', key: 'id', width: 72 },
  { title: t('adminDashboard.username'), key: 'username', width: 140 },
  { title: t('adminDashboard.email'), key: 'email', width: 220, ellipsis: { tooltip: true } },
  {
    title: t('adminDashboard.role'),
    key: 'role',
    width: 90,
    render: row => h(NTag, { type: row.role === 'admin' ? 'error' : 'info', size: 'small' }, () => row.role === 'admin' ? t('adminDashboard.admin') : t('adminDashboard.user')),
  },
  {
    title: t('adminDashboard.status'),
    key: 'status',
    width: 90,
    render: row => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () => row.status === 1 ? t('adminDashboard.normal') : t('adminDashboard.disabled')),
  },
  {
    title: t('adminDashboard.registerTime'),
    key: 'create_time',
    width: 170,
    render: row => formatDateTime(row.create_time),
  },
  {
    title: t('adminUsersDetail.lastLogin'),
    key: 'last_login_time',
    width: 170,
    render: row => formatDateTime(row.last_login_time),
  },
])

const loginUserColumns = computed<DataTableColumns<AdminDashboardRecentUser>>(() => [
  { title: 'ID', key: 'id', width: 72 },
  { title: t('adminDashboard.username'), key: 'username', width: 140 },
  {
    title: t('adminUsers.balance'),
    key: 'money',
    width: 120,
    render: row => `¥${formatCurrency(row.money)}`,
  },
  {
    title: t('adminDashboard.actualPaidAmount'),
    key: 'total_paid_amount',
    width: 140,
    render: row => `¥${formatCurrency(row.total_paid_amount)}`,
  },
  {
    title: t('adminDashboard.rechargeRetentionRatio'),
    key: 'balance_paid_ratio',
    width: 130,
    render: row => formatPercentValue(row.balance_paid_ratio * 100),
  },
  {
    title: t('adminUsersDetail.lastLogin'),
    key: 'last_login_time',
    width: 170,
    render: row => formatDateTime(row.last_login_time),
  },
])

const recentUserSelectableColumnOptions = computed(() => [
  { key: 'id', label: 'ID' },
  { key: 'username', label: t('adminDashboard.username') },
  { key: 'email', label: t('adminDashboard.email') },
  { key: 'role', label: t('adminDashboard.role') },
  { key: 'status', label: t('adminDashboard.status') },
  { key: 'create_time', label: t('adminDashboard.registerTime') },
  { key: 'last_login_time', label: t('adminUsersDetail.lastLogin') },
])

const recentLoginSelectableColumnOptions = computed(() => [
  { key: 'id', label: 'ID' },
  { key: 'username', label: t('adminDashboard.username') },
  { key: 'money', label: t('adminUsers.balance') },
  { key: 'total_paid_amount', label: t('adminDashboard.actualPaidAmount') },
  { key: 'balance_paid_ratio', label: t('adminDashboard.rechargeRetentionRatio') },
  { key: 'last_login_time', label: t('adminUsersDetail.lastLogin') },
])

const {
  columnOptions: recentUserColumnOptions,
  selectedColumnKeys: recentUserSelectedColumnKeys,
  visibleColumns: recentUserVisibleColumns,
  visibleColumnCount: recentUserVisibleColumnCount,
  totalColumnCount: recentUserTotalColumnCount,
  tableScrollX: recentUserTableScrollX,
  resetSelectedColumns: resetRecentUserSelectedColumns,
} = useTableColumnVisibility<AdminDashboardRecentUser>({
  storageKey: 'admin-dashboard-recent-users',
  columns: userColumns,
  options: recentUserSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})

const {
  columnOptions: recentLoginColumnOptions,
  selectedColumnKeys: recentLoginSelectedColumnKeys,
  visibleColumns: recentLoginVisibleColumns,
  visibleColumnCount: recentLoginVisibleColumnCount,
  totalColumnCount: recentLoginTotalColumnCount,
  tableScrollX: recentLoginTableScrollX,
  resetSelectedColumns: resetRecentLoginSelectedColumns,
} = useTableColumnVisibility<AdminDashboardRecentUser>({
  storageKey: 'admin-dashboard-recent-login-users',
  columns: loginUserColumns,
  options: recentLoginSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 700,
})

const totalRechargeRetentionRatio = computed(() => {
  const totalPaid = Number(statistics.total_payment_amount || 0)
  if (totalPaid <= 0)
    return 0
  return Number(statistics.total_user_balance || 0) / totalPaid
})

const topSummaryCards = computed<DashboardSummaryCard[]>(() => [
  {
    key: 'user-overview',
    title: t('adminDashboard.userOverview'),
    icon: markRaw(UserOutlined),
    color: 'var(--info-color)',
    metrics: [
      { key: 'total-users', label: t('adminDashboard.totalUsers'), value: statistics.total_users, precision: 0 },
      { key: 'today-new-users', label: t('adminDashboard.todayNewUsers'), value: statistics.today_new_users, precision: 0 },
      { key: 'today-active-users', label: t('adminDashboard.todayActiveUsers'), value: statistics.today_active_users, precision: 0 },
      { key: 'active-users-7d', label: t('adminDashboard.activeUsers7d'), value: statistics.active_users_7d, precision: 0 },
      { key: 'active-sessions', label: t('adminDashboard.activeSessions'), value: statistics.active_sessions, precision: 0 },
    ],
  },
  {
    key: 'revenue-overview',
    title: t('adminDashboard.revenueOverview'),
    icon: markRaw(DollarOutlined),
    color: 'var(--success-color)',
    metrics: [
      { key: 'today-payment-amount', label: t('adminPaymentOrders.todayRevenue'), value: statistics.today_payment_amount, precision: 2, prefix: '¥' },
      { key: 'month-payment-amount', label: t('adminPaymentOrders.monthRevenue'), value: statistics.month_payment_amount, precision: 2, prefix: '¥' },
      { key: 'year-payment-amount', label: t('adminPaymentOrders.yearRevenue'), value: statistics.year_payment_amount, precision: 2, prefix: '¥' },
      { key: 'total-payment-amount', label: t('adminPaymentOrders.totalRevenue'), value: statistics.total_payment_amount, precision: 2, prefix: '¥' },
      { key: 'total-user-balance', label: t('adminDashboard.totalUserBalance'), value: statistics.total_user_balance, precision: 2, prefix: '¥' },
      { key: 'recharge-retention-ratio', label: t('adminDashboard.rechargeRetentionRatio'), value: totalRechargeRetentionRatio.value * 100, precision: 2, suffix: '%' },
      { key: 'total-payment-orders', label: t('adminPaymentOrders.totalOrders'), value: statistics.total_payment_orders, precision: 0 },
    ],
  },
  {
    key: 'review-overview',
    title: t('adminDashboard.reviewOverview'),
    icon: markRaw(CheckCircleOutlined),
    color: 'var(--warning-color)',
    metrics: [
      { key: 'pending-withdraw-count', label: t('adminDashboard.pendingWithdraw'), value: statistics.pending_withdraw_count, precision: 0 },
      { key: 'pending-realname-count', label: t('adminDashboard.pendingRealname'), value: statistics.pending_realname_count, precision: 0 },
    ],
  },
])

// 快速操作
const quick_actions = computed(() => [
  { label: t('adminDashboard.userManagement'), icon: markRaw(UserOutlined), type: 'primary' as const, path: 'users' },
  { label: t('route.paymentOrders'), icon: markRaw(DollarOutlined), type: 'success' as const, path: 'finance/payment-orders' },
  { label: t('route.withdrawManagement'), icon: markRaw(StarOutlined), type: 'warning' as const, path: 'finance/withdraw' },
  { label: t('route.realnameVerify'), icon: markRaw(CheckCircleOutlined), type: 'info' as const, path: 'realname' },
  { label: t('route.logManagement'), icon: markRaw(FileTextOutlined), type: 'default' as const, path: 'settings/log-management' },
  { label: t('route.serverManagement'), icon: markRaw(SettingOutlined), type: 'error' as const, path: 'settings/server-management' },
])

const businessTrendOptions = computed<ECOption>(() => ({
  tooltip: { trigger: 'axis' },
  legend: { top: 0 },
  grid: { left: 16, right: 16, top: 48, bottom: 16, containLabel: true },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: trends.value.map(item => item.label),
  },
  yAxis: { type: 'value' },
  series: [
    {
      name: t('adminDashboard.newUsers'),
      type: 'line',
      smooth: true,
      data: trends.value.map(item => item.new_users),
    },
    {
      name: t('adminDashboard.activeUsers'),
      type: 'line',
      smooth: true,
      data: trends.value.map(item => item.active_users),
    },
    {
      name: t('adminDashboard.paidOrders'),
      type: 'bar',
      barMaxWidth: 24,
      data: trends.value.map(item => item.paid_orders),
    },
  ],
}))

const revenueTrendOptions = computed<ECOption>(() => ({
  tooltip: { trigger: 'axis' },
  legend: { top: 0 },
  grid: { left: 16, right: 16, top: 48, bottom: 16, containLabel: true },
  xAxis: {
    type: 'category',
    data: trends.value.map(item => item.label),
  },
  yAxis: [
    {
      type: 'value',
      name: t('adminDashboard.paidAmount'),
    },
    {
      type: 'value',
      name: t('adminDashboard.operationLogs'),
    },
  ],
  series: [
    {
      name: t('adminDashboard.paidAmount'),
      type: 'line',
      smooth: true,
      yAxisIndex: 0,
      data: trends.value.map(item => item.paid_amount),
    },
    {
      name: t('adminDashboard.operationLogs'),
      type: 'bar',
      barMaxWidth: 24,
      yAxisIndex: 1,
      data: trends.value.map(item => item.operation_logs),
    },
  ],
}))

const verifyTrendOptions = computed<ECOption>(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0 },
  series: [
    {
      name: t('adminDashboard.verifyDistribution'),
      type: 'pie',
      radius: ['45%', '72%'],
      center: ['50%', '45%'],
      label: { formatter: '{b}\n{c}' },
      data: [
        { value: statistics.pending_realname_count, name: t('adminUsersDetail.pendingReview') },
        { value: statistics.approved_realname_count, name: t('adminUsersDetail.approved') },
        { value: statistics.rejected_realname_count, name: t('adminUsersDetail.rejected') },
      ],
    },
  ],
}))

useEcharts('businessTrendRef', businessTrendOptions)
useEcharts('revenueTrendRef', revenueTrendOptions)
useEcharts('verifyTrendRef', verifyTrendOptions)

const resourceRows = computed(() => {
  const metrics = monitoring.value?.metrics
  const process = monitoring.value?.process
  const totalMemory = Number(metrics?.memory.total_mb || 0)
  const processMemoryPercent = totalMemory > 0 ? (Number(process?.process_rss_mb || 0) / totalMemory) * 100 : 0
  return [
    {
      label: t('adminSettings.cpu'),
      percentage: normalizePercent(metrics?.cpu.usage_percent),
      detail: `${formatPercent(metrics?.cpu.usage_percent)} · ${t('adminSettings.cpuCores', { count: metrics?.cpu.core_count || 0 })}`,
    },
    {
      label: t('adminSettings.systemMemory'),
      percentage: normalizePercent(metrics?.memory.used_percent),
      detail: `${formatPercent(metrics?.memory.used_percent)} · ${formatStorageMB(metrics?.memory.used_mb)} / ${formatStorageMB(metrics?.memory.total_mb)}`,
    },
    {
      label: t('adminSettings.diskUsage'),
      percentage: normalizePercent(metrics?.disk.used_percent),
      detail: `${metrics?.disk.path || '.'} · ${formatStorageGB(metrics?.disk.used_gb)} / ${formatStorageGB(metrics?.disk.total_gb)}`,
    },
    {
      label: t('adminDashboard.processCpu'),
      percentage: normalizePercent(process?.process_cpu),
      detail: formatPercent(process?.process_cpu),
    },
    {
      label: t('adminSettings.processMemory'),
      percentage: normalizePercent(processMemoryPercent),
      detail: `${formatStorageMB(process?.process_rss_mb)} RSS`,
    },
  ]
})

const serviceItems = computed(() => monitoring.value?.services || [])
const operationTasks = computed(() => operations.value?.tasks || [])
const rateLimitItems = computed(() => operations.value?.rate_limits || [])

const runningTaskCount = computed(() => operationTasks.value.filter(task => task.running).length)
const failedTaskCount = computed(() => operationTasks.value.filter(task => !task.running && task.last_status === 'failed').length)
const blockedRequestCount = computed(() => rateLimitItems.value.reduce((sum, item) => sum + Number(item.blocked_count || 0), 0))
const activeVisitorCount = computed(() => rateLimitItems.value.reduce((sum, item) => sum + Number(item.active_visitors || 0), 0))

const operationsSummary = computed(() => [
  { label: t('adminDashboard.runningTasks'), value: runningTaskCount.value },
  { label: t('adminDashboard.failedTasks'), value: failedTaskCount.value },
  { label: t('adminDashboard.blockedRequests'), value: blockedRequestCount.value },
  { label: t('adminDashboard.activeVisitors'), value: activeVisitorCount.value },
])

const taskHighlights = computed(() => {
  return [...operationTasks.value]
    .sort((a, b) => getTaskPriority(b) - getTaskPriority(a))
    .slice(0, 3)
})

const rateLimitHighlights = computed(() => {
  return [...rateLimitItems.value]
    .sort((a, b) => {
      const blockedDiff = Number(b.blocked_count || 0) - Number(a.blocked_count || 0)
      if (blockedDiff !== 0)
        return blockedDiff
      return Number(b.active_visitors || 0) - Number(a.active_visitors || 0)
    })
    .slice(0, 3)
})

const alertItems = computed<DashboardAlert[]>(() => {
  const items: DashboardAlert[] = []
  const metrics = monitoring.value?.metrics

  const cpuUsage = normalizePercent(metrics?.cpu.usage_percent)
  if (cpuUsage >= 85) {
    items.push({
      key: 'cpu',
      type: cpuUsage >= 92 ? 'error' : 'warning',
      title: t('adminSettings.cpu'),
      detail: formatPercent(metrics?.cpu.usage_percent),
      priority: cpuUsage >= 92 ? 100 : 90,
      path: 'settings/server-management',
      query: { tab: 'monitor' },
      actionLabel: t('adminDashboard.viewDetails'),
    })
  }

  const memoryUsage = normalizePercent(metrics?.memory.used_percent)
  if (memoryUsage >= 85) {
    items.push({
      key: 'memory',
      type: memoryUsage >= 92 ? 'error' : 'warning',
      title: t('adminSettings.systemMemory'),
      detail: `${formatPercent(metrics?.memory.used_percent)} · ${formatStorageMB(metrics?.memory.used_mb)} / ${formatStorageMB(metrics?.memory.total_mb)}`,
      priority: memoryUsage >= 92 ? 98 : 88,
      path: 'settings/server-management',
      query: { tab: 'monitor' },
      actionLabel: t('adminDashboard.viewDetails'),
    })
  }

  const diskUsage = normalizePercent(metrics?.disk.used_percent)
  if (diskUsage >= 85) {
    items.push({
      key: 'disk',
      type: diskUsage >= 92 ? 'error' : 'warning',
      title: t('adminSettings.diskUsage'),
      detail: `${formatPercent(metrics?.disk.used_percent)} · ${metrics?.disk.path || '.'}`,
      priority: diskUsage >= 92 ? 96 : 86,
      path: 'settings/server-management',
      query: { tab: 'monitor' },
      actionLabel: t('adminDashboard.viewDetails'),
    })
  }

  if (statistics.pending_realname_count > 0) {
    items.push({
      key: 'realname-pending',
      type: 'warning',
      title: t('adminDashboard.pendingRealname'),
      detail: t('adminDashboard.pendingRealnameDetail', { count: statistics.pending_realname_count }),
      priority: 84,
      path: 'realname',
      actionLabel: t('adminDashboard.handleNow'),
    })
  }

  if (statistics.pending_withdraw_count > 0) {
    items.push({
      key: 'withdraw-pending',
      type: 'warning',
      title: t('adminDashboard.pendingWithdraw'),
      detail: t('adminDashboard.pendingWithdrawDetail', { count: statistics.pending_withdraw_count }),
      priority: 82,
      path: 'finance/withdraw',
      actionLabel: t('adminDashboard.handleNow'),
    })
  }

  if (statistics.pending_payment_orders > 0) {
    items.push({
      key: 'payment-pending',
      type: 'info',
      title: t('adminDashboard.pendingOrders'),
      detail: t('adminDashboard.pendingOrdersDetail', { count: statistics.pending_payment_orders }),
      priority: 72,
      path: 'finance/payment-orders',
      actionLabel: t('adminDashboard.handleNow'),
    })
  }

  for (const service of serviceItems.value) {
    if (service.status && service.status !== 'up') {
      items.push({
        key: `service-${service.name}`,
        type: service.status === 'down' ? 'error' : 'warning',
        title: service.name,
        detail: service.message || formatServiceMeta(service),
        priority: service.status === 'down' ? 94 : 76,
        path: 'settings/server-management',
        query: { tab: 'monitor' },
        actionLabel: t('adminDashboard.viewDetails'),
      })
    }
  }

  for (const task of operationTasks.value) {
    if (!task.running && task.last_status === 'failed') {
      items.push({
        key: `task-${task.key}`,
        type: 'error',
        title: task.label,
        detail: task.last_message || t('adminServer.failed'),
        priority: 92,
        path: 'settings/server-management',
        query: { tab: 'ops' },
        actionLabel: t('adminDashboard.viewDetails'),
      })
    }
  }

  for (const limiter of rateLimitItems.value) {
    if (limiter.enabled && Number(limiter.blocked_count || 0) > 0) {
      items.push({
        key: `rate-limit-${limiter.name}`,
        type: 'warning',
        title: limiter.name,
        detail: formatRateLimitMeta(limiter),
        priority: 74,
        path: 'settings/server-management',
        query: { tab: 'ops' },
        actionLabel: t('adminDashboard.viewDetails'),
      })
    }
  }

  return items
    .sort((a, b) => b.priority - a.priority)
    .slice(0, 6)
})

const displayedAlertItems = computed(() => {
  if (!alertOnlyIssues.value)
    return alertItems.value
  return alertItems.value.filter(item => item.type !== 'info')
})

const alertSeveritySummary = computed(() => ({
  error: alertItems.value.filter(item => item.type === 'error').length,
  warning: alertItems.value.filter(item => item.type === 'warning').length,
  info: alertItems.value.filter(item => item.type === 'info').length,
}))

const alertCountTagType = computed(() => {
  if (displayedAlertItems.value.some(item => item.type === 'error'))
    return 'error' as const
  if (displayedAlertItems.value.length > 0)
    return 'warning' as const
  return 'success' as const
})

const systemShortcutActions = computed<DashboardShortcutAction[]>(() => {
  const actions: DashboardShortcutAction[] = [
    {
      key: 'monitor',
      label: t('adminDashboard.monitorCenter'),
      description: t('adminDashboard.monitorCenterDesc'),
      icon: markRaw(DashboardOutlined),
      type: 'primary',
      actionLabel: t('adminDashboard.viewDetails'),
      path: 'settings/server-management',
      query: { tab: 'monitor' },
    },
    {
      key: 'operations',
      label: t('adminDashboard.operationsConsole'),
      description: t('adminDashboard.operationsConsoleDesc'),
      icon: markRaw(SettingOutlined),
      type: 'warning',
      actionLabel: t('adminDashboard.viewDetails'),
      path: 'settings/server-management',
      query: { tab: 'ops' },
    },
    {
      key: 'api-log',
      label: t('adminDashboard.apiLogCenter'),
      description: t('adminDashboard.apiLogCenterDesc'),
      icon: markRaw(ApiOutlined),
      type: 'info',
      actionLabel: t('adminDashboard.viewDetails'),
      path: 'settings/log-management',
      query: { tab: 'api' },
    },
  ]

  if (debugEnabled.value) {
    actions.push({
      key: 'debug',
      label: t('adminDashboard.debugCenter'),
      description: t('adminDashboard.debugCenterDesc'),
      icon: markRaw(BugOutlined),
      type: 'error',
      actionLabel: t('adminDashboard.viewDetails'),
      path: 'settings/server-management',
      query: { tab: 'debug' },
    })
  }

  return actions
})

function normalizePercent(value?: number | null) {
  return Math.max(0, Math.min(100, Number(value || 0)))
}

function formatPercent(value?: number | null) {
  return `${Number(value || 0).toFixed(1)}%`
}

function formatPercentValue(value?: number | null) {
  return `${Number(value || 0).toFixed(2)}%`
}

function formatCurrency(value?: number | null) {
  return Number(value || 0).toFixed(2)
}

function formatDateTime(value?: string | number | null) {
  if (value === null || value === undefined || value === '')
    return '-'
  if (typeof value === 'string')
    return new Date(value).toLocaleString()
  return new Date(value * 1000).toLocaleString()
}

function formatStorageMB(value?: number | null) {
  const amount = Number(value || 0)
  if (amount >= 1024)
    return `${(amount / 1024).toFixed(2)} GB`
  return `${amount.toFixed(1)} MB`
}

function formatStorageGB(value?: number | null) {
  return `${Number(value || 0).toFixed(2)} GB`
}

function formatTraffic(value?: number | null) {
  const amount = Number(value || 0)
  if (amount >= 1024 ** 3)
    return `${(amount / 1024 ** 3).toFixed(2)} GB`
  if (amount >= 1024 ** 2)
    return `${(amount / 1024 ** 2).toFixed(2)} MB`
  if (amount >= 1024)
    return `${(amount / 1024).toFixed(2)} KB`
  return `${amount.toFixed(0)} B`
}

function formatUptime(seconds?: number | null) {
  const totalSeconds = Math.max(0, Math.floor(Number(seconds || 0)))
  const day = Math.floor(totalSeconds / 86400)
  const hour = Math.floor((totalSeconds % 86400) / 3600)
  const minute = Math.floor((totalSeconds % 3600) / 60)
  const second = totalSeconds % 60
  return t('adminSettings.uptimePreciseFormat', { day, hour, minute, second })
}

function getServiceTagType(status?: ServerMonitoringService['status']) {
  if (status === 'up')
    return 'success' as const
  if (status === 'warning')
    return 'warning' as const
  return 'error' as const
}

function getServiceStatusText(status?: ServerMonitoringService['status']) {
  if (status === 'up')
    return t('adminSettings.statusNormal')
  if (status === 'warning')
    return t('adminSettings.statusWarning')
  return t('adminSettings.statusError')
}

function getTaskPriority(task: BackgroundTaskInfo) {
  if (task.running)
    return 3
  if (task.last_status === 'failed')
    return 2
  if (task.last_status === 'success')
    return 1
  return 0
}

function getTaskTagType(task: BackgroundTaskInfo) {
  if (task.running)
    return 'warning' as const
  if (task.last_status === 'success')
    return 'success' as const
  if (task.last_status === 'failed')
    return 'error' as const
  return 'default' as const
}

function getTaskStatusText(task: BackgroundTaskInfo) {
  if (task.running)
    return t('adminServer.running')
  if (task.last_status === 'success')
    return t('adminServer.success')
  if (task.last_status === 'failed')
    return t('adminServer.failed')
  return t('adminServer.idle')
}

function formatServiceMeta(service: ServerMonitoringService) {
  if (typeof service.open_connections === 'number') {
    return [
      t('adminSettings.serviceConnections', { count: service.open_connections || 0 }),
      t('adminSettings.serviceInUse', { count: service.in_use || 0 }),
      t('adminSettings.serviceIdle', { count: service.idle || 0 }),
    ].join(' · ')
  }
  if (service.host || service.port)
    return [service.host, service.port].filter(Boolean).join(':')
  return service.message || '-'
}

function getRateLimitTagType(item: DynamicRateLimitSnapshot) {
  if (!item.enabled)
    return 'default' as const
  if (Number(item.blocked_count || 0) > 0)
    return 'warning' as const
  return 'success' as const
}

function formatTaskMeta(task: BackgroundTaskInfo) {
  const parts: string[] = []
  if (task.next_run_time)
    parts.push(`${t('adminServer.tasks.nextRun')}: ${formatDateTime(task.next_run_time)}`)
  if (task.last_message)
    parts.push(`${t('adminServer.tasks.message')}: ${task.last_message}`)
  return parts.join(' · ') || '-'
}

function formatTaskDuration(durationMs?: number | null) {
  return `${Number(durationMs || 0)}ms`
}

function formatRateLimitMeta(item: DynamicRateLimitSnapshot) {
  return `${t('adminDashboard.blockedRequests')}: ${item.blocked_count || 0} · ${t('adminDashboard.activeVisitors')}: ${item.active_visitors || 0}`
}

function getAlertLevelLabel(type: DashboardAlert['type']) {
  if (type === 'error')
    return t('adminDashboard.criticalLevel')
  if (type === 'warning')
    return t('adminDashboard.warningLevel')
  return t('adminDashboard.infoLevel')
}

function handleAlertClick(item: DashboardAlert) {
  if (!item.path)
    return
  go_to(item.path, item.query)
}

function handleShortcutAction(action: DashboardShortcutAction) {
  if (action.path) {
    go_to(action.path, action.query)
    return
  }
  action.onClick?.()
}

async function handleRunTaskFromDashboard(task: BackgroundTaskInfo) {
  if (task.running)
    return
  runningTaskKey.value = task.key
  try {
    const res = await adminApi.server.runTask(task.key)
    message.success(res.data?.message || t('adminServer.tasks.runSuccess'))
    await fetchDashboard()
  }
  catch {
    message.error(t('adminServer.tasks.runFailed'))
  }
  finally {
    runningTaskKey.value = ''
  }
}

async function handleForceGCFromDashboard() {
  forceGCLoading.value = true
  try {
    const res = await adminApi.debug.forceGC()
    message.success(res.data?.message || t('adminSettings.gcCompleted', { before: res.data?.goroutines_before || 0, after: res.data?.goroutines_after || 0 }))
    await fetchDashboard()
  }
  catch (error: any) {
    message.error(`${t('adminSettings.operationFailed')}${error?.message || ''}`)
  }
  finally {
    forceGCLoading.value = false
  }
}

// 获取仪表盘数据
async function fetchDashboard() {
  loading.value = true
  let hasData = false
  try {
    const [dashboardResult, monitoringResult, operationsResult] = await Promise.allSettled([
      adminApi.dashboard.getStatistics(),
      adminApi.server.monitoring(),
      adminApi.server.operations(),
    ])

    if (dashboardResult.status === 'fulfilled') {
      const dashboardRes = dashboardResult.value
      if (dashboardRes.isSuccess && dashboardRes.data) {
        const stats = dashboardRes.data.statistics
        if (stats) {
          Object.assign(statistics, stats)
          hasData = true
        }
        recentUsers.value = dashboardRes.data.recent_users || []
        recentLoginUsers.value = dashboardRes.data.recent_login_users || []
        trends.value = dashboardRes.data.trends || []
      }
    }
    else if (import.meta.env.DEV) {
      console.error('[adminDashboard] dashboard request failed', dashboardResult.reason)
    }

    if (monitoringResult.status === 'fulfilled') {
      const monitoringRes = monitoringResult.value
      if (monitoringRes.isSuccess && monitoringRes.data) {
        monitoring.value = monitoringRes.data
        hasData = true
      }
    }
    else if (import.meta.env.DEV) {
      console.error('[adminDashboard] monitoring request failed', monitoringResult.reason)
    }

    if (operationsResult.status === 'fulfilled') {
      const operationsRes = operationsResult.value
      if (operationsRes.isSuccess && operationsRes.data) {
        operations.value = operationsRes.data
        hasData = true
      }
    }
    else if (import.meta.env.DEV) {
      console.error('[adminDashboard] operations request failed', operationsResult.reason)
    }

    if (!hasData)
      message.error(t('adminDashboard.fetchFailed'))
    else
      lastRefreshAt.value = Math.floor(Date.now() / 1000)
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminDashboard] fetch failed', error)
    message.error(t('adminDashboard.fetchFailed'))
  }
  finally {
    loading.value = false
  }
  return hasData
}

function go_to(sub_path: string, query?: Record<string, string>) {
  if (query) {
    router.push({ path: `/${sub_path}`, query })
    return
  }
  router.push(`/${sub_path}`)
}

async function handleRefresh() {
  const success = await fetchDashboard()
  if (success)
    message.success(t('adminDashboard.dataRefreshed'))
}

onMounted(() => {
  void fetchDashboard()
})
</script>

<template>
  <n-space vertical :size="16">
    <!-- 欢迎横幅 -->
    <n-card hoverable>
      <n-flex justify="space-between" align="center" wrap :size="16">
        <n-flex align="center" :size="16">
          <n-icon-wrapper :size="48" :border-radius="12" color="var(--success-color)">
            <n-icon :size="26" color="#fff">
              <CheckCircleOutlined />
            </n-icon>
          </n-icon-wrapper>
          <n-flex vertical>
            <n-text strong>{{ t('adminDashboard.welcomeBack') }}</n-text>
            <n-text depth="3">{{ t('adminDashboard.systemOverview') }}</n-text>
          </n-flex>
        </n-flex>
        <n-flex :size="8">
          <n-button :loading="loading" @click="handleRefresh">{{ t('adminDashboard.refreshData') }}</n-button>
          <n-button type="primary" @click="go_to('finance/payment-orders')">{{ t('route.paymentOrders') }}</n-button>
          <n-button @click="go_to('settings/server-management')">{{ t('route.serverManagement') }}</n-button>
        </n-flex>
      </n-flex>
    </n-card>

    <n-grid :x-gap="16" :y-gap="16" :cols="3" item-responsive responsive="screen">
      <n-gi v-for="card in topSummaryCards" :key="card.key" span="3 m:1">
        <n-card hoverable class="dashboard-summary-card">
          <n-space vertical :size="16">
            <n-flex align="center" :size="12" class="dashboard-summary-header">
              <n-icon-wrapper :size="42" :color="card.color" :border-radius="12">
                <n-icon :size="22" color="#fff">
                  <component :is="card.icon" />
                </n-icon>
              </n-icon-wrapper>
              <n-text strong>{{ card.title }}</n-text>
            </n-flex>

            <n-space vertical :size="12">
              <n-flex v-for="metric in card.metrics" :key="metric.key" justify="space-between" align="center" class="dashboard-summary-row">
                <n-text depth="3">{{ metric.label }}</n-text>
                <div class="dashboard-summary-value">
                  <span v-if="metric.prefix">{{ metric.prefix }}</span>
                  <n-number-animation :from="0" :to="metric.value" :precision="metric.precision" show-separator />
                  <span v-if="metric.suffix">{{ metric.suffix }}</span>
                </div>
              </n-flex>
            </n-space>
          </n-space>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 下半区域 -->
    <n-grid :x-gap="16" :y-gap="16" :cols="12" item-responsive responsive="screen">
      <n-gi span="12 xl:8">
        <n-card :title="t('adminDashboard.businessTrend')" hoverable>
          <div ref="businessTrendRef" style="height: 320px;" />
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.verifyDistribution')" hoverable>
          <template #header-extra>
            <n-text depth="3">{{ t('adminRealname.totalRecords', { total: statistics.total_realname_requests }) }}</n-text>
          </template>
          <div ref="verifyTrendRef" style="height: 320px;" />
        </n-card>
      </n-gi>

      <n-gi span="12 xl:8">
        <n-card :title="t('adminDashboard.revenueTrend')" hoverable>
          <div ref="revenueTrendRef" style="height: 320px;" />
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminSettings.processResources')" hoverable>
          <n-space vertical :size="14">
            <div v-for="item in resourceRows" :key="item.label">
              <n-flex justify="space-between" align="center">
                <n-text>{{ item.label }}</n-text>
                <n-text depth="3">{{ item.detail }}</n-text>
              </n-flex>
              <n-progress type="line" :percentage="item.percentage" :status="item.percentage >= 85 ? 'error' : item.percentage >= 70 ? 'warning' : 'success'" />
            </div>
          </n-space>
        </n-card>
      </n-gi>

      <!-- 运行概览 -->
      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.financeOverview')" hoverable>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item :label="t('adminDashboard.paidOrders')">
              {{ statistics.paid_payment_orders }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.pendingOrders')">
              {{ statistics.pending_payment_orders }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.statApproved')">
              {{ statistics.approved_withdraw_count }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.statPaidCount')">
              {{ statistics.paid_withdraw_count }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.paidWithdrawAmount')">
              ¥{{ formatCurrency(statistics.paid_withdraw_amount) }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminSettings.systemInfo')" hoverable>
          <template #header-extra>
            <n-text depth="3">{{ t('adminSettings.lastRefreshed') }} {{ formatDateTime(monitoring?.generated_at) }}</n-text>
          </template>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item :label="t('adminSettings.appName')">
              {{ monitoring?.app.name || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.appMode')">
              <n-tag size="small" :type="mode === 'production' ? 'success' : 'warning'">{{ monitoring?.app.mode || mode }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.port')">
              {{ monitoring?.app.port || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.pid')">
              {{ monitoring?.process.pid || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.goVersion')">
              {{ monitoring?.app.go_version || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.goroutines')">
              {{ monitoring?.process.goroutines || 0 }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.uptime')">
              {{ formatUptime(monitoring?.uptime_seconds) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.upload')">
              {{ formatTraffic(monitoring?.metrics.network.bytes_sent) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.download')">
              {{ formatTraffic(monitoring?.metrics.network.bytes_recv) }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminSettings.serviceHealthSnapshot')" hoverable>
          <n-space vertical :size="12">
            <n-card v-for="service in serviceItems" :key="service.name" size="small" embedded>
              <n-space vertical :size="8">
                <n-flex justify="space-between" align="center" wrap>
                  <n-space align="center" :size="8">
                    <n-text strong>{{ service.name }}</n-text>
                    <n-tag size="small" :type="getServiceTagType(service.status)">{{ getServiceStatusText(service.status) }}</n-tag>
                  </n-space>
                </n-flex>
                <n-text depth="3">{{ formatServiceMeta(service) }}</n-text>
              </n-space>
            </n-card>
            <n-empty v-if="serviceItems.length === 0 && !loading" />
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.operationsOverview')" hoverable>
          <template #header-extra>
            <n-button text type="primary" @click="go_to('settings/server-management')">{{ t('route.serverManagement') }}</n-button>
          </template>
          <n-space vertical :size="16">
            <n-grid :x-gap="12" :y-gap="12" :cols="2">
              <n-gi v-for="item in operationsSummary" :key="item.label">
                <n-card size="small" embedded>
                  <n-statistic :label="item.label" :value="item.value" />
                </n-card>
              </n-gi>
            </n-grid>

            <n-divider style="margin: 0;" />

            <n-space vertical :size="10">
              <n-text strong>{{ t('adminServer.tasks.title') }}</n-text>
              <n-card v-for="task in taskHighlights" :key="task.key" size="small" embedded>
                <n-space vertical :size="8">
                  <n-flex justify="space-between" align="center" wrap>
                    <n-space align="center" :size="8">
                      <n-text strong>{{ task.label }}</n-text>
                      <n-tag size="small" :type="getTaskTagType(task)">{{ getTaskStatusText(task) }}</n-tag>
                    </n-space>
                    <n-space align="center" :size="8">
                      <n-text depth="3">{{ task.interval_secs }}s</n-text>
                      <n-button size="tiny" tertiary type="primary" :loading="runningTaskKey === task.key" :disabled="task.running" @click="handleRunTaskFromDashboard(task)">
                        {{ t('adminServer.tasks.runNow') }}
                      </n-button>
                    </n-space>
                  </n-flex>
                  <n-space :size="12" wrap>
                    <n-text depth="3">{{ t('adminServer.tasks.lastRun') }}: {{ formatDateTime(task.last_run_time) }}</n-text>
                    <n-text depth="3">{{ t('adminServer.tasks.duration') }}: {{ formatTaskDuration(task.last_duration_ms) }}</n-text>
                  </n-space>
                  <n-text depth="3">{{ formatTaskMeta(task) }}</n-text>
                </n-space>
              </n-card>
              <n-empty v-if="taskHighlights.length === 0 && !loading" />
            </n-space>

            <n-divider style="margin: 0;" />

            <n-space vertical :size="10">
              <n-flex justify="space-between" align="center" wrap>
                <n-text strong>{{ t('adminServer.rateLimit.title') }}</n-text>
                <n-tag size="small" :type="operations?.api_log.enabled ? 'success' : 'default'">
                  {{ t('adminServer.runtimeConfig.apiLog') }}: {{ operations?.api_log.enabled ? t('common.enable') : t('common.disable') }}
                </n-tag>
              </n-flex>
              <n-card v-for="item in rateLimitHighlights" :key="item.name" size="small" embedded>
                <n-space vertical :size="8">
                  <n-flex justify="space-between" align="center" wrap>
                    <n-space align="center" :size="8">
                      <n-text strong>{{ item.name }}</n-text>
                      <n-tag size="small" :type="getRateLimitTagType(item)">{{ item.enabled ? t('common.enable') : t('common.disable') }}</n-tag>
                    </n-space>
                    <n-text depth="3">R{{ item.rate }} / B{{ item.burst }}</n-text>
                  </n-flex>
                  <n-text depth="3">{{ formatRateLimitMeta(item) }}</n-text>
                </n-space>
              </n-card>
              <n-empty v-if="rateLimitHighlights.length === 0 && !loading" />
            </n-space>
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.alertCenter')" hoverable>
          <template #header-extra>
            <n-flex align="center" :size="8" wrap>
              <n-text depth="3">{{ t('adminDashboard.lastUpdated') }} {{ formatDateTime(lastRefreshAt) }}</n-text>
              <n-tag v-if="alertSeveritySummary.error" size="small" type="error">{{ t('adminDashboard.criticalLevel') }} {{ alertSeveritySummary.error }}</n-tag>
              <n-tag v-if="alertSeveritySummary.warning" size="small" type="warning">{{ t('adminDashboard.warningLevel') }} {{ alertSeveritySummary.warning }}</n-tag>
              <n-tag v-if="alertSeveritySummary.info" size="small" type="info">{{ t('adminDashboard.infoLevel') }} {{ alertSeveritySummary.info }}</n-tag>
              <n-tag size="small" :type="alertCountTagType">{{ displayedAlertItems.length }}</n-tag>
              <n-text depth="3">{{ t('adminDashboard.onlyIssues') }}</n-text>
              <n-switch v-model:value="alertOnlyIssues" size="small" />
            </n-flex>
          </template>
          <n-space vertical :size="12">
            <n-alert v-for="item in displayedAlertItems" :key="item.key" :type="item.type" :title="item.title">
              <n-space vertical :size="8">
                <n-tag size="small" :type="item.type === 'info' ? 'info' : item.type">{{ getAlertLevelLabel(item.type) }}</n-tag>
                <span>{{ item.detail }}</span>
                <n-button v-if="item.path" size="tiny" tertiary type="primary" @click="handleAlertClick(item)">
                  {{ item.actionLabel || t('adminDashboard.viewDetails') }}
                </n-button>
              </n-space>
            </n-alert>
            <n-empty v-if="displayedAlertItems.length === 0 && !loading" :description="alertOnlyIssues ? t('adminDashboard.noIssueAlerts') : t('adminDashboard.noAlerts')" />
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.systemActions')" hoverable>
          <n-space vertical :size="12">
            <n-grid :x-gap="12" :y-gap="12" :cols="2" item-responsive>
              <n-gi v-for="action in systemShortcutActions" :key="action.key" span="2 s:1">
                <n-card size="small" embedded>
                  <n-space vertical :size="10">
                    <n-flex justify="space-between" align="center">
                      <n-text strong>{{ action.label }}</n-text>
                      <n-icon>
                        <component :is="action.icon" />
                      </n-icon>
                    </n-flex>
                    <n-text depth="3">{{ action.description }}</n-text>
                    <n-button block :type="action.type" tertiary @click="handleShortcutAction(action)">
                      {{ action.actionLabel }}
                    </n-button>
                  </n-space>
                </n-card>
              </n-gi>
            </n-grid>
            <n-divider style="margin: 0;" />
            <n-space :size="8" wrap>
              <n-button :loading="loading" @click="handleRefresh">
                <template #icon>
                  <n-icon><ReloadOutlined /></n-icon>
                </template>
                {{ t('adminSettings.refresh') }}
              </n-button>
              <n-button v-if="debugEnabled" type="warning" :loading="forceGCLoading" @click="handleForceGCFromDashboard">
                {{ t('adminSettings.forceGC') }}
              </n-button>
            </n-space>
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.quickActions')" hoverable>
          <n-grid :x-gap="12" :y-gap="12" :cols="2" item-responsive>
            <n-gi v-for="action in quick_actions" :key="action.label" span="2 s:1">
              <n-button block :type="action.type" ghost size="large" @click="go_to(action.path)">
                <template #icon>
                  <n-icon><component :is="action.icon" /></n-icon>
                </template>
                {{ action.label }}
              </n-button>
            </n-gi>
          </n-grid>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.recentUsers')" hoverable>
          <template #header-extra>
            <n-space>
              <TableColumnSelector
                v-model="recentUserSelectedColumnKeys"
                :options="recentUserColumnOptions"
                :visible-count="recentUserVisibleColumnCount"
                :total-count="recentUserTotalColumnCount"
                :button-label="t('common.showFields')"
                :title="t('common.visibleFields')"
                :hint="t('common.columnVisibilityHint')"
                :reset-label="t('common.restoreDefaultFields')"
                @reset="resetRecentUserSelectedColumns"
              />
              <n-button type="primary" quaternary @click="go_to('users')">{{ t('adminDashboard.viewAll') }}</n-button>
            </n-space>
          </template>
          <n-spin :show="loading">
            <n-data-table
              :columns="recentUserVisibleColumns"
              :data="recentUsers"
              :bordered="false"
              :single-line="false"
              size="small"
              :pagination="false"
              :scroll-x="recentUserTableScrollX"
            />
            <n-empty v-if="!loading && recentUsers.length === 0" :description="t('adminDashboard.noUserData')" />
          </n-spin>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.recentLoginUsers')" hoverable>
          <template #header-extra>
            <n-space>
              <TableColumnSelector
                v-model="recentLoginSelectedColumnKeys"
                :options="recentLoginColumnOptions"
                :visible-count="recentLoginVisibleColumnCount"
                :total-count="recentLoginTotalColumnCount"
                :button-label="t('common.showFields')"
                :title="t('common.visibleFields')"
                :hint="t('common.columnVisibilityHint')"
                :reset-label="t('common.restoreDefaultFields')"
                @reset="resetRecentLoginSelectedColumns"
              />
              <n-button type="primary" quaternary @click="go_to('users')">{{ t('adminDashboard.viewAll') }}</n-button>
            </n-space>
          </template>
          <n-spin :show="loading">
            <n-data-table
              :columns="recentLoginVisibleColumns"
              :data="recentLoginUsers"
              :bordered="false"
              :single-line="false"
              size="small"
              :pagination="false"
              :scroll-x="recentLoginTableScrollX"
            />
            <n-empty v-if="!loading && recentLoginUsers.length === 0" :description="t('adminDashboard.noUserData')" />
          </n-spin>
        </n-card>
      </n-gi>
    </n-grid>
  </n-space>
</template>

<style scoped>
.dashboard-summary-card {
  height: 100%;
}

.dashboard-summary-header {
  min-height: 42px;
}

.dashboard-summary-row {
  gap: 12px;
}

.dashboard-summary-value {
  display: flex;
  align-items: baseline;
  gap: 2px;
  font-size: 18px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}
</style>
