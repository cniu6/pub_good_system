<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import type { PaymentException } from '@/service/api/admin/payment'
import { useRequestGuard } from '@/hooks'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const listFetchGuard = useRequestGuard()
const loading = ref(false)
const list = ref<PaymentException[]>([])
const query = reactive({ page: 1, page_size: 20, status: '' as number | '', exception_type: '', order_no: '' })
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const statusOptions = [
  { label: t('common.all'), value: '' },
  { label: t('adminPaymentExceptions.statusOpen'), value: 0 },
  { label: t('adminPaymentExceptions.statusResolved'), value: 1 },
  { label: t('adminPaymentExceptions.statusIgnored'), value: 2 },
]

const typeOptions = [
  { label: t('common.all'), value: '' },
  { label: t('adminPaymentExceptions.typeSignFailed'), value: 'sign_failed' },
  { label: t('adminPaymentExceptions.typeAmountMismatch'), value: 'amount_mismatch' },
  { label: t('adminPaymentExceptions.typeBindingMismatch'), value: 'binding_mismatch' },
  { label: t('adminPaymentExceptions.typeLateCallback'), value: 'late_callback_recovered' },
  { label: t('adminPaymentExceptions.typeRemoteSaveFail'), value: 'remote_local_save_failed' },
  { label: t('adminPaymentExceptions.typeReconcilePaid'), value: 'reconcile_paid' },
  { label: t('adminPaymentExceptions.typePermanent'), value: 'permanent_rejected' },
  { label: t('adminPaymentExceptions.typeOrderMissing'), value: 'order_missing' },
  { label: t('adminPaymentExceptions.typeManual'), value: 'manual_resolve' },
]

function formatTime(ts?: number | null) {
  if (!ts)
    return '-'
  return new Date(ts * 1000).toLocaleString()
}

function statusTag(status: number) {
  if (status === 1)
    return { type: 'success' as const, label: t('adminPaymentExceptions.statusResolved') }
  if (status === 2)
    return { type: 'default' as const, label: t('adminPaymentExceptions.statusIgnored') }
  return { type: 'warning' as const, label: t('adminPaymentExceptions.statusOpen') }
}

async function fetchList() {
  const token = listFetchGuard.begin()
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: query.page,
      page_size: query.page_size,
    }
    if (query.status !== '')
      params.status = query.status
    if (query.exception_type)
      params.exception_type = query.exception_type
    if (query.order_no)
      params.order_no = query.order_no
    const res = await adminApi.payment.listExceptions(params)
    if (!listFetchGuard.isLatest(token))
      return
    if (res.isSuccess && res.data) {
      list.value = res.data.list || []
      pagination.itemCount = res.data.total || 0
      pagination.page = query.page
      pagination.pageSize = query.page_size
    }
    else {
      message.error(res.message || t('adminPaymentExceptions.fetchFailed'))
    }
  }
  catch {
    if (listFetchGuard.isLatest(token))
      message.error(t('adminPaymentExceptions.fetchFailed'))
  }
  finally {
    if (listFetchGuard.isLatest(token))
      loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  fetchList()
}

function handleReset() {
  query.page = 1
  query.page_size = 20
  query.status = ''
  query.exception_type = ''
  query.order_no = ''
  fetchList()
}

function handlePageChange(page: number) {
  query.page = page
  fetchList()
}

function handlePageSizeChange(pageSize: number) {
  query.page_size = pageSize
  query.page = 1
  fetchList()
}

function resolveRow(row: PaymentException, action: 'resolve' | 'ignore') {
  dialog.warning({
    title: action === 'resolve' ? t('adminPaymentExceptions.resolveTitle') : t('adminPaymentExceptions.ignoreTitle'),
    content: t('adminPaymentExceptions.resolveConfirm', { id: row.id, orderNo: row.order_no }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res = await adminApi.payment.resolveException(row.id, { action, remark: action })
      if (res.isSuccess) {
        message.success(t('adminPaymentExceptions.resolveSuccess'))
        fetchList()
      }
      else {
        message.error(res.message || t('adminPaymentExceptions.resolveFailed'))
      }
    },
  })
}

const columns: DataTableColumns<PaymentException> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('recharge.orderNo'), key: 'order_no', width: 200, ellipsis: { tooltip: true } },
  { title: t('adminRealname.userId'), key: 'user_id', width: 90 },
  {
    title: t('adminPaymentExceptions.exceptionType'),
    key: 'exception_type',
    width: 160,
    ellipsis: { tooltip: true },
  },
  {
    title: t('recharge.status'),
    key: 'status',
    width: 100,
    render(row) {
      const s = statusTag(row.status)
      return h(NTag, { type: s.type, size: 'small' }, { default: () => s.label })
    },
  },
  { title: t('adminPaymentExceptions.source'), key: 'source', width: 100 },
  { title: t('adminPaymentExceptions.message'), key: 'message', minWidth: 200, ellipsis: { tooltip: true } },
  {
    title: t('recharge.createdAt'),
    key: 'create_time',
    width: 170,
    render: row => formatTime(row.create_time),
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 180,
    fixed: 'right',
    render(row) {
      if (row.status !== 0)
        return '-'
      return h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { size: 'tiny', type: 'primary', onClick: () => resolveRow(row, 'resolve') }, { default: () => t('adminPaymentExceptions.resolve') }),
          h(NButton, { size: 'tiny', onClick: () => resolveRow(row, 'ignore') }, { default: () => t('adminPaymentExceptions.ignore') }),
        ],
      })
    },
  },
]

onMounted(fetchList)
</script>

<template>
  <n-card :title="t('adminPaymentExceptions.title')">
    <NSpace vertical>
      <NSpace>
        <n-input
          v-model:value="query.order_no"
          :placeholder="t('adminPaymentExceptions.orderNoPlaceholder')"
          clearable
          style="width: 220px"
          @keyup.enter="handleSearch"
        />
        <n-select
          v-model:value="query.status"
          :options="statusOptions"
          :placeholder="t('recharge.status')"
          style="width: 140px"
        />
        <n-select
          v-model:value="query.exception_type"
          :options="typeOptions"
          :placeholder="t('adminPaymentExceptions.exceptionType')"
          style="width: 200px"
        />
        <NButton type="primary" @click="handleSearch">
          {{ t('moneyScore.search') }}
        </NButton>
        <NButton @click="handleReset">
          {{ t('common.reset') }}
        </NButton>
      </NSpace>

      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: PaymentException) => row.id"
        :scroll-x="1200"
        striped
        size="small"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </NSpace>
  </n-card>
</template>
