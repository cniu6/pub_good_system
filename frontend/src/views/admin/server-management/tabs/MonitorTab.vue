<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import { NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  formatBytes,
  formatPercent,
  formatStorageFromGB,
  formatStorageFromMB,
  formatUptime,
} from '@/utils/format'
import {
  formatInteger,
  useServerManagement,
  type ServiceHealthRow,
} from '../composables/useServerManagement'

const { t } = useI18n()
const {
  monitoring,
  monitoringLoading,
  operationsLoading,
  autoRefresh,
  toggleAutoRefresh,
  refreshAll,
} = useServerManagement()

const processInfo = computed(() => monitoring.value?.process)
const metricInfo = computed(() => monitoring.value?.metrics)
const serviceList = computed(() => monitoring.value?.services || [])

const serviceHealthColumns: DataTableColumns<ServiceHealthRow> = [
  { title: t('adminSettings.columnService'), key: 'name', minWidth: 140 },
  {
    title: t('adminSettings.columnStatus'),
    key: 'status',
    width: 110,
    render: (row: ServiceHealthRow) => h(
      NTag,
      { type: row.status === 'up' ? 'success' : row.status === 'warning' ? 'warning' : 'error', size: 'small' },
      () => row.status === 'up' ? t('adminSettings.statusNormal') : row.status === 'warning' ? t('adminSettings.statusWarning') : t('adminSettings.statusError'),
    ),
  },
  { title: t('adminSettings.columnMessage'), key: 'message', minWidth: 220 },
]

const serviceHealthSelectableColumnOptions = computed(() => [
  { key: 'name', label: t('adminSettings.columnService') },
  { key: 'status', label: t('adminSettings.columnStatus') },
  { key: 'message', label: t('adminSettings.columnMessage') },
])

const {
  columnOptions: serviceHealthColumnOptions,
  selectedColumnKeys: serviceHealthSelectedColumnKeys,
  visibleColumns: serviceHealthVisibleColumns,
  visibleColumnCount: serviceHealthVisibleColumnCount,
  totalColumnCount: serviceHealthTotalColumnCount,
  tableScrollX: serviceHealthTableScrollX,
  resetSelectedColumns: resetServiceHealthSelectedColumns,
} = useTableColumnVisibility<ServiceHealthRow>({
  storageKey: 'admin-server-service-health',
  columns: serviceHealthColumns,
  options: serviceHealthSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 620,
})
</script>

<template>
  <NSpace vertical :size="16">
    <NCard>
      <template #header>
        <NSpace align="center" justify="space-between" style="width: 100%;">
          <NText strong>{{ t('adminSettings.realtimeMonitor') }}</NText>
          <NSpace align="center">
            <NText depth="3">{{ t('adminSettings.autoRefresh') }}</NText>
            <NSwitch :value="autoRefresh" @update:value="toggleAutoRefresh" />
            <NButton size="small" type="primary" :loading="monitoringLoading || operationsLoading" @click="refreshAll">{{ t('adminSettings.refresh') }}</NButton>
          </NSpace>
        </NSpace>
      </template>

      <NGrid :x-gap="12" :y-gap="12" cols="2 s:2 m:4 l:4" responsive="screen">
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.cpu')"><template #default>{{ formatPercent(metricInfo?.cpu.usage_percent || 0, 2) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ t('adminSettings.cpuCores', { count: metricInfo?.cpu.core_count || 0 }) }}</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.systemMemory')"><template #default>{{ formatPercent(metricInfo?.memory.used_percent || 0, 2) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ formatStorageFromMB(metricInfo?.memory.used_mb || 0) }}/{{ formatStorageFromMB(metricInfo?.memory.total_mb || 0) }}</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.swap')"><template #default>{{ formatPercent(metricInfo?.swap.used_percent || 0, 2) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ formatStorageFromMB(metricInfo?.swap.used_mb || 0) }}/{{ formatStorageFromMB(metricInfo?.swap.total_mb || 0) }}</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.diskUsage')"><template #default>{{ formatPercent(metricInfo?.disk.used_percent || 0, 2) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ formatStorageFromGB(metricInfo?.disk.used_gb || 0) }}/{{ formatStorageFromGB(metricInfo?.disk.total_gb || 0) }}</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.processMemory')"><template #default>{{ formatStorageFromMB(processInfo?.process_rss_mb || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">CPU {{ Number((processInfo?.process_cpu || 0).toFixed(2)) }}%</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.goHeap')"><template #default>{{ formatStorageFromMB(processInfo?.heap_alloc_mb || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">sys {{ formatStorageFromMB(processInfo?.memory_sys_mb || 0) }}</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.goroutines')" :value="formatInteger(processInfo?.goroutines || 0)"><template #suffix><NText depth="3" style="font-size: 10px">GC {{ formatInteger(processInfo?.gc_count || 0) }}</NText></template></NStatistic></NCard></NGi>
        <NGi><NCard size="small" embedded><NStatistic :label="t('adminSettings.uptime')"><template #default>{{ formatUptime(monitoring?.uptime_seconds || 0) }}</template><template #suffix><NText depth="3" style="font-size: 10px">{{ monitoring?.generated_at || '-' }}</NText></template></NStatistic></NCard></NGi>
      </NGrid>

      <NDivider />

      <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 m:2 l:2" responsive="screen">
        <NGi>
          <NCard size="small" :title="t('adminSettings.systemInfo')">
            <NDescriptions bordered :column="2" label-placement="left">
              <NDescriptionsItem :label="t('adminSettings.appName')">{{ monitoring?.app?.name || '-' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('adminSettings.systemVersion')">{{ monitoring?.app?.go_version || '-' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('adminSettings.appMode')">{{ monitoring?.app?.mode || '-' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('adminSettings.port')">{{ monitoring?.app?.port || '-' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('adminSettings.pid')">{{ processInfo?.pid || '-' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('adminSettings.lastRefreshed')">{{ monitoring?.generated_at || '-' }}</NDescriptionsItem>
            </NDescriptions>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.network')">
            <NSpace vertical size="small">
              <NStatistic :label="t('adminSettings.network')"><template #default>{{ formatBytes((metricInfo?.network.bytes_sent || 0) + (metricInfo?.network.bytes_recv || 0)) }}</template></NStatistic>
              <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.upload') }}</NText><NText>{{ formatBytes(metricInfo?.network.bytes_sent || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.download') }}</NText><NText>{{ formatBytes(metricInfo?.network.bytes_recv || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.uploadPackets') }}</NText><NText>{{ formatInteger(metricInfo?.network.packets_sent || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText depth="3">{{ t('adminSettings.downloadPackets') }}</NText><NText>{{ formatInteger(metricInfo?.network.packets_recv || 0) }}</NText></NSpace>
            </NSpace>
          </NCard>
        </NGi>
      </NGrid>

      <NCard size="small" :title="t('adminSettings.memoryDetails')" style="margin-top: 12px;">
        <NGrid :x-gap="10" :y-gap="10" cols="1 s:2 m:4 l:4" responsive="screen">
          <NGi><NStatistic :label="t('adminSettings.goMemoryAlloc')"><template #default>{{ formatStorageFromMB(processInfo?.memory_alloc_mb || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.goMemorySys')"><template #default>{{ formatStorageFromMB(processInfo?.memory_sys_mb || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.heapAlloc')"><template #default>{{ formatStorageFromMB(processInfo?.heap_alloc_mb || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.heapInUse')"><template #default>{{ formatStorageFromMB(processInfo?.heap_inuse_mb || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.heapIdle')"><template #default>{{ formatStorageFromMB(processInfo?.heap_idle_mb || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.stackInUse')"><template #default>{{ formatStorageFromMB(processInfo?.stack_inuse_mb || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.gcCount')"><template #default>{{ formatInteger(processInfo?.gc_count || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.gcCPU')"><template #default>{{ Number(((processInfo?.gc_cpu_fraction || 0) * 100).toFixed(4)) }}%</template></NStatistic></NGi>
        </NGrid>
      </NCard>
    </NCard>

    <NCard :title="t('adminSettings.serviceHealthSnapshot')">
      <template #header-extra>
        <TableColumnSelector
          v-model="serviceHealthSelectedColumnKeys"
          :options="serviceHealthColumnOptions"
          :visible-count="serviceHealthVisibleColumnCount"
          :total-count="serviceHealthTotalColumnCount"
          :button-label="t('common.showFields')"
          :title="t('common.visibleFields')"
          :hint="t('common.columnVisibilityHint')"
          :reset-label="t('common.restoreDefaultFields')"
          @reset="resetServiceHealthSelectedColumns"
        />
      </template>
      <NDataTable :columns="serviceHealthVisibleColumns" :data="serviceList" :pagination="false" :scroll-x="serviceHealthTableScrollX" />
    </NCard>
  </NSpace>
</template>
