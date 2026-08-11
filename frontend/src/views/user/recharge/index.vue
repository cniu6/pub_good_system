<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NEmpty,
  NGrid,
  NGridItem,
  NInputNumber,
  NModal,
  NQrCode,
  NSelect,
  NSpace,
  NSpin,
  NTag,
  NText,
  useDialog,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns, PaginationProps, SelectOption } from 'naive-ui'
import { useRoute } from 'vue-router'
import {
  checkPaymentOrderStatus,
  createPaymentOrder,
  fetchPayGateways,
  fetchPaymentOrderDetail,
  fetchPaymentOrders,
} from '@/service/api/user/payment'
import type { PayGateway, PaymentOrder } from '@/service/api/user/payment'
import { fetchUserProfile } from '@/service/api/user/login'
import { useAuthStore } from '@/store'
import { useBaseCurrency } from '@/composables/useBaseCurrency'
import { useRequestGuard, withSubmitLock } from '@/hooks'

const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()
const { t } = useI18n()
const { currencySymbol } = useBaseCurrency()
const ordersFetchGuard = useRequestGuard()
const route = useRoute()

// ========== 加载状态 ==========
const loading = ref(false)
const balanceLoading = ref(false)
const gatewaysLoading = ref(false)
const creating = ref(false)

// ========== 用户余额 ==========
const userBalance = computed(() => authStore.userInfo?.money ?? 0)

// ========== 充值金额 ==========
const quickAmounts = [10, 50, 100, 200, 500, 1000]
const selectedAmount = ref<number | null>(null)
const customAmount = ref<number | null>(50)

const finalAmount = computed(() => {
  return customAmount.value || selectedAmount.value || 0
})

// ========== 支付通道 ==========
const showPaymentModal = ref(false)
const payGateways = ref<PayGateway[]>([])
const selectedGateway = ref<PayGateway | null>(null)

// ========== 订单详情弹窗 ==========
const showOrderDetail = ref(false)
const detailLoading = ref(false)
const selectedOrder = ref<PaymentOrder | null>(null)

// ========== 订单数据 ==========
const orderData = ref<PaymentOrder[]>([])
const refreshingOrders = ref<Set<number>>(new Set())

// ========== 支付回跳结果展示 ==========
type PaymentReturnResult = 'success' | 'pending' | 'error'

interface PaymentReturnState {
  result: PaymentReturnResult
  orderNo: string
  message: string
}

const paymentReturnState = ref<PaymentReturnState | null>(null)

// ========== 搜索和筛选 ==========
const statusFilter = ref(-1)

const statusOptions: SelectOption[] = [
  { label: t('recharge.allStatus'), value: -1 },
  { label: t('recharge.pending'), value: 0 },
  { label: t('recharge.paid'), value: 1 },
  { label: t('recharge.cancelled'), value: 2 },
  { label: t('recharge.refunded'), value: 3 },
  { label: t('recharge.failed'), value: 4 },
]

// ========== 状态/支付方式映射 ==========
const statusMap: Record<number, { label: string, type: 'default' | 'success' | 'warning' | 'error' | 'info' }> = {
  0: { label: t('recharge.pending'), type: 'warning' },
  1: { label: t('recharge.paid'), type: 'success' },
  2: { label: t('recharge.cancelled'), type: 'default' },
  3: { label: t('recharge.refunded'), type: 'info' },
  4: { label: t('recharge.failed'), type: 'error' },
}

const paymentReturnSummary = computed(() => {
  const state = paymentReturnState.value
  if (!state)
    return null

  const orderNoText = state.orderNo ? t('recharge.paymentReturnOrderNo', { orderNo: state.orderNo }) : ''

  if (state.result === 'success') {
    return {
      type: 'success' as const,
      tagType: 'success' as const,
      tagLabel: t('recharge.paymentReturnSuccessTag'),
      icon: '✅',
      title: t('recharge.paymentReturnSuccessTitle'),
      description: t('recharge.paymentReturnSuccessDesc'),
      detail: orderNoText,
    }
  }

  if (state.result === 'pending') {
    return {
      type: 'warning' as const,
      tagType: 'warning' as const,
      tagLabel: t('recharge.paymentReturnPendingTag'),
      icon: '⏳',
      title: t('recharge.paymentReturnPendingTitle'),
      description: t('recharge.paymentReturnPendingDesc'),
      detail: orderNoText,
    }
  }

  const description = state.message === 'invalid_callback'
    ? t('recharge.paymentReturnInvalidCallback')
    : state.message
      ? t('recharge.paymentReturnErrorReason', { reason: state.message })
      : t('recharge.paymentReturnErrorDesc')

  return {
    type: 'error' as const,
    tagType: 'error' as const,
    tagLabel: t('recharge.paymentReturnErrorTag'),
    icon: '⚠️',
    title: t('recharge.paymentReturnErrorTitle'),
    description,
    detail: orderNoText,
  }
})

const paymentTypeMap: Record<string, string> = {
  alipay: t('recharge.alipay'),
  wxpay: t('recharge.wechatPay'),
  qqpay: t('recharge.qqWallet'),
  bank: t('recharge.bankCard'),
  jdpay: t('recharge.jdPay'),
  manual: t('recharge.manual'),
}

function payTypeIcon(payType: string): string {
  const iconMap: Record<string, string> = {
    alipay: '💙',
    wxpay: '💚',
    qqpay: '🐧',
    bank: '🏦',
    jdpay: '🔴',
  }
  return iconMap[payType] || '💳'
}

function getRouteQueryText(value: unknown): string {
  if (Array.isArray(value))
    return typeof value[0] === 'string' ? value[0] : ''
  return typeof value === 'string' ? value : ''
}

function syncPaymentReturnState() {
  const result = getRouteQueryText(route.query.result)
  const orderNo = getRouteQueryText(route.query.order_no)
  const messageText = getRouteQueryText(route.query.msg)

  if (result === 'success' || result === 'pending' || result === 'error') {
    paymentReturnState.value = {
      result,
      orderNo,
      message: messageText,
    }
    return
  }

  paymentReturnState.value = null
}

function dismissPaymentReturnState() {
  paymentReturnState.value = null
}

function getReturnedOrderRowClass(row: PaymentOrder) {
  const orderNo = paymentReturnState.value?.orderNo
  if (!orderNo)
    return ''
  return row.order_no === orderNo ? 'payment-return-order-row' : ''
}

// ========== 分页 ==========
const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  pageSizes: [20, 40, 80, 100],
  showSizePicker: true,
  prefix(info) {
    return t('recharge.totalItems', { count: info.itemCount ?? 0 })
  },
})

// ========== 表格列 ==========
const columns: DataTableColumns<PaymentOrder> = [
  {
    title: t('recharge.orderTradeNo'),
    key: 'order_no',
    width: 200,
    render(row) {
      return h('div', { style: 'display:flex;flex-direction:column;gap:2px' }, [
        h('span', { style: 'font-size:13px' }, row.order_no),
        h('span', { style: 'font-size:12px;color:#999' }, row.trade_no || '-'),
      ])
    },
  },
  {
    title: t('recharge.amountPaid'),
    key: 'amount',
    width: 140,
    render(row) {
      return h('div', { style: 'display:flex;flex-direction:column;gap:2px' }, [
        h('span', { style: 'color:#18a058;font-weight:500' }, `${currencySymbol.value}${Number(row.amount).toFixed(2)}`),
        h('span', { style: 'font-size:12px;color:var(--primary-color)' }, t('recharge.actualPaid', { amount: Number(row.pay_amount).toFixed(2) })),
      ])
    },
  },
  {
    title: t('recharge.paymentMethod'),
    key: 'payment_type',
    width: 100,
    render(row) {
      return paymentTypeMap[row.payment_type] || row.payment_type
    },
  },
  {
    title: t('recharge.status'),
    key: 'status',
    width: 90,
    render(row) {
      const s = statusMap[row.status] || { label: t('recharge.unknown'), type: 'default' as const }
      return h(NTag, { type: s.type, size: 'small' }, () => s.label)
    },
  },
  {
    title: t('recharge.createdUpdatedAt'),
    key: 'create_time',
    width: 160,
    render(row) {
      return h('div', { style: 'display:flex;flex-direction:column;gap:2px' }, [
        h('span', {}, row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-'),
        h('span', { style: 'font-size:12px;color:#999' }, row.update_time ? new Date(row.update_time * 1000).toLocaleString() : '-'),
      ])
    },
  },
  {
    title: t('recharge.actions'),
    key: 'actions',
    width: 140,
    render(row) {
      const buttons = [
        h(NButton, { size: 'small', onClick: () => handleViewDetails(row) }, () => t('recharge.detail')),
        h(NButton, {
          size: 'small',
          type: 'info',
          ghost: true,
          loading: refreshingOrders.value.has(row.id),
          onClick: () => handleRefreshOrder(row.id),
        }, () => t('common.reload')),
      ]
      return h(NSpace, { size: 'small' }, () => buttons)
    },
  },
]

// ========== 选择金额 ==========
function selectAmount(amount: number) {
  selectedAmount.value = amount
  customAmount.value = amount
}

function onCustomAmountChange(value: number | null) {
  if (value !== null) {
    selectedAmount.value = null
  }
}

// ========== 选择网关 ==========
function selectGateway(gateway: PayGateway) {
  selectedGateway.value = gateway
}

// ========== 刷新余额 ==========
async function refreshBalance() {
  balanceLoading.value = true
  try {
    const res = await fetchUserProfile()
    if (res.isSuccess && res.data) {
      authStore.updateUserInfo({ money: res.data.money, score: res.data.score })
    }
  }
  catch {
    if (import.meta.env.DEV)
      console.error('[recharge] refresh balance failed')
  }
  finally {
    balanceLoading.value = false
  }
}

// ========== 获取支付网关 ==========
async function fetchGateways() {
  gatewaysLoading.value = true
  try {
    const res = await fetchPayGateways()
    if (res.isSuccess && res.data) {
      payGateways.value = (res.data.list || []).filter((gw: PayGateway) => gw.status === 1)
    }
    else {
      message.error(t('recharge.fetchPaymentMethodsFailed'))
    }
  }
  catch {
    message.error(t('recharge.fetchPaymentMethodsFailed'))
  }
  finally {
    gatewaysLoading.value = false
  }
}

// ========== 获取订单列表 ==========
async function fetchOrders() {
  const token = ordersFetchGuard.begin()
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page || 1,
      page_size: pagination.pageSize || 20,
    }
    if (statusFilter.value >= 0) {
      params.status = statusFilter.value
    }
    const res = await fetchPaymentOrders(params)
    if (!ordersFetchGuard.isLatest(token))
      return
    if (res.isSuccess) {
      orderData.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    }
  }
  catch {
    if (ordersFetchGuard.isLatest(token))
      message.error(t('recharge.fetchOrdersFailed'))
  }
  finally {
    if (ordersFetchGuard.isLatest(token))
      loading.value = false
  }
}

// ========== 创建充值订单 ==========
function amountToFen(yuan: number): number {
  return Math.round(Number(yuan) * 100)
}

/** 待支付列表里找最近一笔同金额订单（amount / pay_amount 任一匹配） */
function findSameAmountPendingOrder(amountYuan: number, pendingOrders: PaymentOrder[]): PaymentOrder | null {
  const fen = amountToFen(amountYuan)
  for (const order of pendingOrders) {
    if (amountToFen(order.amount) === fen || amountToFen(order.pay_amount) === fen)
      return order
  }
  return null
}

async function createRechargeOrder() {
  if (creating.value)
    return
  if (!selectedGateway.value) {
    message.warning(t('recharge.selectGateway'))
    return
  }
  if (!finalAmount.value || finalAmount.value <= 0) {
    message.warning(t('recharge.enterAmount'))
    return
  }

  const gw = selectedGateway.value
  if (gw.min_amount > 0 && finalAmount.value < gw.min_amount) {
    message.warning(t('recharge.minAmountError', { amount: gw.min_amount }))
    return
  }
  if (gw.max_amount > 0 && finalAmount.value > gw.max_amount) {
    message.warning(t('recharge.maxAmountError', { amount: gw.max_amount }))
    return
  }

  // 在任何 await 前加锁，避免连点并发建单
  creating.value = true
  let existingPending: PaymentOrder | null = null
  try {
    // 有同金额待支付单则二次确认：去支付旧单 / 继续创建新单
    try {
      const pendingRes = await fetchPaymentOrders({ page: 1, page_size: 50, status: 0 })
      existingPending = pendingRes.isSuccess
        ? findSameAmountPendingOrder(finalAmount.value, pendingRes.data?.list || [])
        : null
    }
    catch {
      // 检测失败不阻断建单
    }

    if (existingPending) {
      // 弹窗前释放锁，否则「继续创建」会被挡住
      creating.value = false
      const existing = existingPending
      dialog.warning({
        title: t('recharge.sameAmountPendingTitle'),
        content: t('recharge.sameAmountPendingContent', {
          amount: finalAmount.value.toFixed(2),
          orderNo: existing.order_no,
        }),
        positiveText: t('recharge.goPayExisting'),
        negativeText: t('recharge.continueCreate'),
        onPositiveClick: () => {
          showPaymentModal.value = false
          void handleViewDetails(existing)
        },
        onNegativeClick: () => {
          void doCreateRechargeOrder()
        },
      })
      return
    }

    await createRechargeOrderCore(gw)
  }
  catch {
    message.error(t('recharge.createOrderRetry'))
  }
  finally {
    creating.value = false
  }
}

async function doCreateRechargeOrder() {
  const gw = selectedGateway.value
  if (!gw)
    return
  await withSubmitLock(creating, async () => {
    try {
      await createRechargeOrderCore(gw)
    }
    catch {
      message.error(t('recharge.createOrderRetry'))
    }
  })
}

/** 实际建单逻辑（调用方负责 creating 锁） */
async function createRechargeOrderCore(gw: PayGateway) {
  const res = await createPaymentOrder({
    gateway_id: gw.id,
    amount: finalAmount.value,
  })
  if (!res.isSuccess || !res.data) {
    message.error((res as { message?: string }).message || t('recharge.createOrderFailed'))
    return
  }

  showPaymentModal.value = false
  message.success(t('recharge.orderCreated'))
  await fetchOrders()

  // 优先详情接口拿完整订单，列表未命中时用建单响应兜底
  const createdOrder = orderData.value.find(o => o.order_no === res.data!.order_no)
  if (createdOrder) {
    const detailRes = await fetchPaymentOrderDetail(createdOrder.id)
    selectedOrder.value = detailRes.isSuccess && detailRes.data
      ? detailRes.data
      : { ...createdOrder, pay_url: res.data.pay_url || createdOrder.pay_url }
  }
  else {
    selectedOrder.value = {
      id: 0,
      user_id: authStore.userInfo?.id || 0,
      gateway_id: gw.id,
      trade_no: '',
      payment_channel: gw.type,
      payment_type: res.data.payment_type || gw.pay_type,
      amount: res.data.amount,
      fee: res.data.fee,
      pay_amount: res.data.pay_amount,
      subject: '',
      status: 0,
      notify_count: 0,
      pay_url: res.data.pay_url,
      paid_at: null,
      expire_at: res.data.expire_at,
      client_ip: '',
      create_time: Math.floor(Date.now() / 1000),
      update_time: Math.floor(Date.now() / 1000),
      order_no: res.data.order_no,
    }
  }

  showOrderDetail.value = true
}

// ========== 支付处理 ==========
function handlePayment(order: PaymentOrder) {
  if (order.pay_url) {
    window.open(order.pay_url, '_blank')
  }
  else {
    message.error(t('recharge.paymentUrlUnavailable'))
  }
}

// ========== 查看订单详情 ==========
async function handleViewDetails(order: PaymentOrder) {
  selectedOrder.value = order
  showOrderDetail.value = true
  detailLoading.value = true
  try {
    const res = await fetchPaymentOrderDetail(order.id)
    if (res.isSuccess && res.data) {
      selectedOrder.value = res.data
    }
    else {
      message.error(res.message || t('recharge.fetchOrderDetailFailed'))
    }
  }
  catch {
    message.error(t('recharge.fetchOrderDetailFailed'))
  }
  finally {
    detailLoading.value = false
  }
}

// ========== 刷新订单状态 ==========
async function handleRefreshOrder(orderId: number) {
  refreshingOrders.value.add(orderId)
  try {
    const res = await checkPaymentOrderStatus(orderId)
    if (res.isSuccess) {
      message.success(t('recharge.orderStatusRefreshed'))
      await fetchOrders()
      await refreshBalance()
      if (selectedOrder.value && selectedOrder.value.id === orderId) {
        const updated = orderData.value.find(o => o.id === orderId)
        if (updated)
          selectedOrder.value = { ...updated }
      }
    }
    else {
      message.error(t('recharge.refreshOrderStatusFailed'))
    }
  }
  catch {
    message.error(t('recharge.refreshOrderStatusFailed'))
  }
  finally {
    refreshingOrders.value.delete(orderId)
  }
}

// ========== 分页处理 ==========
function handlePageChange(page: number) {
  pagination.page = page
  fetchOrders()
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchOrders()
}

function handleSearch() {
  pagination.page = 1
  fetchOrders()
}

function handleReset() {
  statusFilter.value = -1
  pagination.page = 1
  fetchOrders()
}

// ========== 批量刷新待支付 ==========
async function handleRefreshAllPending() {
  const pending = orderData.value.filter(o => o.status === 0)
  if (pending.length === 0) {
    message.info(t('recharge.noPendingOrders'))
    return
  }
  let count = 0
  for (const order of pending) {
    try {
      refreshingOrders.value.add(order.id)
      const res = await checkPaymentOrderStatus(order.id)
      if (res.isSuccess)
        count++
    }
    catch { /* skip */ }
    finally {
      refreshingOrders.value.delete(order.id)
    }
  }
  message.success(t('recharge.bulkRefreshSuccess', { count }))
  await fetchOrders()
  await refreshBalance()
}

// ========== 格式化时间 ==========
function formatTime(ts: number | null | undefined) {
  if (!ts)
    return '-'
  return new Date(ts * 1000).toLocaleString()
}

// ========== 生命周期 ==========
watch(() => showPaymentModal.value, (show) => {
  if (show)
    fetchGateways()
})

watch(
  () => [route.query.result, route.query.order_no, route.query.msg],
  () => {
    syncPaymentReturnState()
  },
  { immediate: true },
)

onMounted(() => {
  refreshBalance()
  fetchOrders()
})
</script>

<template>
  <div class="user-recharge-page">
    <NAlert
      v-if="paymentReturnSummary"
      class="payment-return-alert"
      :type="paymentReturnSummary.type"
      closable
      @close="dismissPaymentReturnState"
    >
      <div class="payment-return-content">
        <div class="payment-return-main">
          <div class="payment-return-head">
            <span class="payment-return-icon">{{ paymentReturnSummary.icon }}</span>
            <NText strong class="payment-return-title">
              {{ paymentReturnSummary.title }}
            </NText>
            <NTag size="small" :type="paymentReturnSummary.tagType">
              {{ paymentReturnSummary.tagLabel }}
            </NTag>
          </div>
          <NText depth="3" class="payment-return-desc">
            {{ paymentReturnSummary.description }}
          </NText>
          <NText v-if="paymentReturnSummary.detail" depth="3" class="payment-return-detail">
            {{ paymentReturnSummary.detail }}
          </NText>
        </div>
        <NSpace class="payment-return-actions" size="small">
          <NButton size="small" @click="refreshBalance">
            {{ t('common.reload') }}
          </NButton>
          <NButton size="small" @click="fetchOrders">
            {{ t('recharge.refreshOrders') }}
          </NButton>
        </NSpace>
      </div>
    </NAlert>

    <!-- 余额显示和充值操作卡片 -->
    <NCard class="balance-card" :title="t('recharge.balanceTitle')">
      <template #header-extra>
        <NButton :loading="balanceLoading" size="small" type="primary" ghost @click="refreshBalance">
          {{ t('common.reload') }}
        </NButton>
      </template>

      <NGrid :cols="24" :x-gap="16" :y-gap="16" responsive="screen">
        <!-- 当前余额 -->
        <NGridItem span="24 800:10">
          <div class="balance-display">
            <NText class="balance-label">
              {{ t('recharge.currentBalance') }}
            </NText>
            <div class="balance-value">
              <span class="balance-currency">{{ currencySymbol }}</span>
              <span class="balance-number">{{ userBalance.toFixed(2) }}</span>
            </div>
          </div>
        </NGridItem>

        <!-- 快速充值 -->
        <NGridItem span="24 800:14">
          <div class="quick-recharge-section">
            <NText class="section-title">
              {{ t('recharge.onlineRecharge') }}
            </NText>

            <!-- 快速选择金额 -->
            <NSpace class="amount-buttons" wrap>
              <NButton
                v-for="amt in quickAmounts"
                :key="amt"
                :type="selectedAmount === amt ? 'primary' : 'default'"
                @click="selectAmount(amt)"
              >
                {{ currencySymbol }}{{ amt }}
              </NButton>
            </NSpace>

            <!-- 充值金额输入 + 按钮 -->
            <div class="recharge-input-row">
              <NInputNumber
                v-model:value="customAmount"
                :min="0.01"
                :max="99999.99"
                :precision="2"
                :placeholder="t('recharge.enterAmount')"
                class="recharge-input"
                @update:value="onCustomAmountChange"
              >
                <template #prefix>
                  {{ currencySymbol }}
                </template>
              </NInputNumber>

              <NButton
                type="primary"
                :disabled="!finalAmount || finalAmount <= 0"
                :loading="creating"
                @click="showPaymentModal = true"
              >
                {{ t('recharge.rechargeNow', { amount: finalAmount?.toFixed(2) || '0.00' }) }}
              </NButton>
            </div>
          </div>
        </NGridItem>
      </NGrid>
    </NCard>

    <!-- 订单记录表格 -->
    <NCard class="records-card" :title="t('recharge.orderRecords')">
      <template #header-extra>
        <NSpace :size="8" align="center">
          <NSelect
            v-model:value="statusFilter"
            :options="statusOptions"
            :placeholder="t('recharge.status')"
            size="small"
            style="width: 120px"
            @update:value="handleSearch"
          />
          <NButton size="small" @click="handleReset">
            {{ t('common.reset') }}
          </NButton>
          <NButton size="small" type="warning" ghost @click="handleRefreshAllPending">
            {{ t('recharge.bulkRefresh') }}
          </NButton>
        </NSpace>
      </template>
      <div class="table-container">
        <NDataTable
          :columns="columns"
          :data="orderData"
          :loading="loading"
          :pagination="pagination"
          :row-class-name="getReturnedOrderRowClass"
          :row-key="(row: PaymentOrder) => row.id"
          striped
          size="small"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </NCard>

    <!-- 选择支付通道弹窗 -->
    <NModal
      v-model:show="showPaymentModal"
      preset="card"
      :title="t('recharge.selectPaymentMethod')"
      class="payment-modal"
      :auto-focus="false"
      :mask-closable="!creating"
    >
      <NSpin :show="gatewaysLoading">
        <div v-if="payGateways.length > 0" class="gateway-grid">
          <div
            v-for="gateway in payGateways"
            :key="gateway.id"
            class="gateway-item"
          >
            <NCard
              size="small"
              :bordered="true"
              class="gateway-card"
              :class="{
                selected: selectedGateway?.id === gateway.id,
              }"
              @click="selectGateway(gateway)"
            >
              <div class="gateway-content">
                <div class="gateway-info">
                  <div class="gateway-header">
                    <img
                      v-if="gateway.logo_url"
                      :src="gateway.logo_url"
                      alt=""
                      class="gateway-logo"
                      referrerpolicy="no-referrer"
                    >
                    <div v-else class="gateway-logo-placeholder">
                      <span>{{ payTypeIcon(gateway.pay_type) }}</span>
                    </div>
                    <div class="gateway-title">
                      <NText class="gateway-name">
                        {{ gateway.name }}
                      </NText>
                      <NText depth="3" class="gateway-type">
                        {{ paymentTypeMap[gateway.pay_type] || gateway.pay_type }}
                      </NText>
                    </div>
                  </div>
                  <NText depth="3" class="gateway-desc">
                    {{ gateway.description || t('recharge.safeConvenientPayment') }}
                  </NText>
                </div>
                <div class="gateway-details">
                  <NText depth="3" class="gateway-range">
                    {{ t('recharge.limitRange', { min: gateway.min_amount, max: gateway.max_amount }) }}
                  </NText>
                  <NText depth="3" class="gateway-fee">
                    {{ t('recharge.feeRate', { rate: gateway.fee_rate || 0 }) }}
                  </NText>
                  <NTag v-if="gateway.min_level > 0" size="small" type="info">
                    Lv.{{ gateway.min_level }}+
                  </NTag>
                </div>
              </div>
            </NCard>
          </div>
        </div>
        <div v-else class="empty-gateways">
          <NEmpty :description="t('recharge.noPaymentMethods')" />
        </div>
      </NSpin>

      <template #footer>
        <NSpace justify="end">
          <NButton :disabled="creating" @click="showPaymentModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton
            type="primary"
            :disabled="!selectedGateway || creating"
            :loading="creating"
            @click="createRechargeOrder"
          >
            {{ t('recharge.confirmRecharge') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 订单详情弹窗 -->
    <NModal v-model:show="showOrderDetail" preset="card" :title="t('recharge.orderDetail')" class="order-detail-modal">
      <NSpin :show="detailLoading">
        <div v-if="selectedOrder" class="order-detail">
          <!-- 二维码区域：仅待支付且有支付链接时显示 -->
          <div v-if="selectedOrder.status === 0 && selectedOrder.pay_url" class="qrcode-section">
            <NQrCode :value="selectedOrder.pay_url" :size="180" />
            <NText depth="3" style="margin-top: 8px; font-size: 13px; text-align: center; display: block">
              {{ t('recharge.scanToPay') }}
            </NText>
            <NDivider />
          </div>

          <NDescriptions :column="1" label-placement="left">
            <NDescriptionsItem :label="t('recharge.orderNo')">
              {{ selectedOrder.order_no }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.tradeNo')">
              {{ selectedOrder.trade_no || '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.rechargeAmount')">
              <NText type="success">
                {{ currencySymbol }}{{ Number(selectedOrder.amount).toFixed(2) }}
              </NText>
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.fee')">
              {{ selectedOrder.fee > 0 ? `${currencySymbol}${Number(selectedOrder.fee).toFixed(2)}` : t('recharge.none') }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.actualPayment')">
              <NText type="info">
                {{ currencySymbol }}{{ Number(selectedOrder.pay_amount).toFixed(2) }}
              </NText>
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.paymentMethod')">
              {{ paymentTypeMap[selectedOrder.payment_type] || selectedOrder.payment_type }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.paymentTime')">
              {{ selectedOrder.paid_at ? formatTime(selectedOrder.paid_at) : t('recharge.unpaid') }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.createdAt')">
              {{ formatTime(selectedOrder.create_time) }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.updatedAt')">
              {{ formatTime(selectedOrder.update_time) }}
            </NDescriptionsItem>
            <NDescriptionsItem :label="t('recharge.orderStatus')">
              <NTag :type="(statusMap[selectedOrder.status] || { type: 'default' }).type">
                {{ (statusMap[selectedOrder.status] || { label: t('recharge.unknown') }).label }}
              </NTag>
            </NDescriptionsItem>
          </NDescriptions>
        </div>
      </NSpin>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showOrderDetail = false">
            {{ t('common.close') }}
          </NButton>
          <NButton
            v-if="selectedOrder"
            type="info"
            ghost
            :loading="refreshingOrders.has(selectedOrder.id)"
            @click="handleRefreshOrder(selectedOrder!.id)"
          >
            {{ t('recharge.refreshOrder') }}
          </NButton>
          <NButton
            v-if="selectedOrder?.status === 0 && selectedOrder?.pay_url"
            type="primary"
            @click="handlePayment(selectedOrder!)"
          >
            {{ t('recharge.payNow') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.balance-card,
.records-card {
  margin-bottom: 16px;
}

.payment-return-alert {
  margin-bottom: 16px;
}

.payment-return-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  flex-wrap: wrap;
}

.payment-return-main {
  flex: 1;
  min-width: 0;
}

.payment-return-head {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.payment-return-icon {
  font-size: 18px;
  line-height: 1;
}

.payment-return-title {
  font-size: 15px;
}

.payment-return-desc {
  display: block;
  margin-top: 8px;
  line-height: 1.6;
}

.payment-return-detail {
  display: block;
  margin-top: 4px;
  font-size: 12px;
}

.payment-return-actions {
  flex-shrink: 0;
}

:deep(.payment-return-order-row td) {
  background: rgba(24, 160, 88, 0.06) !important;
}

.balance-display {
  flex-shrink: 0;
  min-width: 200px;
}

.balance-label {
  display: block;
  font-size: 14px;
  margin-bottom: 8px;
  opacity: 0.7;
}

.balance-value {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.balance-currency {
  font-size: 20px;
  font-weight: 600;
  color: #18a058;
}

.balance-number {
  font-size: 36px;
  font-weight: 700;
  color: #18a058;
  background: linear-gradient(135deg, #18a058, #2dd07a);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: balancePulse 2s ease-in-out infinite;
}

@keyframes balancePulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.8; }
}

.quick-recharge-section {
  flex: 1;
}

.section-title {
  display: block;
  margin-bottom: 12px;
  font-weight: 500;
  font-size: 14px;
}

.amount-buttons {
  margin-top: 8px;
  margin-bottom: 16px;
}

.recharge-input-row {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.recharge-input {
  flex: 1;
  min-width: 200px;
  max-width: 280px;
}

.table-container {
  overflow-x: auto;
}

/* ========== 支付通道弹窗 ========== */
.gateway-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
}

.gateway-item {
  width: 100%;
}

.gateway-card {
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.gateway-card:hover {
  border-color: #18a058;
}

.gateway-card.selected {
  border-color: #18a058;
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.2);
}

.gateway-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 1;
}

.gateway-info {
  flex: 1;
}

.gateway-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.gateway-logo {
  width: 32px;
  height: 32px;
  margin-right: 12px;
  border-radius: 4px;
  object-fit: contain;
}

.gateway-logo-placeholder {
  width: 32px;
  height: 32px;
  margin-right: 12px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  background: #f5f5f5;
}

.gateway-title {
  display: flex;
  flex-direction: column;
}

.gateway-name {
  font-size: 16px;
  font-weight: 500;
  line-height: 1.2;
}

.gateway-type {
  font-size: 12px;
  margin-top: 2px;
}

.gateway-desc {
  display: block;
  font-size: 12px;
  margin-top: 4px;
}

.gateway-details {
  text-align: right;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.gateway-range,
.gateway-fee {
  font-size: 12px;
}

.empty-gateways {
  padding: 40px 0;
  text-align: center;
}

/* ========== 订单详情 ========== */
.order-detail {
  padding: 8px;
}

.qrcode-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 0 0;
}

/* ========== 响应式 ========== */
@media (max-width: 768px) {
  .recharge-input-row {
    flex-direction: column;
    align-items: stretch;
  }

  .recharge-input {
    width: 100%;
    max-width: none;
  }
}

@media (max-width: 480px) {
  .gateway-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .gateway-details {
    align-items: flex-start;
    text-align: left;
  }

  .gateway-grid {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
/* NModal teleports outside scoped scope, so use global styles */
.payment-modal {
  width: 90%;
  max-width: 900px;
}

.order-detail-modal {
  width: 90%;
  max-width: 600px;
}

@media (min-width: 769px) {
  .payment-modal {
    width: 70%;
  }
}

@media (min-width: 1200px) {
  .payment-modal {
    width: 60%;
  }
}
</style>
