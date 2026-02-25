<script setup lang="ts">
import { ref, markRaw } from 'vue'
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
  NTimeline,
  NTimelineItem,
  NThing,
  NAlert,
  NFlex,
  NText,
  NEl,
} from 'naive-ui'
import {
  UserOutlined,
  AppstoreOutlined,
  SwapOutlined,
  SettingOutlined,
  FileTextOutlined,
  TeamOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined,
  WarningOutlined,
} from '@vicons/antd'
import { getAdminPath } from '@/router/admin.loader'

const router = useRouter()
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
    color: 'var(--info-color)',
    trend: 3.2,
    trend_type: 'success' as const,
  },
  {
    label: '今日新增',
    value: 42,
    icon: markRaw(TeamOutlined),
    color: 'var(--success-color)',
    trend: 12.5,
    trend_type: 'success' as const,
  },
  {
    label: '系统插件',
    value: 8,
    icon: markRaw(AppstoreOutlined),
    color: 'var(--warning-color)',
    trend: 0,
    trend_type: 'default' as const,
  },
  {
    label: 'API 调用',
    value: 1247,
    icon: markRaw(SwapOutlined),
    color: 'var(--error-color)',
    trend: -2.1,
    trend_type: 'error' as const,
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
  { id: 1, type: 'success' as const, title: '用户登录', content: '管理员 admin 成功登录系统', time: '2 分钟前', icon: markRaw(CheckCircleOutlined) },
  { id: 2, type: 'info' as const, title: '系统更新', content: '系统配置已更新：启用极验验证', time: '30 分钟前', icon: markRaw(InfoCircleOutlined) },
  { id: 3, type: 'warning' as const, title: '用户操作', content: '用户 test_user 连续3次登录失败', time: '1 小时前', icon: markRaw(WarningOutlined) },
  { id: 4, type: 'success' as const, title: '插件加载', content: 'Demo Plugin v1.0.0 加载成功', time: '3 小时前', icon: markRaw(CheckCircleOutlined) },
  { id: 5, type: 'info' as const, title: '数据库迁移', content: '自动执行数据库迁移完成', time: '1 天前', icon: markRaw(InfoCircleOutlined) },
])

function go_to(sub_path: string) {
  router.push(`${admin_path}/${sub_path}`)
}
</script>

<template>
  <n-space
    vertical
    :size="16"
  >
    <!-- 欢迎横幅 -->
    <n-card hoverable>
      <n-flex
        justify="space-between"
        align="center"
        wrap="wrap"
        :size="16"
      >
        <n-flex
          align="center"
          :size="16"
        >
          <n-icon-wrapper
            :size="48"
            :border-radius="12"
            color="var(--success-color)"
          >
            <n-icon
              :size="26"
              color="#fff"
            >
              <CheckCircleOutlined />
            </n-icon>
          </n-icon-wrapper>
          <n-flex vertical>
            <n-text strong>
              欢迎回来，管理员
            </n-text>
            <n-text depth="3">
              这里是系统运行概览，一切正常运行中。
            </n-text>
          </n-flex>
        </n-flex>
        <n-flex :size="8">
          <n-button
            type="primary"
            @click="go_to('users')"
          >
            用户管理
          </n-button>
          <n-button @click="go_to('settings')">
            系统设置
          </n-button>
        </n-flex>
      </n-flex>
    </n-card>

    <!-- 统计卡片 -->
    <n-grid
      :x-gap="16"
      :y-gap="16"
      :cols="4"
      item-responsive
      responsive="screen"
    >
      <n-gi
        v-for="card in stat_cards"
        :key="card.label"
        span="4 s:2 m:1"
      >
        <n-card hoverable>
          <n-thing>
            <template #avatar>
              <n-el>
                <n-icon-wrapper
                  :size="46"
                  :color="card.color"
                  :border-radius="12"
                >
                  <n-icon
                    :size="24"
                    color="#fff"
                  >
                    <component :is="card.icon" />
                  </n-icon>
                </n-icon-wrapper>
              </n-el>
            </template>
            <template #header>
              <n-statistic :label="card.label">
                <n-number-animation
                  :from="0"
                  :to="card.value"
                  show-separator
                />
              </n-statistic>
            </template>
            <template #description>
              <n-flex
                align="center"
                :size="4"
              >
                <n-tag
                  :type="card.trend_type"
                  size="small"
                  :bordered="false"
                >
                  {{ card.trend > 0 ? '+' : '' }}{{ card.trend }}%
                </n-tag>
                <n-text depth="3">
                  较昨日
                </n-text>
              </n-flex>
            </template>
          </n-thing>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 下半区域 -->
    <n-grid
      :x-gap="16"
      :y-gap="16"
      :cols="12"
      item-responsive
      responsive="screen"
    >
      <!-- 快速操作 -->
      <n-gi span="12 m:6">
        <n-card
          title="快速操作"
          hoverable
        >
          <n-grid
            :x-gap="12"
            :y-gap="12"
            :cols="3"
            item-responsive
          >
            <n-gi
              v-for="action in quick_actions"
              :key="action.label"
              span="3 s:1"
            >
              <n-button
                block
                :type="action.type"
                ghost
                size="large"
                @click="action.handler"
              >
                <template #icon>
                  <n-icon>
                    <component :is="action.icon" />
                  </n-icon>
                </template>
                {{ action.label }}
              </n-button>
            </n-gi>
          </n-grid>
        </n-card>
      </n-gi>

      <!-- 系统信息 -->
      <n-gi span="12 m:6">
        <n-card
          title="系统信息"
          hoverable
        >
          <n-descriptions
            :column="1"
            label-placement="left"
            bordered
            size="small"
          >
            <n-descriptions-item label="系统版本">
              <n-tag
                size="small"
                type="info"
              >
                v1.0.0
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="后端框架">
              Go 1.24 + Gin
            </n-descriptions-item>
            <n-descriptions-item label="前端框架">
              Vue 3 + Naive UI
            </n-descriptions-item>
            <n-descriptions-item label="运行环境">
              <n-tag
                size="small"
                :type="mode === 'production' ? 'success' : 'warning'"
              >
                {{ mode }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="运行时间">
              {{ stats.uptime }}
            </n-descriptions-item>
            <n-descriptions-item label="数据库">
              MySQL 8.0+
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <!-- 最近活动 -->
      <n-gi :span="12">
        <n-card
          title="最近活动"
          hoverable
        >
          <template #header-extra>
            <n-button
              type="primary"
              quaternary
              @click="go_to('logs')"
            >
              查看全部
            </n-button>
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
      </n-gi>
    </n-grid>
  </n-space>
</template>

<style scoped></style>
