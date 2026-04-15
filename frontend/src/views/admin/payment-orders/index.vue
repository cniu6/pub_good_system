<template>
  <n-space vertical :size="16">
    <!-- 统计卡片 -->
    <n-grid :cols="4" :x-gap="12">
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminPaymentOrders.todayRevenue')" :value="stats.today_amount" :precision="2">
            <template #prefix>¥</template>
          </n-statistic>
          <n-text depth="3" style="font-size: 12px">{{ t('adminPaymentOrders.orderCount', { count: stats.today_orders }) }}</n-text>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminPaymentOrders.totalRevenue')" :value="stats.total_amount" :precision="2">
            <template #prefix>¥</template>
          </n-statistic>
          <n-text depth="3" style="font-size: 12px">{{ t('adminPaymentOrders.orderCount', { count: stats.paid_orders }) }}</n-text>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminPaymentOrders.totalOrders')" :value="stats.total_orders" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('recharge.pending')" :value="stats.pending_orders" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 订单列表 -->
    <n-card :title="t('adminPaymentOrders.title')">
      <n-space vertical>
        <n-space>
          <n-input v-model:value="searchForm.keyword" :placeholder="t('adminPaymentOrders.searchPlaceholder')" clearable style="width: 240px" @keyup.enter="handleSearch" />
          <n-input-number v-model:value="searchForm.user_id" :placeholder="t('adminRealname.userId')" style="width: 140px" :show-button="false" />
          <n-select
            v-model:value="searchForm.status"
            :options="statusOptions"
            :placeholder="t('recharge.orderStatus')"
            style="width: 130px"
            clearable
          />
          <n-button type="primary" @click="handleSearch">{{ t('moneyScore.search') }}</n-button>
          <n-button @click="handleReset">{{ t('common.reset') }}</n-button>
        </n-space>

        <n-data-table
          :columns="columns"
          :data="orderList"
          :loading="loading"
          :pagination="pagination"
          :row-key="(row: any) => row.id"
          striped
          size="small"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>

      <!-- 详情弹窗 -->
      <n-modal v-model:show="showDetail" :title="t('recharge.orderDetail')" preset="card" style="width: 600px">
        <template v-if="detailOrder">
          <n-spin :show="detailLoading">
            <n-descriptions :column="2" bordered label-placement="left">
              <n-descriptions-item :label="t('recharge.orderNo')">{{ detailOrder.order_no }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminRealname.userId')">{{ detailOrder.user_id }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.tradeNoThirdParty')">{{ detailOrder.trade_no || '-' }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.paymentChannel')">{{ detailOrder.payment_channel }}</n-descriptions-item>
              <n-descriptions-item :label="t('recharge.paymentMethod')">{{ paymentTypeMap[detailOrder.payment_type] || detailOrder.payment_type }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminUsers.amount')">¥{{ Number(detailOrder.amount).toFixed(2) }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.orderTitle')">{{ detailOrder.subject || '-' }}</n-descriptions-item>
              <n-descriptions-item :label="t('recharge.status')">
                <n-tag :type="(statusMap[detailOrder.status] || {}).type || 'default'" size="small">
                  {{ (statusMap[detailOrder.status] || {}).label || t('recharge.unknown') }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.notifyCount')">{{ detailOrder.notify_count }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.clientIp')">{{ detailOrder.client_ip || '-' }}</n-descriptions-item>
              <n-descriptions-item :label="t('recharge.createdAt')">{{ formatTime(detailOrder.create_time) }}</n-descriptions-item>
              <n-descriptions-item :label="t('recharge.paymentTime')">{{ detailOrder.paid_at ? formatTime(detailOrder.paid_at) : '-' }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.expireAt')">{{ formatTime(detailOrder.expire_at) }}</n-descriptions-item>
            </n-descriptions>
          </n-spin>
        </template>
      </n-modal>

      <!-- 补单弹窗 -->
      <n-modal v-model:show="showComplete" :title="t('adminPaymentOrders.manualComplete')" preset="card" style="width: 450px">
        <n-alert type="warning" style="margin-bottom: 16px">
          {{ t('adminPaymentOrders.manualCompleteWarning') }}
        </n-alert>
        <template v-if="completeOrder">
          <n-descriptions :column="1" bordered label-placement="left" style="margin-bottom: 16px">
            <n-descriptions-item :label="t('recharge.orderNo')">{{ completeOrder.order_no }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminRealname.userId')">{{ completeOrder.user_id }}</n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.amount')">¥{{ Number(completeOrder.amount).toFixed(2) }}</n-descriptions-item>
          </n-descriptions>
          <n-form-item :label="t('adminPaymentOrders.completeRemark')">
            <n-input v-model:value="completeMemo" type="textarea" :placeholder="t('adminPaymentOrders.completeRemarkPlaceholder')" :rows="2" />
          </n-form-item>
        </template>
        <template #footer>
          <n-space justify="end">
            <n-button @click="showComplete = false">{{ t('common.cancel') }}</n-button>
            <n-button type="warning" :loading="submitting" @click="handleCompleteSubmit">{{ t('adminPaymentOrders.confirmComplete') }}</n-button>
          </n-space>
        </template>
      </n-modal>
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, type VNodeChild } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NTag, NSpace as NSpaceComp, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminPaymentApi } from '@/service/api/admin/payment'
import type { PaymentOrder, PaymentStats } from '@/service/api/admin/payment'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()
const loading = ref(false)
const submitting = ref(false)

// 搜索
const searchForm = reactive({
  keyword: '',
  user_id: null as number | null,
  status: null as number | null,
})

const statusOptions = [
  { label: t('recharge.allStatus'), value: -1 },
  { label: t('recharge.pending'), value: 0 },
  { label: t('recharge.paid'), value: 1 },
  { label: t('recharge.cancelled'), value: 2 },
  { label: t('recharge.refunded'), value: 3 },
  { label: t('recharge.failed'), value: 4 },
]

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

// 数据
const orderList = ref<PaymentOrder[]>([])
const stats = reactive<PaymentStats>({
  total_orders: 0,
  paid_orders: 0,
  total_amount: 0,
  today_orders: 0,
  today_amount: 0,
  pending_orders: 0,
})

// 详情弹窗
const showDetail = ref(false)
const detailLoading = ref(false)
const detailOrder = ref<PaymentOrder | null>(null)

// 补单弹窗
const showComplete = ref(false)
const completeOrder = ref<PaymentOrder | null>(null)
const completeMemo = ref('')

// 映射
const statusMap: Record<number, { label: string, type: 'default' | 'success' | 'warning' | 'error' | 'info' }> = {
  0: { label: t('recharge.pending'), type: 'warning' },
  1: { label: t('recharge.paid'), type: 'success' },
  2: { label: t('recharge.cancelled'), type: 'default' },
  3: { label: t('recharge.refunded'), type: 'info' },
  4: { label: t('recharge.failed'), type: 'error' },
}

const paymentTypeMap: Record<string, string> = {
  alipay: t('recharge.alipay'),
  wxpay: t('recharge.wechatPay'),
  qqpay: t('recharge.qqWallet'),
}

function formatTime(ts: number) {
  return ts ? new Date(ts * 1000).toLocaleString() : '-'
}

// 表格列
const columns: DataTableColumns<PaymentOrder> = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: t('recharge.orderNo'),
    key: 'order_no',
    width: 200,
    ellipsis: { tooltip: true },
  },
  { title: t('adminRealname.userId'), key: 'user_id', width: 80 },
  {
    title: t('adminUsers.amount'),
    key: 'amount',
    width: 100,
    render: row => h('span', { style: { color: '#18a058', fontWeight: '500' } }, `¥${Number(row.amount).toFixed(2)}`),
  },
  {
    title: t('recharge.paymentMethod'),
    key: 'payment_type',
    width: 90,
    render: row => paymentTypeMap[row.payment_type] || row.payment_type,
  },
  {
    title: t('recharge.status'),
    key: 'status',
    width: 90,
    render: row => {
      const s = statusMap[row.status] || { label: t('recharge.unknown'), type: 'default' as const }
      return h(NTag, { type: s.type, size: 'small', bordered: false }, () => s.label)
    },
  },
  {
    title: t('adminPaymentOrders.tradeNoThirdParty'),
    key: 'trade_no',
    width: 150,
    ellipsis: { tooltip: true },
    render: row => row.trade_no || '-',
  },
  {
    title: t('recharge.createdAt'),
    key: 'create_time',
    width: 170,
    render: row => formatTime(row.create_time),
  },
  {
    title: t('moneyScore.actions'),
    key: 'actions',
    width: 200,
    fixed: 'right',
    render: (row) => {
      const buttons: VNodeChild[] = []

      buttons.push(h(NButton, {
        size: 'small',
        type: 'info',
        text: true,
        onClick: () => handleViewDetail(row),
      }, { default: () => t('moneyScore.detail') }))

      if (row.status === 0) {
        buttons.push(h(NButton, {
          size: 'small',
          type: 'warning',
          text: true,
          onClick: () => handleComplete(row),
        }, { default: () => t('adminPaymentOrders.completeOrder') }))

        buttons.push(h(NButton, {
          size: 'small',
          type: 'default',
          text: true,
          onClick: () => handleCancel(row),
        }, { default: () => t('recharge.cancelled') }))
      }

      buttons.push(h(NButton, {
        size: 'small',
        type: 'error',
        text: true,
        onClick: () => handleDelete(row),
      }, { default: () => t('adminUsers.delete') }))

      return h(NSpaceComp, { size: 4 }, () => buttons)
    },
  },
]

// 数据加载
async function fetchData() {
  loading.value = true
  try {
    const res = await adminPaymentApi.listOrders({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword || undefined,
      user_id: searchForm.user_id || undefined,
      status: searchForm.status ?? -1,
    })
    if (res.isSuccess) {
      orderList.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    } else {
      message.error(res.message || t('adminPaymentOrders.fetchListFailed'))
    }
  } catch {
    message.error(t('adminPaymentOrders.fetchListFailed'))
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await adminPaymentApi.getStats()
    if (res.isSuccess && res.data) {
      Object.assign(stats, res.data)
    }
  } catch { /* ignore */ }
}

function handleSearch() {
  pagination.page = 1
  fetchData()
}

function handleReset() {
  searchForm.keyword = ''
  searchForm.user_id = null
  searchForm.status = null
  pagination.page = 1
  fetchData()
}

function handlePageChange(page: number) {
  pagination.page = page
  fetchData()
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchData()
}

// 详情
async function handleViewDetail(row: PaymentOrder) {
  detailOrder.value = row
  showDetail.value = true
  detailLoading.value = true
  try {
    const res = await adminPaymentApi.orderDetail(row.id)
    if (res.isSuccess && res.data) {
      detailOrder.value = res.data
    }
    else {
      message.error(res.message || t('recharge.fetchOrderDetailFailed'))
    }
  } catch {
    message.error(t('recharge.fetchOrderDetailFailed'))
  } finally {
    detailLoading.value = false
  }
}

// 补单
function handleComplete(row: PaymentOrder) {
  completeOrder.value = row
  completeMemo.value = ''
  showComplete.value = true
}

async function handleCompleteSubmit() {
  if (!completeOrder.value) return
  submitting.value = true
  try {
    const res = await adminPaymentApi.completeOrder(completeOrder.value.id, { memo: completeMemo.value })
    if (res.isSuccess) {
      message.success(t('adminPaymentOrders.completeSuccess'))
      showComplete.value = false
      fetchData()
      fetchStats()
    } else {
      message.error(res.message || t('adminPaymentOrders.completeFailed'))
    }
  } catch {
    message.error(t('adminPaymentOrders.completeFailed'))
  } finally {
    submitting.value = false
  }
}

// 取消
function handleCancel(row: PaymentOrder) {
  dialog.warning({
    title: t('adminPaymentOrders.confirmCancelTitle'),
    content: t('adminPaymentOrders.confirmCancelContent', { orderNo: row.order_no }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await adminPaymentApi.cancelOrder(row.id)
        if (res.isSuccess) {
          message.success(t('adminPaymentOrders.cancelSuccess'))
          fetchData()
          fetchStats()
        } else {
          message.error(res.message || t('adminPaymentOrders.cancelFailed'))
        }
      } catch {
        message.error(t('adminPaymentOrders.cancelFailed'))
      }
    },
  })
}

// 删除
function handleDelete(row: PaymentOrder) {
  dialog.error({
    title: t('adminUsers.delete'),
    content: t('adminPaymentOrders.confirmDeleteContent', { orderNo: row.order_no }),
    positiveText: t('adminPaymentOrders.confirmDelete'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await adminPaymentApi.deleteOrder(row.id)
        if (res.isSuccess) {
          message.success(t('adminUsers.deleteSuccess'))
          fetchData()
          fetchStats()
        } else {
          message.error(res.message || t('adminUsers.deleteFailed'))
        }
      } catch {
        message.error(t('adminUsers.deleteFailed'))
      }
    },
  })
}

onMounted(() => {
  fetchData()
  fetchStats()
})
</script>
