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

      <n-data-table
        remote
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
      />
    </n-space>
  </n-card>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTag, NButton, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
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

// 获取管理端路径前缀
const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr'

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

// 跳转到用户详情页
function goToUserDetail(userId: number) {
  if (userId) {
    router.push(`${adminPath}/users/${userId}`)
  }
}

// 获取用户显示名称
function getUserDisplayName(userId: number): string {
  const user = userMap.value[userId]
  if (!user) return t('adminLogs.userPrefix', { id: userId })
  return user.nickname || user.username || t('adminLogs.userPrefix', { id: userId })
}

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
]

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

<style scoped></style>
