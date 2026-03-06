<script setup lang="ts">
import { useAuthStore } from '@/store'
import { fetchChangePassword, fetchUpdateProfile } from '@/service'

const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)

const showPasswordModal = ref(false)
const showEmailModal = ref(false)
const showPhoneModal = ref(false)

const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const emailForm = ref({
  email: '',
})

const phoneForm = ref({
  mobile: '',
})

const profileForm = ref({
  nickname: '',
  avatar: '',
  gender: 0 as 0 | 1 | 2,
  birthday: null as number | null,
  motto: '',
})

watchEffect(() => {
  if (userInfo.value) {
    profileForm.value = {
      nickname: userInfo.value.nickname || '',
      avatar: userInfo.value.avatar || '',
      gender: userInfo.value.gender || 0,
      birthday: userInfo.value.birthday ? new Date(userInfo.value.birthday).getTime() : null,
      motto: userInfo.value.motto || '',
    }
    emailForm.value.email = userInfo.value.email || ''
    phoneForm.value.mobile = userInfo.value.mobile || ''
  }
})

const passwordChangeCountdown = ref(0)

async function handlePasswordSubmit() {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    window.$message.error('两次输入的密码不一致')
    return
  }
  if (!passwordForm.value.new_password || passwordForm.value.new_password.length < 6) {
    window.$message.error('新密码长度不能少于6位')
    return
  }
  try {
    const response = await fetchChangePassword({
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password,
    })
    if (response.isSuccess) {
      passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
      passwordChangeCountdown.value = 3
      const countdownInterval = setInterval(() => {
        passwordChangeCountdown.value--
        if (passwordChangeCountdown.value <= 0) {
          clearInterval(countdownInterval)
          showPasswordModal.value = false
          authStore.logout()
        }
      }, 1000)
    }
    else {
      window.$message.error(response.message || '密码修改失败')
    }
  }
  catch (error) {
    window.$message.error(`密码修改失败: ${error}`)
  }
}

async function handleEmailSubmit() {
  try {
    const response = await fetchUpdateProfile({ email: emailForm.value.email })
    if (response.isSuccess) {
      window.$message.success('邮箱修改成功')
      showEmailModal.value = false
      authStore.updateUserInfo({ email: emailForm.value.email })
    }
    else {
      window.$message.error(response.message || '邮箱修改失败')
    }
  }
  catch (error) {
    window.$message.error(`邮箱修改失败: ${error}`)
  }
}

async function handlePhoneSubmit() {
  try {
    const response = await fetchUpdateProfile({ mobile: phoneForm.value.mobile })
    if (response.isSuccess) {
      window.$message.success('手机号修改成功')
      showPhoneModal.value = false
      authStore.updateUserInfo({ mobile: phoneForm.value.mobile })
    }
    else {
      window.$message.error(response.message || '手机号修改失败')
    }
  }
  catch (error) {
    window.$message.error(`手机号修改失败: ${error}`)
  }
}

async function handleProfileSubmit() {
  try {
    const submitData = {
      nickname: profileForm.value.nickname,
      avatar: profileForm.value.avatar,
      gender: profileForm.value.gender,
      birthday: profileForm.value.birthday ? new Date(profileForm.value.birthday).getTime() : null,
      motto: profileForm.value.motto,
    }
    const response = await fetchUpdateProfile(submitData)
    if (response.isSuccess) {
      window.$message.success('个人资料保存成功')
      authStore.updateUserInfo({
        nickname: submitData.nickname,
        avatar: submitData.avatar,
        gender: submitData.gender as 0 | 1 | 2,
        birthday: submitData.birthday,
        motto: submitData.motto,
      })
    }
    else {
      window.$message.error(response.message || '个人资料保存失败')
    }
  }
  catch (error) {
    window.$message.error(`个人资料保存失败: ${error}`)
  }
}
</script>

<template>
  <div class="p-4">
    <n-space vertical size="large">
      <!-- 基本信息 -->
      <div>
        <n-h4>基本信息</n-h4>
        <n-divider />
        <n-grid cols="1 s:2 m:3" :x-gap="32" :y-gap="0" responsive="screen">
          <n-grid-item>
            <n-form-item label="用户ID" label-placement="top">
              <n-input :value="userInfo?.id?.toString()" readonly disabled />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="用户名" label-placement="top">
              <n-input :value="userInfo?.userName" readonly disabled />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="昵称" label-placement="top">
              <n-input v-model:value="profileForm.nickname" placeholder="请输入昵称" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="性别" label-placement="top">
              <n-radio-group v-model:value="profileForm.gender">
                <n-radio :value="0">
                  保密
                </n-radio>
                <n-radio :value="1">
                  男
                </n-radio>
                <n-radio :value="2">
                  女
                </n-radio>
              </n-radio-group>
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="生日" label-placement="top">
              <n-date-picker v-model:value="profileForm.birthday" type="date" placeholder="请选择生日" class="w-full" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item :span="3">
            <n-form-item label="个性签名" label-placement="top">
              <n-input v-model:value="profileForm.motto" type="textarea" placeholder="请输入个性签名" />
            </n-form-item>
          </n-grid-item>
        </n-grid>
        <n-space>
          <n-button type="primary" @click="handleProfileSubmit">
            保存修改
          </n-button>
        </n-space>
      </div>
      <n-divider />

      <!-- 安全设置 -->
      <div>
        <n-h4>安全设置</n-h4>
        <n-space vertical>
          <div class="security-item">
            <div class="security-info">
              <span class="security-label">登录密码</span>
              <span class="security-desc">用于登录账户的密码</span>
            </div>
            <n-button type="warning" @click="showPasswordModal = true">
              修改密码
            </n-button>
          </div>

          <div class="security-item">
            <div class="security-info">
              <span class="security-label">邮箱地址</span>
              <span class="security-desc">{{ userInfo?.email || '未绑定邮箱' }}</span>
            </div>
            <n-button @click="showEmailModal = true">
              {{ userInfo?.email ? '修改邮箱' : '绑定邮箱' }}
            </n-button>
          </div>

          <div class="security-item">
            <div class="security-info">
              <span class="security-label">手机号码</span>
              <span class="security-desc">{{ userInfo?.mobile || '未绑定手机号' }}</span>
            </div>
            <n-button @click="showPhoneModal = true">
              {{ userInfo?.mobile ? '修改手机号' : '绑定手机号' }}
            </n-button>
          </div>
        </n-space>
      </div>

      <!-- 登录信息 -->
      <n-divider />
      <div>
        <n-h4>登录信息</n-h4>
        <n-descriptions :column="1" bordered label-placement="left" class="login-info-desc">
          <n-descriptions-item label="注册时间">
            {{ userInfo?.createTime ? new Date(userInfo.createTime * 1000).toLocaleString() : 'N/A' }}
          </n-descriptions-item>
          <n-descriptions-item label="最后登录">
            {{ userInfo?.lastLoginTime ? new Date(userInfo.lastLoginTime * 1000).toLocaleString() : '从未登录' }}
          </n-descriptions-item>
          <n-descriptions-item label="注册IP">
            {{ userInfo?.joinIp || 'N/A' }}
          </n-descriptions-item>
          <n-descriptions-item label="最后登录IP">
            {{ userInfo?.lastLoginIp || 'N/A' }}
          </n-descriptions-item>
          <n-descriptions-item label="更新时间">
            {{ userInfo?.updateTime ? new Date(userInfo.updateTime * 1000).toLocaleString() : 'N/A' }}
          </n-descriptions-item>
        </n-descriptions>
      </div>
    </n-space>

    <!-- 修改密码弹窗 -->
    <n-modal
      v-model:show="showPasswordModal"
      preset="dialog"
      title="修改密码"
      :mask-closable="passwordChangeCountdown === 0"
      :closable="passwordChangeCountdown === 0"
    >
      <div v-if="passwordChangeCountdown > 0" class="text-center py-6">
        <n-result status="success" title="密码修改成功">
          <template #footer>
            <n-text type="warning">
              {{ passwordChangeCountdown }} 秒后自动退出登录，请使用新密码重新登录
            </n-text>
          </template>
        </n-result>
      </div>
      <n-form v-else :model="passwordForm" label-placement="left" label-width="100px">
        <n-form-item label="当前密码" required>
          <n-input
            v-model:value="passwordForm.old_password"
            type="password"
            placeholder="请输入当前密码"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="新密码" required>
          <n-input
            v-model:value="passwordForm.new_password"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="确认密码" required>
          <n-input
            v-model:value="passwordForm.confirm_password"
            type="password"
            placeholder="请再次输入新密码"
            show-password-on="click"
          />
        </n-form-item>
      </n-form>
      <template v-if="passwordChangeCountdown === 0" #action>
        <n-space>
          <n-button @click="showPasswordModal = false">
            取消
          </n-button>
          <n-button type="primary" @click="handlePasswordSubmit">
            确认修改
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 修改邮箱弹窗 -->
    <n-modal v-model:show="showEmailModal" preset="dialog" title="修改邮箱">
      <n-form :model="emailForm" label-placement="left" label-width="100px">
        <n-form-item label="邮箱地址" required>
          <n-input
            v-model:value="emailForm.email"
            placeholder="请输入邮箱地址"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showEmailModal = false">
            取消
          </n-button>
          <n-button type="primary" @click="handleEmailSubmit">
            确认修改
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 修改手机号弹窗 -->
    <n-modal v-model:show="showPhoneModal" preset="dialog" title="修改手机号">
      <n-form :model="phoneForm" label-placement="left" label-width="100px">
        <n-form-item label="手机号码" required>
          <n-input
            v-model:value="phoneForm.mobile"
            placeholder="请输入手机号码"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showPhoneModal = false">
            取消
          </n-button>
          <n-button type="primary" @click="handlePhoneSubmit">
            确认修改
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  background: var(--n-color);
}

.security-info {
  flex: 1;
}

.security-label {
  display: block;
  font-weight: 500;
  margin-bottom: 4px;
}

.security-desc {
  color: var(--n-text-color-disabled);
  font-size: 14px;
}

.login-info-desc {
  max-width: 600px;
}

@media (min-width: 768px) {
  .login-info-desc {
    --n-column: 2 !important;
  }
}

@media (max-width: 768px) {
  .security-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .security-info {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .security-item {
    padding: 12px;
  }
}
</style>
