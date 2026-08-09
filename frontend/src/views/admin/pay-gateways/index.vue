<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NImage, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns, FormRules, SelectOption } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useRequestGuard, useTableColumnVisibility, withSubmitLock } from '@/hooks'
import {
  createPayGateway,
  deletePayGateway,
  fetchPayGateways,
  fetchPaymentChannelMetas,
  testPayGatewayConnection,
  updatePayGateway,
} from '@/service/api/admin/paygateway'
import type {
  ChannelConfigField,
  ChannelConfigFieldOption,
  ChannelMeta,
  ChannelVersionMeta,
  PayGateway,
  PayGatewayCreateRequest,
} from '@/service/api/admin/paygateway'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()
const listFetchGuard = useRequestGuard()

const loading = ref(false)
const list = ref<PayGateway[]>([])
const keyword = ref('')
const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 弹窗
const showModal = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref()
const advancedMode = ref(false)
const channelMetas = ref<ChannelMeta[]>([])
const channelMeta = ref<ChannelMeta | null>(null)
const extConfigMap = reactive<Record<string, string>>({})
const activeVersion = ref<ChannelVersionMeta | null>(null)

function defaultForm(): PayGatewayCreateRequest {
  return {
    name: '',
    type: 'epay',
    pay_type: 'alipay',
    sign_type: 'MD5',
    version: '',
    device: 'pc',
    currency: 'CNY',
    description: '',
    status: 1,
    api_url: '',
    pid: '',
    key: '',
    ext_config: '',
    logo_url: '',
    sort_order: 0,
    min_amount: 1,
    max_amount: 10000,
    fee_rate: 0,
    fee_mode: 'add',
    min_level: 0,
    notify_url: '',
    expire_minutes: 30,
    active_query_enabled: 1,
    query_interval_seconds: 120,
    query_batch_size: 50,
  }
}

const form = reactive<PayGatewayCreateRequest>(defaultForm())

const formRules: FormRules = {
  name: [{ required: true, message: t('adminPayGateways.enterName'), trigger: 'blur' }],
  type: [{ required: true, message: t('adminPayGateways.selectType'), trigger: 'change' }],
  pay_type: [{ required: true, message: t('adminPayGateways.selectPayType'), trigger: 'change' }],
  api_url: [{ required: true, message: t('adminPayGateways.enterApiUrl'), trigger: 'blur' }],
  version: [{ required: true, message: t('adminPayGateways.selectVersion'), trigger: 'change' }],
}

const feeModeOptions = [
  { label: t('adminPayGateways.feeModeAdd'), value: 'add' },
  { label: t('adminPayGateways.feeModeInclude'), value: 'include' },
]

const defaultPayTypeOptions = [
  { label: t('recharge.alipay'), value: 'alipay' },
  { label: t('recharge.wechatPay'), value: 'wxpay' },
  { label: t('recharge.qqWallet'), value: 'qqpay' },
  { label: t('recharge.bankCard'), value: 'bank' },
  { label: t('recharge.jdPay'), value: 'jdpay' },
]

const payTypeMap: Record<string, string> = {
  alipay: t('recharge.alipay'),
  wxpay: t('recharge.wechatPay'),
  qqpay: t('recharge.qqWallet'),
  bank: t('recharge.bankCard'),
  jdpay: t('recharge.jdPay'),
}

const typeOptions = ref<SelectOption[]>([
  { label: t('adminPayGateways.epay'), value: 'epay' },
])

const versionOptions = ref<SelectOption[]>([])
const signTypeOptions = ref<SelectOption[]>([])
const payTypeOptions = ref<SelectOption[]>([])
const deviceOptions = ref<SelectOption[]>([])
const dynamicConfigFields = ref<ChannelConfigField[]>([])

const columns: DataTableColumns<PayGateway> = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: t('adminPayGateways.logo'),
    key: 'logo_url',
    width: 60,
    render: (row) => {
      if (row.logo_url) {
        return h(NImage, { src: row.logo_url, width: 32, height: 32, objectFit: 'contain', fallbackSrc: '', style: { borderRadius: '4px' } })
      }
      return h('span', { style: { color: '#999', fontSize: '12px' } }, t('recharge.none'))
    },
  },
  {
    title: t('adminPayGateways.gatewayName'),
    key: 'name',
    width: 140,
    ellipsis: { tooltip: true },
  },
  {
    title: t('recharge.paymentMethod'),
    key: 'pay_type',
    width: 90,
    render: row => payTypeMap[row.pay_type] || row.pay_type,
  },
  {
    title: t('adminPayGateways.status'),
    key: 'status',
    width: 70,
    render: row => h(NTag, { type: row.status === 1 ? 'success' : 'default', size: 'small', bordered: false }, () => row.status === 1 ? t('adminUsers.enabled') : t('adminUsers.disabled')),
  },
  {
    title: t('adminPayGateways.amountRange'),
    key: 'amount_range',
    width: 140,
    render: row => `¥${row.min_amount} - ¥${row.max_amount}`,
  },
  {
    title: t('adminPayGateways.fee'),
    key: 'fee_rate',
    width: 80,
    render: row => row.fee_rate > 0 ? `${row.fee_rate}%` : t('recharge.none'),
  },
  {
    title: t('adminPayGateways.minLevel'),
    key: 'min_level',
    width: 80,
    render: row => row.min_level > 0 ? `Lv.${row.min_level}` : t('adminPayGateways.unlimited'),
  },
  {
    title: t('adminPayGateways.sortOrder'),
    key: 'sort_order',
    width: 60,
  },
  {
    title: t('moneyScore.actions'),
    key: 'actions',
    width: 190,
    render: (row) => {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, type: 'info', onClick: () => handleTestConnection(row) }, () => t('adminPayGateways.testConnection')),
        h(NButton, { size: 'small', quaternary: true, type: 'primary', onClick: () => handleEdit(row) }, () => t('adminUsers.edit')),
        h(NButton, { size: 'small', quaternary: true, type: 'error', onClick: () => handleDelete(row) }, () => t('adminPayGateways.delete')),
      ])
    },
  },
]

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'logo_url', label: t('adminPayGateways.logo') },
  { key: 'name', label: t('adminPayGateways.gatewayName') },
  { key: 'pay_type', label: t('recharge.paymentMethod') },
  { key: 'status', label: t('adminPayGateways.status') },
  { key: 'amount_range', label: t('adminPayGateways.amountRange') },
  { key: 'fee_rate', label: t('adminPayGateways.fee') },
  { key: 'min_level', label: t('adminPayGateways.minLevel') },
  { key: 'sort_order', label: t('adminPayGateways.sortOrder') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<PayGateway>({
  storageKey: 'admin-pay-gateways-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 980,
})

async function loadList() {
  const token = listFetchGuard.begin()
  loading.value = true
  try {
    const res = await fetchPayGateways({ page: pagination.page, page_size: pagination.pageSize, keyword: keyword.value })
    if (!listFetchGuard.isLatest(token))
      return
    if (res.isSuccess) {
      list.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    }
  }
  catch {
    // 网络异常：alova onError 已提示
  }
  finally {
    if (listFetchGuard.isLatest(token))
      loading.value = false
  }
}

async function loadChannelMetas() {
  try {
    const res = await fetchPaymentChannelMetas()
    if (res.isSuccess && res.data) {
      channelMetas.value = res.data
      typeOptions.value = res.data.map(m => ({ label: m.name || m.type, value: m.type }))
    }
  }
  catch {
    // 后端不支持时兜底
    typeOptions.value = [
      { label: t('adminPayGateways.epay'), value: 'epay' },
      { label: t('adminPayGateways.wechat'), value: 'wechat' },
      { label: t('adminPayGateways.alipay'), value: 'alipay' },
    ]
  }
}

const baseApiURL = `${window.location.protocol}//${window.location.host}`

function pickChannelMeta(type: string) {
  channelMeta.value = channelMetas.value.find(m => m.type === type) || null
  if (channelMeta.value) {
    payTypeOptions.value = channelMeta.value.pay_types?.length > 0
      ? channelMeta.value.pay_types.map(p => ({ label: p.name || p.value, value: p.value }))
      : defaultPayTypeOptions
    versionOptions.value = channelMeta.value.versions.map(v => ({ label: v.name || v.version, value: v.version }))
    deviceOptions.value = channelMeta.value.devices?.length > 0
      ? channelMeta.value.devices.map(d => ({ label: d.name || d.value, value: d.value }))
      : [
          { label: t('adminPayGateways.pc'), value: 'pc' },
          { label: t('adminPayGateways.mobile'), value: 'mobile' },
        ]

    // 若 notify_url 为空，自动回填默认回调地址
    if (!form.notify_url && channelMeta.value.default_notify_path) {
      form.notify_url = `${baseApiURL}${channelMeta.value.default_notify_path}`
    }
  }
  else {
    payTypeOptions.value = defaultPayTypeOptions
    versionOptions.value = []
    signTypeOptions.value = []
    deviceOptions.value = [
      { label: t('adminPayGateways.pc'), value: 'pc' },
      { label: t('adminPayGateways.mobile'), value: 'mobile' },
    ]
    dynamicConfigFields.value = []
  }
}

function pickVersion(version: string) {
  activeVersion.value = channelMeta.value?.versions.find(v => v.version === version) || null
  if (activeVersion.value) {
    signTypeOptions.value = activeVersion.value.signTypes.map(s => ({ label: s.name || s.value, value: s.value }))
    dynamicConfigFields.value = activeVersion.value.configFields?.map((field) => {
      if (field.type === 'select' && field.options?.length) {
        return {
          ...field,
          options: field.options.map((opt: ChannelConfigFieldOption) => ({ label: opt.label || opt.value, value: opt.value })),
        }
      }
      return field
    }) || []
    // 自动将签名算法选为第一个
    if (!form.sign_type && signTypeOptions.value.length > 0) {
      form.sign_type = signTypeOptions.value[0].value as string
    }
  }
  else {
    signTypeOptions.value = []
    dynamicConfigFields.value = []
  }
}

function syncExtConfigFromMap() {
  if (Object.keys(extConfigMap).length === 0) {
    form.ext_config = ''
    return
  }
  try {
    form.ext_config = JSON.stringify(extConfigMap, null, 2)
  }
  catch {
    // ignore
  }
}

function syncExtConfigToMap() {
  Object.keys(extConfigMap).forEach(k => delete extConfigMap[k])
  if (!form.ext_config)
    return
  try {
    const parsed = JSON.parse(form.ext_config)
    if (parsed && typeof parsed === 'object') {
      Object.keys(parsed).forEach((k) => {
        extConfigMap[k] = parsed[k]
      })
    }
  }
  catch {
    // ignore
  }
}

watch(() => form.type, (type) => {
  pickChannelMeta(type)
  if (versionOptions.value.length > 0 && !form.version) {
    form.version = versionOptions.value[0].value as string
  }
})

watch(() => form.version, (version) => {
  if (version)
    pickVersion(version)
})

watch(() => form.sign_type, () => {
  // 不同签名算法可能有不同配置字段；若后端未来返回，这里可重新渲染
})

watch(extConfigMap, () => {
  if (!advancedMode.value)
    syncExtConfigFromMap()
}, { deep: true })

watch(advancedMode, (v) => {
  if (v) {
    // 切到高级模式：把 map 同步到 JSON 编辑器
    syncExtConfigFromMap()
  }
  else {
    // 切回普通模式：把 JSON 同步到 map
    syncExtConfigToMap()
  }
})

function handleCreate() {
  editingId.value = null
  Object.assign(form, defaultForm())
  Object.keys(extConfigMap).forEach(k => delete extConfigMap[k])
  advancedMode.value = false
  channelMeta.value = null
  activeVersion.value = null
  pickChannelMeta(form.type)
  showModal.value = true
}

function handleEdit(row: PayGateway) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    type: row.type,
    pay_type: row.pay_type,
    sign_type: row.sign_type || 'MD5',
    version: row.version || '',
    device: row.device || 'pc',
    currency: row.currency || 'CNY',
    description: row.description,
    status: row.status,
    api_url: row.api_url,
    pid: row.pid,
    key: row.key,
    ext_config: row.ext_config || '',
    logo_url: row.logo_url,
    sort_order: row.sort_order,
    min_amount: row.min_amount,
    max_amount: row.max_amount,
    fee_rate: row.fee_rate,
    fee_mode: row.fee_mode || 'add',
    min_level: row.min_level,
    notify_url: row.notify_url,
    expire_minutes: row.expire_minutes || 30,
    active_query_enabled: row.active_query_enabled ?? 1,
    query_interval_seconds: row.query_interval_seconds || 120,
    query_batch_size: row.query_batch_size || 50,
  })
  advancedMode.value = false
  syncExtConfigToMap()
  pickChannelMeta(form.type)
  if (form.version)
    pickVersion(form.version)
  showModal.value = true
}

async function handleSubmit() {
  if (submitting.value)
    return
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  if (!advancedMode.value)
    syncExtConfigFromMap()

  // 兼容旧 key：若 ext_config 为空且 key 有值，把 key 写入 ext_config
  const payload: PayGatewayCreateRequest = { ...form }
  if (!payload.ext_config && payload.key) {
    try {
      payload.ext_config = JSON.stringify({ key: payload.key }, null, 2)
    }
    catch {
      // ignore
    }
  }

  await withSubmitLock(submitting, async () => {
    try {
      if (editingId.value) {
        const res = await updatePayGateway(editingId.value, payload)
        if (res.isSuccess) {
          message.success(t('adminUsers.updateSuccess'))
          showModal.value = false
          loadList()
        }
      }
      else {
        const res = await createPayGateway(payload)
        if (res.isSuccess) {
          message.success(t('adminUsers.createSuccess'))
          showModal.value = false
          loadList()
        }
      }
    }
    catch {
      // 网络异常：alova onError 已提示
    }
  })
}

async function handleTestConnection(row: PayGateway) {
  try {
    const res = await testPayGatewayConnection(row.id)
    if (res.isSuccess && res.data) {
      if (res.data.success)
        message.success(res.data.message || t('adminPayGateways.testConnectionSuccess'))
      else
        message.warning(res.data.message || t('adminPayGateways.testConnectionFailed'))
    }
  }
  catch {
    // alova onError 已提示
  }
}

function handleDelete(row: PayGateway) {
  if (submitting.value)
    return
  dialog.warning({
    title: t('adminPayGateways.confirmDeleteTitle'),
    content: t('adminPayGateways.confirmDeleteContent', { name: row.name }),
    positiveText: t('adminPayGateways.delete'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(submitting, async () => {
      try {
        const res = await deletePayGateway(row.id)
        if (res.isSuccess) {
          message.success(t('adminPayGateways.deleteSuccess'))
          loadList()
        }
      }
      catch {
        // 网络异常：alova onError 已提示
      }
    }),
  })
}

onMounted(() => {
  loadList()
  loadChannelMetas()
})
</script>

<template>
  <NSpace vertical :size="16">
    <n-card :title="t('route.admin-pay-gateways')">
      <template #header-extra>
        <NSpace>
          <TableColumnSelector
            v-model="selectedColumnKeys"
            :options="columnOptions"
            :visible-count="visibleColumnCount"
            :total-count="totalColumnCount"
            :button-label="t('common.showFields')"
            :title="t('common.visibleFields')"
            :hint="t('common.columnVisibilityHint')"
            :reset-label="t('common.restoreDefaultFields')"
            @reset="resetSelectedColumns"
          />
          <NButton type="primary" @click="handleCreate">
            <template #icon>
              <n-icon><icon-park-outline-add-one /></n-icon>
            </template>
            {{ t('adminPayGateways.addGatewayShort') }}
          </NButton>
        </NSpace>
      </template>

      <NSpace vertical>
        <NSpace>
          <n-input v-model:value="keyword" :placeholder="t('adminPayGateways.searchPlaceholder')" clearable style="width: 220px" @keyup.enter="loadList" />
          <NButton type="primary" @click="loadList">
            {{ t('moneyScore.search') }}
          </NButton>
        </NSpace>

        <n-data-table
          :columns="visibleColumns"
          :data="list"
          :loading="loading"
          :pagination="pagination"
          :scroll-x="tableScrollX"
          striped
          size="small"
          @update:page="(p: number) => { pagination.page = p; loadList() }"
          @update:page-size="(s: number) => { pagination.pageSize = s; pagination.page = 1; loadList() }"
        />
      </NSpace>
    </n-card>

    <!-- 新增/编辑弹窗 -->
    <n-modal v-model:show="showModal" preset="card" :title="editingId ? t('adminPayGateways.editGateway') : t('adminPayGateways.addGateway')" style="width: 760px; max-width: 92vw" :mask-closable="false">
      <n-form ref="formRef" :model="form" :rules="formRules" label-placement="left" label-width="120">
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item :label="t('adminPayGateways.gatewayName')" path="name">
              <n-input v-model:value="form.name" :placeholder="t('adminPayGateways.gatewayNamePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.gatewayType')" path="type">
              <n-select v-model:value="form.type" :options="typeOptions" :placeholder="t('adminPayGateways.gatewayTypePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('recharge.paymentMethod')" path="pay_type">
              <n-select v-model:value="form.pay_type" :options="payTypeOptions.length ? payTypeOptions : defaultPayTypeOptions" :placeholder="t('adminPayGateways.payTypePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.version')" path="version">
              <n-select v-model:value="form.version" :options="versionOptions" :placeholder="t('adminPayGateways.versionPlaceholder')" :disabled="!channelMeta" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.signType')" path="sign_type">
              <n-select v-model:value="form.sign_type" :options="signTypeOptions" :placeholder="t('adminPayGateways.signTypePlaceholder')" :disabled="!activeVersion" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.device')" path="device">
              <n-select v-model:value="form.device" :options="deviceOptions" :placeholder="t('adminPayGateways.devicePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.currency')" path="currency">
              <n-input v-model:value="form.currency" :placeholder="t('adminPayGateways.currencyPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.status')" path="status">
              <n-switch v-model:value="form.status" :checked-value="1" :unchecked-value="0">
                <template #checked>
                  {{ t('adminUsers.enabled') }}
                </template>
                <template #unchecked>
                  {{ t('adminUsers.disabled') }}
                </template>
              </n-switch>
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.description')" path="description">
              <n-input v-model:value="form.description" type="textarea" :placeholder="t('adminPayGateways.descriptionPlaceholder')" :rows="2" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.apiUrl')" path="api_url">
              <n-input v-model:value="form.api_url" :placeholder="t('adminPayGateways.apiUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.merchantId')" path="pid">
              <n-input v-model:value="form.pid" :placeholder="t('adminPayGateways.merchantIdPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.merchantKey')" path="key">
              <n-input v-model:value="form.key" type="password" show-password-on="click" :placeholder="t('adminPayGateways.merchantKeyPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.logoUrl')" path="logo_url">
              <n-input v-model:value="form.logo_url" :placeholder="t('adminPayGateways.logoUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.minAmount')" path="min_amount">
              <n-input-number v-model:value="form.min_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>
                  ¥
                </template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.maxAmount')" path="max_amount">
              <n-input-number v-model:value="form.max_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>
                  ¥
                </template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.feeRate')" path="fee_rate">
              <n-input-number v-model:value="form.fee_rate" :min="0" :max="100" style="width: 100%">
                <template #suffix>
                  %
                </template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.feeMode')" path="fee_mode">
              <n-select v-model:value="form.fee_mode" :options="feeModeOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.minLevel')" path="min_level">
              <n-input-number v-model:value="form.min_level" :min="0" style="width: 100%" :placeholder="t('adminPayGateways.minLevelPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.sortOrder')" path="sort_order">
              <n-input-number v-model:value="form.sort_order" :min="0" style="width: 100%" :placeholder="t('adminPayGateways.sortOrderPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.expireMinutes')" path="expire_minutes">
              <n-input-number v-model:value="form.expire_minutes" :min="1" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.activeQueryEnabled')" path="active_query_enabled">
              <n-switch v-model:value="form.active_query_enabled" :checked-value="1" :unchecked-value="0" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.queryIntervalSeconds')" path="query_interval_seconds">
              <n-input-number v-model:value="form.query_interval_seconds" :min="10" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.queryBatchSize')" path="query_batch_size">
              <n-input-number v-model:value="form.query_batch_size" :min="1" :max="200" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.notifyUrl')" path="notify_url">
              <n-input v-model:value="form.notify_url" :placeholder="t('adminPayGateways.notifyUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <!-- 动态扩展配置字段 -->
        <n-divider>{{ t('adminPayGateways.dynamicConfig') }}</n-divider>
        <n-grid v-if="dynamicConfigFields.length > 0" :cols="1" :x-gap="16">
          <n-gi v-for="field in dynamicConfigFields" :key="field.name">
            <n-form-item :label="field.label || field.name" :path="`extConfigMap.${field.name}`">
              <n-input
                v-if="field.type === 'input'"
                v-model:value="extConfigMap[field.name]"
                :placeholder="field.placeholder"
                :type="field.secret ? 'password' : 'text'"
                show-password-on="click"
              />
              <n-input
                v-else-if="field.type === 'textarea'"
                v-model:value="extConfigMap[field.name]"
                type="textarea"
                :placeholder="field.placeholder"
                :rows="3"
              />
              <n-select
                v-else-if="field.type === 'select'"
                v-model:value="extConfigMap[field.name]"
                :options="field.options || []"
                :placeholder="field.placeholder"
              />
            </n-form-item>
          </n-gi>
        </n-grid>
        <n-empty v-else :description="t('adminPayGateways.channelTypeNotSelected')" size="small" />

        <!-- ext_config 高级模式 -->
        <NSpace align="center" style="margin: 16px 0 8px">
          <n-switch v-model:value="advancedMode" />
          <span>{{ t('adminPayGateways.advancedMode') }}</span>
        </NSpace>
        <n-form-item v-if="advancedMode" :label="t('adminPayGateways.extConfig')" path="ext_config">
          <n-input v-model:value="form.ext_config" type="textarea" :placeholder="t('adminPayGateways.extConfigPlaceholder')" :rows="8" />
        </n-form-item>
      </n-form>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ t('common.confirm') }}
          </NButton>
        </NSpace>
      </template>
    </n-modal>
  </NSpace>
</template>
