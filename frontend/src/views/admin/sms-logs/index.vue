<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NCode,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NGrid,
  NGi,
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
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useEcharts, useRequestGuard, useTableColumnVisibility } from '@/hooks'
import type { ECOption } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import { adminSMSLogApi, type SMSLog, type SMSLogListParams, type SMSLogStats } from '@/service/api/admin/sms-log'
import { parseBooleanSetting, parseNumberSetting } from '@/utils'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const listFetchGuard = useRequestGuard()

const loading = ref(false)
const runtimeLoading = ref(false)
const runtimeSaving = ref(false)
const logList = ref<SMSLog[]>([])
const total = ref(0)
const statsData = ref<SMSLogStats>({
  total_count: 0,
  today_count: 0,
  success_count: 0,
  fail_count: 0,
  top_templates: [],
  provider_stats: [],
})
const templateNameOptions = ref<{ label: string; value: string }[]>([])

const query = reactive<SMSLogListParams>({
  page: 1,
  page_size: 20,
  phone: '',
  provider: undefined,
  template_name: undefined,
  lang: undefined,
  status: -1,
  start_time: '',
  end_time: '',
})

const dateRange = ref<[number, number] | null>(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
})

const runtimeForm = reactive({
  sms_log_max_count: 1000,
  sms_log_per_user_limit_enabled: false,
  sms_log_per_user_max_count: 1000,
})

const showDetail = ref(false)
const detailData = ref<SMSLog | null>(null)
const detailLoading = ref(false)

const showClean = ref(false)
const cleanBefore = ref('')
const cleaning = ref(false)

const providerOptions = [
  { label: t('adminSMSLogs.all'), value: '' },
  { label: t('adminSMSLogs.aliyun'), value: 'aliyun' },
  { label: t('adminSMSLogs.tencent'), value: 'tencent' },
  { label: t('adminSMSLogs.custom'), value: 'custom' },
  { label: t('adminSMSLogs.console'), value: 'console' },
]

const statusOptions = [
  { label: t('adminSMSLogs.all'), value: -1 },
  { label: t('adminSMSLogs.success'), value: 1 },
  { label: t('adminSMSLogs.failed'), value: 0 },
]

const langOptions = [
  { label: t('adminSMSLogs.all'), value: '' },
  { label: t('adminSMSLogs.chinese'), value: 'zh-CN' },
  { label: t('adminSMSLogs.english'), value: 'en-US' },
]

const providerMap: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'default' }> = {
  aliyun: { label: t('adminSMSLogs.aliyun'), type: 'info' },
  tencent: { label: t('adminSMSLogs.tencent'), type: 'success' },
  custom: { label: t('adminSMSLogs.custom'), type: 'warning' },
  console: { label: t('adminSMSLogs.console'), type: 'default' },
}

const detailStatusText = computed(() => detailData.value?.status === 1 ? t('adminSMSLogs.success') : t('adminSMSLogs.failed'))
const detailStatusType = computed(() => detailData.value?.status === 1 ? 'success' : 'error')
const formattedResponse = computed(() => {
  const raw = detailData.value?.response?.trim()
  if (!raw)
    return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  }
  catch {
    return raw
  }
})

const topTemplateItems = computed(() => (statsData.value.top_templates || []).slice(0, 8))
const topTemplateChartItems = computed(() => [...topTemplateItems.value].reverse())

function normalizeRuntimeMaxCount(value: unknown) {
  return Math.min(200000, Math.max(100, Math.floor(parseNumberSetting(value, 1000))))
}

function normalizePerUserMaxCount(value: unknown) {
  return Math.min(200000, Math.max(1, Math.floor(parseNumberSetting(value, 1000))))
}

function formatTopTemplateAxisLabel(value: string) {
  const normalized = value || '-'
  return normalized.length > 28 ? `${normalized.slice(0, 28)}...` : normalized
}

function goToTemplate(name?: string, lang?: string) {
  if (!name)
    return
  const queryParams: Record<string, string> = { name }
  if (lang)
    queryParams.lang = lang
  router.push({ path: '/settings/sms-templates', query: queryParams })
}

const topTemplateChartOptions = computed<ECOption>(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    confine: true,
    formatter: (params: any) => {
      const index = Array.isArray(params) ? (params[0]?.dataIndex ?? 0) : (params?.dataIndex ?? 0)
      const item = topTemplateChartItems.value[index]
      if (!item)
        return '-'
      return [
        item.template_name || '-',
        `${t('adminSMSLogs.requestCount')}: ${item.count}`,
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
    data: topTemplateChartItems.value.map(item => item.template_name || '-'),
    axisLabel: {
      formatter: (value: string) => formatTopTemplateAxisLabel(value),
    },
  },
  series: [
    {
      type: 'bar',
      barMaxWidth: 18,
      label: { show: true, position: 'right' },
      data: topTemplateChartItems.value.map(item => item.count),
    },
  ],
}))

useEcharts('topTemplateChartRef', topTemplateChartOptions)

async function handleCopyResponse() {
  if (!formattedResponse.value) {
    message.warning(t('adminSMSLogs.noResponseToCopy'))
    return
  }
  try {
    await navigator.clipboard.writeText(formattedResponse.value)
    message.success(t('adminSMSLogs.responseCopied'))
  }
  catch {
    message.error(t('adminSMSLogs.copyFailed'))
  }
}

const columns: DataTableColumns<SMSLog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('adminSMSLogs.phone'), key: 'phone', width: 130 },
  {
    title: t('adminSMSLogs.provider'),
    key: 'provider',
    width: 100,
    render(row) {
      const p = providerMap[row.provider]
      return h(NTag, { type: p?.type || 'default', size: 'small' }, () => p?.label || row.provider)
    },
  },
  { title: t('adminSMSLogs.template'), key: 'template_name', width: 120, ellipsis: { tooltip: true } },
  {
    title: t('adminSMSLogs.lang'),
    key: 'lang',
    width: 80,
    render(row) {
      return row.lang === 'zh-CN' ? t('adminSMSLogs.chinese') : row.lang === 'en-US' ? 'EN' : row.lang
    },
  },
  { title: t('adminSMSLogs.content'), key: 'content', ellipsis: { tooltip: true } },
  {
    title: t('adminSMSLogs.status'),
    key: 'status',
    width: 80,
    render(row) {
      return h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () => row.status === 1 ? t('adminSMSLogs.success') : t('adminSMSLogs.failed'))
    },
  },
  {
    title: t('adminSMSLogs.time'),
    key: 'created_at',
    width: 160,
    render(row) {
      if (!row.created_at)
        return '-'
      return new Date(row.created_at).toLocaleString()
    },
  },
  {
    title: t('adminSMSLogs.actions'),
    key: 'actions',
    width: 80,
    render(row) {
      return h(NButton, { size: 'small', type: 'primary', text: true, onClick: () => handleDetail(row) }, () => t('adminSMSLogs.detail'))
    },
  },
]

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'phone', label: t('adminSMSLogs.phone') },
  { key: 'provider', label: t('adminSMSLogs.provider') },
  { key: 'template_name', label: t('adminSMSLogs.template') },
  { key: 'lang', label: t('adminSMSLogs.lang') },
  { key: 'content', label: t('adminSMSLogs.content') },
  { key: 'status', label: t('adminSMSLogs.status') },
  { key: 'created_at', label: t('adminSMSLogs.time') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<SMSLog>({
  storageKey: 'admin-sms-logs-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 960,
})

async function fetchList() {
  const token = listFetchGuard.begin()
  loading.value = true
  try {
    if (dateRange.value) {
      query.start_time = new Date(dateRange.value[0]).toISOString().slice(0, 19).replace('T', ' ')
      query.end_time = new Date(dateRange.value[1]).toISOString().slice(0, 19).replace('T', ' ')
    }
    else {
      query.start_time = ''
      query.end_time = ''
    }
    const params: any = { ...query }
    if (!params.phone)
      delete params.phone
    if (!params.provider)
      delete params.provider
    if (!params.template_name)
      delete params.template_name
    if (!params.lang)
      delete params.lang
    if (params.status === -1)
      delete params.status
    if (!params.start_time)
      delete params.start_time
    if (!params.end_time)
      delete params.end_time

    const res = await adminSMSLogApi.list(params)
    if (!listFetchGuard.isLatest(token))
      return
    logList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = total.value
  }
  catch {
    if (listFetchGuard.isLatest(token))
      message.error(t('adminSMSLogs.fetchListFailed'))
  }
  finally {
    if (listFetchGuard.isLatest(token))
      loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await adminSMSLogApi.stats()
    if (res.data)
      statsData.value = res.data
  }
  catch { /* ignore */ }
}

async function fetchTemplateNames() {
  try {
    const res = await adminSMSLogApi.templateNames()
    if (res.data) {
      templateNameOptions.value = [
        { label: t('adminSMSLogs.all'), value: '' },
        ...res.data.map(n => ({ label: n, value: n })),
      ]
    }
  }
  catch { /* ignore */ }
}

async function loadRuntimeConfig() {
  runtimeLoading.value = true
  try {
    const res = await adminApi.settings.list()
    const categories = res.data?.categories || []
    for (const category of categories) {
      for (const item of category.items) {
        if (item.key === 'sms_log_max_count')
          runtimeForm.sms_log_max_count = normalizeRuntimeMaxCount(item.value)
        if (item.key === 'sms_log_per_user_limit_enabled')
          runtimeForm.sms_log_per_user_limit_enabled = parseBooleanSetting(item.value, false)
        if (item.key === 'sms_log_per_user_max_count')
          runtimeForm.sms_log_per_user_max_count = normalizePerUserMaxCount(item.value)
      }
    }
  }
  catch {
    message.error(t('adminServer.loadRuntimeFailed'))
  }
  finally {
    runtimeLoading.value = false
  }
}

async function handleSaveRuntimeConfig() {
  runtimeSaving.value = true
  try {
    runtimeForm.sms_log_max_count = normalizeRuntimeMaxCount(runtimeForm.sms_log_max_count)
    runtimeForm.sms_log_per_user_max_count = normalizePerUserMaxCount(runtimeForm.sms_log_per_user_max_count)

    const res = await adminApi.settings.batchUpdate({
      sms_log_max_count: String(runtimeForm.sms_log_max_count),
      sms_log_per_user_limit_enabled: String(runtimeForm.sms_log_per_user_limit_enabled),
      sms_log_per_user_max_count: String(runtimeForm.sms_log_per_user_max_count),
    })
    if (!res.isSuccess)
      throw new Error(res.message || t('adminServer.saveRuntimeFailed'))

    await Promise.all([fetchList(), fetchStats()])
    message.success(res.message || t('adminServer.saveRuntimeSuccess'))
  }
  catch (error: any) {
    message.error(error?.message || t('adminServer.saveRuntimeFailed'))
  }
  finally {
    runtimeSaving.value = false
  }
}

async function handleDetail(row: SMSLog) {
  showDetail.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res = await adminSMSLogApi.detail(row.id)
    detailData.value = res.data || null
    if (!detailData.value)
      message.warning(t('adminSMSLogs.noDetailData'))
  }
  catch {
    showDetail.value = false
    message.error(t('adminSMSLogs.loadDetailFailed'))
  }
  finally {
    detailLoading.value = false
  }
}

function handleSearch() {
  query.page = 1
  pagination.page = 1
  fetchList()
}

function handleReset() {
  query.phone = ''
  query.provider = undefined
  query.template_name = undefined
  query.lang = undefined
  query.status = -1
  dateRange.value = null
  handleSearch()
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchList()
}

async function handleClean() {
  if (!cleanBefore.value) {
    message.warning(t('adminSMSLogs.selectCleanDate'))
    return
  }
  cleaning.value = true
  try {
    const res = await adminSMSLogApi.clean(cleanBefore.value)
    if (res.data) {
      message.success(t('adminSMSLogs.cleanSuccess', { count: res.data.affected }))
      showClean.value = false
      cleanBefore.value = ''
      fetchList()
      fetchStats()
    }
  }
  catch {
    message.error(t('adminSMSLogs.cleanFailed'))
  }
  finally {
    cleaning.value = false
  }
}

function handleCleanDateChange(val: number | null) {
  if (val)
    cleanBefore.value = new Date(val).toISOString().slice(0, 19).replace('T', ' ')
  else
    cleanBefore.value = ''
}

onMounted(() => {
  loadRuntimeConfig()
  fetchList()
  fetchStats()
  fetchTemplateNames()
})
</script>

<template>
  <div class="sms-log-page">
    <NGrid :x-gap="12" :y-gap="12" cols="4" style="margin-bottom: 16px;">
      <NGi><NCard size="small"><NStatistic :label="t('adminSMSLogs.totalCount')" :value="statsData.total_count" /></NCard></NGi>
      <NGi><NCard size="small"><NStatistic :label="t('adminSMSLogs.todayCount')" :value="statsData.today_count" /></NCard></NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminSMSLogs.success')">
            <template #default>
              <NText type="success">{{ statsData.success_count }}</NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminSMSLogs.failed')">
            <template #default>
              <NText type="error">{{ statsData.fail_count }}</NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
    </NGrid>

    <NText depth="3" style="display: block; margin: -4px 0 12px;">{{ t('adminSMSLogs.statsHint') }}</NText>

    <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 l:2" responsive="screen" style="margin-bottom: 16px;">
      <NGi>
        <NCard size="small" :title="t('adminSMSLogs.topTemplates')">
          <div ref="topTemplateChartRef" class="top-path-chart"></div>
          <NText v-if="!topTemplateItems.length" depth="3" style="display: block; text-align: center;">{{ t('adminSMSLogs.noTopTemplates') }}</NText>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small" :title="t('adminSMSLogs.overview')">
          <NDescriptions :column="2" bordered size="small" label-placement="left">
            <NDescriptionsItem :label="t('adminSMSLogs.success')">{{ statsData.success_count }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.failed')">{{ statsData.fail_count }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.providerSummary')" :span="2">
              {{ (statsData.provider_stats || []).map(item => `${providerMap[item.provider]?.label || item.provider}:${item.count}`).join(' / ') || '-' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>
      </NGi>
    </NGrid>

    <NCard :title="t('adminSMSLogs.title')">
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
          <NButton size="small" type="primary" :loading="loading" @click="fetchList">{{ t('adminSMSLogs.refresh') }}</NButton>
          <NButton size="small" type="warning" @click="showClean = true">{{ t('adminSMSLogs.cleanLogs') }}</NButton>
        </NSpace>
      </template>

      <NCard size="small" embedded style="margin-bottom: 12px;">
        <NSpace align="center" justify="space-between" :wrap="true">
          <NSpace align="center" :wrap="true" size="small">
            <NText strong>{{ t('adminSMSLogs.runtimeConfig') }}</NText>
            <NText depth="3">{{ t('adminSMSLogs.maxCount') }}</NText>
            <NInputNumber v-model:value="runtimeForm.sms_log_max_count" :min="100" :max="200000" size="small" style="width: 130px;" />
            <NText depth="3">{{ t('adminSMSLogs.perUserLimitEnabled') }}</NText>
            <NSwitch v-model:value="runtimeForm.sms_log_per_user_limit_enabled" />
            <NText depth="3">{{ t('adminSMSLogs.perUserMaxCount') }}</NText>
            <NInputNumber v-model:value="runtimeForm.sms_log_per_user_max_count" :min="1" :max="200000" size="small" style="width: 130px;" />
          </NSpace>
          <NSpace size="small">
            <NButton size="small" type="primary" :loading="runtimeSaving" @click="handleSaveRuntimeConfig">{{ t('adminServer.runtimeConfig.save') }}</NButton>
            <NButton size="small" :loading="runtimeLoading" @click="loadRuntimeConfig">{{ t('adminSMSLogs.refresh') }}</NButton>
          </NSpace>
        </NSpace>
        <NText depth="3" style="display: block; margin-top: 8px;">{{ t('adminSMSLogs.runtimeHint') }}</NText>
      </NCard>

      <NSpace align="center" style="margin-bottom: 12px;" :wrap="true">
        <NInput v-model:value="query.phone" :placeholder="t('adminSMSLogs.phone')" clearable size="small" style="width: 140px;" @keyup.enter="handleSearch" />
        <NSelect v-model:value="query.provider" :options="providerOptions" :placeholder="t('adminSMSLogs.provider')" clearable size="small" style="width: 120px;" />
        <NSelect v-model:value="query.template_name" :options="templateNameOptions" :placeholder="t('adminSMSLogs.template')" clearable size="small" style="width: 140px;" />
        <NSelect v-model:value="query.lang" :options="langOptions" :placeholder="t('adminSMSLogs.lang')" clearable size="small" style="width: 100px;" />
        <NSelect v-model:value="query.status" :options="statusOptions" :placeholder="t('adminSMSLogs.status')" size="small" style="width: 90px;" />
        <NDatePicker v-model:value="dateRange" type="datetimerange" clearable size="small" style="width: 340px;" />
        <NButton size="small" type="primary" @click="handleSearch">{{ t('adminSMSLogs.search') }}</NButton>
        <NButton size="small" @click="handleReset">{{ t('adminSMSLogs.reset') }}</NButton>
      </NSpace>

      <NDataTable
        remote
        :columns="visibleColumns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="tableScrollX"
        :row-key="(row: SMSLog) => row.id"
        @update:page="handlePageChange"
      />
    </NCard>

    <NModal v-model:show="showDetail" preset="card" :title="t('adminSMSLogs.detailTitle')" style="width: 760px;" :mask-closable="true">
      <NText v-if="detailLoading" depth="3">{{ t('adminSMSLogs.loading') }}</NText>
      <NSpace v-else-if="detailData" vertical :size="16">
        <NGrid cols="2" :x-gap="12" :y-gap="12">
          <NGi>
            <NCard size="small" embedded>
              <NStatistic :label="t('adminSMSLogs.logId')" :value="detailData.id" />
            </NCard>
          </NGi>
          <NGi>
            <NCard size="small" embedded>
              <NStatistic :label="t('adminSMSLogs.sendStatus')">
                <template #default>
                  <NTag :type="detailStatusType" size="small">{{ detailStatusText }}</NTag>
                </template>
              </NStatistic>
            </NCard>
          </NGi>
        </NGrid>

        <NCard size="small" embedded :title="t('adminSMSLogs.basicInfo')">
          <NDescriptions bordered :column="2" label-placement="left">
            <NDescriptionsItem :label="t('adminSMSLogs.phone')">{{ detailData.phone }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.provider')">
              <NTag :type="providerMap[detailData.provider]?.type || 'default'" size="small">
                {{ providerMap[detailData.provider]?.label || detailData.provider }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.templateId')">{{ detailData.template_code || '-' }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.templateName')">
              <NSpace align="center" :size="8">
                <span>{{ detailData.template_name || '-' }}</span>
                <NButton
                  v-if="detailData.template_name"
                  size="tiny"
                  text
                  type="primary"
                  @click="goToTemplate(detailData.template_name, detailData.lang)"
                >
                  {{ t('adminSMSLogs.viewTemplate') }}
                </NButton>
              </NSpace>
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.lang')">{{ detailData.lang || '-' }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.requestId')">{{ detailData.request_id || '-' }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminSMSLogs.sendTime')" :span="2">
              {{ detailData.created_at ? new Date(detailData.created_at).toLocaleString() : '-' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <NCard size="small" embedded :title="t('adminSMSLogs.content')">
          <NCode :code="detailData.content || '-'" word-wrap style="max-height: 280px; overflow: auto;" />
        </NCard>

        <NCard v-if="detailData.error_msg" size="small" embedded :title="t('adminSMSLogs.errorMsg')">
          <NText type="error">{{ detailData.error_msg }}</NText>
        </NCard>

        <NCard v-if="formattedResponse" size="small" embedded :title="t('adminSMSLogs.fullResponse')">
          <template #header-extra>
            <NButton size="small" quaternary @click="handleCopyResponse">{{ t('adminSMSLogs.copyContent') }}</NButton>
          </template>
          <NCode :code="formattedResponse" language="json" word-wrap style="max-height: 320px; overflow: auto;" />
        </NCard>
      </NSpace>
      <NText v-else depth="3">{{ t('adminSMSLogs.noDetailData') }}</NText>
    </NModal>

    <NModal v-model:show="showClean" preset="card" :title="t('adminSMSLogs.cleanModalTitle')" style="width: 400px;" :mask-closable="false">
      <NSpace vertical>
        <NText>{{ t('adminSMSLogs.cleanWarning') }}</NText>
        <NDivider style="margin: 8px 0;" />
        <NText depth="3">{{ t('adminSMSLogs.cleanBeforeLabel') }}</NText>
        <NDatePicker type="datetime" clearable style="width: 100%;" @update:value="handleCleanDateChange" />
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showClean = false">{{ t('common.cancel') }}</NButton>
          <NButton type="error" :loading="cleaning" :disabled="!cleanBefore" @click="handleClean">{{ t('adminSMSLogs.confirmClean') }}</NButton>
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
