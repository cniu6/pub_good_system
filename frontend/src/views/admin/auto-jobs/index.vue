<script setup lang="ts">
/**
 * 自动任务管理器
 * 对照后台 /auto-jobs API：总览卡片 + 全局配置 + 任务定义 / 执行记录 / 当前运行
 */
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NStatistic,
  NSwitch,
  NTabPane,
  NTabs,
  NTag,
  NText,
  useDialog,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRequestGuard } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import type {
  AutoJobDefinition,
  AutoJobGlobalConfig,
  AutoJobOverview,
  AutoJobRun,
  AutoJobUpdateRequest,
} from '@/service/api/admin/auto-job'

const message = useMessage()
const dialog = useDialog()
const runsFetchGuard = useRequestGuard()
const { t } = useI18n()

const loading = ref(false)
const savingConfig = ref(false)
const importing = ref(false)
const activeTab = ref<'jobs' | 'runs' | 'running'>('jobs')

const overview = ref<AutoJobOverview | null>(null)
const config = reactive<AutoJobGlobalConfig>({
  auto_job_enabled: true,
  auto_job_run_max_count: 10000,
  auto_job_retain_errors: true,
  auto_job_auto_prune: true,
  auto_job_stuck_after_sec: 600,
})

const jobList = ref<AutoJobDefinition[]>([])
const runList = ref<AutoJobRun[]>([])
const runTotal = ref(0)
const runningList = ref<AutoJobDefinition[]>([])
const runningTaskCode = ref('')

const jobQuery = reactive({
  keyword: '',
  category: '',
  enabled: '' as '' | '1' | '0',
})

const runQuery = reactive({
  keyword: '',
  status: '',
  category: '',
  job_code: '',
  page: 1,
  page_size: 20,
})

const showEdit = ref(false)
const editSaving = ref(false)
const editForm = reactive<AutoJobUpdateRequest & { job_code?: string }>({
  job_code: '',
  name: '',
  description: '',
  cron_expr: '',
  interval_seconds: 0,
  timezone: 'Asia/Shanghai',
  enabled: true,
  timeout_sec: 60,
  params_json: '{}',
})

const showRunDetail = ref(false)
const runDetail = ref<AutoJobRun | null>(null)

const categoryOptions = computed(() => {
  const set = new Set<string>()
  for (const j of jobList.value) {
    if (j.category)
      set.add(j.category)
  }
  return [
    { label: t('adminAutoJobs.allCategory'), value: '' },
    ...[...set].sort().map(c => ({ label: c, value: c })),
  ]
})

const enabledOptions = [
  { label: t('common.selectPlaceholder'), value: '' },
  { label: t('common.enable'), value: '1' },
  { label: t('common.disable'), value: '0' },
]

const statusOptions = [
  { label: t('adminAutoJobs.selectStatus'), value: '' },
  { label: t('adminAutoJobs.success'), value: 'success' },
  { label: t('adminAutoJobs.failed'), value: 'failed' },
  { label: t('adminAutoJobs.timeout'), value: 'timeout' },
]

function formatTs(sec?: number | null) {
  if (!sec || sec <= 0)
    return '-'
  const d = new Date(sec * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}/${d.getMonth() + 1}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function scheduleText(row: AutoJobDefinition) {
  if (row.cron_expr)
    return row.cron_expr
  if (row.interval_seconds > 0)
    return `${row.interval_seconds}s`
  return '-'
}

function statusTagType(status?: string): 'success' | 'error' | 'warning' | 'info' | 'default' {
  switch (status) {
    case 'success':
      return 'success'
    case 'failed':
    case 'timeout':
      return 'error'
    case 'running':
      return 'info'
    default:
      return 'default'
  }
}

function statusLabel(status?: string) {
  switch (status) {
    case 'success':
      return t('adminAutoJobs.success')
    case 'failed':
      return t('adminAutoJobs.failed')
    case 'running':
      return t('adminAutoJobs.running')
    case 'timeout':
      return t('adminAutoJobs.timeout')
    default:
      return status || '-'
  }
}

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadConfig(), loadJobs()])
    if (activeTab.value === 'runs')
      await loadRuns()
    if (activeTab.value === 'running')
      await loadRunning()
  }
  finally {
    loading.value = false
  }
}

async function loadOverview() {
  try {
    const res = await adminApi.autoJob.overview()
    overview.value = res.data || null
  }
  catch {
    message.error(t('adminAutoJobs.loadFailed'))
  }
}

async function loadConfig() {
  try {
    const res = await adminApi.autoJob.getConfig()
    if (res.data)
      Object.assign(config, res.data)
  }
  catch {
    message.error(t('adminAutoJobs.loadFailed'))
  }
}

async function loadJobs() {
  try {
    const params: { keyword?: string, category?: string, enabled?: string } = {}
    if (jobQuery.keyword)
      params.keyword = jobQuery.keyword
    if (jobQuery.category)
      params.category = jobQuery.category
    if (jobQuery.enabled)
      params.enabled = jobQuery.enabled
    const res = await adminApi.autoJob.listJobs(params)
    jobList.value = res.data?.list || []
  }
  catch {
    message.error(t('adminAutoJobs.loadFailed'))
  }
}

async function loadRuns() {
  const token = runsFetchGuard.begin()
  try {
    const params: Record<string, string | number> = {
      page: runQuery.page,
      page_size: runQuery.page_size,
    }
    if (runQuery.keyword)
      params.keyword = runQuery.keyword
    if (runQuery.status)
      params.status = runQuery.status
    if (runQuery.category)
      params.category = runQuery.category
    if (runQuery.job_code)
      params.job_code = runQuery.job_code
    const res = await adminApi.autoJob.listRuns(params)
    if (!runsFetchGuard.isLatest(token))
      return
    runList.value = res.data?.list || []
    runTotal.value = res.data?.total || 0
  }
  catch {
    if (runsFetchGuard.isLatest(token))
      message.error(t('adminAutoJobs.loadFailed'))
  }
}

async function loadRunning() {
  try {
    const res = await adminApi.autoJob.listRunning()
    runningList.value = res.data?.list || []
  }
  catch {
    message.error(t('adminAutoJobs.loadFailed'))
  }
}

async function handleSaveConfig() {
  savingConfig.value = true
  try {
    const payload = { ...config }
    const res = await adminApi.autoJob.saveConfig(payload)
    if (res.data)
      Object.assign(config, res.data)
    message.success(t('adminAutoJobs.saveSuccess'))
    await loadOverview()
  }
  catch {
    message.error(t('adminAutoJobs.saveFailed'))
  }
  finally {
    savingConfig.value = false
  }
}

async function handleImportPresets() {
  importing.value = true
  try {
    await adminApi.autoJob.importPresets('skip')
    message.success(t('adminAutoJobs.importSuccess'))
    await Promise.all([loadOverview(), loadJobs()])
  }
  catch {
    message.error(t('adminAutoJobs.importFailed'))
  }
  finally {
    importing.value = false
  }
}

async function handleToggleEnabled(row: AutoJobDefinition, enabled: boolean) {
  try {
    if (enabled)
      await adminApi.autoJob.enableJob(row.job_code)
    else
      await adminApi.autoJob.disableJob(row.job_code)
    row.enabled = enabled ? 1 : 0
    await loadOverview()
  }
  catch {
    message.error(t('adminAutoJobs.toggleFailed'))
    await loadJobs()
  }
}

async function handleRunNow(row: AutoJobDefinition) {
  dialog.warning({
    title: t('adminAutoJobs.confirmRunTitle'),
    content: t('adminAutoJobs.confirmRunContent', {
      name: row.name || row.job_code,
      code: row.job_code,
    }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      runningTaskCode.value = row.job_code
      try {
        await adminApi.autoJob.runJob(row.job_code)
        message.success(t('adminAutoJobs.runSuccess'))
        await Promise.all([loadOverview(), loadJobs()])
        if (activeTab.value === 'runs')
          await loadRuns()
        if (activeTab.value === 'running')
          await loadRunning()
      }
      catch {
        message.error(t('adminAutoJobs.runFailed'))
      }
      finally {
        runningTaskCode.value = ''
      }
    },
  })
}

function openEdit(row: AutoJobDefinition) {
  editForm.job_code = row.job_code
  editForm.name = row.name
  editForm.description = row.description
  editForm.cron_expr = row.cron_expr
  editForm.interval_seconds = row.interval_seconds
  editForm.timezone = row.timezone || 'Asia/Shanghai'
  editForm.enabled = row.enabled === 1
  editForm.timeout_sec = row.timeout_sec
  editForm.params_json = row.params_json || '{}'
  showEdit.value = true
}

async function handleSaveEdit() {
  if (!editForm.job_code)
    return
  editSaving.value = true
  try {
    const payload: AutoJobUpdateRequest = {
      name: editForm.name,
      description: editForm.description,
      cron_expr: editForm.cron_expr,
      interval_seconds: editForm.interval_seconds,
      timezone: editForm.timezone,
      enabled: editForm.enabled,
      timeout_sec: editForm.timeout_sec,
      params_json: editForm.params_json,
    }
    await adminApi.autoJob.updateJob(editForm.job_code, payload)
    message.success(t('adminAutoJobs.saveSuccess'))
    showEdit.value = false
    await loadJobs()
  }
  catch {
    message.error(t('adminAutoJobs.saveFailed'))
  }
  finally {
    editSaving.value = false
  }
}

async function openRunDetail(row: AutoJobRun) {
  showRunDetail.value = true
  runDetail.value = null
  try {
    const res = await adminApi.autoJob.getRun(row.id)
    runDetail.value = res.data || row
  }
  catch {
    runDetail.value = row
  }
}

async function handleCleanSuccessRuns() {
  dialog.warning({
    title: t('adminAutoJobs.confirmCleanTitle'),
    content: t('adminAutoJobs.confirmCleanContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await adminApi.autoJob.cleanRuns({ scope: 'success', include_errors: false })
        message.success(t('adminAutoJobs.cleanSuccess', { n: res.data?.affected || 0 }))
        await Promise.all([loadOverview(), loadRuns()])
      }
      catch {
        message.error(t('adminAutoJobs.cleanFailed'))
      }
    },
  })
}

async function handleMarkKeep(row: AutoJobRun) {
  try {
    const res = await adminApi.autoJob.markKeep([row.id], true)
    message.success(t('adminAutoJobs.markKeepSuccess', { n: res.data?.affected || 0 }))
    await loadRuns()
  }
  catch {
    message.error(t('adminAutoJobs.saveFailed'))
  }
}

function onTabChange(name: string) {
  if (name === 'runs')
    loadRuns()
  else if (name === 'running')
    loadRunning()
  else if (name === 'jobs')
    loadJobs()
}

const jobColumns = computed<DataTableColumns<AutoJobDefinition>>(() => [
  { title: t('adminAutoJobs.jobCode'), key: 'job_code', minWidth: 200, ellipsis: { tooltip: true } },
  { title: t('adminAutoJobs.name'), key: 'name', minWidth: 140, ellipsis: { tooltip: true } },
  { title: t('adminAutoJobs.category'), key: 'category', width: 110 },
  {
    title: t('adminAutoJobs.schedule'),
    key: 'schedule',
    width: 140,
    render: row => scheduleText(row),
  },
  {
    title: t('adminAutoJobs.enabled'),
    key: 'enabled',
    width: 90,
    render: row => h(NSwitch, {
      value: row.enabled === 1,
      size: 'small',
      onUpdateValue: (v: boolean) => handleToggleEnabled(row, v),
    }),
  },
  {
    title: t('adminAutoJobs.lastStatus'),
    key: 'last_status',
    width: 100,
    render: row => h(NTag, { size: 'small', type: statusTagType(row.last_status), bordered: false }, {
      default: () => statusLabel(row.last_status),
    }),
  },
  {
    title: t('adminAutoJobs.lastFinished'),
    key: 'last_finished_at',
    width: 170,
    render: row => formatTs(row.last_finished_at),
  },
  {
    title: t('adminAutoJobs.lifetime'),
    key: 'lifetime_run_count',
    width: 90,
    render: row => row.lifetime_run_count || '0',
  },
  {
    title: t('adminAutoJobs.actions'),
    key: 'actions',
    width: 180,
    fixed: 'right',
    render: row => h(NSpace, { size: 8 }, {
      default: () => [
        h(NButton, {
          size: 'small',
          type: 'primary',
          secondary: true,
          loading: runningTaskCode.value === row.job_code,
          onClick: () => handleRunNow(row),
        }, { default: () => t('adminAutoJobs.runNow') }),
        h(NButton, {
          size: 'small',
          onClick: () => openEdit(row),
        }, { default: () => t('adminAutoJobs.edit') }),
      ],
    }),
  },
])

const runColumns = computed<DataTableColumns<AutoJobRun>>(() => [
  { title: t('adminAutoJobs.runId'), key: 'id', width: 80 },
  { title: t('adminAutoJobs.jobCode'), key: 'job_code', minWidth: 180, ellipsis: { tooltip: true } },
  { title: t('adminAutoJobs.category'), key: 'category', width: 100 },
  { title: t('adminAutoJobs.trigger'), key: 'trigger', width: 100 },
  {
    title: t('adminAutoJobs.status'),
    key: 'status',
    width: 100,
    render: row => h(NTag, { size: 'small', type: statusTagType(row.status), bordered: false }, {
      default: () => statusLabel(row.status),
    }),
  },
  {
    title: t('adminAutoJobs.startedAt'),
    key: 'started_at',
    width: 170,
    render: row => formatTs(row.started_at),
  },
  {
    title: t('adminAutoJobs.duration'),
    key: 'duration_ms',
    width: 90,
    render: row => (row.duration_ms > 0 ? `${row.duration_ms}ms` : '-'),
  },
  {
    title: t('adminAutoJobs.message'),
    key: 'message',
    minWidth: 160,
    ellipsis: { tooltip: true },
  },
  {
    title: t('adminAutoJobs.actions'),
    key: 'actions',
    width: 180,
    fixed: 'right',
    render: row => h(NSpace, { size: 8 }, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => openRunDetail(row) }, { default: () => t('adminAutoJobs.detail') }),
        h(NButton, {
          size: 'small',
          secondary: true,
          disabled: row.keep_forever === 1,
          onClick: () => handleMarkKeep(row),
        }, { default: () => t('adminAutoJobs.markKeep') }),
      ],
    }),
  },
])

/** 当前运行：看定义表 last_status=running（执行记录是跑完才写） */
const runningColumns = computed<DataTableColumns<AutoJobDefinition>>(() => [
  { title: t('adminAutoJobs.jobCode'), key: 'job_code', minWidth: 180, ellipsis: { tooltip: true } },
  { title: t('adminAutoJobs.name'), key: 'name', minWidth: 140, ellipsis: { tooltip: true } },
  { title: t('adminAutoJobs.category'), key: 'category', width: 100 },
  {
    title: t('adminAutoJobs.status'),
    key: 'last_status',
    width: 100,
    render: row => h(NTag, { size: 'small', type: statusTagType(row.last_status), bordered: false }, {
      default: () => statusLabel(row.last_status),
    }),
  },
  {
    title: t('adminAutoJobs.startedAt'),
    key: 'last_started_at',
    width: 170,
    render: row => formatTs(row.last_started_at),
  },
  {
    title: t('adminAutoJobs.duration'),
    key: 'elapsed',
    width: 100,
    render: (row) => {
      if (!row.last_started_at || row.last_started_at <= 0)
        return '-'
      const sec = Math.max(0, Math.floor(Date.now() / 1000) - row.last_started_at)
      return `${sec}s`
    },
  },
  {
    title: t('adminAutoJobs.timeoutSec'),
    key: 'timeout_sec',
    width: 90,
  },
])

onMounted(() => {
  refreshAll()
})
</script>

<template>
  <div class="auto-jobs-page">
    <!-- 总览卡片 -->
    <NGrid cols="2 s:3 m:6" responsive="screen" :x-gap="12" :y-gap="12">
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAutoJobs.enabledJobs')" :value="overview?.enabled_jobs ?? 0" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAutoJobs.todaySuccess')" :value="overview?.today_success ?? 0" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAutoJobs.todayFailed')" :value="overview?.today_failed ?? 0" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAutoJobs.running')" :value="overview?.running_count ?? 0" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAutoJobs.recordUsage')">
            <template #default>
              {{ overview?.run_row_count ?? 0 }} / {{ overview?.run_max_count ?? config.auto_job_run_max_count }}
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminAutoJobs.lifetimeRuns')" :value="overview?.lifetime_run_total || '0'" />
        </NCard>
      </NGi>
    </NGrid>

    <!-- 运行配置 -->
    <NCard :title="t('adminAutoJobs.configTitle')" size="small" class="mt-3">
      <NSpace align="center" wrap :size="16">
        <NSpace align="center" :size="8">
          <NText depth="3">
            {{ t('adminAutoJobs.masterSwitch') }}
          </NText>
          <NSwitch v-model:value="config.auto_job_enabled" />
        </NSpace>
        <NSpace align="center" :size="8">
          <NText depth="3">
            {{ t('adminAutoJobs.recordLimit') }}
          </NText>
          <NInputNumber v-model:value="config.auto_job_run_max_count" :min="100" :max="1000000" :step="100" style="width: 140px" />
        </NSpace>
        <NSpace align="center" :size="8">
          <NText depth="3">
            {{ t('adminAutoJobs.autoPrune') }}
          </NText>
          <NSwitch v-model:value="config.auto_job_auto_prune" />
        </NSpace>
        <NSpace align="center" :size="8">
          <NText depth="3">
            {{ t('adminAutoJobs.retainErrors') }}
          </NText>
          <NSwitch v-model:value="config.auto_job_retain_errors" />
        </NSpace>
        <NSpace align="center" :size="8">
          <NText depth="3">
            {{ t('adminAutoJobs.stuckAfterSec') }}
          </NText>
          <NInputNumber v-model:value="config.auto_job_stuck_after_sec" :min="60" :max="86400" :step="60" style="width: 140px" />
        </NSpace>
        <NSpace :size="8">
          <NButton type="primary" :loading="savingConfig" @click="handleSaveConfig">
            {{ t('adminAutoJobs.saveConfig') }}
          </NButton>
          <NButton :loading="loading" @click="refreshAll">
            {{ t('adminAutoJobs.refresh') }}
          </NButton>
          <NButton :loading="importing" @click="handleImportPresets">
            {{ t('adminAutoJobs.importPresets') }}
          </NButton>
        </NSpace>
      </NSpace>
      <div class="scheduler-meta mt-3">
        <NText depth="3">
          {{ t('adminAutoJobs.schedulerRunning') }}:
          {{ overview?.scheduler_running ? t('adminAutoJobs.yes') : t('adminAutoJobs.no') }}
          · {{ t('adminAutoJobs.uptime') }}: {{ overview?.scheduler_uptime_sec ?? 0 }}s
          · {{ t('adminAutoJobs.lastTick') }}: {{ formatTs(overview?.last_tick_at) }}
          · {{ t('adminAutoJobs.defaultTimezone') }}: Asia/Shanghai
        </NText>
      </div>
    </NCard>

    <!-- 任务 / 记录 / 运行中 -->
    <NCard size="small" class="mt-3">
      <NTabs v-model:value="activeTab" type="line" @update:value="onTabChange">
        <NTabPane name="jobs" :tab="t('adminAutoJobs.tabJobs')">
          <NSpace class="mb-3" wrap>
            <NInput
              v-model:value="jobQuery.keyword"
              clearable
              :placeholder="t('adminAutoJobs.keywordPlaceholder')"
              style="width: 220px"
              @keyup.enter="loadJobs"
            />
            <NSelect
              v-model:value="jobQuery.category"
              :options="categoryOptions"
              style="width: 140px"
            />
            <NSelect
              v-model:value="jobQuery.enabled"
              :options="enabledOptions"
              style="width: 140px"
            />
            <NButton type="primary" @click="loadJobs">
              {{ t('adminAutoJobs.search') }}
            </NButton>
          </NSpace>
          <NDataTable
            :columns="jobColumns"
            :data="jobList"
            :loading="loading"
            :scroll-x="1200"
            size="small"
            :bordered="false"
          />
        </NTabPane>

        <NTabPane name="runs" :tab="t('adminAutoJobs.tabRuns')">
          <NSpace class="mb-3" wrap>
            <NInput
              v-model:value="runQuery.keyword"
              clearable
              :placeholder="t('adminAutoJobs.keywordPlaceholder')"
              style="width: 220px"
              @keyup.enter="() => { runQuery.page = 1; loadRuns() }"
            />
            <NSelect v-model:value="runQuery.status" :options="statusOptions" style="width: 140px" />
            <NSelect v-model:value="runQuery.category" :options="categoryOptions" style="width: 140px" />
            <NButton type="primary" @click="() => { runQuery.page = 1; loadRuns() }">
              {{ t('adminAutoJobs.search') }}
            </NButton>
            <NButton @click="handleCleanSuccessRuns">
              {{ t('adminAutoJobs.cleanRuns') }}
            </NButton>
          </NSpace>
          <NDataTable
            :columns="runColumns"
            :data="runList"
            :loading="loading"
            :scroll-x="1200"
            size="small"
            :bordered="false"
            :pagination="{
              page: runQuery.page,
              pageSize: runQuery.page_size,
              itemCount: runTotal,
              showSizePicker: true,
              pageSizes: [20, 50, 100],
              onUpdatePage: (p: number) => { runQuery.page = p; loadRuns() },
              onUpdatePageSize: (ps: number) => { runQuery.page_size = ps; runQuery.page = 1; loadRuns() },
            }"
          />
        </NTabPane>

        <NTabPane name="running" :tab="t('adminAutoJobs.tabRunning')">
          <NDataTable
            v-if="runningList.length"
            :columns="runningColumns"
            :data="runningList"
            :loading="loading"
            :scroll-x="900"
            size="small"
            :bordered="false"
          />
          <NText v-else depth="3">
            {{ t('adminAutoJobs.noRunning') }}
          </NText>
        </NTabPane>
      </NTabs>
    </NCard>

    <!-- 编辑任务 -->
    <NModal
      v-model:show="showEdit"
      preset="card"
      :title="t('adminAutoJobs.editTitle')"
      style="width: 560px"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="100">
        <NFormItem :label="t('adminAutoJobs.jobCode')">
          <NInput :value="editForm.job_code" disabled />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.name')">
          <NInput v-model:value="editForm.name" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.description')">
          <NInput v-model:value="editForm.description" type="textarea" :rows="2" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.cronExpr')">
          <NInput v-model:value="editForm.cron_expr" placeholder="0 0 * * *" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.intervalSeconds')">
          <NInputNumber v-model:value="editForm.interval_seconds" :min="0" style="width: 100%" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.timezone')">
          <NInput v-model:value="editForm.timezone" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.timeoutSec')">
          <NInputNumber v-model:value="editForm.timeout_sec" :min="1" style="width: 100%" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.enabled')">
          <NSwitch v-model:value="editForm.enabled" />
        </NFormItem>
        <NFormItem :label="t('adminAutoJobs.paramsJson')">
          <NInput v-model:value="editForm.params_json" type="textarea" :rows="3" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEdit = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="editSaving" @click="handleSaveEdit">
            {{ t('adminAutoJobs.saveConfig') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 执行详情 -->
    <NModal
      v-model:show="showRunDetail"
      preset="card"
      :title="t('adminAutoJobs.detail')"
      style="width: 640px"
    >
      <template v-if="runDetail">
        <p>
          <NText depth="3">
            ID:
          </NText> {{ runDetail.id }} / {{ runDetail.run_uid }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.jobCode') }}:
          </NText> {{ runDetail.job_code }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.status') }}:
          </NText> {{ statusLabel(runDetail.status) }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.trigger') }}:
          </NText> {{ runDetail.trigger }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.startedAt') }}:
          </NText> {{ formatTs(runDetail.started_at) }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.finishedAt') }}:
          </NText> {{ formatTs(runDetail.finished_at) }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.duration') }}:
          </NText> {{ runDetail.duration_ms }}ms
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.message') }}:
          </NText> {{ runDetail.message || '-' }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.operator') }}:
          </NText> {{ runDetail.operator || '-' }}
        </p>
        <p v-if="runDetail.error_text">
          <NText type="error">
            {{ runDetail.error_text }}
          </NText>
        </p>
        <pre v-if="runDetail.detail_json" class="detail-json">{{ runDetail.detail_json }}</pre>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.auto-jobs-page {
  padding: 4px;
}
.mt-3 {
  margin-top: 12px;
}
.mb-3 {
  margin-bottom: 12px;
}
.scheduler-meta {
  line-height: 1.6;
}
.detail-json {
  margin-top: 8px;
  max-height: 240px;
  overflow: auto;
  padding: 8px;
  border-radius: 6px;
  background: rgba(128, 128, 128, 0.12);
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
}
</style>
