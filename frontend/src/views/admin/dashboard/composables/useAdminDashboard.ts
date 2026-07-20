/**
 * 管理端仪表盘：数据加载与展示计算（只读概览 + 跳转）
 */
import { computed, h, markRaw, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  ApiOutlined,
  BugOutlined,
  CheckCircleOutlined,
  DashboardOutlined,
  DollarOutlined,
  FileTextOutlined,
  SettingOutlined,
  StarOutlined,
  UserOutlined,
} from '@vicons/antd'
import { useEcharts, useTableColumnVisibility } from '@/hooks'
import type { ECOption } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import type {
  AdminDashboardRecentUser,
  AdminDashboardStatistics,
  AdminDashboardTrendPoint,
} from '@/service/api/admin/dashboard'
import type { BackgroundTaskInfo, DynamicRateLimitSnapshot, ServerOperationsStatusResponse } from '@/service/api/admin/server'
import type { ServerMonitoringService, ServerMonitoringStatusResponse } from '@/service/api/admin/settings'
import {
  formatBytes,
  formatPercent,
  formatStorageFromGB,
  formatStorageFromMB,
  formatUptime,
  normalizePercent,
} from '@/utils/format'

interface DashboardAlert {
  key: string
  type: 'error' | 'warning' | 'info'
  title: string
  detail: string
  priority: number
  path?: string
  query?: Record<string, string>
  actionLabel?: string
}

interface DashboardShortcutAction {
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

interface DashboardSummaryMetric {
  key: string
  label: string
  value: number
  precision: number
  prefix?: string
  suffix?: string
}

interface DashboardSummaryCard {
  key: string
  title: string
  icon: any
  color: string
  metrics: DashboardSummaryMetric[]
}

export function useAdminDashboard() {
  const router = useRouter()
  const message = useMessage()
  const { t } = useI18n()
  const mode = import.meta.env.MODE
  const loading = ref(false)
  const alertOnlyIssues = ref(false)
  const lastRefreshAt = ref<number | null>(null)

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

  const recentUsers = ref<AdminDashboardRecentUser[]>([])
  const recentLoginUsers = ref<AdminDashboardRecentUser[]>([])
  const trends = ref<AdminDashboardTrendPoint[]>([])
  const monitoring = ref<ServerMonitoringStatusResponse | null>(null)
  const operations = ref<ServerOperationsStatusResponse | null>(null)
  const debugEnabled = computed(() => mode !== 'production')

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

  const quick_actions = computed(() => [
    { label: t('adminDashboard.userManagement'), icon: markRaw(UserOutlined), type: 'primary' as const, path: 'users' },
    { label: t('route.paymentOrders'), icon: markRaw(DollarOutlined), type: 'success' as const, path: 'finance/payment-orders' },
    { label: t('route.withdrawManagement'), icon: markRaw(StarOutlined), type: 'warning' as const, path: 'finance/withdraw' },
    { label: t('route.realnameVerify'), icon: markRaw(CheckCircleOutlined), type: 'info' as const, path: 'users/realname' },
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
      { type: 'value', name: t('adminDashboard.paidAmount') },
      { type: 'value', name: t('adminDashboard.operationLogs') },
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
        detail: `${formatPercent(metrics?.memory.used_percent)} · ${formatStorageFromMB(metrics?.memory.used_mb)} / ${formatStorageFromMB(metrics?.memory.total_mb)}`,
      },
      {
        label: t('adminSettings.diskUsage'),
        percentage: normalizePercent(metrics?.disk.used_percent),
        detail: `${metrics?.disk.path || '.'} · ${formatStorageFromGB(metrics?.disk.used_gb)} / ${formatStorageFromGB(metrics?.disk.total_gb)}`,
      },
      {
        label: t('adminDashboard.processCpu'),
        percentage: normalizePercent(process?.process_cpu),
        detail: formatPercent(process?.process_cpu),
      },
      {
        label: t('adminSettings.processMemory'),
        percentage: normalizePercent(processMemoryPercent),
        detail: `${formatStorageFromMB(process?.process_rss_mb)} RSS`,
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
        detail: `${formatPercent(metrics?.memory.used_percent)} · ${formatStorageFromMB(metrics?.memory.used_mb)} / ${formatStorageFromMB(metrics?.memory.total_mb)}`,
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
        path: 'users/realname',
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

  function go_to(sub_path: string, query?: Record<string, string>) {
    if (query) {
      router.push({ path: `/${sub_path}`, query })
      return
    }
    router.push(`/${sub_path}`)
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

  async function handleRefresh() {
    const success = await fetchDashboard()
    if (success)
      message.success(t('adminDashboard.dataRefreshed'))
  }

  onMounted(() => {
    void fetchDashboard()
  })

  return {
    t,
    mode,
    loading,
    alertOnlyIssues,
    lastRefreshAt,
    statistics,
    recentUsers,
    recentLoginUsers,
    monitoring,
    operations,
    topSummaryCards,
    quick_actions,
    resourceRows,
    serviceItems,
    operationsSummary,
    taskHighlights,
    rateLimitHighlights,
    displayedAlertItems,
    alertSeveritySummary,
    alertCountTagType,
    systemShortcutActions,
    recentUserColumnOptions,
    recentUserSelectedColumnKeys,
    recentUserVisibleColumns,
    recentUserVisibleColumnCount,
    recentUserTotalColumnCount,
    recentUserTableScrollX,
    resetRecentUserSelectedColumns,
    recentLoginColumnOptions,
    recentLoginSelectedColumnKeys,
    recentLoginVisibleColumns,
    recentLoginVisibleColumnCount,
    recentLoginTotalColumnCount,
    recentLoginTableScrollX,
    resetRecentLoginSelectedColumns,
    formatCurrency,
    formatDateTime,
    formatBytes,
    formatUptime,
    getServiceTagType,
    getServiceStatusText,
    formatServiceMeta,
    getTaskTagType,
    getTaskStatusText,
    formatTaskMeta,
    formatTaskDuration,
    getRateLimitTagType,
    formatRateLimitMeta,
    getAlertLevelLabel,
    handleAlertClick,
    handleShortcutAction,
    handleRefresh,
    go_to,
  }
}
