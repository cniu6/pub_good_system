<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminScoreLogApi, adminUserApi } from '@/service/api/admin/user'
import { parseMemo } from '@/utils/memo'
import I18nMemoEditor from '@/components/common/I18nMemoEditor.vue'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()
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
  memo: {} as Record<string, string>,
})

const columns: DataTableColumns<Entity.UserScoreLog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('adminRealname.userId'), key: 'user_id', width: 80 },
  {
    title: t('moneyScore.scoreChange'),
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
    title: t('moneyScore.beforeChange'),
    key: 'before',
    width: 100,
    render: row => `${Number(row.before) || 0}`,
  },
  {
    title: t('moneyScore.afterChange'),
    key: 'after',
    width: 100,
    render: row => `${Number(row.after) || 0}`,
  },
  {
    title: t('moneyScore.remark'),
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: t('moneyScore.time'),
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
  {
    title: t('moneyScore.actions'),
    key: 'actions',
    width: 80,
    render: row => h(NButton, {
      size: 'small',
      type: 'error',
      text: true,
      onClick: () => handleDelete(row.id),
    }, { default: () => t('adminUsers.delete') }),
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
    }
    else {
      message.error(res.message || t('moneyScore.fetchScoreFailed'))
    }
  }
  catch {
    message.error(t('moneyScore.fetchScoreFailed'))
  }
  finally {
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
  addForm.memo = {}
  showModal.value = true
}

async function handleSubmit() {
  if (!addForm.user_id) {
    message.error(t('adminMoneyLogs.enterUserId'))
    return
  }
  if (addForm.score === 0) {
    message.error(t('adminUsers.scoreCannotBeZero'))
    return
  }
  submitting.value = true
  try {
    const memoStr = Object.keys(addForm.memo).length > 0 ? JSON.stringify(addForm.memo) : ''
    const res = await adminUserApi.changeScore(addForm.user_id, {
      score: addForm.score,
      memo: memoStr,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminUsers.scoreChangedSuccess'))
      showModal.value = false
      fetchData()
    }
    else {
      message.error(res.message || t('adminUsers.scoreChangedFailed'))
    }
  }
  catch (e: unknown) {
    message.error((e instanceof Error ? e.message : null) || t('adminUsers.operationFailed'))
  }
  finally {
    submitting.value = false
  }
}

function handleDelete(id: number) {
  dialog.warning({
    title: t('adminMoneyLogs.confirmDeleteTitle'),
    content: t('adminScoreLogs.confirmDeleteContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await adminScoreLogApi.delete(id)
        if (res.isSuccess) {
          message.success(res.message || t('adminUsers.deleteSuccess'))
          fetchData()
        }
        else {
          message.error(res.message || t('adminUsers.deleteFailed'))
        }
      }
      catch {
        message.error(t('adminUsers.deleteFailed'))
      }
    },
  })
}

onMounted(() => fetchData())
</script>

<template>
  <n-card :title="t('adminScoreLogs.title')">
    <n-space vertical>
      <n-space>
        <n-input v-model:value="searchForm.keyword" :placeholder="t('adminScoreLogs.searchPlaceholder')" clearable style="width: 200px" @keyup.enter="handleSearch" />
        <n-input-number v-model:value="searchForm.user_id" :placeholder="t('adminRealname.userId')" style="width: 140px" :show-button="false" />
        <NButton type="primary" @click="handleSearch">
          {{ t('moneyScore.search') }}
        </NButton>
        <NButton @click="handleReset">
          {{ t('common.reset') }}
        </NButton>
        <NButton type="success" @click="handleAdd">
          {{ t('adminScoreLogs.changeScore') }}
        </NButton>
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

    <n-modal v-model:show="showModal" :title="t('adminScoreLogs.changeUserScore')" preset="card" style="width: 500px">
      <n-form :model="addForm" label-placement="left" label-width="80px">
        <n-form-item :label="t('adminRealname.userId')" required>
          <n-input-number v-model:value="addForm.user_id" :placeholder="t('adminMoneyLogs.enterUserIdPlaceholder')" :show-button="false" style="width: 100%" />
        </n-form-item>
        <n-form-item :label="t('adminUsers.score')" required>
          <n-input-number v-model:value="addForm.score" :placeholder="t('adminScoreLogs.scorePlaceholder')" :step="1" style="width: 100%" />
        </n-form-item>
        <n-form-item :label="t('moneyScore.remark')">
          <I18nMemoEditor v-model="addForm.memo" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <NButton @click="showModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ t('common.confirm') }}
          </NButton>
        </n-space>
      </template>
    </n-modal>
  </n-card>
</template>
