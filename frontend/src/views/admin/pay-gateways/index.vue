<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NImage, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns, FormRules, SelectOption } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useRequestGuard, useTableColumnVisibility, withSubmitLock } from '@/hooks'
import {
  createExchangeRate,
  createPayGateway,
  deleteExchangeRate,
  deletePayGateway,
  fetchExchangeRates,
  fetchPayGateways,
  fetchPaymentChannelMetas,
  getBaseCurrency,
  marshalExtConfig,
  parseExtConfig,
  refreshExchangeRate,
  refreshExchangeRates,
  setBaseCurrency,
  testPayGatewayConnection,
  updateExchangeRate,
  updatePayGateway,
} from '@/service/api/admin/paygateway'
import type {
  ChannelConfigField,
  ChannelMeta,
  ChannelVersionMeta,
  ExchangeRate,
  PayGateway,
  PayGatewayCreateRequest,
} from '@/service/api/admin/paygateway'
import CurrencyPair from '@/components/common/CurrencyPair.vue'
import { useBaseCurrency } from '@/composables/useBaseCurrency'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()
const { currencySymbol } = useBaseCurrency()
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
const extConfigMap = reactive<Record<string, any>>({})
const activeVersion = ref<ChannelVersionMeta | null>(null)

function defaultForm(): PayGatewayCreateRequest {
  return {
    name: '',
    type: 'epay',
    pay_type: 'alipay',
    version: '',
    device: 'pc',
    currency: 'CNY',
    description: '',
    status: 1,
    api_url: '',
    pid: '',
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
  }
}

const baseCurrency = ref('CNY')

const form = reactive<PayGatewayCreateRequest>(defaultForm())

const activeQueryEnabled = computed(() => {
  const v = extConfigMap.active_query_enabled
  return v === 1 || v === '1' || v === true
})

// 是否需要展示「目标币种/汇率/目标手续费」相关配置
const needExchange = computed(() => {
  const targetCurrency = String(extConfigMap.target_currency || '').trim().toUpperCase()
  return targetCurrency && targetCurrency !== String(form.currency || '').trim().toUpperCase()
})

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

const exchangeRateModeOptions = [
  { label: t('adminPayGateways.exchangeRateModeSystem'), value: 'system' },
  { label: t('adminPayGateways.exchangeRateModeFixed'), value: 'fixed' },
  { label: t('adminPayGateways.exchangeRateModeDynamic'), value: 'dynamic' },
]

// 汇率管理弹窗
const showCurrencyModal = ref(false)
const currencyLoading = ref(false)
const currencyRefreshingId = ref<number | null>(null)
const exchangeRates = ref<ExchangeRate[]>([])
const currencyEditingId = ref<number | null>(null)
const currencyForm = reactive({
  from_currency: 'CNY',
  to_currency: 'USD',
  rate: 0,
  fixed_amount: 0,
  rate_type: 'fixed',
  source: '',
})

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
    render: row => `${currencySymbol.value}${row.min_amount} - ${currencySymbol.value}${row.max_amount}`,
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

const exchangeRateColumns: DataTableColumns<ExchangeRate> = [
  {
    title: t('adminPayGateways.currencyPair'),
    key: 'pair',
    render: (row: ExchangeRate) => h(CurrencyPair, { from: row.from_currency, to: row.to_currency }),
  },
  { title: t('adminPayGateways.exchangeRate'), key: 'rate', render: (row: ExchangeRate) => row.rate.toFixed(8) },
  { title: t('adminPayGateways.fixedAmount'), key: 'fixed_amount', render: (row: ExchangeRate) => (row.fixed_amount || 0).toFixed(8) },
  {
    title: t('adminPayGateways.rateType'),
    key: 'rate_type',
    render: (row: ExchangeRate) => ({ fixed: t('adminPayGateways.rateTypeFixed'), dynamic: t('adminPayGateways.rateTypeDynamic') }[row.rate_type] || row.rate_type),
  },
  { title: t('adminPayGateways.source'), key: 'source' },
  {
    title: t('moneyScore.actions'),
    key: 'actions',
    render: (row: ExchangeRate) => h(NSpace, { size: 8 }, {
      default: () => [
        h(NButton, { size: 'small', quaternary: true, loading: isCurrencyRefreshing(row), onClick: () => handleRefreshExchangeRate(row) }, () => t('adminPayGateways.refreshRate')),
        h(NButton, { size: 'small', quaternary: true, onClick: () => handleEditExchangeRate(row) }, () => t('common.edit')),
        h(NButton, { size: 'small', quaternary: true, type: 'error', onClick: () => handleDeleteExchangeRate(row) }, () => t('adminPayGateways.delete')),
      ],
    }),
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

function fillNotifyUrl() {
  if (channelMeta.value?.default_notify_path) {
    form.notify_url = `${baseApiURL}${channelMeta.value.default_notify_path}`
  }
  else {
    message.warning(t('adminPayGateways.noDefaultNotifyUrl'))
  }
}

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

function normalizeConfigField(field: ChannelConfigField): ChannelConfigField {
  const f = { ...field }
  if (f.type === 'select' && f.options?.length) {
    f.options = f.options.map((opt: any) => ({ label: opt.label || opt.value, value: opt.value }))
  }
  return f
}

const exchangeSectionFields = new Set(['exchange_rate_mode', 'exchange_rate', 'exchange_fixed_amount', 'exchange_rate_source'])
const targetFeeSectionFields = new Set(['target_fee_rate', 'target_fee_fixed', 'target_fee_mode'])

function isFieldVisible(field: ChannelConfigField): boolean {
  // 查单间隔/批次只在开启主动查单时显示
  if (field.name === 'query_interval_seconds' || field.name === 'query_batch_size')
    return activeQueryEnabled.value

  // 汇率/目标手续费相关字段只在目标币种与结算币种不同且已设置目标币种时显示
  if (exchangeSectionFields.has(field.name) || targetFeeSectionFields.has(field.name)) {
    if (!needExchange.value)
      return false
    // 固定汇率：显示「固定汇率」和「固定加额」；动态汇率：显示「汇率源标识」
    const mode = String(extConfigMap.exchange_rate_mode || '').trim().toLowerCase()
    if (field.name === 'exchange_rate' || field.name === 'exchange_fixed_amount')
      return mode === 'fixed'
    if (field.name === 'exchange_rate_source')
      return mode === 'dynamic'
    return true
  }

  return true
}

function fillDynamicDefaults() {
  const defaultValues: Record<string, any> = {
    sign_type: signTypeOptions.value[0]?.value || 'MD5',
    active_query_enabled: '1',
    query_interval_seconds: '120',
    query_batch_size: '50',
    exchange_rate_mode: 'system',
    target_fee_mode: 'add',
  }

  dynamicConfigFields.value.forEach((field) => {
    const current = extConfigMap[field.name]
    if (current !== undefined && current !== null && String(current) !== '')
      return

    if (field.options?.length) {
      if (field.name === 'sign_type' && signTypeOptions.value.length > 0)
        extConfigMap[field.name] = signTypeOptions.value[0].value
      else
        extConfigMap[field.name] = field.options[0].value
      return
    }

    if (defaultValues[field.name] !== undefined)
      extConfigMap[field.name] = defaultValues[field.name]
  })
}

function pickVersion(version: string) {
  activeVersion.value = channelMeta.value?.versions.find(v => v.version === version) || null
  if (activeVersion.value) {
    signTypeOptions.value = activeVersion.value.signTypes.map(s => ({ label: s.name || s.value, value: s.value }))
  }
  else {
    signTypeOptions.value = []
  }

  const topFields = (channelMeta.value?.config_fields || channelMeta.value?.configFields || [])
    .map(normalizeConfigField)
  const versionFields = (activeVersion.value?.config_fields || activeVersion.value?.configFields || [])
    .map(normalizeConfigField)

  // 版本字段放前面，通用字段跟在后面做兜底/覆盖，这样 V1 的 key 会出现在表单顶部
  const fieldMap = new Map<string, ChannelConfigField>()
  versionFields.forEach(f => fieldMap.set(f.name, f))
  topFields.forEach(f => fieldMap.set(f.name, f))

  if (fieldMap.has('sign_type') && signTypeOptions.value.length > 0) {
    const signField = fieldMap.get('sign_type')!
    signField.options = signTypeOptions.value.map(opt => ({ label: opt.label as string, value: opt.value as string }))
    fieldMap.set('sign_type', signField)
  }

  dynamicConfigFields.value = Array.from(fieldMap.values())
  fillDynamicDefaults()
}

function syncExtConfigFromMap() {
  form.ext_config = Object.keys(extConfigMap).length === 0 ? '' : marshalExtConfig(extConfigMap)
}

function syncExtConfigToMap() {
  Object.keys(extConfigMap).forEach(k => delete extConfigMap[k])
  Object.assign(extConfigMap, parseExtConfig(form.ext_config))
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
    version: row.version || '',
    device: row.device || 'pc',
    currency: row.currency || 'CNY',
    description: row.description,
    status: row.status,
    api_url: row.api_url,
    pid: row.pid,
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
  if (!payload.ext_config && extConfigMap.key) {
    try {
      payload.ext_config = marshalExtConfig({ key: String(extConfigMap.key) })
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

async function openCurrencyModal() {
  showCurrencyModal.value = true
  resetCurrencyForm()
  currencyLoading.value = true
  try {
    const [ratesRes, baseRes] = await Promise.all([fetchExchangeRates(), getBaseCurrency()])
    if (ratesRes.isSuccess) {
      exchangeRates.value = ratesRes.data?.list || []
    }
    if (baseRes.isSuccess) {
      baseCurrency.value = baseRes.data?.base_currency || 'CNY'
    }
  }
  catch {
    // ignore
  }
  finally {
    currencyLoading.value = false
  }
}

async function handleSaveBaseCurrency() {
  try {
    const res = await setBaseCurrency(baseCurrency.value)
    if (res.isSuccess)
      message.success(t('adminPayGateways.saveSuccess'))
  }
  catch {
    // ignore
  }
}

function resetCurrencyForm() {
  currencyEditingId.value = null
  currencyForm.from_currency = baseCurrency.value || 'CNY'
  currencyForm.to_currency = 'USD'
  currencyForm.rate = 0
  currencyForm.fixed_amount = 0
  currencyForm.rate_type = 'fixed'
  currencyForm.source = ''
}

async function handleAddExchangeRate() {
  if (!currencyForm.from_currency || !currencyForm.to_currency || currencyForm.rate <= 0) {
    message.warning(t('adminPayGateways.exchangeRateFormInvalid'))
    return
  }
  try {
    const res = currencyEditingId.value
      ? await updateExchangeRate(currencyEditingId.value, { ...currencyForm })
      : await createExchangeRate({ ...currencyForm })
    if (res.isSuccess) {
      message.success(t('adminPayGateways.exchangeRateSaved'))
      resetCurrencyForm()
      await openCurrencyModal()
    }
  }
  catch {
    // ignore
  }
}

function handleEditExchangeRate(row: ExchangeRate) {
  currencyEditingId.value = row.id
  currencyForm.from_currency = row.from_currency
  currencyForm.to_currency = row.to_currency
  currencyForm.rate = row.rate
  currencyForm.fixed_amount = row.fixed_amount || 0
  currencyForm.rate_type = row.rate_type || 'fixed'
  currencyForm.source = row.source || ''
}

function handleDeleteExchangeRate(row: ExchangeRate) {
  dialog.warning({
    title: t('adminPayGateways.confirmDeleteTitle'),
    content: t('adminPayGateways.confirmDeleteExchangeRate', { from: row.from_currency, to: row.to_currency }),
    positiveText: t('adminPayGateways.delete'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await deleteExchangeRate(row.id)
        if (res.isSuccess) {
          message.success(t('adminPayGateways.deleteSuccess'))
          await openCurrencyModal()
        }
      }
      catch {
        // ignore
      }
    },
  })
}

async function handleRefreshRates() {
  try {
    const res = await refreshExchangeRates()
    if (res.isSuccess) {
      message.success(t('adminPayGateways.exchangeRateRefreshed'))
      await openCurrencyModal()
    }
  }
  catch {
    // ignore
  }
}

function isCurrencyRefreshing(row: ExchangeRate) {
  return currencyRefreshingId.value === row.id
}

async function handleRefreshExchangeRate(row: ExchangeRate) {
  currencyRefreshingId.value = row.id
  try {
    const res = await refreshExchangeRate(row.id)
    if (res.isSuccess) {
      message.success(t('adminPayGateways.exchangeRateRefreshed'))
      await openCurrencyModal()
    }
    else {
      message.error(res.message || t('adminPayGateways.exchangeRateRefreshFailed'))
    }
  }
  finally {
    currencyRefreshingId.value = null
  }
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

watch(() => form.currency, (val) => {
  const currency = (val || '').toUpperCase()
  if (!currency)
    return
  if (currency !== baseCurrency.value.toUpperCase()) {
    if (!extConfigMap.target_currency)
      extConfigMap.target_currency = baseCurrency.value
  }
  else {
    extConfigMap.target_currency = ''
    extConfigMap.exchange_rate_mode = 'system'
    extConfigMap.exchange_rate = '0'
    extConfigMap.exchange_fixed_amount = '0'
    extConfigMap.exchange_rate_source = ''
    extConfigMap.target_fee_rate = '0'
    extConfigMap.target_fee_fixed = '0'
    extConfigMap.target_fee_mode = 'add'
  }
})

async function loadBaseCurrency() {
  try {
    const res = await getBaseCurrency()
    if (res.isSuccess)
      baseCurrency.value = res.data?.base_currency || 'CNY'
  }
  catch {
    // ignore
  }
}

onMounted(() => {
  loadList()
  loadChannelMetas()
  loadBaseCurrency()
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
          <NButton @click="openCurrencyModal">
            {{ t('adminPayGateways.currencyManager') }}
          </NButton>
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
          <template v-for="field in dynamicConfigFields" :key="field.name">
            <n-gi v-if="field.name === 'target_currency'" :span="2" style="padding: 4px 0;">
              <n-divider title-placement="left">
                {{ t('adminPayGateways.exchangeSectionTitle') }}
              </n-divider>
              <n-alert v-if="!needExchange" type="info" :bordered="false" size="small">
                {{ t('adminPayGateways.exchangeSectionHint') }}
              </n-alert>
            </n-gi>
            <n-gi v-if="field.name === 'active_query_enabled'" :span="2" style="padding: 4px 0;">
              <n-divider title-placement="left">
                {{ t('adminPayGateways.querySectionTitle') }}
              </n-divider>
            </n-gi>
            <n-gi v-if="isFieldVisible(field)" :span="field.type === 'textarea' ? 2 : 1">
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
          </template>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.logoUrl')" path="logo_url">
              <n-input v-model:value="form.logo_url" :placeholder="t('adminPayGateways.logoUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.minAmount')" path="min_amount">
              <n-input-number v-model:value="form.min_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>
                  {{ currencySymbol }}
                </template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.maxAmount')" path="max_amount">
              <n-input-number v-model:value="form.max_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>
                  {{ currencySymbol }}
                </template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi :span="2" style="padding: 4px 0;">
            <n-divider title-placement="left">
              {{ t('adminPayGateways.channelFeeSectionTitle') }}
            </n-divider>
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
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.notifyUrl')" path="notify_url">
              <n-input-group>
                <n-input v-model:value="form.notify_url" :placeholder="t('adminPayGateways.notifyUrlPlaceholder')" />
                <NButton @click="fillNotifyUrl">
                  <template #icon>
                    <NIcon><icon-park-outline-refresh /></NIcon>
                  </template>
                  {{ t('adminPayGateways.fillNotifyUrl') }}
                </NButton>
              </n-input-group>
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-alert type="info" :title="t('adminPayGateways.feeExplainTitle')" :bordered="false" style="margin-top: 12px; white-space: pre-line;">
          {{ t('adminPayGateways.feeExplainContent') }}
        </n-alert>

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

    <!-- 汇率管理弹窗 -->
    <n-modal v-model:show="showCurrencyModal" preset="card" :title="t('adminPayGateways.currencyManager')" style="width: 760px; max-width: 92vw" :mask-closable="false">
      <NSpace vertical :size="16">
        <n-form inline :label-width="80">
          <n-form-item :label="t('adminPayGateways.baseCurrency')">
            <n-input v-model:value="baseCurrency" style="width: 120px" />
          </n-form-item>
          <n-form-item>
            <NButton type="primary" @click="handleSaveBaseCurrency">
              {{ t('adminPayGateways.saveBaseCurrency') }}
            </NButton>
          </n-form-item>
          <n-form-item>
            <NButton @click="handleRefreshRates">
              {{ t('adminPayGateways.refreshRates') }}
            </NButton>
          </n-form-item>
        </n-form>

        <n-form :label-width="80">
          <n-grid :cols="3" :x-gap="12">
            <n-gi>
              <n-form-item :label="t('adminPayGateways.fromCurrency')">
                <n-input v-model:value="currencyForm.from_currency" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item :label="t('adminPayGateways.toCurrency')">
                <n-input v-model:value="currencyForm.to_currency" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item :label="t('adminPayGateways.exchangeRate')">
                <n-input-number v-model:value="currencyForm.rate" :min="0" :precision="8" style="width: 100%" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item :label="t('adminPayGateways.rateType')">
                <n-select v-model:value="currencyForm.rate_type" :options="exchangeRateModeOptions" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item :label="t('adminPayGateways.fixedAmount')">
                <n-input-number v-model:value="currencyForm.fixed_amount" :min="0" :precision="8" style="width: 100%" />
              </n-form-item>
            </n-gi>
            <n-gi :span="2">
              <n-form-item :label="t('adminPayGateways.exchangeRateSource')">
                <n-input v-model:value="currencyForm.source" :placeholder="t('adminPayGateways.exchangeRateSourcePlaceholder')" />
              </n-form-item>
            </n-gi>
          </n-grid>
          <NButton type="primary" @click="handleAddExchangeRate">
            {{ currencyEditingId ? t('common.save') : t('adminPayGateways.addExchangeRate') }}
          </NButton>
          <NButton v-if="currencyEditingId" @click="resetCurrencyForm">
            {{ t('common.cancel') }}
          </NButton>
        </n-form>

        <n-data-table
          :columns="exchangeRateColumns"
          :data="exchangeRates"
          :loading="currencyLoading"
          size="small"
          :scroll-x="560"
        />
      </NSpace>
    </n-modal>
  </NSpace>
</template>
