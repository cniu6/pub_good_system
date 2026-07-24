<script setup lang="ts">
import { computed, h, reactive, ref, watch } from 'vue'
import { NButton, NSpace } from 'naive-ui'
import type { DataTableColumns, PaginationProps } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { withSubmitLock } from '@/hooks'
import type { DbColumnMeta } from '@/service/api/admin/db'

const props = defineProps<{
  columns: string[]
  loading: boolean
  metaColumns: DbColumnMeta[]
  page: number
  pageSize: number
  rows: Record<string, unknown>[]
  total: number
  writeEnabled: boolean
  /** 父级写库（返回是否成功），子组件 await 后再关弹窗，避免假锁 */
  createRow: (values: Record<string, unknown>) => Promise<boolean>
  updateRow: (primaryKey: Record<string, unknown>, values: Record<string, unknown>) => Promise<boolean>
  deleteRow: (primaryKey: Record<string, unknown>) => Promise<boolean>
}>()

const emit = defineEmits<{
  pageChange: [page: number]
  pageSizeChange: [size: number]
}>()

const { t } = useI18n()
const dialog = useDialog()
const editVisible = ref(false)
const viewVisible = ref(false)
const submitting = ref(false)
const editing = ref(false)
const editPrimaryKey = ref<Record<string, unknown>>({})
const viewingRow = ref<Record<string, unknown> | null>(null)
const draft = reactive<Record<string, string>>({})
const visibleKeys = ref<string[]>([])

const primaryColumns = computed(() => props.metaColumns.filter(column => column.primary_key))
const hasPrimaryKey = computed(() => primaryColumns.value.length > 0)
const canWriteRow = computed(() => props.writeEnabled && hasPrimaryKey.value)
const editableColumns = computed(() => props.metaColumns.filter((column) => {
  if (column.primary_key && (editing.value || column.auto_increment))
    return false
  return !/password|passwd|secret|token|api[_ -]?key|access[_ -]?key|private[_ -]?key|credential/i.test(column.name)
}))
const columnOptions = computed(() => props.columns.map(key => ({ key, label: key })))

/** 查看弹窗：优先按表结构字段顺序，否则用当前列 */
const viewEntries = computed(() => {
  const row = viewingRow.value
  if (!row)
    return [] as Array<{ key: string, value: string, long: boolean }>
  const keys = props.metaColumns.length
    ? props.metaColumns.map(column => column.name)
    : props.columns
  const seen = new Set<string>()
  const entries: Array<{ key: string, value: string, long: boolean }> = []
  for (const key of keys) {
    if (seen.has(key))
      continue
    seen.add(key)
    const value = formatCellValue(row[key])
    entries.push({ key, value, long: value.length > 120 || value.includes('\n') })
  }
  // 补上结构里没有、但行数据里有的字段
  for (const key of Object.keys(row)) {
    if (seen.has(key))
      continue
    const value = formatCellValue(row[key])
    entries.push({ key, value, long: value.length > 120 || value.includes('\n') })
  }
  return entries
})

watch(() => props.columns, (columns) => {
  visibleKeys.value = columns.slice()
}, { immediate: true })

const tableColumns = computed<DataTableColumns<Record<string, unknown>>>(() => {
  const columns: DataTableColumns<Record<string, unknown>> = props.columns
    .filter(key => visibleKeys.value.includes(key))
    .map(key => ({
      title: key,
      key,
      minWidth: 140,
      ellipsis: { tooltip: true },
      render: (row: Record<string, unknown>) => h('span', String(row[key] ?? '')),
    }))
  if (props.columns.length) {
    columns.push({
      title: t('common.actions'),
      key: '__actions__',
      fixed: 'right',
      width: canWriteRow.value ? 200 : 80,
      render: row => h(NSpace, { size: 4 }, () => {
        const buttons = [
          h(NButton, { size: 'tiny', tertiary: true, onClick: () => openView(row) }, () => t('common.view')),
        ]
        if (canWriteRow.value) {
          buttons.push(
            h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, () => t('common.edit')),
            h(NButton, { size: 'tiny', type: 'error', secondary: true, onClick: () => confirmDelete(row) }, () => t('common.delete')),
          )
        }
        return buttons
      }),
    })
  }
  return columns
})

const pagination = computed<PaginationProps>(() => ({
  page: props.page,
  pageSize: props.pageSize,
  itemCount: props.total,
  pageSizes: [10, 20, 50, 100],
  showSizePicker: true,
  // 与其它管理端列表一致：显式展示总条数（remote 模式下依赖 itemCount）
  prefix: ({ itemCount }) => t('adminServer.dbPaginationTotal', { total: itemCount ?? 0 }),
  onUpdatePage: (page: number) => emit('pageChange', page),
  onUpdatePageSize: (size: number) => emit('pageSizeChange', size),
}))

function formatCellValue(value: unknown): string {
  if (value == null)
    return ''
  if (typeof value === 'string')
    return value
  if (typeof value === 'number' || typeof value === 'boolean' || typeof value === 'bigint')
    return String(value)
  try {
    return JSON.stringify(value, null, 2)
  }
  catch {
    return String(value)
  }
}

function resetDraft(row?: Record<string, unknown>) {
  Object.keys(draft).forEach(key => delete draft[key])
  for (const column of editableColumns.value)
    draft[column.name] = row?.[column.name] == null ? '' : String(row[column.name])
}

function primaryKeyFrom(row: Record<string, unknown>) {
  return Object.fromEntries(primaryColumns.value.map(column => [column.name, row[column.name]]))
}

function openCreate() {
  editing.value = false
  editPrimaryKey.value = {}
  resetDraft()
  editVisible.value = true
}

function openView(row: Record<string, unknown>) {
  viewingRow.value = { ...row }
  viewVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editing.value = true
  editPrimaryKey.value = primaryKeyFrom(row)
  resetDraft(row)
  editVisible.value = true
}

function confirmDelete(row: Record<string, unknown>) {
  dialog.warning({
    title: t('common.confirmDelete'),
    content: t('adminServer.dbDeleteRowConfirm'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(submitting, async () => {
      await props.deleteRow(primaryKeyFrom(row))
    }),
  })
}

async function submit() {
  const values = Object.fromEntries(Object.entries(draft).filter(([, value]) => value !== ''))
  if (!Object.keys(values).length) {
    window.$message?.warning(t('adminServer.dbRowValuesRequired'))
    return
  }
  await withSubmitLock(submitting, async () => {
    const ok = editing.value
      ? await props.updateRow(editPrimaryKey.value, values)
      : await props.createRow(values)
    if (ok)
      editVisible.value = false
  })
}
</script>

<template>
  <NCard size="small" :title="t('adminServer.dbData')" :segmented="{ content: true }">
    <template #header-extra>
      <NSpace>
        <TableColumnSelector
          v-if="columns.length"
          v-model="visibleKeys"
          :button-label="t('common.showFields')"
          :hint="t('common.columnVisibilityHint')"
          :options="columnOptions"
          :reset-label="t('common.restoreDefaultFields')"
          :title="t('common.visibleFields')"
          :total-count="columns.length"
          :visible-count="visibleKeys.length"
          @reset="visibleKeys = columns.slice()"
        />
        <NButton v-if="canWriteRow" type="primary" size="small" @click="openCreate">
          {{ t('common.add') }}
        </NButton>
      </NSpace>
    </template>
    <NAlert v-if="!hasPrimaryKey && metaColumns.length" type="warning" :bordered="false" class="mb-12px">
      {{ t('adminServer.dbNoPrimaryKeyWrite') }}
    </NAlert>
    <!-- remote：服务端分页，必须用 itemCount=total，否则只会按当前页行数算出 1 页 -->
    <NDataTable
      remote
      :bordered="false"
      :columns="tableColumns"
      :data="rows"
      :loading="loading"
      :pagination="pagination"
      :scroll-x="Math.max(760, columns.length * 150)"
      size="small"
    />
  </NCard>

  <NModal
    v-model:show="viewVisible"
    preset="card"
    :title="t('adminServer.dbViewRow')"
    style="width: min(760px, calc(100vw - 32px));"
    :segmented="{ content: true }"
  >
    <div class="db-row-view">
      <NDescriptions
        v-if="viewEntries.length"
        bordered
        :column="1"
        label-placement="left"
        size="small"
        :label-style="{ width: '160px', verticalAlign: 'top' }"
      >
        <NDescriptionsItem v-for="item in viewEntries" :key="item.key" :label="item.key">
          <div class="db-row-view__value" :class="{ 'db-row-view__value--long': item.long }">
            {{ item.value || t('adminServer.dbEmptyValue') }}
          </div>
        </NDescriptionsItem>
      </NDescriptions>
      <NEmpty v-else :description="t('adminServer.dbViewEmpty')" />
    </div>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="viewVisible = false">
          {{ t('common.close') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>

  <NModal v-model:show="editVisible" preset="card" :title="editing ? t('adminServer.dbEditRow') : t('adminServer.dbCreateRow')" style="width: min(620px, calc(100vw - 32px));">
    <NForm label-placement="top">
      <NFormItem v-for="column in editableColumns" :key="column.name" :label="`${column.name} (${column.type})`">
        <NInput v-model:value="draft[column.name]" :placeholder="column.comment || t('common.inputPlaceholder')" />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="editVisible = false">
          {{ t('common.cancel') }}
        </NButton>
        <NButton type="primary" :loading="submitting" @click="submit">
          {{ t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.db-row-view {
  max-height: min(70vh, 640px);
  overflow: auto;
}

.db-row-view__value {
  word-break: break-word;
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
}

.db-row-view__value--long {
  max-height: 220px;
  overflow: auto;
  padding: 4px 0;
}
</style>
