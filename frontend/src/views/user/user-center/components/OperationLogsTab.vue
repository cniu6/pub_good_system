<script setup lang="ts">
/**
 * 用户中心 - 本人操作日志（风格对齐管理端操作日志详情 NCard+NCode）
 */
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NCode, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { fetchMyOperationLogDetail, fetchMyOperationLogs } from '@/service/api/user/logs'
import type { UserOperationLog } from '@/service/api/user/logs'
import { formatPrettyJSON } from '@/utils/format'

const { t } = useI18n()
const message = useMessage()

const loading = ref(false)
const logList = ref<UserOperationLog[]>([])
const total = ref(0)
const showDetail = ref(false)
const detailLoading = ref(false)
const detailData = ref<UserOperationLog | null>(null)

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

const formattedRequestBody = computed(() => formatPrettyJSON(detailData.value?.request_body))
const formattedResponseBody = computed(() => formatPrettyJSON(detailData.value?.response_body))

const columns: DataTableColumns<UserOperationLog> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: t('userLogs.module'), key: 'module', width: 100 },
  { title: t('userLogs.action'), key: 'action', width: 80 },
  {
    title: t('userLogs.method'),
    key: 'method',
    width: 80,
    render(row) {
      return h(NTag, { type: methodColors[row.method || ''] ?? 'info', size: 'small' }, () => row.method || '-')
    },
  },
  { title: t('userLogs.path'), key: 'path', ellipsis: { tooltip: true } },
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
    const res = await fetchMyOperationLogDetail(id)
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
  query.start_time = now - 30 * 24 * 60 * 60
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await fetchMyOperationLogs(query)
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
})
</script>

<template>
  <div class="p-2">
    <n-space vertical>
      <n-text depth="3">
        {{ t('userLogs.operationHint') }}
      </n-text>
      <n-text depth="3">
        {{ t('userLogs.totalLogs', { total }) }}
      </n-text>
      <n-data-table
        remote
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="960"
      />
    </n-space>

    <n-modal v-model:show="showDetail" preset="card" :title="t('userLogs.operationDetailTitle')" style="width: 860px;" :mask-closable="true">
      <n-text v-if="detailLoading" depth="3">
        {{ t('userLogs.loading') }}
      </n-text>
      <n-space v-else-if="detailData" vertical :size="16">
        <n-descriptions bordered :column="2" label-placement="left">
          <n-descriptions-item :label="t('userLogs.module')">
            {{ detailData.module || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.action')">
            {{ detailData.action || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.method')">
            {{ detailData.method || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.path')">
            {{ detailData.path || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.handlerName')" :span="2">
            {{ detailData.handler_name || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.ip')">
            {{ detailData.ip || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.duration')">
            {{ detailData.duration || 0 }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.statusCode')">
            {{ detailData.status_code || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('userLogs.time')">
            {{ detailData.create_time ? new Date(detailData.create_time * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
        </n-descriptions>

        <n-card v-if="formattedRequestBody" size="small" embedded :title="t('userLogs.requestBody')">
          <NCode :code="formattedRequestBody" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </n-card>
        <n-card v-if="formattedResponseBody" size="small" embedded :title="t('userLogs.responseBody')">
          <NCode :code="formattedResponseBody" language="json" word-wrap style="max-height: 280px; overflow: auto;" />
        </n-card>
      </n-space>
      <n-text v-else depth="3">
        {{ t('userLogs.noDetailData') }}
      </n-text>
    </n-modal>
  </div>
</template>
