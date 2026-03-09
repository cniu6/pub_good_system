<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch, h } from 'vue'
import {
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
  useMessage,
} from 'naive-ui'
import type { DataTableColumns, PaginationProps, SelectOption } from 'naive-ui'
import {
  fetchPayGateways,
  createPaymentOrder,
  fetchPaymentOrders,
  checkPaymentOrderStatus,
} from '@/service/api/user/payment'
import type { PayGateway, PaymentOrder } from '@/service/api/user/payment'
import { fetchUserProfile } from '@/service/api/user/login'
import { useAuthStore } from '@/store'

const message = useMessage()
const authStore = useAuthStore()

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
const selectedOrder = ref<PaymentOrder | null>(null)

// ========== 订单数据 ==========
const orderData = ref<PaymentOrder[]>([])
const refreshingOrders = ref<Set<number>>(new Set())
const autoRefreshTimer = ref<ReturnType<typeof setInterval> | null>(null)

// ========== 搜索和筛选 ==========
const statusFilter = ref(-1)

const statusOptions: SelectOption[] = [
  { label: '全部状态', value: -1 },
  { label: '待支付', value: 0 },
  { label: '已支付', value: 1 },
  { label: '已取消', value: 2 },
  { label: '已退款', value: 3 },
  { label: '支付失败', value: 4 },
]

// ========== 状态/支付方式映射 ==========
const statusMap: Record<number, { label: string; type: 'default' | 'success' | 'warning' | 'error' | 'info' }> = {
  0: { label: '待支付', type: 'warning' },
  1: { label: '已支付', type: 'success' },
  2: { label: '已取消', type: 'default' },
  3: { label: '已退款', type: 'info' },
  4: { label: '支付失败', type: 'error' },
}

const paymentTypeMap: Record<string, string> = {
  alipay: '支付宝',
  wxpay: '微信支付',
  qqpay: 'QQ钱包',
  bank: '银行卡',
  jdpay: '京东支付',
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

// ========== 分页 ==========
const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  pageSizes: [20, 40, 80, 100],
  showSizePicker: true,
  prefix(info) {
    return `共 ${info.itemCount ?? 0} 条`
  },
})

// ========== 表格列 ==========
const columns: DataTableColumns<PaymentOrder> = [
  {
    title: '订单号/交易号',
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
    title: '充值金额/实付',
    key: 'amount',
    width: 140,
    render(row) {
      return h('div', { style: 'display:flex;flex-direction:column;gap:2px' }, [
        h('span', { style: 'color:#18a058;font-weight:500' }, `¥${Number(row.amount).toFixed(2)}`),
        h('span', { style: 'font-size:12px;color:var(--primary-color)' }, `实付 ¥${Number(row.pay_amount).toFixed(2)}`),
      ])
    },
  },
  {
    title: '支付方式',
    key: 'payment_type',
    width: 100,
    render(row) {
      return paymentTypeMap[row.payment_type] || row.payment_type
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render(row) {
      const s = statusMap[row.status] || { label: '未知', type: 'default' as const }
      return h(NTag, { type: s.type, size: 'small' }, () => s.label)
    },
  },
  {
    title: '创建/更新时间',
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
    title: '操作',
    key: 'actions',
    width: 140,
    render(row) {
      const buttons = [
        h(NButton, { size: 'small', onClick: () => handleViewDetails(row) }, () => '详情'),
        h(NButton, {
          size: 'small',
          type: 'info',
          ghost: true,
          loading: refreshingOrders.value.has(row.id),
          onClick: () => handleRefreshOrder(row.id),
        }, () => '刷新'),
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
  } catch {
    console.error('刷新余额失败')
  } finally {
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
    } else {
      message.error('获取支付方式失败')
    }
  } catch {
    message.error('获取支付方式失败')
  } finally {
    gatewaysLoading.value = false
  }
}

// ========== 获取订单列表 ==========
async function fetchOrders() {
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
    if (res.isSuccess) {
      orderData.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    }
  } catch {
    message.error('获取订单记录失败')
  } finally {
    loading.value = false
  }
}

// ========== 自动刷新 ==========
function startAutoRefresh() {
  stopAutoRefresh()
  autoRefreshTimer.value = setInterval(async () => {
    const currentOrder = selectedOrder.value
    if (currentOrder && currentOrder.status === 0) {
      await handleRefreshOrder(currentOrder.id)
      const refreshedOrder = selectedOrder.value
      if (refreshedOrder?.status === 1) {
        stopAutoRefresh()
        showOrderDetail.value = false
        await refreshBalance()
        message.success('付款已完成，余额已更新')
      }
    } else {
      stopAutoRefresh()
    }
  }, 5000)
}

function stopAutoRefresh() {
  if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value)
    autoRefreshTimer.value = null
  }
}

// ========== 创建充值订单 ==========
async function createRechargeOrder() {
  if (!selectedGateway.value) {
    message.warning('请选择支付通道')
    return
  }
  if (!finalAmount.value || finalAmount.value <= 0) {
    message.warning('请输入充值金额')
    return
  }

  const gw = selectedGateway.value
  if (gw.min_amount > 0 && finalAmount.value < gw.min_amount) {
    message.warning(`该通道最低充值金额为 ¥${gw.min_amount}`)
    return
  }
  if (gw.max_amount > 0 && finalAmount.value > gw.max_amount) {
    message.warning(`该通道最高充值金额为 ¥${gw.max_amount}`)
    return
  }

  creating.value = true
  try {
    const res = await createPaymentOrder({
      gateway_id: gw.id,
      amount: finalAmount.value,
    })
    if (res.isSuccess && res.data) {
      showPaymentModal.value = false
      message.success('订单创建成功')

      // 刷新订单列表
      await fetchOrders()

      // 找到刚创建的订单并弹出详情（含二维码）
      const newOrder = orderData.value.find(o => o.order_no === res.data!.order_no)
      if (newOrder) {
        selectedOrder.value = { ...newOrder, pay_url: res.data.pay_url || newOrder.pay_url }
        showOrderDetail.value = true
        startAutoRefresh()
      }
    } else {
      message.error((res as any).message || '创建订单失败')
    }
  } catch {
    message.error('创建订单失败，请稍后重试')
  } finally {
    creating.value = false
  }
}

// ========== 支付处理 ==========
function handlePayment(order: PaymentOrder) {
  if (order.pay_url) {
    window.open(order.pay_url, '_blank')
  } else {
    message.error('支付链接不可用')
  }
}

// ========== 查看订单详情 ==========
function handleViewDetails(order: PaymentOrder) {
  selectedOrder.value = order
  showOrderDetail.value = true
  if (order.status === 0) {
    startAutoRefresh()
  }
}

// ========== 刷新订单状态 ==========
async function handleRefreshOrder(orderId: number) {
  refreshingOrders.value.add(orderId)
  try {
    const res = await checkPaymentOrderStatus(orderId)
    if (res.isSuccess) {
      message.success('订单状态已刷新')
      await fetchOrders()
      await refreshBalance()
      if (selectedOrder.value && selectedOrder.value.id === orderId) {
        const updated = orderData.value.find(o => o.id === orderId)
        if (updated) selectedOrder.value = { ...updated }
      }
    } else {
      message.error('刷新订单状态失败')
    }
  } catch {
    message.error('刷新订单状态失败')
  } finally {
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
    message.info('没有待支付的订单')
    return
  }
  let count = 0
  for (const order of pending) {
    try {
      refreshingOrders.value.add(order.id)
      const res = await checkPaymentOrderStatus(order.id)
      if (res.isSuccess) count++
    } catch { /* skip */ } finally {
      refreshingOrders.value.delete(order.id)
    }
  }
  message.success(`成功刷新 ${count} 个订单状态`)
  await fetchOrders()
  await refreshBalance()
}

// ========== 格式化时间 ==========
function formatTime(ts: number | null | undefined) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString()
}

// ========== 生命周期 ==========
watch(() => showPaymentModal.value, (show) => {
  if (show) fetchGateways()
})

watch(() => showOrderDetail.value, (show) => {
  if (!show) stopAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})

onMounted(() => {
  refreshBalance()
  fetchOrders()
})
</script>

<template>
  <div class="user-recharge-page">
    <!-- 余额显示和充值操作卡片 -->
    <NCard class="balance-card" title="账户余额">
      <template #header-extra>
        <NButton :loading="balanceLoading" size="small" type="primary" ghost @click="refreshBalance">
          刷新
        </NButton>
      </template>

      <NGrid :cols="24" :x-gap="16" :y-gap="16" responsive="screen">
        <!-- 当前余额 -->
        <NGridItem span="24 800:10">
          <div class="balance-display">
            <NText class="balance-label">当前余额</NText>
            <div class="balance-value">
              <span class="balance-currency">¥</span>
              <span class="balance-number">{{ userBalance.toFixed(2) }}</span>
            </div>
          </div>
        </NGridItem>

        <!-- 快速充值 -->
        <NGridItem span="24 800:14">
          <div class="quick-recharge-section">
            <NText class="section-title">在线充值</NText>

            <!-- 快速选择金额 -->
            <NSpace class="amount-buttons" wrap>
              <NButton
                v-for="amt in quickAmounts"
                :key="amt"
                :type="selectedAmount === amt ? 'primary' : 'default'"
                @click="selectAmount(amt)"
              >
                ¥{{ amt }}
              </NButton>
            </NSpace>

            <!-- 充值金额输入 + 按钮 -->
            <div class="recharge-input-row">
              <NInputNumber
                v-model:value="customAmount"
                :min="0.01"
                :max="99999.99"
                :precision="2"
                placeholder="请输入充值金额"
                class="recharge-input"
                @update:value="onCustomAmountChange"
              >
                <template #prefix>¥</template>
              </NInputNumber>

              <NButton
                type="primary"
                :disabled="!finalAmount || finalAmount <= 0"
                :loading="creating"
                @click="showPaymentModal = true"
              >
                立即充值 ¥{{ finalAmount?.toFixed(2) || '0.00' }}
              </NButton>
            </div>
          </div>
        </NGridItem>
      </NGrid>
    </NCard>

    <!-- 订单记录表格 -->
    <NCard class="records-card" title="订单记录">
      <template #header-extra>
        <NSpace :size="8" align="center">
          <NSelect
            v-model:value="statusFilter"
            :options="statusOptions"
            placeholder="状态"
            size="small"
            style="width: 120px"
            @update:value="handleSearch"
          />
          <NButton size="small" @click="handleReset">重置</NButton>
          <NButton size="small" type="warning" ghost @click="handleRefreshAllPending">批量刷新</NButton>
        </NSpace>
      </template>
      <div class="table-container">
        <NDataTable
          :columns="columns"
          :data="orderData"
          :loading="loading"
          :pagination="pagination"
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
      title="选择支付方式"
      class="payment-modal"
      :auto-focus="false"
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
                      <NText class="gateway-name">{{ gateway.name }}</NText>
                      <NText depth="3" class="gateway-type">{{ paymentTypeMap[gateway.pay_type] || gateway.pay_type }}</NText>
                    </div>
                  </div>
                  <NText depth="3" class="gateway-desc">
                    {{ gateway.description || '安全便捷的支付方式' }}
                  </NText>
                </div>
                <div class="gateway-details">
                  <NText depth="3" class="gateway-range">
                    限额: ¥{{ gateway.min_amount }} - ¥{{ gateway.max_amount }}
                  </NText>
                  <NText depth="3" class="gateway-fee">
                    手续费: {{ gateway.fee_rate || 0 }}%
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
          <NEmpty description="暂无可用的支付方式" />
        </div>
      </NSpin>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPaymentModal = false">取消</NButton>
          <NButton
            type="primary"
            :disabled="!selectedGateway"
            :loading="creating"
            @click="createRechargeOrder"
          >
            确认充值
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 订单详情弹窗 -->
    <NModal v-model:show="showOrderDetail" preset="card" title="订单详情" class="order-detail-modal">
      <template #header-extra>
        <NSpace v-if="selectedOrder?.status === 0 && autoRefreshTimer" align="center" :size="4">
          <NSpin size="small" />
          <NText depth="3" style="font-size: 12px">自动刷新中</NText>
        </NSpace>
      </template>

      <div v-if="selectedOrder" class="order-detail">
        <!-- 二维码区域：仅待支付且有支付链接时显示 -->
        <div v-if="selectedOrder.status === 0 && selectedOrder.pay_url" class="qrcode-section">
          <NQrCode :value="selectedOrder.pay_url" :size="180" />
          <NText depth="3" style="margin-top: 8px; font-size: 13px; text-align: center; display: block">
            请使用手机扫码完成支付
          </NText>
          <NDivider />
        </div>

        <NDescriptions :column="1" label-placement="left">
          <NDescriptionsItem label="订单号">
            {{ selectedOrder.order_no }}
          </NDescriptionsItem>
          <NDescriptionsItem label="交易号">
            {{ selectedOrder.trade_no || '-' }}
          </NDescriptionsItem>
          <NDescriptionsItem label="充值金额">
            <NText type="success">¥{{ Number(selectedOrder.amount).toFixed(2) }}</NText>
          </NDescriptionsItem>
          <NDescriptionsItem label="手续费">
            {{ selectedOrder.fee > 0 ? `¥${Number(selectedOrder.fee).toFixed(2)}` : '无' }}
          </NDescriptionsItem>
          <NDescriptionsItem label="实际支付">
            <NText type="info">¥{{ Number(selectedOrder.pay_amount).toFixed(2) }}</NText>
          </NDescriptionsItem>
          <NDescriptionsItem label="支付方式">
            {{ paymentTypeMap[selectedOrder.payment_type] || selectedOrder.payment_type }}
          </NDescriptionsItem>
          <NDescriptionsItem label="支付时间">
            {{ selectedOrder.paid_at ? formatTime(selectedOrder.paid_at) : '未支付' }}
          </NDescriptionsItem>
          <NDescriptionsItem label="创建时间">
            {{ formatTime(selectedOrder.create_time) }}
          </NDescriptionsItem>
          <NDescriptionsItem label="更新时间">
            {{ formatTime(selectedOrder.update_time) }}
          </NDescriptionsItem>
          <NDescriptionsItem label="订单状态">
            <NTag :type="(statusMap[selectedOrder.status] || { type: 'default' }).type">
              {{ (statusMap[selectedOrder.status] || { label: '未知' }).label }}
            </NTag>
          </NDescriptionsItem>
        </NDescriptions>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showOrderDetail = false">关闭</NButton>
          <NButton
            v-if="selectedOrder"
            type="info"
            ghost
            :loading="refreshingOrders.has(selectedOrder.id)"
            @click="handleRefreshOrder(selectedOrder!.id)"
          >
            刷新订单
          </NButton>
          <NButton
            v-if="selectedOrder?.status === 0 && selectedOrder?.pay_url"
            type="primary"
            @click="handlePayment(selectedOrder!)"
          >
            立即支付
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
