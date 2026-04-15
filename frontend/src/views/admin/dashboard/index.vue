<script setup lang="ts">
import { ref, reactive, onMounted, markRaw, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import {
  NGrid,
  NGi,
  NCard,
  NStatistic,
  NNumberAnimation,
  NIcon,
  NIconWrapper,
  NSpace,
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NThing,
  NFlex,
  NText,
  NEl,
  NDataTable,
  NEmpty,
  NSpin,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  UserOutlined,
  SettingOutlined,
  FileTextOutlined,
  TeamOutlined,
  CheckCircleOutlined,
  DollarOutlined,
  StarOutlined,
  FieldTimeOutlined,
} from '@vicons/antd'
import { adminApi } from '@/service/api/admin'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const mode = import.meta.env.MODE
const loading = ref(false)

// 统计数据（从后端获取）
const statistics = reactive({
  total_users: 0,
  today_new_users: 0,
  active_users_7d: 0,
  total_money_logs: 0,
  total_score_logs: 0,
  total_operation_logs: 0,
  today_operation_logs: 0,
  active_sessions: 0,
})

// 最近注册用户
const recentUsers = ref<any[]>([])

// 用户表格列
const userColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 60 },
  { title: t('adminDashboard.username'), key: 'username', width: 120 },
  { title: t('adminDashboard.email'), key: 'email', width: 180, ellipsis: { tooltip: true } },
  {
    title: t('adminDashboard.role'), key: 'role', width: 80,
    render: (row) => h(NTag, { type: row.role === 'admin' ? 'error' : 'info', size: 'small' }, () => row.role === 'admin' ? t('adminDashboard.admin') : t('adminDashboard.user')),
  },
  {
    title: t('adminDashboard.status'), key: 'status', width: 80,
    render: (row) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () => row.status === 1 ? t('adminDashboard.normal') : t('adminDashboard.disabled')),
  },
  {
    title: t('adminDashboard.registerTime'), key: 'create_time', width: 160,
    render: (row) => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

// 统计卡片（动态值）
const stat_cards = [
  { label: t('adminDashboard.totalUsers'), key: 'total_users', icon: markRaw(UserOutlined), color: 'var(--info-color)' },
  { label: t('adminDashboard.todayNewUsers'), key: 'today_new_users', icon: markRaw(TeamOutlined), color: 'var(--success-color)' },
  { label: t('adminDashboard.activeUsers7d'), key: 'active_users_7d', icon: markRaw(FieldTimeOutlined), color: 'var(--warning-color)' },
  { label: t('adminDashboard.activeSessions'), key: 'active_sessions', icon: markRaw(CheckCircleOutlined), color: 'var(--error-color)' },
]

// 快速操作
const quick_actions = [
  { label: t('adminDashboard.userManagement'), icon: markRaw(UserOutlined), type: 'primary' as const, path: 'users' },
  { label: t('adminDashboard.moneyLogs'), icon: markRaw(DollarOutlined), type: 'success' as const, path: 'finance/money-logs' },
  { label: t('adminDashboard.scoreLogs'), icon: markRaw(StarOutlined), type: 'info' as const, path: 'finance/score-logs' },
  { label: t('adminDashboard.operationLogs'), icon: markRaw(FileTextOutlined), type: 'warning' as const, path: 'logs' },
  { label: t('adminDashboard.systemSettings'), icon: markRaw(SettingOutlined), type: 'default' as const, path: 'settings' },
]

// 获取仪表盘数据
async function fetchDashboard() {
  loading.value = true
  try {
    const res = await adminApi.dashboard.getStatistics()
    if (res.isSuccess && res.data) {
      const stats = res.data.statistics
      if (stats) {
        Object.assign(statistics, stats)
      }
      if (res.data.recent_users) {
        recentUsers.value = res.data.recent_users
      }
    }
  } catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminDashboard] fetch failed', error)
  } finally {
    loading.value = false
  }
}

function go_to(sub_path: string) {
  router.push(`/${sub_path}`)
}

function handleRefresh() {
  fetchDashboard()
  message.success(t('adminDashboard.dataRefreshed'))
}

onMounted(() => {
  fetchDashboard()
})
</script>

<template>
  <n-space vertical :size="16">
    <!-- 欢迎横幅 -->
    <n-card hoverable>
      <n-flex justify="space-between" align="center" wrap :size="16">
        <n-flex align="center" :size="16">
          <n-icon-wrapper :size="48" :border-radius="12" color="var(--success-color)">
            <n-icon :size="26" color="#fff">
              <CheckCircleOutlined />
            </n-icon>
          </n-icon-wrapper>
          <n-flex vertical>
            <n-text strong>{{ t('adminDashboard.welcomeBack') }}</n-text>
            <n-text depth="3">{{ t('adminDashboard.systemOverview') }}</n-text>
          </n-flex>
        </n-flex>
        <n-flex :size="8">
          <n-button :loading="loading" @click="handleRefresh">{{ t('adminDashboard.refreshData') }}</n-button>
          <n-button type="primary" @click="go_to('users')">{{ t('adminDashboard.userManagement') }}</n-button>
          <n-button @click="go_to('settings')">{{ t('adminDashboard.systemSettings') }}</n-button>
        </n-flex>
      </n-flex>
    </n-card>

    <!-- 统计卡片 -->
    <n-grid :x-gap="16" :y-gap="16" :cols="4" item-responsive responsive="screen">
      <n-gi v-for="card in stat_cards" :key="card.label" span="4 s:2 m:1">
        <n-card hoverable>
          <n-thing>
            <template #avatar>
              <n-el>
                <n-icon-wrapper :size="46" :color="card.color" :border-radius="12">
                  <n-icon :size="24" color="#fff">
                    <component :is="card.icon" />
                  </n-icon>
                </n-icon-wrapper>
              </n-el>
            </template>
            <template #header>
              <n-statistic :label="card.label">
                <n-number-animation :from="0" :to="(statistics as any)[card.key]" show-separator />
              </n-statistic>
            </template>
          </n-thing>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 下半区域 -->
    <n-grid :x-gap="16" :y-gap="16" :cols="12" item-responsive responsive="screen">
      <!-- 快速操作 + 日志统计 -->
      <n-gi span="12 m:6">
        <n-card :title="t('adminDashboard.quickActions')" hoverable>
          <n-grid :x-gap="12" :y-gap="12" :cols="3" item-responsive>
            <n-gi v-for="action in quick_actions" :key="action.label" span="3 s:1">
              <n-button block :type="action.type" ghost size="large" @click="go_to(action.path)">
                <template #icon>
                  <n-icon><component :is="action.icon" /></n-icon>
                </template>
                {{ action.label }}
              </n-button>
            </n-gi>
          </n-grid>
        </n-card>
      </n-gi>

      <!-- 系统信息 -->
      <n-gi span="12 m:6">
        <n-card :title="t('adminDashboard.systemInfo')" hoverable>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item :label="t('adminDashboard.systemVersion')">
              <n-tag size="small" type="info">v1.0.0</n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.backendFramework')">Go 1.24 + Gin</n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.frontendFramework')">Vue 3 + Naive UI</n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.environment')">
              <n-tag size="small" :type="mode === 'production' ? 'success' : 'warning'">{{ mode }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.operationLogs')">
              {{ t('adminDashboard.operationLogsSummary', { total: statistics.total_operation_logs, today: statistics.today_operation_logs }) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.moneyScoreLogs')">
              {{ t('adminDashboard.moneyScoreLogsSummary', { money: statistics.total_money_logs, score: statistics.total_score_logs }) }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <!-- 最近注册用户 -->
      <n-gi :span="12">
        <n-card :title="t('adminDashboard.recentUsers')" hoverable>
          <template #header-extra>
            <n-button type="primary" quaternary @click="go_to('users')">{{ t('adminDashboard.viewAll') }}</n-button>
          </template>
          <n-spin :show="loading">
            <n-data-table
              :columns="userColumns"
              :data="recentUsers"
              :bordered="false"
              :single-line="false"
              size="small"
              :pagination="false"
            />
            <n-empty v-if="!loading && recentUsers.length === 0" :description="t('adminDashboard.noUserData')" />
          </n-spin>
        </n-card>
      </n-gi>
    </n-grid>
  </n-space>
</template>

<style scoped></style>
