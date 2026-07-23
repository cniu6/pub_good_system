/**
 * 管理端个人设置：修改密码
 */
<script setup lang="ts">
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { adminProfileApi } from '@/service/api/admin/profile'

const { t } = useI18n()
const message = useMessage()

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm: '',
})

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
</script>

<template>
  <n-space vertical :size="16">
    <n-card :title="t('adminProfile.passwordTitle')" :bordered="false">
      <n-form label-placement="left" label-width="120">
        <n-form-item :label="t('adminProfile.oldPassword')">
          <n-input
            v-model:value="pwdForm.old_password"
            type="password"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item :label="t('adminProfile.newPassword')">
          <n-input
            v-model:value="pwdForm.new_password"
            type="password"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item :label="t('adminProfile.confirmPassword')">
          <n-input
            v-model:value="pwdForm.confirm"
            type="password"
            show-password-on="click"
          />
        </n-form-item>
        <n-button type="primary" @click="changePassword">
          {{ t('adminProfile.changePassword') }}
        </n-button>
      </n-form>
    </n-card>
  </n-space>
</template>
