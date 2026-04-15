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
}

onMounted(() => {
  loadApiKey()
})
</script>

<template>
  <div class="p-4">
    <n-h4>{{ t('apiTab.title') }}</n-h4>
    <n-divider />

    <div class="api-section">
      <n-text depth="3" class="api-desc">
        {{ t('apiTab.description') }}
      </n-text>

      <div class="api-key-container">
        <n-input
          :loading="apiKeyLoading"
          :value="userInfo?.apikey || t('apiTab.noApiKey')"
          :type="showApiKey ? 'text' : 'password'"
          readonly
          :placeholder="t('apiTab.noApiKey')"
          class="api-key-input"
        >
          <template #suffix>
            <n-space style="margin-top: 5px;">
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

        <n-button
          type="warning"
          class="reset-btn"
          @click="showResetConfirm = true"
        >
          {{ t('apiTab.resetKey') }}
        </n-button>
      </div>

      <n-alert type="warning" class="mt-4">
        <template #header>
          {{ t('apiTab.notes') }}
        </template>
        <ul class="alert-list">
          <li>{{ t('apiTab.note1') }}</li>
          <li>{{ t('apiTab.note2') }}</li>
        </ul>
      </n-alert>
    </div>

    <!-- 重置确认对话框 -->
    <n-modal v-model:show="showResetConfirm">
      <n-card
        style="width: 400px"
        :title="t('apiTab.confirmResetTitle')"
        :bordered="false"
        size="huge"
        role="dialog"
        aria-modal="true"
      >
        <div class="confirm-content">
          <n-alert type="warning" :show-icon="false">
            {{ t('apiTab.confirmResetContent') }}
          </n-alert>
        </div>
        <template #footer>
          <div class="dialog-footer">
            <n-button @click="showResetConfirm = false">
              {{ t('common.cancel') }}
            </n-button>
            <n-button type="warning" @click="confirmResetApiKey">
              {{ t('apiTab.confirmReset') }}
            </n-button>
          </div>
        </template>
      </n-card>
    </n-modal>
  </div>
</template>

<style scoped>
.api-section {
  max-width: 600px;
}

.api-desc {
  display: block;
  margin: 8px 0 16px 0;
}

.api-key-container {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.api-key-input {
  flex: 1;
}

.reset-btn {
  flex-shrink: 0;
}

.alert-list {
  margin: 0;
  padding-left: 20px;
}

.alert-list li {
  margin-bottom: 4px;
}

.confirm-content {
  margin: 16px 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .api-key-container {
    flex-direction: column;
    gap: 12px;
  }

  .reset-btn {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .api-section {
    max-width: 100%;
  }
}
</style>
