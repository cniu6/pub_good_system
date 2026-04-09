<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { createWithdrawRequest, fetchMyMoneyLogs, fetchMyScoreLogs, fetchMyWithdrawDetail, fetchMyWithdrawRecords } from '@/service/api/user/user-center'
import { useSettingsStore } from '@/store'
import { parseMemo } from '@/utils/memo'

const message = useMessage()
const settingsStore = useSettingsStore()

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
    title: '变动金额',
    key: 'money',
    width: 120,
    render: (row) => {
      const money = Number(row.money) || 0
      const isPositive = money > 0
      return h('span', {
        style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: '500' },
      }, `${isPositive ? '+' : ''}¥${money.toFixed(2)}`)
    },
  },
  {
    title: '变动前',
    key: 'before',
    width: 110,
    render: row => `¥${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: '变动后',
    key: 'after',
    width: 110,
    render: row => `¥${(Number(row.after) || 0).toFixed(2)}`,
  },
  {
    title: '备注',
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: '时间',
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

async function fetchMoneyLogs() {
  moneyLoading.value = true
  try {
    const res = await fetchMyMoneyLogs({
      page: moneyPagination.page,
      page_size: moneyPagination.pageSize,
      keyword: moneyKeyword.value || undefined,
    })
    if (res.isSuccess) {
      moneyLogs.value = res.data?.list || []
      moneyPagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || '获取余额记录失败')
    }
  }
  catch {
    message.error('获取余额记录失败')
  }
  finally {
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
    title: '积分变动',
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
    title: '变动前',
    key: 'before',
    width: 100,
    render: row => `${Number(row.before) || 0}`,
  },
  {
    title: '变动后',
    key: 'after',
    width: 100,
    render: row => `${Number(row.after) || 0}`,
  },
  {
    title: '备注',
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: '时间',
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

async function fetchScoreLogs() {
  scoreLoading.value = true
  try {
    const res = await fetchMyScoreLogs({
      page: scorePagination.page,
      page_size: scorePagination.pageSize,
      keyword: scoreKeyword.value || undefined,
    })
    if (res.isSuccess) {
      scoreLogs.value = res.data?.list || []
      scorePagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || '获取积分记录失败')
    }
  }
  catch {
    message.error('获取积分记录失败')
  }
  finally {
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
  bank: '银行卡',
  alipay: '支付宝',
  wechat: '微信',
  usdt: 'USDT',
}

const withdrawEnabled = computed(() => settingsStore.withdrawEnabled)
const withdrawMinAmount = computed(() => Number(settingsStore.withdrawMinAmount) || 10)
const withdrawNotifyText = computed(() => settingsStore.withdrawNotifyText || '提现申请提交后需管理员审核，通过后人工打款。')
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
    0: { type: 'warning', label: '待审核' },
    1: { type: 'info', label: '待打款' },
    2: { type: 'error', label: '已拒绝' },
    3: { type: 'success', label: '已打款' },
  }
  return map[status] || { type: 'info', label: '未知' }
}

function getWithdrawStatusHint(status: number) {
  const map: Record<number, string> = {
    0: '已提交，等待管理员审核',
    1: '审核已通过，等待管理员人工打款',
    2: '申请已被拒绝，请查看审核备注',
    3: '管理员已完成打款，请留意到账情况',
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
    title: '提现金额',
    key: 'amount',
    width: 120,
    render: row => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const meta = getWithdrawStatusMeta(row.status)
      return h(NTag, { type: meta.type, bordered: false }, () => meta.label)
    },
  },
  {
    title: '收款方式',
    key: 'account_type',
    width: 100,
    render: row => accountTypeLabelMap[row.account_type] || row.account_type,
  },
  { title: '账户名称', key: 'account_name', width: 120, ellipsis: { tooltip: true } },
  {
    title: '收款账号',
    key: 'account_no',
    width: 180,
    ellipsis: { tooltip: true },
    render: row => maskAccountNo(row.account_no),
  },
  {
    title: '审核时间',
    key: 'reviewed_at',
    width: 170,
    render: row => row.reviewed_at ? new Date(row.reviewed_at * 1000).toLocaleString() : '-',
  },
  {
    title: '打款时间',
    key: 'paid_at',
    width: 170,
    render: row => row.paid_at ? new Date(row.paid_at * 1000).toLocaleString() : '-',
  },
  {
    title: '审核备注',
    key: 'review_remark',
    ellipsis: { tooltip: true },
    render: row => row.review_remark || '-',
  },
  {
    title: '打款备注',
    key: 'transfer_remark',
    ellipsis: { tooltip: true },
    render: row => row.transfer_remark || '-',
  },
  {
    title: '申请时间',
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    render: row => h(
      'a',
      {
        style: 'color: var(--n-primary-color); cursor: pointer;',
        onClick: () => openWithdrawDetail(row.id),
      },
      '详情',
    ),
  },
]

async function fetchWithdrawLogs() {
  withdrawLoading.value = true
  try {
    const res = await fetchMyWithdrawRecords({
      page: withdrawPagination.page,
      page_size: withdrawPagination.pageSize,
    })
    if (res.isSuccess) {
      withdrawLogs.value = res.data?.list || []
      withdrawPagination.itemCount = res.data?.total || 0
    }
    else {
      message.error(res.message || '获取提现记录失败')
    }
  }
  catch {
    message.error('获取提现记录失败')
  }
  finally {
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
      message.error(res.message || '获取提现详情失败')
    }
  }
  catch {
    message.error('获取提现详情失败')
  }
  finally {
    withdrawDetailLoading.value = false
  }
}

async function handleWithdrawSubmit() {
  if (!withdrawEnabled.value) {
    message.error('提现功能暂未开启')
    showWithdrawModal.value = false
    return
  }
  if (!withdrawForm.amount || withdrawForm.amount <= 0) {
    message.error('请输入正确的提现金额')
    return
  }
  if (withdrawForm.amount < withdrawMinAmount.value) {
    message.error(`提现金额不能低于 ¥${withdrawMinAmount.value.toFixed(2)}`)
    return
  }
  if (!withdrawForm.account_type || !accountTypeOptions.value.some(option => option.value === withdrawForm.account_type)) {
    message.error('请选择有效的收款方式')
    return
  }
  if (!withdrawForm.account_name.trim() || !withdrawForm.account_no.trim() || !withdrawForm.real_name.trim()) {
    message.error('请完整填写收款信息')
    return
  }
  if (withdrawForm.account_name.trim().length > 100) {
    message.error('账户名称不能超过100个字符')
    return
  }
  if (withdrawForm.account_no.trim().length > 128) {
    message.error('收款账号不能超过128个字符')
    return
  }
  if (withdrawForm.real_name.trim().length > 100) {
    message.error('收款人不能超过100个字符')
    return
  }
  if (withdrawForm.remark.trim().length > 255) {
    message.error('备注不能超过255个字符')
    return
  }
  withdrawSubmitting.value = true
  try {
    const res = await createWithdrawRequest({
      amount: withdrawForm.amount,
      account_type: withdrawForm.account_type,
      account_name: withdrawForm.account_name.trim(),
      account_no: withdrawForm.account_no.trim(),
      real_name: withdrawForm.real_name.trim(),
      remark: withdrawForm.remark.trim(),
    })
    if (res.isSuccess) {
      message.success(res.message || '提现申请已提交')
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
    }
    else {
      message.error(res.message || '提现申请失败')
    }
  }
  catch {
    message.error('提现申请失败')
  }
  finally {
    withdrawSubmitting.value = false
  }
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
    <n-tab-pane name="money" tab="余额记录">
      <n-space vertical>
        <n-space>
          <n-input v-model:value="moneyKeyword" placeholder="搜索备注" clearable style="width: 200px" @keyup.enter="fetchMoneyLogs" />
          <n-button type="primary" @click="fetchMoneyLogs">
            搜索
          </n-button>
          <n-button @click="moneyKeyword = ''; moneyPagination.page = 1; fetchMoneyLogs()">
            重置
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

    <n-tab-pane name="score" tab="积分记录">
      <n-space vertical>
        <n-space>
          <n-input v-model:value="scoreKeyword" placeholder="搜索备注" clearable style="width: 200px" @keyup.enter="fetchScoreLogs" />
          <n-button type="primary" @click="fetchScoreLogs">
            搜索
          </n-button>
          <n-button @click="scoreKeyword = ''; scorePagination.page = 1; fetchScoreLogs()">
            重置
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

    <n-tab-pane name="withdraw" tab="提现记录">
      <n-space vertical>
        <n-space justify="space-between" style="width: 100%">
          <n-alert type="info" :show-icon="false" style="flex: 1">
            {{ withdrawNotifyText }}
          </n-alert>
          <n-button type="primary" :disabled="!withdrawEnabled || accountTypeOptions.length === 0" @click="showWithdrawModal = true">
            申请提现
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

  <n-modal v-model:show="showWithdrawModal" preset="card" title="申请提现" style="width: 520px" :mask-closable="!withdrawSubmitting">
    <n-form :model="withdrawForm" label-placement="left" label-width="90">
      <n-form-item label="最低提现金额">
        <n-text depth="3">
          ¥{{ withdrawMinAmount.toFixed(2) }}
        </n-text>
      </n-form-item>
      <n-form-item label="提现金额" required>
        <n-input-number v-model:value="withdrawForm.amount" :min="0" :precision="2" :step="10" style="width: 100%" placeholder="请输入提现金额" />
      </n-form-item>
      <n-form-item label="收款方式" required>
        <n-select v-model:value="withdrawForm.account_type" :options="accountTypeOptions" :disabled="accountTypeOptions.length === 0" />
      </n-form-item>
      <n-form-item label="账户名称" required>
        <n-input v-model:value="withdrawForm.account_name" maxlength="100" show-count placeholder="如：招商银行 / 支付宝" />
      </n-form-item>
      <n-form-item label="收款账号" required>
        <n-input v-model:value="withdrawForm.account_no" maxlength="128" show-count placeholder="请输入银行卡号/支付宝账号等" />
      </n-form-item>
      <n-form-item label="收款人" required>
        <n-input v-model:value="withdrawForm.real_name" maxlength="100" show-count placeholder="请输入收款人姓名" />
      </n-form-item>
      <n-form-item label="备注">
        <n-input v-model:value="withdrawForm.remark" type="textarea" :rows="3" maxlength="255" show-count placeholder="可填写补充说明" />
      </n-form-item>
    </n-form>
    <template #footer>
      <n-space justify="end">
        <n-button :disabled="withdrawSubmitting" @click="showWithdrawModal = false">
          取消
        </n-button>
        <n-button type="primary" :loading="withdrawSubmitting" @click="handleWithdrawSubmit">
          提交申请
        </n-button>
      </n-space>
    </template>
  </n-modal>

  <n-modal v-model:show="showWithdrawDetailModal" preset="card" title="提现详情" style="width: 620px">
    <template v-if="currentWithdraw">
      <n-spin :show="withdrawDetailLoading">
        <n-descriptions :column="1" bordered label-placement="left">
          <n-descriptions-item label="申请ID">
            {{ currentWithdraw.id }}
          </n-descriptions-item>
          <n-descriptions-item label="提现金额">
            ¥{{ Number(currentWithdraw.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <NTag :type="getWithdrawStatusMeta(currentWithdraw.status).type" :bordered="false">
              {{ getWithdrawStatusMeta(currentWithdraw.status).label }}
            </NTag>
          </n-descriptions-item>
          <n-descriptions-item label="状态说明">
            {{ getWithdrawStatusHint(currentWithdraw.status) }}
          </n-descriptions-item>
          <n-descriptions-item label="收款方式">
            {{ accountTypeLabelMap[currentWithdraw.account_type] || currentWithdraw.account_type }}
          </n-descriptions-item>
          <n-descriptions-item label="账户名称">
            {{ currentWithdraw.account_name }}
          </n-descriptions-item>
          <n-descriptions-item label="收款账号">
            {{ currentWithdraw.account_no }}
          </n-descriptions-item>
          <n-descriptions-item label="收款人">
            {{ currentWithdraw.real_name }}
          </n-descriptions-item>
          <n-descriptions-item label="用户备注">
            {{ currentWithdraw.remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="审核备注">
            {{ currentWithdraw.review_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="打款备注">
            {{ currentWithdraw.transfer_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="申请时间">
            {{ currentWithdraw.create_time ? new Date(currentWithdraw.create_time * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="审核时间">
            {{ currentWithdraw.reviewed_at ? new Date(currentWithdraw.reviewed_at * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="打款时间">
            {{ currentWithdraw.paid_at ? new Date(currentWithdraw.paid_at * 1000).toLocaleString() : '-' }}
          </n-descriptions-item>
        </n-descriptions>
      </n-spin>
    </template>
  </n-modal>
</template>
