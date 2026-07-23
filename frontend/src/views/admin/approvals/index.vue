/**
 * 财务审批工作台：待审批列表，另一管理员批准/拒绝
 */
<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import { adminApprovalApi } from '@/service/api/admin/approval'
import type { ApprovalRequestItem } from '@/service/api/admin/approval'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const list = ref<ApprovalRequestItem[]>([])

function formatTime(ts?: number | null) {
  if (!ts)
    return '-'
  return new Date(ts * 1000).toLocaleString()
}

function typeLabel(type: string) {
  if (type === 'force_payment_complete')
    return t('adminApprovals.typeForceComplete')
  if (type === 'money_adjust')
    return t('adminApprovals.typeMoneyAdjust')
  return type
}

async function fetchList() {
  loading.value = true
  try {
    const res = await adminApprovalApi.listPending()
    if (res.isSuccess && res.data)
      list.value = res.data.list || []
    else
      message.error(res.message || t('adminApprovals.fetchFailed'))
  }
  catch {
    message.error(t('adminApprovals.fetchFailed'))
  }
  finally {
    loading.value = false
  }
}

function handleApprove(row: ApprovalRequestItem) {
  dialog.warning({
    title: t('adminApprovals.approveTitle'),
    content: t('adminApprovals.approveConfirm', { id: row.id }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res = await adminApprovalApi.approve(row.id, { comment: 'ok' })
      if (res.isSuccess) {
        message.success(t('adminApprovals.approveSuccess'))
        fetchList()
      }
      else {
        message.error(res.message || t('adminApprovals.actionFailed'))
        return false
      }
    },
  })
}

function handleReject(row: ApprovalRequestItem) {
  dialog.warning({
    title: t('adminApprovals.rejectTitle'),
    content: t('adminApprovals.rejectConfirm', { id: row.id }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res = await adminApprovalApi.reject(row.id, { comment: 'reject' })
      if (res.isSuccess) {
        message.success(t('adminApprovals.rejectSuccess'))
        fetchList()
      }
      else {
        message.error(res.message || t('adminApprovals.actionFailed'))
        return false
      }
    },
  })
}

const columns: DataTableColumns<ApprovalRequestItem> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminApprovals.type'),
    key: 'type',
    render: row => h(NTag, { size: 'small', type: 'warning' }, { default: () => typeLabel(row.type) }),
  },
  { title: t('adminApprovals.requester'), key: 'requester_id', width: 110 },
  {
    title: t('adminApprovals.payload'),
    key: 'payload_json',
    ellipsis: { tooltip: true },
  },
  {
    title: t('adminApprovals.createTime'),
    key: 'create_time',
    width: 170,
    render: row => formatTime(row.create_time),
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 180,
    render: row => h(NSpace, null, {
      default: () => [
        h(NButton, { size: 'small', type: 'primary', onClick: () => handleApprove(row) }, { default: () => t('adminApprovals.approve') }),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleReject(row) }, { default: () => t('adminApprovals.reject') }),
      ],
    }),
  },
]

onMounted(fetchList)
</script>

<template>
  <n-card :title="t('adminApprovals.title')" :bordered="false">
    <n-space class="mb-12px">
      <n-button :loading="loading" @click="fetchList">
        {{ t('common.refresh') }}
      </n-button>
    </n-space>
    <n-alert type="info" class="mb-12px">
      {{ t('adminApprovals.hint') }}
    </n-alert>
    <n-alert type="warning" class="mb-12px">
      {{ t('adminApprovals.autoBizHint') }}
    </n-alert>
    <n-data-table :columns="columns" :data="list" :loading="loading" :bordered="false" :single-line="false" />
  </n-card>
</template>
