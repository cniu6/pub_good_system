<script setup lang="ts">
/**
 * 保留记录：被标记为永久保留的执行记录副本，独立表持久化，支持搜索。
 */
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NDataTable, NDatePicker, NInput, NSelect, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRequestGuard } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import type { AutoJobRunKeep } from '@/service/api/admin/auto-job'

const { t } = useI18n()
const message = useMessage()
const fetchGuard = useRequestGuard()

const loading = ref(false)
const keptList = ref<AutoJobRunKeep[]>([])
const keptTotal = ref(0)

const query = reactive({
  keyword: '',
  status: '',
  category: '',
  job_code: '',
  start_time: null as number | null,
  end_time: null as number | null,
  page: 1,
  page_size: 20,
})

const showDetail = ref(false)
const detail = ref<AutoJobRunKeep | null>(null)

const statusOptions = [
  { label: t('adminAutoJobs.selectStatus'), value: '' },
  { label: t('adminAutoJobs.success'), value: 'success' },
  { label: t('adminAutoJobs.failed'), value: 'failed' },
  { label: t('adminAutoJobs.timeout'), value: 'timeout' },
]

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

function formatTs(sec?: number | null) {
  if (!sec || sec <= 0)
    return '-'
  const d = new Date(sec * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}/${d.getMonth() + 1}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

async function loadKeptRuns() {
  const token = fetchGuard.begin()
  loading.value = true
  try {
    const params: Record<string, string | number | undefined> = {
      page: query.page,
      page_size: query.page_size,
    }
    if (query.keyword)
      params.keyword = query.keyword
    if (query.status)
      params.status = query.status
    if (query.category)
      params.category = query.category
    if (query.job_code)
      params.job_code = query.job_code
    if (query.start_time)
      params.start_time = Math.floor(query.start_time / 1000)
    if (query.end_time)
      params.end_time = Math.floor(query.end_time / 1000)
    const res = await adminApi.autoJob.listKeptRuns(params)
    if (!fetchGuard.isLatest(token))
      return
    keptList.value = res.data?.list || []
    keptTotal.value = res.data?.total || 0
  }
  catch {
    if (fetchGuard.isLatest(token))
      message.error(t('adminAutoJobs.loadFailed'))
  }
  finally {
    if (fetchGuard.isLatest(token))
      loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  loadKeptRuns()
}

function handleReset() {
  query.keyword = ''
  query.status = ''
  query.category = ''
  query.job_code = ''
  query.start_time = null
  query.end_time = null
  query.page = 1
  loadKeptRuns()
}

function openDetail(row: AutoJobRunKeep) {
  detail.value = row
  showDetail.value = true
}

const columns = computed<DataTableColumns<AutoJobRunKeep>>(() => [
  { title: t('adminAutoJobs.runId'), key: 'id', width: 80 },
  { title: t('adminAutoJobs.sourceRunId'), key: 'source_run_id', width: 100 },
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
    title: t('adminAutoJobs.runAt'),
    key: 'run_timestamp',
    width: 170,
    render: row => formatTs(row.run_timestamp),
  },
  {
    title: t('adminAutoJobs.keptAt'),
    key: 'kept_at',
    width: 170,
    render: row => formatTs(row.kept_at),
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
    width: 120,
    fixed: 'right',
    render: row => h(NButton, { size: 'small', onClick: () => openDetail(row) }, { default: () => t('adminAutoJobs.detail') }),
  },
])

onMounted(() => {
  loadKeptRuns()
})
</script>

<template>
  <div>
    <NSpace class="mb-3" wrap>
      <NInput
        v-model:value="query.keyword"
        clearable
        :placeholder="t('adminAutoJobs.keywordPlaceholder')"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <NInput
        v-model:value="query.job_code"
        clearable
        :placeholder="t('adminAutoJobs.jobCode')"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <NInput
        v-model:value="query.category"
        clearable
        :placeholder="t('adminAutoJobs.category')"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <NSelect
        v-model:value="query.status"
        :options="statusOptions"
        :placeholder="t('adminAutoJobs.selectStatus')"
        clearable
        style="width: 130px"
        @update:value="handleSearch"
      />
      <NDatePicker
        v-model:value="query.start_time"
        type="datetime"
        clearable
        :placeholder="t('adminAutoJobs.startTime')"
        @update:value="handleSearch"
      />
      <NDatePicker
        v-model:value="query.end_time"
        type="datetime"
        clearable
        :placeholder="t('adminAutoJobs.endTime')"
        @update:value="handleSearch"
      />
      <NButton type="primary" @click="handleSearch">
        {{ t('adminAutoJobs.search') }}
      </NButton>
      <NButton @click="handleReset">
        {{ t('common.reset') }}
      </NButton>
    </NSpace>

    <NDataTable
      :columns="columns"
      :data="keptList"
      :loading="loading"
      :scroll-x="1300"
      size="small"
      :bordered="false"
      remote
      :pagination="{
        page: query.page,
        pageSize: query.page_size,
        itemCount: keptTotal,
        showSizePicker: true,
        pageSizes: [20, 50, 100],
        onUpdatePage: (p: number) => { query.page = p; loadKeptRuns() },
        onUpdatePageSize: (ps: number) => { query.page_size = ps; query.page = 1; loadKeptRuns() },
      }"
    />

    <NModal
      v-model:show="showDetail"
      preset="card"
      :title="t('adminAutoJobs.detail')"
      style="width: 640px"
    >
      <template v-if="detail">
        <p>
          <NText depth="3">
            ID:
          </NText> {{ detail.id }} / {{ detail.source_run_id }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.runUID') }}:
          </NText> {{ detail.run_uid }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.jobCode') }}:
          </NText> {{ detail.job_code }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.status') }}:
          </NText> {{ statusLabel(detail.status) }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.trigger') }}:
          </NText> {{ detail.trigger }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.runAt') }}:
          </NText> {{ formatTs(detail.run_timestamp) }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.keptAt') }}:
          </NText> {{ formatTs(detail.kept_at) }}
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.duration') }}:
          </NText> {{ detail.duration_ms }}ms
        </p>
        <p>
          <NText depth="3">
            {{ t('adminAutoJobs.operator') }}:
          </NText> {{ detail.operator || '-' }}
        </p>
        <p v-if="detail.error_text">
          <NText type="error">
            {{ detail.error_text }}
          </NText>
        </p>
        <pre v-if="detail.detail_json" class="detail-json">{{ detail.detail_json }}</pre>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
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
