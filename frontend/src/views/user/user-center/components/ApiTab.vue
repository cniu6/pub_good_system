<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store'
import { fetchResetApiKey, fetchUserApiKey } from '@/service'
import NovaIcon from '@/components/common/NovaIcon.vue'

const authStore = useAuthStore()
const { t } = useI18n()

const userInfo = computed(() => authStore.userInfo)

const showResetConfirm = ref(false)
const showApiKey = ref(false)
const apiKeyLoading = ref(false)
const resettingApiKey = ref(false)

async function loadApiKey() {
  apiKeyLoading.value = true
  try {
    const response = await fetchUserApiKey()
    if (response.isSuccess) {
      authStore.updateUserInfo({ apikey: response.data?.apikey || null })
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
  if (userInfo.value?.apikey) {
    navigator.clipboard.writeText(userInfo.value.apikey)
    window.$message.success(t('apiTab.apiKeyCopied'))
  }
  else {
    window.$message.warning(t('apiTab.apiKeyEmpty'))
  }
}

async function confirmResetApiKey() {
  if (resettingApiKey.value)
    return
  resettingApiKey.value = true
  try {
    const response = await fetchResetApiKey()
    if (response.isSuccess) {
      window.$message.success(t('apiTab.apiKeyResetSuccess'))
      authStore.updateUserInfo({ apikey: response.data.apikey })
      showResetConfirm.value = false
    }
    else {
      window.$message.error(response.message || t('apiTab.apiKeyResetFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[apiTab] reset api key failed', error)
    window.$message.error(t('apiTab.apiKeyResetFailed'))
  }
  finally {
    resettingApiKey.value = false
  }
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

      <!-- 使用 naive-ui 原生 input-group，保证输入框与重置按钮纵向对齐 -->
      <n-input-group>
        <n-input
          :loading="apiKeyLoading"
          :value="userInfo?.apikey || t('apiTab.noApiKey')"
          :type="showApiKey ? 'text' : 'password'"
          readonly
          :placeholder="t('apiTab.noApiKey')"
        >
          <template #suffix>
            <n-space :size="4" align="center">
              <n-button
                text
                type="primary"
                :disabled="!userInfo?.apikey"
                @click="showApiKey = !showApiKey"
              >
                <template #icon>
                  <NovaIcon v-if="!showApiKey" icon="icon-park-outline:preview-open" :size="16" />
                  <NovaIcon v-else icon="icon-park-outline:preview-close" :size="16" />
                </template>
              </n-button>
              <n-button
                text
                type="primary"
                :disabled="!userInfo?.apikey"
                @click="copyApiKey"
              >
                <template #icon>
                  <NovaIcon icon="icon-park-outline:copy" :size="16" />
                </template>
                {{ t('apiTab.copy') }}
              </n-button>
            </n-space>
          </template>
        </n-input>
        <n-button type="warning" @click="showResetConfirm = true">
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

    <!-- 重置确认对话框 -->
    <n-modal v-model:show="showResetConfirm" preset="card" :title="t('apiTab.confirmResetTitle')" style="width: 400px;" :bordered="false" size="huge">
      <n-alert type="warning" :show-icon="false">
        {{ t('apiTab.confirmResetContent') }}
      </n-alert>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showResetConfirm = false">
            {{ t('common.cancel') }}
          </n-button>
          <n-button type="warning" :loading="resettingApiKey" :disabled="resettingApiKey" @click="confirmResetApiKey">
            {{ t('apiTab.confirmReset') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
