<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NCode,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NGi,
  NGrid,
  NInputNumber,
  NModal,
  NSpace,
  NStatistic,
  NSwitch,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useEcharts, useTableColumnVisibility } from '@/hooks'
import type { ECOption } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import { adminLogApi, type OperationLogStats } from '@/service/api/admin/log'
import type { UserSimpleInfo } from '@/service/api/admin/user'
import { parseBooleanSetting, parseNumberSetting } from '@/utils'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const runtimeLoading = ref(false)
const runtimeSaving = ref(false)
const logList = ref<any[]>([])
const userMap = ref<Record<number, UserSimpleInfo>>({})
const total = ref(0)
const statsData = ref<OperationLogStats>({
  total_count: 0,
  today_count: 0,
  success_count: 0,
  client_error_count: 0,
  server_error_count: 0,
  avg_duration: 0,
  top_modules: [],
  top_actions: [],
  method_stats: [],
})

const showDetail = ref(false)
const detailLoading = ref(false)
const detailData = ref<any | null>(null)

const query = reactive({
  page: 1,
  page_size: 20,
  start_time: 0,
  end_time: 0,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
})

const runtimeForm = reactive({
  operation_log_query_days: 30,
  operation_log_max_count: 1000,
  operation_log_per_user_limit_enabled: false,
  operation_log_per_user_max_count: 1000,
})

const methodColors: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  GET: 'info',
  POST: 'success',
  PUT: 'warning',
  DELETE: 'error',
}

const topModuleItems = computed(() => (statsData.value.top_modules || []).slice(0, 8))
const topModuleChartItems = computed(() => [...topModuleItems.value].reverse())

function goToUserDetail(userId: number) {
  if (userId)
    router.push(`/users/${userId}`)
}

function getUserDisplayName(userId: number): string {
  const user = userMap.value[userId]
  if (!user)
    return t('adminLogs.userPrefix', { id: userId })
  return user.nickname || user.username || t('adminLogs.userPrefix', { id: userId })
}

function formatPayload(raw?: string) {
  const value = raw?.trim()
  if (!value)
    return ''
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  }
  catch {
    return value
  }
}

const formattedRequestBody = computed(() => formatPayload(detailData.value?.request_body))
const formattedResponseBody = computed(() => formatPayload(detailData.value?.response_body))

function normalizeRuntimeQueryDays(value: unknown) {
  return Math.min(365, Math.max(1, Math.floor(parseNumberSetting(value, 30))))
}

function normalizeRuntimeMaxCount(value: unknown) {
  return Math.min(200000, Math.max(100, Math.floor(parseNumberSetting(value, 1000))))
}

function normalizePerUserMaxCount(value: unknown) {
  return Math.min(200000, Math.max(1, Math.floor(parseNumberSetting(value, 1000))))
}

function applyDateRange(days = runtimeForm.operation_log_query_days) {
  const now = Math.floor(Date.now() / 1000)
  const safeDays = Math.max(1, Math.floor(days || 1))
  query.end_time = now
  query.start_time = now - safeDays * 24 * 60 * 60
}

function formatTopModuleAxisLabel(value: string) {
  const normalized = value || '-'
  return normalized.length > 28 ? `${normalized.slice(0, 28)}...` : normalized
}

const topModuleChartOptions = computed<ECOption>(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    confine: true,
    formatter: (params: any) => {
      const index = Array.isArray(params) ? (params[0]?.dataIndex ?? 0) : (params?.dataIndex ?? 0)
      const item = topModuleChartItems.value[index]
      if (!item)
        return '-'
      return [
        item.module || '-',
        `${t('adminLogs.requestCount')}: ${item.count}`,
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
    data: topModuleChartItems.value.map(item => item.module || '-'),
    axisLabel: {
      formatter: (value: string) => formatTopModuleAxisLabel(value),
    },
  },
  series: [
    {
      type: 'bar',
      barMaxWidth: 18,
      label: { show: true, position: 'right' },
      data: topModuleChartItems.value.map(item => item.count),
    },
  ],
}))

useEcharts('topModuleChartRef', topModuleChartOptions)

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminLogs.user'),
    key: 'user_id',
    width: 120,
    render(row) {
      const userId = row.user_id
      if (!userId)
        return '-'
      const displayName = getUserDisplayName(userId)
      return h(
        NButton,
        {
          text: true,
          type: 'primary',
          onClick: () => goToUserDetail(userId),
        },
        { default: () => displayName },
      )
    },
  },
  { title: t('adminLogs.module'), key: 'module', width: 100 },
  { title: t('adminLogs.action'), key: 'action', width: 80 },
  {
    title: t('adminLogs.method'),
    key: 'method',
    width: 80,
    render(row) {
      return h(NTag, { type: methodColors[row.method] ?? 'info', size: 'small' }, () => row.method)
    },
  },
  { title: t('adminLogs.path'), key: 'path', ellipsis: { tooltip: true } },
  { title: t('adminLogs.ip'), key: 'ip', width: 120 },
  { title: t('adminLogs.duration'), key: 'duration', width: 90 },
  {
    title: t('adminLogs.time'),
    key: 'create_time',
    width: 160,
    render(row) {
      if (!row.create_time)
        return '-'
      return new Date(row.create_time * 1000).toLocaleString()
    },
  },
  {
    title: t('adminLogs.actions'),
    key: 'actions',
    width: 80,
    render(row) {
      return h(
        NButton,
        {
          text: true,
          type: 'primary',
          onClick: () => handleDetail(row.id),
        },
        { default: () => t('adminLogs.detail') },
      )
    },
  },
]

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'user_id', label: t('adminLogs.user') },
  { key: 'module', label: t('adminLogs.module') },
  { key: 'action', label: t('adminLogs.action') },
  { key: 'method', label: t('adminLogs.method') },
  { key: 'path', label: t('adminLogs.path') },
  { key: 'ip', label: t('adminLogs.ip') },
  { key: 'duration', label: t('adminLogs.duration') },
  { key: 'create_time', label: t('adminLogs.time') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility({
  storageKey: 'admin-operation-logs-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 960,
})

async function handleDetail(id: number) {
  showDetail.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res = await adminLogApi.detail(id)
    detailData.value = res.data || null
  }
  catch {
    showDetail.value = false
    message.error(t('adminLogs.loadDetailFailed'))
  }
  finally {
    detailLoading.value = false
  }
}

async function fetchUserInfos(logs: any[]) {
  const userIds = [...new Set(logs.map(log => log.user_id).filter(Boolean))]
  if (userIds.length === 0)
    return

  try {
    userMap.value = await adminApi.user.batchSimpleInfo(userIds)
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminLogs] fetch user info failed', error)
  }
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await adminLogApi.list(query)
    logList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = res.data?.total || 0
    await fetchUserInfos(logList.value)
  }
  catch {
    message.error(t('adminLogs.fetchLogsFailed'))
  }
  finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await adminLogApi.stats()
    if (res.data)
      statsData.value = res.data
  }
  catch {}
}

/** 从系统设置读取运行时配置 */
async function loadRuntimeConfig() {
  runtimeLoading.value = true
  try {
    const res = await adminApi.settings.list()
    const categories = res.data?.categories || []
    for (const category of categories) {
      for (const item of category.items) {
        if (item.key === 'operation_log_query_days')
          runtimeForm.operation_log_query_days = normalizeRuntimeQueryDays(item.value)
        if (item.key === 'operation_log_max_count')
          runtimeForm.operation_log_max_count = normalizeRuntimeMaxCount(item.value)
        if (item.key === 'operation_log_per_user_limit_enabled')
          runtimeForm.operation_log_per_user_limit_enabled = parseBooleanSetting(item.value, false)
        if (item.key === 'operation_log_per_user_max_count')
          runtimeForm.operation_log_per_user_max_count = normalizePerUserMaxCount(item.value)
      }
    }
  }
  catch {
    message.error(t('adminServer.loadRuntimeFailed'))
  }
  finally {
    applyDateRange(runtimeForm.operation_log_query_days)
    runtimeLoading.value = false
  }
}

async function handleSaveRuntimeConfig() {
  runtimeSaving.value = true
  try {
    runtimeForm.operation_log_query_days = normalizeRuntimeQueryDays(runtimeForm.operation_log_query_days)
    runtimeForm.operation_log_max_count = normalizeRuntimeMaxCount(runtimeForm.operation_log_max_count)
    runtimeForm.operation_log_per_user_max_count = normalizePerUserMaxCount(runtimeForm.operation_log_per_user_max_count)

    const res = await adminApi.settings.batchUpdate({
      operation_log_query_days: String(runtimeForm.operation_log_query_days),
      operation_log_max_count: String(runtimeForm.operation_log_max_count),
      operation_log_per_user_limit_enabled: String(runtimeForm.operation_log_per_user_limit_enabled),
      operation_log_per_user_max_count: String(runtimeForm.operation_log_per_user_max_count),
    })
    if (!res.isSuccess)
      throw new Error(res.message || t('adminServer.saveRuntimeFailed'))

    query.page = 1
    pagination.page = 1
    applyDateRange(runtimeForm.operation_log_query_days)
    await Promise.all([fetchLogs(), fetchStats()])
    message.success(res.message || t('adminServer.saveRuntimeSuccess'))
  }
  catch (error: any) {
    message.error(error?.message || t('adminServer.saveRuntimeFailed'))
  }
  finally {
    runtimeSaving.value = false
  }
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchLogs()
}

onMounted(() => {
  loadRuntimeConfig().finally(() => {
    fetchLogs()
    fetchStats()
  })
})
</script>

<template>
  <div class="operation-log-page">
    <NGrid :x-gap="12" :y-gap="12" cols="4" style="margin-bottom: 16px;">
      <NGi><NCard size="small"><NStatistic :label="t('adminLogs.totalCount')" :value="statsData.total_count" /></NCard></NGi>
      <NGi><NCard size="small"><NStatistic :label="t('adminLogs.todayCount')" :value="statsData.today_count" /></NCard></NGi>
      <NGi><NCard size="small"><NStatistic :label="t('adminLogs.clientErrors')"><template #default><NText type="warning">{{ statsData.client_error_count }}</NText></template></NStatistic></NCard></NGi>
      <NGi><NCard size="small"><NStatistic :label="t('adminLogs.serverErrors')"><template #default><NText type="error">{{ statsData.server_error_count }}</NText></template></NStatistic></NCard></NGi>
    </NGrid>

    <NText depth="3" style="display: block; margin: -4px 0 12px;">{{ t('adminLogs.statsHint') }}</NText>

    <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 l:2" responsive="screen" style="margin-bottom: 16px;">
      <NGi>
        <NCard size="small" :title="t('adminLogs.topModules')">
          <div ref="topModuleChartRef" class="top-path-chart"></div>
          <NText v-if="!topModuleItems.length" depth="3" style="display: block; text-align: center;">{{ t('adminLogs.noTopModules') }}</NText>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small" :title="t('adminLogs.overview')">
          <NDescriptions :column="2" bordered size="small" label-placement="left">
            <NDescriptionsItem :label="t('adminLogs.successCount')">{{ statsData.success_count }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminLogs.avgDuration')">{{ Number(statsData.avg_duration || 0).toFixed(1) }} ms</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminLogs.methodSummary')" :span="2">
              {{ (statsData.method_stats || []).map(item => `${item.method}:${item.count}`).join(' / ') || '-' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>
      </NGi>
    </NGrid>

    <NCard :title="t('adminLogs.title')">
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
          <NButton size="small" type="primary" :loading="loading" @click="fetchLogs">{{ t('adminLogs.refresh') }}</NButton>
        </NSpace>
      </template>

      <NCard size="small" embedded style="margin-bottom: 12px;">
        <NSpace align="center" justify="space-between" :wrap="true">
          <NSpace align="center" :wrap="true" size="small">
            <NText strong>{{ t('adminLogs.runtimeConfig') }}</NText>
            <NText depth="3">{{ t('adminLogs.queryDays') }}</NText>
            <NInputNumber v-model:value="runtimeForm.operation_log_query_days" :min="1" :max="365" size="small" style="width: 110px;" />
            <NText depth="3">{{ t('adminLogs.maxCount') }}</NText>
            <NInputNumber v-model:value="runtimeForm.operation_log_max_count" :min="100" :max="200000" size="small" style="width: 130px;" />
            <NText depth="3">{{ t('adminLogs.perUserLimitEnabled') }}</NText>
            <NSwitch v-model:value="runtimeForm.operation_log_per_user_limit_enabled" />
            <NText depth="3">{{ t('adminLogs.perUserMaxCount') }}</NText>
            <NInputNumber v-model:value="runtimeForm.operation_log_per_user_max_count" :min="1" :max="200000" size="small" style="width: 130px;" />
          </NSpace>
          <NSpace size="small">
            <NButton size="small" type="primary" :loading="runtimeSaving" @click="handleSaveRuntimeConfig">{{ t('adminServer.runtimeConfig.save') }}</NButton>
            <NButton size="small" :loading="runtimeLoading" @click="loadRuntimeConfig">{{ t('adminLogs.refresh') }}</NButton>
          </NSpace>
        </NSpace>
        <NText depth="3" style="display: block; margin-top: 8px;">{{ t('adminLogs.runtimeHint') }}</NText>
      </NCard>

      <NText depth="3" style="display: block; margin-bottom: 12px;">{{ t('adminLogs.totalLogs', { total }) }}</NText>

      <NDataTable
        remote
        :columns="visibleColumns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="tableScrollX"
        @update:page="handlePageChange"
      />
    </NCard>

    <NModal v-model:show="showDetail" preset="card" :title="t('adminLogs.detailTitle')" style="width: 860px;" :mask-closable="true">
      <NText v-if="detailLoading" depth="3">{{ t('adminLogs.loading') }}</NText>
      <NSpace v-else-if="detailData" vertical :size="16">
        <NDescriptions bordered :column="2" label-placement="left">
          <NDescriptionsItem :label="t('adminLogs.module')">{{ detailData.module || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.action')">{{ detailData.action || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.method')">{{ detailData.method || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.path')">{{ detailData.path || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.handlerName')" :span="2">{{ detailData.handler_name || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.ip')">{{ detailData.ip || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.duration')">{{ detailData.duration || 0 }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.statusCode')">{{ detailData.status_code || '-' }}</NDescriptionsItem>
          <NDescriptionsItem :label="t('adminLogs.time')">{{ detailData.create_time ? new Date(detailData.create_time * 1000).toLocaleString() : '-' }}</NDescriptionsItem>
        </NDescriptions>

        <NCard size="small" embedded :title="t('adminLogs.requestBody')">
          <NCode :code="formattedRequestBody || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </NCard>

        <NCard size="small" embedded :title="t('adminLogs.responseBody')">
          <NCode :code="formattedResponseBody || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </NCard>
      </NSpace>
      <NText v-else depth="3">{{ t('adminLogs.noDetailData') }}</NText>
    </NModal>
  </div>
</template>

<style scoped>
.top-path-chart {
  width: 100%;
  height: 280px;
}
</style>
