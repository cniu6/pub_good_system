<template>
  <div class="admin-logs">
    <n-card title="操作日志">
      <n-space class="mb-4">
        <n-input v-model:value="query.username" placeholder="用户名" clearable style="width: 150px" />
        <n-input v-model:value="query.module" placeholder="模块" clearable style="width: 120px" />
        <n-select v-model:value="query.method" :options="methodOptions" placeholder="方法" clearable style="width: 100px" />
        <n-button type="primary" @click="fetchLogs">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
      </n-space>

      <n-data-table
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { NCard, NSpace, NInput, NSelect, NButton, NDataTable, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminApi } from '@/service/api/admin'

const message = useMessage()
const loading = ref(false)
const logList = ref<any[]>([])

const query = reactive({
  username: '',
  module: '',
  method: null as string | null,
  page: 1,
  page_size: 20,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
})

const methodOptions = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
]

const methodColors: Record<string, string> = {
  GET: 'info',
  POST: 'success',
  PUT: 'warning',
  DELETE: 'error',
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户', key: 'username', width: 100 },
  { title: '模块', key: 'module', width: 100 },
  { title: '操作', key: 'action', width: 80 },
  {
    title: '方法',
    key: 'method',
    width: 80,
    render(row) {
      return h(NTag, { type: methodColors[row.method] as any, size: 'small' }, () => row.method)
    },
  },
  { title: '路径', key: 'path', ellipsis: { tooltip: true } },
  { title: 'IP', key: 'ip', width: 120 },
  { title: '耗时(ms)', key: 'duration', width: 80 },
  {
    title: '时间',
    key: 'create_time',
    width: 180,
    render(row) {
      if (!row.create_time) return '-'
      return new Date(row.create_time * 1000).toLocaleString()
    },
  },
]

async function fetchLogs() {
  loading.value = true
  try {
    const res = await adminApi.log.list(query)
    logList.value = res.data?.list || []
    pagination.itemCount = res.data?.total || 0
  } catch {
    message.error('获取日志列表失败')
  } finally {
    loading.value = false
  }
}

function handleReset() {
  query.username = ''
  query.module = ''
  query.method = null
  query.page = 1
  fetchLogs()
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchLogs()
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.admin-logs { padding: 16px; }
.mb-4 { margin-bottom: 16px; }
</style>
