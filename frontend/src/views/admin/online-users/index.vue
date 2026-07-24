<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NSpace, NTag, NText, NTooltip, useDialog, useMessage } from 'naive-ui'
import { adminOnlineApi } from '@/service/api/admin/online'
import type { OnlineSession, OnlineSessionListParams, OnlineUserRow } from '@/service/api/admin/online'
import { fetchSetting, updateSetting } from '@/service/api/admin/settings'
import { useAuthStore, useSettingsStore } from '@/store'
import { useRequestGuard, withSubmitLock } from '@/hooks'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const presenceEnabled = computed(() => settingsStore.presenceEnabled)
const sessionsFetchGuard = useRequestGuard()
const loading = ref(false)
/** 踢下线写操作防连点 */
const kicking = ref(false)
const stats = ref({ online_users: 0, online_sessions: 0 })
const userRows = ref<OnlineUserRow[]>([])
// client_type 默认值为空字符串，语义上代表「全部」
const query = reactive<OnlineSessionListParams>({ page: 1, page_size: 20, keyword: '', client_type: '' })
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

// ── 在线心跳「上报周期」设置（默认30秒，直接在本页配置） ──
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
      // 热更新：同步 settingsStore，并通过认证层统一重启当前 Presence。
      // 不在页面里直接调组合式函数，避免未来认证/多标签协调逻辑变更后出现两套重启路径。
      settingsStore.updateConfig({ online_report_interval_seconds: val })
      authStore.startPresence()
      message.success(t('adminOnlineUsers.reportIntervalSaveSuccessHot'))
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

function formatUserLabel(row: OnlineUserRow) {
  const name = row.username || row.nickname
  return name ? `${name} (#${row.user_id})` : `#${row.user_id}`
}

function formatGuard(guard?: string) {
  if (guard === 'admin')
    return t('adminOnlineUsers.guardAdmin')
  return t('adminOnlineUsers.guardUser')
}

function deviceShortLabel(device?: OnlineSession) {
  if (!device)
    return '-'
  const parts = [device.client_type || '-', device.device || '-']
  if (device.ip)
    parts.push(device.ip)
  return parts.join(' · ')
}

function deviceTooltip(device?: OnlineSession) {
  if (!device)
    return ''
  const ua = device.user_agent
    ? (device.user_agent.length > 120 ? `${device.user_agent.slice(0, 120)}…` : device.user_agent)
    : ''
  return [
    `${t('adminOnlineUsers.device')}: ${device.device || '-'}`,
    `IP: ${device.ip || '-'}`,
    `${t('adminOnlineUsers.loginAt')}: ${formatTime(device.login_at)}`,
    `${t('adminOnlineUsers.lastSeenAt')}: ${formatTime(device.last_seen_at)}`,
    ua ? `UA: ${ua}` : '',
  ].filter(Boolean).join('\n')
}

const columns: DataTableColumns<OnlineUserRow> = [
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
  {
    title: t('adminOnlineUsers.devices'),
    key: 'devices',
    minWidth: 360,
    render: (row) => {
      const devices = row.devices || []
      if (!devices.length)
        return h(NText, { depth: 3 }, () => '-')
      return h(
        NSpace,
        { size: 8, vertical: true, style: 'padding: 4px 0;' },
        () => devices.map(device =>
          h(
            NSpace,
            { key: device.id, align: 'center', size: 8, wrap: false },
            () => [
              h(
                NTooltip,
                { trigger: 'hover' },
                {
                  trigger: () => h(
                    NTag,
                    {
                      size: 'small',
                      type: device.is_online ? 'success' : 'default',
                      round: true,
                      style: 'max-width: 280px;',
                    },
                    () => deviceShortLabel(device),
                  ),
                  default: () => h('div', { style: 'white-space: pre-line; max-width: 360px;' }, deviceTooltip(device)),
                },
              ),
              h(
                NButton,
                {
                  size: 'tiny',
                  type: 'error',
                  quaternary: true,
                  onClick: () => handleKick(device, row),
                },
                () => t('adminOnlineUsers.kick'),
              ),
            ],
          ),
        ),
      )
    },
  },
  {
    title: t('adminOnlineUsers.deviceCount'),
    key: 'device_count',
    width: 90,
    render: row => row.device_count || (row.devices?.length ?? 0),
  },
  {
    title: t('adminOnlineUsers.lastSeenAt'),
    key: 'last_seen_at',
    width: 170,
    render: row => formatTime(row.last_seen_at),
  },
  {
    title: t('adminOnlineUsers.status'),
    key: 'is_online',
    width: 90,
    render: row => h(NTag, { type: row.is_online ? 'success' : 'default', size: 'small' }, () => row.is_online ? t('adminOnlineUsers.online') : t('adminOnlineUsers.offline')),
  },
]

async function fetchStats() {
  if (!presenceEnabled.value)
    return
  try {
    const res = await adminOnlineApi.stats()
    if (res.isSuccess && res.data)
      stats.value = res.data
  }
  catch {
    // 统计请求失败不影响会话列表使用
  }
}

async function fetchSessions() {
  if (!presenceEnabled.value)
    return
  const token = sessionsFetchGuard.begin()
  loading.value = true
  try {
    const params = { ...query }
    if (!params.keyword)
      delete params.keyword
    if (!params.client_type)
      delete params.client_type
    const res = await adminOnlineApi.sessions(params)
    if (!sessionsFetchGuard.isLatest(token))
      return
    if (res.isSuccess && res.data) {
      userRows.value = res.data.list || []
      pagination.itemCount = res.data.total || 0
    }
    else {
      message.error(res.message || t('adminOnlineUsers.loadFailed'))
    }
  }
  catch {
    if (sessionsFetchGuard.isLatest(token))
      message.error(t('adminOnlineUsers.loadFailed'))
  }
  finally {
    if (sessionsFetchGuard.isLatest(token))
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

function handleKick(session: OnlineSession, userRow?: OnlineUserRow) {
  if (kicking.value)
    return
  const label = session.device || `#${session.id}`
  const userHint = userRow ? formatUserLabel(userRow) : ''
  dialog.warning({
    title: t('adminOnlineUsers.kickTitle'),
    content: userHint
      ? t('adminOnlineUsers.kickContentWithUser', { user: userHint, device: label })
      : t('adminOnlineUsers.kickContent', { device: label }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(kicking, async () => {
      try {
        const res = await adminOnlineApi.kick(session.id)
        if (!res.isSuccess) {
          message.error(res.message || t('adminOnlineUsers.kickFailed'))
          return false
        }
        message.success(t('adminOnlineUsers.kickSuccess'))
        await Promise.all([fetchSessions(), fetchStats()])
      }
      catch {
        message.error(t('adminOnlineUsers.kickFailed'))
        return false
      }
    }),
  })
}

onMounted(() => {
  if (!presenceEnabled.value)
    return
  void Promise.all([fetchSessions(), fetchStats(), loadReportInterval()])
})
</script>

<template>
  <div class="online-users-page">
    <n-alert
      v-if="!presenceEnabled"
      type="warning"
      :title="t('adminOnlineUsers.featureDisabledTitle')"
      style="margin-bottom: 16px;"
    >
      {{ t('adminOnlineUsers.featureDisabledDesc') }}
    </n-alert>
    <div v-else>
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
          <NButton size="small" type="primary" :loading="loading" @click="refreshAll">
            {{ t('common.refresh') }}
          </NButton>
        </template>
        <NSpace align="center" :wrap="true" style="margin-bottom: 12px;">
          <n-input v-model:value="query.keyword" clearable size="small" :placeholder="t('adminOnlineUsers.keyword')" style="width: 180px;" @keyup.enter="handleSearch" />
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
          <NButton size="small" type="primary" @click="handleSearch">
            {{ t('adminOnlineUsers.search') }}
          </NButton>
          <NButton size="small" @click="handleReset">
            {{ t('adminOnlineUsers.reset') }}
          </NButton>
        </NSpace>
        <n-alert type="default" :show-icon="true" style="margin-bottom: 12px;">
          {{ t('adminOnlineUsers.multiDeviceHint') }}
        </n-alert>
        <n-data-table
          remote
          :columns="columns"
          :data="userRows"
          :loading="loading"
          :pagination="pagination"
          :scroll-x="1100"
          :row-key="row => `${row.user_id}-${row.auth_guard}`"
          @update:page="handlePageChange"
        />
      </n-card>

      <n-card :title="t('adminOnlineUsers.reportIntervalTitle')" style="margin-top: 16px;" size="small">
        <NSpace align="center" :wrap="true">
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
          <NButton size="small" type="primary" :loading="reportIntervalSaving" @click="saveReportInterval">
            {{ t('adminOnlineUsers.reportIntervalSave') }}
          </NButton>
        </NSpace>
        <NText depth="3" style="display: block; margin-top: 8px; font-size: 12px;">
          {{ t('adminOnlineUsers.reportIntervalHint') }}
        </NText>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.online-users-page {
  padding: 16px;
}
</style>
