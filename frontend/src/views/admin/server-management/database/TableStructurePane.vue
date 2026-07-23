<script setup lang="ts">
import { computed, h } from 'vue'
import { NTag, NText } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { DbColumnMeta, DbIndexMeta, DbTableMeta } from '@/service/api/admin/db'

defineProps<{
  loading: boolean
  meta: DbTableMeta | null
}>()

const { t } = useI18n()

const columnColumns = computed<DataTableColumns<DbColumnMeta>>(() => [
  { title: t('adminServer.dbColumnName'), key: 'name', minWidth: 150 },
  { title: t('adminServer.dbColumnType'), key: 'type', minWidth: 140 },
  {
    title: t('adminServer.dbNullable'),
    key: 'nullable',
    width: 90,
    render: row => h(NTag, { size: 'small', type: row.nullable ? 'warning' : 'success' }, () => row.nullable ? t('adminServer.dbYes') : t('adminServer.dbNo')),
  },
  { title: t('adminServer.dbDefaultValue'), key: 'default_value', minWidth: 140, ellipsis: { tooltip: true } },
  {
    title: t('adminServer.dbConstraints'),
    key: 'constraints',
    minWidth: 140,
    render: (row) => {
      const values = [
        row.primary_key ? t('adminServer.dbPrimaryKey') : '',
        row.auto_increment ? t('adminServer.dbAutoIncrement') : '',
      ].filter(Boolean)
      return values.length ? h(NText, null, () => values.join(' · ')) : '-'
    },
  },
  { title: t('adminServer.dbColumnComment'), key: 'comment', minWidth: 180, ellipsis: { tooltip: true } },
])

const indexColumns = computed<DataTableColumns<DbIndexMeta>>(() => [
  { title: t('adminServer.dbIndexName'), key: 'name', minWidth: 180 },
  { title: t('adminServer.dbIndexColumns'), key: 'columns', minWidth: 220, render: row => row.columns.join(', ') },
  {
    title: t('adminServer.dbIndexType'),
    key: 'type',
    minWidth: 140,
    render: (row) => {
      const labels = [
        row.primary_key ? t('adminServer.dbPrimaryKey') : '',
        row.unique ? t('adminServer.dbUniqueIndex') : '',
        row.columns.length > 1 ? t('adminServer.dbCompositeIndex') : '',
      ].filter(Boolean)
      return h(NTag, { size: 'small', type: row.primary_key ? 'success' : row.unique ? 'info' : 'default' }, () => labels.join(' · ') || t('adminServer.dbNormalIndex'))
    },
  },
])
</script>

<template>
  <NSpace vertical :size="16">
    <NAlert v-if="meta?.comment" type="info" :bordered="false">
      {{ meta.comment }}
    </NAlert>
    <NCard size="small" :title="t('adminServer.dbColumns')" :segmented="{ content: true }">
      <NDataTable
        :bordered="false"
        :columns="columnColumns"
        :data="meta?.columns || []"
        :loading="loading"
        :scroll-x="900"
        size="small"
      />
    </NCard>
    <NCard size="small" :title="t('adminServer.dbIndexes')" :segmented="{ content: true }">
      <NDataTable
        :bordered="false"
        :columns="indexColumns"
        :data="meta?.indexes || []"
        :loading="loading"
        :scroll-x="640"
        size="small"
      />
    </NCard>
    <NCard v-if="meta?.foreign_keys?.length" size="small" :title="t('adminServer.dbForeignKeys')" :segmented="{ content: true }">
      <NSpace vertical size="small">
        <NText v-for="foreignKey in meta.foreign_keys" :key="foreignKey.name" depth="2">
          {{ foreignKey.name }}: {{ foreignKey.columns.join(', ') }} → {{ foreignKey.ref_table }} ({{ foreignKey.ref_columns.join(', ') }})
        </NText>
      </NSpace>
    </NCard>
  </NSpace>
</template>
