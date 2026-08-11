<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns, FormRules } from 'naive-ui'
import { withSubmitLock } from '@/hooks'
import { adminApi } from '@/service/api/admin'
import type { BatchRefreshPreviewItem, ExchangeRate } from '@/service/api/admin/currency'
import CurrencyPair from '@/components/common/CurrencyPair.vue'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const saving = ref(false)
const refreshing = ref(false)
const refreshingId = ref<number | null>(null)
const exchangeRates = ref<ExchangeRate[]>([])
const checkedRowKeys = ref<number[]>([])

const configForm = reactive({
  base_currency: 'CNY',
  currency_dynamic_source: '',
  currency_dynamic_source_url: '',
})

const showEditModal = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref()

const showBatchRefreshModal = ref(false)
const batchRefreshPreviewItems = ref<BatchRefreshPreviewItem[]>([])
const batchRefreshing = ref(false)
const isBatchAll = ref(false)

const rateForm = reactive({
  from_currency: '',
  to_currency: '',
  rate: 0,
  fixed_amount: 0,
  rate_type: 'fixed',
  source: '',
})

const rateTypeOptions = [
  { label: t('adminCurrency.rateTypeFixed'), value: 'fixed' },
  { label: t('adminCurrency.rateTypeDynamic'), value: 'dynamic' },
]

const baseCurrencyOptions = [
  { label: 'CNY', value: 'CNY' },
  { label: 'USD', value: 'USD' },
]

const dynamicSourceOptions = [
  { label: t('adminCurrency.dynamicSourceExchangerateApi'), value: 'exchangerate-api' },
  { label: t('adminCurrency.dynamicSourceCustom'), value: 'custom' },
]

const formRules: FormRules = {
  from_currency: [{ required: true, message: () => t('adminCurrency.fromCurrencyRequired'), trigger: 'blur' }],
  to_currency: [{ required: true, message: () => t('adminCurrency.toCurrencyRequired'), trigger: 'blur' }],
  rate: [{ required: true, type: 'number', min: 0.00000001, message: () => t('adminCurrency.rateRequired'), trigger: 'blur' }],
}

function defaultRateForm(): ExchangeRate {
  return {
    id: 0,
    from_currency: '',
    to_currency: '',
    rate: 0,
    fixed_amount: 0,
    rate_type: 'fixed',
    source: '',
    create_time: 0,
    update_time: 0,
  }
}

async function loadData() {
  loading.value = true
  try {
    const [configRes, ratesRes] = await Promise.all([
      adminApi.currency.getCurrencyConfig(),
      adminApi.currency.fetchExchangeRates(),
    ])
    if (configRes.isSuccess) {
      const data = configRes.data
      configForm.base_currency = data?.base_currency || 'CNY'
      configForm.currency_dynamic_source = data?.currency_dynamic_source || ''
      configForm.currency_dynamic_source_url = data?.currency_dynamic_source_url || ''
    }
    if (ratesRes.isSuccess) {
      exchangeRates.value = ratesRes.data?.list || []
      checkedRowKeys.value = checkedRowKeys.value.filter(id => exchangeRates.value.some(r => r.id === id))
    }
  }
  catch {
    message.error(t('adminCurrency.loadFailed'))
  }
  finally {
    loading.value = false
  }
}

async function handleSaveConfig() {
  await withSubmitLock(saving, async () => {
    const res = await adminApi.currency.updateCurrencyConfig({
      base_currency: configForm.base_currency.trim().toUpperCase(),
      currency_dynamic_source: configForm.currency_dynamic_source,
      currency_dynamic_source_url: configForm.currency_dynamic_source_url.trim(),
    })
    if (res.isSuccess) {
      message.success(t('adminCurrency.configSaved'))
    }
    else {
      message.error(res.message || t('adminCurrency.configSaveFailed'))
    }
  })
}

function openAddModal() {
  editingId.value = null
  Object.assign(rateForm, defaultRateForm())
  showEditModal.value = true
}

function openEditModal(row: ExchangeRate) {
  editingId.value = row.id
  rateForm.from_currency = row.from_currency
  rateForm.to_currency = row.to_currency
  rateForm.rate = row.rate
  rateForm.fixed_amount = row.fixed_amount || 0
  rateForm.rate_type = row.rate_type || 'fixed'
  rateForm.source = row.source || ''
  showEditModal.value = true
}

async function handleSaveRate() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  const data = {
    from_currency: rateForm.from_currency.trim().toUpperCase(),
    to_currency: rateForm.to_currency.trim().toUpperCase(),
    rate: rateForm.rate,
    fixed_amount: rateForm.fixed_amount,
    rate_type: rateForm.rate_type,
    source: rateForm.source.trim(),
  }

  await withSubmitLock(submitting, async () => {
    const res = editingId.value
      ? await adminApi.currency.updateExchangeRate(editingId.value, data)
      : await adminApi.currency.createExchangeRate(data)
    if (res.isSuccess) {
      message.success(res.message || t('adminCurrency.rateSaved'))
      showEditModal.value = false
      await loadData()
    }
    else {
      message.error(res.message || t('adminCurrency.rateSaveFailed'))
    }
  })
}

function handleDeleteRate(row: ExchangeRate) {
  dialog.warning({
    title: t('adminCurrency.confirmDeleteTitle'),
    content: t('adminCurrency.confirmDeleteContent', { from: row.from_currency, to: row.to_currency }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res = await adminApi.currency.deleteExchangeRate(row.id)
      if (res.isSuccess) {
        message.success(res.message || t('adminCurrency.deleteSuccess'))
        await loadData()
      }
      else {
        message.error(res.message || t('adminCurrency.deleteFailed'))
      }
    },
  })
}

async function handleRefreshRates() {
  await withSubmitLock(refreshing, async () => {
    const res = await adminApi.currency.refreshExchangeRates()
    if (res.isSuccess) {
      message.success(res.message || t('adminCurrency.refreshSuccess'))
      await loadData()
    }
    else {
      message.error(res.message || t('adminCurrency.refreshFailed'))
    }
  })
}

async function handleRefreshRate(row: ExchangeRate) {
  refreshingId.value = row.id
  try {
    const res = await adminApi.currency.refreshExchangeRate(row.id)
    if (res.isSuccess) {
      message.success(res.message || t('adminCurrency.refreshRateSuccess'))
      await loadData()
    }
    else {
      message.error(res.message || t('adminCurrency.refreshRateFailed'))
    }
  }
  finally {
    refreshingId.value = null
  }
}

function getSelectedIds(): number[] {
  if (isBatchAll.value) {
    return exchangeRates.value.map(r => r.id)
  }
  return checkedRowKeys.value
}

async function openBatchRefreshPreview(all: boolean) {
  isBatchAll.value = all
  const ids = getSelectedIds()
  if (ids.length === 0) {
    message.warning(t('adminCurrency.pleaseSelectRates'))
    return
  }

  batchRefreshing.value = true
  try {
    const res = await adminApi.currency.batchRefreshExchangeRatesPreview(ids)
    if (res.isSuccess) {
      batchRefreshPreviewItems.value = res.data?.items || []
      showBatchRefreshModal.value = true
    }
    else {
      message.error(res.message || t('adminCurrency.batchRefreshPreviewFailed'))
    }
  }
  finally {
    batchRefreshing.value = false
  }
}

async function handleConfirmBatchRefresh() {
  const ids = batchRefreshPreviewItems.value
    .filter(item => !item.error && item.new_rate > 0)
    .map(item => item.id)
  if (ids.length === 0) {
    message.warning(t('adminCurrency.noRatesToRefresh'))
    return
  }

  batchRefreshing.value = true
  try {
    const res = await adminApi.currency.batchRefreshExchangeRates(ids)
    if (res.isSuccess) {
      message.success(t('adminCurrency.batchRefreshSuccess'))
      showBatchRefreshModal.value = false
      checkedRowKeys.value = []
      isBatchAll.value = false
      await loadData()
    }
    else {
      message.error(res.message || t('adminCurrency.batchRefreshFailed'))
    }
  }
  finally {
    batchRefreshing.value = false
  }
}

function isPreviewChanged(): boolean {
  return batchRefreshPreviewItems.value.some(item => !item.error && Math.abs(item.new_rate - item.old_rate) > 0)
}

const columns: DataTableColumns<ExchangeRate> = [
  {
    type: 'selection',
  },
  {
    title: t('adminCurrency.currencyPair'),
    key: 'pair',
    render: (row: ExchangeRate) => h(CurrencyPair, { from: row.from_currency, to: row.to_currency }),
  },
  { title: t('adminCurrency.rate'), key: 'rate' },
  { title: t('adminCurrency.fixedAmount'), key: 'fixed_amount' },
  {
    title: t('adminCurrency.rateType'),
    key: 'rate_type',
    render: (row: ExchangeRate) => {
      const map: Record<string, string> = {
        fixed: t('adminCurrency.rateTypeFixed'),
        dynamic: t('adminCurrency.rateTypeDynamic'),
      }
      return map[row.rate_type] || row.rate_type
    },
  },
  { title: t('adminCurrency.source'), key: 'source', ellipsis: { tooltip: true } },
  {
    title: t('adminCurrency.action'),
    key: 'action',
    width: 190,
    render(row) {
      return h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { size: 'small', loading: refreshingId.value === row.id, onClick: () => handleRefreshRate(row) }, { default: () => t('adminCurrency.refreshRate') }),
          h(NButton, { size: 'small', onClick: () => openEditModal(row) }, { default: () => t('common.edit') }),
          h(NButton, { size: 'small', type: 'error', onClick: () => handleDeleteRate(row) }, { default: () => t('common.delete') }),
        ],
      })
    },
  },
]

const previewColumns: DataTableColumns<BatchRefreshPreviewItem> = [
  {
    title: t('adminCurrency.currencyPair'),
    key: 'pair',
    render: (row: BatchRefreshPreviewItem) => h(CurrencyPair, { from: row.from_currency, to: row.to_currency }),
  },
  {
    title: t('adminCurrency.oldRate'),
    key: 'old_rate',
    render: (row: BatchRefreshPreviewItem) => row.old_rate.toFixed(8),
  },
  {
    title: t('adminCurrency.newRate'),
    key: 'new_rate',
    render: (row: BatchRefreshPreviewItem) => h('span', { style: { color: getRateColor(row) } }, row.new_rate.toFixed(8)),
  },
  {
    title: t('adminCurrency.source'),
    key: 'source',
    ellipsis: { tooltip: true },
  },
  {
    title: t('adminCurrency.error'),
    key: 'error',
    render: (row: BatchRefreshPreviewItem) => row.error || '-',
  },
]

function getRateColor(row: BatchRefreshPreviewItem): string {
  if (row.error)
    return 'var(--n-text-color-disabled)'
  if (row.new_rate > row.old_rate)
    return 'var(--n-success-color)'
  if (row.new_rate < row.old_rate)
    return 'var(--n-error-color)'
  return 'var(--n-text-color-base)'
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <n-spin :show="loading">
    <NSpace vertical :size="24">
      <!-- 全局货币配置 -->
      <n-card :title="t('adminCurrency.globalConfig')" :bordered="false">
        <n-form :model="configForm" label-placement="left" label-width="180px" style="max-width: 720px;">
          <n-form-item :label="t('adminCurrency.baseCurrency')">
            <n-select v-model:value="configForm.base_currency" :options="baseCurrencyOptions" :placeholder="t('adminCurrency.baseCurrencyPlaceholder')" style="width: 200px;" />
          </n-form-item>
          <n-form-item :label="t('adminCurrency.dynamicSource')">
            <n-select v-model:value="configForm.currency_dynamic_source" :options="dynamicSourceOptions" :placeholder="t('adminCurrency.dynamicSourcePlaceholder')" clearable style="width: 280px;" />
          </n-form-item>
          <n-form-item :label="t('adminCurrency.dynamicSourceUrl')">
            <n-input v-model:value="configForm.currency_dynamic_source_url" :placeholder="t('adminCurrency.dynamicSourceUrlPlaceholder')" />
          </n-form-item>
          <n-form-item>
            <NButton type="primary" :loading="saving" @click="handleSaveConfig">
              {{ t('adminCurrency.saveConfig') }}
            </NButton>
          </n-form-item>
        </n-form>
      </n-card>

      <!-- 汇率列表 -->
      <n-card :title="t('adminCurrency.exchangeRateList')" :bordered="false">
        <NSpace vertical :size="16">
          <NSpace>
            <NButton type="primary" @click="openAddModal">
              <template #icon>
                <n-icon><icon-park-outline-add-one /></n-icon>
              </template>
              {{ t('adminCurrency.addRate') }}
            </NButton>
            <NButton :loading="batchRefreshing" @click="openBatchRefreshPreview(false)">
              <template #icon>
                <n-icon><icon-park-outline-refresh /></n-icon>
              </template>
              {{ t('adminCurrency.syncSelected') }}
            </NButton>
            <NButton :loading="batchRefreshing" @click="openBatchRefreshPreview(true)">
              <template #icon>
                <n-icon><icon-park-outline-refresh /></n-icon>
              </template>
              {{ t('adminCurrency.syncAll') }}
            </NButton>
            <NButton :loading="refreshing" @click="handleRefreshRates">
              <template #icon>
                <n-icon><icon-park-outline-refresh /></n-icon>
              </template>
              {{ t('adminCurrency.refreshRates') }}
            </NButton>
          </NSpace>
          <n-data-table
            :columns="columns"
            :data="exchangeRates"
            :row-key="(row: ExchangeRate) => row.id"
            :checked-row-keys="checkedRowKeys"
            :loading="loading"
            size="small"
            @update:checked-row-keys="(keys: any[]) => checkedRowKeys = keys as number[]"
          />
        </NSpace>
      </n-card>
    </NSpace>
  </n-spin>

  <!-- 新增/编辑汇率 -->
  <n-modal v-model:show="showEditModal" preset="card" :title="editingId ? t('adminCurrency.editRate') : t('adminCurrency.addRate')" style="width: 560px; max-width: 92vw" :mask-closable="false">
    <n-form ref="formRef" :model="rateForm" :rules="formRules" label-placement="left" label-width="140px">
      <n-form-item :label="t('adminCurrency.fromCurrency')" path="from_currency">
        <n-input v-model:value="rateForm.from_currency" :placeholder="t('adminCurrency.fromCurrencyPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminCurrency.toCurrency')" path="to_currency">
        <n-input v-model:value="rateForm.to_currency" :placeholder="t('adminCurrency.toCurrencyPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminCurrency.rate')" path="rate">
        <n-input-number v-model:value="rateForm.rate" :min="0" :precision="8" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminCurrency.fixedAmount')">
        <n-input-number v-model:value="rateForm.fixed_amount" :min="0" :precision="8" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminCurrency.rateType')">
        <n-select v-model:value="rateForm.rate_type" :options="rateTypeOptions" />
      </n-form-item>
      <n-form-item :label="t('adminCurrency.source')">
        <n-input v-model:value="rateForm.source" :placeholder="t('adminCurrency.sourcePlaceholder')" />
      </n-form-item>
    </n-form>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="showEditModal = false">
          {{ t('common.cancel') }}
        </NButton>
        <NButton type="primary" :loading="submitting" @click="handleSaveRate">
          {{ t('common.save') }}
        </NButton>
      </NSpace>
    </template>
  </n-modal>

  <!-- 批量刷新预览确认 -->
  <n-modal v-model:show="showBatchRefreshModal" preset="card" :title="t('adminCurrency.batchRefreshPreviewTitle')" style="width: 720px; max-width: 92vw" :mask-closable="false">
    <NSpace vertical :size="16">
      <n-alert v-if="!isPreviewChanged()" type="warning" :bordered="false">
        {{ t('adminCurrency.batchRefreshNoChange') }}
      </n-alert>
      <n-alert v-else type="info" :bordered="false">
        {{ t('adminCurrency.batchRefreshTip') }}
      </n-alert>
      <n-data-table :columns="previewColumns" :data="batchRefreshPreviewItems" size="small" :scroll-x="600" />
    </NSpace>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="showBatchRefreshModal = false">
          {{ t('common.cancel') }}
        </NButton>
        <NButton type="primary" :loading="batchRefreshing" :disabled="!isPreviewChanged()" @click="handleConfirmBatchRefresh">
          {{ t('adminCurrency.batchRefreshConfirm') }}
        </NButton>
      </NSpace>
    </template>
  </n-modal>
</template>
