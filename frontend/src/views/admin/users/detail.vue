<template>
  <div class="admin-user-detail-page">
    <!-- 页面头部 -->
    <n-card :bordered="false" class="header-card">
      <div class="header-content">
        <div class="header-title">
          <span class="title-text">用户详情</span>
        </div>
        <n-space>
          <n-button @click="$router.back()">返回</n-button>
          <n-button @click="fetchUser">刷新</n-button>
        </n-space>
      </div>
    </n-card>

    <!-- 用户信息卡片 -->
    <n-card :bordered="false" class="user-info-card">
      <n-spin :show="loading">
        <div v-if="user" class="user-info-content">
          <div class="user-avatar">
            <n-avatar :size="72" :src="user.avatar || undefined">
              {{ user.username?.charAt(0).toUpperCase() }}
            </n-avatar>
          </div>
          <div class="user-details">
            <div class="user-name">{{ user.username }} <span class="user-id">#{{ user.id }}</span></div>
            <div class="user-meta">
              <n-tag :type="user.status === 1 ? 'success' : 'error'" size="small">
                {{ user.status === 1 ? '正常' : '禁用' }}
              </n-tag>
              <n-tag type="info" size="small">{{ user.role === 'admin' ? '管理员' : '用户' }}</n-tag>
              <n-tag type="warning" size="small">等级 {{ user.level || 0 }}</n-tag>
            </div>
            <div class="user-contact">
              <span v-if="user.email" class="contact-item">📧 {{ user.email }}</span>
              <span v-if="user.mobile" class="contact-item">📱 {{ user.mobile }}</span>
            </div>
          </div>
          <div class="user-stats">
            <div class="stat-item">
              <div class="stat-value money">¥{{ (Number(user.money) || 0).toFixed(2) }}</div>
              <div class="stat-label">余额</div>
            </div>
            <div class="stat-item">
              <div class="stat-value score">{{ Number(user.score) || 0 }}</div>
              <div class="stat-label">积分</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ user.create_time ? new Date(user.create_time * 1000).toLocaleDateString() : '-' }}</div>
              <div class="stat-label">注册时间</div>
            </div>
          </div>
        </div>
        <n-empty v-else-if="!loading" description="未找到用户信息" />
      </n-spin>
    </n-card>

    <!-- 数据标签页 -->
    <n-card :bordered="false" class="data-tabs-card">
      <n-tabs type="line" animated @update:value="handleTabChange">
        <n-tab-pane name="basic" tab="基本信息">
          <div v-if="user" class="user-info-sections">
            <!-- 基础信息 -->
            <n-card title="基础信息" size="small">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="用户ID">{{ user.id }}</n-descriptions-item>
                <n-descriptions-item label="用户名">{{ user.username }}</n-descriptions-item>
                <n-descriptions-item label="昵称">{{ user.nickname || '-' }}</n-descriptions-item>
                <n-descriptions-item label="邮箱">{{ user.email || '-' }}</n-descriptions-item>
                <n-descriptions-item label="手机">{{ user.mobile || '-' }}</n-descriptions-item>
                <n-descriptions-item label="角色">{{ user.role === 'admin' ? '管理员' : '用户' }}</n-descriptions-item>
                <n-descriptions-item label="等级">{{ user.level || 0 }}</n-descriptions-item>
                <n-descriptions-item label="状态">
                  <n-tag :type="user.status === 1 ? 'success' : 'error'" size="small">
                    {{ user.status === 1 ? '正常' : '禁用' }}
                  </n-tag>
                </n-descriptions-item>
                <n-descriptions-item label="性别">{{ ['未知', '男', '女'][user.gender] || '未知' }}</n-descriptions-item>
                <n-descriptions-item label="座右铭">{{ user.motto || '-' }}</n-descriptions-item>
              </n-descriptions>
            </n-card>

            <!-- 资产信息 -->
            <n-card title="资产信息" size="small">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="余额">
                  <span class="money">¥{{ user.money ? Number(user.money).toFixed(2) : '0.00' }}</span>
                </n-descriptions-item>
                <n-descriptions-item label="积分">
                  <span class="score">{{ user.score || '0' }}</span>
                </n-descriptions-item>
                <n-descriptions-item label="API密钥" :span="2">
                  <n-text code>{{ user.apikey || '-' }}</n-text>
                </n-descriptions-item>
              </n-descriptions>
            </n-card>

            <!-- 登录信息 -->
            <n-card title="登录信息" size="small">
              <n-descriptions :column="2" bordered>
                <n-descriptions-item label="注册时间">
                  {{ user.create_time ? new Date(user.create_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="最后登录">
                  {{ user.last_login_time ? new Date(user.last_login_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
                <n-descriptions-item label="注册IP">{{ user.join_ip || '-' }}</n-descriptions-item>
                <n-descriptions-item label="最后登录IP">{{ user.last_login_ip || '-' }}</n-descriptions-item>
                <n-descriptions-item label="登录失败次数">{{ user.login_failure || '0' }}</n-descriptions-item>
                <n-descriptions-item label="更新时间">
                  {{ user.update_time ? new Date(user.update_time * 1000).toLocaleString() : '-' }}
                </n-descriptions-item>
              </n-descriptions>
            </n-card>
          </div>

          <div class="action-buttons">
            <n-button type="primary" @click="handleEdit">编辑用户</n-button>
            <n-button type="warning" @click="handleToggleStatus">
              {{ user?.status === 1 ? '禁用用户' : '启用用户' }}
            </n-button>
            <n-button type="info" @click="handleResetApikey">重置API密钥</n-button>
            <n-button type="info" @click="showResetPasswordModal = true; newPassword = ''">重置密码</n-button>
            <n-button type="success" @click="handleLoginAs">登录为此用户</n-button>
            <n-button type="error" @click="handleDelete">删除用户</n-button>
          </div>
        </n-tab-pane>

        <n-tab-pane name="money" tab="余额记录">
          <n-data-table
            :columns="moneyColumns"
            :data="moneyData"
            :loading="moneyLoading"
            :pagination="moneyPagination"
            size="small"
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
            size="small"
            @update:page="handleScorePageChange"
            @update:page-size="handleScorePageSizeChange"
          />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 重置密码对话框 -->
    <n-modal v-model:show="showResetPasswordModal" preset="dialog" title="重置密码" positive-text="确认" negative-text="取消" @positive-click="confirmResetPassword">
      <div style="margin-bottom: 16px;">
        <p>您正在重置用户 <strong>{{ user?.username }}</strong> (ID: {{ user?.id }}) 的密码</p>
        <n-form-item label="新密码" required>
          <n-input v-model:value="newPassword" type="password" placeholder="请输入新密码（至少6位）" show-password-on="click" />
        </n-form-item>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, useDialog, NTag } from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import { adminMoneyLogApi, adminScoreLogApi } from '@/service/api/admin/user'
import { local } from '@/utils'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const userId = ref(Number(route.params.id))
const loading = ref(false)
const user = ref<any>(null)

const showResetPasswordModal = ref(false)
const newPassword = ref('')

// 余额记录
const moneyLoading = ref(false)
const moneyData = ref<any[]>([])
const moneyPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 积分记录
const scoreLoading = ref(false)
const scoreData = ref<any[]>([])
const scorePagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 余额记录表格列
const moneyColumns = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: '变动金额',
    key: 'money',
    width: 120,
    render: (row: any) => {
      const money = Number(row.money) || 0
      return h('span', { style: { color: money > 0 ? '#18a058' : '#d03050', fontWeight: '500' } },
        `${money > 0 ? '+' : ''}¥${money.toFixed(2)}`)
    },
  },
  {
    title: '变动前', key: 'before', width: 120,
    render: (row: any) => `¥${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: '变动后', key: 'after', width: 120,
    render: (row: any) => `¥${(Number(row.after) || 0).toFixed(2)}`,
  },
  { title: '备注', key: 'memo', ellipsis: { tooltip: true } },
  {
    title: '时间', key: 'create_time', width: 170,
    render: (row: any) => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
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
      return h('span', { style: { color: score > 0 ? '#18a058' : '#d03050', fontWeight: '500' } },
        `${score > 0 ? '+' : ''}${score}`)
    },
  },
  {
    title: '变动前', key: 'before', width: 120,
    render: (row: any) => String(Number(row.before) || 0),
  },
  {
    title: '变动后', key: 'after', width: 120,
    render: (row: any) => String(Number(row.after) || 0),
  },
  { title: '备注', key: 'memo', ellipsis: { tooltip: true } },
  {
    title: '时间', key: 'create_time', width: 170,
    render: (row: any) => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

// 获取用户信息
async function fetchUser() {
  if (!userId.value) return
  loading.value = true
  try {
    const res = await adminApi.user.detail(userId.value)
    if (res.code === 200 && res.data?.user) {
      user.value = res.data.user
    } else {
      message.error('获取用户信息失败')
    }
  } catch (error) {
    message.error('获取用户信息失败')
  } finally {
    loading.value = false
  }
}

// 获取余额记录
async function fetchMoneyData() {
  if (!userId.value) return
  moneyLoading.value = true
  try {
    const res = await adminMoneyLogApi.list({
      page: moneyPagination.page,
      page_size: moneyPagination.pageSize,
      user_id: userId.value,
    })
    if (res.isSuccess && res.data) {
      moneyData.value = res.data.list || []
      moneyPagination.itemCount = res.data.total || 0
    }
  } catch (error) {
    message.error('获取余额记录失败')
  } finally {
    moneyLoading.value = false
  }
}

// 获取积分记录
async function fetchScoreData() {
  if (!userId.value) return
  scoreLoading.value = true
  try {
    const res = await adminScoreLogApi.list({
      page: scorePagination.page,
      page_size: scorePagination.pageSize,
      user_id: userId.value,
    })
    if (res.isSuccess && res.data) {
      scoreData.value = res.data.list || []
      scorePagination.itemCount = res.data.total || 0
    }
  } catch (error) {
    message.error('获取积分记录失败')
  } finally {
    scoreLoading.value = false
  }
}

function handleTabChange(tabName: string) {
  if (tabName === 'money') fetchMoneyData()
  else if (tabName === 'score') fetchScoreData()
}

function handleMoneyPageChange(page: number) {
  moneyPagination.page = page
  fetchMoneyData()
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

// 编辑用户 - 跳回用户列表并带上编辑参数
function handleEdit() {
  router.push({ path: '/users', query: { edit: userId.value } })
}

// 切换用户状态
function handleToggleStatus() {
  if (!user.value) return
  const newStatus = user.value.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  dialog.warning({
    title: `确认${action}`,
    content: `确定要${action}用户 "${user.value.username}" 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminApi.user.updateStatus(user.value.id, newStatus)
        if (res.code === 200) {
          message.success(`${action}成功`)
          fetchUser()
        } else {
          message.error(res.message || `${action}失败`)
        }
      } catch (error) {
        message.error(`${action}失败`)
      }
    },
  })
}

// 重置API密钥
function handleResetApikey() {
  if (!user.value) return
  dialog.warning({
    title: '确认重置',
    content: '确定要重置用户的API密钥吗？重置后旧的密钥将失效。',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminApi.user.resetApiKey(user.value.id)
        if (res.code === 200) {
          message.success('API密钥重置成功')
          fetchUser()
        } else {
          message.error(res.message || '重置失败')
        }
      } catch (error) {
        message.error('重置API密钥失败')
      }
    },
  })
}

// 确认重置密码
async function confirmResetPassword() {
  if (!user.value || !newPassword.value) {
    message.error('请输入新密码')
    return false
  }
  if (newPassword.value.length < 6) {
    message.error('密码至少6位')
    return false
  }
  try {
    const res = await adminApi.user.resetPassword(user.value.id, newPassword.value)
    if (res.code === 200) {
      message.success(`用户 ${user.value.username} 的密码已重置`)
      showResetPasswordModal.value = false
    } else {
      message.error(res.message || '密码重置失败')
    }
  } catch (error) {
    message.error('密码重置失败')
  }
  return false
}

// 登录为此用户
function handleLoginAs() {
  if (!user.value) return
  dialog.warning({
    title: '确认登录',
    content: `确定要以用户 "${user.value.username}" 的身份登录吗？将在新标签页打开用户面板。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminApi.user.loginAsUser(user.value.id)
        if (res.code === 200 && res.data) {
          const token = res.data.token
          const userData = res.data.user
          if (!token) {
            message.error('获取用户 token 失败')
            return
          }
          local.set('accessToken', token)
          local.set('userInfo', {
            userName: userData.username,
            role: [userData.role],
            accessToken: token,
          })
          local.set('role', userData.role || 'user')
          message.success(`正在以 ${user.value.username} 身份打开用户面板...`)
          setTimeout(() => {
            window.open('/', '_blank')
          }, 500)
        } else {
          message.error(res.message || '登录失败')
        }
      } catch (error) {
        message.error('登录失败')
      }
    },
  })
}

// 删除用户
function handleDelete() {
  if (!user.value) return
  dialog.warning({
    title: '确认删除',
    content: `确定要删除用户 "${user.value.username}" 吗？此操作不可恢复。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminApi.user.delete(user.value.id)
        if (res.code === 200) {
          message.success('删除成功')
          router.push('/users')
        } else {
          message.error(res.message || '删除失败')
        }
      } catch (error) {
        message.error('删除失败')
      }
    },
  })
}

onMounted(() => {
  fetchUser()
})
</script>

<style scoped>
.admin-user-detail-page {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-text {
  font-size: 18px;
  font-weight: 600;
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
  color: #666;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.user-stats {
  display: flex;
  gap: 32px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
}

.stat-value.money { color: #18a058; }
.stat-value.score { color: #2080f0; }

.stat-label {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.money { font-weight: bold; color: #18a058; font-size: 16px; }
.score { font-weight: bold; color: #2080f0; font-size: 16px; }

.user-info-sections {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.data-tabs-card {
  min-height: 400px;
}

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
</style>
