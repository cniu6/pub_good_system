<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NTag, useDialog, useMessage } from 'naive-ui'
import { adminOnlineApi, type OnlineSession, type OnlineSessionListParams } from '@/service/api/admin/online'
import { fetchSetting, updateSetting } from '@/service/api/admin/settings'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const stats = ref({ online_users: 0, online_sessions: 0 })
const sessions = ref<OnlineSession[]>([])
// client_type 默认值为空字符串，语义上代表「全部」，配合下方 clientTypeOptions 的「全部」选项显式展示，而非空白占位符
const query = reactive<OnlineSessionListParams>({ page: 1, page_size: 20, keyword: '', client_type: '' })
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

// ── 在线心跳「上报周期」设置（默认30秒，直接在本页配置，无需跳转系统设置） ──
const REPORT_INTERVAL_KEY = 'online_report_interval_seconds'
const reportInterval = ref(30)
const reportIntervalSaving = ref(false)
const reportIntervalLoading = ref(false)

async function loadReportInterval() {
  reportIntervalLoading.value = true
  try {
    const res = await fetchSetting(REPORT_INTERVAL_KEY)
    if (res.isSuccess && res.data) {
      const val = Number(res.data.value)
      if (Number.isFinite(val) && val > 0)
        reportInterval.value = val
    }
  }
  catch {
    // 静默失败，保留默认值 30
  }
  finally {
    reportIntervalLoading.value = false
  }
}

async function saveReportInterval() {
  const val = Math.round(reportInterval.value)
  if (!Number.isFinite(val) || val < 10 || val > 300) {
    message.error(t('adminOnlineUsers.reportIntervalSaveFailed'))
    return
  }
  reportIntervalSaving.value = true
  try {
    const res = await updateSetting(REPORT_INTERVAL_KEY, String(val))
    if (res.isSuccess) {
      reportInterval.value = val
      message.success(t('adminOnlineUsers.reportIntervalSaveSuccess'))
    }
    else {
      message.error(res.message || t('adminOnlineUsers.reportIntervalSaveFailed'))
    }
  }
  catch {
    message.error(t('adminOnlineUsers.reportIntervalSaveFailed'))
  }
  finally {
    reportIntervalSaving.value = false
  }
}

function formatTime(timestamp?: number) {
  return timestamp ? new Date(timestamp * 1000).toLocaleString() : '-'
}

function formatUserLabel(row: OnlineSession) {
  const name = row.username || row.nickname
  return name ? `${name} (#${row.user_id})` : `#${row.user_id}`
}

function formatDeviceDetail(row: OnlineSession) {
  const parts = [row.device || '-']
  if (row.user_agent) {
    const ua = row.user_agent.length > 80 ? `${row.user_agent.slice(0, 80)}…` : row.user_agent
    parts.push(ua)
  }
  return parts.join(' · ')
}

function formatGuard(guard?: string) {
  if (guard === 'admin')
    return t('adminOnlineUsers.guardAdmin')
  return t('adminOnlineUsers.guardUser')
}

const columns: DataTableColumns<OnlineSession> = [
  {
    title: t('adminOnlineUsers.user'),
    key: 'username',
    minWidth: 180,
    ellipsis: { tooltip: true },
    render: row => formatUserLabel(row),
  },
  {
    title: t('adminOnlineUsers.authGuard'),
    key: 'auth_guard',
    width: 100,
    render: row => formatGuard(row.auth_guard),
  },
  { title: t('adminOnlineUsers.clientType'), key: 'client_type', width: 90 },
  {
    title: t('adminOnlineUsers.device'),
    key: 'device',
    minWidth: 220,
    ellipsis: { tooltip: true },
    render: row => formatDeviceDetail(row),
  },
  { title: 'IP', key: 'ip', width: 130 },
  {
    title: t('adminOnlineUsers.loginAt'),
    key: 'login_at',
    width: 170,
    render: row => formatTime(row.login_at),
  },
  { title: t('adminOnlineUsers.lastSeenAt'), key: 'last_seen_at', width: 170, render: row => formatTime(row.last_seen_at) },
  {
    title: t('adminOnlineUsers.status'),
    key: 'is_online',
    width: 90,
    render: row => h(NTag, { type: row.is_online ? 'success' : 'default', size: 'small' }, () => row.is_online ? t('adminOnlineUsers.online') : t('adminOnlineUsers.offline')),
  },
  {
    title: t('adminOnlineUsers.actions'),
    key: 'actions',
    width: 100,
    render: row => h(NButton, { size: 'small', type: 'error', secondary: true, onClick: () => handleKick(row) }, () => t('adminOnlineUsers.kick')),
  },
]

async function fetchStats() {
  try {
    const res = await adminOnlineApi.stats()
    if (res.isSuccess && res.data)
      stats.value = res.data
  }
  catch {
    // 统计请求失败不影响会话列表使用。
  }
}

async function fetchSessions() {
  loading.value = true
  try {
    const params = { ...query }
    if (!params.keyword) delete params.keyword
    if (!params.client_type) delete params.client_type
    const res = await adminOnlineApi.sessions(params)
    if (res.isSuccess && res.data) {
      sessions.value = res.data.list || []
      pagination.itemCount = res.data.total || 0
    }
    else {
      message.error(res.message || t('adminOnlineUsers.loadFailed'))
    }
  }
  catch {
    message.error(t('adminOnlineUsers.loadFailed'))
  }
  finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  pagination.page = 1
  fetchSessions()
}

function handleReset() {
  query.keyword = ''
  query.client_type = ''
  handleSearch()
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchSessions()
}

async function refreshAll() {
  await Promise.all([fetchSessions(), fetchStats()])
}

function handleKick(session: OnlineSession) {
  dialog.warning({
    title: t('adminOnlineUsers.kickTitle'),
    content: t('adminOnlineUsers.kickContent', { device: session.device || `#${session.id}` }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await adminOnlineApi.kick(session.id)
        if (!res.isSuccess) {
          message.error(res.message || t('adminOnlineUsers.kickFailed'))
          return
        }
        message.success(t('adminOnlineUsers.kickSuccess'))
        await Promise.all([fetchSessions(), fetchStats()])
      }
      catch {
        message.error(t('adminOnlineUsers.kickFailed'))
      }
    },
  })
}

onMounted(() => {
  void Promise.all([fetchSessions(), fetchStats(), loadReportInterval()])
})
</script>

<template>
  <div class="online-users-page">
    <n-grid cols="2" :x-gap="12" :y-gap="12" style="margin-bottom: 16px;">
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminOnlineUsers.onlineUsers')" :value="stats.online_users" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small">
          <n-statistic :label="t('adminOnlineUsers.onlineDevices')" :value="stats.online_sessions" />
        </n-card>
      </n-gi>
    </n-grid>

    <n-card :title="t('adminOnlineUsers.title')">
      <template #header-extra>
        <n-button size="small" type="primary" :loading="loading" @click="refreshAll">
          {{ t('common.refresh') }}
        </n-button>
      </template>
      <n-space align="center" :wrap="true" style="margin-bottom: 12px;">
        <n-input v-model:value="query.keyword" clearable size="small" :placeholder="t('adminOnlineUsers.keyword')" style="width: 180px;" @keyup.enter="handleSearch" />
        <!-- 首个选项即「全部」，默认选中它而非留空占位符，避免看起来像没设置默认值 -->
        <n-select
          v-model:value="query.client_type"
          size="small"
          style="width: 120px;"
          :options="[
            { label: t('adminOnlineUsers.allClientTypes'), value: '' },
            { label: t('adminOnlineUsers.web'), value: 'web' },
            { label: t('adminOnlineUsers.app'), value: 'app' },
          ]"
          @update:value="handleSearch"
        />
        <n-button size="small" type="primary" @click="handleSearch">{{ t('adminOnlineUsers.search') }}</n-button>
        <n-button size="small" @click="handleReset">{{ t('adminOnlineUsers.reset') }}</n-button>
      </n-space>
      <n-alert type="default" :show-icon="true" style="margin-bottom: 12px;">
        {{ t('adminOnlineUsers.multiDeviceHint') }}
      </n-alert>
      <n-data-table
        remote
        :columns="columns"
        :data="sessions"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="1200"
        :row-key="row => row.id"
        @update:page="handlePageChange"
      />
    </n-card>

    <!-- 在线心跳「上报周期」设置：直接在本页配置，默认30秒，无需跳转系统设置 -->
    <n-card :title="t('adminOnlineUsers.reportIntervalTitle')" style="margin-top: 16px;" size="small">
      <n-space align="center" :wrap="true">
        <span>{{ t('adminOnlineUsers.reportIntervalLabel') }}</span>
        <n-input-number
          v-model:value="reportInterval"
          :min="10"
          :max="300"
          :step="5"
          size="small"
          style="width: 140px;"
          :disabled="reportIntervalLoading"
        >
          <template #suffix>
            {{ t('adminOnlineUsers.reportIntervalUnit') }}
          </template>
        </n-input-number>
        <n-button size="small" type="primary" :loading="reportIntervalSaving" @click="saveReportInterval">
          {{ t('adminOnlineUsers.reportIntervalSave') }}
        </n-button>
      </n-space>
      <n-text depth="3" style="display: block; margin-top: 8px; font-size: 12px;">
        {{ t('adminOnlineUsers.reportIntervalHint') }}
      </n-text>
    </n-card>
  </div>
</template>

<style scoped>
.online-users-page {
  padding: 16px;
}
</style>
