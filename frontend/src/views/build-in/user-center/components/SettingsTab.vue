<script setup lang="ts">
import { useAuthStore } from '@/store'
import { fetchUserSettings, updateUserSettings } from '@/service'

const authStore = useAuthStore()

const loading = ref(false)
const saving = ref(false)

const settingsForm = ref({
  language: 'zh-CN',
  theme: 'light',
  notify_email: true,
})

const languageOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' },
]

const themeOptions = [
  { label: '浅色模式', value: 'light' },
  { label: '深色模式', value: 'dark' },
  { label: '跟随系统', value: 'auto' },
]

async function loadSettings() {
  loading.value = true
  try {
    const response = await fetchUserSettings()
    if (response.isSuccess && response.data) {
      settingsForm.value = {
        language: response.data.language || 'zh-CN',
        theme: response.data.theme || 'light',
        notify_email: response.data.notify_email ?? true,
      }
    }
  }
  catch (error) {
    console.error('获取设置失败', error)
  }
  finally {
    loading.value = false
  }
}

async function handleSaveSettings() {
  saving.value = true
  try {
    const response = await updateUserSettings({
      language: settingsForm.value.language,
      theme: settingsForm.value.theme,
      notify_email: settingsForm.value.notify_email,
    })
    if (response.isSuccess) {
      window.$message.success('设置保存成功')
    }
    else {
      window.$message.error(response.message || '设置保存失败')
    }
  }
  catch (error) {
    window.$message.error(`设置保存失败: ${error}`)
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
          <n-h4>显示偏好</n-h4>
          <n-divider />
          <n-grid cols="1 s:2" :x-gap="32" :y-gap="0" responsive="screen">
            <n-grid-item>
              <n-form-item label="界面语言" label-placement="top">
                <n-select
                  v-model:value="settingsForm.language"
                  :options="languageOptions"
                  placeholder="选择语言"
                />
              </n-form-item>
            </n-grid-item>
            <n-grid-item>
              <n-form-item label="主题模式" label-placement="top">
                <n-select
                  v-model:value="settingsForm.theme"
                  :options="themeOptions"
                  placeholder="选择主题"
                />
              </n-form-item>
            </n-grid-item>
          </n-grid>
        </div>

        <n-divider />

        <!-- 通知偏好 -->
        <div>
          <n-h4>通知偏好</n-h4>
          <n-divider />
          <n-space vertical>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">邮件通知</span>
                <span class="setting-desc">接收系统通知和安全提醒邮件</span>
              </div>
              <n-switch v-model:value="settingsForm.notify_email" />
            </div>
          </n-space>
        </div>

        <n-divider />

        <n-space>
          <n-button type="primary" :loading="saving" @click="handleSaveSettings">
            保存设置
          </n-button>
          <n-button @click="loadSettings">
            重置
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
