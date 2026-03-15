<script setup lang="ts">
import { useAuthStore } from '@/store'
import { fetchUserProfile, fetchUpdateProfile } from '@/service'
import ProfileTab from './components/ProfileTab.vue'
import ApiTab from './components/ApiTab.vue'
import SettingsTab from './components/SettingsTab.vue'
import SecurityTab from './components/SecurityTab.vue'
import MoneyScoreTab from './components/MoneyScoreTab.vue'
import NovaIcon from '@/components/common/NovaIcon.vue'

const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)

const headerCardStyle = computed(() => {
  const bg = userInfo.value?.backGround
  if (!bg || !/^https?:\/\//i.test(bg)) return {}
  const safeBg = globalThis.CSS?.escape?.(bg) ?? bg.replace(/[()'"\\]/g, '')
  return {
    backgroundImage: `linear-gradient(rgba(255,255,255,0.85), rgba(255,255,255,0.92)), url("${safeBg}")`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  }
})

const activeTab = ref('profile')

const showAvatarModal = ref(false)
const avatarForm = ref({
  currentAvatar: '',
  newAvatar: '',
})

const showMottoModal = ref(false)
const mottoForm = ref('')

const showBgModal = ref(false)
const bgForm = ref({
  currentBg: '',
  newBg: '',
})

function openAvatarModal() {
  avatarForm.value.currentAvatar = userInfo.value?.avatar || ''
  avatarForm.value.newAvatar = userInfo.value?.avatar || ''
  showAvatarModal.value = true
}

async function handleAvatarSubmit() {
  try {
    const nextAvatar = avatarForm.value.newAvatar.trim()
    if (nextAvatar && !/^https?:\/\//i.test(nextAvatar)) {
      window.$message.error('头像 URL 仅支持 http/https 协议')
      return
    }
    const response = await fetchUpdateProfile({ avatar: nextAvatar })
    if (response.isSuccess) {
      authStore.updateUserInfo({ avatar: nextAvatar })
      showAvatarModal.value = false
      window.$message.success('头像更新成功')
    }
  }
  catch (error) {
    console.error('更新头像失败', error)
    window.$message.error('更新头像失败')
  }
}

function openMottoModal() {
  mottoForm.value = userInfo.value?.motto || ''
  showMottoModal.value = true
}

async function handleMottoSubmit() {
  try {
    const nextMotto = mottoForm.value.trim()
    if (nextMotto.length > 200) {
      window.$message.error('个性签名不能超过200个字符')
      return
    }
    const response = await fetchUpdateProfile({ motto: nextMotto })
    if (response.isSuccess) {
      authStore.updateUserInfo({ motto: nextMotto })
      showMottoModal.value = false
      window.$message.success('签名更新成功')
    }
  }
  catch (error) {
    console.error('更新签名失败', error)
    window.$message.error('更新签名失败')
  }
}

function openBgModal() {
  bgForm.value.currentBg = userInfo.value?.backGround || ''
  bgForm.value.newBg = userInfo.value?.backGround || ''
  showBgModal.value = true
}

async function handleBgSubmit() {
  try {
    const nextBg = bgForm.value.newBg.trim()
    if (nextBg && !/^https?:\/\//i.test(nextBg)) {
      window.$message.error('背景图 URL 仅支持 http/https 协议')
      return
    }
    const response = await fetchUpdateProfile({ back_ground: nextBg })
    if (response.isSuccess) {
      authStore.updateUserInfo({ backGround: nextBg })
      showBgModal.value = false
      window.$message.success('背景图更新成功')
    }
  }
  catch (error) {
    console.error('更新背景图失败', error)
    window.$message.error('更新背景图失败')
  }
}

async function refreshUserInfo() {
  try {
    const response = await fetchUserProfile()
    if (response.isSuccess && response.data) {
      authStore.updateUserInfo(response.data)
    }
  }
  catch (error) {
    console.error('获取用户信息失败', error)
  }
}

watch(activeTab, () => {
  refreshUserInfo()
})

onMounted(() => {
  refreshUserInfo()
})

onActivated(() => {
  refreshUserInfo()
})
</script>

<template>
  <div class="user-center">
    <!-- 用户信息头部 -->
    <n-card class="user-header mb-4" :style="headerCardStyle">
      <div class="user-info-container">
        <div class="user-avatar-section">
          <n-avatar
            :size="80"
            :src="userInfo?.avatar"
            :img-props="{ referrerpolicy: 'no-referrer' }"
            class="user-avatar clickable-avatar"
            @click="openAvatarModal"
          >
            <NovaIcon v-if="!userInfo?.avatar" icon="icon-park-outline:user" :size="40" />
          </n-avatar>
        </div>

        <div class="user-details-section">
          <n-h3 class="user-name">
            <span class="user-name-text">{{ userInfo?.nickname || userInfo?.userName || '用户' }}</span>
            <n-text v-if="userInfo?.userName" depth="3" class="user-name-account">
              ({{ userInfo.userName }})
            </n-text>
          </n-h3>
          <n-text depth="3" class="user-email ml-2">
            {{ userInfo?.email || '暂无邮箱' }}
          </n-text>
          <n-space class="ml-3" size="small">
            <n-tag type="info" size="small">
              ID: {{ userInfo?.id || 'N/A' }}
            </n-tag>
            <n-tag type="warning" size="small">
              余额: ¥{{ userInfo?.money ? Number(userInfo.money).toFixed(2) : '0.00' }}
            </n-tag>
            <n-tag type="primary" size="small">
              积分: {{ userInfo?.score || '0' }}
            </n-tag>
          </n-space>

          <n-grid class="mt-2 ml-3" cols="2 m:4" :x-gap="3" :y-gap="12" responsive="screen" style="padding: 1%;">
            <n-grid-item>
              <n-text depth="3" class="info-item">
                <NovaIcon class="info-icon" icon="icon-park-outline:level" :size="16" />
                等级: {{ userInfo?.level || '0' }}
              </n-text>
            </n-grid-item>
            <n-grid-item>
              <n-text depth="3" class="info-item">
                <NovaIcon class="info-icon" icon="icon-park-outline:crown" :size="16" />
                角色: {{ userInfo?.role?.includes('admin') ? '管理员' : '普通用户' }}
              </n-text>
            </n-grid-item>
            <n-grid-item>
              <n-text depth="3" class="info-item">
                <NovaIcon class="info-icon" icon="icon-park-outline:check-one" :size="16" />
                状态: {{ userInfo?.status === 1 ? '正常' : '禁用' }}
              </n-text>
            </n-grid-item>
            <n-grid-item>
              <n-text depth="3" class="info-item clickable-item" @click="openMottoModal">
                <NovaIcon class="info-icon" icon="icon-park-outline:quote" :size="16" />
                {{ userInfo?.motto || '暂无签名' }}
                <NovaIcon class="edit-icon" icon="icon-park-outline:edit" :size="12" />
              </n-text>
            </n-grid-item>
          </n-grid>
        </div>

        <div class="user-actions-section">
          <n-space vertical class="w-full">
            <n-button type="primary" block @click="activeTab = 'profile'">
              编辑资料
            </n-button>
            <n-button block @click="openBgModal">
              设置背景图
            </n-button>
          </n-space>
        </div>
      </div>
    </n-card>

    <!-- 标签栏导航 -->
    <n-card>
      <n-tabs
        v-model:value="activeTab"
        type="line"
        animated
      >
        <n-tab-pane name="profile" tab="个人资料">
          <ProfileTab />
        </n-tab-pane>
        <n-tab-pane name="settings" tab="偏好设置">
          <SettingsTab />
        </n-tab-pane>
        <n-tab-pane name="security" tab="安全管理">
          <SecurityTab />
        </n-tab-pane>
        <n-tab-pane name="api" tab="API 管理">
          <ApiTab />
        </n-tab-pane>
        <n-tab-pane name="moneyScore" tab="余额与积分">
          <MoneyScoreTab />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 头像修改对话框 -->
    <n-modal v-model:show="showAvatarModal" preset="dialog" title="修改头像">
      <n-space vertical size="large">
        <div>
          <n-text depth="3">
            当前头像
          </n-text>
          <div class="avatar-preview mt-2">
            <n-avatar :size="80" :src="avatarForm.currentAvatar" :img-props="{ referrerpolicy: 'no-referrer' }">
              <NovaIcon v-if="!avatarForm.currentAvatar" icon="icon-park-outline:user" :size="40" />
            </n-avatar>
            <n-input
              :value="avatarForm.currentAvatar"
              readonly
              placeholder="当前头像 URL"
              class="mt-2"
              disabled
            />
          </div>
        </div>

        <n-divider />

        <div>
          <n-text depth="3">
            新头像
          </n-text>
          <div class="avatar-preview mt-2">
            <n-avatar :size="80" :src="avatarForm.newAvatar" :img-props="{ referrerpolicy: 'no-referrer' }">
              <NovaIcon v-if="!avatarForm.newAvatar" icon="icon-park-outline:user" :size="40" />
            </n-avatar>
            <n-input
              v-model:value="avatarForm.newAvatar"
              type="textarea"
              placeholder="请输入新的头像 URL（最多250字符）"
              :maxlength="250"
              :rows="3"
              class="mt-2"
            />
          </div>
        </div>
      </n-space>

      <template #action>
        <n-space>
          <n-button @click="showAvatarModal = false">
            取消
          </n-button>
          <n-button type="primary" @click="handleAvatarSubmit">
            保存
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 个性签名修改弹窗 -->
    <n-modal v-model:show="showMottoModal" preset="dialog" title="修改个性签名">
      <n-input
        v-model:value="mottoForm"
        type="textarea"
        placeholder="请输入个性签名（最多200字符）"
        :maxlength="200"
        show-count
        :rows="3"
      />
      <template #action>
        <n-space>
          <n-button @click="showMottoModal = false">
            取消
          </n-button>
          <n-button type="primary" @click="handleMottoSubmit">
            保存
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 背景图修改弹窗 -->
    <n-modal v-model:show="showBgModal" preset="dialog" title="设置背景图">
      <n-space vertical size="large">
        <div v-if="bgForm.currentBg">
          <n-text depth="3">当前背景</n-text>
          <div class="bg-preview mt-2">
            <img :src="bgForm.currentBg" referrerpolicy="no-referrer" class="bg-preview-img" />
          </div>
        </div>
        <div>
          <n-text depth="3">新背景图 URL</n-text>
          <n-input
            v-model:value="bgForm.newBg"
            type="textarea"
            placeholder="请输入背景图 URL（http/https，最多500字符，留空则清除）"
            :maxlength="500"
            :rows="3"
            class="mt-2"
          />
          <div v-if="bgForm.newBg" class="bg-preview mt-2">
            <n-text depth="3" class="mb-1">预览</n-text>
            <img :src="bgForm.newBg" referrerpolicy="no-referrer" class="bg-preview-img" />
          </div>
        </div>
      </n-space>
      <template #action>
        <n-space>
          <n-button @click="showBgModal = false">
            取消
          </n-button>
          <n-button type="primary" @click="handleBgSubmit">
            保存
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.user-info-container {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.user-avatar-section {
  flex-shrink: 0;
}

.user-details-section {
  flex: 1;
  min-width: 0;
}

.user-name {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  word-break: break-word;
}

.user-name-text {
  display: inline-flex;
  align-items: center;
}

.user-name-account {
  display: inline-flex;
  align-items: center;
  font-size: 14px;
}

.user-email {
  display: block;
  margin-bottom: 12px;
  word-break: break-all;
}

.user-actions-section {
  flex-shrink: 0;
  min-width: 120px;
}

.info-item {
  display: flex;
  align-items: center;
  font-size: 12px;
}

.info-icon {
  margin-right: 6px;
  font-size: 14px;
}

.clickable-avatar {
  cursor: pointer;
  transition: all 0.3s ease;
}

.clickable-avatar:hover {
  transform: scale(1.05);
  filter: brightness(1.1);
}

.clickable-item {
  cursor: pointer;
  transition: color 0.2s ease;
}

.clickable-item:hover {
  color: var(--n-text-color);
}

.edit-icon {
  margin-left: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.clickable-item:hover .edit-icon {
  opacity: 1;
}

.avatar-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.bg-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.bg-preview-img {
  max-width: 100%;
  max-height: 160px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid var(--n-border-color);
}

@media (max-width: 768px) {
  .user-header :deep(.n-card__content) {
    padding: 16px;
  }

  .user-info-container {
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 16px;
  }

  .user-avatar-section {
    order: 1;
  }

  .user-details-section {
    order: 2;
    width: 100%;
  }

  .user-actions-section {
    order: 3;
    width: 100%;
    min-width: unset;
  }

  .user-actions-section .n-space {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .user-header :deep(.n-card__content) {
    padding: 12px;
  }

  .user-avatar-section .user-avatar {
    --n-size: 60px !important;
  }

  .user-name {
    font-size: 18px;
  }
}
</style>
