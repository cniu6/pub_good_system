<script setup lang="ts">
import { useAppStore } from '@/store'
import { useI18n } from 'vue-i18n'
import { fetchUserSettings, updateUserSettings } from '@/service'
import { langToBackendFormat, langToFrontendFormat } from '@/utils'

const appStore = useAppStore()
const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)

const settingsForm = ref({
  language: appStore.lang as App.lang,
  theme: 'light',
  notify_email: true,
})

const languageOptions = [
  { label: t('settingsTab.simplifiedChinese'), value: 'zhCN' as App.lang },
  { label: t('settingsTab.english'), value: 'enUS' as App.lang },
]

const themeOptions = [
  { label: t('settingsTab.lightMode'), value: 'light' },
  { label: t('settingsTab.darkMode'), value: 'dark' },
  { label: t('settingsTab.followSystem'), value: 'auto' },
]

async function loadSettings() {
  loading.value = true
  try {
    const response = await fetchUserSettings()
    if (response.isSuccess && response.data) {
      settingsForm.value = {
        language: langToFrontendFormat(response.data.language || 'zh-CN'),
        theme: response.data.theme || 'light',
        notify_email: response.data.notify_email ?? true,
      }
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[settingsTab] fetch failed', error)
  }
  finally {
    loading.value = false
  }
}

async function handleSaveSettings() {
  saving.value = true
  try {
    const response = await updateUserSettings({
      language: langToBackendFormat(settingsForm.value.language),
      theme: settingsForm.value.theme,
      notify_email: settingsForm.value.notify_email,
    })
    if (response.isSuccess) {
      appStore.setAppLang(settingsForm.value.language)
      window.$message.success(t('settingsTab.saveSuccess'))
    }
    else {
      window.$message.error(response.message || t('settingsTab.saveFailed'))
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[settingsTab] save failed', error)
    window.$message.error(t('settingsTab.saveFailed'))
  }
  finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <div class="p-4">
    <n-spin :show="loading">
      <n-space vertical size="large">
        <!-- 显示偏好 -->
        <div>
          <n-h4>{{ t('settingsTab.displayPreference') }}</n-h4>
          <n-divider />
          <n-grid cols="1 s:2" :x-gap="32" :y-gap="0" responsive="screen">
            <n-grid-item>
              <n-form-item :label="t('settingsTab.interfaceLanguage')" label-placement="top">
                <n-select
                  v-model:value="settingsForm.language"
                  :options="languageOptions"
                  :placeholder="t('settingsTab.selectLanguage')"
                />
              </n-form-item>
            </n-grid-item>
            <n-grid-item>
              <n-form-item :label="t('settingsTab.themeMode')" label-placement="top">
                <n-select
                  v-model:value="settingsForm.theme"
                  :options="themeOptions"
                  :placeholder="t('settingsTab.selectTheme')"
                />
              </n-form-item>
            </n-grid-item>
          </n-grid>
        </div>

        <n-divider />

        <!-- 通知偏好 -->
        <div>
          <n-h4>{{ t('settingsTab.notificationPreference') }}</n-h4>
          <n-divider />
          <n-space vertical>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">{{ t('settingsTab.emailNotification') }}</span>
                <span class="setting-desc">{{ t('settingsTab.emailNotificationDesc') }}</span>
              </div>
              <n-switch v-model:value="settingsForm.notify_email" />
            </div>
          </n-space>
        </div>

        <n-divider />

        <n-space>
          <n-button type="primary" :loading="saving" @click="handleSaveSettings">
            {{ t('settingsTab.saveSettings') }}
          </n-button>
          <n-button @click="loadSettings">
            {{ t('settingsTab.reset') }}
          </n-button>
        </n-space>
      </n-space>
    </n-spin>
  </div>
</template>

<style scoped>
.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  background: var(--n-color);
}

.setting-info {
  flex: 1;
}

.setting-label {
  display: block;
  font-weight: 500;
  margin-bottom: 4px;
}

.setting-desc {
  color: var(--n-text-color-disabled);
  font-size: 14px;
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
