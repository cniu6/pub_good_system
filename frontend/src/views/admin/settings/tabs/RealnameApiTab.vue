<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  realnameApiForm,
  switchLoading,
  savingRealnameApi,
  realnameApiProviderOptions,
  handleUpdateRealnameApiEnabled,
  handleSaveRealnameApi,
} = useAdminSettings()
</script>

<template>
  <n-space vertical>
    <n-form :model="realnameApiForm" label-placement="left" label-width="180px" style="max-width: 640px;">
      <n-form-item :label="t('adminSettings.realnameApiEnabled')">
        <n-space align="center">
          <n-switch
            :value="realnameApiForm.realname_api_enabled"
            :loading="switchLoading.realname_api_enabled"
            @update:value="handleUpdateRealnameApiEnabled"
          />
          <n-text depth="3">{{ realnameApiForm.realname_api_enabled ? t('adminSettings.realnameApiEnabledText') : t('adminSettings.realnameApiDisabledText') }}</n-text>
        </n-space>
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.realnameApiProvider')">
        <n-select
          v-model:value="realnameApiForm.realname_api_provider"
          :options="realnameApiProviderOptions"
          :placeholder="t('adminSettings.realnameApiProviderPlaceholder')"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.realnameApiAppKey')">
        <n-input
          v-model:value="realnameApiForm.realname_api_app_key"
          :placeholder="t('adminSettings.realnameApiAppKeyPlaceholder')"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.realnameApiAppSecret')">
        <n-input
          v-model:value="realnameApiForm.realname_api_app_secret"
          type="password"
          show-password-on="click"
          :placeholder="t('adminSettings.realnameApiAppSecretPlaceholder')"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.realnameApiEndpoint')">
        <n-input
          v-model:value="realnameApiForm.realname_api_endpoint"
          :placeholder="t('adminSettings.realnameApiEndpointPlaceholder')"
        />
      </n-form-item>
      <n-form-item>
        <n-button type="primary" :loading="savingRealnameApi" @click="handleSaveRealnameApi">{{ t('adminSettings.saveSettings') }}</n-button>
      </n-form-item>
    </n-form>
    <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
      <p>{{ t('adminSettings.realnameApiAlert1') }}</p>
      <p>{{ t('adminSettings.realnameApiAlert2') }}</p>
    </n-alert>
  </n-space>
</template>
