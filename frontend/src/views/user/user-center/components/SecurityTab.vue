<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { withSubmitLock } from '@/hooks'
import { deactivateAccount, fetchUserSessions, fetchUserStats, revokeAllSessions, revokeSession } from '@/service'
import { useAuthStore, useSettingsStore } from '@/store'
import { useBaseCurrency } from '@/composables/useBaseCurrency'
import NovaIcon from '@/components/common/NovaIcon.vue'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'

const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()
const { currencySymbol } = useBaseCurrency()
const allowDeleteAccount = computed(() => settingsStore.allowDeleteAccount)

interface SessionItem {
  id: number | string
  device?: string
  user_agent?: string
  ip?: string
  login_at?: number | null
  last_seen_at?: number | null
  client_type?: string
  is_online?: boolean
  is_current?: boolean
}

interface UserSecurityStats {
  daysJoined?: number
  loginCount?: number
  money?: number
  score?: number
}

const isGeetestEnabled = computed(() => geetestManager.isEnabled())
const deactivateGeetestRef = ref<any>(null)
const deactivateCaptchaKey = ref(0)

const sessions = ref<SessionItem[]>([])
const sessionsLoading = ref(false)
const stats = ref<UserSecurityStats | null>(null)

const showDeactivateModal = ref(false)
const deactivateForm = ref({
  password: '',
  reason: '',
})
const deactivating = ref(false)
/** 踢会话 / 踢全部 防连点 */
const sessionActionLock = ref(false)

function parseBrowser(ua?: string): string {
  if (!ua)
    return ''
  if (ua.includes('Edg/'))
    return 'Edge'
  if (ua.includes('Chrome/') && !ua.includes('Edg/'))
    return 'Chrome'
  if (ua.includes('Firefox/'))
    return 'Firefox'
  if (ua.includes('Safari/') && !ua.includes('Chrome/'))
    return 'Safari'
  if (ua.includes('OPR/') || ua.includes('Opera/'))
    return 'Opera'
  return ''
}

async function loadSessions() {
  sessionsLoading.value = true
  try {
    const response = await fetchUserSessions()
    if (response.isSuccess) {
      const data = response.data
      sessions.value = Array.isArray(data) ? data : []
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[securityTab] load sessions failed', error)
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
    if (import.meta.env.DEV)
      console.error('[securityTab] load stats failed', error)
  }
}

async function handleRevokeSession(sessionId: number | string) {
  await withSubmitLock(sessionActionLock, async () => {
    try {
      const response = await revokeSession(sessionId)
      if (response.isSuccess) {
        window.$message.success(t('securityTab.revokedSession'))
        loadSessions()
      }
      else {
        window.$message.error(response.message || t('adminUsers.operationFailed'))
      }
    }
    catch (error) {
      if (import.meta.env.DEV)
        console.error('[securityTab] revoke session failed', error)
      window.$message.error(t('securityTab.revokeFailed'))
    }
  })
}

async function handleRevokeAll() {
  window.$dialog.warning({
    title: t('securityTab.revokeAllTitle'),
    content: t('securityTab.revokeAllContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(sessionActionLock, async () => {
      try {
        const response = await revokeAllSessions()
        if (response.isSuccess) {
          window.$message.success(t('securityTab.revokedAllSessions'))
          loadSessions()
        }
        else {
          window.$message.error(response.message || t('adminUsers.operationFailed'))
        }
      }
      catch (error) {
        if (import.meta.env.DEV)
          console.error('[securityTab] revoke all sessions failed', error)
        window.$message.error(t('securityTab.revokeAllFailed'))
      }
    }),
  })
}

function triggerDeactivate() {
  if (!deactivateForm.value.password) {
    window.$message.error(t('securityTab.enterPasswordConfirm'))
    return
  }
  if (isGeetestEnabled.value) {
    deactivateGeetestRef.value?.showCaptcha()
  }
  else {
    doDeactivate()
  }
}

function onDeactivateGeetestSuccess() {
  doDeactivate()
}

function onDeactivateGeetestError() {
  geetestManager.clearCaptchaResult()
  deactivateCaptchaKey.value++
}

async function doDeactivate() {
  await withSubmitLock(deactivating, async () => {
    try {
      const response = await deactivateAccount({
        password: deactivateForm.value.password,
        reason: deactivateForm.value.reason,
      })
      if (response.isSuccess) {
        window.$message.success(t('securityTab.accountDeactivated'))
        showDeactivateModal.value = false
        setTimeout(() => {
          authStore.logout()
        }, 1500)
      }
      else {
        window.$message.error(response.message || t('securityTab.deactivateFailed'))
      }
    }
    catch (error) {
      if (import.meta.env.DEV)
        console.error('[securityTab] deactivate failed', error)
      window.$message.error(t('securityTab.deactivateFailed'))
    }
  })
}

function formatTime(timestamp?: number | null) {
  if (!timestamp)
    return t('profile.na')
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
        <n-h4>{{ t('securityTab.loginStats') }}</n-h4>
        <n-divider />
        <n-grid cols="2 m:4" :x-gap="16" :y-gap="16" responsive="screen">
          <n-grid-item>
            <n-statistic :label="t('securityTab.daysJoined')" :value="stats?.daysJoined || 0">
              <template #suffix>
                {{ t('securityTab.dayUnit') }}
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic :label="t('securityTab.loginCount')" :value="stats?.loginCount || 0">
              <template #suffix>
                {{ t('securityTab.timesUnit') }}
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic :label="t('securityTab.accountBalance')">
              <template #default>
                {{ currencySymbol }}{{ stats?.money ? Number(stats.money).toFixed(2) : '0.00' }}
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic :label="t('adminUsers.score')" :value="stats?.score || 0" />
          </n-grid-item>
        </n-grid>
      </div>

      <n-divider />

      <!-- 登录设备管理 -->
      <div>
        <div class="section-header">
          <n-h4>{{ t('securityTab.sessionManagement') }}</n-h4>
          <n-space>
            <n-button size="small" @click="loadSessions">
              <template #icon>
                <NovaIcon icon="icon-park-outline:refresh" :size="14" />
              </template>
              {{ t('common.reload') }}
            </n-button>
            <n-button size="small" type="warning" @click="handleRevokeAll">
              {{ t('securityTab.revokeAllDevices') }}
            </n-button>
          </n-space>
        </div>
        <n-divider />
        <n-spin :show="sessionsLoading">
          <n-space v-if="sessions.length > 0" vertical>
            <div v-for="session in sessions" :key="session.id" class="session-item" :class="{ 'is-online': session.is_online }">
              <div class="session-info">
                <div class="session-device">
                  <span class="online-dot" :class="{ offline: !session.is_online }" />
                  <NovaIcon icon="icon-park-outline:computer" :size="16" class="mr-1" />
                  {{ session.device || t('securityTab.unknownDevice') }}
                  <n-tag v-if="session.is_current" size="small" type="primary" :bordered="false" class="ml-2">
                    {{ t('securityTab.currentDevice') }}
                  </n-tag>
                  <n-tag v-if="session.client_type" size="small" :bordered="false" class="ml-2">
                    {{ session.client_type }}
                  </n-tag>
                  <n-tag v-if="parseBrowser(session.user_agent)" size="small" :bordered="false" class="ml-2">
                    {{ parseBrowser(session.user_agent) }}
                  </n-tag>
                </div>
                <n-text depth="3" class="session-detail">
                  {{ t('securityTab.sessionDetail', { ip: session.ip || t('securityTab.unknown'), time: formatTime(session.login_at) }) }}
                  <span v-if="session.last_seen_at"> · {{ t('securityTab.lastSeenAt') }}: {{ formatTime(session.last_seen_at) }}</span>
                </n-text>
              </div>
              <n-button size="small" type="error" @click="handleRevokeSession(session.id)">
                {{ t('securityTab.revoke') }}
              </n-button>
            </div>
          </n-space>
          <n-empty v-else :description="t('securityTab.noSessionRecords')" />
        </n-spin>
      </div>

      <n-divider />

      <!-- 账号注销 -->
      <div>
        <n-h4>{{ t('securityTab.dangerZone') }}</n-h4>
        <n-divider />
        <div class="danger-zone">
          <div class="danger-info">
            <span class="danger-label">{{ t('securityTab.deactivateAccount') }}</span>
            <span v-if="allowDeleteAccount" class="danger-desc">{{ t('securityTab.deactivateDesc') }}</span>
            <span v-else class="danger-desc">{{ t('securityTab.deactivateDisabledDesc') }}</span>
          </div>
          <n-tooltip :disabled="allowDeleteAccount">
            <template #trigger>
              <n-button type="error" :disabled="!allowDeleteAccount" @click="showDeactivateModal = true">
                {{ t('securityTab.deactivateAccount') }}
              </n-button>
            </template>
            {{ t('securityTab.deactivateDisabledTooltip') }}
          </n-tooltip>
        </div>
      </div>
    </n-space>

    <!-- 注销确认弹窗 -->
    <n-modal v-model:show="showDeactivateModal" preset="dialog" :title="t('securityTab.deactivateAccount')" type="error">
      <n-alert type="error" class="mb-4">
        {{ t('securityTab.deactivateWarning') }}
      </n-alert>
      <n-form :model="deactivateForm" label-placement="left" label-width="100px">
        <n-form-item :label="t('securityTab.confirmPassword')" required>
          <n-input
            v-model:value="deactivateForm.password"
            type="password"
            :placeholder="t('securityTab.confirmPasswordPlaceholder')"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item :label="t('securityTab.deactivateReason')">
          <n-input
            v-model:value="deactivateForm.reason"
            type="textarea"
            :placeholder="t('securityTab.deactivateReasonPlaceholder')"
            :rows="3"
          />
        </n-form-item>
        <GeetestCaptcha
          v-if="isGeetestEnabled"
          ref="deactivateGeetestRef"
          :key="deactivateCaptchaKey"
          :config="{ product: 'bind' }"
          @success="onDeactivateGeetestSuccess"
          @error="onDeactivateGeetestError"
        />
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showDeactivateModal = false">
            {{ t('common.cancel') }}
          </n-button>
          <n-button type="error" :loading="deactivating" @click="triggerDeactivate">
            {{ t('securityTab.confirmDeactivate') }}
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

.session-item.is-online {
  border-color: var(--n-success-color);
  background: color-mix(in srgb, var(--n-success-color) 8%, var(--n-color));
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

.online-dot {
  width: 8px;
  height: 8px;
  margin-right: 8px;
  border-radius: 50%;
  background: var(--n-success-color);
  box-shadow: 0 0 8px var(--n-success-color);
}

.online-dot.offline {
  background: var(--n-text-color-disabled);
  box-shadow: none;
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
