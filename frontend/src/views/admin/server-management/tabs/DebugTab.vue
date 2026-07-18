<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import { NTag, NText } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { RuntimeGoroutineInfo } from '@/service/api/admin/debug'
import { formatBytes, formatStorageFromMB, normalizePercent } from '@/utils/format'
import {
  extractProfileMetric,
  formatInteger,
  formatNsDuration,
  formatRuntimeTimestamp,
  getRuntimeStateTagType,
  useServerManagement,
  type PprofResultPanel,
  type RuntimeStateCategory,
} from '../composables/useServerManagement'

const { t } = useI18n()
const {
  monitoring,
  goroutineStats,
  runtimeStacks,
  longRunningStacks,
  potentialLeakStacks,
  runtimeStateSummaryMap,
  runtimeStacksLoaded,
  stackFilterMinWaitMinutes,
  cpuSeconds,
  traceSeconds,
  debugResults,
  debugLoading,
  debugAutoRefresh,
  forcingGC,
  getRuntimeStateSummaryLabel,
  getRuntimeStateSummaryTagType,
  loadGoroutineStats,
  handleForceGC,
  captureProfile,
  previewTraceProfile,
  downloadTraceProfile,
  clearAllPprofResults,
  loadRuntimeStacks,
  clearRuntimeStacks,
  toggleDebugAutoRefresh,
} = useServerManagement()

const processInfo = computed(() => monitoring.value?.process)
const hasAnyPprofResult = computed(() => Object.values(debugResults).some(value => Boolean(value)))
const potentialLeakIdSet = computed(() => new Set(potentialLeakStacks.value.map(stack => stack.id)))

const runtimeStateSummary = computed(() => {
  const orderedKeys: RuntimeStateCategory[] = ['running', 'waiting', 'channel', 'syscall', 'mutex', 'other']
  return orderedKeys
    .map(key => ({
      key,
      count: Number(runtimeStateSummaryMap.value[key] || 0),
      label: getRuntimeStateSummaryLabel(key),
      type: getRuntimeStateSummaryTagType(key),
    }))
    .filter(item => item.count > 0)
})

const heapProfileStats = computed(() => {
  if (!debugResults.heapText)
    return null
  const alloc = extractProfileMetric(debugResults.heapText, [
    /#\s*Alloc\s*=\s*(\d+)/i,
    /heap profile:\s*\d+:\s*(\d+)/i,
  ])
  const objects = extractProfileMetric(debugResults.heapText, [
    /#\s*HeapObjects\s*=\s*(\d+)/i,
    /#\s*objects\s*=\s*(\d+)/i,
  ])
  if (alloc == null && objects == null)
    return null
  return { alloc: alloc || 0, objects: objects || 0 }
})

const goroutineProfileCount = computed(() => {
  return (debugResults.goroutineText.match(/^goroutine\s+\d+\s+\[/gm) || []).length
})

const pprofResultPanels = computed<PprofResultPanel[]>(() => {
  const panels: PprofResultPanel[] = []
  if (debugResults.cpuText) {
    panels.push({
      key: 'cpu',
      title: t('adminSettings.cpuProfileResult'),
      text: debugResults.cpuText,
      maxHeight: 500,
      tags: [{ label: `${cpuSeconds.value}s`, type: 'success' }],
    })
  }
  if (debugResults.heapText) {
    const tags: PprofResultPanel['tags'] = []
    if (heapProfileStats.value) {
      if (heapProfileStats.value.alloc > 0)
        tags.push({ label: formatBytes(heapProfileStats.value.alloc), type: 'info' })
      if (heapProfileStats.value.objects > 0)
        tags.push({ label: `${formatInteger(heapProfileStats.value.objects)} ${t('adminSettings.heapObjects')}` })
    }
    panels.push({
      key: 'heap',
      title: t('adminSettings.heapProfileResult'),
      text: debugResults.heapText,
      maxHeight: 420,
      tags,
    })
  }
  if (debugResults.goroutineText) {
    panels.push({
      key: 'goroutine',
      title: t('adminSettings.goroutineProfileResult'),
      text: debugResults.goroutineText,
      maxHeight: 500,
      tags: [{ label: `${formatInteger(goroutineProfileCount.value)} ${t('adminSettings.goroutines')}`, type: 'info' }],
    })
  }
  if (debugResults.allocsText) {
    panels.push({ key: 'allocs', title: t('adminSettings.allocsProfileResult'), text: debugResults.allocsText, maxHeight: 420, tags: [] })
  }
  if (debugResults.blockText) {
    panels.push({ key: 'block', title: t('adminSettings.blockProfileResult'), text: debugResults.blockText, maxHeight: 420, tags: [] })
  }
  if (debugResults.mutexText) {
    panels.push({ key: 'mutex', title: t('adminSettings.mutexProfileResult'), text: debugResults.mutexText, maxHeight: 420, tags: [] })
  }
  if (debugResults.threadcreateText) {
    panels.push({ key: 'threadcreate', title: t('adminSettings.threadCreateProfileResult'), text: debugResults.threadcreateText, maxHeight: 420, tags: [] })
  }
  if (debugResults.traceText) {
    panels.push({ key: 'trace', title: t('adminSettings.traceProfileResult'), text: debugResults.traceText, maxHeight: 360, tags: [] })
  }
  return panels
})

const runtimeStackColumns: DataTableColumns<RuntimeGoroutineInfo> = [
  {
    title: '#',
    key: 'id',
    width: 72,
    render(row) {
      return h(NText, { code: true }, () => `#${row.id}`)
    },
  },
  {
    title: t('adminSettings.stackFunction'),
    key: 'function',
    minWidth: 240,
    ellipsis: { tooltip: true },
  },
  {
    title: t('adminSettings.columnStatus'),
    key: 'state',
    width: 150,
    render(row) {
      return h(NTag, { type: getRuntimeStateTagType(row.state) as any, size: 'small' }, () => row.state)
    },
  },
  {
    title: t('adminSettings.waitTime'),
    key: 'wait_time',
    width: 140,
    render(row) {
      return row.wait_time || '-'
    },
  },
  {
    title: t('adminSettings.createdBy'),
    key: 'created_by',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render(row) {
      return row.created_by || '-'
    },
  },
]

const runtimeStackSelectableColumnOptions = computed(() => [
  { key: 'id', label: '#' },
  { key: 'function', label: t('adminSettings.stackFunction') },
  { key: 'state', label: t('adminSettings.columnStatus') },
  { key: 'wait_time', label: t('adminSettings.waitTime') },
  { key: 'created_by', label: t('adminSettings.createdBy') },
])

const {
  columnOptions: runtimeStackColumnOptions,
  selectedColumnKeys: runtimeStackSelectedColumnKeys,
  visibleColumns: runtimeStackVisibleColumns,
  visibleColumnCount: runtimeStackVisibleColumnCount,
  totalColumnCount: runtimeStackTotalColumnCount,
  tableScrollX: runtimeStackTableScrollX,
  resetSelectedColumns: resetRuntimeStackSelectedColumns,
} = useTableColumnVisibility<RuntimeGoroutineInfo>({
  storageKey: 'admin-server-runtime-stack-preview',
  columns: computed(() => runtimeStackColumns),
  options: runtimeStackSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})
</script>

<template>
  <NSpace vertical :size="16">
    <NCard :title="t('adminSettings.systemOverview')" size="small">
      <template #header-extra>
        <NSpace>
          <NButton size="small" :type="debugAutoRefresh ? 'primary' : 'default'" @click="toggleDebugAutoRefresh(!debugAutoRefresh)">
            {{ debugAutoRefresh ? t('adminSettings.stopRefresh') : t('adminSettings.autoRefresh') }}
          </NButton>
          <NButton size="small" :loading="debugLoading.goroutineStats" @click="loadGoroutineStats">{{ t('adminSettings.refresh') }}</NButton>
          <NButton size="small" type="warning" :loading="forcingGC" @click="handleForceGC">{{ t('adminSettings.forceGC') }}</NButton>
        </NSpace>
      </template>
      <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 m:2 l:2" responsive="screen">
        <NGi>
          <NCard size="small" :title="t('adminSettings.processResources')">
            <NSpace vertical size="small">
              <div>
                <NSpace justify="space-between">
                  <NText>{{ t('adminSettings.cpu') }}</NText>
                  <NText>{{ Number((processInfo?.process_cpu || 0).toFixed(1)) }}%</NText>
                </NSpace>
                <NProgress
                  type="line"
                  :percentage="normalizePercent(processInfo?.process_cpu || 0)"
                  :status="(processInfo?.process_cpu ?? 0) > 80 ? 'error' : 'success'"
                  :show-indicator="false"
                  style="margin-top: 4px; transform: scaleY(0.7); transform-origin: center;"
                />
              </div>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.memory') }}</NText><NText>{{ formatStorageFromMB(processInfo?.process_rss_mb || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.goroutines') }}</NText><NText>{{ formatInteger(processInfo?.goroutines || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.pid') }}</NText><NText>{{ processInfo?.pid || '-' }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.gomaxprocs') }}</NText><NText>{{ formatInteger(goroutineStats?.gomaxprocs || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.numCPU') }}</NText><NText>{{ formatInteger(goroutineStats?.num_cpu || 0) }}</NText></NSpace>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.goroutineStats')">
            <NSpace vertical size="small">
              <NSpace justify="space-between"><NText>{{ t('adminSettings.runtimeTotal') }}</NText><NText>{{ formatInteger(goroutineStats?.total_count || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.tracked') }}</NText><NText>{{ formatInteger(goroutineStats?.tracked_count || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.heapAlloc') }}</NText><NText>{{ formatBytes(goroutineStats?.mem_stats?.heap_alloc || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.heapInUse') }}</NText><NText>{{ formatBytes(goroutineStats?.mem_stats?.heap_inuse || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.gcCount') }}</NText><NText>{{ formatInteger(goroutineStats?.mem_stats?.num_gc || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.totalAlloc') }}</NText><NText>{{ formatBytes(goroutineStats?.mem_stats?.total_alloc || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.heapObjects') }}</NText><NText>{{ formatInteger(goroutineStats?.mem_stats?.heap_objects || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.numCgoCall') }}</NText><NText>{{ formatInteger(goroutineStats?.num_cgo_call || 0) }}</NText></NSpace>
              <NSpace justify="space-between"><NText>{{ t('adminSettings.potentialLeaks') }}</NText><NText type="error">{{ formatInteger(goroutineStats?.potential_leaks || 0) }}</NText></NSpace>
            </NSpace>
          </NCard>
        </NGi>
      </NGrid>

      <NCard size="small" :title="t('adminSettings.memoryDetails')" style="margin-top: 12px;">
        <NGrid :x-gap="10" :y-gap="10" cols="1 s:2 m:4 l:4" responsive="screen">
          <NGi><NStatistic :label="t('adminSettings.heapSys')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.heap_sys || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.heapIdle')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.heap_idle || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.heapReleased')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.heap_released || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.stackInUse')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.stack_inuse || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.stackSys')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.stack_sys || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.runtimeSys')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.sys || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.mallocs')"><template #default>{{ formatInteger(goroutineStats?.mem_stats?.mallocs || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.frees')"><template #default>{{ formatInteger(goroutineStats?.mem_stats?.frees || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.nextGC')"><template #default>{{ formatBytes(goroutineStats?.mem_stats?.next_gc || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.lastGC')"><template #default>{{ formatRuntimeTimestamp(goroutineStats?.mem_stats?.last_gc || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.pauseTotal')"><template #default>{{ formatNsDuration(goroutineStats?.mem_stats?.pause_total_ns || 0) }}</template></NStatistic></NGi>
          <NGi><NStatistic :label="t('adminSettings.forcedGC')"><template #default>{{ formatInteger(goroutineStats?.mem_stats?.num_forced_gc || 0) }}</template></NStatistic></NGi>
        </NGrid>
      </NCard>
    </NCard>

    <NCard :title="t('adminSettings.pprofTitle')" size="small">
      <template #header-extra>
        <NButton size="small" @click="clearAllPprofResults">{{ t('adminSettings.clearResults') }}</NButton>
      </template>

      <NGrid :x-gap="12" :y-gap="12" cols="1 s:2 m:3 l:3" responsive="screen">
        <NGi>
          <NCard size="small" :title="t('adminSettings.cpuProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.cpuProfileDesc') }}</NText>
              <NSpace>
                <NInputNumber v-model:value="cpuSeconds" :min="5" :max="120" size="small" style="width: 90px" />
                <NButton size="small" type="primary" :loading="debugLoading.cpu" @click="captureProfile('cpu')">{{ t('adminSettings.capture') }}</NButton>
              </NSpace>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.heapProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.heapProfileDesc') }}</NText>
              <NButton size="small" type="primary" :loading="debugLoading.heap" @click="captureProfile('heap')">{{ t('adminSettings.capture') }}</NButton>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.goroutineProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.goroutineProfileDesc') }}</NText>
              <NButton size="small" type="primary" :loading="debugLoading.goroutine" @click="captureProfile('goroutine')">{{ t('adminSettings.capture') }}</NButton>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.allocsProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.allocsProfileDesc') }}</NText>
              <NButton size="small" type="primary" :loading="debugLoading.allocs" @click="captureProfile('allocs')">{{ t('adminSettings.capture') }}</NButton>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.blockProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.blockProfileDesc') }}</NText>
              <NButton size="small" type="primary" :loading="debugLoading.block" @click="captureProfile('block')">{{ t('adminSettings.capture') }}</NButton>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.mutexProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.mutexProfileDesc') }}</NText>
              <NButton size="small" type="primary" :loading="debugLoading.mutex" @click="captureProfile('mutex')">{{ t('adminSettings.capture') }}</NButton>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.threadCreateProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.threadCreateProfileDesc') }}</NText>
              <NButton size="small" type="primary" :loading="debugLoading.threadcreate" @click="captureProfile('threadcreate')">{{ t('adminSettings.capture') }}</NButton>
            </NSpace>
          </NCard>
        </NGi>
        <NGi>
          <NCard size="small" :title="t('adminSettings.traceProfile')">
            <NSpace vertical size="small">
              <NText depth="3">{{ t('adminSettings.traceProfileDesc') }}</NText>
              <NSpace>
                <NInputNumber v-model:value="traceSeconds" :min="1" :max="30" size="small" style="width: 90px" />
                <NButton size="small" type="primary" :loading="debugLoading.trace" @click="previewTraceProfile">{{ t('common.preview') }}</NButton>
                <NButton size="small" :loading="debugLoading.trace" @click="downloadTraceProfile">{{ t('adminSettings.downloadBinary') }}</NButton>
              </NSpace>
            </NSpace>
          </NCard>
        </NGi>
      </NGrid>

      <NEmpty v-if="!hasAnyPprofResult" :description="t('adminSettings.clickToCapture')" style="margin-top: 16px" />
      <NCollapse v-else style="margin-top: 16px">
        <NCollapseItem
          v-for="panel in pprofResultPanels"
          :key="panel.key"
          :title="panel.title"
          :name="panel.key"
        >
          <template #header-extra>
            <NSpace v-if="panel.tags.length > 0" size="small">
              <NTag v-for="tag in panel.tags" :key="`${panel.key}-${tag.label}`" :type="tag.type as any" size="small">
                {{ tag.label }}
              </NTag>
            </NSpace>
          </template>
          <NCode
            :code="panel.text"
            language="text"
            word-wrap
            :style="{ maxHeight: `${panel.maxHeight}px`, overflow: 'auto' }"
          />
        </NCollapseItem>
      </NCollapse>
    </NCard>

    <NCard :title="t('adminSettings.runtimeStacks')" size="small">
      <template #header-extra>
        <NSpace>
          <TableColumnSelector
            v-model="runtimeStackSelectedColumnKeys"
            :options="runtimeStackColumnOptions"
            :visible-count="runtimeStackVisibleColumnCount"
            :total-count="runtimeStackTotalColumnCount"
            :button-label="t('common.showFields')"
            :title="t('common.visibleFields')"
            :hint="t('common.columnVisibilityHint')"
            :reset-label="t('common.restoreDefaultFields')"
            @reset="resetRuntimeStackSelectedColumns"
          />
          <NTooltip trigger="hover">
            <template #trigger>
              <NInputNumber v-model:value="stackFilterMinWaitMinutes" :min="0" size="small" style="width: 140px" :placeholder="t('adminSettings.minWaitMinutes')" />
            </template>
            {{ t('adminSettings.filterTooltip') }}
          </NTooltip>
          <NButton size="small" :loading="debugLoading.stacks" @click="loadRuntimeStacks">{{ t('adminSettings.loadStacks') }}</NButton>
          <NButton size="small" @click="clearRuntimeStacks">{{ t('adminSettings.clearStacks') }}</NButton>
        </NSpace>
      </template>
      <NAlert v-if="potentialLeakStacks.length > 0" type="error" :title="t('adminSettings.suspectedLeakTitle')" style="margin-bottom: 12px">
        {{ t('adminSettings.suspectedLeakHint', { count: potentialLeakStacks.length }) }}
        <NDataTable
          :columns="runtimeStackVisibleColumns"
          :data="potentialLeakStacks"
          size="small"
          :bordered="false"
          :pagination="{ pageSize: 5 }"
          :scroll-x="runtimeStackTableScrollX"
          style="margin-top: 12px"
        />
      </NAlert>

      <NCollapse v-if="longRunningStacks.length > 0" style="margin-bottom: 12px">
        <NCollapseItem :title="t('adminSettings.longRunningTitle', { count: longRunningStacks.length })" name="long-running-preview">
          <NDataTable
            :columns="runtimeStackVisibleColumns"
            :data="longRunningStacks"
            size="small"
            :bordered="false"
            :pagination="{ pageSize: 10 }"
            :scroll-x="runtimeStackTableScrollX"
          />
        </NCollapseItem>
      </NCollapse>

      <NEmpty v-if="!runtimeStacksLoaded" :description="t('adminSettings.clickToLoadStacks')" />
      <template v-else>
        <NSpace wrap size="small" style="margin-bottom: 12px">
          <NTag type="info">{{ t('adminSettings.runtimeTotal') }}: {{ formatInteger(runtimeStacks.length) }}</NTag>
          <NTag
            v-for="summary in runtimeStateSummary"
            :key="summary.key"
            :type="summary.type as any"
            size="small"
          >
            {{ summary.label }}: {{ formatInteger(summary.count) }}
          </NTag>
        </NSpace>

        <NCollapse v-if="runtimeStacks.length > 0" accordion>
          <NCollapseItem v-for="stack in runtimeStacks" :key="stack.id" :name="stack.id">
            <template #header>
              <NSpace align="center">
                <NText code>#{{ stack.id }}</NText>
                <NText>{{ stack.function }}</NText>
                <NTag v-if="potentialLeakIdSet.has(stack.id)" type="error" size="small">{{ t('adminSettings.suspectedLeakTag') }}</NTag>
                <NTag v-if="stack.locked_to_thread" type="warning" size="small">{{ t('adminSettings.lockedToThread') }}</NTag>
              </NSpace>
            </template>
            <template #header-extra>
              <NSpace size="small">
                <NTag :type="getRuntimeStateTagType(stack.state) as any" size="small">{{ stack.state }}</NTag>
                <NText v-if="stack.wait_time" depth="3" style="font-size: 12px">{{ stack.wait_time }}</NText>
                <NText depth="3" style="font-size: 11px">{{ t('adminSettings.stackLines', { count: stack.stack_lines }) }}</NText>
              </NSpace>
            </template>
            <NSpace vertical>
              <NText v-if="stack.created_by" depth="3" style="font-size: 12px">
                {{ t('adminSettings.createdBy') }}: {{ stack.created_by }}
              </NText>
              <NCode :code="stack.stack" language="text" word-wrap style="max-height: 420px; overflow: auto; font-size: 12px;" />
            </NSpace>
          </NCollapseItem>
        </NCollapse>
        <NEmpty v-else :description="t('adminSettings.noStacksMatchFilter')" />
      </template>
    </NCard>
  </NSpace>
</template>
