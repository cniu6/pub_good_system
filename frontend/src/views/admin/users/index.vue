<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { NButton, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns, FormRules } from 'naive-ui'
import NovaIcon from '@/components/common/NovaIcon.vue'
import {
  createUser,
  deleteUser,
  fetchAdminUserPage,
  loginAsUser,
  openLoginAsUserWindow,
  resetUserApikey,
  updateAdminUserProfile,
  type AdminUser,
} from '@/service/api/admin/user'
import { addScoreLog, generateNos, operateUserMoney, updateUserScore, type MoneyOperationPayload } from '@/service/api/admin/finance'

const route = useRoute()
const message = useMessage()
const dialog = useDialog()

// 搜索表单
const searchForm = reactive({
  keyword: '',
})

// 角色选项 - 移除"全部角色"选项
const roleOptions = [
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
]

// 状态选项 - 编辑用（使用数字类型）
const userStatusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
]

// 性别选项（使用数字类型）
const genderOptions = [
  { label: '未知', value: 0 },
  { label: '男', value: 1 },
  { label: '女', value: 2 },
]

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10, // 改为默认10个/页
  total: 0,
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
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3-20个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email' as any, message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
  level: [
    { type: 'number' as any, message: '请输入有效的等级', trigger: 'change' },
  ],
  status: [
    {
      required: true,
      validator: (_rule: any, value: any) => {
        if (value === null || value === undefined || value === '') {
          return new Error('请选择状态')
        }
        return true
      },
      trigger: 'change',
    },
  ],
}

// 密码验证规则：新建用户时必填，编辑时可选
const passwordRule = computed(() => {
  if (isEdit.value) {
    return [
      { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' },
    ]
  }
  return [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' },
  ]
})

// 用户详情相关
const showUserDetailModal = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const resettingApikey = ref(false)

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

const orderStatusOptions = [
  { label: '待支付(0)', value: 0 },
  { label: '已支付(1)', value: 1 },
  { label: '已取消(2)', value: 2 },
  { label: '已退款(3)', value: 3 },
  { label: '支付失败(4)', value: 4 },
]

// 积分管理相关
const scoreForm = reactive({
  amount: 0,
  operation: 'modify', // 'modify', 'log', 'both'
  memo: '',
})

// 重置密码表单验证规则
const resetPasswordRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string) => {
        if (value !== resetPasswordForm.newPassword) {
          return new Error('两次输入的密码不一致')
        }
        return true
      },
      trigger: 'blur',
    },
  ],
}

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
    title: '用户名',
    key: 'username',
    width: 120,
  },
  {
    title: '昵称',
    key: 'nickname',
    width: 120,
  },
  {
    title: '用户组',
    key: 'group_id',
    width: 80,
    render: (row: AdminUser) => row.group_id?.toString() || '-',
  },
  {
    title: '邮箱',
    key: 'email',
    width: 200,
    ellipsis: true,
  },
  {
    title: '手机',
    key: 'mobile',
    width: 120,
  },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render: (row: AdminUser) => {
      const roleMap: Record<string, { type: string, label: string }> = {
        admin: { type: 'error', label: '管理员' },
        user: { type: 'success', label: '用户' },
      }
      const role = roleMap[row.role] || { type: 'default', label: row.role }
      return h(NTag, { type: role.type as any }, () => role.label)
    },
  },
  {
    title: '等级',
    key: 'level',
    width: 80,
  },
  {
    title: '余额',
    key: 'money',
    width: 100,
    render: (row: AdminUser) => `¥${(Number(row.money) || 0).toFixed(2)}`,
  },
  {
    title: '积分',
    key: 'score',
    width: 80,
    render: (row: AdminUser) => (Number(row.score) || 0).toString(),
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: AdminUser) => {
      const statusMap: Record<string, { type: string, text: string }> = {
        '1': { type: 'success', text: '启用' },
        '0': { type: 'error', text: '禁用' },
      }
      const statusKey = String(row.status)
      const status = statusMap[statusKey] || { type: 'default', text: '未知' }
      return h(NTag, { type: status.type as any }, { default: () => status.text })
    },
  },
  {
    title: '注册时间',
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
    title: '更新时间',
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
    title: '操作',
    key: 'actions',
    width: 250,
    render: (row: AdminUser) => {
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, {
          size: 'small',
          type: 'info',
          onClick: () => handleViewUserDetail(row),
        }, () => '详情'),
        h(NButton, {
          size: 'small',
          type: 'primary',
          onClick: () => handleEdit(row),
        }, () => '编辑'),
        h(NButton, {
          size: 'small',
          type: 'success',
          onClick: () => handleLoginAsUser(row),
        }, () => '登录'),
        h(NButton, {
          size: 'small',
          type: 'error',
          onClick: () => handleDelete(row.id),
        }, () => '删除'),
      ])
    },
  },
]

// 获取用户数据
async function fetchData() {
  loading.value = true
  try {
    const response: any = await fetchAdminUserPage({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword || undefined,
    })

    if (response.isSuccess) {
      const data = response.data
      if (Array.isArray(data)) {
        userData.value = data
        pagination.total = response.total || 0
      }
      else if (data && data.list) {
        userData.value = data.list || []
        pagination.total = data.total || response.total || 0
      }
      else {
        userData.value = []
        pagination.total = response.total || 0
      }
    }
    else {
      message.error(response.message || '获取用户列表失败')
    }
  }
  catch {
    message.error('获取用户列表失败')
  }
  finally {
    loading.value = false
  }
}

// 查看用户详情
function handleViewUserDetail(user: AdminUser) {
  selectedUser.value = user
  showUserDetailModal.value = true
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
    password: '',
    role: user.role,
    level: user.level,
    status: statusValue,
    avatar: user.avatar || '',
    gender: user.gender || 0,
    birthday: user.birthday ? new Date(user.birthday as any) : null,
    motto: user.motto || '',
  })
  showUserModal.value = true
}

// 删除用户
function handleDelete(userId: number) {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除这个用户吗？此操作不可恢复。',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response: any = await deleteUser(userId)
        if (response.isSuccess) {
          message.success('删除成功')
          fetchData()
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

// 重置API密钥
async function handleResetApikey() {
  if (!selectedUser.value)
    return

  dialog.warning({
    title: '确认重置',
    content: '确定要重置用户的API密钥吗？重置后旧的密钥将失效。',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        resettingApikey.value = true
        const response: any = await resetUserApikey(selectedUser.value!.id)
        if (response.isSuccess) {
          message.success('API密钥重置成功')
          showUserDetailModal.value = false
          fetchData()
        }
        else {
          message.error(response.message || 'API密钥重置失败')
        }
      }
      catch {
        message.error('API密钥重置失败')
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
      message.error('请输入新密码')
      return
    }
    if (resetPasswordForm.newPassword.length < 6) {
      message.error('密码长度不能少于6个字符')
      return
    }
    if (resetPasswordForm.newPassword !== resetPasswordForm.confirmPassword) {
      message.error('两次输入的密码不一致')
      return
    }

    resettingPassword.value = true

    const { resetUserPassword } = await import('@/service/api/admin/user')
    const response: any = await resetUserPassword(selectedUser.value!.id, {
      password: resetPasswordForm.newPassword,
    })

    if (response.isSuccess) {
      message.success('密码重置成功')
      showResetPasswordModal.value = false
      showUserDetailModal.value = false
      fetchData()
    }
    else {
      message.error(response.message || '密码重置失败')
    }
  }
  catch {
    message.error('密码重置失败')
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

      if (userForm.username !== originalUser?.username) {
        changedData.username = userForm.username
      }
      if (userForm.nickname !== (originalUser?.nickname || '')) {
        changedData.nickname = userForm.nickname
      }
      if (userForm.email !== (originalUser?.email || '')) {
        changedData.email = userForm.email
      }
      if (userForm.mobile !== (originalUser?.mobile || '')) {
        changedData.mobile = userForm.mobile
      }
      if (userForm.password && userForm.password.trim()) {
        changedData.password = userForm.password
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

      const originalBirthday = originalUser?.birthday ? new Date(originalUser.birthday as any).getTime() : null
      const formBirthday = userForm.birthday ? userForm.birthday.getTime?.() : null
      if (originalBirthday !== formBirthday) {
        changedData.birthday = userForm.birthday
      }

      if (userForm.motto !== (originalUser?.motto || '')) {
        changedData.motto = userForm.motto
      }

      if (Object.keys(changedData).length === 0) {
        message.warning('没有检测到任何变化')
        submitting.value = false
        return
      }

      const response: any = await updateAdminUserProfile(selectedUser.value?.id as number, changedData)
      if (response.isSuccess) {
        message.success('更新成功')
        showUserModal.value = false
        fetchData()
      }
      else {
        message.error(response.message || '更新失败')
      }
    }
    else {
      const userPayload = { ...userForm }
      const response: any = await createUser(userPayload as any)
      if (response.isSuccess) {
        message.success('创建成功')
        showUserModal.value = false
        fetchData()
      }
      else {
        message.error(response.message || '创建失败')
      }
    }
  }
  catch (error) {
    // eslint-disable-next-line no-console
    console.error('表单验证失败:', error)
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

// 以用户身份登录（在新标签页打开，不污染管理员登录态）
function handleLoginAsUser(user: AdminUser) {
  dialog.warning({
    title: '确认登录',
    content: `确定要以用户 "${user.username}" 的身份登录吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res: any = await loginAsUser(user.id)
      if (res.isSuccess && res.data?.user && res.data?.token) {
        const targetUrl = `/user/dashboard?_t=${Date.now()}`
        openLoginAsUserWindow(res.data.user, res.data.token, res.data.refreshToken, res.data.expiresAt, targetUrl)
        message.success('已在新标签页打开用户后台')
      }
      else {
        message.error(res.message || '登录失败')
      }
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
  // eslint-disable-next-line no-console
  console.log('头像加载失败')
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

  activeTab.value = 'details'
}

// 处理余额操作
async function handleBalanceOperation() {
  if (!selectedUser.value)
    return

  try {
    submitting.value = true

    const isOrder = ['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)
    const needsAmount = balanceForm.operation !== 'order_only'

    if (needsAmount && balanceForm.amount === null) {
      message.warning('金额不能为空')
      return
    }

    if (needsAmount && Number(balanceForm.amount) === 0) {
      message.warning('涉及余额或日志操作时，金额不能为 0')
      return
    }

    if (isOrder && !balanceForm.orderNo) {
      message.warning('涉及订单操作时，订单号不能为空')
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
      message.success('余额操作成功')
      fetchData()
      showUserModal.value = false
    }
    else {
      message.error(response.message || '余额操作失败')
    }
  }
  catch (error) {
    // eslint-disable-next-line no-console
    console.error('余额操作失败:', error)
    message.error('操作失败')
  }
  finally {
    submitting.value = false
  }
}

// 自动填充订单号（从后端生成）
async function handleAutoFillOrderNo() {
  try {
    const res: any = await generateNos()
    if (res.code === 200 || res.code === 0) {
      balanceForm.orderNo = res.data.order_no
      message.success('订单号已自动生成')
    }
    else {
      message.error(res.message || '生成订单号失败')
    }
  }
  catch {
    message.error('生成订单号失败')
  }
}

// 自动填充交易号（从后端生成）
async function handleAutoFillTradeNo() {
  try {
    const res: any = await generateNos()
    if (res.code === 200 || res.code === 0) {
      balanceForm.tradeNo = res.data.trade_no
      message.success('交易号已自动生成')
    }
    else {
      message.error(res.message || '生成交易号失败')
    }
  }
  catch {
    message.error('生成交易号失败')
  }
}

// 处理积分操作
async function handleScoreOperation() {
  if (!selectedUser.value)
    return

  try {
    submitting.value = true

    if (scoreForm.operation === 'modify') {
      const response: any = await updateUserScore(selectedUser.value.id, {
        score: scoreForm.amount,
      })
      if (response.isSuccess) {
        message.success('积分修改成功')
        fetchData()
        showUserModal.value = false
      }
      else {
        message.error(response.message || '积分修改失败')
      }
    }
    else if (scoreForm.operation === 'log') {
      const response: any = await addScoreLog(selectedUser.value.id, {
        score: scoreForm.amount,
        memo: scoreForm.memo,
      })
      if (response.isSuccess) {
        message.success('积分日志添加成功')
        showUserModal.value = false
      }
      else {
        message.error(response.message || '积分日志添加失败')
      }
    }
    else if (scoreForm.operation === 'both') {
      const logResponse: any = await addScoreLog(selectedUser.value.id, {
        score: scoreForm.amount,
        memo: scoreForm.memo,
      })

      if (logResponse.isSuccess) {
        const updateResponse: any = await updateUserScore(selectedUser.value.id, {
          score: scoreForm.amount,
        })

        if (updateResponse.isSuccess) {
          message.success('积分日志添加并修改成功')
          fetchData()
          showUserModal.value = false
        }
        else {
          message.error('日志添加成功，但积分修改失败')
        }
      }
      else {
        message.error(logResponse.message || '积分日志添加失败')
      }
    }
  }
  catch (error) {
    // eslint-disable-next-line no-console
    console.error('积分操作失败:', error)
    message.error('操作失败')
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
          <span>用户管理</span>
        </div>
        <NSpace :wrap="false" :size="12" class="header-actions">
          <NButton @click="handleRefresh">
            <template #icon>
              <NovaIcon icon="icon-park-outline:refresh" />
            </template>
            刷新
          </NButton>
          <NButton type="primary" @click="handleAdd">
            <template #icon>
              <NovaIcon icon="icon-park-outline:plus" />
            </template>
            添加用户
          </NButton>
        </NSpace>
      </div>
    </n-card>

    <!-- 搜索和筛选 -->
    <n-card class="search-card" :bordered="false">
      <n-form :model="searchForm" label-placement="left" :label-width="80">
        <n-grid :cols="24" :x-gap="16" responsive="screen">
          <n-form-item-gi span="24 600:12 800:12" label="关键词">
            <n-input
              v-model:value="searchForm.keyword"
              placeholder="搜索ID/用户名/邮箱/手机/昵称/角色/状态/等级"
              clearable
              @keyup.enter="handleSearch"
            />
          </n-form-item-gi>
          <n-form-item-gi span="24 600:12 800:12" class="search-actions">
            <NSpace justify="center">
              <NButton type="primary" class="search-btn" @click="handleSearch">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:search" />
                </template>
                搜索
              </NButton>
              <NButton class="reset-btn" @click="handleReset">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:refresh" />
                </template>
                重置
              </NButton>
            </NSpace>
          </n-form-item-gi>
        </n-grid>
      </n-form>
    </n-card>

    <!-- 用户列表 -->
    <n-card class="table-card" :bordered="false">
      <n-data-table
        :columns="columns"
        :data="userData"
        :loading="loading"
        :pagination="false"
        :row-key="(row) => row.id"
        :scrollbar-props="{ trigger: 'hover' }"
        :scroll-x="1800"
      />

      <!-- 外部分页组件 -->
      <div class="pagination-container">
        <div class="pagination-info">
          <n-text depth="3">
            共 {{ pagination.total }} 条记录，当前第 {{ pagination.page }} 页，每页显示 {{ pagination.pageSize }} 条
          </n-text>
        </div>
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :item-count="pagination.total"
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
      :title="isEdit ? '编辑用户' : '添加用户'"
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

      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 详情标签页 -->
        <n-tab-pane name="details" tab="详情">
          <n-form
            ref="formRef"
            :model="userForm"
            :rules="rules"
            label-placement="left"
            :label-width="100"
          >
            <n-grid :cols="2" :x-gap="16">
              <n-form-item-gi label="用户名" path="username">
                <n-input
                  v-model:value="userForm.username"
                  placeholder="请输入用户名"
                  :disabled="isEdit"
                />
              </n-form-item-gi>
              <n-form-item-gi label="昵称" path="nickname">
                <n-input
                  v-model:value="userForm.nickname"
                  placeholder="请输入昵称"
                />
              </n-form-item-gi>
              <n-form-item-gi label="邮箱" path="email">
                <n-input
                  v-model:value="userForm.email"
                  placeholder="请输入邮箱"
                  @blur="userForm.email = userForm.email.includes('@') ? userForm.email.split('@')[0] + '@' + userForm.email.split('@')[1].toLowerCase() : userForm.email"
                />
              </n-form-item-gi>
              <n-form-item-gi label="手机" path="mobile">
                <n-input
                  v-model:value="userForm.mobile"
                  placeholder="请输入手机号"
                />
              </n-form-item-gi>
              <n-form-item-gi span="2" label="密码" path="password" :rule="passwordRule">
                <n-input
                  v-model:value="userForm.password"
                  type="password"
                  placeholder="请输入新密码"
                  show-password-on="click"
                />
                <template #feedback>
                  <span class="password-tip">{{ isEdit ? '留空则不修改密码' : '请设置密码' }}</span>
                </template>
              </n-form-item-gi>
              <n-form-item-gi label="角色" path="role">
                <n-select
                  v-model:value="userForm.role"
                  :options="roleOptions"
                  placeholder="选择角色"
                />
              </n-form-item-gi>
              <n-form-item-gi label="等级" path="level">
                <n-input-number
                  v-model:value="userForm.level"
                  placeholder="请输入用户等级"
                  :min="0"
                  :max="100"
                />
              </n-form-item-gi>
              <n-form-item-gi label="状态" path="status">
                <n-select
                  v-model:value="userForm.status"
                  :options="userStatusOptions"
                  placeholder="选择状态"
                />
              </n-form-item-gi>
              <n-form-item-gi label="性别" path="gender">
                <n-select
                  v-model:value="userForm.gender"
                  :options="genderOptions"
                  placeholder="选择性别"
                />
              </n-form-item-gi>
              <n-form-item-gi span="2" label="生日" path="birthday">
                <n-date-picker
                  v-model:value="userForm.birthday"
                  type="date"
                  placeholder="选择生日"
                  clearable
                />
              </n-form-item-gi>
              <n-form-item-gi span="2" label="头像" path="avatar">
                <NSpace vertical>
                  <n-input
                    v-model:value="userForm.avatar"
                    placeholder="请输入头像URL"
                  />
                  <!-- 头像预览 -->
                  <div v-if="userForm.avatar" class="avatar-preview">
                    <n-text depth="3" style="font-size: 12px;">
                      预览：
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
              <n-form-item-gi span="2" label="座右铭" path="motto">
                <n-input
                  v-model:value="userForm.motto"
                  type="textarea"
                  placeholder="请输入座右铭"
                  :rows="3"
                />
              </n-form-item-gi>
            </n-grid>
          </n-form>
        </n-tab-pane>

        <!-- 余额标签页 -->
        <n-tab-pane v-if="isEdit" name="balance" tab="余额">
          <div class="balance-management">
            <NSpace vertical size="large">
              <!-- 当前余额显示 -->
              <n-card title="当前余额" size="small">
                <n-statistic
                  label="用户余额"
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
                <n-form-item label="金额">
                  <n-input-number
                    v-model:value="balanceForm.amount"
                    placeholder="请输入金额（正数增加，负数减少）"
                    :precision="2"
                    :step="0.01"
                  />
                </n-form-item>

                <n-form-item label="操作类型">
                  <n-radio-group v-model:value="balanceForm.operation">
                    <NSpace wrap>
                      <n-radio value="balance_only">仅修改余额</n-radio>
                      <n-radio value="log_only">仅添加日志</n-radio>
                      <n-radio value="order_only">仅操作订单</n-radio>
                      <n-radio value="balance_log">修改余额 + 添加日志</n-radio>
                      <n-radio value="balance_order">修改余额 + 操作订单</n-radio>
                      <n-radio value="log_order">添加日志 + 操作订单</n-radio>
                      <n-radio value="both">余额 + 日志 + 订单</n-radio>
                    </NSpace>
                  </n-radio-group>
                </n-form-item>

                <n-form-item v-if="!['balance_only', 'order_only'].includes(balanceForm.operation)" label="备注">
                  <n-input
                    v-model:value="balanceForm.memo"
                    type="textarea"
                    placeholder="请输入操作备注"
                    :rows="3"
                  />
                </n-form-item>

                <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" label="订单号">
                  <n-input-group>
                    <n-input
                      v-model:value="balanceForm.orderNo"
                      placeholder="请输入订单号（必填）"
                      style="flex: 1"
                    />
                    <NButton type="primary" ghost @click="handleAutoFillOrderNo">
                      自动生成
                    </NButton>
                  </n-input-group>
                </n-form-item>

                <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" label="交易号">
                  <n-input-group>
                    <n-input
                      v-model:value="balanceForm.tradeNo"
                      placeholder="请输入第三方交易号（可选）"
                      style="flex: 1"
                    />
                    <NButton type="primary" ghost @click="handleAutoFillTradeNo">
                      自动生成
                    </NButton>
                  </n-input-group>
                </n-form-item>

                <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" label="订单状态">
                  <n-select
                    v-model:value="balanceForm.orderStatus"
                    :options="orderStatusOptions"
                    placeholder="请选择订单状态"
                  />
                </n-form-item>

                <n-form-item>
                  <NButton type="primary" :loading="submitting" @click="handleBalanceOperation">
                    确定操作
                  </NButton>
                </n-form-item>
              </n-form>
            </NSpace>
          </div>
        </n-tab-pane>

        <!-- 积分标签页 -->
        <n-tab-pane v-if="isEdit" name="score" tab="积分">
          <div class="score-management">
            <NSpace vertical size="large">
              <!-- 当前积分显示 -->
              <n-card title="当前积分" size="small">
                <n-statistic
                  label="用户积分"
                  :value="selectedUser?.score || 0"
                />
              </n-card>

              <!-- 积分操作 -->
              <n-form label-placement="left" :label-width="100">
                <n-form-item label="积分">
                  <n-input-number
                    v-model:value="scoreForm.amount"
                    placeholder="请输入积分（正数增加，负数减少）"
                    :step="1"
                  />
                </n-form-item>

                <n-form-item label="操作类型">
                  <n-radio-group v-model:value="scoreForm.operation">
                    <NSpace>
                      <n-radio value="modify">
                        直接修改
                      </n-radio>
                      <n-radio value="log">
                        直接加日志
                      </n-radio>
                      <n-radio value="both">
                        同时修改+日志
                      </n-radio>
                    </NSpace>
                  </n-radio-group>
                </n-form-item>

                <n-form-item v-if="scoreForm.operation !== 'modify'" label="备注">
                  <n-input
                    v-model:value="scoreForm.memo"
                    type="textarea"
                    placeholder="请输入操作备注"
                    :rows="3"
                  />
                </n-form-item>

                <n-form-item>
                  <NButton type="primary" :loading="submitting" @click="handleScoreOperation">
                    确定操作
                  </NButton>
                </n-form-item>
              </n-form>
            </NSpace>
          </div>
        </n-tab-pane>
      </n-tabs>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showUserModal = false">
            取消
          </NButton>
          <NButton v-if="activeTab === 'details'" type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '更新' : '创建' }}
          </NButton>
        </NSpace>
      </template>
    </n-modal>

    <!-- 用户详情模态框 -->
    <n-modal
      v-model:show="showUserDetailModal"
      preset="card"
      title="用户详情"
      style="width: 700px;"
      :bordered="false"
    >
      <div class="user-details-container">
        <!-- 基本信息区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            基本信息
          </h3>
          <n-descriptions :column="3" bordered size="small">
            <n-descriptions-item label="用户ID">
              {{ selectedUser?.id }}
            </n-descriptions-item>
            <n-descriptions-item label="用户名">
              {{ selectedUser?.username }}
            </n-descriptions-item>
            <n-descriptions-item label="昵称">
              {{ selectedUser?.nickname || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="邮箱">
              {{ selectedUser?.email || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="手机">
              {{ selectedUser?.mobile || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="角色">
              {{ selectedUser?.role === 'admin' ? '管理员' : '用户' }}
            </n-descriptions-item>
            <n-descriptions-item label="用户组">
              {{ selectedUser?.group_id || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="性别">
              {{ selectedUser?.gender === 1 ? '男' : selectedUser?.gender === 2 ? '女' : '未知' }}
            </n-descriptions-item>
            <n-descriptions-item label="生日">
              {{ selectedUser?.birthday ? new Date(selectedUser.birthday).toLocaleDateString() : '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- 账户状态区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            账户状态
          </h3>
          <n-descriptions :column="2" bordered size="small">
            <n-descriptions-item label="等级">
              {{ selectedUser?.level || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <NTag :type="(selectedUser?.status === 1) ? 'success' : 'error'">
                {{ (selectedUser?.status === 1) ? '启用' : '禁用' }}
              </NTag>
            </n-descriptions-item>
            <n-descriptions-item label="余额">
              <n-text type="success">
                ¥{{ selectedUser?.money ? Number(selectedUser.money).toFixed(2) : '0.00' }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item label="积分">
              <n-text type="info">
                {{ selectedUser?.score || '0' }}
              </n-text>
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- 登录信息区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            登录信息
          </h3>
          <n-descriptions :column="2" bordered size="small">
            <n-descriptions-item label="注册时间">
              {{ selectedUser?.create_time ? new Date(selectedUser.create_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="最后登录">
              {{ selectedUser?.last_login_time ? new Date(selectedUser.last_login_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="注册IP">
              {{ selectedUser?.join_ip || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="最后登录IP">
              {{ selectedUser?.last_login_ip || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="登录失败次数">
              {{ selectedUser?.login_failure || '0' }}
            </n-descriptions-item>
            <n-descriptions-item label="更新时间">
              {{ selectedUser?.update_time ? new Date(selectedUser.update_time * 1000).toLocaleString() : '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- 其他信息区域 -->
        <div class="detail-section">
          <h3 class="section-title">
            其他信息
          </h3>
          <n-descriptions :column="1" bordered size="small">
            <n-descriptions-item label="API密钥">
              <NSpace align="center" size="small">
                <n-text code style="font-size: 12px;">
                  {{ selectedUser?.apikey || '-' }}
                </n-text>
                <NButton size="tiny" type="warning" :loading="resettingApikey" @click="handleResetApikey">
                  重置
                </NButton>
              </NSpace>
            </n-descriptions-item>
            <n-descriptions-item label="座右铭">
              {{ selectedUser?.motto || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="头像">
              <NSpace v-if="selectedUser?.avatar" vertical size="small">
                <n-avatar :src="selectedUser.avatar" size="large" />
                <n-text depth="3" style="font-size: 12px;">
                  URL: {{ selectedUser.avatar }}
                </n-text>
              </NSpace>
              <span v-else>-</span>
            </n-descriptions-item>
            <n-descriptions-item label="背景">
              {{ selectedUser?.back_ground || '-' }}
            </n-descriptions-item>
          </n-descriptions>
        </div>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showUserDetailModal = false">
            关闭
          </NButton>
          <NButton type="error" @click="handleShowResetPassword">
            重置密码
          </NButton>
        </NSpace>
      </template>
    </n-modal>

    <!-- 重置密码模态框 -->
    <n-modal
      v-model:show="showResetPasswordModal"
      preset="card"
      title="重置密码"
      style="width: 500px;"
      :bordered="false"
    >
      <n-form
        :model="resetPasswordForm"
        :rules="resetPasswordRules"
        label-placement="left"
        :label-width="100"
      >
        <n-form-item label="新密码" path="newPassword">
          <n-input
            v-model:value="resetPasswordForm.newPassword"
            type="password"
            placeholder="请输入新密码"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="确认密码" path="confirmPassword">
          <n-input
            v-model:value="resetPasswordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password-on="click"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showResetPasswordModal = false">
            取消
          </NButton>
          <NButton type="primary" :loading="resettingPassword" @click="handleResetPassword">
            确定重置
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
