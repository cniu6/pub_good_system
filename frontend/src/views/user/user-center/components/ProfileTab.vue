<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store'
import { fetchChangePassword, fetchUpdateProfile } from '@/service'
import { sendEmailChangeCode, verifyEmailChange, sendPhoneChangeCode, verifyPhoneChange } from '@/service'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'

const authStore = useAuthStore()
const { t } = useI18n()

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
  code: '',
})
const emailStep = ref<'input' | 'verify'>('input')
const emailCodeCountdown = ref(0)
let emailCodeTimer: ReturnType<typeof setInterval> | null = null

const phoneForm = ref({
  mobile: '',
  code: '',
})
const phoneStep = ref<'input' | 'verify'>('input')
const phoneCodeCountdown = ref(0)
let phoneCodeTimer: ReturnType<typeof setInterval> | null = null

const profileForm = ref({
  nickname: '',
  avatar: '',
  gender: 0 as 0 | 1 | 2,
  birthday: null as number | null,
  motto: '',
  back_ground: '',
})

// 极验相关
const isGeetestEnabled = computed(() => geetestManager.isEnabled())
const emailGeetestRef = ref<any>(null)
const phoneGeetestRef = ref<any>(null)
const emailCaptchaKey = ref(0)
const phoneCaptchaKey = ref(0)

watchEffect(() => {
  if (userInfo.value) {
    profileForm.value = {
      nickname: userInfo.value.nickname || '',
      avatar: userInfo.value.avatar || '',
      gender: userInfo.value.gender ?? 0,
      birthday: userInfo.value.birthday ? new Date(userInfo.value.birthday).getTime() : null,
      motto: userInfo.value.motto || '',
      back_ground: userInfo.value.backGround || '',
    }
    emailForm.value.email = userInfo.value.email || ''
    phoneForm.value.mobile = userInfo.value.mobile || ''
  }
})

const passwordChangeCountdown = ref(0)

async function handlePasswordSubmit() {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    window.$message.error(t('profile.passwordMismatch'))
    return
  }
  if (!passwordForm.value.new_password || passwordForm.value.new_password.length < 8) {
    window.$message.error(t('profile.passwordTooShort'))
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
      window.$message.error(response.message || t('profile.changePasswordFailed'))
    }
  }
  catch (error) {
    window.$message.error(`${t('profile.changePasswordFailed')}: ${error}`)
  }
}

// ========== 邮箱验证流程 ==========

function openEmailModal() {
  emailForm.value = { email: '', code: '' }
  emailStep.value = 'input'
  emailCodeCountdown.value = 0
  showEmailModal.value = true
}

function triggerSendEmailCode() {
  if (!emailForm.value.email) {
    window.$message.error(t('profile.enterNewEmail'))
    return
  }
  if (isGeetestEnabled.value) {
    emailGeetestRef.value?.showCaptcha()
  }
  else {
    doSendEmailCode()
  }
}

function onEmailGeetestSuccess() {
  doSendEmailCode()
}

function onEmailGeetestError() {
  geetestManager.clearCaptchaResult()
  emailCaptchaKey.value++
}

async function doSendEmailCode() {
  try {
    const response = await sendEmailChangeCode({ new_email: emailForm.value.email })
    if (response.isSuccess) {
      // 验证关闭时后端直接修改完成
      if (response.data?.verified) {
        window.$message.success(t('profile.emailUpdated'))
        showEmailModal.value = false
        authStore.updateUserInfo({ email: emailForm.value.email })
        return
      }
      window.$message.success(t('profile.emailCodeSent'))
      emailStep.value = 'verify'
      emailCodeCountdown.value = 60
      emailCodeTimer = setInterval(() => {
        emailCodeCountdown.value--
        if (emailCodeCountdown.value <= 0) {
          if (emailCodeTimer) clearInterval(emailCodeTimer)
          emailCodeTimer = null
        }
      }, 1000)
    }
    else {
      window.$message.error(response.message || t('profile.sendCodeFailed'))
    }
  }
  catch (error) {
    window.$message.error(`${t('profile.sendCodeFailed')}: ${error}`)
  }
}

async function handleVerifyEmailChange() {
  if (!emailForm.value.code) {
    window.$message.error(t('profile.enterVerificationCode'))
    return
  }
  try {
    const response = await verifyEmailChange({
      new_email: emailForm.value.email,
      code: emailForm.value.code,
    })
    if (response.isSuccess) {
      window.$message.success(t('profile.emailUpdated'))
      showEmailModal.value = false
      authStore.updateUserInfo({ email: emailForm.value.email })
      if (emailCodeTimer) clearInterval(emailCodeTimer)
    }
    else {
      window.$message.error(response.message || t('profile.invalidOrExpiredCode'))
    }
  }
  catch (error) {
    window.$message.error(`${t('profile.emailVerifyFailed')}: ${error}`)
  }
}

// ========== 手机验证流程 ==========

function openPhoneModal() {
  phoneForm.value = { mobile: '', code: '' }
  phoneStep.value = 'input'
  phoneCodeCountdown.value = 0
  showPhoneModal.value = true
}

function triggerSendPhoneCode() {
  if (!phoneForm.value.mobile) {
    window.$message.error(t('profile.enterNewPhone'))
    return
  }
  if (isGeetestEnabled.value) {
    phoneGeetestRef.value?.showCaptcha()
  }
  else {
    doSendPhoneCode()
  }
}

function onPhoneGeetestSuccess() {
  doSendPhoneCode()
}

function onPhoneGeetestError() {
  geetestManager.clearCaptchaResult()
  phoneCaptchaKey.value++
}

async function doSendPhoneCode() {
  try {
    const response = await sendPhoneChangeCode({ new_mobile: phoneForm.value.mobile })
    if (response.isSuccess) {
      // 验证关闭时后端直接修改完成
      if (response.data?.verified) {
        window.$message.success(t('profile.phoneUpdated'))
        showPhoneModal.value = false
        authStore.updateUserInfo({ mobile: phoneForm.value.mobile })
        return
      }
      window.$message.success(t('profile.codeSent'))
      phoneStep.value = 'verify'
      phoneCodeCountdown.value = 60
      phoneCodeTimer = setInterval(() => {
        phoneCodeCountdown.value--
        if (phoneCodeCountdown.value <= 0) {
          if (phoneCodeTimer) clearInterval(phoneCodeTimer)
          phoneCodeTimer = null
        }
      }, 1000)
    }
    else {
      window.$message.error(response.message || t('profile.sendCodeFailed'))
    }
  }
  catch (error) {
    window.$message.error(`${t('profile.sendCodeFailed')}: ${error}`)
  }
}

async function handleVerifyPhoneChange() {
  if (!phoneForm.value.code) {
    window.$message.error(t('profile.enterVerificationCode'))
    return
  }
  try {
    const response = await verifyPhoneChange({
      new_mobile: phoneForm.value.mobile,
      code: phoneForm.value.code,
    })
    if (response.isSuccess) {
      window.$message.success(t('profile.phoneUpdated'))
      showPhoneModal.value = false
      authStore.updateUserInfo({ mobile: phoneForm.value.mobile })
      if (phoneCodeTimer) clearInterval(phoneCodeTimer)
    }
    else {
      window.$message.error(response.message || t('profile.invalidOrExpiredCode'))
    }
  }
  catch (error) {
    window.$message.error(`${t('profile.phoneVerifyFailed')}: ${error}`)
  }
}

// ========== 基本资料 ==========

async function handleProfileSubmit() {
  try {
    const submitData = {
      nickname: profileForm.value.nickname,
      avatar: profileForm.value.avatar,
      gender: profileForm.value.gender,
      birthday: profileForm.value.birthday ? Math.floor(Number(profileForm.value.birthday) / 1000) : null,
      motto: profileForm.value.motto,
      back_ground: profileForm.value.back_ground,
    }
    const response = await fetchUpdateProfile(submitData)
    if (response.isSuccess) {
      window.$message.success(t('profile.profileSaved'))
      authStore.updateUserInfo({
        nickname: submitData.nickname,
        avatar: submitData.avatar,
        gender: submitData.gender as 0 | 1 | 2,
        birthday: submitData.birthday,
        motto: submitData.motto,
        backGround: submitData.back_ground,
      })
    }
    else {
      window.$message.error(response.message || t('profile.profileSaveFailed'))
    }
  }
  catch (error) {
    window.$message.error(`${t('profile.profileSaveFailed')}: ${error}`)
  }
}
</script>

<template>
  <div class="p-4">
    <n-space vertical size="large">
      <!-- 基本信息 -->
      <div>
        <n-h4>{{ t('profile.basicInfo') }}</n-h4>
        <n-divider />
        <n-grid cols="1 s:2 m:3" :x-gap="32" :y-gap="0" responsive="screen">
          <n-grid-item>
            <n-form-item :label="t('profile.userId')" label-placement="top">
              <n-input :value="userInfo?.id?.toString()" readonly disabled />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item :label="t('profile.username')" label-placement="top">
              <n-input :value="userInfo?.userName" readonly disabled />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item :label="t('profile.nickname')" label-placement="top">
              <n-input v-model:value="profileForm.nickname" :placeholder="t('profile.enterNickname')" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item :label="t('profile.gender')" label-placement="top">
              <n-radio-group v-model:value="profileForm.gender">
                <n-radio :value="0">
                  {{ t('profile.genderSecret') }}
                </n-radio>
                <n-radio :value="1">
                  {{ t('profile.genderMale') }}
                </n-radio>
                <n-radio :value="2">
                  {{ t('profile.genderFemale') }}
                </n-radio>
              </n-radio-group>
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item :label="t('profile.birthday')" label-placement="top">
              <n-date-picker v-model:value="profileForm.birthday" type="date" :placeholder="t('profile.selectBirthday')" class="w-full" />
            </n-form-item>
          </n-grid-item>
        </n-grid>
        <n-space>
          <n-button type="primary" @click="handleProfileSubmit">
            {{ t('profile.saveChanges') }}
          </n-button>
        </n-space>
      </div>
      <n-divider />

      <!-- 安全设置 -->
      <div>
        <n-h4>{{ t('profile.securitySettings') }}</n-h4>
        <n-space vertical>
          <div class="security-item">
            <div class="security-info">
              <span class="security-label">{{ t('profile.loginPassword') }}</span>
              <span class="security-desc">{{ t('profile.loginPasswordDesc') }}</span>
            </div>
            <n-button type="warning" @click="showPasswordModal = true">
              {{ t('profile.changePassword') }}
            </n-button>
          </div>

          <div class="security-item">
            <div class="security-info">
              <span class="security-label">{{ t('profile.emailAddress') }}</span>
              <span class="security-desc">{{ userInfo?.email || t('profile.emailUnbound') }}</span>
            </div>
            <n-button @click="openEmailModal">
              {{ userInfo?.email ? t('profile.changeEmail') : t('profile.bindEmail') }}
            </n-button>
          </div>

          <div class="security-item">
            <div class="security-info">
              <span class="security-label">{{ t('profile.phoneNumber') }}</span>
              <span class="security-desc">{{ userInfo?.mobile || t('profile.phoneUnbound') }}</span>
            </div>
            <n-button @click="openPhoneModal">
              {{ userInfo?.mobile ? t('profile.changePhone') : t('profile.bindPhone') }}
            </n-button>
          </div>
        </n-space>
      </div>

      <!-- 登录信息 -->
      <n-divider />
      <div>
        <n-h4>{{ t('profile.loginInfo') }}</n-h4>
        <n-descriptions :column="1" bordered label-placement="left" class="login-info-desc">
          <n-descriptions-item :label="t('profile.registerTime')">
            {{ userInfo?.createTime ? new Date(userInfo.createTime * 1000).toLocaleString() : t('profile.na') }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('profile.lastLogin')">
            {{ userInfo?.lastLoginTime ? new Date(userInfo.lastLoginTime * 1000).toLocaleString() : t('profile.neverLoggedIn') }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('profile.registerIp')">
            {{ userInfo?.joinIp || t('profile.na') }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('profile.lastLoginIp')">
            {{ userInfo?.lastLoginIp || t('profile.na') }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('profile.updateTime')">
            {{ userInfo?.updateTime ? new Date(userInfo.updateTime * 1000).toLocaleString() : t('profile.na') }}
          </n-descriptions-item>
        </n-descriptions>
      </div>
    </n-space>

    <!-- 修改密码弹窗 -->
    <n-modal
      v-model:show="showPasswordModal"
      preset="dialog"
      :title="t('profile.changePassword')"
      :mask-closable="passwordChangeCountdown === 0"
      :closable="passwordChangeCountdown === 0"
    >
      <div v-if="passwordChangeCountdown > 0" class="text-center py-6">
        <n-result status="success" :title="t('profile.passwordChanged')">
          <template #footer>
            <n-text type="warning">
              {{ t('profile.autoLogoutCountdown', { seconds: passwordChangeCountdown }) }}
            </n-text>
          </template>
        </n-result>
      </div>
      <n-form v-else :model="passwordForm" label-placement="left" label-width="100px">
        <n-form-item :label="t('profile.currentPassword')" required>
          <n-input
            v-model:value="passwordForm.old_password"
            type="password"
            :placeholder="t('profile.enterCurrentPassword')"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item :label="t('profile.newPassword')" required>
          <n-input
            v-model:value="passwordForm.new_password"
            type="password"
            :placeholder="t('profile.enterNewPassword')"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item :label="t('profile.confirmPassword')" required>
          <n-input
            v-model:value="passwordForm.confirm_password"
            type="password"
            :placeholder="t('profile.enterConfirmPassword')"
            show-password-on="click"
          />
        </n-form-item>
      </n-form>
      <template v-if="passwordChangeCountdown === 0" #action>
        <n-space>
          <n-button @click="showPasswordModal = false">
            {{ t('common.cancel') }}
          </n-button>
          <n-button type="primary" @click="handlePasswordSubmit">
            {{ t('profile.confirmChange') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 修改邮箱弹窗（验证码流程） -->
    <n-modal v-model:show="showEmailModal" preset="dialog" :title="t('profile.changeEmail')">
      <n-form :model="emailForm" label-placement="left" label-width="100px">
        <n-form-item :label="t('profile.newEmail')" required>
          <n-input
            v-model:value="emailForm.email"
            :placeholder="t('profile.enterNewEmail')"
            :disabled="emailStep === 'verify'"
          />
        </n-form-item>
        <n-form-item v-if="emailStep === 'verify'" :label="t('profile.verificationCode')" required>
          <n-input-group>
            <n-input
              v-model:value="emailForm.code"
              :placeholder="t('profile.enterSixDigitCode')"
              :maxlength="6"
            />
            <n-button
              :disabled="emailCodeCountdown > 0"
              @click="triggerSendEmailCode"
            >
              {{ emailCodeCountdown > 0 ? `${emailCodeCountdown}s` : t('profile.resend') }}
            </n-button>
          </n-input-group>
        </n-form-item>
        <GeetestCaptcha
          v-if="isGeetestEnabled"
          ref="emailGeetestRef"
          :key="emailCaptchaKey"
          :config="{ product: 'bind' }"
          @success="onEmailGeetestSuccess"
          @error="onEmailGeetestError"
        />
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showEmailModal = false">
            {{ t('common.cancel') }}
          </n-button>
          <n-button v-if="emailStep === 'input'" type="primary" @click="triggerSendEmailCode">
            {{ t('profile.sendCode') }}
          </n-button>
          <n-button v-else type="primary" @click="handleVerifyEmailChange">
            {{ t('profile.confirmChange') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 修改手机号弹窗（验证码流程） -->
    <n-modal v-model:show="showPhoneModal" preset="dialog" :title="t('profile.changePhone')">
      <n-form :model="phoneForm" label-placement="left" label-width="100px">
        <n-form-item :label="t('profile.newPhone')" required>
          <n-input
            v-model:value="phoneForm.mobile"
            :placeholder="t('profile.enterNewPhone')"
            :disabled="phoneStep === 'verify'"
          />
        </n-form-item>
        <n-form-item v-if="phoneStep === 'verify'" :label="t('profile.verificationCode')" required>
          <n-input-group>
            <n-input
              v-model:value="phoneForm.code"
              :placeholder="t('profile.enterSixDigitCode')"
              :maxlength="6"
            />
            <n-button
              :disabled="phoneCodeCountdown > 0"
              @click="triggerSendPhoneCode"
            >
              {{ phoneCodeCountdown > 0 ? `${phoneCodeCountdown}s` : t('profile.resend') }}
            </n-button>
          </n-input-group>
        </n-form-item>
        <GeetestCaptcha
          v-if="isGeetestEnabled"
          ref="phoneGeetestRef"
          :key="phoneCaptchaKey"
          :config="{ product: 'bind' }"
          @success="onPhoneGeetestSuccess"
          @error="onPhoneGeetestError"
        />
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showPhoneModal = false">
            {{ t('common.cancel') }}
          </n-button>
          <n-button v-if="phoneStep === 'input'" type="primary" @click="triggerSendPhoneCode">
            {{ t('profile.sendCode') }}
          </n-button>
          <n-button v-else type="primary" @click="handleVerifyPhoneChange">
            {{ t('profile.confirmChange') }}
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
