<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import type { VNodeChild } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NSpace as NSpaceComp, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useRequestGuard, useTableColumnVisibility, withSubmitLock } from '@/hooks'
import { adminPaymentApi } from '@/service/api/admin/payment'
import type { PaymentOrder, PaymentStats } from '@/service/api/admin/payment'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()
const listFetchGuard = useRequestGuard()
const detailFetchGuard = useRequestGuard()
const statsFetchGuard = useRequestGuard()
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
  bank: t('recharge.bankCard'),
  jdpay: t('recharge.jdPay'),
  manual: t('recharge.manual'),
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
    render: (row) => {
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
        }, { default: () => t('adminPaymentOrders.cancelOrder') }))
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

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'order_no', label: t('recharge.orderNo') },
  { key: 'user_id', label: t('adminRealname.userId') },
  { key: 'amount', label: t('adminUsers.amount') },
  { key: 'payment_type', label: t('recharge.paymentMethod') },
  { key: 'status', label: t('recharge.status') },
  { key: 'trade_no', label: t('adminPaymentOrders.tradeNoThirdParty') },
  { key: 'create_time', label: t('recharge.createdAt') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<PaymentOrder>({
  storageKey: 'admin-payment-orders-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 1080,
})

// 数据加载
async function fetchData() {
  const token = listFetchGuard.begin()
  loading.value = true
  try {
    const res = await adminPaymentApi.listOrders({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword || undefined,
      user_id: searchForm.user_id || undefined,
      status: searchForm.status ?? -1,
    })
    if (!listFetchGuard.isLatest(token))
      return
    if (res.isSuccess) {
      orderList.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || t('adminPaymentOrders.fetchListFailed'))
    }
  }
  catch {
    if (listFetchGuard.isLatest(token))
      message.error(t('adminPaymentOrders.fetchListFailed'))
  }
  finally {
    if (listFetchGuard.isLatest(token))
      loading.value = false
  }
}

async function fetchStats() {
  const token = statsFetchGuard.begin()
  try {
    const res = await adminPaymentApi.getStats()
    if (!statsFetchGuard.isLatest(token))
      return
    if (res.isSuccess && res.data) {
      Object.assign(stats, res.data)
    }
  }
  catch { /* ignore */ }
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

// 详情（带请求序号保护：快速连点不同订单时丢弃过期响应）
async function handleViewDetail(row: PaymentOrder) {
  const token = detailFetchGuard.begin()
  detailOrder.value = row
  showDetail.value = true
  detailLoading.value = true
  try {
    const res = await adminPaymentApi.orderDetail(row.id)
    if (!detailFetchGuard.isLatest(token))
      return
    if (res.isSuccess && res.data) {
      detailOrder.value = res.data
    }
    else {
      message.error(res.message || t('recharge.fetchOrderDetailFailed'))
    }
  }
  catch {
    if (!detailFetchGuard.isLatest(token))
      return
    message.error(t('recharge.fetchOrderDetailFailed'))
  }
  finally {
    if (detailFetchGuard.isLatest(token))
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
  if (!completeOrder.value)
    return
  await withSubmitLock(submitting, async () => {
    try {
      const res = await adminPaymentApi.completeOrder(completeOrder.value!.id, { memo: completeMemo.value })
      if (res.isSuccess) {
        message.success(t('adminPaymentOrders.completeSuccess'))
        showComplete.value = false
        fetchData()
        fetchStats()
        return
      }
      message.error(res.message || t('adminPaymentOrders.completeFailed'))
    }
    catch {
      message.error(t('adminPaymentOrders.completeFailed'))
    }
  })
}

// 取消
function handleCancel(row: PaymentOrder) {
  if (submitting.value)
    return
  dialog.warning({
    title: t('adminPaymentOrders.confirmCancelTitle'),
    content: t('adminPaymentOrders.confirmCancelContent', { orderNo: row.order_no }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(submitting, async () => {
      try {
        const res = await adminPaymentApi.cancelOrder(row.id)
        if (res.isSuccess) {
          message.success(t('adminPaymentOrders.cancelSuccess'))
          fetchData()
          fetchStats()
          return
        }
        message.error(res.message || t('adminPaymentOrders.cancelFailed'))
        return false
      }
      catch {
        message.error(t('adminPaymentOrders.cancelFailed'))
        return false
      }
    }),
  })
}

// 删除
function handleDelete(row: PaymentOrder) {
  if (submitting.value)
    return
  dialog.error({
    title: t('adminUsers.delete'),
    content: t('adminPaymentOrders.confirmDeleteContent', { orderNo: row.order_no }),
    positiveText: t('adminPaymentOrders.confirmDelete'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(submitting, async () => {
      try {
        const res = await adminPaymentApi.deleteOrder(row.id)
        if (res.isSuccess) {
          message.success(t('adminUsers.deleteSuccess'))
          fetchData()
          fetchStats()
          return
        }
        message.error(res.message || t('adminUsers.deleteFailed'))
        return false
      }
      catch {
        message.error(t('adminUsers.deleteFailed'))
        return false
      }
    }),
  })
}

onMounted(() => {
  fetchData()
  fetchStats()
})
</script>

<template>
  <n-space vertical :size="16">
    <!-- 统计卡片 -->
    <n-grid :cols="4" :x-gap="12">
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminPaymentOrders.todayRevenue')" :value="stats.today_amount" :precision="2">
            <template #prefix>
              ¥
            </template>
          </n-statistic>
          <n-text depth="3" style="font-size: 12px">
            {{ t('adminPaymentOrders.orderCount', { count: stats.today_orders }) }}
          </n-text>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminPaymentOrders.totalRevenue')" :value="stats.total_amount" :precision="2">
            <template #prefix>
              ¥
            </template>
          </n-statistic>
          <n-text depth="3" style="font-size: 12px">
            {{ t('adminPaymentOrders.orderCount', { count: stats.paid_orders }) }}
          </n-text>
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
      <template #header-extra>
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
      </template>
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
          <NButton type="primary" @click="handleSearch">
            {{ t('moneyScore.search') }}
          </NButton>
          <NButton @click="handleReset">
            {{ t('common.reset') }}
          </NButton>
        </n-space>

        <n-data-table
          :columns="visibleColumns"
          :data="orderList"
          :loading="loading"
          :pagination="pagination"
          :row-key="(row: any) => row.id"
          :scroll-x="tableScrollX"
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
              <n-descriptions-item :label="t('recharge.orderNo')">
                {{ detailOrder.order_no }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminRealname.userId')">
                {{ detailOrder.user_id }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.tradeNoThirdParty')">
                {{ detailOrder.trade_no || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.paymentChannel')">
                {{ detailOrder.payment_channel }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('recharge.paymentMethod')">
                {{ paymentTypeMap[detailOrder.payment_type] || detailOrder.payment_type }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminUsers.amount')">
                ¥{{ Number(detailOrder.amount).toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.orderTitle')">
                {{ detailOrder.subject || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('recharge.status')">
                <NTag :type="(statusMap[detailOrder.status] || {}).type || 'default'" size="small">
                  {{ (statusMap[detailOrder.status] || {}).label || t('recharge.unknown') }}
                </NTag>
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.notifyCount')">
                {{ detailOrder.notify_count }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.clientIp')">
                {{ detailOrder.client_ip || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('recharge.createdAt')">
                {{ formatTime(detailOrder.create_time) }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('recharge.paymentTime')">
                {{ detailOrder.paid_at ? formatTime(detailOrder.paid_at) : '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="t('adminPaymentOrders.expireAt')">
                {{ formatTime(detailOrder.expire_at) }}
              </n-descriptions-item>
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
            <n-descriptions-item :label="t('recharge.orderNo')">
              {{ completeOrder.order_no }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminRealname.userId')">
              {{ completeOrder.user_id }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.amount')">
              ¥{{ Number(completeOrder.amount).toFixed(2) }}
            </n-descriptions-item>
          </n-descriptions>
          <n-form-item :label="t('adminPaymentOrders.completeRemark')">
            <n-input v-model:value="completeMemo" type="textarea" :placeholder="t('adminPaymentOrders.completeRemarkPlaceholder')" :rows="2" />
          </n-form-item>
        </template>
        <template #footer>
          <n-space justify="end">
            <NButton @click="showComplete = false">
              {{ t('common.cancel') }}
            </NButton>
            <NButton type="warning" :loading="submitting" @click="handleCompleteSubmit">
              {{ t('adminPaymentOrders.confirmComplete') }}
            </NButton>
          </n-space>
        </template>
      </n-modal>
    </n-card>
  </n-space>
</template>
