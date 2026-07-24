<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NCode,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NGi,
  NGrid,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NStatistic,
  NSwitch,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import { useEcharts, useRequestGuard, useTableColumnVisibility, withSubmitLock } from '@/hooks'
import type { ECOption } from '@/hooks'
import { adminAPILogApi } from '@/service/api/admin/api-log'
import type { APIAccessLog, APIAccessLogListParams, APIAccessLogStats } from '@/service/api/admin/api-log'
import { normalizeAPILogCleanupIntervalSeconds, normalizeLogMaxCount, normalizeLogPerUserMaxCount, normalizeLogQueryDays, parseBooleanSetting } from '@/utils'
import { formatPrettyJSON } from '@/utils/format'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const runtimeLoading = ref(false)
const runtimeSaving = ref(false)
const logList = ref<APIAccessLog[]>([])
const total = ref(0)
const statsData = ref<APIAccessLogStats>({
  total_count: 0,
  today_count: 0,
  success_count: 0,
  client_error_count: 0,
  server_error_count: 0,
  distinct_ip_count: 0,
  avg_duration: 0,
  top_paths: [],
  method_stats: [],
  scene_stats: [],
})

const query = reactive<APIAccessLogListParams>({
  page: 1,
  page_size: 20,
  keyword: '',
  path: '',
  scene: undefined,
  auth_method: undefined,
  transport: undefined,
  method: undefined,
  status_code: undefined,
  start_time: 0,
  end_time: 0,
})

const dateRange = ref<[number, number] | null>(null)
const listFetchGuard = useRequestGuard()
const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
})
const runtimeForm = reactive({
  api_access_log_enabled: true,
  api_log_query_days: 7,
  api_log_max_count: 5000,
  api_log_cleanup_interval_seconds: 600,
  api_log_per_user_limit_enabled: false,
  api_log_per_user_max_count: 1000,
  user_api_log_visible: true,
})

const showDetail = ref(false)
const detailLoading = ref(false)
const detailData = ref<APIAccessLog | null>(null)
/** 详情弹窗内是否解锁展示敏感字段（请求头/请求体/响应体） */
const showSensitiveDetail = ref(false)

const showClean = ref(false)
const cleanBefore = ref<number | null>(null)
const cleaning = ref(false)

const sceneOptions = [
  { label: t('adminAPILogs.all'), value: '' },
  { label: t('adminAPILogs.admin'), value: 'admin' },
  { label: t('adminAPILogs.user'), value: 'user' },
  { label: t('adminAPILogs.public'), value: 'public' },
  { label: t('adminAPILogs.system'), value: 'system' },
  { label: t('adminAPILogs.plugin'), value: 'plugin' },
]

const methodOptions = [
  { label: t('adminAPILogs.all'), value: '' },
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
  { label: 'PATCH', value: 'PATCH' },
]

const transportOptions = [
  { label: t('adminAPILogs.all'), value: '' },
  { label: t('adminAPILogs.transportHttp'), value: 'http' },
  { label: t('adminAPILogs.transportWebsocket'), value: 'websocket' },
  { label: t('adminAPILogs.transportSse'), value: 'sse' },
  { label: t('adminAPILogs.transportStream'), value: 'stream' },
]

const statusOptions = [
  { label: t('adminAPILogs.all'), value: 0 },
  { label: '200', value: 200 },
  { label: '401', value: 401 },
  { label: '403', value: 403 },
  { label: '404', value: 404 },
  { label: '429', value: 429 },
  { label: '500', value: 500 },
]

const formattedRequestHeaders = computed(() => formatPrettyJSON(detailData.value?.request_headers))
const formattedQueryString = computed(() => formatPrettyJSON(detailData.value?.query_string))
const formattedPathParams = computed(() => formatPrettyJSON(detailData.value?.path_params))
const formattedRequestBody = computed(() => formatPrettyJSON(detailData.value?.request_body))
const formattedResponseBody = computed(() => formatPrettyJSON(detailData.value?.response_body))
const topPathItems = computed(() => statsData.value.top_paths.slice(0, 8))
const topPathChartItems = computed(() => [...topPathItems.value].reverse())

function resolveStatusType(statusCode: number) {
  if (statusCode >= 500)
    return 'error'
  if (statusCode >= 400)
    return 'warning'
  return 'success'
}

function resolveTransportLabel(transport?: string) {
  switch ((transport || '').toLowerCase()) {
    case 'websocket':
      return t('adminAPILogs.transportWebsocket')
    case 'sse':
      return t('adminAPILogs.transportSse')
    case 'stream':
      return t('adminAPILogs.transportStream')
    case 'http':
      return t('adminAPILogs.transportHttp')
    default:
      return transport || '-'
  }
}

function resolveTransportTagType(transport?: string) {
  switch ((transport || '').toLowerCase()) {
    case 'websocket':
      return 'primary'
    case 'sse':
      return 'success'
    case 'stream':
      return 'warning'
    case 'http':
      return 'info'
    default:
      return 'default'
  }
}

function formatDuration(duration?: number) {
  if (typeof duration !== 'number' || Number.isNaN(duration))
    return '-'
  return `${duration} ms`
}

function formatByteSize(size?: number) {
  if (typeof size !== 'number' || Number.isNaN(size) || size < 0)
    return '-'
  if (size < 1024)
    return `${size} B`
  if (size < 1024 * 1024)
    return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024)
    return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

function applyDefaultDateRange(days = runtimeForm.api_log_query_days) {
  const end = Date.now()
  dateRange.value = [end - days * 24 * 60 * 60 * 1000, end]
}

function formatTopPathAxisLabel(value: string) {
  const normalized = value || '-'
  return normalized.length > 28 ? `${normalized.slice(0, 28)}...` : normalized
}

const topPathChartOptions = computed<ECOption>(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    confine: true,
    formatter: (params: any) => {
      const index = Array.isArray(params) ? (params[0]?.dataIndex ?? 0) : (params?.dataIndex ?? 0)
      const item = topPathChartItems.value[index]
      if (!item)
        return '-'
      return [
        item.route_path || '-',
        `${t('adminAPILogs.requestCount')}: ${item.count}`,
        `${t('adminAPILogs.avgDuration')}: ${Number(item.avg_duration || 0).toFixed(1)} ms`,
      ].join('<br/>')
    },
  },
  grid: { left: 16, right: 24, top: 12, bottom: 12, containLabel: true },
  xAxis: {
    type: 'value',
    splitLine: { show: true },
  },
  yAxis: {
    type: 'category',
    axisTick: { show: false },
    data: topPathChartItems.value.map(item => item.route_path || '-'),
    axisLabel: {
      formatter: (value: string) => formatTopPathAxisLabel(value),
    },
  },
  series: [
    {
      type: 'bar',
      barMaxWidth: 18,
      label: { show: true, position: 'right' },
      data: topPathChartItems.value.map(item => item.count),
    },
  ],
}))

useEcharts('topPathChartRef', topPathChartOptions)

const columns: DataTableColumns<APIAccessLog> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('adminAPILogs.scene'),
    key: 'scene',
    width: 90,
    render(row) {
      return h(NTag, { size: 'small', type: 'info' }, () => row.scene || '-')
    },
  },
  {
    title: t('adminAPILogs.method'),
    key: 'method',
    width: 90,
    render(row) {
      const type = row.method === 'GET' ? 'info' : row.method === 'POST' ? 'success' : row.method === 'DELETE' ? 'error' : 'warning'
      return h(NTag, { size: 'small', type }, () => row.method)
    },
  },
  {
    title: t('adminAPILogs.transport'),
    key: 'transport',
    width: 110,
    render(row) {
      return h(NTag, { size: 'small', type: resolveTransportTagType(row.transport) as any }, () => resolveTransportLabel(row.transport || 'http'))
    },
  },
  { title: t('adminAPILogs.requestId'), key: 'request_id', width: 220, ellipsis: { tooltip: true } },
  {
    title: t('adminAPILogs.callTarget'),
    key: 'path',
    minWidth: 280,
    render(row) {
      return h(NSpace, { vertical: true, size: 2 }, {
        default: () => [
          h(NText, { style: 'word-break: break-all;' }, () => row.path || '-'),
          row.route_path && row.route_path !== row.path
            ? h(NText, { depth: 3, style: 'font-size: 12px; word-break: break-all;' }, () => `${t('adminAPILogs.routePath')}: ${row.route_path}`)
            : null,
        ],
      })
    },
  },
  { title: t('adminAPILogs.handlerName'), key: 'handler_name', width: 240, ellipsis: { tooltip: true } },
  { title: t('adminAPILogs.username'), key: 'username', width: 120, ellipsis: { tooltip: true } },
  { title: t('adminAPILogs.ip'), key: 'ip', width: 140 },
  {
    title: t('adminAPILogs.statusCode'),
    key: 'status_code',
    width: 90,
    render(row) {
      return h(NTag, { size: 'small', type: resolveStatusType(row.status_code) as any }, () => String(row.status_code))
    },
  },
  {
    title: t('adminAPILogs.duration'),
    key: 'duration',
    width: 110,
    render(row) {
      return formatDuration(row.duration)
    },
  },
  {
    title: t('adminAPILogs.time'),
    key: 'create_time',
    width: 170,
    render(row) {
      return row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-'
    },
  },
  {
    title: t('adminAPILogs.actions'),
    key: 'actions',
    width: 80,
    render(row) {
      return h(NButton, { size: 'small', text: true, type: 'primary', onClick: () => handleDetail(row.request_id || row.id) }, () => t('adminAPILogs.detail'))
    },
  },
]

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'scene', label: t('adminAPILogs.scene') },
  { key: 'method', label: t('adminAPILogs.method') },
  { key: 'transport', label: t('adminAPILogs.transport') },
  { key: 'request_id', label: t('adminAPILogs.requestId') },
  { key: 'path', label: t('adminAPILogs.callTarget') },
  { key: 'handler_name', label: t('adminAPILogs.handlerName') },
  { key: 'username', label: t('adminAPILogs.username') },
  { key: 'ip', label: t('adminAPILogs.ip') },
  { key: 'status_code', label: t('adminAPILogs.statusCode') },
  { key: 'duration', label: t('adminAPILogs.duration') },
  { key: 'create_time', label: t('adminAPILogs.time') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<APIAccessLog>({
  storageKey: 'admin-api-logs-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 960,
})

async function loadRuntimeConfig() {
  runtimeLoading.value = true
  try {
    const res = await adminApi.server.operations()
    const apiLogConfig = res.data?.api_log
    if (apiLogConfig) {
      runtimeForm.api_access_log_enabled = parseBooleanSetting(apiLogConfig.enabled, true)
      runtimeForm.api_log_query_days = normalizeLogQueryDays(apiLogConfig.query_days, 7)
      runtimeForm.api_log_max_count = normalizeLogMaxCount(apiLogConfig.max_count, 5000)
      runtimeForm.api_log_cleanup_interval_seconds = normalizeAPILogCleanupIntervalSeconds(apiLogConfig.cleanup_interval_seconds, 600)
      if (apiLogConfig.per_user_limit_enabled !== undefined)
        runtimeForm.api_log_per_user_limit_enabled = parseBooleanSetting(apiLogConfig.per_user_limit_enabled, false)
      if (apiLogConfig.per_user_max_count !== undefined)
        runtimeForm.api_log_per_user_max_count = normalizeLogPerUserMaxCount(apiLogConfig.per_user_max_count)
    }

    // 服务端 ops 若未返回 per-user 字段，则从系统设置兜底读取
    const settingsRes = await adminApi.settings.list()
    const categories = settingsRes.data?.categories || []
    for (const category of categories) {
      for (const item of category.items) {
        if (item.key === 'api_log_per_user_limit_enabled')
          runtimeForm.api_log_per_user_limit_enabled = parseBooleanSetting(item.value, false)
        if (item.key === 'api_log_per_user_max_count')
          runtimeForm.api_log_per_user_max_count = normalizeLogPerUserMaxCount(item.value)
        if (item.key === 'api_log_cleanup_interval_seconds')
          runtimeForm.api_log_cleanup_interval_seconds = normalizeAPILogCleanupIntervalSeconds(item.value, 600)
        if (item.key === 'user_api_log_visible')
          runtimeForm.user_api_log_visible = parseBooleanSetting(item.value, true)
      }
    }
  }
  catch {
    message.error(t('adminServer.loadRuntimeFailed'))
  }
  finally {
    if (!dateRange.value)
      applyDefaultDateRange(runtimeForm.api_log_query_days)
    runtimeLoading.value = false
  }
}

async function fetchList() {
  const token = listFetchGuard.begin()
  loading.value = true
  try {
    if (dateRange.value) {
      query.start_time = Math.floor(dateRange.value[0] / 1000)
      query.end_time = Math.floor(dateRange.value[1] / 1000)
    }
    else {
      query.start_time = 0
      query.end_time = 0
    }

    const params: Record<string, any> = { ...query }
    if (typeof params.path === 'string')
      params.path = params.path.trim()
    if (!params.keyword)
      delete params.keyword
    if (!params.path)
      delete params.path
    if (!params.scene)
      delete params.scene
    if (!params.transport)
      delete params.transport
    if (!params.method)
      delete params.method
    if (!params.status_code)
      delete params.status_code
    if (!params.start_time)
      delete params.start_time
    if (!params.end_time)
      delete params.end_time

    const res = await adminAPILogApi.list(params)
    if (!listFetchGuard.isLatest(token))
      return
    logList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = total.value
  }
  catch {
    if (listFetchGuard.isLatest(token))
      message.error(t('adminAPILogs.fetchListFailed'))
  }
  finally {
    if (listFetchGuard.isLatest(token))
      loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await adminAPILogApi.stats()
    if (res.data)
      statsData.value = res.data
  }
  catch {}
}

async function handleSaveRuntimeConfig() {
  await withSubmitLock(runtimeSaving, async () => {
    try {
      runtimeForm.api_log_query_days = normalizeLogQueryDays(runtimeForm.api_log_query_days, 7)
      runtimeForm.api_log_max_count = normalizeLogMaxCount(runtimeForm.api_log_max_count, 5000)
      runtimeForm.api_log_cleanup_interval_seconds = normalizeAPILogCleanupIntervalSeconds(runtimeForm.api_log_cleanup_interval_seconds, 600)
      runtimeForm.api_log_per_user_max_count = normalizeLogPerUserMaxCount(runtimeForm.api_log_per_user_max_count)

      const res = await adminApi.settings.batchUpdate({
        api_access_log_enabled: String(runtimeForm.api_access_log_enabled),
        api_log_query_days: String(runtimeForm.api_log_query_days),
        api_log_max_count: String(runtimeForm.api_log_max_count),
        api_log_cleanup_interval_seconds: String(runtimeForm.api_log_cleanup_interval_seconds),
        api_log_per_user_limit_enabled: String(runtimeForm.api_log_per_user_limit_enabled),
        api_log_per_user_max_count: String(runtimeForm.api_log_per_user_max_count),
        user_api_log_visible: String(runtimeForm.user_api_log_visible),
      })
      if (!res.isSuccess)
        throw new Error(res.message || t('adminServer.saveRuntimeFailed'))

      applyDefaultDateRange(runtimeForm.api_log_query_days)
      query.page = 1
      pagination.page = 1
      await Promise.all([fetchList(), fetchStats()])
      message.success(res.message || t('adminServer.saveRuntimeSuccess'))
    }
    catch (error: any) {
      message.error(error?.message || t('adminServer.saveRuntimeFailed'))
    }
  })
}

async function handleDetail(id: number | string) {
  showDetail.value = true
  detailLoading.value = true
  detailData.value = null
  // 每次打开详情默认隐藏敏感正文，由用户主动点击解锁。
  showSensitiveDetail.value = false
  try {
    const res = await adminAPILogApi.detail(id)
    detailData.value = res.data || null
  }
  catch {
    showDetail.value = false
    message.error(t('adminAPILogs.loadDetailFailed'))
  }
  finally {
    detailLoading.value = false
  }
}

/** 用户主动点击“查看敏感详情”后直接展示请求头 / 请求体 / 响应体。 */
function handleRevealSensitiveDetail() {
  showSensitiveDetail.value = true
}

function handleSearch() {
  query.page = 1
  pagination.page = 1
  fetchList()
}

function handleReset() {
  query.keyword = ''
  query.path = ''
  query.scene = undefined
  query.transport = undefined
  query.method = undefined
  query.status_code = undefined
  applyDefaultDateRange(runtimeForm.api_log_query_days)
  handleSearch()
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchList()
}

function handleCleanDateChange(val: number | null) {
  cleanBefore.value = val
}

async function handleClean() {
  if (!cleanBefore.value) {
    message.warning(t('adminAPILogs.selectCleanDate'))
    return
  }
  await withSubmitLock(cleaning, async () => {
    try {
      const beforeTime = Math.floor(cleanBefore.value! / 1000)
      const res = await adminAPILogApi.clean(beforeTime)
      message.success(t('adminAPILogs.cleanSuccess', { count: res.data?.affected || 0 }))
      showClean.value = false
      cleanBefore.value = null
      fetchList()
      fetchStats()
    }
    catch {
      message.error(t('adminAPILogs.cleanFailed'))
    }
  })
}

onMounted(() => {
  loadRuntimeConfig().finally(() => {
    fetchList()
    fetchStats()
  })
})
</script>

<template>
  <div class="api-log-page">
    <NGrid :x-gap="12" :y-gap="12" cols="4" style="margin-bottom: 16px;">
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAPILogs.totalCount')" :value="statsData.total_count" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAPILogs.todayCount')" :value="statsData.today_count" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAPILogs.clientErrors')">
            <template #default>
              <NText type="warning">
                {{ statsData.client_error_count }}
              </NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAPILogs.serverErrors')">
            <template #default>
              <NText type="error">
                {{ statsData.server_error_count }}
              </NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
    </NGrid>

    <NText depth="3" style="display: block; margin: -4px 0 12px;">
      {{ t('adminAPILogs.statsHint') }}
    </NText>

    <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 l:2" responsive="screen" style="margin-bottom: 16px;">
      <NGi>
        <NCard size="small" :title="t('adminAPILogs.topPaths')">
          <div ref="topPathChartRef" class="top-path-chart" />
          <NText v-if="!topPathItems.length" depth="3" style="display: block; text-align: center;">
            {{ t('adminAPILogs.noTopPaths') }}
          </NText>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small" :title="t('adminAPILogs.overview')">
          <NDescriptions :column="2" bordered size="small" label-placement="left">
            <NDescriptionsItem :label="t('adminAPILogs.successCount')">
              {{ statsData.success_count }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.distinctIPs')">
              {{ statsData.distinct_ip_count }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.avgDuration')">
              {{ Number(statsData.avg_duration || 0).toFixed(1) }} ms
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.sceneSummary')">
              {{ statsData.scene_stats.map(item => `${item.scene}:${item.count}`).join(' / ') || '-' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>
      </NGi>
    </NGrid>

    <NCard :title="t('adminAPILogs.title')">
      <template #header-extra>
        <NSpace>
          <TableColumnSelector
            v-model="selectedColumnKeys"
            :options="columnOptions"
            :visible-count="visibleColumnCount"
            :total-count="totalColumnCount"
            :button-label="t('common.showFields')"
            :title="t('common.visibleFields')"
            :hint="t('common.columnVisibilityHint')"
            :reset-label="t('common.restoreDefaultFields')"
            @reset="resetSelectedColumns"
          />
          <NButton size="small" type="primary" @click="fetchList">
            {{ t('adminAPILogs.refresh') }}
          </NButton>
          <NButton size="small" type="warning" @click="showClean = true">
            {{ t('adminAPILogs.cleanLogs') }}
          </NButton>
        </NSpace>
      </template>

      <NCard size="small" embedded style="margin-bottom: 12px;">
        <NSpace align="center" justify="space-between" :wrap="true">
          <NSpace align="center" :wrap="true" size="small">
            <NText strong>
              {{ t('adminServer.runtimeConfig.apiLog') }}
            </NText>
            <NSwitch v-model:value="runtimeForm.api_access_log_enabled" />
            <NText depth="3">
              {{ t('adminAPILogs.userVisible') }}
            </NText>
            <NSwitch v-model:value="runtimeForm.user_api_log_visible" />
            <NText depth="3">
              {{ t('adminServer.runtimeConfig.queryDays') }}
            </NText>
            <NInputNumber v-model:value="runtimeForm.api_log_query_days" :min="1" :max="365" size="small" style="width: 110px;" />
            <NText depth="3">
              {{ t('adminServer.runtimeConfig.maxCount') }}
            </NText>
            <NInputNumber v-model:value="runtimeForm.api_log_max_count" :min="100" :max="200000" size="small" style="width: 130px;" />
            <NText depth="3">
              {{ t('adminServer.runtimeConfig.cleanupInterval') }}
            </NText>
            <NInputNumber v-model:value="runtimeForm.api_log_cleanup_interval_seconds" :min="60" :max="86400" size="small" style="width: 130px;" />
            <NText depth="3">
              {{ t('adminAPILogs.perUserLimitEnabled') }}
            </NText>
            <NSwitch v-model:value="runtimeForm.api_log_per_user_limit_enabled" />
            <NText depth="3">
              {{ t('adminAPILogs.perUserMaxCount') }}
            </NText>
            <NInputNumber v-model:value="runtimeForm.api_log_per_user_max_count" :min="1" :max="200000" size="small" style="width: 130px;" />
          </NSpace>
          <NSpace size="small">
            <NButton size="small" type="primary" :loading="runtimeSaving" @click="handleSaveRuntimeConfig">
              {{ t('adminServer.runtimeConfig.save') }}
            </NButton>
            <NButton size="small" :loading="runtimeLoading" @click="loadRuntimeConfig">
              {{ t('adminAPILogs.refresh') }}
            </NButton>
          </NSpace>
        </NSpace>
        <NText depth="3" style="display: block; margin-top: 8px;">
          {{ t('adminAPILogs.runtimeHint') }}
        </NText>
      </NCard>

      <NSpace align="center" style="margin-bottom: 12px;" :wrap="true">
        <NInput v-model:value="query.keyword" :placeholder="t('adminAPILogs.keywordPlaceholder')" clearable size="small" style="width: 220px;" @keyup.enter="handleSearch" />
        <NInput v-model:value="query.path" :placeholder="t('adminAPILogs.pathPlaceholder')" clearable size="small" style="width: 280px;" @keyup.enter="handleSearch" />
        <NSelect v-model:value="query.scene" :options="sceneOptions" :placeholder="t('adminAPILogs.scene')" clearable size="small" style="width: 120px;" />
        <NSelect v-model:value="query.transport" :options="transportOptions" :placeholder="t('adminAPILogs.transport')" clearable size="small" style="width: 130px;" />
        <NSelect v-model:value="query.method" :options="methodOptions" :placeholder="t('adminAPILogs.method')" clearable size="small" style="width: 110px;" />
        <NSelect v-model:value="query.status_code" :options="statusOptions" :placeholder="t('adminAPILogs.statusCode')" size="small" style="width: 110px;" />
        <NDatePicker v-model:value="dateRange" type="datetimerange" clearable size="small" style="width: 340px;" />
        <NButton size="small" type="primary" @click="handleSearch">
          {{ t('adminAPILogs.search') }}
        </NButton>
        <NButton size="small" @click="handleReset">
          {{ t('adminAPILogs.reset') }}
        </NButton>
      </NSpace>
      <NText depth="3" style="display: block; margin: -4px 0 12px;">
        {{ t('adminAPILogs.transportHint') }}
      </NText>

      <NDataTable
        remote
        :columns="visibleColumns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="tableScrollX"
        :row-key="(row: APIAccessLog) => row.request_id || row.id"
        @update:page="handlePageChange"
      />
    </NCard>

    <NModal v-model:show="showDetail" preset="card" :title="t('adminAPILogs.detailTitle')" style="width: 1100px;" :mask-closable="true">
      <NText v-if="detailLoading" depth="3">
        {{ t('adminAPILogs.loading') }}
      </NText>
      <NSpace v-else-if="detailData" vertical :size="16">
        <NCard size="small" embedded :title="t('adminAPILogs.basicInfo')">
          <NDescriptions bordered :column="2" label-placement="left">
            <NDescriptionsItem :label="t('adminAPILogs.id')">
              {{ detailData.id }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.requestId')">
              {{ detailData.request_id || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.scene')">
              {{ detailData.scene || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.transport')">
              <NTag size="small" :type="resolveTransportTagType(detailData.transport)">
                {{ resolveTransportLabel(detailData.transport || 'http') }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.routePath')">
              {{ detailData.route_path || detailData.path }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.path')">
              {{ detailData.path }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.method')">
              {{ detailData.method }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.protocol')">
              {{ detailData.protocol || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.statusCode')">
              {{ detailData.status_code }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.duration')">
              {{ formatDuration(detailData.duration) }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.username')">
              {{ detailData.username || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.ip')">
              {{ detailData.ip }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.sourceIP')">
              {{ detailData.source_ip || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.handlerName')">
              {{ detailData.handler_name || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.requestContentType')">
              {{ detailData.request_content_type || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.responseContentType')">
              {{ detailData.response_content_type || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.xIP')">
              {{ detailData.x_ip || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.xRealIP')">
              {{ detailData.x_real_ip || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.xForwardedFor')" :span="2">
              {{ detailData.x_forwarded_for || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.requestSize')">
              {{ formatByteSize(detailData.request_size) }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.responseSize')">
              {{ formatByteSize(detailData.response_size) }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.time')">
              {{ detailData.create_time ? new Date(detailData.create_time * 1000).toLocaleString() : '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.referer')">
              {{ detailData.referer || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminAPILogs.userAgent')" :span="2">
              {{ detailData.user_agent || '-' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <NCard size="small" embedded :title="t('adminAPILogs.queryString')">
          <NCode :code="formattedQueryString || '-'" language="json" word-wrap style="max-height: 220px; overflow: auto;" />
        </NCard>

        <NCard size="small" embedded :title="t('adminAPILogs.pathParams')">
          <NCode :code="formattedPathParams || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </NCard>

        <!-- 敏感字段默认隐藏，需「查看敏感详情」二次确认 -->
        <NCard v-if="!showSensitiveDetail" size="small" embedded :title="t('adminAPILogs.sensitiveSection')">
          <NSpace vertical :size="8">
            <NText depth="3">
              {{ t('adminAPILogs.sensitiveHiddenHint') }}
            </NText>
            <NButton type="warning" size="small" @click="handleRevealSensitiveDetail">
              {{ t('adminAPILogs.viewSensitiveDetail') }}
            </NButton>
          </NSpace>
        </NCard>
        <template v-else>
          <NCard size="small" embedded :title="t('adminAPILogs.requestHeaders')">
            <NCode :code="formattedRequestHeaders || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
          </NCard>

          <NCard size="small" embedded :title="t('adminAPILogs.requestBody')">
            <NCode :code="formattedRequestBody || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
          </NCard>

          <NCard size="small" embedded :title="t('adminAPILogs.responseBody')">
            <NCode :code="formattedResponseBody || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
          </NCard>
        </template>
      </NSpace>
      <NText v-else depth="3">
        {{ t('adminAPILogs.noDetailData') }}
      </NText>
    </NModal>

    <NModal v-model:show="showClean" preset="card" :title="t('adminAPILogs.cleanModalTitle')" style="width: 400px;" :mask-closable="false">
      <NSpace vertical>
        <NText>{{ t('adminAPILogs.cleanWarning') }}</NText>
        <NDivider style="margin: 8px 0;" />
        <NText depth="3">
          {{ t('adminAPILogs.cleanBeforeLabel') }}
        </NText>
        <NDatePicker type="datetime" clearable style="width: 100%;" @update:value="handleCleanDateChange" />
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showClean = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="error" :loading="cleaning" :disabled="!cleanBefore" @click="handleClean">
            {{ t('adminAPILogs.confirmClean') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.top-path-chart {
  width: 100%;
  height: 280px;
}
</style>
