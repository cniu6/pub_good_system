<template>
  <div class="admin-dashboard p-4">
    <!-- 欢迎横幅 -->
    <n-card :bordered="false" class="welcome-card mb-4">
      <div class="welcome-inner">
        <div class="welcome-text">
          <h2 class="welcome-title">欢迎回来，管理员 👋</h2>
          <p class="welcome-desc">这里是系统运行概览，一切正常运行中。</p>
        </div>
        <div class="welcome-actions">
          <n-button type="primary" @click="go_to('users')">用户管理</n-button>
          <n-button @click="go_to('settings')">系统设置</n-button>
        </div>
      </div>
    </n-card>

    <!-- 统计卡片 -->
    <n-grid :cols="4" :x-gap="16" :y-gap="16" item-responsive class="mb-4">
      <n-grid-item v-for="card in stat_cards" :key="card.label" span="4 s:2 m:1">
        <n-card :bordered="false" class="stat-card" hoverable>
          <div class="stat-inner">
            <div class="stat-icon-wrap" :style="{ background: card.bg }">
              <n-icon :size="24" :color="card.color">
                <component :is="card.icon" />
              </n-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">
                <n-number-animation :from="0" :to="card.value" show-separator />
                <span v-if="card.suffix" class="stat-suffix">{{ card.suffix }}</span>
              </div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
          <div v-if="card.trend" class="stat-trend" :class="card.trend > 0 ? 'up' : 'down'">
            {{ card.trend > 0 ? '↑' : '↓' }} {{ Math.abs(card.trend) }}% 较昨日
          </div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 下半区域 -->
    <n-grid :cols="2" :x-gap="16" :y-gap="16" item-responsive>
      <!-- 快速操作 -->
      <n-grid-item span="2 m:1">
        <n-card title="快速操作" :bordered="false" class="h-full" hoverable>
          <n-grid :cols="2" :x-gap="12" :y-gap="12">
            <n-grid-item v-for="action in quick_actions" :key="action.label">
              <n-button
                block
                :type="action.type"
                ghost
                class="quick-action-btn"
                @click="action.handler"
              >
                <template #icon>
                  <n-icon><component :is="action.icon" /></n-icon>
                </template>
                {{ action.label }}
              </n-button>
            </n-grid-item>
          </n-grid>
        </n-card>
      </n-grid-item>

      <!-- 系统信息 -->
      <n-grid-item span="2 m:1">
        <n-card title="系统信息" :bordered="false" class="h-full" hoverable>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item label="系统版本">
              <n-tag size="small" type="info">v1.0.0</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="后端框架">Go 1.24 + Gin</n-descriptions-item>
            <n-descriptions-item label="前端框架">Vue 3 + Naive UI</n-descriptions-item>
            <n-descriptions-item label="运行环境">
              <n-tag size="small" :type="mode === 'production' ? 'success' : 'warning'">{{ mode }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="运行时间">{{ stats.uptime }}</n-descriptions-item>
            <n-descriptions-item label="数据库">MySQL 8.0+</n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-grid-item>

      <!-- 最近活动 -->
      <n-grid-item span="2">
        <n-card title="最近活动" :bordered="false" hoverable>
          <template #header-extra>
            <n-button type="primary" text @click="go_to('logs')">查看全部</n-button>
          </template>
          <n-timeline>
            <n-timeline-item
              v-for="log in recent_logs"
              :key="log.id"
              :type="log.type"
              :title="log.title"
              :content="log.content"
              :time="log.time"
            />
          </n-timeline>
        </n-card>
      </n-grid-item>
    </n-grid>
  </div>
</template>

<script setup lang="ts">
import { ref, markRaw } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard, NGrid, NGridItem, NButton, NIcon, NDescriptions,
  NDescriptionsItem, NTag, NTimeline, NTimelineItem, NNumberAnimation,
  useMessage,
} from 'naive-ui'
import {
  UserOutlined, AppstoreOutlined, ClockCircleOutlined,
  SwapOutlined, SettingOutlined,
  FileTextOutlined, TeamOutlined,
} from '@vicons/antd'
import { getAdminPath } from '@/router/admin.loader'

const router = useRouter()
const message = useMessage()
const admin_path = getAdminPath()
const mode = import.meta.env.MODE

const stats = ref({
  uptime: '12d 4h 23m',
})

// 统计卡片
const stat_cards = ref([
  {
    label: '用户总数',
    value: 1256,
    icon: markRaw(UserOutlined),
    color: '#3b82f6',
    bg: 'rgba(59,130,246,0.1)',
    trend: 3.2,
    suffix: '',
  },
  {
    label: '今日新增',
    value: 42,
    icon: markRaw(TeamOutlined),
    color: '#10b981',
    bg: 'rgba(16,185,129,0.1)',
    trend: 12.5,
    suffix: '',
  },
  {
    label: '系统插件',
    value: 8,
    icon: markRaw(AppstoreOutlined),
    color: '#8b5cf6',
    bg: 'rgba(139,92,246,0.1)',
    trend: 0,
    suffix: ' 个',
  },
  {
    label: 'API 调用',
    value: 1247,
    icon: markRaw(SwapOutlined),
    color: '#f59e0b',
    bg: 'rgba(245,158,11,0.1)',
    trend: -2.1,
    suffix: '',
  },
])

// 快速操作
const quick_actions = ref([
  {
    label: '用户管理',
    icon: markRaw(UserOutlined),
    type: 'primary' as const,
    handler: () => go_to('users'),
  },
  {
    label: '操作日志',
    icon: markRaw(FileTextOutlined),
    type: 'warning' as const,
    handler: () => go_to('logs'),
  },
  {
    label: '系统设置',
    icon: markRaw(SettingOutlined),
    type: 'default' as const,
    handler: () => go_to('settings'),
  },
])

// 最近活动
const recent_logs = ref([
  { id: 1, type: 'success' as const, title: '用户登录', content: '管理员 admin 成功登录系统', time: '2 分钟前' },
  { id: 2, type: 'info' as const, title: '系统更新', content: '系统配置已更新：启用极验验证', time: '30 分钟前' },
  { id: 3, type: 'warning' as const, title: '用户操作', content: '用户 test_user 连续3次登录失败', time: '1 小时前' },
  { id: 4, type: 'success' as const, title: '插件加载', content: 'Demo Plugin v1.0.0 加载成功', time: '3 小时前' },
  { id: 5, type: 'info' as const, title: '数据库迁移', content: '自动执行数据库迁移完成', time: '1 天前' },
])

function go_to(sub_path: string) {
  router.push(`${admin_path}/${sub_path}`)
}
</script>

<style scoped>
.admin-dashboard { padding: 16px; }
.mb-4 { margin-bottom: 16px; }
.h-full { height: 100%; }

.welcome-card {
  background: linear-gradient(135deg, #10b981, #059669) !important;
  border-radius: 12px;
}
.welcome-card :deep(.n-card__content) { color: #fff; }
.welcome-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}
.welcome-title {
  font-size: 22px;
  font-weight: 700;
  margin: 0 0 6px;
  color: #fff;
}
.welcome-desc {
  margin: 0;
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
}
.welcome-actions {
  display: flex;
  gap: 8px;
}

.stat-card { border-radius: 12px; }
.stat-inner {
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.stat-value {
  font-size: 26px;
  font-weight: 700;
  line-height: 1;
}
.stat-suffix {
  font-size: 14px;
  font-weight: 400;
  opacity: 0.6;
}
.stat-label {
  font-size: 13px;
  color: var(--n-text-color-3);
  margin-top: 4px;
}
.stat-trend {
  font-size: 12px;
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px solid var(--n-border-color);
}
.stat-trend.up { color: #10b981; }
.stat-trend.down { color: #ef4444; }

.quick-action-btn {
  height: 44px;
  font-weight: 500;
}
</style>
