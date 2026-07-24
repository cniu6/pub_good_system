import { computed, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { withSubmitLock } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import { authStorage } from '@/utils'
import type { DbTableMeta } from '@/service/api/admin/db'

export function useDatabaseConsole() {
  const { t } = useI18n()
  const message = useMessage()
  const loading = ref(false)
  const metaLoading = ref(false)
  const ddlLoading = ref(false)
  const backupLoading = ref(false)
  /** 增删改行串行锁，防止连点并发写库 */
  const writeLock = ref(false)
  const driver = ref('')
  const backupSupported = ref(false)
  const writeEnabled = ref(false)
  const tables = ref<string[]>([])
  const selectedTable = ref<string | null>(null)
  const rows = ref<Record<string, unknown>[]>([])
  const rowColumns = ref<string[]>([])
  const meta = ref<DbTableMeta | null>(null)
  const ddl = ref('')
  const query = reactive({ page: 1, page_size: 20, total: 0 })

  const selectedTableLabel = computed(() => selectedTable.value || t('adminServer.dbNoSelection'))

  async function loadInfoAndTables() {
    loading.value = true
    try {
      const [infoRes, tablesRes] = await Promise.all([adminApi.db.info(), adminApi.db.tables()])
      if (infoRes.isSuccess && infoRes.data) {
        driver.value = infoRes.data.driver || ''
        backupSupported.value = Boolean(infoRes.data.backup_supported)
        writeEnabled.value = Boolean(infoRes.data.write_enabled)
      }
      if (!tablesRes.isSuccess || !tablesRes.data) {
        message.error(tablesRes.message || t('adminServer.dbLoadFailed'))
        return
      }
      tables.value = tablesRes.data.list || []
      if (selectedTable.value && !tables.value.includes(selectedTable.value))
        resetSelection()
    }
    catch {
      message.error(t('adminServer.dbLoadFailed'))
    }
    finally {
      loading.value = false
    }
  }

  function resetSelection() {
    selectedTable.value = null
    rows.value = []
    rowColumns.value = []
    meta.value = null
    ddl.value = ''
    query.total = 0
  }

  async function selectTable(name: string | null) {
    if (!name) {
      resetSelection()
      return
    }
    selectedTable.value = name
    query.page = 1
    await Promise.all([loadRows(), loadMeta(), loadDdl()])
  }

  async function loadRows() {
    if (!selectedTable.value)
      return
    loading.value = true
    try {
      const res = await adminApi.db.tableRows(selectedTable.value, { page: query.page, page_size: query.page_size })
      if (!res.isSuccess || !res.data) {
        message.error(res.message || t('adminServer.dbLoadFailed'))
        return
      }
      rowColumns.value = res.data.columns || []
      rows.value = res.data.rows || []
      query.total = Number(res.data.total || 0)
    }
    finally {
      loading.value = false
    }
  }

  async function loadMeta() {
    if (!selectedTable.value)
      return
    metaLoading.value = true
    try {
      const res = await adminApi.db.tableMeta(selectedTable.value)
      if (res.isSuccess && res.data)
        meta.value = res.data
      else
        message.error(res.message || t('adminServer.dbMetaLoadFailed'))
    }
    finally {
      metaLoading.value = false
    }
  }

  async function loadDdl() {
    if (!selectedTable.value)
      return
    ddlLoading.value = true
    try {
      const res = await adminApi.db.tableDdl(selectedTable.value)
      if (res.isSuccess && res.data)
        ddl.value = res.data.ddl || ''
      else
        message.error(res.message || t('adminServer.dbDdlLoadFailed'))
    }
    finally {
      ddlLoading.value = false
    }
  }

  async function createRow(values: Record<string, unknown>) {
    if (!selectedTable.value)
      return false
    const ok = await withSubmitLock(writeLock, async () => {
      const res = await adminApi.db.createTableRow(selectedTable.value!, values)
      if (!res.isSuccess) {
        message.error(res.message || t('adminServer.dbRowCreateFailed'))
        return false
      }
      message.success(t('adminServer.dbRowCreateSuccess'))
      await loadRows()
      return true
    })
    return ok ?? false
  }

  async function updateRow(primaryKey: Record<string, unknown>, values: Record<string, unknown>) {
    if (!selectedTable.value)
      return false
    const ok = await withSubmitLock(writeLock, async () => {
      const res = await adminApi.db.updateTableRow(selectedTable.value!, primaryKey, values)
      if (!res.isSuccess) {
        message.error(res.message || t('adminServer.dbRowUpdateFailed'))
        return false
      }
      message.success(t('adminServer.dbRowUpdateSuccess'))
      await loadRows()
      return true
    })
    return ok ?? false
  }

  async function deleteRow(primaryKey: Record<string, unknown>) {
    if (!selectedTable.value)
      return false
    const ok = await withSubmitLock(writeLock, async () => {
      const res = await adminApi.db.deleteTableRow(selectedTable.value!, primaryKey)
      if (!res.isSuccess) {
        message.error(res.message || t('adminServer.dbRowDeleteFailed'))
        return false
      }
      message.success(t('adminServer.dbRowDeleteSuccess'))
      await loadRows()
      return true
    })
    return ok ?? false
  }

  async function downloadBackup() {
    if (!backupSupported.value)
      return
    await withSubmitLock(backupLoading, async () => {
      try {
        const headers: Record<string, string> = {}
        const token = authStorage.get('accessToken')
        if (token)
          headers.Authorization = `Bearer ${token}`
        const response = await fetch(await adminApi.db.backupUrl(), { headers })
        if (!response.ok)
          throw new Error((await response.json() as { message?: string }).message || t('adminServer.dbBackupFailed'))
        const objectUrl = URL.createObjectURL(await response.blob())
        const link = document.createElement('a')
        link.href = objectUrl
        link.download = `fst-backup-${Date.now()}.db`
        link.click()
        URL.revokeObjectURL(objectUrl)
        message.success(t('adminServer.dbBackupSuccess'))
      }
      catch (error: any) {
        message.error(error?.message || t('adminServer.dbBackupFailed'))
      }
    })
  }

  return {
    backupLoading,
    backupSupported,
    ddl,
    ddlLoading,
    driver,
    loadDdl,
    loadInfoAndTables,
    loadMeta,
    loadRows,
    loading,
    meta,
    metaLoading,
    query,
    rowColumns,
    rows,
    selectedTable,
    selectedTableLabel,
    selectTable,
    tables,
    writeEnabled,
    createRow,
    updateRow,
    deleteRow,
    downloadBackup,
  }
}
