<script setup lang="ts">
/**
 * 服务器管理 · 运维与限流
 * 只负责全局/管理端限流配置与运行时快照，不掺杂日志、任务、重启等其它能力。
 */
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import type { DynamicRateLimitSnapshot } from '@/service/api/admin/server'
import { useServerManagement } from '../composables/useServerManagement'

const { t } = useI18n()
const {
  operations,
  operationsLoading,
  rateLimitForm,
  savingRateLimit,
  loadOperations,
  saveRateLimitConfig,
} = useServerManagement()

const rateLimits = computed(() => operations.value?.rate_limits || [])

const rateLimitColumns = computed<DataTableColumns<DynamicRateLimitSnapshot>>(() => [
  { title: t('adminServer.rateLimit.name'), key: 'name', minWidth: 140 },
  {
    title: t('adminServer.rateLimit.enabled'),
    key: 'enabled',
    width: 90,
    render(row) {
      return h(
        NTag,
        { type: row.enabled ? 'success' : 'default', size: 'small', bordered: false },
        () => (row.enabled ? t('common.enable') : t('common.disable')),
      )
    },
  },
  { title: t('adminServer.rateLimit.rate'), key: 'rate', width: 100 },
  { title: t('adminServer.rateLimit.burst'), key: 'burst', width: 100 },
  { title: t('adminServer.rateLimit.allowed'), key: 'allowed_count', width: 100 },
  { title: t('adminServer.rateLimit.blocked'), key: 'blocked_count', width: 100 },
  { title: t('adminServer.rateLimit.visitors'), key: 'active_visitors', width: 120 },
  {
    title: t('adminServer.rateLimit.lastReload'),
    key: 'last_config_reload',
    minWidth: 180,
    render: row => row.last_config_reload || '-',
  },
])

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
    <NCard :title="t('adminServer.rateLimit.configTitle')">
      <NAlert type="info" :bordered="false" style="margin-bottom: 16px;">
        {{ t('adminServer.rateLimit.hint') }}
      </NAlert>

      <NGrid cols="1 s:2" :x-gap="24" :y-gap="18" responsive="screen">
        <NGi>
          <NCard size="small" :title="t('adminServer.rateLimit.globalRateLimit')" embedded>
            <NSpace vertical :size="14">
              <NSpace align="center" justify="space-between">
                <NText>{{ t('adminServer.rateLimit.enabled') }}</NText>
                <NSwitch v-model:value="rateLimitForm.api_rate_limit_enabled" />
              </NSpace>
              <NSpace align="center" justify="space-between">
                <NText>{{ t('adminServer.rateLimit.rateWithUnit') }}</NText>
                <NInputNumber
                  v-model:value="rateLimitForm.api_rate_limit_rate"
                  :min="1"
                  :max="10000"
                  :disabled="!rateLimitForm.api_rate_limit_enabled"
                  style="width: 140px;"
                />
              </NSpace>
              <NSpace align="center" justify="space-between">
                <NText>{{ t('adminServer.rateLimit.burstWithUnit') }}</NText>
                <NInputNumber
                  v-model:value="rateLimitForm.api_rate_limit_burst"
                  :min="1"
                  :max="20000"
                  :disabled="!rateLimitForm.api_rate_limit_enabled"
                  style="width: 140px;"
                />
              </NSpace>
            </NSpace>
          </NCard>
        </NGi>

        <NGi>
          <NCard size="small" :title="t('adminServer.rateLimit.adminRateLimit')" embedded>
            <NSpace vertical :size="14">
              <NSpace align="center" justify="space-between">
                <NText>{{ t('adminServer.rateLimit.enabled') }}</NText>
                <NSwitch v-model:value="rateLimitForm.admin_rate_limit_enabled" />
              </NSpace>
              <NSpace align="center" justify="space-between">
                <NText>{{ t('adminServer.rateLimit.rateWithUnit') }}</NText>
                <NInputNumber
                  v-model:value="rateLimitForm.admin_rate_limit_rate"
                  :min="1"
                  :max="10000"
                  :disabled="!rateLimitForm.admin_rate_limit_enabled"
                  style="width: 140px;"
                />
              </NSpace>
              <NSpace align="center" justify="space-between">
                <NText>{{ t('adminServer.rateLimit.burstWithUnit') }}</NText>
                <NInputNumber
                  v-model:value="rateLimitForm.admin_rate_limit_burst"
                  :min="1"
                  :max="20000"
                  :disabled="!rateLimitForm.admin_rate_limit_enabled"
                  style="width: 140px;"
                />
              </NSpace>
            </NSpace>
          </NCard>
        </NGi>
      </NGrid>

      <NSpace style="margin-top: 16px;">
        <NButton type="primary" :loading="savingRateLimit" @click="saveRateLimitConfig">
          {{ t('adminServer.rateLimit.save') }}
        </NButton>
      </NSpace>
    </NCard>

    <NCard :title="t('adminServer.rateLimit.snapshotTitle')">
      <template #header-extra>
        <NSpace>
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
          <NButton :loading="operationsLoading" @click="loadOperations">
            {{ t('common.refresh') }}
          </NButton>
        </NSpace>
      </template>

      <NDataTable
        :columns="rateLimitVisibleColumns"
        :data="rateLimits"
        :loading="operationsLoading"
        :pagination="false"
        :scroll-x="rateLimitTableScrollX"
        :bordered="false"
      />
    </NCard>
  </NSpace>
</template>
