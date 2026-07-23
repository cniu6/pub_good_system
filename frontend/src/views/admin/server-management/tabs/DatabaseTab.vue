<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import { authStorage } from '@/utils'

const { t } = useI18n()
const message = useMessage()

const loading = ref(false)
const sqlLoading = ref(false)
const backupLoading = ref(false)
const driver = ref('')
const backupSupported = ref(false)
const tables = ref<string[]>([])
const selectedTable = ref<string | null>(null)
const previewRows = ref<Record<string, unknown>[]>([])
const previewColumns = ref<string[]>([])
const sqlText = ref('SELECT 1')
const allowWrite = ref(false)
const sqlResultRows = ref<Record<string, unknown>[]>([])
const sqlResultColumns = ref<string[]>([])
const sqlMeta = ref('')

const query = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

const tableOptions = computed(() => tables.value.map(name => ({ label: name, value: name })))

const previewTableColumns = computed<DataTableColumns<Record<string, unknown>>>(() => {
  return previewColumns.value.map(col => ({
    title: col,
    key: col,
    ellipsis: { tooltip: true },
    render: (row: Record<string, unknown>) => h('span', String(row[col] ?? '')),
  }))
})

const sqlTableColumns = computed<DataTableColumns<Record<string, unknown>>>(() => {
  return sqlResultColumns.value.map(col => ({
    title: col,
    key: col,
    ellipsis: { tooltip: true },
    render: (row: Record<string, unknown>) => h('span', String(row[col] ?? '')),
  }))
})

const pagination = computed(() => ({
  page: query.page,
  pageSize: query.page_size,
  itemCount: query.total,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onUpdatePage: (page: number) => {
    query.page = page
    void loadPreview()
  },
  onUpdatePageSize: (size: number) => {
    query.page_size = size
    query.page = 1
    void loadPreview()
  },
}))

async function loadInfoAndTables() {
  loading.value = true
  try {
    const [infoRes, tablesRes] = await Promise.all([
      adminApi.db.info(),
      adminApi.db.tables(),
    ])
    if (infoRes.isSuccess && infoRes.data) {
      driver.value = infoRes.data.driver || ''
      backupSupported.value = Boolean(infoRes.data.backup_supported)
    }
    if (tablesRes.isSuccess && tablesRes.data)
      tables.value = tablesRes.data.list || []
    else
      message.error(tablesRes.message || t('adminServer.dbLoadFailed'))
  }
  catch {
    message.error(t('adminServer.dbLoadFailed'))
  }
  finally {
    loading.value = false
  }
}

async function loadPreview() {
  if (!selectedTable.value)
    return
  loading.value = true
  try {
    const res = await adminApi.db.tableRows(selectedTable.value, {
      page: query.page,
      page_size: query.page_size,
    })
    if (!res.isSuccess || !res.data) {
      message.error(res.message || t('adminServer.dbLoadFailed'))
      return
    }
    previewColumns.value = res.data.columns || []
    previewRows.value = res.data.rows || []
    query.total = Number(res.data.total || 0)
  }
  catch {
    message.error(t('adminServer.dbLoadFailed'))
  }
  finally {
    loading.value = false
  }
}

function onTableChange(name: string | null) {
  selectedTable.value = name
  query.page = 1
  previewRows.value = []
  previewColumns.value = []
  if (name)
    void loadPreview()
}

async function runSql() {
  const sql = sqlText.value.trim()
  if (!sql) {
    message.warning(t('adminServer.dbSqlEmpty'))
    return
  }
  sqlLoading.value = true
  sqlResultRows.value = []
  sqlResultColumns.value = []
  sqlMeta.value = ''
  try {
    const res = await adminApi.db.execSql({ sql, allow_write: allowWrite.value })
    if (!res.isSuccess || !res.data) {
      message.error(res.message || t('adminServer.dbSqlFailed'))
      return
    }
    sqlResultColumns.value = res.data.columns || []
    sqlResultRows.value = res.data.rows || []
    const parts: string[] = []
    if (res.data.duration_ms != null)
      parts.push(`${res.data.duration_ms}ms`)
    if (res.data.row_count != null)
      parts.push(`rows=${res.data.row_count}`)
    if (res.data.rows_affected != null)
      parts.push(`affected=${res.data.rows_affected}`)
    if (res.data.truncated)
      parts.push(t('adminServer.dbTruncated'))
    sqlMeta.value = parts.join(' · ')
    message.success(t('adminServer.dbSqlSuccess'))
  }
  catch (e: any) {
    message.error(e?.message || t('adminServer.dbSqlFailed'))
  }
  finally {
    sqlLoading.value = false
  }
}

async function downloadBackup() {
  if (!backupSupported.value) {
    message.warning(t('adminServer.dbBackupUnsupported'))
    return
  }
  backupLoading.value = true
  try {
    const url = await adminApi.db.backupUrl()
    const token = authStorage.get('accessToken')
    const headers: Record<string, string> = {}
    if (token)
      headers.Authorization = `Bearer ${token}`
    const res = await fetch(url, { headers })
    if (!res.ok) {
      const text = await res.text()
      try {
        const payload = JSON.parse(text)
        throw new Error(String(payload?.message || text))
      }
      catch (err) {
        if (err instanceof SyntaxError)
          throw new Error(text)
        throw err
      }
    }
    const blob = await res.blob()
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = `fst-backup-${Date.now()}.db`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(objectUrl)
    message.success(t('adminServer.dbBackupSuccess'))
  }
  catch (e: any) {
    message.error(e?.message || t('adminServer.dbBackupFailed'))
  }
  finally {
    backupLoading.value = false
  }
}

onMounted(() => {
  void loadInfoAndTables()
})
</script>

<template>
  <n-space vertical :size="16">
    <n-card :title="t('adminServer.dbTab')" :bordered="false" size="small">
      <n-space align="center" class="mb-12px">
        <n-tag type="info">
          {{ t('adminServer.dbDriver') }}: {{ driver || '-' }}
        </n-tag>
        <n-button
          type="primary"
          secondary
          :loading="backupLoading"
          :disabled="!backupSupported"
          @click="downloadBackup"
        >
          {{ t('adminServer.dbBackup') }}
        </n-button>
        <n-button quaternary @click="loadInfoAndTables">
          {{ t('common.refresh') }}
        </n-button>
      </n-space>
      <n-alert v-if="!backupSupported" type="warning" class="mb-12px">
        {{ t('adminServer.dbBackupUnsupported') }}
      </n-alert>

      <n-space vertical :size="12">
        <n-select
          :value="selectedTable"
          :options="tableOptions"
          :placeholder="t('adminServer.dbSelectTable')"
          filterable
          clearable
          style="max-width: 360px"
          @update:value="onTableChange"
        />
        <n-data-table
          :loading="loading"
          :columns="previewTableColumns"
          :data="previewRows"
          :pagination="selectedTable ? pagination : false"
          :bordered="false"
          size="small"
          :scroll-x="Math.max(800, previewColumns.length * 140)"
        />
      </n-space>
    </n-card>

    <n-card :title="t('adminServer.dbSqlTitle')" :bordered="false" size="small">
      <n-input
        v-model:value="sqlText"
        type="textarea"
        :rows="5"
        :placeholder="t('adminServer.dbSqlPlaceholder')"
        class="mb-12px"
      />
      <n-space align="center" class="mb-12px">
        <n-checkbox v-model:checked="allowWrite">
          {{ t('adminServer.dbAllowWrite') }}
        </n-checkbox>
        <n-button type="primary" :loading="sqlLoading" @click="runSql">
          {{ t('adminServer.dbRunSql') }}
        </n-button>
        <span v-if="sqlMeta" class="opacity-60 text-13px">{{ sqlMeta }}</span>
      </n-space>
      <n-data-table
        :columns="sqlTableColumns"
        :data="sqlResultRows"
        :bordered="false"
        size="small"
        :scroll-x="Math.max(600, sqlResultColumns.length * 140)"
      />
    </n-card>
  </n-space>
</template>
