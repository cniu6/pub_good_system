<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  emailForm,
  switchLoading,
  savingEmail,
  testingEmail,
  handleUpdateEmailVerifyEnabled,
  handleUpdateSmtpSSL,
  handleSaveEmail,
  handleTestEmail,
} = useAdminSettings()
</script>

<template>
  <n-space vertical>
    <n-form :model="emailForm" label-placement="left" label-width="120px" style="max-width: 640px;">
      <n-form-item :label="t('adminSettings.emailVerification')">
        <n-space align="center">
          <n-switch
            :value="emailForm.email_verify_enabled"
            :loading="switchLoading.email_verify_enabled"
            @update:value="handleUpdateEmailVerifyEnabled"
          />
          <n-text depth="3">{{ emailForm.email_verify_enabled ? t('adminSettings.emailVerifyEnabled') : t('adminSettings.emailVerifyDisabled') }}</n-text>
        </n-space>
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.smtpHost')">
        <n-input v-model:value="emailForm.smtp_host" :placeholder="t('adminSettings.smtpHostPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpPort')">
        <n-input-number v-model:value="emailForm.smtp_port" :min="1" :max="65535" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpUsername')">
        <n-input v-model:value="emailForm.smtp_username" :placeholder="t('adminSettings.smtpUsernamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpPassword')">
        <n-input
          v-model:value="emailForm.smtp_password"
          type="password"
          show-password-on="click"
          :placeholder="t('adminSettings.smtpPasswordPlaceholder')"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.systemEmailName')">
        <n-input v-model:value="emailForm.system_email_name" :placeholder="t('adminSettings.systemEmailNamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpSSL')">
        <n-space align="center">
          <n-switch :value="emailForm.smtp_ssl" :loading="switchLoading.smtp_ssl" @update:value="handleUpdateSmtpSSL" />
          <n-text depth="3">{{ emailForm.smtp_ssl ? t('adminSettings.sslEnabled') : t('adminSettings.sslDisabled') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item>
        <n-space>
          <n-button type="primary" :loading="savingEmail" @click="handleSaveEmail">{{ t('adminSettings.save') }}</n-button>
          <n-button :loading="testingEmail" @click="handleTestEmail">{{ t('adminSettings.sendTestEmail') }}</n-button>
        </n-space>
      </n-form-item>
    </n-form>
  </n-space>
</template>
