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
import { adminPaymentApi } from '@/service/api/admin/payment'
import type { PaymentOrder } from '@/service/api/admin/payment'
import { adminOnlineApi } from '@/service/api/admin/online'
import type { OnlineSession } from '@/service/api/admin/online'
import { adminApi } from '@/service/api/admin'
import type { WithdrawRecord } from '@/service/api/admin/finance'
import { useRequestGuard, withSubmitLock } from '@/hooks'
import WithdrawDetailModal from './components/WithdrawDetailModal.vue'
import {
  formatCurrency,
  formatLanguage,
  formatRechargeRetentionRatio,
  formatTime,
  getRealnameStatusText,
  getRealnameStatusType,
  getWithdrawStatusMeta,
  maskAccountNo,
  maskCertificateNo,
  getAdminDisplayName as resolveAdminDisplayName,
} from './utils/userDisplay'
import { setPendingUserEditId } from './utils/pendingEdit'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()

// 各 Tab 分页拉取各自独立的竞态保护（快速切页/切 Tab 时丢弃过期响应）
const orderFetchGuard = useRequestGuard()
const moneyFetchGuard = useRequestGuard()
const scoreFetchGuard = useRequestGuard()
const withdrawFetchGuard = useRequestGuard()
const sessionFetchGuard = useRequestGuard()
const userFetchGuard = useRequestGuard()
/** 危险写操作（踢下线/删用户/改状态等）防连点 */
const actionLock = ref(false)

/**
 * API Key：列表/详情只显示掩码；重置成功后的明文只在本地短暂展示一次。
 */
const showApiKey = ref(false)

function displayApiKey(key?: string | null) {
  const value = String(key || '').trim()
  if (!value)
    return '-'
  if (value.startsWith('********'))
    return value
  if (showApiKey.value)
    return value
  return `********${value.slice(-4)}`
}

function isTemporaryPlainApiKey(key?: string | null) {
  const value = String(key || '').trim()
  return Boolean(value) && !value.startsWith('********')
}

/** 返回用户列表（hash 路由用 name） */
function goUserList() {
  router.push({ name: 'admin-users' })
}

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
const withdrawDetail = ref<WithdrawRecord | null>(null)
const adminUserMap = ref<Record<number, UserSimpleInfo>>({})

/** 绑定当前页 adminUserMap 的展示名 */
function getAdminDisplayName(adminId?: number | null) {
  return resolveAdminDisplayName(adminId, adminUserMap.value)
}

// 重置密码相关
const showResetPasswordModal = ref(false)
const newPassword = ref('')
const resettingPassword = ref(false)

// 订单数据
const orderData = ref<PaymentOrder[]>([])
const orderPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 余额记录数据
const moneyData = ref<Entity.UserMoneyLog[]>([])
const moneyPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 积分记录数据
const scoreData = ref<Entity.UserScoreLog[]>([])
const scorePagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 提现记录数据
const withdrawData = ref<WithdrawRecord[]>([])
const sessionLoading = ref(false)
const sessionData = ref<OnlineSession[]>([])
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
    render: (row: PaymentOrder) => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: t('adminUsersDetail.status'),
    key: 'status',
    width: 100,
    render: (row: PaymentOrder) => {
      const statusMap: Record<number, { type: 'default' | 'error' | 'info' | 'success' | 'warning', label: string }> = {
        0: { type: 'warning', label: t('adminUsersDetail.pendingPayment') },
        1: { type: 'success', label: t('adminUsersDetail.paid') },
        2: { type: 'default', label: t('adminUsersDetail.cancelled') },
        3: { type: 'info', label: t('adminUsersDetail.refunded') },
        4: { type: 'error', label: t('adminUsersDetail.paymentFailed') },
      }
      const status = statusMap[row.status] || { type: 'default' as const, label: String(row.status) }
      return h(NTag, { type: status.type }, () => status.label)
    },
  },
  { title: t('adminUsersDetail.paymentMethod'), key: 'payment_channel', width: 120 },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: PaymentOrder) => formatTime(row.create_time),
  },
]

// 余额记录表格列
const moneyColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsersDetail.moneyChange'),
    key: 'money',
    width: 120,
    render: (row: Entity.UserMoneyLog) => {
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
    render: (row: Entity.UserMoneyLog) => `¥${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: t('adminUsersDetail.afterChange'),
    key: 'after',
    width: 120,
    render: (row: Entity.UserMoneyLog) => `¥${(Number(row.after) || 0).toFixed(2)}`,
  },
  { title: t('adminUsersDetail.remark'), key: 'memo', ellipsis: { tooltip: true }, render: (row: Entity.UserMoneyLog) => parseMemo(row.memo) },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: Entity.UserMoneyLog) => formatTime(row.create_time),
  },
]

// 积分记录表格列
const scoreColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsersDetail.scoreChange'),
    key: 'score',
    width: 120,
    render: (row: Entity.UserScoreLog) => {
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
    render: (row: Entity.UserScoreLog) => (Number(row.before) || 0).toString(),
  },
  {
    title: t('adminUsersDetail.afterChange'),
    key: 'after',
    width: 120,
    render: (row: Entity.UserScoreLog) => (Number(row.after) || 0).toString(),
  },
  { title: t('adminUsersDetail.remark'), key: 'memo', ellipsis: { tooltip: true }, render: (row: Entity.UserScoreLog) => parseMemo(row.memo) },
  {
    title: t('adminUsersDetail.createTime'),
    key: 'create_time',
    width: 180,
    render: (row: Entity.UserScoreLog) => formatTime(row.create_time),
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

const sessionColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsersDetail.device'),
    key: 'device',
    minWidth: 150,
    render: (row: OnlineSession) => row.device || '-',
  },
  { title: t('adminUsersDetail.clientType'), key: 'client_type', width: 100 },
  { title: 'IP', key: 'ip', width: 140 },
  {
    title: t('adminUsersDetail.lastSeenAt'),
    key: 'last_seen_at',
    width: 180,
    render: (row: OnlineSession) => formatTime(row.last_seen_at),
  },
  {
    title: t('adminUsersDetail.onlineStatus'),
    key: 'is_online',
    width: 100,
    render: (row: OnlineSession) => h(NTag, { type: row.is_online ? 'success' : 'default' }, () => row.is_online ? t('adminUsersDetail.online') : t('adminUsersDetail.offline')),
  },
  {
    title: t('adminUsersDetail.actions'),
    key: 'actions',
    width: 100,
    render: (row: OnlineSession) => h(
      'a',
      { style: 'color: var(--n-error-color); cursor: pointer;', onClick: () => handleKickSession(row) },
      t('adminUsersDetail.kick'),
    ),
  },
]

// 获取用户信息
async function fetchUserData() {
  if (!userId.value)
    return

  const token = userFetchGuard.begin()
  loading.value = true
  try {
    const response = await adminUserApi.detail(userId.value)
    if (!userFetchGuard.isLatest(token))
      return
    if (response.isSuccess) {
      user.value = response.data?.user || null
      realname.value = response.data?.realname || { has_verification: false }
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchUserFailed'))
    }
  }
  catch (error) {
    if (!userFetchGuard.isLatest(token))
      return
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch user failed', error)
    message.error(t('adminUsersDetail.fetchUserFailed'))
  }
  finally {
    if (userFetchGuard.isLatest(token))
      loading.value = false
  }
}

// 获取订单数据
async function fetchOrderData() {
  if (!userId.value)
    return

  const token = orderFetchGuard.begin()
  orderLoading.value = true
  try {
    const response = await adminPaymentApi.listOrders({
      page: orderPagination.page,
      page_size: orderPagination.pageSize,
      user_id: userId.value,
    })
    if (!orderFetchGuard.isLatest(token))
      return

    if (response.isSuccess) {
      orderData.value = response.data.list || []
      orderPagination.itemCount = response.data.total || 0
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchOrdersFailed'))
    }
  }
  catch (error) {
    if (!orderFetchGuard.isLatest(token))
      return
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch orders failed', error)
    message.error(t('adminUsersDetail.fetchOrdersFailed'))
  }
  finally {
    if (orderFetchGuard.isLatest(token))
      orderLoading.value = false
  }
}

// 获取余额记录
async function fetchMoneyData() {
  if (!userId.value)
    return

  const token = moneyFetchGuard.begin()
  moneyLoading.value = true
  try {
    const response = await adminApi.finance.fetchAllMoneyLogs({
      page: moneyPagination.page,
      page_size: moneyPagination.pageSize,
      user_id: userId.value,
    })
    if (!moneyFetchGuard.isLatest(token))
      return

    if (response.isSuccess) {
      moneyData.value = response.data.list || []
      moneyPagination.itemCount = response.data.total || 0
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchMoneyFailed'))
    }
  }
  catch (error) {
    if (!moneyFetchGuard.isLatest(token))
      return
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch money failed', error)
    message.error(t('adminUsersDetail.fetchMoneyFailed'))
  }
  finally {
    if (moneyFetchGuard.isLatest(token))
      moneyLoading.value = false
  }
}

// 获取积分记录
async function fetchScoreData() {
  if (!userId.value)
    return

  const token = scoreFetchGuard.begin()
  scoreLoading.value = true
  try {
    const response = await adminApi.finance.fetchAllScoreLogs({
      page: scorePagination.page,
      page_size: scorePagination.pageSize,
      user_id: userId.value,
    })
    if (!scoreFetchGuard.isLatest(token))
      return

    if (response.isSuccess) {
      scoreData.value = response.data.list || []
      scorePagination.itemCount = response.data.total || 0
    }
    else {
      message.error(response.message || t('adminUsersDetail.fetchScoreFailed'))
    }
  }
  catch (error) {
    if (!scoreFetchGuard.isLatest(token))
      return
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch score failed', error)
    message.error(t('adminUsersDetail.fetchScoreFailed'))
  }
  finally {
    if (scoreFetchGuard.isLatest(token))
      scoreLoading.value = false
  }
}

// 获取提现记录
async function fetchWithdrawData() {
  if (!userId.value)
    return

  const token = withdrawFetchGuard.begin()
  withdrawLoading.value = true
  try {
    const response = await adminApi.finance.fetchWithdrawRecords({
      page: withdrawPagination.page,
      page_size: withdrawPagination.pageSize,
      user_id: userId.value,
    })
    if (!withdrawFetchGuard.isLatest(token))
      return

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
    if (!withdrawFetchGuard.isLatest(token))
      return
    if (import.meta.env.DEV)
      console.error('[adminUsersDetail] fetch withdraw failed', error)
    message.error(t('adminUsersDetail.fetchWithdrawFailed'))
  }
  finally {
    if (withdrawFetchGuard.isLatest(token))
      withdrawLoading.value = false
  }
}

async function fetchSessionData() {
  if (!userId.value)
    return
  const token = sessionFetchGuard.begin()
  sessionLoading.value = true
  try {
    const response = await adminOnlineApi.userSessions(userId.value)
    if (!sessionFetchGuard.isLatest(token))
      return

    if (response.isSuccess)
      sessionData.value = response.data || []
    else
      message.error(response.message || t('adminUsersDetail.fetchSessionsFailed'))
  }
  catch {
    if (!sessionFetchGuard.isLatest(token))
      return
    message.error(t('adminUsersDetail.fetchSessionsFailed'))
  }
  finally {
    if (sessionFetchGuard.isLatest(token))
      sessionLoading.value = false
  }
}

function handleKickSession(session: OnlineSession) {
  if (actionLock.value)
    return
  dialog.warning({
    title: t('adminUsersDetail.kickSessionTitle'),
    content: t('adminUsersDetail.kickSessionContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(actionLock, async () => {
      const response = await adminOnlineApi.kick(session.id)
      if (response.isSuccess) {
        message.success(t('adminUsersDetail.kickSuccess'))
        fetchSessionData()
        return
      }
      message.error(response.message || t('adminUsersDetail.kickFailed'))
      return false
    }),
  })
}

function handleRevokeAllSessions() {
  if (actionLock.value)
    return
  dialog.warning({
    title: t('adminUsersDetail.revokeAllSessionsTitle'),
    content: t('adminUsersDetail.revokeAllSessionsContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(actionLock, async () => {
      const response = await adminOnlineApi.revokeAllUserSessions(userId.value)
      if (response.isSuccess) {
        message.success(t('adminUsersDetail.revokeAllSessionsSuccess'))
        fetchSessionData()
        return
      }
      message.error(response.message || t('adminUsersDetail.revokeAllSessionsFailed'))
      return false
    }),
  })
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
    case 'sessions':
      fetchSessionData()
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
  goUserList()
}

// 刷新数据
function handleRefresh() {
  fetchUserData()
}

// 编辑用户：先记下待编辑 id，再回列表（不用 ?edit=，否则 fullPath 变化会重建页面把弹窗冲掉）
function handleEdit() {
  setPendingUserEditId(userId.value)
  goUserList()
}

// 切换用户状态
function handleToggleStatus() {
  if (!user.value || actionLock.value)
    return

  const newStatus = user.value.status === 1 ? 0 : 1
  const action = newStatus === 1 ? t('adminUsersDetail.enable') : t('adminUsersDetail.disable')

  dialog.warning({
    title: t('adminUsersDetail.confirmAction', { action }),
    content: t('adminUsersDetail.confirmActionContent', { action, username: user.value.username }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(actionLock, async () => {
      const res = await updateUserStatus(user.value!.id, { status: newStatus })
      if (res.isSuccess) {
        message.success(t('adminUsersDetail.actionSuccess', { action }))
        fetchUserData()
        return
      }
      message.error(res.message || t('adminUsersDetail.actionFailed', { action }))
      return false
    }),
  })
}

// 重置API密钥
function handleResetApikey() {
  if (!user.value || actionLock.value)
    return

  dialog.warning({
    title: t('adminUsersDetail.confirmReset'),
    content: t('adminUsersDetail.confirmResetContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(actionLock, async () => {
      const res = await resetUserApikey(user.value!.id)
      if (res.isSuccess) {
        message.success(t('adminUsersDetail.apiKeyResetSuccess'))
        showApiKey.value = true
        // 重置接口的明文只此一次返回：先刷新其余字段，再用本次明文覆盖回显，
        // 避免被 fetchUserData() 拉回来的掩码值覆盖掉，导致管理员连一次都看不到完整密钥。
        await fetchUserData()
        const newPlainKey = res.data?.apikey
        if (user.value && newPlainKey)
          user.value.apikey = newPlainKey
        return
      }
      message.error(res.message || t('adminUsersDetail.apiKeyResetFailed'))
      return false
    }),
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
    return false
  }
  if (newPassword.value.length < 8) {
    message.error(t('adminUsersDetail.passwordMinLength'))
    return false
  }

  const ok = await withSubmitLock(resettingPassword, async () => {
    try {
      const response = await resetUserPassword({
        user_id: user.value!.id,
        password: newPassword.value,
      })

      if (response.isSuccess) {
        message.success(t('adminUsersDetail.passwordResetSuccess', { username: user.value!.username }))
        showResetPasswordModal.value = false
        return true
      }
      message.error(response.message || t('adminUsersDetail.passwordResetFailed'))
      return false
    }
    catch (error) {
      if (import.meta.env.DEV)
        console.error('[adminUsersDetail] password reset failed', error)
      message.error(t('adminUsersDetail.passwordResetFailed'))
      return false
    }
  })
  // 进行中忽略 / 失败时保持弹窗
  return ok === true
}

function cancelResetPassword() {
  showResetPasswordModal.value = false
  newPassword.value = ''
}

// 登录为此用户
// 删除用户
function handleDelete() {
  if (!user.value || actionLock.value)
    return

  dialog.warning({
    title: t('adminUsersDetail.confirmDelete'),
    content: t('adminUsersDetail.confirmDeleteContent', { username: user.value.username }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(actionLock, async () => {
      try {
        const response = await deleteUser(user.value!.id)
        if (response.isSuccess) {
          message.success(t('adminUsersDetail.deleteSuccess'))
          goUserList()
          return
        }
        message.error(response.message || t('adminUsersDetail.deleteFailed'))
        return false
      }
      catch {
        message.error(t('adminUsersDetail.deleteFailed'))
        return false
      }
    }),
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
          <!-- 不设外部占位图地址：加载失败/未设置头像时直接回退到用户名首字母，避免依赖外部图床 -->
          <n-avatar :size="80" :src="user.avatar || undefined">
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
                  <n-space align="center">
                    <n-text code>
                      {{ displayApiKey(user?.apikey) }}
                    </n-text>
                    <n-button
                      v-if="isTemporaryPlainApiKey(user?.apikey)"
                      size="tiny"
                      quaternary
                      @click="showApiKey = !showApiKey"
                    >
                      {{ showApiKey ? t('common.hide') : t('common.show') }}
                    </n-button>
                  </n-space>
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

        <n-tab-pane name="sessions" :tab="t('adminUsersDetail.loginDevices')">
          <n-space justify="end" style="margin-bottom: 12px;">
            <n-button size="small" @click="fetchSessionData">
              {{ t('common.refresh') }}
            </n-button>
            <n-button size="small" type="error" @click="handleRevokeAllSessions">
              {{ t('adminUsersDetail.revokeAllDevices') }}
            </n-button>
          </n-space>
          <n-data-table
            :columns="sessionColumns"
            :data="sessionData"
            :loading="sessionLoading"
            :pagination="false"
            :scroll-x="800"
            :row-key="(row: OnlineSession) => row.id"
            :row-class-name="(row: OnlineSession) => row.is_online ? 'online-session-row' : ''"
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

    <WithdrawDetailModal
      v-model:show="showWithdrawDetailModal"
      :detail="withdrawDetail"
      :admin-user-map="adminUserMap"
    />
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

:deep(.online-session-row td) {
  background: color-mix(in srgb, var(--n-success-color) 8%, transparent);
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
