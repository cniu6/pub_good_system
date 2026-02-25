<template>
  <n-card title="操作日志">
    <n-space vertical>
      <n-space align="center">
        <n-text depth="3">共 {{ total }} 条日志</n-text>
        <n-divider vertical />
        <n-text depth="3">查询天数</n-text>
        <n-input-number
          v-model:value="queryDays"
          :min="1"
          :max="365"
          size="small"
          style="width: 120px"
        >
          <template #suffix>天</template>
        </n-input-number>
        <n-text depth="3">最大数量</n-text>
        <n-input-number
          v-model:value="maxCount"
          :min="1"
          :max="500"
          size="small"
          style="width: 120px"
        />
        <n-button size="small" type="primary" :loading="savingQuerySettings" @click="handleApplyQuerySettings">
          应用
        </n-button>
      </n-space>

      <n-data-table
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
import { NTag, NButton, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import { adminApi } from '@/service/api/admin'
import type { UserSimpleInfo } from '@/service/api/admin/user'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const logList = ref<any[]>([])
const userMap = ref<Record<number, UserSimpleInfo>>({})
const total = ref(0)
const queryDays = ref(30)
const maxCount = ref(20)
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
  if (!user) return `用户#${userId}`
  return user.nickname || user.username || `用户#${userId}`
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: '用户',
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
  { title: '模块', key: 'module', width: 100 },
  { title: '操作', key: 'action', width: 80 },
  {
    title: '方法',
    key: 'method',
    width: 80,
    render(row) {
      return h(NTag, { type: methodColors[row.method] ?? 'info', size: 'small' }, () => row.method)
    },
  },
  { title: '路径', key: 'path', ellipsis: { tooltip: true } },
  { title: 'IP', key: 'ip', width: 120 },
  { title: '耗时(ms)', key: 'duration', width: 90 },
  {
    title: '时间',
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
  catch {
    console.error('获取用户信息失败')
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
    message.error('获取日志列表失败')
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
          maxCount.value = Math.max(1, Math.min(500, Number(item.value) || 20))
        }
      }
    }

    if (!hasQueryDays) {
      await adminApi.settings.create({
        key: 'operation_log_query_days',
        value: String(queryDays.value),
        type: 'number',
        category: 'custom',
        label: '操作日志查询天数',
        description: '操作日志页面默认查询时间范围（天）',
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
        label: '操作日志最大数量',
        description: '操作日志页面单页最大查询数量',
        is_public: false,
        is_editable: true,
      })
    }

    query.page_size = maxCount.value
    pagination.pageSize = maxCount.value
  }
  catch {
    query.page_size = maxCount.value
    pagination.pageSize = maxCount.value
  }
}

async function handleApplyQuerySettings() {
  savingQuerySettings.value = true
  try {
    queryDays.value = Math.max(1, Math.floor(queryDays.value || 1))
    maxCount.value = Math.max(1, Math.min(500, Math.floor(maxCount.value || 1)))

    await adminApi.settings.batchUpdate({
      operation_log_query_days: String(queryDays.value),
      operation_log_max_count: String(maxCount.value),
    })

    query.page = 1
    pagination.page = 1
    query.page_size = maxCount.value
    pagination.pageSize = maxCount.value
    applyDateRange()
    await fetchLogs()
    message.success('查询设置已更新')
  }
  catch {
    message.error('更新查询设置失败')
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
