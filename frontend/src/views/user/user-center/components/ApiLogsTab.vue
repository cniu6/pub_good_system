<script setup lang="ts">
/**
 * 用户中心 - 本人 API 访问日志（详情弹窗对齐管理端 NCard+NCode）
 */
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NCode, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { fetchMyAPILogDetail, fetchMyAPILogs, fetchMyAPILogStats } from '@/service/api/user/logs'
import type { UserAPIAccessLog, UserAPILogStats } from '@/service/api/user/logs'
import { formatPrettyJSON } from '@/utils/format'

const { t } = useI18n()
const message = useMessage()

const loading = ref(false)
const logList = ref<UserAPIAccessLog[]>([])
const total = ref(0)
const showDetail = ref(false)
const detailLoading = ref(false)
const detailData = ref<UserAPIAccessLog | null>(null)

const statsData = ref<UserAPILogStats>({
  total_count: 0,
  today_count: 0,
  success_count: 0,
  client_error_count: 0,
  server_error_count: 0,
  avg_duration: 0,
  top_paths: [],
  method_stats: [],
})

async function fetchStats() {
  try {
    const res = await fetchMyAPILogStats()
    if (res.data)
      statsData.value = res.data
  }
  catch {}
}

const query = reactive({
  page: 1,
  page_size: 20,
  start_time: 0,
  end_time: 0,
})

function authMethodLabel(method?: string) {
  if (method === 'apikey')
    return t('userLogs.authMethodApiKey')
  return method || '-'
}

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  onChange: (page: number) => {
    query.page = page
    pagination.page = page
    fetchLogs()
  },
})

const methodColors: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  GET: 'info',
  POST: 'success',
  PUT: 'warning',
  DELETE: 'error',
}

const formattedQueryString = computed(() => formatPrettyJSON(detailData.value?.query_string))
const formattedPathParams = computed(() => formatPrettyJSON(detailData.value?.path_params))
const formattedRequestHeaders = computed(() => formatPrettyJSON(detailData.value?.request_headers))
const formattedRequestBody = computed(() => formatPrettyJSON(detailData.value?.request_body))
const formattedResponseBody = computed(() => formatPrettyJSON(detailData.value?.response_body))

const columns: DataTableColumns<UserAPIAccessLog> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('userLogs.method'),
    key: 'method',
    width: 80,
    render(row) {
      return h(NTag, { type: methodColors[row.method || ''] ?? 'info', size: 'small' }, () => row.method || '-')
    },
  },
  {
    title: t('userLogs.authMethod'),
    key: 'auth_method',
    width: 100,
    render(row) {
      const type = row.auth_method === 'apikey' ? 'success' : row.auth_method === 'jwt' ? 'info' : 'default'
      return h(NTag, { type, size: 'small' }, () => authMethodLabel(row.auth_method))
    },
  },
  { title: t('userLogs.path'), key: 'path', ellipsis: { tooltip: true } },
  { title: t('userLogs.statusCode'), key: 'status_code', width: 90 },
  { title: t('userLogs.ip'), key: 'ip', width: 120 },
  { title: t('userLogs.duration'), key: 'duration', width: 90 },
  {
    title: t('userLogs.time'),
    key: 'create_time',
    width: 160,
    render(row) {
      if (!row.create_time)
        return '-'
      return new Date(row.create_time * 1000).toLocaleString()
    },
  },
  {
    title: t('userLogs.actions'),
    key: 'actions',
    width: 80,
    render(row) {
      return h(NButton, { text: true, type: 'primary', onClick: () => handleDetail(row.id) }, { default: () => t('userLogs.detail') })
    },
  },
]

async function handleDetail(id: number) {
  showDetail.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res = await fetchMyAPILogDetail(id)
    detailData.value = res.data || null
  }
  catch {
    showDetail.value = false
    message.error(t('userLogs.loadDetailFailed'))
  }
  finally {
    detailLoading.value = false
  }
}

function applyDateRange() {
  const now = Math.floor(Date.now() / 1000)
  query.end_time = now
  query.start_time = now - 7 * 24 * 60 * 60
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await fetchMyAPILogs({
      ...query,
      auth_method: 'apikey',
    })
    logList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = res.data?.total || 0
  }
  catch {
    message.error(t('userLogs.fetchLogsFailed'))
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  applyDateRange()
  fetchLogs()
  fetchStats()
})
</script>

<template>
  <div class="p-2">
    <n-space vertical>
      <n-text depth="3">
        {{ t('userLogs.apiHint') }}
      </n-text>
      <n-grid :x-gap="12" :y-gap="12" cols="2 s:4" responsive="screen">
        <n-gi>
          <n-card size="small">
            <n-statistic :label="t('userLogs.statsTotal')" :value="statsData.total_count" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic :label="t('userLogs.statsToday')" :value="statsData.today_count" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic :label="t('userLogs.statsErrors')">
              <n-text type="warning">
                {{ statsData.client_error_count + statsData.server_error_count }}
              </n-text>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic :label="t('userLogs.statsAvgDuration')">
              {{ Number(statsData.avg_duration || 0).toFixed(1) }} ms
            </n-statistic>
          </n-card>
        </n-gi>
      </n-grid>
      <n-space align="center">
        <NTag type="info" size="small">
          {{ t('userLogs.authMethodApiKey') }}
        </NTag>
        <n-text depth="3">
          {{ t('userLogs.totalLogs', { total }) }}
        </n-text>
      </n-space>
      <n-data-table
        remote
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="1060"
      />
    </n-space>

    <n-modal v-model:show="showDetail" preset="card" :title="t('userLogs.apiDetailTitle')" style="width: 1100px;" :mask-closable="true">
      <n-text v-if="detailLoading" depth="3">
        {{ t('userLogs.loading') }}
      </n-text>
      <n-space v-else-if="detailData" vertical :size="16">
        <n-card size="small" embedded :title="t('userLogs.basicInfo')">
          <n-descriptions bordered :column="2" label-placement="left">
            <n-descriptions-item :label="t('userLogs.id')">
              {{ detailData.id }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.requestId')">
              {{ detailData.request_id || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.method')">
              {{ detailData.method || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.authMethod')">
              {{ authMethodLabel(detailData.auth_method) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.statusCode')">
              {{ detailData.status_code }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.path')">
              {{ detailData.path || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.routePath')">
              {{ detailData.route_path || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.ip')">
              {{ detailData.ip || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.duration')">
              {{ detailData.duration || 0 }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.time')" :span="2">
              {{ detailData.create_time ? new Date(detailData.create_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('userLogs.userAgent')" :span="2">
              {{ detailData.user_agent || '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>

        <n-card size="small" embedded :title="t('userLogs.queryString')">
          <NCode :code="formattedQueryString || '-'" language="json" word-wrap style="max-height: 220px; overflow: auto;" />
        </n-card>
        <n-card size="small" embedded :title="t('userLogs.pathParams')">
          <NCode :code="formattedPathParams || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </n-card>
        <n-card size="small" embedded :title="t('userLogs.requestHeaders')">
          <NCode :code="formattedRequestHeaders || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </n-card>
        <n-card size="small" embedded :title="t('userLogs.requestBody')">
          <NCode :code="formattedRequestBody || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </n-card>
        <n-card size="small" embedded :title="t('userLogs.responseBody')">
          <NCode :code="formattedResponseBody || '-'" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </n-card>
      </n-space>
      <n-text v-else depth="3">
        {{ t('userLogs.noDetailData') }}
      </n-text>
    </n-modal>
  </div>
</template>
