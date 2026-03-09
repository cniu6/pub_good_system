<template>
  <n-space vertical :size="16">
    <!-- 统计卡片 -->
    <n-grid :cols="4" :x-gap="12">
      <n-gi>
        <n-card size="small">
          <n-statistic label="今日收入" :value="stats.today_amount" :precision="2">
            <template #prefix>¥</template>
          </n-statistic>
          <n-text depth="3" style="font-size: 12px">{{ stats.today_orders }} 笔</n-text>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic label="总收入" :value="stats.total_amount" :precision="2">
            <template #prefix>¥</template>
          </n-statistic>
          <n-text depth="3" style="font-size: 12px">{{ stats.paid_orders }} 笔</n-text>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic label="总订单" :value="stats.total_orders" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic label="待支付" :value="stats.pending_orders" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 订单列表 -->
    <n-card title="支付订单管理">
      <n-space vertical>
        <n-space>
          <n-input v-model:value="searchForm.keyword" placeholder="搜索订单号/交易号/标题" clearable style="width: 240px" @keyup.enter="handleSearch" />
          <n-input-number v-model:value="searchForm.user_id" placeholder="用户ID" style="width: 140px" :show-button="false" />
          <n-select
            v-model:value="searchForm.status"
            :options="statusOptions"
            placeholder="订单状态"
            style="width: 130px"
            clearable
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
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
      <n-modal v-model:show="showDetail" title="订单详情" preset="card" style="width: 600px">
        <template v-if="detailOrder">
          <n-descriptions :column="2" bordered label-placement="left">
            <n-descriptions-item label="订单号">{{ detailOrder.order_no }}</n-descriptions-item>
            <n-descriptions-item label="用户ID">{{ detailOrder.user_id }}</n-descriptions-item>
            <n-descriptions-item label="第三方交易号">{{ detailOrder.trade_no || '-' }}</n-descriptions-item>
            <n-descriptions-item label="支付通道">{{ detailOrder.payment_channel }}</n-descriptions-item>
            <n-descriptions-item label="支付方式">{{ paymentTypeMap[detailOrder.payment_type] || detailOrder.payment_type }}</n-descriptions-item>
            <n-descriptions-item label="金额">¥{{ Number(detailOrder.amount).toFixed(2) }}</n-descriptions-item>
            <n-descriptions-item label="订单标题">{{ detailOrder.subject || '-' }}</n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="(statusMap[detailOrder.status] || {}).type || 'default'" size="small">
                {{ (statusMap[detailOrder.status] || {}).label || '未知' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="通知次数">{{ detailOrder.notify_count }}</n-descriptions-item>
            <n-descriptions-item label="客户端IP">{{ detailOrder.client_ip || '-' }}</n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ formatTime(detailOrder.create_time) }}</n-descriptions-item>
            <n-descriptions-item label="支付时间">{{ detailOrder.paid_at ? formatTime(detailOrder.paid_at) : '-' }}</n-descriptions-item>
            <n-descriptions-item label="过期时间">{{ formatTime(detailOrder.expire_at) }}</n-descriptions-item>
          </n-descriptions>
        </template>
      </n-modal>

      <!-- 补单弹窗 -->
      <n-modal v-model:show="showComplete" title="手动补单" preset="card" style="width: 450px">
        <n-alert type="warning" style="margin-bottom: 16px">
          手动补单将直接为用户充值对应金额，请确认订单信息无误。
        </n-alert>
        <template v-if="completeOrder">
          <n-descriptions :column="1" bordered label-placement="left" style="margin-bottom: 16px">
            <n-descriptions-item label="订单号">{{ completeOrder.order_no }}</n-descriptions-item>
            <n-descriptions-item label="用户ID">{{ completeOrder.user_id }}</n-descriptions-item>
            <n-descriptions-item label="金额">¥{{ Number(completeOrder.amount).toFixed(2) }}</n-descriptions-item>
          </n-descriptions>
          <n-form-item label="补单备注">
            <n-input v-model:value="completeMemo" type="textarea" placeholder="输入补单原因（可选）" :rows="2" />
          </n-form-item>
        </template>
        <template #footer>
          <n-space justify="end">
            <n-button @click="showComplete = false">取消</n-button>
            <n-button type="warning" :loading="submitting" @click="handleCompleteSubmit">确认补单</n-button>
          </n-space>
        </template>
      </n-modal>
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, type VNodeChild } from 'vue'
import { NButton, NTag, NSpace as NSpaceComp, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminPaymentApi } from '@/service/api/admin/payment'
import type { PaymentOrder, PaymentStats } from '@/service/api/admin/payment'

const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const submitting = ref(false)

// 搜索
const searchForm = reactive({
  keyword: '',
  user_id: null as number | null,
  status: null as number | null,
})

const statusOptions = [
  { label: '全部', value: -1 },
  { label: '待支付', value: 0 },
  { label: '已支付', value: 1 },
  { label: '已取消', value: 2 },
  { label: '已退款', value: 3 },
  { label: '支付失败', value: 4 },
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
const detailOrder = ref<PaymentOrder | null>(null)

// 补单弹窗
const showComplete = ref(false)
const completeOrder = ref<PaymentOrder | null>(null)
const completeMemo = ref('')

// 映射
const statusMap: Record<number, { label: string, type: 'default' | 'success' | 'warning' | 'error' | 'info' }> = {
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
}

function formatTime(ts: number) {
  return ts ? new Date(ts * 1000).toLocaleString() : '-'
}

// 表格列
const columns: DataTableColumns<PaymentOrder> = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '订单号',
    key: 'order_no',
    width: 200,
    ellipsis: { tooltip: true },
  },
  { title: '用户ID', key: 'user_id', width: 80 },
  {
    title: '金额',
    key: 'amount',
    width: 100,
    render: row => h('span', { style: { color: '#18a058', fontWeight: '500' } }, `¥${Number(row.amount).toFixed(2)}`),
  },
  {
    title: '支付方式',
    key: 'payment_type',
    width: 90,
    render: row => paymentTypeMap[row.payment_type] || row.payment_type,
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: row => {
      const s = statusMap[row.status] || { label: '未知', type: 'default' as const }
      return h(NTag, { type: s.type, size: 'small', bordered: false }, () => s.label)
    },
  },
  {
    title: '第三方交易号',
    key: 'trade_no',
    width: 150,
    ellipsis: { tooltip: true },
    render: row => row.trade_no || '-',
  },
  {
    title: '创建时间',
    key: 'create_time',
    width: 170,
    render: row => formatTime(row.create_time),
  },
  {
    title: '操作',
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
      }, { default: () => '详情' }))

      if (row.status === 0) {
        buttons.push(h(NButton, {
          size: 'small',
          type: 'warning',
          text: true,
          onClick: () => handleComplete(row),
        }, { default: () => '补单' }))

        buttons.push(h(NButton, {
          size: 'small',
          type: 'default',
          text: true,
          onClick: () => handleCancel(row),
        }, { default: () => '取消' }))
      }

      buttons.push(h(NButton, {
        size: 'small',
        type: 'error',
        text: true,
        onClick: () => handleDelete(row),
      }, { default: () => '删除' }))

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
      message.error(res.message || '获取订单列表失败')
    }
  } catch {
    message.error('获取订单列表失败')
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
function handleViewDetail(row: PaymentOrder) {
  detailOrder.value = row
  showDetail.value = true
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
      message.success('补单成功')
      showComplete.value = false
      fetchData()
      fetchStats()
    } else {
      message.error(res.message || '补单失败')
    }
  } catch {
    message.error('补单失败')
  } finally {
    submitting.value = false
  }
}

// 取消
function handleCancel(row: PaymentOrder) {
  dialog.warning({
    title: '确认取消',
    content: `确定取消订单 ${row.order_no}？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminPaymentApi.cancelOrder(row.id)
        if (res.isSuccess) {
          message.success('订单已取消')
          fetchData()
          fetchStats()
        } else {
          message.error(res.message || '取消失败')
        }
      } catch {
        message.error('取消失败')
      }
    },
  })
}

// 删除
function handleDelete(row: PaymentOrder) {
  dialog.error({
    title: '确认删除',
    content: `确定删除订单 ${row.order_no}？此操作不可恢复。`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminPaymentApi.deleteOrder(row.id)
        if (res.isSuccess) {
          message.success('删除成功')
          fetchData()
          fetchStats()
        } else {
          message.error(res.message || '删除失败')
        }
      } catch {
        message.error('删除失败')
      }
    },
  })
}

onMounted(() => {
  fetchData()
  fetchStats()
})
</script>
