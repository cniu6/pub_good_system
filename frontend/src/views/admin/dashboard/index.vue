<script setup lang="ts">
import { ref, reactive, onMounted, markRaw, h } from 'vue'
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
  { title: '用户名', key: 'username', width: 120 },
  { title: '邮箱', key: 'email', width: 180, ellipsis: { tooltip: true } },
  {
    title: '角色', key: 'role', width: 80,
    render: (row) => h(NTag, { type: row.role === 'admin' ? 'error' : 'info', size: 'small' }, () => row.role === 'admin' ? '管理员' : '用户'),
  },
  {
    title: '状态', key: 'status', width: 80,
    render: (row) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () => row.status === 1 ? '正常' : '禁用'),
  },
  {
    title: '注册时间', key: 'create_time', width: 160,
    render: (row) => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

// 统计卡片（动态值）
const stat_cards = [
  { label: '用户总数', key: 'total_users', icon: markRaw(UserOutlined), color: 'var(--info-color)' },
  { label: '今日新增', key: 'today_new_users', icon: markRaw(TeamOutlined), color: 'var(--success-color)' },
  { label: '7日活跃', key: 'active_users_7d', icon: markRaw(FieldTimeOutlined), color: 'var(--warning-color)' },
  { label: '活跃会话', key: 'active_sessions', icon: markRaw(CheckCircleOutlined), color: 'var(--error-color)' },
]

// 快速操作
const quick_actions = [
  { label: '用户管理', icon: markRaw(UserOutlined), type: 'primary' as const, path: 'users' },
  { label: '余额日志', icon: markRaw(DollarOutlined), type: 'success' as const, path: 'finance/money-logs' },
  { label: '积分日志', icon: markRaw(StarOutlined), type: 'info' as const, path: 'finance/score-logs' },
  { label: '操作日志', icon: markRaw(FileTextOutlined), type: 'warning' as const, path: 'logs' },
  { label: '系统设置', icon: markRaw(SettingOutlined), type: 'default' as const, path: 'settings' },
]

// 获取仪表盘数据
async function fetchDashboard() {
  loading.value = true
  try {
    const res = await adminApi.dashboard.getStatistics()
    if (res.code === 200 && res.data) {
      const stats = res.data.statistics
      if (stats) {
        Object.assign(statistics, stats)
      }
      if (res.data.recent_users) {
        recentUsers.value = res.data.recent_users
      }
    }
  } catch (error) {
    console.error('获取仪表盘数据失败:', error)
  } finally {
    loading.value = false
  }
}

function go_to(sub_path: string) {
  router.push(`/${sub_path}`)
}

function handleRefresh() {
  fetchDashboard()
  message.success('数据已刷新')
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
            <n-text strong>欢迎回来，管理员</n-text>
            <n-text depth="3">这里是系统运行概览，一切正常运行中。</n-text>
          </n-flex>
        </n-flex>
        <n-flex :size="8">
          <n-button :loading="loading" @click="handleRefresh">刷新数据</n-button>
          <n-button type="primary" @click="go_to('users')">用户管理</n-button>
          <n-button @click="go_to('settings')">系统设置</n-button>
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
        <n-card title="快速操作" hoverable>
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
        <n-card title="系统信息" hoverable>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item label="系统版本">
              <n-tag size="small" type="info">v1.0.0</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="后端框架">Go 1.24 + Gin</n-descriptions-item>
            <n-descriptions-item label="前端框架">Vue 3 + Naive UI</n-descriptions-item>
            <n-descriptions-item label="运行环境">
              <n-tag size="small" :type="mode === 'production' ? 'success' : 'warning'">{{ mode }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="操作日志">
              总 {{ statistics.total_operation_logs }} 条，今日 {{ statistics.today_operation_logs }} 条
            </n-descriptions-item>
            <n-descriptions-item label="余额/积分日志">
              余额 {{ statistics.total_money_logs }} 条 / 积分 {{ statistics.total_score_logs }} 条
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <!-- 最近注册用户 -->
      <n-gi :span="12">
        <n-card title="最近注册用户" hoverable>
          <template #header-extra>
            <n-button type="primary" quaternary @click="go_to('users')">查看全部</n-button>
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
            <n-empty v-if="!loading && recentUsers.length === 0" description="暂无用户数据" />
          </n-spin>
        </n-card>
      </n-gi>
    </n-grid>
  </n-space>
</template>

<style scoped></style>
