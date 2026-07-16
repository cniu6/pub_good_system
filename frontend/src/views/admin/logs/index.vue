<template>
  <n-card :title="t('adminLogs.title')">
    <n-space vertical>
      <n-space align="center">
        <n-text depth="3">{{ t('adminLogs.totalLogs', { total }) }}</n-text>
        <n-divider vertical />
        <n-text depth="3">{{ t('adminLogs.queryDays') }}</n-text>
        <n-input-number
          v-model:value="queryDays"
          :min="1"
          :max="365"
          size="small"
          style="width: 120px"
        >
          <template #suffix>{{ t('adminLogs.days') }}</template>
        </n-input-number>
        <n-text depth="3">{{ t('adminLogs.logRetentionLimit') }}</n-text>
        <n-input-number
          v-model:value="maxCount"
          :min="20"
          :max="10000"
          size="small"
          style="width: 120px"
        />
        <n-text depth="3" style="font-size:12px;color:#999;">{{ t('adminLogs.autoCleanupHint') }}</n-text>
        <n-button size="small" type="primary" :loading="savingQuerySettings" @click="handleApplyQuerySettings">
          {{ t('adminLogs.apply') }}
        </n-button>
      </n-space>

      <n-space justify="end">
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
      </n-space>

      <n-data-table
        remote
        :columns="visibleColumns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="tableScrollX"
        @update:page="handlePageChange"
      />

      <n-modal v-model:show="showDetail" preset="card" :title="t('adminLogs.detailTitle')" style="width: 860px;" :mask-closable="true">
        <n-text v-if="detailLoading" depth="3">{{ t('adminLogs.loading') }}</n-text>
        <n-space v-else-if="detailData" vertical :size="16">
          <n-descriptions bordered :column="2" label-placement="left">
            <n-descriptions-item :label="t('adminLogs.module')">{{ detailData.module || '-' }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminLogs.action')">{{ detailData.action || '-' }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminLogs.method')">{{ detailData.method || '-' }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminLogs.path')">{{ detailData.path || '-' }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminLogs.ip')">{{ detailData.ip || '-' }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminLogs.duration')">{{ detailData.duration || 0 }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminLogs.time')" :span="2">{{ detailData.create_time ? new Date(detailData.create_time * 1000).toLocaleString() : '-' }}</n-descriptions-item>
          </n-descriptions>

          <n-card size="small" embedded :title="t('adminLogs.requestBody')">
            <div class="payload-block">{{ formattedRequestBody || '-' }}</div>
          </n-card>

          <n-card size="small" embedded :title="t('adminLogs.responseBody')">
            <div class="payload-block dark">{{ formattedResponseBody || '-' }}</div>
          </n-card>
        </n-space>
        <n-text v-else depth="3">{{ t('adminLogs.noDetailData') }}</n-text>
      </n-modal>
    </n-space>
  </n-card>
</template>

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTag, NButton, useMessage, NDescriptions, NDescriptionsItem, NCard, NText, NSpace, NModal } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import type { UserSimpleInfo } from '@/service/api/admin/user'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const loading = ref(false)
const logList = ref<any[]>([])
const userMap = ref<Record<number, UserSimpleInfo>>({})
const total = ref(0)
const queryDays = ref(30)
const maxCount = ref(500)
const savingQuerySettings = ref(false)
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

const methodColors: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  GET: 'info',
  POST: 'success',
  PUT: 'warning',
  DELETE: 'error',
}

// 跳转到用户详情页（admin hash 路由内部路径）
function goToUserDetail(userId: number) {
  if (userId) {
    router.push(`/users/${userId}`)
  }
}

// 获取用户显示名称
function getUserDisplayName(userId: number): string {
  const user = userMap.value[userId]
  if (!user) return t('adminLogs.userPrefix', { id: userId })
  return user.nickname || user.username || t('adminLogs.userPrefix', { id: userId })
}

function formatPayload(raw?: string) {
  const value = raw?.trim()
  if (!value) return ''
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  }
  catch {
    return value
  }
}

const formattedRequestBody = computed(() => formatPayload(detailData.value?.request_body))
const formattedResponseBody = computed(() => formatPayload(detailData.value?.response_body))

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminLogs.user'),
    key: 'user_id',
    width: 120,
    render(row) {
      const userId = row.user_id
      if (!userId) return '-'
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
    const res = await adminApi.log.detail(id)
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

// 批量获取日志中的用户信息
async function fetchUserInfos(logs: any[]) {
  const userIds = [...new Set(logs.map(log => log.user_id).filter(Boolean))]
  if (userIds.length === 0) return

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
    const res = await adminApi.log.list(query)
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

function applyDateRange() {
  const now = Math.floor(Date.now() / 1000)
  const safeDays = Math.max(1, Math.floor(queryDays.value || 1))
  query.end_time = now
  query.start_time = now - safeDays * 24 * 60 * 60
}

async function loadQuerySettings() {
  let hasQueryDays = false
  let hasMaxCount = false
  try {
    const res = await adminApi.settings.list()
    const categories = res.data?.categories || []
    for (const category of categories) {
      for (const item of category.items) {
        if (item.key === 'operation_log_query_days') {
          hasQueryDays = true
          queryDays.value = Math.max(1, Number(item.value) || 30)
        }
        if (item.key === 'operation_log_max_count') {
          hasMaxCount = true
          maxCount.value = Math.max(20, Math.min(10000, Number(item.value) || 500))
        }
      }
    }

    if (!hasQueryDays) {
      await adminApi.settings.create({
        key: 'operation_log_query_days',
        value: String(queryDays.value),
        type: 'number',
        category: 'custom',
        label: t('adminLogs.queryDaysLabel'),
        description: t('adminLogs.queryDaysDesc'),
        is_public: false,
        is_editable: true,
      })
    }

    if (!hasMaxCount) {
      await adminApi.settings.create({
        key: 'operation_log_max_count',
        value: String(maxCount.value),
        type: 'number',
        category: 'custom',
        label: t('adminLogs.maxCountLabel'),
        description: t('adminLogs.maxCountDesc'),
        is_public: false,
        is_editable: true,
      })
    }

  }
  catch {
    // use defaults
  }
}

async function handleApplyQuerySettings() {
  savingQuerySettings.value = true
  try {
    queryDays.value = Math.max(1, Math.floor(queryDays.value || 1))
    maxCount.value = Math.max(20, Math.min(10000, Math.floor(maxCount.value || 500)))

    const res = await adminApi.settings.batchUpdate({
      operation_log_query_days: String(queryDays.value),
      operation_log_max_count: String(maxCount.value),
    })
    if (res.isSuccess) {
      query.page = 1
      pagination.page = 1
      applyDateRange()
      await fetchLogs()
      message.success(res.message || t('adminLogs.querySettingsUpdated'))
    }
    else {
      message.error(res.message || t('adminLogs.updateQuerySettingsFailed'))
    }
  }
  catch {
    message.error(t('adminLogs.updateQuerySettingsFailed'))
  }
  finally {
    savingQuerySettings.value = false
  }
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchLogs()
}

onMounted(() => {
  loadQuerySettings().then(() => {
    applyDateRange()
    fetchLogs()
  })
})
</script>

<style scoped>
.payload-block {
  max-height: 260px;
  overflow-y: auto;
  padding: 12px;
  border-radius: 10px;
  background: rgb(250 250 252);
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.payload-block.dark {
  background: rgb(17 24 39);
  color: rgb(229 231 235);
}
</style>
