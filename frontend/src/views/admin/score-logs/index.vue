<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility, withSubmitLock } from '@/hooks'
import { adminScoreLogApi, adminUserApi } from '@/service/api/admin/user'
import { parseMemo } from '@/utils/memo'
import I18nMemoEditor from '@/components/common/I18nMemoEditor.vue'
import { useLedgerLogPage } from '../composables/useLedgerLogPage'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()
const submitting = ref(false)

const {
  loading,
  logList,
  searchForm,
  pagination,
  fetchData,
  handleSearch,
  handleReset,
  handlePageChange,
  handlePageSizeChange,
  handleDelete,
} = useLedgerLogPage<Entity.UserScoreLog>({
  fetchList: params => adminScoreLogApi.list(params),
  deleteItem: id => adminScoreLogApi.delete(id),
  fetchErrorMessage: t('moneyScore.fetchScoreFailed'),
  deleteSuccessMessage: t('adminUsers.deleteSuccess'),
  deleteFailedMessage: t('adminUsers.deleteFailed'),
  deleteConfirmTitle: t('adminScoreLogs.confirmDeleteTitle'),
  deleteConfirmContent: t('adminScoreLogs.confirmDeleteContent'),
})

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

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'user_id', label: t('adminRealname.userId') },
  { key: 'score', label: t('moneyScore.scoreChange') },
  { key: 'before', label: t('moneyScore.beforeChange') },
  { key: 'after', label: t('moneyScore.afterChange') },
  { key: 'memo', label: t('moneyScore.remark') },
  { key: 'create_time', label: t('moneyScore.time') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<Entity.UserScoreLog>({
  storageKey: 'admin-score-logs-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 900,
})

function handleAdd() {
  addForm.user_id = null
  addForm.score = 0
  addForm.memo = {}
  showModal.value = true
}

async function handleSubmit() {
  if (submitting.value)
    return
  if (!addForm.user_id) {
    message.error(t('adminMoneyLogs.enterUserId'))
    return
  }
  if (addForm.score === 0) {
    message.error(t('adminUsers.scoreCannotBeZero'))
    return
  }
  const userId = addForm.user_id
  const score = addForm.score
  const memo = { ...addForm.memo }
  dialog.warning({
    title: t('adminScoreLogs.confirmChangeTitle'),
    content: t('adminScoreLogs.confirmChangeContent', {
      userId,
      score: `${score > 0 ? '+' : ''}${score}`,
    }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(submitting, async () => {
      try {
        const memoStr = Object.keys(memo).length > 0 ? JSON.stringify(memo) : ''
        const res = await adminUserApi.changeScore(userId, {
          score,
          memo: memoStr,
        })
        if (res.isSuccess) {
          message.success(res.message || t('adminUsers.scoreChangedSuccess'))
          showModal.value = false
          fetchData()
          return
        }
        message.error(res.message || t('adminUsers.scoreChangedFailed'))
        return false
      }
      catch (e: unknown) {
        message.error((e instanceof Error ? e.message : null) || t('adminUsers.operationFailed'))
        return false
      }
    }),
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
        :columns="visibleColumns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: Entity.UserScoreLog) => row.id"
        :scroll-x="tableScrollX"
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
