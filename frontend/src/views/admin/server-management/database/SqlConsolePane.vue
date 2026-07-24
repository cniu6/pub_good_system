<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { withSubmitLock } from '@/hooks'
import { adminApi } from '@/service/api/admin'

const props = defineProps<{
  selectedTable: string | null
  writeEnabled: boolean
}>()

const { t } = useI18n()
const message = useMessage()
const loading = ref(false)
const allowWrite = ref(false)
const sqlText = ref('SELECT 1')
const resultRows = ref<Record<string, unknown>[]>([])
const resultColumns = ref<string[]>([])
const metaText = ref('')

watch(() => props.selectedTable, (table) => {
  if (table)
    sqlText.value = `SELECT * FROM ${table} LIMIT 20`
})

const columns = computed<DataTableColumns<Record<string, unknown>>>(() =>
  resultColumns.value.map(key => ({
    title: key,
    key,
    minWidth: 140,
    ellipsis: { tooltip: true },
    render: row => h('span', String(row[key] ?? '')),
  })),
)

async function runSql() {
  const sql = sqlText.value.trim()
  if (!sql) {
    message.warning(t('adminServer.dbSqlEmpty'))
    return
  }
  await withSubmitLock(loading, async () => {
    const res = await adminApi.db.execSql({ sql, allow_write: allowWrite.value && props.writeEnabled })
    if (!res.isSuccess || !res.data) {
      message.error(res.message || t('adminServer.dbSqlFailed'))
      return
    }
    resultColumns.value = res.data.columns || []
    resultRows.value = res.data.rows || []
    const parts = [
      res.data.duration_ms != null ? `${res.data.duration_ms}ms` : '',
      res.data.row_count != null ? `${t('adminServer.dbResultRows')}: ${res.data.row_count}` : '',
      res.data.rows_affected != null ? `${t('adminServer.dbAffectedRows')}: ${res.data.rows_affected}` : '',
      res.data.truncated ? t('adminServer.dbTruncated') : '',
    ].filter(Boolean)
    metaText.value = parts.join(' · ')
    message.success(t('adminServer.dbSqlSuccess'))
  })
}
</script>

<template>
  <NSpace vertical :size="12">
    <NAlert type="info" :bordered="false">
      {{ writeEnabled ? t('adminServer.dbWriteEnabledHint') : t('adminServer.dbReadOnlyHint') }}
    </NAlert>
    <NInput
      v-model:value="sqlText"
      :placeholder="t('adminServer.dbSqlPlaceholder')"
      :rows="7"
      type="textarea"
    />
    <NSpace align="center">
      <NCheckbox v-model:checked="allowWrite" :disabled="!writeEnabled">
        {{ t('adminServer.dbAllowWrite') }}
      </NCheckbox>
      <NButton type="primary" :loading="loading" @click="runSql">
        {{ t('adminServer.dbRunSql') }}
      </NButton>
      <NText v-if="metaText" depth="3">
        {{ metaText }}
      </NText>
    </NSpace>
    <NDataTable
      :bordered="false"
      :columns="columns"
      :data="resultRows"
      :scroll-x="Math.max(640, resultColumns.length * 150)"
      size="small"
    />
  </NSpace>
</template>
