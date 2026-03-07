<script setup lang="ts">
import { useAuthStore } from '@/store'
import { fetchResetApiKey } from '@/service'
import NovaIcon from '@/components/common/NovaIcon.vue'

const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)

const showResetConfirm = ref(false)
const showApiKey = ref(false)

function copyApiKey() {
  if (userInfo.value?.apikey) {
    navigator.clipboard.writeText(userInfo.value.apikey)
    window.$message.success('API Key 已复制到剪贴板')
  }
  else {
    window.$message.warning('API Key 为空')
  }
}

async function confirmResetApiKey() {
  try {
    const response = await fetchResetApiKey()
    if (response.isSuccess) {
      window.$message.success('API Key 重置成功')
      authStore.updateUserInfo({ apikey: response.data.apikey })
      showResetConfirm.value = false
    }
    else {
      window.$message.error(response.message || '重置 API Key 失败')
    }
  }
  catch (error) {
    window.$message.error(`重置 API Key 失败: ${error}`)
  }
}
</script>

<template>
  <div class="p-4">
    <n-h4>API 管理</n-h4>
    <n-divider />

    <div class="api-section">
      <n-text depth="3" class="api-desc">
        API 密钥用于调用系统接口，请妥善保管，不要泄露给他人
      </n-text>

      <div class="api-key-container">
        <n-input
          :value="userInfo?.apikey || '暂无 API 密钥'"
          :type="showApiKey ? 'text' : 'password'"
          readonly
          placeholder="暂无 API 密钥"
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
                复制
              </n-button>
            </n-space>
          </template>
        </n-input>

        <n-button
          type="warning"
          class="reset-btn"
          @click="showResetConfirm = true"
        >
          重置密钥
        </n-button>
      </div>

      <n-alert type="warning" class="mt-4">
        <template #header>
          注意事项
        </template>
        <ul class="alert-list">
          <li>重置 API 密钥后，原密钥将立即失效</li>
          <li>请及时更新使用该密钥的应用程序</li>
        </ul>
      </n-alert>
    </div>

    <!-- 重置确认对话框 -->
    <n-modal v-model:show="showResetConfirm">
      <n-card
        style="width: 400px"
        title="确认重置 API 密钥"
        :bordered="false"
        size="huge"
        role="dialog"
        aria-modal="true"
      >
        <div class="confirm-content">
          <n-alert type="warning" :show-icon="false">
            重置后原 API 密钥将失效，请确认是否继续？
          </n-alert>
        </div>
        <template #footer>
          <div class="dialog-footer">
            <n-button @click="showResetConfirm = false">
              取消
            </n-button>
            <n-button type="warning" @click="confirmResetApiKey">
              确认重置
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
