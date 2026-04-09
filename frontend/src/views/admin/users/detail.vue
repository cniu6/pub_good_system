<script setup lang="ts">
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
    return '中文'
  if (language === 'en-US')
    return 'English'
  return language
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
  { title: '订单号', key: 'order_no', width: 180 },
  {
    title: '金额',
    key: 'amount',
    width: 100,
    render: (row: any) => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => {
      const statusMap: Record<string, { type: 'default' | 'error' | 'info' | 'success' | 'warning', label: string }> = {
        0: { type: 'warning', label: '待支付' },
        1: { type: 'success', label: '已支付' },
        2: { type: 'default', label: '已取消' },
        3: { type: 'info', label: '已退款' },
        4: { type: 'error', label: '支付失败' },
      }
      const status = statusMap[row.status] || { type: 'default', label: row.status }
      return h(NTag, { type: status.type }, () => status.label)
    },
  },
  { title: '支付方式', key: 'payment_channel', width: 120 },
  {
    title: '创建时间',
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
    title: '变动金额',
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
    title: '变动前',
    key: 'before',
    width: 120,
    render: (row: any) => `¥${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: '变动后',
    key: 'after',
    width: 120,
    render: (row: any) => `¥${(Number(row.after) || 0).toFixed(2)}`,
  },
  { title: '备注', key: 'memo', ellipsis: { tooltip: true }, render: (row: any) => parseMemo(row.memo) },
  {
    title: '创建时间',
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
    title: '变动积分',
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
    title: '变动前',
    key: 'before',
    width: 120,
    render: (row: any) => (Number(row.before) || 0).toString(),
  },
  {
    title: '变动后',
    key: 'after',
    width: 120,
    render: (row: any) => (Number(row.after) || 0).toString(),
  },
  { title: '备注', key: 'memo', ellipsis: { tooltip: true }, render: (row: any) => parseMemo(row.memo) },
  {
    title: '创建时间',
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
    title: '提现金额',
    key: 'amount',
    width: 120,
    render: (row: WithdrawRecord) => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: WithdrawRecord) => {
      const status = getWithdrawStatusMeta(row.status)
      return h(NTag, { type: status.type }, () => status.label)
    },
  },
  { title: '提现方式', key: 'account_type', width: 100 },
  { title: '账户名称', key: 'account_name', width: 120, ellipsis: { tooltip: true } },
  { title: '收款账号', key: 'account_no', width: 180, ellipsis: { tooltip: true }, render: (row: WithdrawRecord) => maskAccountNo(row.account_no) },
  { title: '收款人', key: 'real_name', width: 100 },
  { title: '审核人', key: 'reviewed_by', width: 120, render: (row: WithdrawRecord) => getAdminDisplayName(row.reviewed_by) },
  { title: '打款人', key: 'paid_by', width: 120, render: (row: WithdrawRecord) => getAdminDisplayName(row.paid_by) },
  {
    title: '创建时间',
    key: 'create_time',
    width: 180,
    render: (row: WithdrawRecord) => formatTime(row.create_time),
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row: WithdrawRecord) => h(
      'a',
      {
        style: 'color: var(--n-primary-color); cursor: pointer;',
        onClick: () => openWithdrawDetail(row),
      },
      '详情',
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
      message.error(response.message || '获取用户信息失败')
    }
  }
  catch (error) {
    console.error('获取用户信息失败:', error)
    message.error('获取用户信息失败')
  }
  finally {
    loading.value = false
  }
}

function getRealnameStatusText(status?: number) {
  const map: Record<number, string> = {
    0: '待审核',
    1: '已通过',
    2: '已拒绝',
  }
  return status !== undefined ? map[status] || '未知' : '未认证'
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
    0: { type: 'warning', label: '待审核' },
    1: { type: 'info', label: '待打款' },
    2: { type: 'error', label: '已拒绝' },
    3: { type: 'success', label: '已打款' },
  }
  return status !== undefined ? map[status] || { type: 'info', label: '未知' } : { type: 'info', label: '未知' }
}

function getAdminDisplayName(adminId?: number | null) {
  if (!adminId)
    return '-'
  const admin = adminUserMap.value[adminId]
  return admin?.nickname || admin?.username || `管理员#${adminId}`
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
      message.error(response.message || '获取订单记录失败')
    }
  }
  catch (error) {
    console.error('获取订单记录失败:', error)
    message.error('获取订单记录失败')
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
      message.error(response.message || '获取余额记录失败')
    }
  }
  catch (error) {
    console.error('获取余额记录失败:', error)
    message.error('获取余额记录失败')
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
      message.error(response.message || '获取积分记录失败')
    }
  }
  catch (error) {
    console.error('获取积分记录失败:', error)
    message.error('获取积分记录失败')
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
      message.error(response.message || '获取提现记录失败')
    }
  }
  catch (error) {
    console.error('获取提现记录失败:', error)
    message.error('获取提现记录失败')
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
  const action = newStatus === 1 ? '启用' : '禁用'

  dialog.warning({
    title: `确认${action}`,
    content: `确定要${action}用户 "${user.value.username}" 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res: any = await updateUserStatus(user.value!.id, { status: newStatus })
      if (res.isSuccess) {
        message.success(`${action}成功`)
        fetchUserData()
      }
      else {
        message.error(res.message || `${action}失败`)
      }
    },
  })
}

// 重置API密钥
function handleResetApikey() {
  if (!user.value)
    return

  dialog.warning({
    title: '确认重置',
    content: '确定要重置用户的API密钥吗？重置后旧的密钥将失效。',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res: any = await resetUserApikey(user.value!.id)
      if (res.isSuccess) {
        message.success('API密钥重置成功')
        fetchUserData()
      }
      else {
        message.error(res.message || 'API密钥重置失败')
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
    message.error('请输入新密码')
    return
  }
  if (newPassword.value.length < 6) {
    message.error('密码长度不能少于6个字符')
    return
  }

  resettingPassword.value = true
  try {
    const response: any = await resetUserPassword({
      user_id: user.value.id,
      password: newPassword.value,
    })

    if (response.isSuccess) {
      message.success(`用户 ${user.value.username} 的密码已重置`)
      showResetPasswordModal.value = false
    }
    else {
      message.error(response.message || '密码重置失败')
    }
  }
  catch (error) {
    console.error('密码重置失败:', error)
    message.error('密码重置失败')
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
    title: '确认删除',
    content: `确定要删除用户 "${user.value.username}" 吗？此操作不可恢复。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response: any = await deleteUser(user.value!.id)
        if (response.isSuccess) {
          message.success('删除成功')
          const adminBasePath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
          router.push(`${adminBasePath}/user-management/users`)
        }
        else {
          message.error(response.message || '删除失败')
        }
      }
      catch {
        message.error('删除失败')
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
          <span>用户详情</span>
        </div>
        <n-space>
          <n-button @click="handleBack">
            <template #icon>
              <NovaIcon icon="icon-park-outline:back" />
            </template>
            返回
          </n-button>
          <n-button @click="handleRefresh">
            <template #icon>
              <NovaIcon icon="icon-park-outline:refresh" />
            </template>
            刷新
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
              {{ user.status === 1 ? '正常' : '禁用' }}
            </NTag>
            <NTag type="info">
              {{ user.role === 'admin' ? '管理员' : '用户' }}
            </NTag>
            <NTag type="warning">
              等级 {{ user.level }}
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
              余额
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-value">
              {{ Number(user.score) || 0 }}
            </div>
            <div class="stat-label">
              积分
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-value">
              {{ user.create_time ? new Date(user.create_time * 1000).toLocaleDateString() : '-' }}
            </div>
            <div class="stat-label">
              注册时间
            </div>
          </div>
        </div>
      </div>
      <n-empty v-else description="未找到用户信息" />
    </n-card>

    <!-- 数据标签页 -->
    <n-card class="data-tabs-card" :bordered="false">
      <n-tabs type="line" animated @update:value="handleTabChange">
        <n-tab-pane name="basic" tab="基本信息">
          <div class="user-info-sections">
            <!-- 基础信息 -->
            <n-card title="基础信息" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="用户ID">
                  {{ user?.id }}
                </n-descriptions-item>
                <n-descriptions-item label="用户名">
                  {{ user?.username }}
                </n-descriptions-item>
                <n-descriptions-item label="昵称">
                  {{ user?.nickname || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="邮箱">
                  {{ user?.email || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="手机">
                  {{ user?.mobile || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="语言">
                  {{ formatLanguage(user?.language) }}
                </n-descriptions-item>
                <n-descriptions-item label="国家">
                  {{ user?.country || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="角色">
                  {{ user?.role === 'admin' ? '管理员' : '用户' }}
                </n-descriptions-item>
                <n-descriptions-item label="等级">
                  {{ user?.level || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="状态">
                  <NTag :type="user?.status === 1 ? 'success' : 'error'">
                    {{ user?.status === 1 ? '正常' : '禁用' }}
                  </NTag>
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <!-- 资产信息 -->
            <n-card title="资产信息" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="余额">
                  <span class="money-amount">¥{{ user?.money ? Number(user.money).toFixed(2) : '0.00' }}</span>
                </n-descriptions-item>
                <n-descriptions-item label="积分">
                  <span class="score-amount">{{ user?.score || '0' }}</span>
                </n-descriptions-item>
                <n-descriptions-item label="API密钥" :span="2">
                  <n-text code>
                    {{ user?.apikey || '-' }}
                  </n-text>
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <!-- 登录信息 -->
            <n-card title="登录信息" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="注册时间">
                  {{ user?.create_time ? new Date(user.create_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="最后登录">
                  {{ user?.last_login_time ? new Date(user.last_login_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="注册IP">
                  {{ user?.join_ip || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="最后登录IP">
                  {{ user?.last_login_ip || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="登录失败次数">
                  {{ user?.login_failure || '0' }}
                </n-descriptions-item>
                <n-descriptions-item label="座右铭" :span="1">
                  {{ user?.motto || '-' }}
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <n-card title="实名认证" class="info-section">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="认证状态">
                  <NTag :type="getRealnameStatusType(realname?.status)">
                    {{ realname?.has_verification ? getRealnameStatusText(realname?.status) : '未认证' }}
                  </NTag>
                </n-descriptions-item>
                <n-descriptions-item label="真实姓名">
                  {{ realname?.real_name || '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="证件号码">
                  {{ maskCertificateNo(realname?.certificate_no) }}
                </n-descriptions-item>
                <n-descriptions-item label="提交时间">
                  {{ realname?.submitted_at ? new Date(realname.submitted_at * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="审核时间">
                  {{ realname?.reviewed_at ? new Date(realname.reviewed_at * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="拒绝原因">
                  {{ realname?.reject_reason || '-' }}
                </n-descriptions-item>
              </n-descriptions>
            </n-card>
          </div>

          <div class="action-buttons">
            <n-button type="primary" @click="handleEdit">
              编辑用户
            </n-button>
            <n-button type="warning" @click="handleToggleStatus">
              {{ user?.status === 1 ? '禁用用户' : '启用用户' }}
            </n-button>
            <n-button type="info" @click="handleResetApikey">
              重置API密钥
            </n-button>
            <n-button type="info" @click="handleResetPassword">
              重置密码
            </n-button>
            <n-button type="error" @click="handleDelete">
              删除用户
            </n-button>
          </div>
        </n-tab-pane>

        <n-tab-pane name="orders" tab="订单记录">
          <n-data-table
            :columns="orderColumns"
            :data="orderData"
            :loading="orderLoading"
            :pagination="orderPagination"
            @update:page="handleOrderPageChange"
            @update:page-size="handleOrderPageSizeChange"
          />
        </n-tab-pane>

        <n-tab-pane name="money" tab="余额记录">
          <n-data-table
            :columns="moneyColumns"
            :data="moneyData"
            :loading="moneyLoading"
            :pagination="moneyPagination"
            @update:page="handleMoneyPageChange"
            @update:page-size="handleMoneyPageSizeChange"
          />
        </n-tab-pane>

        <n-tab-pane name="score" tab="积分记录">
          <n-data-table
            :columns="scoreColumns"
            :data="scoreData"
            :loading="scoreLoading"
            :pagination="scorePagination"
            @update:page="handleScorePageChange"
            @update:page-size="handleScorePageSizeChange"
          />
        </n-tab-pane>

        <n-tab-pane name="withdraw" tab="提现记录">
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
      title="重置密码"
      positive-text="确认"
      negative-text="取消"
      @positive-click="confirmResetPassword"
      @negative-click="cancelResetPassword"
    >
      <template #default>
        <div style="margin-bottom: 16px;">
          <p>您正在重置用户 <strong>{{ user?.username }}</strong> (ID: {{ user?.id }}) 的密码</p>
          <n-form-item label="新密码" required>
            <n-input v-model:value="newPassword" type="password" placeholder="请输入新密码" show-password-on="click" />
          </n-form-item>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showWithdrawDetailModal" preset="card" title="提现记录详情" style="width: 620px;">
      <template v-if="withdrawDetail">
        <n-descriptions :column="1" bordered label-placement="left">
          <n-descriptions-item label="申请ID">
            {{ withdrawDetail.id }}
          </n-descriptions-item>
          <n-descriptions-item label="用户ID">
            {{ withdrawDetail.user_id }}
          </n-descriptions-item>
          <n-descriptions-item label="提现金额">
            ¥{{ Number(withdrawDetail.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <NTag :type="getWithdrawStatusMeta(withdrawDetail.status).type">
              {{ getWithdrawStatusMeta(withdrawDetail.status).label }}
            </NTag>
          </n-descriptions-item>
          <n-descriptions-item label="收款方式">
            {{ withdrawDetail.account_type }}
          </n-descriptions-item>
          <n-descriptions-item label="账户名称">
            {{ withdrawDetail.account_name }}
          </n-descriptions-item>
          <n-descriptions-item label="收款账号">
            {{ withdrawDetail.account_no }}
          </n-descriptions-item>
          <n-descriptions-item label="收款人">
            {{ withdrawDetail.real_name }}
          </n-descriptions-item>
          <n-descriptions-item label="用户备注">
            {{ withdrawDetail.remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="审核备注">
            {{ withdrawDetail.review_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="打款备注">
            {{ withdrawDetail.transfer_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="申请时间">
            {{ formatTime(withdrawDetail.create_time) }}
          </n-descriptions-item>
          <n-descriptions-item label="审核时间">
            {{ formatTime(withdrawDetail.reviewed_at) }}
          </n-descriptions-item>
          <n-descriptions-item label="审核人">
            {{ getAdminDisplayName(withdrawDetail.reviewed_by) }}
          </n-descriptions-item>
          <n-descriptions-item label="打款时间">
            {{ formatTime(withdrawDetail.paid_at) }}
          </n-descriptions-item>
          <n-descriptions-item label="打款人">
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
