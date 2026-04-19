<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { h, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NTag, useDialog, useMessage } from 'naive-ui'
import NovaIcon from '@/components/common/NovaIcon.vue'
import { parseMemo } from '@/utils/memo'
import {

  adminUserApi,

  deleteUser,
  resetUserApikey,
  resetUserPassword,
  updateUserStatus,

} from '@/service/api/admin/user'
import type { AdminUser, AdminUserRealnameSummary, UserSimpleInfo } from '@/service/api/admin/user'
import { fetchAllPayOrders } from '@/service/api/admin/order'
import { fetchAllMoneyLogs, fetchAllScoreLogs, fetchWithdrawRecords } from '@/service/api/admin/finance'
import type { WithdrawRecord } from '@/service/api/admin/finance'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()

// 用户ID
const userId = ref(Number(route.params.id))

// 加载状态
const loading = ref(false)
const orderLoading = ref(false)
const moneyLoading = ref(false)
const scoreLoading = ref(false)
const withdrawLoading = ref(false)
const showWithdrawDetailModal = ref(false)

// 用户数据
const user = ref<AdminUser | null>(null)
const realname = ref<AdminUserRealnameSummary | null>(null)
const defaultAvatar = 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'
const withdrawDetail = ref<WithdrawRecord | null>(null)
const adminUserMap = ref<Record<number, UserSimpleInfo>>({})

function formatLanguage(language?: string) {
  if (!language)
    return '-'
  if (language === 'zh-CN')
    return t('adminUsersDetail.chinese')
  if (language === 'en-US')
    return t('adminUsersDetail.english')
  return language
}

function formatCurrency(value?: number | null) {
  return Number(value || 0).toFixed(2)
}

function formatRechargeRetentionRatio(currentUser?: AdminUser | null) {
  const totalPaid = Number(currentUser?.total_paid_amount || 0)
  if (totalPaid <= 0)
    return '-'
  return `${(Number(currentUser?.balance_paid_ratio || 0) * 100).toFixed(2)}%`
}

// 重置密码相关
const showResetPasswordModal = ref(false)
const newPassword = ref('')
const resettingPassword = ref(false)

// 订单数据
const orderData = ref<any[]>([])
const orderPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 余额记录数据
const moneyData = ref<any[]>([])
const moneyPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 积分记录数据
const scoreData = ref<any[]>([])
const scorePagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 提现记录数据
const withdrawData = ref<WithdrawRecord[]>([])
const withdrawPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 订单表格列
const orderColumns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: t('adminUsersDetail.orderNo'), key: 'order_no', width: 180 },
  {
    title: t('adminUsersDetail.amount'),
    key: 'amount',
    width: 100,
    render: (row: any) => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: t('adminUsersDetail.status'),
    key: 'status',
    width: 100,
    render: (row: any) => {
      const statusMap: Record<string, { type: 'default' | 'error' | 'info' | 'success' | 'warning', label: string }> = {
        0: { type: 'warning', label: t('adminUsersDetail.pendingPayment') },
        1: { type: 'success', label: t('adminUsersDetail.paid') },
        2: { type: 'default', label: t('adminUsersDetail.cancelled') },
        3: { type: 'info', label: t('adminUsersDetail.refunded') },
        4: { type: 'error', label: t('adminUsersDetail.paymentFailed') },
      }
      const status = statusMap[row.status] || { type: 'default', label: row.status }
      return h(NTag, { type: status.type }, () => status.label)
    },
  },
  { title: t('adminUsersDetail.paymentMethod'), key: 'payment_channel', width: 120 },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: any) => {
      if (!row.create_time)
        return '-'
      try {
        return new Date(row.create_time * 1000).toLocaleString()
      }
      catch {
        return row.create_time
      }
    },
  },
]

// 余额记录表格列
const moneyColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsersDetail.moneyChange'),
    key: 'money',
    width: 120,
    render: (row: any) => {
      const money = Number(row.money) || 0
      const isPositive = money > 0
      return h(
        'span',
        {
          style: {
            color: isPositive ? '#52c41a' : '#ff4d4f',
            fontWeight: '500',
          },
        },
        `${isPositive ? '+' : ''}¥${money.toFixed(2)}`,
      )
    },
  },
  {
    title: t('adminUsersDetail.beforeChange'),
    key: 'before',
    width: 120,
    render: (row: any) => `¥${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: t('adminUsersDetail.afterChange'),
    key: 'after',
    width: 120,
    render: (row: any) => `¥${(Number(row.after) || 0).toFixed(2)}`,
  },
  { title: t('adminUsersDetail.remark'), key: 'memo', ellipsis: { tooltip: true }, render: (row: any) => parseMemo(row.memo) },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: any) => {
      if (!row.create_time)
        return '-'
      try {
        return new Date(row.create_time * 1000).toLocaleString()
      }
      catch {
        return row.create_time
      }
    },
  },
]

// 积分记录表格列
const scoreColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsersDetail.scoreChange'),
    key: 'score',
    width: 120,
    render: (row: any) => {
      const score = Number(row.score) || 0
      const isPositive = score > 0
      return h(
        'span',
        {
          style: {
            color: isPositive ? '#52c41a' : '#ff4d4f',
            fontWeight: '500',
          },
        },
        `${isPositive ? '+' : ''}${score}`,
      )
    },
  },
  {
    title: t('adminUsersDetail.beforeChange'),
    key: 'before',
    width: 120,
    render: (row: any) => (Number(row.before) || 0).toString(),
  },
  {
    title: t('adminUsersDetail.afterChange'),
    key: 'after',
    width: 120,
    render: (row: any) => (Number(row.after) || 0).toString(),
  },
  { title: t('adminUsersDetail.remark'), key: 'memo', ellipsis: { tooltip: true }, render: (row: any) => parseMemo(row.memo) },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: any) => {
      if (!row.create_time)
        return '-'
      try {
        return new Date(row.create_time * 1000).toLocaleString()
      }
      catch {
        return row.create_time
      }
    },
  },
]

// 提现记录表格列
const withdrawColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsersDetail.withdrawAmount'),
    key: 'amount',
    width: 120,
    render: (row: WithdrawRecord) => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: t('adminUsersDetail.status'),
    key: 'status',
    width: 100,
    render: (row: WithdrawRecord) => {
      const status = getWithdrawStatusMeta(row.status)
      return h(NTag, { type: status.type }, () => status.label)
    },
  },
  { title: t('adminUsersDetail.withdrawMethod'), key: 'account_type', width: 100 },
  { title: t('adminUsersDetail.accountName'), key: 'account_name', width: 120, ellipsis: { tooltip: true } },
  { title: t('adminUsersDetail.accountNo'), key: 'account_no', width: 180, ellipsis: { tooltip: true }, render: (row: WithdrawRecord) => maskAccountNo(row.account_no) },
  { title: t('adminUsersDetail.payee'), key: 'real_name', width: 100 },
  { title: t('adminUsersDetail.reviewer'), key: 'reviewed_by', width: 120, render: (row: WithdrawRecord) => getAdminDisplayName(row.reviewed_by) },
  { title: t('adminUsersDetail.payer'), key: 'paid_by', width: 120, render: (row: WithdrawRecord) => getAdminDisplayName(row.paid_by) },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: WithdrawRecord) => formatTime(row.create_time),
  },
  {
    title: t('adminUsersDetail.actions'),
    key: 'actions',
    width: 100,
    render: (row: WithdrawRecord) => h(
      'a',
      {
        style: 'color: var(--n-primary-color); cursor: pointer;',
        onClick: () => openWithdrawDetail(row),
      },
      t('adminUsersDetail.detail'),
    ),
  },
]

// 获取用户信息
async function fetchUserData() {
  if (!userId.value)
    return

  loading.value = true
  try {
    const response = await adminUserApi.detail(userId.value)
    if (response.isSuccess) {
      user.value = response.data?.user || null
      realname.value = response.data?.realname || { has_verification: false }
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchUserFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch user failed', error)
    message.error(t('adminUsersDetail.fetchUserFailed'))
  }
  finally {
    loading.value = false
  }
}

function getRealnameStatusText(status?: number) {
  const map: Record<number, string> = {
    0: t('adminUsersDetail.pendingReview'),
    1: t('adminUsersDetail.approved'),
    2: t('adminUsersDetail.rejected'),
  }
  return status !== undefined ? map[status] || t('adminUsersDetail.unknown') : t('adminUsersDetail.notVerified')
}

function getRealnameStatusType(status?: number): 'default' | 'warning' | 'success' | 'error' {
  const map: Record<number, 'warning' | 'success' | 'error'> = {
    0: 'warning',
    1: 'success',
    2: 'error',
  }
  return status !== undefined ? map[status] || 'default' : 'default'
}

function maskCertificateNo(no?: string) {
  if (!no)
    return '-'
  if (no.length < 8)
    return no
  return `${no.slice(0, 4)}****${no.slice(-4)}`
}

function maskAccountNo(accountNo?: string) {
  if (!accountNo)
    return '-'
  if (accountNo.length <= 8)
    return accountNo
  return `${accountNo.slice(0, 4)}****${accountNo.slice(-4)}`
}

function formatTime(ts?: number | null) {
  return ts ? new Date(ts * 1000).toLocaleString() : '-'
}

function getWithdrawStatusMeta(status?: number): { type: 'warning' | 'info' | 'error' | 'success', label: string } {
  const map: Record<number, { type: 'warning' | 'info' | 'error' | 'success', label: string }> = {
    0: { type: 'warning', label: t('adminUsersDetail.pendingReview') },
    1: { type: 'info', label: t('adminUsersDetail.pendingTransfer') },
    2: { type: 'error', label: t('adminUsersDetail.rejected') },
    3: { type: 'success', label: t('adminUsersDetail.transferred') },
  }
  return status !== undefined ? map[status] || { type: 'info', label: t('adminUsersDetail.unknown') } : { type: 'info', label: t('adminUsersDetail.unknown') }
}

function getAdminDisplayName(adminId?: number | null) {
  if (!adminId)
    return '-'
  const admin = adminUserMap.value[adminId]
  return admin?.nickname || admin?.username || t('adminUsersDetail.adminPrefix', { id: adminId })
}

// 获取订单数据
async function fetchOrderData() {
  if (!userId.value)
    return

  orderLoading.value = true
  try {
    const response: any = await fetchAllPayOrders({
      page: orderPagination.page,
      page_size: orderPagination.pageSize,
      user_id: userId.value,
    })

    if (response.isSuccess) {
      orderData.value = response.data.list || []
      orderPagination.itemCount = response.data.total || 0
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchOrdersFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch orders failed', error)
    message.error(t('adminUsersDetail.fetchOrdersFailed'))
  }
  finally {
    orderLoading.value = false
  }
}

// 获取余额记录
async function fetchMoneyData() {
  if (!userId.value)
    return

  moneyLoading.value = true
  try {
    const response: any = await fetchAllMoneyLogs({
      page: moneyPagination.page,
      page_size: moneyPagination.pageSize,
      user_id: userId.value,
    })

    if (response.isSuccess) {
      moneyData.value = response.data.list || []
      moneyPagination.itemCount = response.data.total || 0
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchMoneyFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch money failed', error)
    message.error(t('adminUsersDetail.fetchMoneyFailed'))
  }
  finally {
    moneyLoading.value = false
  }
}

// 获取积分记录
async function fetchScoreData() {
  if (!userId.value)
    return

  scoreLoading.value = true
  try {
    const response: any = await fetchAllScoreLogs({
      page: scorePagination.page,
      page_size: scorePagination.pageSize,
      user_id: userId.value,
    })

    if (response.isSuccess) {
      scoreData.value = response.data.list || []
      scorePagination.itemCount = response.data.total || 0
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchScoreFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch score failed', error)
    message.error(t('adminUsersDetail.fetchScoreFailed'))
  }
  finally {
    scoreLoading.value = false
  }
}

// 获取提现记录
async function fetchWithdrawData() {
  if (!userId.value)
    return

  withdrawLoading.value = true
  try {
    const response: any = await fetchWithdrawRecords({
      page: withdrawPagination.page,
      page_size: withdrawPagination.pageSize,
      user_id: userId.value,
    })

    if (response.isSuccess) {
      withdrawData.value = response.data.list || []
      withdrawPagination.itemCount = response.data.total || 0
      const adminIds = Array.from(new Set(withdrawData.value.flatMap(item => [item.reviewed_by, item.paid_by]).filter(Boolean) as number[]))
      adminUserMap.value = await adminUserApi.batchSimpleInfo(adminIds)
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchWithdrawFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch withdraw failed', error)
    message.error(t('adminUsersDetail.fetchWithdrawFailed'))
  }
  finally {
    withdrawLoading.value = false
  }
}

function openWithdrawDetail(row: WithdrawRecord) {
  withdrawDetail.value = row
  showWithdrawDetailModal.value = true
}

// 处理标签页切换
function handleTabChange(tabName: string) {
  switch (tabName) {
    case 'orders':
      fetchOrderData()
      break
    case 'money':
      fetchMoneyData()
      break
    case 'score':
      fetchScoreData()
      break
    case 'withdraw':
      fetchWithdrawData()
      break
  }
}

// 处理分页变化
function handleOrderPageChange(page: number) {
  orderPagination.page = page
  fetchOrderData()
}

function handleMoneyPageChange(page: number) {
  moneyPagination.page = page
  fetchMoneyData()
}

function handleOrderPageSizeChange(pageSize: number) {
  orderPagination.pageSize = pageSize
  orderPagination.page = 1
  fetchOrderData()
}

function handleMoneyPageSizeChange(pageSize: number) {
  moneyPagination.pageSize = pageSize
  moneyPagination.page = 1
  fetchMoneyData()
}

function handleScorePageChange(page: number) {
  scorePagination.page = page
  fetchScoreData()
}

function handleScorePageSizeChange(pageSize: number) {
  scorePagination.pageSize = pageSize
  scorePagination.page = 1
  fetchScoreData()
}

function handleWithdrawPageChange(page: number) {
  withdrawPagination.page = page
  fetchWithdrawData()
}

function handleWithdrawPageSizeChange(pageSize: number) {
  withdrawPagination.pageSize = pageSize
  withdrawPagination.page = 1
  fetchWithdrawData()
}

// 返回用户列表
function handleBack() {
  const adminBasePath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
  router.push(`${adminBasePath}/user-management/users`)
}

// 刷新数据
function handleRefresh() {
  fetchUserData()
}

// 编辑用户
function handleEdit() {
  const adminBasePath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
  router.push({
    path: `${adminBasePath}/user-management/users`,
    query: { edit: userId.value },
  })
}

// 切换用户状态
function handleToggleStatus() {
  if (!user.value)
    return

  const newStatus = user.value.status === 1 ? 0 : 1
  const action = newStatus === 1 ? t('adminUsersDetail.enable') : t('adminUsersDetail.disable')

  dialog.warning({
    title: t('adminUsersDetail.confirmAction', { action }),
    content: t('adminUsersDetail.confirmActionContent', { action, username: user.value.username }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res: any = await updateUserStatus(user.value!.id, { status: newStatus })
      if (res.isSuccess) {
        message.success(t('adminUsersDetail.actionSuccess', { action }))
        fetchUserData()
      }
      else {
        message.error(res.message || t('adminUsersDetail.actionFailed', { action }))
      }
    },
  })
}

// 重置API密钥
function handleResetApikey() {
  if (!user.value)
    return

  dialog.warning({
    title: t('adminUsersDetail.confirmReset'),
    content: t('adminUsersDetail.confirmResetContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res: any = await resetUserApikey(user.value!.id)
      if (res.isSuccess) {
        message.success(t('adminUsersDetail.apiKeyResetSuccess'))
        fetchUserData()
      }
      else {
        message.error(res.message || t('adminUsersDetail.apiKeyResetFailed'))
      }
    },
  })
}

// 处理重置密码
function handleResetPassword() {
  if (!user.value)
    return

  showResetPasswordModal.value = true
  newPassword.value = ''
}

async function confirmResetPassword() {
  if (!user.value || !newPassword.value) {
    message.error(t('adminUsersDetail.enterNewPassword'))
    return
  }
  if (newPassword.value.length < 8) {
    message.error(t('adminUsersDetail.passwordMinLength'))
    return
  }

  resettingPassword.value = true
  try {
    const response: any = await resetUserPassword({
      user_id: user.value.id,
      password: newPassword.value,
    })

    if (response.isSuccess) {
      message.success(t('adminUsersDetail.passwordResetSuccess', { username: user.value.username }))
      showResetPasswordModal.value = false
    }
    else {
      message.error(response.message || t('adminUsersDetail.passwordResetFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] password reset failed', error)
    message.error(t('adminUsersDetail.passwordResetFailed'))
  }
  finally {
    resettingPassword.value = false
  }
}

function cancelResetPassword() {
  showResetPasswordModal.value = false
  newPassword.value = ''
}

// 登录为此用户
// 删除用户
function handleDelete() {
  if (!user.value)
    return

  dialog.warning({
    title: t('adminUsersDetail.confirmDelete'),
    content: t('adminUsersDetail.confirmDeleteContent', { username: user.value.username }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const response: any = await deleteUser(user.value!.id)
        if (response.isSuccess) {
          message.success(t('adminUsersDetail.deleteSuccess'))
          const adminBasePath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
          router.push(`${adminBasePath}/user-management/users`)
        }
        else {
          message.error(response.message || t('adminUsersDetail.deleteFailed'))
        }
      }
      catch {
        message.error(t('adminUsersDetail.deleteFailed'))
      }
    },
  })
}

onMounted(() => {
  fetchUserData()
})
</script>

<template>
  <div class="admin-user-detail-page">
    <!-- 页面头部 -->
    <n-card class="header-card" :bordered="false">
      <div class="header-content">
        <div class="header-title">
          <NovaIcon :size="24" class="title-icon" icon="icon-park-outline:user-info" />
          <span>{{ t('adminUsersDetail.userDetail') }}</span>
        </div>
        <n-space>
          <n-button @click="handleBack">
            <template #icon>
              <NovaIcon icon="icon-park-outline:back" />
            </template>
            {{ t('common.back') }}
          </n-button>
          <n-button @click="handleRefresh">
            <template #icon>
              <NovaIcon icon="icon-park-outline:refresh" />
            </template>
            {{ t('common.refresh') }}
          </n-button>
        </n-space>
      </div>
    </n-card>

    <!-- 用户信息卡片 -->
    <n-card class="user-info-card" :bordered="false" :loading="loading">
      <div v-if="user" class="user-info-content">
        <div class="user-avatar">
          <n-avatar :size="80" :src="user.avatar" :fallback-src="defaultAvatar">
            {{ user.username?.charAt(0).toUpperCase() }}
          </n-avatar>
        </div>
        <div class="user-details">
          <div class="user-name">
            {{ user.username }} <span class="user-id">#{{ user.id }}</span>
          </div>
          <div class="user-meta">
            <NTag :type="user.status === 1 ? 'success' : 'error'">
              {{ user.status === 1 ? t('adminUsersDetail.normal') : t('adminUsersDetail.disabled') }}
            </NTag>
            <NTag type="info">
              {{ user.role === 'admin' ? t('adminUsersDetail.admin') : t('adminUsersDetail.user') }}
            </NTag>
            <NTag type="warning">
              {{ t('adminUsersDetail.level', { level: user.level }) }}
            </NTag>
          </div>
          <div class="user-contact">
            <div v-if="user.email" class="contact-item">
              <NovaIcon icon="icon-park-outline:email" />
              <span>{{ user.email }}</span>
            </div>
            <div v-if="user.mobile" class="contact-item">
              <NovaIcon icon="icon-park-outline:phone" />
              <span>{{ user.mobile }}</span>
            </div>
          </div>
        </div>
        <div class="user-stats">
          <div class="stat-item">
            <div class="stat-value">
              ¥{{ (Number(user.money) || 0).toFixed(2) }}
            </div>
            <div class="stat-label">
              {{ t('adminUsersDetail.balance') }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-value">
              {{ Number(user.score) || 0 }}
            </div>
            <div class="stat-label">
              {{ t('adminUsersDetail.score') }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-value">
              {{ user.create_time ? new Date(user.create_time * 1000).toLocaleDateString() : '-' }}
            </div>
            <div class="stat-label">
              {{ t('adminUsersDetail.registerTime') }}
            </div>
          </div>
        </div>
      </div>
      <n-empty v-else :description="t('adminUsersDetail.userNotFound')" />
    </n-card>

    <!-- 数据标签页 -->
    <n-card class="data-tabs-card" :bordered="false">
      <n-tabs type="line" animated @update:value="handleTabChange">
        <n-tab-pane name="basic" :tab="t('adminUsersDetail.basicInfo')">
          <div class="user-info-sections">
            <!-- 基础信息 -->
            <n-card :title="t('adminUsersDetail.basicInfo')" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item :label="t('adminUsersDetail.userId')">
                  {{ user?.id }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.username')">
                  {{ user?.username }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.nickname')">
                  {{ user?.nickname || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.email')">
                  {{ user?.email || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.mobile')">
                  {{ user?.mobile || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.language')">
                  {{ formatLanguage(user?.language) }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.country')">
                  {{ user?.country || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.role')">
                  {{ user?.role === 'admin' ? t('adminUsersDetail.admin') : t('adminUsersDetail.user') }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.level')">
                  {{ user?.level || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.status')">
                  <NTag :type="user?.status === 1 ? 'success' : 'error'">
                    {{ user?.status === 1 ? t('adminUsersDetail.normal') : t('adminUsersDetail.disabled') }}
                  </NTag>
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsers.adminRemark')" :span="2">
                  {{ user?.admin_remark || '-' }}
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <!-- 资产信息 -->
            <n-card :title="t('adminUsersDetail.assetInfo')" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item :label="t('adminUsersDetail.balance')">
                  <span class="money-amount">¥{{ formatCurrency(user?.money) }}</span>
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.score')">
                  <span class="score-amount">{{ user?.score || '0' }}</span>
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsers.totalPaidAmount')">
                  {{ Number(user?.total_paid_amount || 0) > 0 ? `¥${formatCurrency(user?.total_paid_amount)}` : '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsers.rechargeRetentionRatio')">
                  {{ formatRechargeRetentionRatio(user) }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.apiKey')" :span="2">
                  <n-text code>
                    {{ user?.apikey || '-' }}
                  </n-text>
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <!-- 登录信息 -->
            <n-card :title="t('adminUsersDetail.loginInfo')" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item :label="t('adminUsersDetail.registerTime')">
                  {{ user?.create_time ? new Date(user.create_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.lastLogin')">
                  {{ user?.last_login_time ? new Date(user.last_login_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.registerIp')">
                  {{ user?.join_ip || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.lastLoginIp')">
                  {{ user?.last_login_ip || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.loginFailures')">
                  {{ user?.login_failure || '0' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.motto')" :span="1">
                  {{ user?.motto || '-' }}
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <n-card :title="t('adminUsersDetail.realnameAuth')" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item :label="t('adminUsersDetail.authStatus')">
                  <NTag :type="getRealnameStatusType(realname?.status)">
                    {{ realname?.has_verification ? getRealnameStatusText(realname?.status) : t('adminUsersDetail.notVerified') }}
                  </NTag>
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.realName')">
                  {{ realname?.real_name || '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.certificateNo')">
                  {{ maskCertificateNo(realname?.certificate_no) }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.submitTime')">
                  {{ realname?.submitted_at ? new Date(realname.submitted_at * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.reviewTime')">
                  {{ realname?.reviewed_at ? new Date(realname.reviewed_at * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item :label="t('adminUsersDetail.rejectReason')">
                  {{ realname?.reject_reason || '-' }}
                </n-descriptions-item>
              </n-descriptions>
            </n-card>
          </div>

          <div class="action-buttons">
            <n-button type="primary" @click="handleEdit">
              {{ t('adminUsersDetail.editUser') }}
            </n-button>
            <n-button type="warning" @click="handleToggleStatus">
              {{ user?.status === 1 ? t('adminUsersDetail.disableUser') : t('adminUsersDetail.enableUser') }}
            </n-button>
            <n-button type="info" @click="handleResetApikey">
              {{ t('adminUsersDetail.resetApiKey') }}
            </n-button>
            <n-button type="info" @click="handleResetPassword">
              {{ t('adminUsersDetail.resetPassword') }}
            </n-button>
            <n-button type="error" @click="handleDelete">
              {{ t('adminUsersDetail.deleteUser') }}
            </n-button>
          </div>
        </n-tab-pane>

        <n-tab-pane name="orders" :tab="t('adminUsersDetail.orderRecords')">
          <n-data-table
            :columns="orderColumns"
            :data="orderData"
            :loading="orderLoading"
            :pagination="orderPagination"
            @update:page="handleOrderPageChange"
            @update:page-size="handleOrderPageSizeChange"
          />
        </n-tab-pane>

        <n-tab-pane name="money" :tab="t('adminUsersDetail.moneyRecords')">
          <n-data-table
            :columns="moneyColumns"
            :data="moneyData"
            :loading="moneyLoading"
            :pagination="moneyPagination"
            @update:page="handleMoneyPageChange"
            @update:page-size="handleMoneyPageSizeChange"
          />
        </n-tab-pane>

        <n-tab-pane name="score" :tab="t('adminUsersDetail.scoreRecords')">
          <n-data-table
            :columns="scoreColumns"
            :data="scoreData"
            :loading="scoreLoading"
            :pagination="scorePagination"
            @update:page="handleScorePageChange"
            @update:page-size="handleScorePageSizeChange"
          />
        </n-tab-pane>

        <n-tab-pane name="withdraw" :tab="t('adminUsersDetail.withdrawRecords')">
          <n-data-table
            :columns="withdrawColumns"
            :data="withdrawData"
            :loading="withdrawLoading"
            :pagination="withdrawPagination"
            @update:page="handleWithdrawPageChange"
            @update:page-size="handleWithdrawPageSizeChange"
          />
        </n-tab-pane>
      </n-tabs>
    </n-card>
    <!-- 重置密码对话框 -->
    <n-modal
      v-model:show="showResetPasswordModal"
      preset="dialog"
      :title="t('adminUsersDetail.resetPasswordTitle')"
      :positive-text="t('common.confirm')"
      :negative-text="t('common.cancel')"
      @positive-click="confirmResetPassword"
      @negative-click="cancelResetPassword"
    >
      <template #default>
        <div style="margin-bottom: 16px;">
          <p>{{ t('adminUsersDetail.resetPasswordDesc', { username: user?.username, id: user?.id }) }}</p>
          <n-form-item :label="t('adminUsersDetail.newPassword')" required>
            <n-input v-model:value="newPassword" type="password" :placeholder="t('adminUsersDetail.newPasswordPlaceholder')" show-password-on="click" />
          </n-form-item>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showWithdrawDetailModal" preset="card" :title="t('adminUsersDetail.withdrawDetailTitle')" style="width: 620px;">
      <template v-if="withdrawDetail">
        <n-descriptions :column="1" bordered label-placement="left">
          <n-descriptions-item :label="t('adminUsersDetail.applyId')">
            {{ withdrawDetail.id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.userId')">
            {{ withdrawDetail.user_id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.withdrawAmount')">
            ¥{{ Number(withdrawDetail.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.status')">
            <NTag :type="getWithdrawStatusMeta(withdrawDetail.status).type">
              {{ getWithdrawStatusMeta(withdrawDetail.status).label }}
            </NTag>
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.withdrawMethod')">
            {{ withdrawDetail.account_type }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.accountName')">
            {{ withdrawDetail.account_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.accountNo')">
            {{ withdrawDetail.account_no }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.payee')">
            {{ withdrawDetail.real_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.userRemark')">
            {{ withdrawDetail.remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.reviewRemark')">
            {{ withdrawDetail.review_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.transferRemark')">
            {{ withdrawDetail.transfer_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.applyTime')">
            {{ formatTime(withdrawDetail.create_time) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.reviewTime')">
            {{ formatTime(withdrawDetail.reviewed_at) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.reviewer')">
            {{ getAdminDisplayName(withdrawDetail.reviewed_by) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.transferTime')">
            {{ formatTime(withdrawDetail.paid_at) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsersDetail.payer')">
            {{ getAdminDisplayName(withdrawDetail.paid_by) }}
          </n-descriptions-item>
        </n-descriptions>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.admin-user-detail-page {
  padding: 16px;
}

.header-card {
  margin-bottom: 16px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #ffffff;
}

.title-icon {
  color: #ffffff;
}

.user-info-card {
  margin-bottom: 16px;
}

.user-info-content {
  display: flex;
  align-items: center;
  gap: 24px;
}

.user-details {
  flex: 1;
}

.user-name {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 8px;
}

.user-id {
  font-size: 14px;
  font-weight: normal;
  color: #999;
  margin-left: 8px;
}

.user-meta {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.user-contact {
  display: flex;
  gap: 16px;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666;
}

.user-stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

.data-tabs-card {
  min-height: 400px;
}

.user-info-sections {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-section {
  margin-bottom: 0;
}

.money-amount {
  font-weight: bold;
  color: #18a058;
  font-size: 16px;
}

.score-amount {
  font-weight: bold;
  color: #2080f0;
  font-size: 16px;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .user-info-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .user-stats {
    width: 100%;
    justify-content: space-around;
    margin-top: 16px;
  }

  .action-buttons {
    flex-wrap: wrap;
  }
}

@media (max-width: 480px) {
  .user-meta,
  .user-contact {
    flex-wrap: wrap;
  }
}
</style>
