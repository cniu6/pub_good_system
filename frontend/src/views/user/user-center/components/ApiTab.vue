<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store'
import { fetchResetApiKey, fetchUserApiKey } from '@/service'
import { withSubmitLock } from '@/hooks'

const authStore = useAuthStore()
const { t } = useI18n()
const dialog = useDialog()
const message = useMessage()

const apiKeyLoading = ref(false)
const resettingApiKey = ref(false)
/** 用户端完整密钥（明文）；前端用 password 眼睛控制显隐 */
const apiKey = ref('')

async function loadApiKey() {
  apiKeyLoading.value = true
  try {
    const response = await fetchUserApiKey()
    if (response.isSuccess) {
      const key = String(response.data?.apikey || '').trim()
      apiKey.value = key
      authStore.updateUserInfo({ apikey: key || null })
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[apiTab] fetch api key failed', error)
  }
  finally {
    apiKeyLoading.value = false
  }
}

function copyApiKey() {
  const key = apiKey.value.trim()
  if (!key) {
    message.warning(t('apiTab.apiKeyEmpty'))
    return
  }
  navigator.clipboard.writeText(key)
  message.success(t('apiTab.apiKeyCopied'))
}

function handleResetApiKey() {
  if (resettingApiKey.value)
    return
  dialog.warning({
    title: t('apiTab.confirmResetTitle'),
    content: t('apiTab.confirmResetContent'),
    positiveText: t('apiTab.confirmReset'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => withSubmitLock(resettingApiKey, async () => {
      try {
        const response = await fetchResetApiKey()
        if (!response.isSuccess) {
          message.error(response.message || t('apiTab.apiKeyResetFailed'))
          return false
        }
        const plain = String(response.data?.apikey || '').trim()
        if (!plain) {
          message.error(t('apiTab.apiKeyResetFailed'))
          return false
        }
        apiKey.value = plain
        authStore.updateUserInfo({ apikey: plain })
        message.success(t('apiTab.apiKeyResetSuccess'))
      }
      catch (error) {
        if (import.meta.env.DEV)
          console.error('[apiTab] reset api key failed', error)
        message.error(t('apiTab.apiKeyResetFailed'))
        return false
      }
    }),
  })
}

onMounted(() => {
  loadApiKey()
})
</script>

<template>
  <div class="p-4">
    <n-h4>{{ t('apiTab.title') }}</n-h4>
    <n-divider />

    <n-space vertical :size="16" style="max-width: 640px;">
      <n-text depth="3">
        {{ t('apiTab.description') }}
      </n-text>

      <n-input-group>
        <n-input
          :loading="apiKeyLoading"
          :value="apiKey"
          type="password"
          show-password-on="click"
          readonly
          :placeholder="t('apiTab.noApiKey')"
        />
        <n-button type="primary" :disabled="!apiKey" @click="copyApiKey">
          {{ t('apiTab.copy') }}
        </n-button>
        <n-button type="warning" :loading="resettingApiKey" :disabled="resettingApiKey" @click="handleResetApiKey">
          {{ t('apiTab.resetKey') }}
        </n-button>
      </n-input-group>

      <n-alert type="warning">
        <template #header>
          {{ t('apiTab.notes') }}
        </template>
        <n-space vertical :size="4">
          <n-text>{{ t('apiTab.note1') }}</n-text>
          <n-text>{{ t('apiTab.note2') }}</n-text>
          <n-text>{{ t('apiTab.note3') }}</n-text>
          <n-text>{{ t('apiTab.note4') }}</n-text>
          <n-text>{{ t('apiTab.note5') }}</n-text>
        </n-space>
      </n-alert>
    </n-space>
  </div>
</template>
