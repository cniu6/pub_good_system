<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { NButton, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns, FormRules } from 'naive-ui'
import NovaIcon from '@/components/common/NovaIcon.vue'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import {

  adminUserApi,

  createUser,
  deleteUser,
  fetchAdminUserPage,
  loginAsUser,
  openLoginAsUserWindow,
  resetUserApikey,
  updateAdminUserProfile,

} from '@/service/api/admin/user'
import type { AdminUser, AdminUserRealnameSummary, LoginAsAuthGuard, UserSimpleInfo } from '@/service/api/admin/user'
import { addScoreLog, fetchWithdrawRecords, generateNos, operateUserMoney } from '@/service/api/admin/finance'
import type { MoneyOperationPayload, WithdrawRecord } from '@/service/api/admin/finance'
import { buildAdminEntryUrl } from '@/router/constants'

const route = useRoute()
const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()

function reportAdminUsersError(message: string, error?: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

// 搜索表单
const searchForm = reactive({
  keyword: '',
  realnameStatus: null as 0 | 1 | 2 | null,
})

// 角色选项 - 移除"全部角色"选项
const roleOptions = computed(() => [
  { label: t('adminUsers.admin'), value: 'admin' },
  { label: t('adminUsers.normalUser'), value: 'user' },
])

// 状态选项 - 编辑用（使用数字类型）
const userStatusOptions = computed(() => [
  { label: t('adminUsers.enabled'), value: 1 },
  { label: t('adminUsers.disabled'), value: 0 },
])

const realnameFilterOptions = computed(() => [
  { label: t('realname.pending'), value: 0 },
  { label: t('realname.approved'), value: 1 },
  { label: t('realname.rejected'), value: 2 },
])

// 性别选项（使用数字类型）
const genderOptions = computed(() => [
  { label: t('adminUsers.unknown'), value: 0 },
  { label: t('adminUsers.male'), value: 1 },
  { label: t('adminUsers.female'), value: 2 },
])

const languageOptions = computed(() => [
  { label: t('adminUsersDetail.chinese'), value: 'zh-CN' },
  { label: t('adminUsersDetail.english'), value: 'en-US' },
])

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10, // 改为默认10个/页
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

// 用户数据
const userData = ref<AdminUser[]>([])
const loading = ref(false)

// 用户表单相关
const showUserModal = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const userForm = reactive({
  username: '',
  nickname: '',
  email: '',
  mobile: '',
  language: 'zh-CN',
  country: '',
  admin_remark: '',
  password: '',
  role: 'user',
  level: 1,
  status: 1,
  avatar: '',
  gender: 0,
  birthday: null as any,
  motto: '',
})

// 表单验证规则
const rules: FormRules = {
  username: [
    { required: true, message: t('adminUsers.enterUsername'), trigger: 'blur' },
    { min: 3, max: 20, message: t('adminUsers.usernameLength'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: t('adminUsers.enterEmail'), trigger: 'blur' },
    { type: 'email' as any, message: t('adminUsers.invalidEmail'), trigger: 'blur' },
  ],
  role: [
    { required: true, message: t('adminUsers.selectRole'), trigger: 'change' },
  ],
  level: [
    { type: 'number' as any, message: t('adminUsers.enterValidLevel'), trigger: 'change' },
  ],
  status: [
    {
      required: true,
      validator: (_rule: any, value: any) => {
        if (value === null || value === undefined || value === '') {
          return new Error(t('adminUsers.selectStatus'))
        }
        return true
      },
      trigger: 'change',
    },
  ],
}

// 密码验证规则：仅新建用户时填写，编辑密码请使用专用“重置密码”流程
const passwordRule = computed(() => {
  if (isEdit.value)
    return []
  return [
    { required: true, message: t('adminUsers.enterPassword'), trigger: 'blur' },
    { min: 8, message: t('adminUsers.passwordLength'), trigger: 'blur' },
  ]
})

// 用户详情相关
const showUserDetailModal = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const selectedUserRealname = ref<AdminUserRealnameSummary | null>(null)
const resettingApikey = ref(false)
const showWithdrawDetailModal = ref(false)
const withdrawDetail = ref<WithdrawRecord | null>(null)
const adminUserMap = ref<Record<number, UserSimpleInfo>>({})

// 重置密码相关
const showResetPasswordModal = ref(false)
const resettingPassword = ref(false)
const resetPasswordForm = reactive({
  newPassword: '',
  confirmPassword: '',
})

// 标签页相关
const activeTab = ref('details')
const isFullscreen = ref(false)

// 余额管理相关
const balanceForm = reactive({
  amount: 0,
  operation: 'balance_only', // 'balance_only', 'log_only', 'order_only', 'balance_log', 'balance_order', 'log_order', 'both'
  memo: '',
  orderNo: '',
  tradeNo: '',
  orderStatus: 1,
})

const orderStatusOptions = computed(() => [
  { label: `${t('adminUsersDetail.pendingPayment')}(0)`, value: 0 },
  { label: `${t('adminUsersDetail.paid')}(1)`, value: 1 },
  { label: `${t('adminUsersDetail.cancelled')}(2)`, value: 2 },
  { label: `${t('adminUsersDetail.refunded')}(3)`, value: 3 },
  { label: `${t('adminUsersDetail.paymentFailed')}(4)`, value: 4 },
])

// 积分管理相关
const scoreForm = reactive({
  amount: 0,
  operation: 'modify', // 'modify', 'log'
  memo: '',
})

const withdrawLoading = ref(false)
const withdrawData = ref<WithdrawRecord[]>([])
const withdrawPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 重置密码表单验证规则
const resetPasswordRules = {
  newPassword: [
    { required: true, message: t('adminUsers.enterNewPassword'), trigger: 'blur' },
    { min: 8, message: t('adminUsers.passwordLength'), trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: t('adminUsers.confirmPasswordPrompt'), trigger: 'blur' },
    {
      validator: (_rule: any, value: string) => {
        if (value !== resetPasswordForm.newPassword) {
          return new Error(t('adminUsers.passwordMismatch'))
        }
        return true
      },
      trigger: 'blur',
    },
  ],
}

const balanceAmountLabel = computed(() => {
  if (['log_only', 'log_order'].includes(balanceForm.operation)) {
    return t('adminUsers.logAmount')
  }
  return t('adminUsers.amount')
})

const balanceAmountPlaceholder = computed(() => {
  if (['log_only', 'log_order'].includes(balanceForm.operation)) {
    return t('adminUsers.logAmountPlaceholder')
  }
  return t('adminUsers.amountPlaceholder')
})

// 表格列配置
const columns: DataTableColumns<AdminUser> = [
  {
    type: 'selection',
    width: 50,
  },
  {
    title: 'ID',
    key: 'id',
    width: 80,
  },
  {
    title: t('adminUsers.username'),
    key: 'username',
    width: 120,
  },
  {
    title: t('adminUsers.nickname'),
    key: 'nickname',
    width: 120,
  },
  {
    title: t('adminUsers.userGroup'),
    key: 'group_id',
    width: 80,
    render: (row: AdminUser) => row.group_id?.toString() || '-',
  },
  {
    title: t('adminUsers.email'),
    key: 'email',
    width: 200,
    ellipsis: true,
  },
  {
    title: t('adminUsers.mobile'),
    key: 'mobile',
    width: 120,
  },
  {
    title: t('adminUsers.language'),
    key: 'language',
    width: 100,
    render: (row: AdminUser) => formatLanguage(row.language),
  },
  {
    title: t('adminUsers.role'),
    key: 'role',
    width: 100,
    render: (row: AdminUser) => {
      const roleMap: Record<string, { type: string, label: string }> = {
        admin: { type: 'error', label: t('adminUsers.admin') },
        user: { type: 'success', label: t('adminUsers.user') },
      }
      const role = roleMap[row.role] || { type: 'default', label: row.role }
      return h(NTag, { type: role.type as any }, () => role.label)
    },
  },
  {
    title: t('adminUsers.level'),
    key: 'level',
    width: 80,
  },
  {
    title: t('adminUsers.balance'),
    key: 'money',
    width: 100,
    render: (row: AdminUser) => `¥${formatCurrency(row.money)}`,
  },
  {
    title: t('adminUsers.rechargeRetentionRatio'),
    key: 'balance_paid_ratio',
    width: 130,
    render: (row: AdminUser) => formatRechargeRetentionRatio(row),
  },
  {
    title: t('adminUsers.score'),
    key: 'score',
    width: 80,
    render: (row: AdminUser) => (Number(row.score) || 0).toString(),
  },
  {
    title: t('adminUsers.status'),
    key: 'status',
    width: 100,
    render: (row: AdminUser) => {
      const statusMap: Record<string, { type: string, text: string }> = {
        1: { type: 'success', text: t('adminUsers.enabled') },
        0: { type: 'error', text: t('adminUsers.disabled') },
      }
      const statusKey = String(row.status)
      const status = statusMap[statusKey] || { type: 'default', text: t('adminUsers.unknown') }
      return h(NTag, { type: status.type as any }, { default: () => status.text })
    },
  },
  {
    title: t('adminUsers.realnameStatus'),
    key: 'realname_status',
    width: 110,
    render: (row: AdminUser) => {
      const status = row.realname_status
      return h(
        NTag,
        { type: getRealnameStatusType(status === null ? undefined : status), size: 'small' },
        {
          default: () => getRealnameStatusText(status === null ? undefined : status),
        },
      )
    },
  },
  {
    title: t('adminUsers.adminRemark'),
    key: 'admin_remark',
    width: 220,
    ellipsis: { tooltip: true },
    render: (row: AdminUser) => row.admin_remark || '-',
  },
  {
    title: t('adminUsers.registerTime'),
    key: 'create_time',
    width: 180,
    render: (row: AdminUser) => {
      if (!row.create_time)
        return '-'
      try {
        return new Date(row.create_time * 1000).toLocaleString()
      }
      catch {
        return row.create_time as any
      }
    },
  },
  {
    title: t('adminUsers.updateTime'),
    key: 'update_time',
    width: 180,
    render: (row: AdminUser) => {
      if (!row.update_time)
        return '-'
      try {
        return new Date(row.update_time * 1000).toLocaleString()
      }
      catch {
        return row.update_time as any
      }
    },
  },
  {
    title: t('adminUsers.actions'),
    key: 'actions',
    width: 250,
    render: (row: AdminUser) => {
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, {
          size: 'small',
          type: 'info',
          onClick: () => handleViewUserDetail(row),
        }, () => t('adminUsers.detail')),
        h(NButton, {
          size: 'small',
          type: 'primary',
          onClick: () => handleEdit(row),
        }, () => t('adminUsers.edit')),
        h(NButton, {
          size: 'small',
          type: 'success',
          onClick: () => handleLoginAsUser(row),
        }, () => t('adminUsers.loginAs')),
        h(NButton, {
          size: 'small',
          type: 'error',
          onClick: () => handleDelete(row.id),
        }, () => t('adminUsers.delete')),
      ])
    },
  },
]

 const selectableColumnOptions = [
   { key: 'id', label: 'ID' },
   { key: 'username', label: t('adminUsers.username') },
   { key: 'nickname', label: t('adminUsers.nickname') },
   { key: 'group_id', label: t('adminUsers.userGroup') },
   { key: 'email', label: t('adminUsers.email') },
   { key: 'mobile', label: t('adminUsers.mobile') },
   { key: 'language', label: t('adminUsers.language') },
   { key: 'role', label: t('adminUsers.role') },
   { key: 'level', label: t('adminUsers.level') },
   { key: 'money', label: t('adminUsers.balance') },
   { key: 'balance_paid_ratio', label: t('adminUsers.rechargeRetentionRatio') },
   { key: 'score', label: t('adminUsers.score') },
   { key: 'status', label: t('adminUsers.status') },
   { key: 'realname_status', label: t('adminUsers.realnameStatus') },
   { key: 'admin_remark', label: t('adminUsers.adminRemark') },
   { key: 'create_time', label: t('adminUsers.registerTime') },
   { key: 'update_time', label: t('adminUsers.updateTime') },
 ]

 const {
   columnOptions,
   selectedColumnKeys,
   visibleColumns,
   visibleColumnCount,
   totalColumnCount,
   tableScrollX,
   resetSelectedColumns,
 } = useTableColumnVisibility<AdminUser>({
   storageKey: 'admin-users-list',
   columns,
   options: selectableColumnOptions,
   minVisibleCount: 1,
   minScrollX: 1280,
 })

// 获取用户数据
async function fetchData() {
  loading.value = true
  try {
    const response: any = await fetchAdminUserPage({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword || undefined,
      realname_status: searchForm.realnameStatus ?? undefined,
    })

    if (response.isSuccess) {
      const data = response.data
      if (Array.isArray(data)) {
        userData.value = data
        pagination.itemCount = response.total || 0
      }
      else if (data && data.list) {
        userData.value = data.list || []
        pagination.itemCount = data.total || response.total || 0
      }
      else {
        userData.value = []
        pagination.itemCount = response.total || 0
      }
    }
    else {
      message.error(response.message || t('adminUsers.fetchListFailed'))
    }
  }
  catch {
    message.error(t('adminUsers.fetchListFailed'))
  }
  finally {
    loading.value = false
  }
}

// 查看用户详情
async function handleViewUserDetail(user: AdminUser) {
  try {
    const response: any = await adminUserApi.detail(user.id)
    if (!response.isSuccess || !response.data?.user) {
      message.error(response.message || t('adminUsers.fetchDetailFailed'))
      return
    }
    selectedUser.value = response.data.user
    selectedUserRealname.value = response.data.realname || { has_verification: false }
    showUserDetailModal.value = true
  }
  catch {
    message.error(t('adminUsers.fetchDetailFailed'))
  }
}

function getRealnameStatusText(status?: number) {
  const map: Record<number, string> = {
    0: t('realname.pending'),
    1: t('realname.approved'),
    2: t('realname.rejected'),
  }
  return status !== undefined ? map[status] || t('adminUsers.unknown') : t('adminUsers.unverified')
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

function formatCurrency(value?: number | null) {
  return Number(value || 0).toFixed(2)
}

function formatRechargeRetentionRatio(row: AdminUser) {
  const totalPaid = Number(row.total_paid_amount || 0)
  if (totalPaid <= 0)
    return '-'
  return `${(Number(row.balance_paid_ratio || 0) * 100).toFixed(2)}%`
}

function getWithdrawStatusMeta(status?: number): { type: 'warning' | 'info' | 'error' | 'success', label: string } {
  const map: Record<number, { type: 'warning' | 'info' | 'error' | 'success', label: string }> = {
    0: { type: 'warning', label: t('moneyScore.statusPending') },
    1: { type: 'info', label: t('moneyScore.statusApproved') },
    2: { type: 'error', label: t('moneyScore.statusRejected') },
    3: { type: 'success', label: t('moneyScore.statusPaid') },
  }
  return status !== undefined ? map[status] || { type: 'info', label: t('adminUsers.unknown') } : { type: 'info', label: t('adminUsers.unknown') }
}

function getAdminDisplayName(adminId?: number | null) {
  if (!adminId)
    return '-'
  const admin = adminUserMap.value[adminId]
  return admin?.nickname || admin?.username || t('adminUsers.adminFallback', { id: adminId })
}

function formatLanguage(language?: string) {
  if (!language)
    return '-'
  if (language === 'zh-CN')
    return t('adminUsersDetail.chinese')
  if (language === 'en-US')
    return t('adminUsersDetail.english')
  return language
}

// 添加用户
function handleAdd() {
  isEdit.value = false
  resetUserForm()
  showUserModal.value = true
}

// 编辑用户
function handleEdit(user: AdminUser) {
  isEdit.value = true
  selectedUser.value = user

  resetForms()

  const statusValue = Number(user.status) || 0

  Object.assign(userForm, {
    username: user.username,
    nickname: user.nickname || '',
    email: user.email || '',
    mobile: user.mobile || '',
    language: user.language || 'zh-CN',
    country: user.country || '',
    admin_remark: user.admin_remark || '',
    password: '',
    role: user.role,
    level: user.level,
    status: statusValue,
    avatar: user.avatar || '',
    gender: user.gender || 0,
    birthday: user.birthday ? Number(user.birthday) * 1000 : null,
    motto: user.motto || '',
  })
  showUserModal.value = true
  fetchWithdrawData()
}

// 删除用户
function handleDelete(userId: number) {
  dialog.warning({
    title: t('adminUsers.confirmDeleteTitle'),
    content: t('adminUsers.confirmDeleteContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const response: any = await deleteUser(userId)
        if (response.isSuccess) {
          message.success(t('adminUsers.deleteSuccess'))
          fetchData()
        }
        else {
          message.error(response.message || t('adminUsers.deleteFailed'))
        }
      }
      catch {
        message.error(t('adminUsers.deleteFailed'))
      }
    },
  })
}

// 重置API密钥
async function handleResetApikey() {
  if (!selectedUser.value)
    return

  dialog.warning({
    title: t('adminUsers.confirmResetApiKeyTitle'),
    content: t('adminUsers.confirmResetApiKeyContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        resettingApikey.value = true
        const response: any = await resetUserApikey(selectedUser.value!.id)
        if (response.isSuccess) {
          message.success(t('adminUsers.resetApiKeySuccess'))
          showUserDetailModal.value = false
          fetchData()
        }
        else {
          message.error(response.message || t('adminUsers.resetApiKeyFailed'))
        }
      }
      catch {
        message.error(t('adminUsers.resetApiKeyFailed'))
      }
      finally {
        resettingApikey.value = false
      }
    },
  })
}

// 显示重置密码弹窗
function handleShowResetPassword() {
  if (!selectedUser.value)
    return

  resetPasswordForm.newPassword = ''
  resetPasswordForm.confirmPassword = ''
  showResetPasswordModal.value = true
}

// 重置密码
async function handleResetPassword() {
  try {
    if (!resetPasswordForm.newPassword) {
      message.error(t('adminUsers.enterNewPassword'))
      return
    }
    if (resetPasswordForm.newPassword.length < 8) {
      message.error(t('adminUsers.passwordLength'))
      return
    }
    if (resetPasswordForm.newPassword !== resetPasswordForm.confirmPassword) {
      message.error(t('adminUsers.passwordMismatch'))
      return
    }

    resettingPassword.value = true

    const { resetUserPassword } = await import('@/service/api/admin/user')
    const response: any = await resetUserPassword(selectedUser.value!.id, {
      password: resetPasswordForm.newPassword,
    })

    if (response.isSuccess) {
      message.success(t('adminUsers.resetPasswordSuccess'))
      showResetPasswordModal.value = false
      showUserDetailModal.value = false
      fetchData()
    }
    else {
      message.error(response.message || t('adminUsers.resetPasswordFailed'))
    }
  }
  catch {
    message.error(t('adminUsers.resetPasswordFailed'))
  }
  finally {
    resettingPassword.value = false
  }
}

// 重置用户表单
function resetUserForm() {
  Object.assign(userForm, {
    username: '',
    nickname: '',
    email: '',
    mobile: '',
    language: 'zh-CN',
    country: '',
    admin_remark: '',
    password: '',
    role: 'user',
    level: 1,
    status: 1,
    avatar: '',
    gender: 0,
    birthday: null,
    motto: '',
  })
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value?.validate()
    submitting.value = true

    if (isEdit.value) {
      const originalUser = selectedUser.value
      const changedData: any = {}

      if (userForm.nickname !== (originalUser?.nickname || '')) {
        changedData.nickname = userForm.nickname
      }
      if (userForm.email !== (originalUser?.email || '')) {
        changedData.email = userForm.email
      }
      if (userForm.mobile !== (originalUser?.mobile || '')) {
        changedData.mobile = userForm.mobile
      }
      if (userForm.language !== (originalUser?.language || 'zh-CN')) {
        changedData.language = userForm.language
      }
      if (userForm.country !== (originalUser?.country || '')) {
        changedData.country = userForm.country
      }
      if (userForm.admin_remark !== (originalUser?.admin_remark || '')) {
        changedData.admin_remark = userForm.admin_remark
      }
      if (userForm.role !== originalUser?.role) {
        changedData.role = userForm.role
      }
      if (userForm.level !== originalUser?.level) {
        changedData.level = userForm.level
      }

      const originalStatus = Number(originalUser?.status) || 0
      if (userForm.status !== originalStatus) {
        changedData.status = userForm.status
      }

      if (userForm.avatar !== (originalUser?.avatar || '')) {
        changedData.avatar = userForm.avatar
      }
      if (userForm.gender !== (originalUser?.gender || 0)) {
        changedData.gender = userForm.gender
      }

      const originalBirthday = originalUser?.birthday ? Number(originalUser.birthday) * 1000 : null
      const formBirthday = userForm.birthday ? Number(userForm.birthday) : null
      if (originalBirthday !== formBirthday) {
        changedData.birthday = formBirthday ? Math.floor(formBirthday / 1000) : null
      }

      if (userForm.motto !== (originalUser?.motto || '')) {
        changedData.motto = userForm.motto
      }

      if (Object.keys(changedData).length === 0) {
        message.warning(t('adminUsers.noChangesDetected'))
        submitting.value = false
        return
      }

      const response: any = await updateAdminUserProfile(selectedUser.value?.id as number, changedData)
      if (response.isSuccess) {
        message.success(t('adminUsers.updateSuccess'))
        showUserModal.value = false
        fetchData()
      }
      else {
        message.error(response.message || t('adminUsers.updateFailed'))
      }
    }
    else {
      const userPayload = {
        username: userForm.username,
        password: userForm.password,
        email: userForm.email,
        nickname: userForm.nickname,
        mobile: userForm.mobile,
        language: userForm.language,
        country: userForm.country,
        admin_remark: userForm.admin_remark,
        level: userForm.level,
        role: userForm.role,
        status: userForm.status,
      }
      const response: any = await createUser(userPayload as any)
      if (response.isSuccess) {
        message.success(t('adminUsers.createSuccess'))
        showUserModal.value = false
        fetchData()
      }
      else {
        message.error(response.message || t('adminUsers.createFailed'))
      }
    }
  }
  catch (error) {
    reportAdminUsersError('[adminUsers] form validation failed', error)
  }
  finally {
    submitting.value = false
  }
}

// 刷新数据
function handleRefresh() {
  fetchData()
}

// 搜索
function handleSearch() {
  pagination.page = 1
  fetchData()
}

// 重置搜索
function handleReset() {
  Object.assign(searchForm, {
    keyword: '',
    realnameStatus: null,
  })
  pagination.page = 1
  fetchData()
}

// 分页变化
function handlePageChange(page: number) {
  pagination.page = page
  fetchData()
}

// 每页大小变化
function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchData()
}

function handleRealnameStatusChange() {
  pagination.page = 1
  fetchData()
}

// 以用户身份登录（在新标签页打开，不污染管理员登录态）
function isAdminRole(role?: string) {
  return String(role || '').trim().toLowerCase() === 'admin'
}

async function doLoginAsUser(user: AdminUser, authGuard: LoginAsAuthGuard) {
  const res: any = await loginAsUser(user.id, { auth_guard: authGuard })
  if (!(res.isSuccess && res.data?.user && res.data?.token))
    return

  const targetUrl = authGuard === 'admin'
    ? `${buildAdminEntryUrl('/dashboard')}?_t=${Date.now()}`
    : `/user/dashboard?_t=${Date.now()}`

  openLoginAsUserWindow(res.data.user, res.data.token, res.data.refreshToken, res.data.expiresAt, targetUrl)
  message.success(
    authGuard === 'admin'
      ? t('adminUsers.openedAdminConsole')
      : t('adminUsers.openedUserConsole'),
  )
}

function confirmAdminLoginTarget(user: AdminUser) {
  // 目标是管理员时二次确认：明确进入管理后台还是用户前端
  const d = dialog.create({
    type: 'warning',
    title: t('adminUsers.confirmAdminLoginTitle'),
    content: t('adminUsers.confirmAdminLoginContent', { username: user.username }),
    closable: true,
    maskClosable: true,
    action: () => h(NSpace, { justify: 'end' }, () => [
      h(NButton, {
        size: 'small',
        onClick: () => d.destroy(),
      }, () => t('common.cancel')),
      h(NButton, {
        size: 'small',
        type: 'info',
        onClick: () => {
          d.destroy()
          void doLoginAsUser(user, 'user')
        },
      }, () => t('adminUsers.loginAsUserFrontend')),
      h(NButton, {
        size: 'small',
        type: 'warning',
        onClick: () => {
          d.destroy()
          void doLoginAsUser(user, 'admin')
        },
      }, () => t('adminUsers.loginAsAdminConsole')),
    ]),
  })
}

function handleLoginAsUser(user: AdminUser) {
  dialog.warning({
    title: t('adminUsers.confirmLoginTitle'),
    content: t('adminUsers.confirmLoginContent', { username: user.username }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      if (isAdminRole(user.role)) {
        confirmAdminLoginTarget(user)
        return
      }
      void doLoginAsUser(user, 'user')
    },
  })
}

// 组件挂载时获取数据
onMounted(() => {
  const searchKeyword = route.query.search as string
  if (searchKeyword) {
    searchForm.keyword = searchKeyword
  }
  fetchData()
})

// 头像加载错误处理
function handleAvatarError() {
  reportAdminUsersError('[adminUsers] avatar load failed')
}

// 切换全屏
function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
}

// 重置表单数据
function resetForms() {
  balanceForm.amount = 0
  balanceForm.operation = 'balance_only'
  balanceForm.memo = ''
  balanceForm.orderNo = ''
  balanceForm.tradeNo = ''
  balanceForm.orderStatus = 1

  scoreForm.amount = 0
  scoreForm.operation = 'modify'
  scoreForm.memo = ''

  withdrawData.value = []
  withdrawPagination.page = 1
  withdrawPagination.pageSize = 10
  withdrawPagination.itemCount = 0

  activeTab.value = 'details'
}

function handleUserModalTabChange(tab: string) {
  activeTab.value = tab
  if (tab === 'withdraw' && isEdit.value) {
    fetchWithdrawData()
  }
}

async function fetchWithdrawData() {
  if (!selectedUser.value)
    return

  withdrawLoading.value = true
  try {
    const response: any = await fetchWithdrawRecords({
      page: withdrawPagination.page,
      page_size: withdrawPagination.pageSize,
      user_id: selectedUser.value.id,
    })
    if (response.isSuccess) {
      withdrawData.value = response.data?.list || []
      withdrawPagination.itemCount = response.data?.total || 0
      const adminIds = Array.from(new Set(withdrawData.value.flatMap(item => [item.reviewed_by, item.paid_by]).filter(Boolean) as number[]))
      adminUserMap.value = await adminUserApi.batchSimpleInfo(adminIds)
    }
    else {
      message.error(response.message || t('moneyScore.fetchWithdrawFailed'))
    }
  }
  catch {
    message.error(t('moneyScore.fetchWithdrawFailed'))
  }
  finally {
    withdrawLoading.value = false
  }
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

function openWithdrawDetail(row: WithdrawRecord) {
  withdrawDetail.value = row
  showWithdrawDetailModal.value = true
}

const withdrawColumns: DataTableColumns<WithdrawRecord> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminUsers.amount'),
    key: 'amount',
    width: 100,
    render: row => `¥${(Number(row.amount) || 0).toFixed(2)}`,
  },
  {
    title: t('adminUsers.status'),
    key: 'status',
    width: 100,
    render: (row) => {
      const status = getWithdrawStatusMeta(row.status)
      return h(NTag, { type: status.type }, () => status.label)
    },
  },
  { title: t('adminUsers.method'), key: 'account_type', width: 90 },
  { title: t('moneyScore.accountName'), key: 'account_name', width: 120, ellipsis: true },
  {
    title: t('moneyScore.accountNo'),
    key: 'account_no',
    width: 160,
    ellipsis: true,
    render: row => maskAccountNo(row.account_no),
  },
  { title: t('adminUsers.reviewer'), key: 'reviewed_by', width: 110, render: row => getAdminDisplayName(row.reviewed_by) },
  { title: t('adminUsers.payer'), key: 'paid_by', width: 110, render: row => getAdminDisplayName(row.paid_by) },
  {
    title: t('moneyScore.createdAt'),
    key: 'create_time',
    width: 170,
    render: row => formatTime(row.create_time),
  },
  {
    title: t('adminUsers.actions'),
    key: 'actions',
    width: 90,
    render: row => h(
      'a',
      {
        style: 'color: var(--n-primary-color); cursor: pointer;',
        onClick: () => openWithdrawDetail(row),
      },
      t('adminUsers.detail'),
    ),
  },
]

// 处理余额操作
async function handleBalanceOperation() {
  if (!selectedUser.value)
    return

  try {
    submitting.value = true

    const isOrder = ['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)
    const needsAmount = balanceForm.operation !== 'order_only'

    if (needsAmount && balanceForm.amount === null) {
      message.warning(t('adminUsers.amountRequired'))
      return
    }

    if (needsAmount && Number(balanceForm.amount) === 0) {
      message.warning(t('adminUsers.amountCannotBeZero'))
      return
    }

    if (isOrder && !balanceForm.orderNo) {
      message.warning(t('adminUsers.orderNoRequired'))
      return
    }

    const response: any = await operateUserMoney(selectedUser.value.id, {
      money: Number(balanceForm.amount || 0),
      memo: balanceForm.memo,
      operation: balanceForm.operation as MoneyOperationPayload['operation'],
      order_no: balanceForm.orderNo || undefined,
      trade_no: balanceForm.tradeNo || undefined,
      order_status: isOrder ? balanceForm.orderStatus : undefined,
    })

    if (response.isSuccess) {
      message.success(t('adminUsers.balanceOperationSuccess'))
      fetchData()
      showUserModal.value = false
    }
    else {
      message.error(response.message || t('adminUsers.balanceOperationFailed'))
    }
  }
  catch (error) {
    reportAdminUsersError('[adminUsers] balance operation failed', error)
    message.error(t('adminUsers.operationFailed'))
  }
  finally {
    submitting.value = false
  }
}

// 自动填充订单号（从后端生成）
async function handleAutoFillOrderNo() {
  try {
    const res: any = await generateNos()
    if (res.isSuccess && res.data?.order_no) {
      balanceForm.orderNo = res.data.order_no
      message.success(t('adminUsers.orderNoGenerated'))
    }
    else {
      message.error(res.message || t('adminUsers.generateOrderNoFailed'))
    }
  }
  catch {
    message.error(t('adminUsers.generateOrderNoFailed'))
  }
}

// 自动填充交易号（从后端生成）
async function handleAutoFillTradeNo() {
  try {
    const res: any = await generateNos()
    if (res.isSuccess && res.data?.trade_no) {
      balanceForm.tradeNo = res.data.trade_no
      message.success(t('adminUsers.tradeNoGenerated'))
    }
    else {
      message.error(res.message || t('adminUsers.generateTradeNoFailed'))
    }
  }
  catch {
    message.error(t('adminUsers.generateTradeNoFailed'))
  }
}

// 处理积分操作
async function handleScoreOperation() {
  if (!selectedUser.value)
    return

  try {
    submitting.value = true

    if (Number(scoreForm.amount) === 0) {
      message.warning(t('adminUsers.scoreCannotBeZero'))
      return
    }

    if (scoreForm.operation === 'modify') {
      const response: any = await adminUserApi.changeScore(selectedUser.value.id, {
        score: scoreForm.amount,
        memo: scoreForm.memo,
      })
      if (response.isSuccess) {
        message.success(t('adminUsers.scoreChangedSuccess'))
        fetchData()
        showUserModal.value = false
      }
      else {
        message.error(response.message || t('adminUsers.scoreChangedFailed'))
      }
    }
    else if (scoreForm.operation === 'log') {
      const response: any = await addScoreLog(selectedUser.value.id, {
        score: scoreForm.amount,
        memo: scoreForm.memo,
      })
      if (response.isSuccess) {
        message.success(t('adminUsers.scoreLogAddedSuccess'))
        showUserModal.value = false
      }
      else {
        message.error(response.message || t('adminUsers.scoreLogAddedFailed'))
      }
    }
  }
  catch (error) {
    reportAdminUsersError('[adminUsers] score operation failed', error)
    message.error(t('adminUsers.operationFailed'))
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <div>
    <!-- 页面头部 -->
    <n-card class="header-card" :bordered="false">
      <div class="header-content">
        <div class="header-title">
          <NovaIcon :size="24" class="title-icon" icon="icon-park-outline:user" />
          <span>{{ t('adminUsers.userManagement') }}</span>
        </div>
        <NSpace :wrap="false" :size="12" class="header-actions">
          <NButton @click="handleRefresh">
            <template #icon>
              <NovaIcon icon="icon-park-outline:refresh" />
            </template>
            {{ t('common.reload') }}
          </NButton>
          <NButton type="primary" @click="handleAdd">
            <template #icon>
              <NovaIcon icon="icon-park-outline:plus" />
            </template>
            {{ t('adminUsers.addUser') }}
          </NButton>
        </NSpace>
      </div>
    </n-card>

    <!-- 搜索和筛选 -->
    <n-card class="search-card" :bordered="false">
      <n-form :model="searchForm" label-placement="left" :label-width="80">
        <n-grid :cols="24" :x-gap="16" responsive="screen">
          <n-form-item-gi span="24 600:10 800:10" :label="t('adminRealname.keyword')">
            <n-input
              v-model:value="searchForm.keyword"
              :placeholder="t('adminUsers.searchPlaceholder')"
              clearable
              @keyup.enter="handleSearch"
            />
          </n-form-item-gi>
          <n-form-item-gi span="24 600:6 800:6" :label="t('adminUsers.realnameStatus')">
            <n-select
              v-model:value="searchForm.realnameStatus"
              :options="realnameFilterOptions"
              clearable
              :placeholder="t('common.all')"
              @update:value="handleRealnameStatusChange"
            />
          </n-form-item-gi>
          <n-form-item-gi span="24 600:8 800:8" class="search-actions">
            <NSpace justify="center">
              <NButton type="primary" class="search-btn" @click="handleSearch">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:search" />
                </template>
                {{ t('moneyScore.search') }}
              </NButton>
              <NButton class="reset-btn" @click="handleReset">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:refresh" />
                </template>
                {{ t('common.reset') }}
              </NButton>
            </NSpace>
          </n-form-item-gi>
        </n-grid>
      </n-form>
    </n-card>

    <!-- 用户列表 -->
    <n-card class="table-card" :bordered="false">
      <NSpace justify="end" style="margin-bottom: 12px;">
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
      </NSpace>
      <n-data-table
        :columns="visibleColumns"
        :data="userData"
        :loading="loading"
        :pagination="false"
        :row-key="(row) => row.id"
        :scrollbar-props="{ trigger: 'hover' }"
        :scroll-x="tableScrollX"
      />

      <!-- 外部分页组件 -->
      <div class="pagination-container">
        <div class="pagination-info">
          <n-text depth="3">
            {{ t('adminUsers.paginationInfo', { total: pagination.itemCount, page: pagination.page, pageSize: pagination.pageSize }) }}
          </n-text>
        </div>
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :item-count="pagination.itemCount"
          :page-sizes="pagination.pageSizes"
          :show-size-picker="pagination.showSizePicker"
          show-quick-jumper
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </n-card>

    <!-- 添加/编辑用户模态框 -->
    <n-modal
      v-model:show="showUserModal"
      preset="card"
      :title="isEdit ? t('adminUsers.editUser') : t('adminUsers.addUser')"
      :style="isFullscreen ? 'width: 100vw; height: 100vh; max-width: none; max-height: none;' : 'width: 800px;'"
      :bordered="false"
      :closable="!isFullscreen"
      :mask-closable="!isFullscreen"
    >
      <template #header-extra>
        <NButton quaternary circle @click="toggleFullscreen">
          <template #icon>
            <NovaIcon :icon="isFullscreen ? 'icon-park-outline:off-screen' : 'icon-park-outline:full-screen'" />
          </template>
        </NButton>
      </template>

      <n-tabs v-model:value="activeTab" type="line" animated @update:value="handleUserModalTabChange">
        <!-- 详情标签页 -->
        <n-tab-pane name="details" :tab="t('adminUsers.detail')">
          <n-form
            ref="formRef"
            :model="userForm"
            :rules="rules"
            label-placement="left"
            :label-width="100"
          >
            <n-grid :cols="2" :x-gap="16">
              <n-form-item-gi :label="t('adminUsers.username')" path="username">
                <n-input
                  v-model:value="userForm.username"
                  :placeholder="t('adminUsers.enterUsername')"
                  :disabled="isEdit"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.nickname')" path="nickname">
                <n-input
                  v-model:value="userForm.nickname"
                  :placeholder="t('adminUsers.enterNickname')"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.email')" path="email">
                <n-input
                  v-model:value="userForm.email"
                  :placeholder="t('adminUsers.enterEmail')"
                  @blur="userForm.email = userForm.email.includes('@') ? `${userForm.email.split('@')[0]}@${userForm.email.split('@')[1].toLowerCase()}` : userForm.email"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.mobile')" path="mobile">
                <n-input
                  v-model:value="userForm.mobile"
                  :placeholder="t('adminUsers.enterMobile')"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.language')" path="language">
                <n-select
                  v-model:value="userForm.language"
                  :options="languageOptions"
                  :placeholder="t('adminUsers.selectLanguage')"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.country')" path="country">
                <n-input
                  v-model:value="userForm.country"
                  :placeholder="t('adminUsers.enterCountry')"
                />
              </n-form-item-gi>
              <n-form-item-gi span="2" :label="t('adminUsers.adminRemark')" path="admin_remark">
                <n-input
                  v-model:value="userForm.admin_remark"
                  type="textarea"
                  :placeholder="t('adminUsers.enterAdminRemark')"
                  :rows="3"
                />
              </n-form-item-gi>
              <n-form-item-gi v-if="!isEdit" span="2" :label="t('adminUsers.password')" path="password" :rule="passwordRule">
                <n-input
                  v-model:value="userForm.password"
                  type="password"
                  :placeholder="t('adminUsers.enterNewPassword')"
                  show-password-on="click"
                />
                <template #feedback>
                  <span class="password-tip">{{ t('adminUsers.setPasswordTip') }}</span>
                </template>
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.role')" path="role">
                <n-select
                  v-model:value="userForm.role"
                  :options="roleOptions"
                  :placeholder="t('adminUsers.selectRole')"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.level')" path="level">
                <n-input-number
                  v-model:value="userForm.level"
                  :placeholder="t('adminUsers.enterUserLevel')"
                  :min="0"
                  :max="100"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.status')" path="status">
                <n-select
                  v-model:value="userForm.status"
                  :options="userStatusOptions"
                  :placeholder="t('adminUsers.selectStatus')"
                />
              </n-form-item-gi>
              <n-form-item-gi :label="t('adminUsers.gender')" path="gender">
                <n-select
                  v-model:value="userForm.gender"
                  :options="genderOptions"
                  :placeholder="t('adminUsers.selectGender')"
                />
              </n-form-item-gi>
              <n-form-item-gi span="2" :label="t('adminUsers.birthday')" path="birthday">
                <n-date-picker
                  v-model:value="userForm.birthday"
                  type="date"
                  :placeholder="t('adminUsers.selectBirthday')"
                  clearable
                />
              </n-form-item-gi>
              <n-form-item-gi span="2" :label="t('adminUsers.avatar')" path="avatar">
                <NSpace vertical>
                  <n-input
                    v-model:value="userForm.avatar"
                    :placeholder="t('adminUsers.enterAvatarUrl')"
                  />
                  <!-- 头像预览 -->
                  <div v-if="userForm.avatar" class="avatar-preview">
                    <n-text depth="3" style="font-size: 12px;">
                      {{ t('adminUsers.preview') }}:
                    </n-text>
                    <n-avatar
                      :src="userForm.avatar"
                      size="large"
                      fallback-src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjQiIGhlaWdodD0iNjQiIHZpZXdCb3g9IjAgMCA2NCA2NCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iMzIiIGN5PSIzMiIgcj0iMzIiIGZpbGw9IiNGNUY1RjUiLz4KPHN2ZyB3aWR0aD0iMzIiIGhlaWdodD0iMzIiIHZpZXdCb3g9IjAgMCAzMiAzMiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4PSIxNiIgeT0iMTYiPgo8cGF0aCBkPSJNMTYgMTZDMTguMjA5MSAxNiAyMCAxNC4yMDkxIDIwIDEyQzIwIDkuNzkwODYgMTguMjA5MSA4IDE2IDhDMTMuNzkwOSA4IDEyIDkuNzkwODYgMTIgMTJDMTIgMTQuMjA5MSAxMy43OTA5IDE2IDE2IDE2WiIgZmlsbD0iIzk5OTk5OSIvPgo8cGF0aCBkPSJNMjQgMjRWMjJDMjQgMTkuNzkwOSAyMi4yMDkxIDE4IDIwIDE4SDEyQzkuNzkwODYgMTggOCAxOS43OTA5IDggMjJWMMjQiIGZpbGw9IiM5OTk5OTkiLz4KPC9zdmc+Cjwvc3ZnPgo="
                      @error="handleAvatarError"
                    />
                  </div>
                </NSpace>
              </n-form-item-gi>
              <n-form-item-gi span="2" :label="t('adminUsers.motto')" path="motto">
                <n-input
                  v-model:value="userForm.motto"
                  type="textarea"
                  :placeholder="t('adminUsers.enterMotto')"
                  :rows="3"
                />
              </n-form-item-gi>
            </n-grid>
          </n-form>
        </n-tab-pane>

        <!-- 余额标签页 -->
        <n-tab-pane v-if="isEdit" name="balance" :tab="t('adminUsers.balance')">
          <div class="balance-management">
            <NSpace vertical size="large">
              <!-- 当前余额显示 -->
              <n-card :title="t('adminUsers.currentBalance')" size="small">
                <n-statistic
                  :label="t('adminUsers.userBalance')"
                  :value="selectedUser?.money || 0"
                  :precision="2"
                >
                  <template #prefix>
                    ¥
                  </template>
                </n-statistic>
              </n-card>

              <!-- 余额操作 -->
              <n-form label-placement="left" :label-width="100">
                <n-form-item v-if="balanceForm.operation !== 'order_only'" :label="balanceAmountLabel">
                  <n-input-number
                    v-model:value="balanceForm.amount"
                    :placeholder="balanceAmountPlaceholder"
                    :precision="2"
                    :step="0.01"
                  />
                </n-form-item>

                <n-form-item :label="t('adminUsers.operationType')">
                  <n-radio-group v-model:value="balanceForm.operation">
                    <NSpace wrap>
                      <n-radio value="balance_only">
                        {{ t('adminUsers.balanceOnly') }}
                      </n-radio>
                      <n-radio value="log_only">
                        {{ t('adminUsers.logOnly') }}
                      </n-radio>
                      <n-radio value="order_only">
                        {{ t('adminUsers.orderOnly') }}
                      </n-radio>
                      <n-radio value="balance_log">
                        {{ t('adminUsers.balanceLog') }}
                      </n-radio>
                      <n-radio value="balance_order">
                        {{ t('adminUsers.balanceOrder') }}
                      </n-radio>
                      <n-radio value="log_order">
                        {{ t('adminUsers.logOrder') }}
                      </n-radio>
                      <n-radio value="both">
                        {{ t('adminUsers.allInOne') }}
                      </n-radio>
                    </NSpace>
                  </n-radio-group>
                </n-form-item>

                <n-form-item v-if="['log_only', 'balance_log', 'log_order', 'both'].includes(balanceForm.operation)" :label="t('moneyScore.remark')">
                  <n-input
                    v-model:value="balanceForm.memo"
                    type="textarea"
                    :placeholder="t('adminUsers.enterOperationRemark')"
                    :rows="3"
                  />
                </n-form-item>

                <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" :label="t('recharge.orderNo')">
                  <n-input-group>
                    <n-input
                      v-model:value="balanceForm.orderNo"
                      :placeholder="t('adminUsers.enterOrderNoRequired')"
                      style="flex: 1"
                    />
                    <NButton type="primary" ghost @click="handleAutoFillOrderNo">
                      {{ t('adminUsers.autoGenerate') }}
                    </NButton>
                  </n-input-group>
                </n-form-item>

                <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" :label="t('recharge.tradeNo')">
                  <n-input-group>
                    <n-input
                      v-model:value="balanceForm.tradeNo"
                      :placeholder="t('adminUsers.enterTradeNoOptional')"
                      style="flex: 1"
                    />
                    <NButton type="primary" ghost @click="handleAutoFillTradeNo">
                      {{ t('adminUsers.autoGenerate') }}
                    </NButton>
                  </n-input-group>
                </n-form-item>

                <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" :label="t('recharge.orderStatus')">
                  <n-select
                    v-model:value="balanceForm.orderStatus"
                    :options="orderStatusOptions"
                    :placeholder="t('adminUsers.selectOrderStatus')"
                  />
                </n-form-item>

                <n-alert
                  v-if="balanceForm.operation === 'order_only'"
                  type="info"
                  :show-icon="false"
                  style="margin-bottom: 16px;"
                >
                  {{ t('adminUsers.orderOnlyHint') }}
                </n-alert>

                <n-alert
                  v-if="balanceForm.operation === 'balance_order'"
                  type="info"
                  :show-icon="false"
                  style="margin-bottom: 16px;"
                >
                  {{ t('adminUsers.balanceOrderHint') }}
                </n-alert>

                <n-alert
                  v-if="balanceForm.operation === 'log_order'"
                  type="info"
                  :show-icon="false"
                  style="margin-bottom: 16px;"
                >
                  {{ t('adminUsers.logOrderHint') }}
                </n-alert>

                <n-alert
                  v-if="balanceForm.operation === 'both'"
                  type="info"
                  :show-icon="false"
                  style="margin-bottom: 16px;"
                >
                  {{ t('adminUsers.allInOneHint') }}
                </n-alert>

                <n-form-item>
                  <NButton type="primary" :loading="submitting" @click="handleBalanceOperation">
                    {{ t('adminUsers.confirmOperation') }}
                  </NButton>
                </n-form-item>
              </n-form>
            </NSpace>
          </div>
        </n-tab-pane>

        <!-- 积分标签页 -->
        <n-tab-pane v-if="isEdit" name="score" :tab="t('adminUsers.score')">
          <div class="score-management">
            <NSpace vertical size="large">
              <!-- 当前积分显示 -->
              <n-card :title="t('adminUsers.currentScore')" size="small">
                <n-statistic
                  :label="t('adminUsers.userScore')"
                  :value="selectedUser?.score || 0"
                />
              </n-card>

              <!-- 积分操作 -->
              <n-form label-placement="left" :label-width="100">
                <n-form-item :label="t('adminUsers.score')">
                  <n-input-number
                    v-model:value="scoreForm.amount"
                    :placeholder="t('adminUsers.scorePlaceholder')"
                    :step="1"
                  />
                </n-form-item>

                <n-form-item :label="t('adminUsers.operationType')">
                  <n-radio-group v-model:value="scoreForm.operation">
                    <NSpace>
                      <n-radio value="modify">
                        {{ t('adminUsers.modifyScore') }}
                      </n-radio>
                      <n-radio value="log">
                        {{ t('adminUsers.logOnly') }}
                      </n-radio>
                    </NSpace>
                  </n-radio-group>
                </n-form-item>

                <n-form-item :label="t('moneyScore.remark')">
                  <n-input
                    v-model:value="scoreForm.memo"
                    type="textarea"
                    :placeholder="t('adminUsers.enterOperationRemark')"
                    :rows="3"
                  />
                </n-form-item>

                <n-form-item>
                  <NButton type="primary" :loading="submitting" @click="handleScoreOperation">
                    {{ t('adminUsers.confirmOperation') }}
                  </NButton>
                </n-form-item>
              </n-form>
            </NSpace>
          </div>
        </n-tab-pane>

        <n-tab-pane v-if="isEdit" name="withdraw" :tab="t('moneyScore.withdrawRecords')">
          <n-data-table
            :columns="withdrawColumns"
            :data="withdrawData"
            :loading="withdrawLoading"
            :pagination="withdrawPagination"
            size="small"
            :row-key="(row: WithdrawRecord) => row.id"
            @update:page="handleWithdrawPageChange"
            @update:page-size="handleWithdrawPageSizeChange"
          />
        </n-tab-pane>
      </n-tabs>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showUserModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton v-if="activeTab === 'details'" type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? t('adminUsers.update') : t('adminUsers.create') }}
          </NButton>
        </NSpace>
      </template>
    </n-modal>

    <!-- 用户详情模态框 -->
    <n-modal
      v-model:show="showUserDetailModal"
      preset="card"
      :title="t('adminUsers.userDetail')"
      style="width: 700px;"
      :bordered="false"
    >
      <div class="user-details-container">
        <!-- 基本信息区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            {{ t('profile.basicInfo') }}
          </h3>
          <n-descriptions :column="3" bordered size="small">
            <n-descriptions-item :label="t('profile.userId')">
              {{ selectedUser?.id }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.username')">
              {{ selectedUser?.username }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.nickname')">
              {{ selectedUser?.nickname || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.emailAddress')">
              {{ selectedUser?.email || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.phoneNumber')">
              {{ selectedUser?.mobile || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.language')">
              {{ formatLanguage(selectedUser?.language) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.country')">
              {{ selectedUser?.country || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.role')">
              {{ selectedUser?.role === 'admin' ? t('adminUsers.admin') : t('adminUsers.user') }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.userGroup')">
              {{ selectedUser?.group_id || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.gender')">
              {{ selectedUser?.gender === 1 ? t('adminUsers.male') : selectedUser?.gender === 2 ? t('adminUsers.female') : t('adminUsers.unknown') }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.birthday')">
              {{ selectedUser?.birthday ? new Date(Number(selectedUser.birthday) * 1000).toLocaleDateString() : '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- 账户状态区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            {{ t('adminUsers.accountStatus') }}
          </h3>
          <n-descriptions :column="2" bordered size="small">
            <n-descriptions-item :label="t('adminUsers.level')">
              {{ selectedUser?.level || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.status')">
              <NTag :type="(selectedUser?.status === 1) ? 'success' : 'error'">
                {{ (selectedUser?.status === 1) ? t('adminUsers.enabled') : t('adminUsers.disabled') }}
              </NTag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.balance')">
              <n-text type="success">
                ¥{{ formatCurrency(selectedUser?.money) }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.totalPaidAmount')">
              <n-text>
                {{ Number(selectedUser?.total_paid_amount || 0) > 0 ? `¥${formatCurrency(selectedUser?.total_paid_amount)}` : '-' }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.score')">
              <n-text type="info">
                {{ selectedUser?.score || '0' }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.rechargeRetentionRatio')">
              {{ selectedUser ? formatRechargeRetentionRatio(selectedUser) : '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <div class="detail-section">
          <h3 class="section-title">
            {{ t('route.admin-realname') }}
          </h3>
          <n-descriptions :column="2" bordered size="small">
            <n-descriptions-item :label="t('adminUsers.realnameVerifyStatus')">
              <NTag :type="getRealnameStatusType(selectedUserRealname?.status)">
                {{ selectedUserRealname?.has_verification ? getRealnameStatusText(selectedUserRealname?.status) : t('adminUsers.unverified') }}
              </NTag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('realname.realName')">
              {{ selectedUserRealname?.real_name || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('realname.certificateNo')">
              {{ maskCertificateNo(selectedUserRealname?.certificate_no) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('realname.submittedAt')">
              {{ selectedUserRealname?.submitted_at ? new Date(selectedUserRealname.submitted_at * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('realname.reviewedAt')">
              {{ selectedUserRealname?.reviewed_at ? new Date(selectedUserRealname.reviewed_at * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('realname.rejectReason')">
              {{ selectedUserRealname?.reject_reason || '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- 登录信息区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            {{ t('profile.loginInfo') }}
          </h3>
          <n-descriptions :column="2" bordered size="small">
            <n-descriptions-item :label="t('profile.registerTime')">
              {{ selectedUser?.create_time ? new Date(selectedUser.create_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.lastLogin')">
              {{ selectedUser?.last_login_time ? new Date(selectedUser.last_login_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.registerIp')">
              {{ selectedUser?.join_ip || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.lastLoginIp')">
              {{ selectedUser?.last_login_ip || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.loginFailureCount')">
              {{ selectedUser?.login_failure || '0' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('profile.updateTime')">
              {{ selectedUser?.update_time ? new Date(selectedUser.update_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- 其他信息区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            {{ t('adminUsers.otherInfo') }}
          </h3>
          <n-descriptions :column="1" bordered size="small">
            <n-descriptions-item :label="t('adminUsers.apiKey')">
              <NSpace align="center" size="small">
                <n-text code style="font-size: 12px;">
                  {{ selectedUser?.apikey || '-' }}
                </n-text>
                <NButton size="tiny" type="warning" :loading="resettingApikey" @click="handleResetApikey">
                  {{ t('common.reset') }}
                </NButton>
              </NSpace>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.motto')">
              {{ selectedUser?.motto || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.adminRemark')">
              {{ selectedUser?.admin_remark || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.avatar')">
              <NSpace v-if="selectedUser?.avatar" vertical size="small">
                <n-avatar :src="selectedUser.avatar" size="large" />
                <n-text depth="3" style="font-size: 12px;">
                  URL: {{ selectedUser.avatar }}
                </n-text>
              </NSpace>
              <span v-else>-</span>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminUsers.background')">
              {{ selectedUser?.back_ground || '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showUserDetailModal = false">
            {{ t('common.close') }}
          </NButton>
          <NButton type="error" @click="handleShowResetPassword">
            {{ t('adminUsers.resetPassword') }}
          </NButton>
        </NSpace>
      </template>
    </n-modal>

    <n-modal v-model:show="showWithdrawDetailModal" preset="card" :title="t('adminUsers.withdrawDetail')" style="width: 620px;" :bordered="false">
      <template v-if="withdrawDetail">
        <n-descriptions :column="1" bordered label-placement="left">
          <n-descriptions-item :label="t('moneyScore.applicationId')">
            {{ withdrawDetail.id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminRealname.userId')">
            {{ withdrawDetail.user_id }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.withdrawAmount')">
            ¥{{ Number(withdrawDetail.amount).toFixed(2) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.status')">
            <NTag :type="getWithdrawStatusMeta(withdrawDetail.status).type">
              {{ getWithdrawStatusMeta(withdrawDetail.status).label }}
            </NTag>
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.accountType')">
            {{ withdrawDetail.account_type }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.accountName')">
            {{ withdrawDetail.account_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.accountNo')">
            {{ withdrawDetail.account_no }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.realName')">
            {{ withdrawDetail.real_name }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.userRemark')">
            {{ withdrawDetail.remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.reviewRemark')">
            {{ withdrawDetail.review_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.transferRemark')">
            {{ withdrawDetail.transfer_remark || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.createdAt')">
            {{ formatTime(withdrawDetail.create_time) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.reviewedAt')">
            {{ formatTime(withdrawDetail.reviewed_at) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsers.reviewer')">
            {{ getAdminDisplayName(withdrawDetail.reviewed_by) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('moneyScore.paidAt')">
            {{ formatTime(withdrawDetail.paid_at) }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminUsers.payer')">
            {{ getAdminDisplayName(withdrawDetail.paid_by) }}
          </n-descriptions-item>
        </n-descriptions>
      </template>
    </n-modal>

    <!-- 重置密码模态框 -->
    <n-modal
      v-model:show="showResetPasswordModal"
      preset="card"
      :title="t('adminUsers.resetPassword')"
      style="width: 500px;"
      :bordered="false"
    >
      <n-form
        :model="resetPasswordForm"
        :rules="resetPasswordRules"
        label-placement="left"
        :label-width="100"
      >
        <n-form-item :label="t('profile.newPassword')" path="newPassword">
          <n-input
            v-model:value="resetPasswordForm.newPassword"
            type="password"
            :placeholder="t('adminUsers.enterNewPassword')"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item :label="t('profile.confirmPassword')" path="confirmPassword">
          <n-input
            v-model:value="resetPasswordForm.confirmPassword"
            type="password"
            :placeholder="t('profile.enterConfirmPassword')"
            show-password-on="click"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showResetPasswordModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="resettingPassword" @click="handleResetPassword">
            {{ t('adminUsers.confirmReset') }}
          </NButton>
        </NSpace>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.header-card {
  margin-bottom: 16px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
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

.search-card {
  margin-bottom: 16px;
}

.search-actions {
  display: flex;
  align-items: flex-end;
}

.table-card {
  min-height: 400px;
}

/* 用户详情样式 */
.user-details-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-section {
  padding: 0;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #2080f0;
  border-left: 3px solid #2080f0;
  padding-left: 10px;
}

.password-tip {
  font-size: 12px;
  color: #909399;
  font-style: italic;
}

.avatar-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.balance-management,
.score-management {
  padding: 16px 0;
}

.balance-management .n-card,
.score-management .n-card {
  margin-bottom: 16px;
}

.balance-management .n-statistic,
.score-management .n-statistic {
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {

  .header-content {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .search-card :deep(.n-grid) {
    grid-template-columns: repeat(12, 1fr) !important;
  }

  .search-card :deep(.n-form-item-gi) {
    grid-column: span 12 !important;
  }

  .search-actions {
    margin-top: 8px;
  }

  .search-card :deep(.n-space) {
    width: 100%;
    display: flex;
    justify-content: space-between;
    gap: 8px;
  }

  .search-btn,
  .reset-btn {
    flex: 1;
  }

  .search-card :deep(.n-button) {
    flex: 1;
    min-width: 120px;
  }

  /* 操作按钮的响应式处理 */
  .table-card :deep(.n-data-table .n-data-table__td--last) .n-space {
    flex-wrap: wrap;
    gap: 8px;
    justify-content: flex-start;
  }

  .table-card :deep(.n-data-table .n-data-table__td--last) .n-button {
    margin-bottom: 4px;
    margin-right: 4px;
    font-size: 12px !important;
    padding: 0 8px !important;
  }

  /* 改进表格在移动端的显示 */
  .table-card :deep(.n-data-table) {
    overflow-x: auto;
  }

  .table-card :deep(.n-data-table-td) {
    padding: 8px !important;
    white-space: nowrap;
  }

  /* 改进模态框在移动端的显示 */
  :deep(.n-modal-body-wrapper) {
    width: 95vw !important;
    max-width: 600px;
  }
}

@media (max-width: 480px) {
  .header-title {
    font-size: 16px;
  }

  .search-card :deep(.n-button) {
    width: 100%;
    margin-bottom: 8px;
  }

  /* 移动端操作按钮优化 */
  .table-card :deep(.n-data-table .n-data-table__td--last) .n-space {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    width: 100%;
  }

  .table-card :deep(.n-data-table .n-data-table__td--last) .n-button {
    margin: 2px;
    width: 100%;
    padding: 4px 0 !important;
  }

  .table-card :deep(.n-data-table) {
    font-size: 12px;
  }

  /* 进一步缩小模态框内部元素间距 */
  :deep(.n-modal-body) {
    padding: 16px !important;
  }

  :deep(.n-form-item) {
    margin-bottom: 16px !important;
  }
}

/* 分页容器样式 */
.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 12px 0;
  border-top: 1px solid var(--n-border-color);
}

.pagination-info {
  display: flex;
  align-items: center;
}
</style>
