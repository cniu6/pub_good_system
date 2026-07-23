<script setup lang="ts">
import { computed, h, reactive, ref, watch } from 'vue'
import { NButton, NSpace } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
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
}>()

const emit = defineEmits<{
  create: [values: Record<string, unknown>]
  delete: [primaryKey: Record<string, unknown>]
  pageChange: [page: number]
  pageSizeChange: [size: number]
  update: [primaryKey: Record<string, unknown>, values: Record<string, unknown>]
}>()

const { t } = useI18n()
const dialog = useDialog()
const editVisible = ref(false)
const submitting = ref(false)
const editing = ref(false)
const editPrimaryKey = ref<Record<string, unknown>>({})
const draft = reactive<Record<string, string>>({})
const visibleKeys = ref<string[]>([])

const primaryColumns = computed(() => props.metaColumns.filter(column => column.primary_key))
const hasPrimaryKey = computed(() => primaryColumns.value.length > 0)
const editableColumns = computed(() => props.metaColumns.filter((column) => {
  if (column.primary_key && (editing.value || column.auto_increment))
    return false
  return !/password|passwd|secret|token|api[_ -]?key|access[_ -]?key|private[_ -]?key|credential/i.test(column.name)
}))
const columnOptions = computed(() => props.columns.map(key => ({ key, label: key })))

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
  if (props.writeEnabled && hasPrimaryKey.value) {
    columns.push({
      title: t('common.actions'),
      key: '__actions__',
      fixed: 'right',
      width: 130,
      render: row => h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, () => t('common.edit')),
        h(NButton, { size: 'tiny', type: 'error', secondary: true, onClick: () => confirmDelete(row) }, () => t('common.delete')),
      ]),
    })
  }
  return columns
})

const pagination = computed(() => ({
  page: props.page,
  pageSize: props.pageSize,
  itemCount: props.total,
  pageSizes: [10, 20, 50, 100],
  showSizePicker: true,
  onUpdatePage: (page: number) => emit('pageChange', page),
  onUpdatePageSize: (size: number) => emit('pageSizeChange', size),
}))

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
    onPositiveClick: () => emit('delete', primaryKeyFrom(row)),
  })
}

async function submit() {
  const values = Object.fromEntries(Object.entries(draft).filter(([, value]) => value !== ''))
  if (!Object.keys(values).length) {
    window.$message?.warning(t('adminServer.dbRowValuesRequired'))
    return
  }
  submitting.value = true
  try {
    if (editing.value)
      emit('update', editPrimaryKey.value, values)
    else
      emit('create', values)
    editVisible.value = false
  }
  finally {
    submitting.value = false
  }
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
        <NButton v-if="writeEnabled && hasPrimaryKey" type="primary" size="small" @click="openCreate">
          {{ t('common.add') }}
        </NButton>
      </NSpace>
    </template>
    <NAlert v-if="!hasPrimaryKey && metaColumns.length" type="warning" :bordered="false" class="mb-12px">
      {{ t('adminServer.dbNoPrimaryKeyWrite') }}
    </NAlert>
    <NDataTable
      :bordered="false"
      :columns="tableColumns"
      :data="rows"
      :loading="loading"
      :pagination="pagination"
      :scroll-x="Math.max(760, columns.length * 150)"
      size="small"
    />
  </NCard>

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
