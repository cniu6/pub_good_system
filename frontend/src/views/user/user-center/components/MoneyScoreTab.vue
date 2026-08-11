<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { createWithdrawRequest, fetchMyMoneyLogs, fetchMyScoreLogs, fetchMyWithdrawDetail, fetchMyWithdrawRecords } from '@/service/api/user/user-center'
import { useSettingsStore } from '@/store'
import { useBaseCurrency } from '@/composables/useBaseCurrency'
import { parseMemo } from '@/utils/memo'
import { useRequestGuard, withSubmitLock } from '@/hooks'

const message = useMessage()
const settingsStore = useSettingsStore()
const { t } = useI18n()
const { currencySymbol } = useBaseCurrency()
const moneyFetchGuard = useRequestGuard()
const scoreFetchGuard = useRequestGuard()
const withdrawFetchGuard = useRequestGuard()

const activeTab = ref('money')

const moneyLoading = ref(false)
const moneyKeyword = ref('')
const moneyLogs = ref<Entity.UserMoneyLog[]>([])
const moneyPagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const moneyColumns: DataTableColumns<Entity.UserMoneyLog> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('moneyScore.moneyChange'),
    key: 'money',
    width: 120,
    render: (row) => {
      const money = Number(row.money) || 0
      const isPositive = money > 0
      return h('span', {
        style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: '500' },
      }, `${isPositive ? '+' : ''}${currencySymbol.value}${money.toFixed(2)}`)
    },
  },
  {
    title: t('moneyScore.beforeChange'),
    key: 'before',
    width: 110,
    render: row => `${currencySymbol.value}${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: t('moneyScore.afterChange'),
    key: 'after',
    width: 110,
    render: row => `${currencySymbol.value}${(Number(row.after) || 0).toFixed(2)}`,
  },
  {
    title: t('moneyScore.remark'),
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: t('moneyScore.time'),
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

async function fetchMoneyLogs() {
  const token = moneyFetchGuard.begin()
  moneyLoading.value = true
  try {
    const res = await fetchMyMoneyLogs({
      page: moneyPagination.page,
      page_size: moneyPagination.pageSize,
      keyword: moneyKeyword.value || undefined,
    })
    if (!moneyFetchGuard.isLatest(token))
      return
    if (res.isSuccess) {
      moneyLogs.value = res.data?.list || []
      moneyPagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || t('moneyScore.fetchMoneyFailed'))
    }
  }
  catch {
    if (moneyFetchGuard.isLatest(token))
      message.error(t('moneyScore.fetchMoneyFailed'))
  }
  finally {
    if (moneyFetchGuard.isLatest(token))
      moneyLoading.value = false
  }
}

const scoreLoading = ref(false)
const scoreKeyword = ref('')
const scoreLogs = ref<Entity.UserScoreLog[]>([])
const scorePagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const scoreColumns: DataTableColumns<Entity.UserScoreLog> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('moneyScore.scoreChange'),
    key: 'score',
    width: 120,
    render: (row) => {
      const score = Number(row.score) || 0
      const isPositive = score > 0
      return h('span', {
        style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: '500' },
      }, `${isPositive ? '+' : ''}${score}`)
    },
  },
  {
    title: t('moneyScore.beforeChange'),
    key: 'before',
    width: 100,
    render: row => `${Number(row.before) || 0}`,
  },
  {
    title: t('moneyScore.afterChange'),
    key: 'after',
    width: 100,
    render: row => `${Number(row.after) || 0}`,
  },
  {
    title: t('moneyScore.remark'),
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: t('moneyScore.time'),
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

async function fetchScoreLogs() {
  const token = scoreFetchGuard.begin()
  scoreLoading.value = true
  try {
    const res = await fetchMyScoreLogs({
      page: scorePagination.page,
      page_size: scorePagination.pageSize,
      keyword: scoreKeyword.value || undefined,
    })
    if (!scoreFetchGuard.isLatest(token))
      return
    if (res.isSuccess) {
      scoreLogs.value = res.data?.list || []
      scorePagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || t('moneyScore.fetchScoreFailed'))
    }
  }
  catch {
    if (scoreFetchGuard.isLatest(token))
      message.error(t('moneyScore.fetchScoreFailed'))
  }
  finally {
    if (scoreFetchGuard.isLatest(token))
      scoreLoading.value = false
  }
}

const withdrawLoading = ref(false)
const withdrawSubmitting = ref(false)
const showWithdrawModal = ref(false)
const showWithdrawDetailModal = ref(false)
const withdrawDetailLoading = ref(false)
const withdrawLogs = ref<Entity.WithdrawRecord[]>([])
const currentWithdraw = ref<Entity.WithdrawRecord | null>(null)
const withdrawPagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const withdrawForm = reactive({
  amount: 0 as number | null,
  account_type: 'bank',
  account_name: '',
  account_no: '',
  real_name: '',
  remark: '',
})

const accountTypeLabelMap: Record<string, string> = {
  bank: t('moneyScore.accountTypeBank'),
  alipay: t('moneyScore.accountTypeAlipay'),
  wechat: t('moneyScore.accountTypeWechat'),
  usdt: 'USDT',
}

const withdrawEnabled = computed(() => settingsStore.withdrawEnabled)
const withdrawMinAmount = computed(() => Number(settingsStore.withdrawMinAmount) || 10)
const withdrawNotifyText = computed(() => settingsStore.withdrawNotifyText || t('moneyScore.withdrawNotifyText'))
const accountTypeOptions = computed(() =>
  settingsStore.withdrawAccountTypes.map((item) => {
    const value = String(item)
    return {
      label: accountTypeLabelMap[value] || value,
      value,
    }
  }))

watch(accountTypeOptions, (options) => {
  if (!options.length) {
    withdrawForm.account_type = ''
    return
  }
  if (!options.some(option => option.value === withdrawForm.account_type)) {
    withdrawForm.account_type = options[0].value
  }
}, { immediate: true })

function getWithdrawStatusMeta(status: number) {
  const map: Record<number, { type: 'warning' | 'success' | 'error' | 'info', label: string }> = {
    0: { type: 'warning', label: t('moneyScore.statusPending') },
    1: { type: 'info', label: t('moneyScore.statusApproved') },
    2: { type: 'error', label: t('moneyScore.statusRejected') },
    3: { type: 'success', label: t('moneyScore.statusPaid') },
  }
  return map[status] || { type: 'info', label: t('moneyScore.unknown') }
}

function getWithdrawStatusHint(status: number) {
  const map: Record<number, string> = {
    0: t('moneyScore.hintPending'),
    1: t('moneyScore.hintApproved'),
    2: t('moneyScore.hintRejected'),
    3: t('moneyScore.hintPaid'),
  }
  return map[status] || '-'
}

function maskAccountNo(accountNo?: string) {
  if (!accountNo)
    return '-'
  if (accountNo.length <= 8)
    return accountNo
  return `${accountNo.slice(0, 4)}****${accountNo.slice(-4)}`
}

const withdrawColumns: DataTableColumns<Entity.WithdrawRecord> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('moneyScore.withdrawAmount'),
    key: 'amount',
    width: 120,
    render: row => `${currencySymbol.value}${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: t('moneyScore.status'),
    key: 'status',
    width: 100,
    render: (row) => {
      const meta = getWithdrawStatusMeta(row.status)
      return h(NTag, { type: meta.type, bordered: false }, () => meta.label)
    },
  },
  {
    title: t('moneyScore.accountType'),
    key: 'account_type',
    width: 100,
    render: row => accountTypeLabelMap[row.account_type] || row.account_type,
  },
  { title: t('moneyScore.accountName'), key: 'account_name', width: 120, ellipsis: { tooltip: true } },
  {
    title: t('moneyScore.accountNo'),
    key: 'account_no',
    width: 180,
    ellipsis: { tooltip: true },
    render: row => maskAccountNo(row.account_no),
  },
  {
    title: t('moneyScore.reviewedAt'),
    key: 'reviewed_at',
    width: 170,
    render: row => row.reviewed_at ? new Date(row.reviewed_at * 1000).toLocaleString() : '-',
  },
  {
    title: t('moneyScore.paidAt'),
    key: 'paid_at',
    width: 170,
    render: row => row.paid_at ? new Date(row.paid_at * 1000).toLocaleString() : '-',
  },
  {
    title: t('moneyScore.reviewRemark'),
    key: 'review_remark',
    ellipsis: { tooltip: true },
    render: row => row.review_remark || '-',
  },
  {
    title: t('moneyScore.transferRemark'),
    key: 'transfer_remark',
    ellipsis: { tooltip: true },
    render: row => row.transfer_remark || '-',
  },
  {
    title: t('moneyScore.createdAt'),
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
  {
    title: t('moneyScore.actions'),
    key: 'actions',
    width: 90,
    render: row => h(
      'a',
      {
        style: 'color: var(--n-primary-color); cursor: pointer;',
        onClick: () => openWithdrawDetail(row.id),
      },
      t('moneyScore.detail'),
    ),
  },
]

async function fetchWithdrawLogs() {
  const token = withdrawFetchGuard.begin()
  withdrawLoading.value = true
  try {
    const res = await fetchMyWithdrawRecords({
      page: withdrawPagination.page,
      page_size: withdrawPagination.pageSize,
    })
    if (!withdrawFetchGuard.isLatest(token))
      return
    if (res.isSuccess) {
      withdrawLogs.value = res.data?.list || []
      withdrawPagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || t('moneyScore.fetchWithdrawFailed'))
    }
  }
  catch {
    if (withdrawFetchGuard.isLatest(token))
      message.error(t('moneyScore.fetchWithdrawFailed'))
  }
  finally {
    if (withdrawFetchGuard.isLatest(token))
      withdrawLoading.value = false
  }
}

async function openWithdrawDetail(id: number) {
  showWithdrawDetailModal.value = true
  currentWithdraw.value = null
  withdrawDetailLoading.value = true
  try {
    const res = await fetchMyWithdrawDetail(id)
    if (res.isSuccess && res.data) {
      currentWithdraw.value = res.data
    }
    else {
      message.error(res.message || t('moneyScore.fetchWithdrawDetailFailed'))
    }
  }
  catch {
    message.error(t('moneyScore.fetchWithdrawDetailFailed'))
  }
  finally {
    withdrawDetailLoading.value = false
  }
}

async function handleWithdrawSubmit() {
  if (withdrawSubmitting.value)
    return
  if (!withdrawEnabled.value) {
    message.error(t('moneyScore.withdrawDisabled'))
    showWithdrawModal.value = false
    return
  }
  if (!withdrawForm.amount || withdrawForm.amount <= 0) {
    message.error(t('moneyScore.enterValidWithdrawAmount'))
    return
  }
  if (withdrawForm.amount < withdrawMinAmount.value) {
    message.error(t('moneyScore.withdrawMinAmountError', { amount: withdrawMinAmount.value.toFixed(2) }))
    return
  }
  if (!withdrawForm.account_type || !accountTypeOptions.value.some(option => option.value === withdrawForm.account_type)) {
    message.error(t('moneyScore.selectValidAccountType'))
    return
  }
  if (!withdrawForm.account_name.trim() || !withdrawForm.account_no.trim() || !withdrawForm.real_name.trim()) {
    message.error(t('moneyScore.completeAccountInfo'))
    return
  }
  if (withdrawForm.account_name.trim().length > 100) {
    message.error(t('moneyScore.accountNameTooLong'))
    return
  }
  if (withdrawForm.account_no.trim().length > 128) {
    message.error(t('moneyScore.accountNoTooLong'))
    return
  }
  if (withdrawForm.real_name.trim().length > 100) {
    message.error(t('moneyScore.realNameTooLong'))
    return
  }
  if (withdrawForm.remark.trim().length > 255) {
    message.error(t('moneyScore.remarkTooLong'))
    return
  }
  // 校验后固化入参，避免闭包内 reactive 字段被 TS 判成 number | null
  const payload = {
    amount: Number(withdrawForm.amount),
    account_type: withdrawForm.account_type,
    account_name: withdrawForm.account_name.trim(),
    account_no: withdrawForm.account_no.trim(),
    real_name: withdrawForm.real_name.trim(),
    remark: withdrawForm.remark.trim(),
  }
  await withSubmitLock(withdrawSubmitting, async () => {
    try {
      const res = await createWithdrawRequest(payload)
      if (res.isSuccess) {
        message.success(res.message || t('moneyScore.withdrawSubmitted'))
        showWithdrawModal.value = false
        withdrawForm.amount = 0
        withdrawForm.account_type = accountTypeOptions.value[0]?.value || 'bank'
        withdrawForm.account_name = ''
        withdrawForm.account_no = ''
        withdrawForm.real_name = ''
        withdrawForm.remark = ''
        withdrawPagination.page = 1
        fetchWithdrawLogs()
        fetchMoneyLogs()
        return
      }
      message.error(res.message || t('moneyScore.withdrawSubmitFailed'))
    }
    catch {
      message.error(t('moneyScore.withdrawSubmitFailed'))
    }
  })
}

watch(activeTab, (val) => {
  if (val === 'money' && moneyLogs.value.length === 0)
    fetchMoneyLogs()
  if (val === 'score' && scoreLogs.value.length === 0)
    fetchScoreLogs()
  if (val === 'withdraw' && withdrawLogs.value.length === 0)
    fetchWithdrawLogs()
})

onMounted(() => fetchMoneyLogs())
</script>

<template>
  <n-tabs v-model:value="activeTab" type="line" animated>
    <n-tab-pane name="money" :tab="t('moneyScore.moneyRecords')">
      <n-space vertical>
        <n-space>
          <n-input v-model:value="moneyKeyword" :placeholder="t('moneyScore.searchRemark')" clearable style="width: 200px" @keyup.enter="fetchMoneyLogs" />
          <n-button type="primary" @click="fetchMoneyLogs">
            {{ t('moneyScore.search') }}
          </n-button>
          <n-button @click="moneyKeyword = ''; moneyPagination.page = 1; fetchMoneyLogs()">
            {{ t('common.reset') }}
          </n-button>
        </n-space>
        <n-data-table
          :columns="moneyColumns"
          :data="moneyLogs"
          :loading="moneyLoading"
          :pagination="moneyPagination"
          striped
          size="small"
          @update:page="(p: number) => { moneyPagination.page = p; fetchMoneyLogs() }"
          @update:page-size="(s: number) => { moneyPagination.pageSize = s; moneyPagination.page = 1; fetchMoneyLogs() }"
        />
      </n-space>
    </n-tab-pane>

    <n-tab-pane name="score" :tab="t('moneyScore.scoreRecords')">
      <n-space vertical>
        <n-space>
          <n-input v-model:value="scoreKeyword" :placeholder="t('moneyScore.searchRemark')" clearable style="width: 200px" @keyup.enter="fetchScoreLogs" />
          <n-button type="primary" @click="fetchScoreLogs">
            {{ t('moneyScore.search') }}
          </n-button>
          <n-button @click="scoreKeyword = ''; scorePagination.page = 1; fetchScoreLogs()">
            {{ t('common.reset') }}
          </n-button>
        </n-space>
        <n-data-table
          :columns="scoreColumns"
          :data="scoreLogs"
          :loading="scoreLoading"
          :pagination="scorePagination"
          striped
          size="small"
          @update:page="(p: number) => { scorePagination.page = p; fetchScoreLogs() }"
          @update:page-size="(s: number) => { scorePagination.pageSize = s; scorePagination.page = 1; fetchScoreLogs() }"
        />
      </n-space>
    </n-tab-pane>

    <n-tab-pane name="withdraw" :tab="t('moneyScore.withdrawRecords')">
      <n-space vertical>
        <n-space justify="space-between" style="width: 100%">
          <n-alert type="info" :show-icon="false" style="flex: 1">
            {{ withdrawNotifyText }}
          </n-alert>
          <n-button type="primary" :disabled="!withdrawEnabled || accountTypeOptions.length === 0" @click="showWithdrawModal = true">
            {{ t('moneyScore.applyWithdraw') }}
          </n-button>
        </n-space>
        <n-data-table
          :columns="withdrawColumns"
          :data="withdrawLogs"
          :loading="withdrawLoading"
          :pagination="withdrawPagination"
          striped
          size="small"
          @update:page="(p: number) => { withdrawPagination.page = p; fetchWithdrawLogs() }"
          @update:page-size="(s: number) => { withdrawPagination.pageSize = s; withdrawPagination.page = 1; fetchWithdrawLogs() }"
        />
      </n-space>
    </n-tab-pane>
  </n-tabs>

  <n-modal v-model:show="showWithdrawModal" preset="card" :title="t('moneyScore.applyWithdraw')" style="width: 520px" :mask-closable="!withdrawSubmitting">
    <n-form :model="withdrawForm" label-placement="left" label-width="90">
      <n-form-item :label="t('moneyScore.minWithdrawAmount')">
        <n-text depth="3">
          {{ currencySymbol }}{{ withdrawMinAmount.toFixed(2) }}
        </n-text>
      </n-form-item>
      <n-form-item :label="t('moneyScore.withdrawAmount')" required>
        <n-input-number v-model:value="withdrawForm.amount" :min="0" :precision="2" :step="10" style="width: 100%" :placeholder="t('moneyScore.enterWithdrawAmount')" />
      </n-form-item>
      <n-form-item :label="t('moneyScore.accountType')" required>
        <n-select v-model:value="withdrawForm.account_type" :options="accountTypeOptions" :disabled="accountTypeOptions.length === 0" />
      </n-form-item>
      <n-form-item :label="t('moneyScore.accountName')" required>
        <n-input v-model:value="withdrawForm.account_name" maxlength="100" show-count :placeholder="t('moneyScore.accountNamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('moneyScore.accountNo')" required>
        <n-input v-model:value="withdrawForm.account_no" maxlength="128" show-count :placeholder="t('moneyScore.accountNoPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('moneyScore.realName')" required>
        <n-input v-model:value="withdrawForm.real_name" maxlength="100" show-count :placeholder="t('moneyScore.realNamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('moneyScore.remark')">
        <n-input v-model:value="withdrawForm.remark" type="textarea" :rows="3" maxlength="255" show-count :placeholder="t('moneyScore.remarkPlaceholder')" />
      </n-form-item>
    </n-form>
    <template #footer>
      <n-space justify="end">
        <n-button :disabled="withdrawSubmitting" @click="showWithdrawModal = false">
          {{ t('common.cancel') }}
        </n-button>
        <n-button type="primary" :loading="withdrawSubmitting" @click="handleWithdrawSubmit">
          {{ t('moneyScore.submitApplication') }}
        </n-button>
      </n-space>
    </template>
  </n-modal>

  <n-modal v-model:show="showWithdrawDetailModal" preset="card" :title="t('moneyScore.withdrawDetail')" style="width: 620px">
    <template v-if="currentWithdraw">
      <n-spin :show="withdrawDetailLoading">
        <n-descriptions :column="1" bordered label-placement="left">
          <n-descriptions-item :label="t('moneyScore.applicationId')">
            {{ currentWithdraw.id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.withdrawAmount')">
            {{ currencySymbol }}{{ Number(currentWithdraw.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.status')">
            <NTag :type="getWithdrawStatusMeta(currentWithdraw.status).type" :bordered="false">
              {{ getWithdrawStatusMeta(currentWithdraw.status).label }}
            </NTag>
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.statusHint')">
            {{ getWithdrawStatusHint(currentWithdraw.status) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.accountType')">
            {{ accountTypeLabelMap[currentWithdraw.account_type] || currentWithdraw.account_type }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.accountName')">
            {{ currentWithdraw.account_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.accountNo')">
            {{ currentWithdraw.account_no }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.realName')">
            {{ currentWithdraw.real_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.userRemark')">
            {{ currentWithdraw.remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.reviewRemark')">
            {{ currentWithdraw.review_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.transferRemark')">
            {{ currentWithdraw.transfer_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.createdAt')">
            {{ currentWithdraw.create_time ? new Date(currentWithdraw.create_time * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.reviewedAt')">
            {{ currentWithdraw.reviewed_at ? new Date(currentWithdraw.reviewed_at * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.paidAt')">
            {{ currentWithdraw.paid_at ? new Date(currentWithdraw.paid_at * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
        </n-descriptions>
      </n-spin>
    </template>
  </n-modal>
</template>
