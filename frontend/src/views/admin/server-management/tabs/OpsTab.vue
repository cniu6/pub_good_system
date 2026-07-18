<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import { NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { DynamicRateLimitSnapshot } from '@/service/api/admin/server'
import { useServerManagement } from '../composables/useServerManagement'

const { t } = useI18n()
const {
  operations,
  operationsLoading,
  runtimeForm,
  savingRuntime,
  restartLoading,
  saveRuntimeConfig,
  handleRestartBackend,
  goToAutoJobs,
} = useServerManagement()

const rateLimits = computed(() => operations.value?.rate_limits || [])

const rateLimitColumns: DataTableColumns<DynamicRateLimitSnapshot> = [
  { title: t('adminServer.rateLimit.name'), key: 'name', minWidth: 130 },
  {
    title: t('adminServer.rateLimit.enabled'),
    key: 'enabled',
    width: 90,
    render(row) {
      return h(NTag, { type: row.enabled ? 'success' : 'default', size: 'small' }, () => row.enabled ? t('common.enable') : t('common.disable'))
    },
  },
  { title: t('adminServer.rateLimit.rate'), key: 'rate', width: 90 },
  { title: t('adminServer.rateLimit.burst'), key: 'burst', width: 90 },
  { title: t('adminServer.rateLimit.allowed'), key: 'allowed_count', width: 100 },
  { title: t('adminServer.rateLimit.blocked'), key: 'blocked_count', width: 100 },
  { title: t('adminServer.rateLimit.visitors'), key: 'active_visitors', width: 110 },
  { title: t('adminServer.rateLimit.lastReload'), key: 'last_config_reload', minWidth: 180, render: row => row.last_config_reload || '-' },
]

const rateLimitSelectableColumnOptions = computed(() => [
  { key: 'name', label: t('adminServer.rateLimit.name') },
  { key: 'enabled', label: t('adminServer.rateLimit.enabled') },
  { key: 'rate', label: t('adminServer.rateLimit.rate') },
  { key: 'burst', label: t('adminServer.rateLimit.burst') },
  { key: 'allowed_count', label: t('adminServer.rateLimit.allowed') },
  { key: 'blocked_count', label: t('adminServer.rateLimit.blocked') },
  { key: 'active_visitors', label: t('adminServer.rateLimit.visitors') },
  { key: 'last_config_reload', label: t('adminServer.rateLimit.lastReload') },
])

const {
  columnOptions: rateLimitColumnOptions,
  selectedColumnKeys: rateLimitSelectedColumnKeys,
  visibleColumns: rateLimitVisibleColumns,
  visibleColumnCount: rateLimitVisibleColumnCount,
  totalColumnCount: rateLimitTotalColumnCount,
  tableScrollX: rateLimitTableScrollX,
  resetSelectedColumns: resetRateLimitSelectedColumns,
} = useTableColumnVisibility<DynamicRateLimitSnapshot>({
  storageKey: 'admin-server-rate-limits',
  columns: rateLimitColumns,
  options: rateLimitSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 920,
})
</script>

<template>
  <NSpace vertical :size="16">
    <NCard :title="t('adminServer.tasks.title')">
      <NAlert type="info" :bordered="false" style="margin-bottom: 12px;">
        {{ t('adminServer.tasks.movedHint') }}
      </NAlert>
      <NButton type="primary" @click="goToAutoJobs">{{ t('adminServer.tasks.goToAutoJobs') }}</NButton>
    </NCard>

    <NCard :title="t('adminServer.rateLimit.title')">
      <template #header-extra>
        <TableColumnSelector
          v-model="rateLimitSelectedColumnKeys"
          :options="rateLimitColumnOptions"
          :visible-count="rateLimitVisibleColumnCount"
          :total-count="rateLimitTotalColumnCount"
          :button-label="t('common.showFields')"
          :title="t('common.visibleFields')"
          :hint="t('common.columnVisibilityHint')"
          :reset-label="t('common.restoreDefaultFields')"
          @reset="resetRateLimitSelectedColumns"
        />
      </template>
      <NDataTable :columns="rateLimitVisibleColumns" :data="rateLimits" :loading="operationsLoading" :pagination="false" :scroll-x="rateLimitTableScrollX" />
    </NCard>

    <NCard :title="t('adminServer.runtimeConfig.title')">
      <NGrid cols="2" :x-gap="24" :y-gap="18">
        <NGi>
          <NSpace vertical>
            <NSpace align="center" justify="space-between"><NText strong>{{ t('adminServer.runtimeConfig.apiLog') }}</NText><NSwitch v-model:value="runtimeForm.api_access_log_enabled" /></NSpace>
            <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.runtimeConfig.queryDays') }}</NText><NInputNumber v-model:value="runtimeForm.api_log_query_days" :min="1" :max="365" /></NSpace>
            <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.runtimeConfig.maxCount') }}</NText><NInputNumber v-model:value="runtimeForm.api_log_max_count" :min="100" :max="200000" /></NSpace>
          </NSpace>
        </NGi>
        <NGi>
          <NSpace vertical>
            <NSpace align="center" justify="space-between"><NText strong>{{ t('adminServer.runtimeConfig.globalRateLimit') }}</NText><NSwitch v-model:value="runtimeForm.api_rate_limit_enabled" /></NSpace>
            <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.rate') }}</NText><NInputNumber v-model:value="runtimeForm.api_rate_limit_rate" :min="1" :max="10000" /></NSpace>
            <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.burst') }}</NText><NInputNumber v-model:value="runtimeForm.api_rate_limit_burst" :min="1" :max="20000" /></NSpace>
          </NSpace>
        </NGi>
        <NGi>
          <NSpace vertical>
            <NSpace align="center" justify="space-between"><NText strong>{{ t('adminServer.runtimeConfig.adminRateLimit') }}</NText><NSwitch v-model:value="runtimeForm.admin_rate_limit_enabled" /></NSpace>
            <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.rate') }}</NText><NInputNumber v-model:value="runtimeForm.admin_rate_limit_rate" :min="1" :max="10000" /></NSpace>
            <NSpace align="center" justify="space-between"><NText>{{ t('adminServer.rateLimit.burst') }}</NText><NInputNumber v-model:value="runtimeForm.admin_rate_limit_burst" :min="1" :max="20000" /></NSpace>
          </NSpace>
        </NGi>
        <NGi>
          <NSpace vertical>
            <NText depth="3">{{ t('adminServer.runtimeConfig.hint') }}</NText>
            <NSpace>
              <NButton type="primary" :loading="savingRuntime" @click="saveRuntimeConfig">{{ t('adminServer.runtimeConfig.save') }}</NButton>
              <NButton :loading="restartLoading" @click="handleRestartBackend">{{ t('adminServer.runtimeConfig.restartBackend') }}</NButton>
            </NSpace>
          </NSpace>
        </NGi>
      </NGrid>
    </NCard>
  </NSpace>
</template>
