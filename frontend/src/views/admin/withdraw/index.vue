<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import { fetchWithdrawDetail, fetchWithdrawRecords, fetchWithdrawStats, payWithdraw, reviewWithdraw } from '@/service/api/admin/finance'
import type { WithdrawRecord, WithdrawStats } from '@/service/api/admin/finance'
import { adminUserApi } from '@/service/api/admin/user'
import type { UserSimpleInfo } from '@/service/api/admin/user'

const message = useMessage()
const { t } = useI18n()
const loading = ref(false)
const submitting = ref(false)

const list = ref<WithdrawRecord[]>([])
const currentRow = ref<WithdrawRecord | null>(null)
const showReviewModal = ref(false)
const showPayModal = ref(false)
const showDetailModal = ref(false)
const detailLoading = ref(false)
const adminUserMap = ref<Record<number, UserSimpleInfo>>({})

const searchForm = reactive({
  keyword: '',
  user_id: null as number | null,
  status: null as number | null,
})

const reviewForm = reactive({
  status: 1 as 1 | 2,
  review_remark: '',
})

const payForm = reactive({
  transfer_remark: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const statusOptions = [
  { label: t('adminWithdraw.statusPending'), value: 0 },
  { label: t('adminWithdraw.statusApproved'), value: 1 },
  { label: t('adminWithdraw.statusRejected'), value: 2 },
  { label: t('adminWithdraw.statusPaid'), value: 3 },
]

const stats = ref<WithdrawStats>({
  pending_count: 0,
  approved_count: 0,
  rejected_count: 0,
  paid_count: 0,
  paid_amount: 0,
})

function getStatusMeta(status: number) {
  const map: Record<number, { label: string, type: 'warning' | 'info' | 'error' | 'success' }> = {
    0: { label: t('adminWithdraw.statusPending'), type: 'warning' },
    1: { label: t('adminWithdraw.statusApproved'), type: 'info' },
    2: { label: t('adminWithdraw.statusRejected'), type: 'error' },
    3: { label: t('adminWithdraw.statusPaid'), type: 'success' },
  }
  return map[status] || { label: t('adminWithdraw.unknown'), type: 'info' }
}

function formatTime(ts?: number | null) {
  return ts ? new Date(ts * 1000).toLocaleString() : '-'
}

function maskAccountNo(accountNo: string) {
  if (!accountNo || accountNo.length <= 8) {
    return accountNo || '-'
  }
  return `${accountNo.slice(0, 4)}****${accountNo.slice(-4)}`
}

function getAdminDisplayName(adminId?: number | null) {
  if (!adminId) {
    return '-'
  }
  const admin = adminUserMap.value[adminId]
  return admin?.nickname || admin?.username || t('adminWithdraw.adminPrefix', { id: adminId })
}

const columns: DataTableColumns<WithdrawRecord> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('adminWithdraw.userId'), key: 'user_id', width: 80 },
  {
    title: t('adminWithdraw.withdrawAmount'),
    key: 'amount',
    width: 120,
    render: row => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: t('adminWithdraw.status'),
    key: 'status',
    width: 100,
    render: (row) => {
      const meta = getStatusMeta(row.status)
      return h(NTag, { type: meta.type, bordered: false }, () => meta.label)
    },
  },
  { title: t('adminWithdraw.accountType'), key: 'account_type', width: 100 },
  { title: t('adminWithdraw.accountName'), key: 'account_name', width: 140, ellipsis: { tooltip: true } },
  {
    title: t('adminWithdraw.accountNo'),
    key: 'account_no',
    width: 180,
    ellipsis: { tooltip: true },
    render: row => maskAccountNo(row.account_no),
  },
  { title: t('adminWithdraw.realName'), key: 'real_name', width: 100 },
  {
    title: t('adminWithdraw.reviewedAt'),
    key: 'reviewed_at',
    width: 170,
    render: row => formatTime(row.reviewed_at),
  },
  {
    title: t('adminWithdraw.paidAt'),
    key: 'paid_at',
    width: 170,
    render: row => formatTime(row.paid_at),
  },
  {
    title: t('adminWithdraw.createdAt'),
    key: 'create_time',
    width: 170,
    render: row => formatTime(row.create_time),
  },
  {
    title: t('adminWithdraw.actions'),
    key: 'actions',
    width: 220,
    render: (row) => {
      const buttons = [
        h(NButton, {
          size: 'small',
          text: true,
          type: 'info',
          onClick: async () => {
            currentRow.value = row
            showDetailModal.value = true
            detailLoading.value = true
            try {
              const res = await fetchWithdrawDetail(row.id)
              if (res.isSuccess && res.data) {
                currentRow.value = res.data
              }
              else {
                message.error(res.message || t('adminWithdraw.fetchDetailFailed'))
              }
            }
            catch {
              message.error(t('adminWithdraw.fetchDetailFailed'))
            }
            finally {
              detailLoading.value = false
            }
          },
        }, { default: () => t('adminWithdraw.detail') }),
      ]

      if (row.status === 0) {
        buttons.push(h(NButton, {
          size: 'small',
          text: true,
          type: 'primary',
          onClick: () => openReview(row),
        }, { default: () => t('adminWithdraw.review') }))
      }

      if (row.status === 1) {
        buttons.push(h(NButton, {
          size: 'small',
          text: true,
          type: 'warning',
          onClick: () => openPay(row),
        }, { default: () => t('adminWithdraw.markPaid') }))
      }

      return h('div', { style: 'display:flex;gap:8px;' }, buttons)
    },
  },
]

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'user_id', label: t('adminWithdraw.userId') },
  { key: 'amount', label: t('adminWithdraw.withdrawAmount') },
  { key: 'status', label: t('adminWithdraw.status') },
  { key: 'account_type', label: t('adminWithdraw.accountType') },
  { key: 'account_name', label: t('adminWithdraw.accountName') },
  { key: 'account_no', label: t('adminWithdraw.accountNo') },
  { key: 'real_name', label: t('adminWithdraw.realName') },
  { key: 'reviewed_at', label: t('adminWithdraw.reviewedAt') },
  { key: 'paid_at', label: t('adminWithdraw.paidAt') },
  { key: 'create_time', label: t('adminWithdraw.createdAt') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<WithdrawRecord>({
  storageKey: 'admin-withdraw-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 1180,
})

async function fetchData() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword || undefined,
      user_id: searchForm.user_id || undefined,
      status: searchForm.status ?? undefined,
    }
    const [res, statsRes] = await Promise.all([
      fetchWithdrawRecords(params),
      fetchWithdrawStats({
        keyword: params.keyword,
        user_id: params.user_id,
        status: params.status,
      }),
    ])
    if (res.isSuccess) {
      list.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
      const adminIds = Array.from(new Set(list.value.flatMap(item => [item.reviewed_by, item.paid_by]).filter(Boolean) as number[]))
      adminUserMap.value = await adminUserApi.batchSimpleInfo(adminIds)
    }
    else {
      message.error(res.message || t('adminWithdraw.fetchListFailed'))
    }
    if (statsRes.isSuccess && statsRes.data) {
      stats.value = statsRes.data
    }
  }
  catch {
    message.error(t('adminWithdraw.fetchListFailed'))
  }
  finally {
    loading.value = false
  }
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

function openReview(row: WithdrawRecord) {
  currentRow.value = row
  reviewForm.status = 1
  reviewForm.review_remark = row.review_remark || ''
  showReviewModal.value = true
}

function openPay(row: WithdrawRecord) {
  currentRow.value = row
  payForm.transfer_remark = row.transfer_remark || ''
  showPayModal.value = true
}

async function handleSubmitReview() {
  if (!currentRow.value)
    return
  if (reviewForm.review_remark.trim().length > 255) {
    message.error(t('adminWithdraw.reviewRemarkTooLong'))
    return
  }
  submitting.value = true
  try {
    const res = await reviewWithdraw(currentRow.value.id, reviewForm)
    if (res.isSuccess) {
      message.success(res.message || t('adminWithdraw.reviewSuccess'))
      showReviewModal.value = false
      fetchData()
    }
    else {
      message.error(res.message || t('adminWithdraw.reviewFailed'))
    }
  }
  catch {
    message.error(t('adminWithdraw.reviewFailed'))
  }
  finally {
    submitting.value = false
  }
}

async function handleSubmitPay() {
  if (!currentRow.value)
    return
  if (payForm.transfer_remark.trim().length > 255) {
    message.error(t('adminWithdraw.transferRemarkTooLong'))
    return
  }
  submitting.value = true
  try {
    const res = await payWithdraw(currentRow.value.id, payForm)
    if (res.isSuccess) {
      message.success(res.message || t('adminWithdraw.markPaidSuccess'))
      showPayModal.value = false
      fetchData()
    }
    else {
      message.error(res.message || t('adminWithdraw.operationFailed'))
    }
  }
  catch {
    message.error(t('adminWithdraw.operationFailed'))
  }
  finally {
    submitting.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="admin-withdraw-page">
    <n-card :title="t('adminWithdraw.title')">
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
        <n-grid :cols="5" :x-gap="12">
          <n-gi>
            <n-card size="small">
              <n-statistic :label="t('adminWithdraw.statPending')" :value="stats.pending_count" />
            </n-card>
          </n-gi>
          <n-gi>
            <n-card size="small">
              <n-statistic :label="t('adminWithdraw.statApproved')" :value="stats.approved_count" />
            </n-card>
          </n-gi>
          <n-gi>
            <n-card size="small">
              <n-statistic :label="t('adminWithdraw.statRejected')" :value="stats.rejected_count" />
            </n-card>
          </n-gi>
          <n-gi>
            <n-card size="small">
              <n-statistic :label="t('adminWithdraw.statPaidCount')" :value="stats.paid_count" />
            </n-card>
          </n-gi>
          <n-gi>
            <n-card size="small">
              <n-statistic :label="t('adminWithdraw.statPaidAmount')" :value="stats.paid_amount" :precision="2">
                <template #prefix>
                  ¥
                </template>
              </n-statistic>
            </n-card>
          </n-gi>
        </n-grid>
        <n-space>
          <n-input v-model:value="searchForm.keyword" :placeholder="t('adminWithdraw.searchPlaceholder')" clearable style="width: 240px" @keyup.enter="fetchData" />
          <n-input-number v-model:value="searchForm.user_id" :placeholder="t('adminWithdraw.userIdPlaceholder')" style="width: 140px" :show-button="false" />
          <n-select v-model:value="searchForm.status" :options="statusOptions" clearable :placeholder="t('adminWithdraw.statusPlaceholder')" style="width: 140px" />
          <NButton type="primary" @click="handleSearch">
            {{ t('adminWithdraw.search') }}
          </NButton>
          <NButton @click="handleReset">
            {{ t('adminWithdraw.reset') }}
          </NButton>
        </n-space>

        <n-data-table
          :columns="visibleColumns"
          :data="list"
          :loading="loading"
          :pagination="pagination"
          :scroll-x="tableScrollX"
          striped
          size="small"
          :row-key="(row: WithdrawRecord) => row.id"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>
    </n-card>

    <n-modal v-model:show="showDetailModal" preset="card" :title="t('adminWithdraw.detailTitle')" style="width: 620px">
      <template v-if="currentRow">
        <n-spin :show="detailLoading">
          <n-descriptions :column="1" bordered label-placement="left">
            <n-descriptions-item :label="t('adminWithdraw.applicationId')">
              {{ currentRow.id }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.userId')">
              {{ currentRow.user_id }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.withdrawAmount')">
              ¥{{ Number(currentRow.amount).toFixed(2) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.status')">
              <NTag :type="getStatusMeta(currentRow.status).type">
                {{ getStatusMeta(currentRow.status).label }}
              </NTag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.accountType')">
              {{ currentRow.account_type }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.accountName')">
              {{ currentRow.account_name }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.accountNo')">
              {{ currentRow.account_no }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.realName')">
              {{ currentRow.real_name }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.userRemark')">
              {{ currentRow.remark || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.reviewRemark')">
              {{ currentRow.review_remark || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.transferRemark')">
              {{ currentRow.transfer_remark || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.createdAt')">
              {{ formatTime(currentRow.create_time) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.reviewedAt')">
              {{ formatTime(currentRow.reviewed_at) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.reviewer')">
              {{ getAdminDisplayName(currentRow.reviewed_by) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.paidAt')">
              {{ formatTime(currentRow.paid_at) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.payer')">
              {{ getAdminDisplayName(currentRow.paid_by) }}
            </n-descriptions-item>
          </n-descriptions>
        </n-spin>
      </template>
    </n-modal>

    <n-modal v-model:show="showReviewModal" preset="card" :title="t('adminWithdraw.reviewModalTitle')" style="width: 520px" :mask-closable="!submitting">
      <template v-if="currentRow">
        <n-descriptions :column="1" bordered label-placement="left" style="margin-bottom: 16px">
          <n-descriptions-item :label="t('adminWithdraw.applicationId')">
            {{ currentRow.id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.userId')">
            {{ currentRow.user_id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.withdrawAmount')">
            ¥{{ Number(currentRow.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.accountType')">
            {{ currentRow.account_type }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.accountName')">
            {{ currentRow.account_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.accountNo')">
            {{ currentRow.account_no }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.realName')">
            {{ currentRow.real_name }}
          </n-descriptions-item>
        </n-descriptions>
        <n-form label-placement="left" label-width="80">
          <n-form-item :label="t('adminWithdraw.reviewResult')">
            <n-radio-group v-model:value="reviewForm.status">
              <n-space>
                <n-radio :value="1">
                  {{ t('adminWithdraw.approve') }}
                </n-radio>
                <n-radio :value="2">
                  {{ t('adminWithdraw.reject') }}
                </n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>
          <n-form-item :label="t('adminWithdraw.reviewRemark')">
            <n-input v-model:value="reviewForm.review_remark" type="textarea" :rows="3" maxlength="255" show-count :placeholder="t('adminWithdraw.reviewRemarkPlaceholder')" />
          </n-form-item>
        </n-form>
      </template>
      <template #footer>
        <n-space justify="end">
          <NButton :disabled="submitting" @click="showReviewModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="submitting" @click="handleSubmitReview">
            {{ t('adminWithdraw.submitReview') }}
          </NButton>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showPayModal" preset="card" :title="t('adminWithdraw.payModalTitle')" style="width: 520px" :mask-closable="!submitting">
      <template v-if="currentRow">
        <n-alert type="warning" style="margin-bottom: 16px">
          {{ t('adminWithdraw.payWarning') }}
        </n-alert>
        <n-descriptions :column="1" bordered label-placement="left" style="margin-bottom: 16px">
          <n-descriptions-item :label="t('adminWithdraw.applicationId')">
            {{ currentRow.id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.userId')">
            {{ currentRow.user_id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.withdrawAmount')">
            ¥{{ Number(currentRow.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.accountType')">
            {{ currentRow.account_type }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminWithdraw.accountNo')">
            {{ currentRow.account_no }}
          </n-descriptions-item>
        </n-descriptions>
        <n-form-item :label="t('adminWithdraw.transferRemark')">
          <n-input v-model:value="payForm.transfer_remark" type="textarea" :rows="3" maxlength="255" show-count :placeholder="t('adminWithdraw.transferRemarkPlaceholder')" />
        </n-form-item>
      </template>
      <template #footer>
        <n-space justify="end">
          <NButton :disabled="submitting" @click="showPayModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="warning" :loading="submitting" @click="handleSubmitPay">
            {{ t('adminWithdraw.confirmPaid') }}
          </NButton>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
