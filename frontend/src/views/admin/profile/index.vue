/**
 * 管理端个人设置：修改密码
 */
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import type { AdminMeInfo } from '@/service/api/admin/profile'
import { adminProfileApi } from '@/service/api/admin/profile'

const { t } = useI18n()
const message = useMessage()

const me = ref<AdminMeInfo | null>(null)
const meLoading = ref(false)

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm: '',
})

async function loadMe() {
  meLoading.value = true
  try {
    const res = await adminProfileApi.me()
    if (res.isSuccess && res.data) {
      me.value = res.data
      return
    }
    message.error(res.message || t('adminProfile.loadFailed'))
  }
  catch {
    message.error(t('adminProfile.loadFailed'))
  }
  finally {
    meLoading.value = false
  }
}

onMounted(loadMe)

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
    <n-card :title="t('adminProfile.accountTitle')" :bordered="false">
      <n-spin :show="meLoading">
        <n-descriptions :column="2" label-placement="left" bordered size="small">
          <n-descriptions-item :label="t('adminProfile.username')">
            {{ me?.username || '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminProfile.userId')">
            {{ me?.id ?? '-' }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminProfile.role')">
            {{ me?.role === 'admin' ? t('adminUsers.admin') : (me?.role || '-') }}
          </n-descriptions-item>
          <n-descriptions-item :label="t('adminProfile.sessionHint')">
            {{ t('adminProfile.sessionHintText') }}
          </n-descriptions-item>
        </n-descriptions>
      </n-spin>
    </n-card>
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
