<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  securityForm,
  switchLoading,
  savingSecurity,
  restartingBackend,
  handleUpdateGeetestEnabled,
  handleUpdateAllowDeleteAccount,
  handleUpdateRealnameEnabled,
  handleUpdateRealnameReviewRequired,
  handleSaveSecurity,
  handleRestartBackend,
} = useAdminSettings()
</script>

<template>
  <n-space vertical>
    <n-form :model="securityForm" label-placement="left" label-width="180px" style="max-width: 640px;">
      <n-form-item :label="t('adminSettings.geetestEnabled')">
        <n-space align="center">
          <n-switch
            :value="securityForm.geetest_enabled"
            :loading="switchLoading.geetest_enabled"
            @update:value="handleUpdateGeetestEnabled"
          />
          <n-text depth="3">{{ securityForm.geetest_enabled ? t('adminSettings.enabled') : t('adminSettings.disabled') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.allowDeleteAccount')">
        <n-space align="center">
          <n-switch
            :value="securityForm.allow_delete_account"
            :loading="switchLoading.allow_delete_account"
            @update:value="handleUpdateAllowDeleteAccount"
          />
          <n-text depth="3">{{ securityForm.allow_delete_account ? t('adminSettings.allowDeleteAccountEnabled') : t('adminSettings.allowDeleteAccountDisabled') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.geetestCaptchaId')">
        <n-input
          v-model:value="securityForm.geetest_captcha_id"
          type="password"
          show-password-on="click"
          :placeholder="t('adminSettings.geetestCaptchaIdPlaceholder')"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.geetestCaptchaKey')">
        <n-input
          v-model:value="securityForm.geetest_captcha_key"
          type="password"
          show-password-on="click"
          :placeholder="t('adminSettings.geetestCaptchaKeyPlaceholder')"
        />
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.jwtAccessExpire')">
        <n-input-number v-model:value="securityForm.jwt_access_expire" :min="300" :step="300" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.jwtRefreshExpire')">
        <n-input-number v-model:value="securityForm.jwt_refresh_expire" :min="3600" :step="3600" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.loginMaxFailure')">
        <n-input-number v-model:value="securityForm.login_max_failure" :min="3" :max="20" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.loginLockDuration')">
        <n-input-number v-model:value="securityForm.login_lock_duration" :min="1" :max="1440" style="width: 100%;" />
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.realnameEnabled')">
        <n-space align="center">
          <n-switch
            :value="securityForm.realname_enabled"
            :loading="switchLoading.realname_enabled"
            @update:value="handleUpdateRealnameEnabled"
          />
          <n-text depth="3">{{ securityForm.realname_enabled ? t('adminSettings.realnameEnabledText') : t('adminSettings.realnameDisabledText') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.realnameReview')">
        <n-space align="center">
          <n-switch
            :value="securityForm.realname_review_required"
            :loading="switchLoading.realname_review_required"
            @update:value="handleUpdateRealnameReviewRequired"
          />
          <n-text depth="3">{{ securityForm.realname_review_required ? t('adminSettings.realnameReviewRequired') : t('adminSettings.realnameReviewNotRequired') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.realnameNotifyText')">
        <n-input
          v-model:value="securityForm.realname_notify_text"
          type="textarea"
          :placeholder="t('adminSettings.realnameNotifyTextPlaceholder')"
          :rows="3"
        />
      </n-form-item>
      <n-form-item>
        <n-space>
          <n-button type="primary" :loading="savingSecurity" @click="handleSaveSecurity">{{ t('adminSettings.saveSettings') }}</n-button>
          <n-button type="warning" :loading="restartingBackend" @click="handleRestartBackend">{{ t('adminSettings.restartBackend') }}</n-button>
        </n-space>
      </n-form-item>
    </n-form>
  </n-space>
</template>
