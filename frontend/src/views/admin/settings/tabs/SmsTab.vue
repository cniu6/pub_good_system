<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  smsForm,
  switchLoading,
  savingSms,
  smsProviderOptions,
  smsBodyFormatOptions,
  smsProviderNeedsSignName,
  smsProviderNeedsTemplateCode,
  smsAccessKeyPlaceholder,
  smsSecretKeyPlaceholder,
  smsTemplateLabel,
  smsTemplatePlaceholder,
  smsTemplateEnLabel,
  smsTemplateEnPlaceholder,
  handleUpdateSmsVerifyEnabled,
  handleSaveSms,
} = useAdminSettings()
</script>

<template>
  <n-space vertical>
    <n-form :model="smsForm" label-placement="left" label-width="120px" style="max-width: 640px;">
      <n-form-item :label="t('adminSettings.smsVerification')">
        <n-space align="center">
          <n-switch
            :value="smsForm.sms_verify_enabled"
            :loading="switchLoading.sms_verify_enabled"
            @update:value="handleUpdateSmsVerifyEnabled"
          />
          <n-text depth="3">{{ smsForm.sms_verify_enabled ? t('adminSettings.smsVerifyEnabled') : t('adminSettings.smsVerifyDisabled') }}</n-text>
        </n-space>
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.smsProvider')">
        <n-select
          v-model:value="smsForm.sms_provider"
          :options="smsProviderOptions"
          :placeholder="t('adminSettings.smsProviderPlaceholder')"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smsAccessKey')">
        <n-input
          v-model:value="smsForm.sms_access_key"
          type="password"
          show-password-on="click"
          :placeholder="smsAccessKeyPlaceholder"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smsSecretKey')">
        <n-input
          v-model:value="smsForm.sms_secret_key"
          type="password"
          show-password-on="click"
          :placeholder="smsSecretKeyPlaceholder"
        />
      </n-form-item>
      <n-form-item v-if="smsProviderNeedsSignName" :label="t('adminSettings.smsSignName')">
        <n-input v-model:value="smsForm.sms_sign_name" :placeholder="t('adminSettings.smsSignNamePlaceholder')" />
      </n-form-item>
      <n-form-item v-if="smsProviderNeedsTemplateCode" :label="smsTemplateLabel">
        <n-input v-model:value="smsForm.sms_template_code" :placeholder="smsTemplatePlaceholder" />
      </n-form-item>
      <n-form-item v-if="smsProviderNeedsTemplateCode" :label="smsTemplateEnLabel">
        <n-input v-model:value="smsForm.sms_template_code_en" :placeholder="smsTemplateEnPlaceholder" />
      </n-form-item>
      <n-form-item v-if="smsForm.sms_provider === 'aliyun'" :label="t('adminSettings.smsRegion')">
        <n-input v-model:value="smsForm.sms_region" :placeholder="t('adminSettings.smsRegionPlaceholder')" />
      </n-form-item>
      <n-form-item v-if="smsForm.sms_provider === 'tencent'" :label="t('adminSettings.smsSdkAppId')">
        <n-input v-model:value="smsForm.sms_sdk_app_id" :placeholder="t('adminSettings.smsSdkAppIdPlaceholder')" />
      </n-form-item>
      <n-form-item v-if="smsForm.sms_provider === 'custom'" :label="t('adminSettings.smsEndpoint')">
        <n-input v-model:value="smsForm.sms_endpoint" :placeholder="t('adminSettings.smsEndpointPlaceholder')" />
      </n-form-item>
      <n-form-item v-if="smsForm.sms_provider === 'custom'" :label="t('adminSettings.smsBodyFormat')">
        <n-select
          v-model:value="smsForm.sms_body_format"
          :options="smsBodyFormatOptions"
          :placeholder="t('adminSettings.smsBodyFormatPlaceholder')"
        />
      </n-form-item>
      <n-form-item>
        <n-button type="primary" :loading="savingSms" @click="handleSaveSms">{{ t('adminSettings.saveSettings') }}</n-button>
      </n-form-item>
    </n-form>
    <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
      {{ t('adminSettings.smsAlert') }}
    </n-alert>
  </n-space>
</template>
