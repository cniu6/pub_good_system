<template>
  <n-card title="积分日志管理">
    <n-space vertical>
      <n-space>
        <n-input v-model:value="searchForm.keyword" placeholder="搜索备注/积分" clearable style="width: 200px" @keyup.enter="handleSearch" />
        <n-input-number v-model:value="searchForm.user_id" placeholder="用户ID" style="width: 140px" :show-button="false" />
        <n-button type="primary" @click="handleSearch">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
        <n-button type="success" @click="handleAdd">变更积分</n-button>
      </n-space>

      <n-data-table
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: Entity.UserScoreLog) => row.id"
        striped
        size="small"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-space>

    <n-modal v-model:show="showModal" title="变更用户积分" preset="card" style="width: 500px">
      <n-form :model="addForm" label-placement="left" label-width="80px">
        <n-form-item label="用户ID" required>
          <n-input-number v-model:value="addForm.user_id" placeholder="输入用户ID" :show-button="false" style="width: 100%" />
        </n-form-item>
        <n-form-item label="积分" required>
          <n-input-number v-model:value="addForm.score" placeholder="正数增加，负数扣减" :step="1" style="width: 100%" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="addForm.memo" type="textarea" placeholder="输入备注" :rows="3" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { NButton, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminUserApi, adminScoreLogApi } from '@/service/api/admin/user'

const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const submitting = ref(false)

const searchForm = reactive({
  keyword: '',
  user_id: null as number | null,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const logList = ref<Entity.UserScoreLog[]>([])

const showModal = ref(false)
const addForm = reactive({
  user_id: null as number | null,
  score: 0,
  memo: '',
})

const columns: DataTableColumns<Entity.UserScoreLog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '用户ID', key: 'user_id', width: 80 },
  {
    title: '积分变动',
    key: 'score',
    width: 120,
    render: (row) => {
      const score = Number(row.score) || 0
      const isPositive = score > 0
      return h('span', {
        style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: '500' },
      }, `${isPositive ? '+' : ''}${score}`)
    },
  },
  {
    title: '变动前',
    key: 'before',
    width: 100,
    render: row => `${Number(row.before) || 0}`,
  },
  {
    title: '变动后',
    key: 'after',
    width: 100,
    render: row => `${Number(row.after) || 0}`,
  },
  {
    title: '备注',
    key: 'memo',
    ellipsis: { tooltip: true },
  },
  {
    title: '时间',
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: row => h(NButton, {
      size: 'small',
      type: 'error',
      text: true,
      onClick: () => handleDelete(row.id),
    }, { default: () => '删除' }),
  },
]

async function fetchData() {
  loading.value = true
  try {
    const res = await adminScoreLogApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword || undefined,
      user_id: searchForm.user_id || undefined,
    })
    if (res.isSuccess) {
      logList.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    } else {
      message.error(res.message || '获取积分日志失败')
    }
  } catch (e) {
    message.error('获取积分日志失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchData()
}

function handleReset() {
  searchForm.keyword = ''
  searchForm.user_id = null
  pagination.page = 1
  fetchData()
}

function handlePageChange(page: number) {
  pagination.page = page
  fetchData()
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchData()
}

function handleAdd() {
  addForm.user_id = null
  addForm.score = 0
  addForm.memo = ''
  showModal.value = true
}

async function handleSubmit() {
  if (!addForm.user_id) {
    message.error('请输入用户ID')
    return
  }
  if (addForm.score === 0) {
    message.error('积分不能为0')
    return
  }
  submitting.value = true
  try {
    await adminUserApi.changeScore(addForm.user_id, {
      score: addForm.score,
      memo: addForm.memo,
    })
    message.success('积分变更成功')
    showModal.value = false
    fetchData()
  } catch (e: unknown) {
    message.error((e instanceof Error ? e.message : null) || '操作失败')
  } finally {
    submitting.value = false
  }
}

function handleDelete(id: number) {
  dialog.warning({
    title: '确认删除',
    content: '删除日志记录不会影响用户积分，确定删除？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminScoreLogApi.delete(id)
        message.success('删除成功')
        fetchData()
      } catch {
        message.error('删除失败')
      }
    },
  })
}

onMounted(() => fetchData())
</script>
