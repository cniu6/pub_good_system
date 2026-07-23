/**
 * 管理端个人设置：改密 + TOTP 2FA 骨架
 */
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { adminProfileApi } from '@/service/api/admin/profile'

const { t } = useI18n()
const message = useMessage()

const loading = ref(false)
const totpEnabled = ref(false)
const setupSecret = ref('')
const setupUrl = ref('')

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm: '',
})
const totpCode = ref('')
const disableCode = ref('')

async function loadMe() {
  loading.value = true
  try {
    const res = await adminProfileApi.me()
    if (res.isSuccess && res.data)
      totpEnabled.value = !!res.data.totp_enabled
  }
  catch (e) {
    if (import.meta.env.DEV)
      console.error(e)
  }
  finally {
    loading.value = false
  }
}

async function changePassword() {
  if (!pwdForm.old_password || !pwdForm.new_password) {
    message.warning(t('adminProfile.fillPassword'))
    return
  }
  if (pwdForm.new_password !== pwdForm.confirm) {
    message.warning(t('adminProfile.passwordMismatch'))
    return
  }
  const res = await adminProfileApi.changePassword({
    old_password: pwdForm.old_password,
    new_password: pwdForm.new_password,
  })
  if (res.isSuccess) {
    message.success(res.message || t('adminProfile.passwordChanged'))
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm = ''
  }
  else {
    message.error(res.message || t('adminProfile.actionFailed'))
  }
}

async function setupTotp() {
  const res = await adminProfileApi.setupTotp()
  if (res.isSuccess && res.data) {
    setupSecret.value = res.data.secret || ''
    setupUrl.value = res.data.otpauth_url || ''
    message.success(t('adminProfile.totpSetupReady'))
  }
  else {
    message.error(res.message || t('adminProfile.actionFailed'))
  }
}

async function enableTotp() {
  if (!totpCode.value) {
    message.warning(t('adminProfile.fillTotpCode'))
    return
  }
  const res = await adminProfileApi.enableTotp({ code: totpCode.value })
  if (res.isSuccess) {
    message.success(t('adminProfile.totpEnabled'))
    totpEnabled.value = true
    setupSecret.value = ''
    setupUrl.value = ''
    totpCode.value = ''
    const { invalidateSensitiveTotpCache } = await import('@/composables/useSensitiveTotp')
    invalidateSensitiveTotpCache()
  }
  else {
    message.error(res.message || t('adminProfile.actionFailed'))
  }
}

async function disableTotp() {
  if (!disableCode.value) {
    message.warning(t('adminProfile.fillTotpCode'))
    return
  }
  const res = await adminProfileApi.disableTotp({ code: disableCode.value })
  if (res.isSuccess) {
    message.success(t('adminProfile.totpDisabled'))
    totpEnabled.value = false
    disableCode.value = ''
    const { invalidateSensitiveTotpCache } = await import('@/composables/useSensitiveTotp')
    invalidateSensitiveTotpCache()
  }
  else {
    message.error(res.message || t('adminProfile.actionFailed'))
  }
}

onMounted(loadMe)
</script>

<template>
  <n-space vertical :size="16">
    <n-card :title="t('adminProfile.passwordTitle')" :bordered="false">
      <n-form label-placement="left" label-width="120">
        <n-form-item :label="t('adminProfile.oldPassword')">
          <n-input v-model:value="pwdForm.old_password" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item :label="t('adminProfile.newPassword')">
          <n-input v-model:value="pwdForm.new_password" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item :label="t('adminProfile.confirmPassword')">
          <n-input v-model:value="pwdForm.confirm" type="password" show-password-on="click" />
        </n-form-item>
        <n-button type="primary" @click="changePassword">
          {{ t('adminProfile.changePassword') }}
        </n-button>
      </n-form>
    </n-card>

    <n-card :title="t('adminProfile.totpTitle')" :bordered="false" :loading="loading">
      <n-alert v-if="totpEnabled" type="success" class="mb-12px">
        {{ t('adminProfile.totpOn') }}
      </n-alert>
      <n-alert v-else type="info" class="mb-12px">
        {{ t('adminProfile.totpOff') }}
      </n-alert>

      <n-space vertical v-if="!totpEnabled">
        <n-button @click="setupTotp">
          {{ t('adminProfile.totpSetup') }}
        </n-button>
        <div v-if="setupSecret" class="text-13px">
          <div>{{ t('adminProfile.totpSecret') }}: <code>{{ setupSecret }}</code></div>
          <div class="mt-6px break-all opacity-70">{{ setupUrl }}</div>
          <n-input
            v-model:value="totpCode"
            class="mt-12px max-w-240px"
            :placeholder="t('adminProfile.totpCodePlaceholder')"
          />
          <n-button type="primary" class="mt-8px" @click="enableTotp">
            {{ t('adminProfile.totpEnable') }}
          </n-button>
        </div>
      </n-space>

      <n-space v-else vertical>
        <n-input
          v-model:value="disableCode"
          class="max-w-240px"
          :placeholder="t('adminProfile.totpCodePlaceholder')"
        />
        <n-button type="warning" @click="disableTotp">
          {{ t('adminProfile.totpDisable') }}
        </n-button>
      </n-space>
    </n-card>
  </n-space>
</template>
