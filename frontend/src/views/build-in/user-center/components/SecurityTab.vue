<script setup lang="ts">
import { useAuthStore } from '@/store'
import { fetchUserSessions, revokeSession, revokeAllSessions, deactivateAccount, fetchUserStats } from '@/service'
import NovaIcon from '@/components/common/NovaIcon.vue'

const authStore = useAuthStore()

const sessions = ref<any[]>([])
const sessionsLoading = ref(false)
const stats = ref<any>(null)

const showDeactivateModal = ref(false)
const deactivateForm = ref({
  password: '',
  reason: '',
})
const deactivating = ref(false)

async function loadSessions() {
  sessionsLoading.value = true
  try {
    const response = await fetchUserSessions()
    if (response.isSuccess && response.data) {
      sessions.value = Array.isArray(response.data) ? response.data : []
    }
  }
  catch (error) {
    console.error('获取会话列表失败', error)
  }
  finally {
    sessionsLoading.value = false
  }
}

async function loadStats() {
  try {
    const response = await fetchUserStats()
    if (response.isSuccess && response.data) {
      stats.value = response.data
    }
  }
  catch (error) {
    console.error('获取统计失败', error)
  }
}

async function handleRevokeSession(sessionId: number) {
  try {
    const response = await revokeSession(sessionId)
    if (response.isSuccess) {
      window.$message.success('已踢出该会话')
      loadSessions()
    }
    else {
      window.$message.error(response.message || '操作失败')
    }
  }
  catch (error) {
    window.$message.error(`操作失败: ${error}`)
  }
}

async function handleRevokeAll() {
  window.$dialog.warning({
    title: '踢出所有其他会话',
    content: '确定要踢出除当前会话外的所有登录会话吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response = await revokeAllSessions()
        if (response.isSuccess) {
          window.$message.success('已踢出所有其他会话')
          loadSessions()
        }
        else {
          window.$message.error(response.message || '操作失败')
        }
      }
      catch (error) {
        window.$message.error(`操作失败: ${error}`)
      }
    },
  })
}

async function handleDeactivate() {
  if (!deactivateForm.value.password) {
    window.$message.error('请输入密码确认')
    return
  }
  deactivating.value = true
  try {
    const response = await deactivateAccount({
      password: deactivateForm.value.password,
      reason: deactivateForm.value.reason,
    })
    if (response.isSuccess) {
      window.$message.success('账号已注销')
      showDeactivateModal.value = false
      setTimeout(() => {
        authStore.logout()
      }, 1500)
    }
    else {
      window.$message.error(response.message || '注销失败')
    }
  }
  catch (error) {
    window.$message.error(`注销失败: ${error}`)
  }
  finally {
    deactivating.value = false
  }
}

function formatTime(timestamp: number) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp * 1000).toLocaleString()
}

onMounted(() => {
  loadSessions()
  loadStats()
})
</script>

<template>
  <div class="p-4">
    <n-space vertical size="large">
      <!-- 登录统计 -->
      <div>
        <n-h4>登录统计</n-h4>
        <n-divider />
        <n-grid cols="2 m:4" :x-gap="16" :y-gap="16" responsive="screen">
          <n-grid-item>
            <n-statistic label="加入天数" :value="stats?.daysJoined || 0">
              <template #suffix>
                天
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="登录次数" :value="stats?.loginCount || 0">
              <template #suffix>
                次
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="账户余额">
              <template #default>
                ¥{{ stats?.money ? Number(stats.money).toFixed(2) : '0.00' }}
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="积分" :value="stats?.score || 0" />
          </n-grid-item>
        </n-grid>
      </div>

      <n-divider />

      <!-- 登录设备管理 -->
      <div>
        <div class="section-header">
          <n-h4>登录设备管理</n-h4>
          <n-space>
            <n-button size="small" @click="loadSessions">
              <template #icon>
                <NovaIcon icon="icon-park-outline:refresh" :size="14" />
              </template>
              刷新
            </n-button>
            <n-button size="small" type="warning" @click="handleRevokeAll">
              踢出所有其他设备
            </n-button>
          </n-space>
        </div>
        <n-divider />
        <n-spin :show="sessionsLoading">
          <n-space v-if="sessions.length > 0" vertical>
            <div v-for="session in sessions" :key="session.id" class="session-item">
              <div class="session-info">
                <div class="session-device">
                  <NovaIcon icon="icon-park-outline:computer" :size="16" class="mr-1" />
                  {{ session.device || '未知设备' }}
                </div>
                <n-text depth="3" class="session-detail">
                  IP: {{ session.ip || '未知' }} · 登录时间: {{ formatTime(session.login_at) }}
                </n-text>
              </div>
              <n-button size="small" type="error" @click="handleRevokeSession(session.id)">
                踢出
              </n-button>
            </div>
          </n-space>
          <n-empty v-else description="暂无登录会话记录" />
        </n-spin>
      </div>

      <n-divider />

      <!-- 账号注销 -->
      <div>
        <n-h4>危险操作</n-h4>
        <n-divider />
        <div class="danger-zone">
          <div class="danger-info">
            <span class="danger-label">注销账号</span>
            <span class="danger-desc">注销后账号将无法恢复，请谨慎操作</span>
          </div>
          <n-button type="error" @click="showDeactivateModal = true">
            注销账号
          </n-button>
        </div>
      </div>
    </n-space>

    <!-- 注销确认弹窗 -->
    <n-modal v-model:show="showDeactivateModal" preset="dialog" title="注销账号" type="error">
      <n-alert type="error" class="mb-4">
        注销账号后，您的所有数据将被永久删除且无法恢复。请确认此操作。
      </n-alert>
      <n-form :model="deactivateForm" label-placement="left" label-width="100px">
        <n-form-item label="确认密码" required>
          <n-input
            v-model:value="deactivateForm.password"
            type="password"
            placeholder="请输入当前密码以确认"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="注销原因">
          <n-input
            v-model:value="deactivateForm.reason"
            type="textarea"
            placeholder="可选：告诉我们您注销的原因"
            :rows="3"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showDeactivateModal = false">
            取消
          </n-button>
          <n-button type="error" :loading="deactivating" @click="handleDeactivate">
            确认注销
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.session-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  background: var(--n-color);
}

.session-info {
  flex: 1;
}

.session-device {
  display: flex;
  align-items: center;
  font-weight: 500;
  margin-bottom: 4px;
}

.session-detail {
  font-size: 12px;
}

.danger-zone {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--n-error-color);
  border-radius: 6px;
  background: var(--n-color);
}

.danger-info {
  flex: 1;
}

.danger-label {
  display: block;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--n-error-color);
}

.danger-desc {
  color: var(--n-text-color-disabled);
  font-size: 14px;
}

@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .session-item,
  .danger-zone {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
