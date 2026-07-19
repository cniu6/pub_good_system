<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  emailForm,
  testEmailTo,
  switchLoading,
  savingEmail,
  testingEmail,
  handleUpdateEmailVerifyEnabled,
  handleUpdateSmtpSSL,
  handleUpdateSmtpProxyEnabled,
  handleSaveEmail,
  handleTestEmail,
} = useAdminSettings()

const proxyTypeOptions = [
  { label: 'SOCKS5', value: 'socks5' },
  { label: 'SOCKS5H', value: 'socks5h' },
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
]

const smtpPort = computed(() => Number(emailForm.smtp_port) || 0)
const showSslOnConflict = computed(() => smtpPort.value === 587 && emailForm.smtp_ssl)
const showSslOffConflict = computed(() => smtpPort.value === 465 && !emailForm.smtp_ssl)

/** 市面常见端口一键填：顺带把 SSL 开关配成对应习惯 */
function applyPortPreset(port: number) {
  emailForm.smtp_port = port
  if (port === 465) {
    if (!emailForm.smtp_ssl)
      handleUpdateSmtpSSL(true)
  }
  else if (port === 587 || port === 25) {
    if (emailForm.smtp_ssl)
      handleUpdateSmtpSSL(false)
  }
}
</script>

<template>
  <n-space vertical :size="16">
    <n-form :model="emailForm" label-placement="left" label-width="120px" style="max-width: 760px;">
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
        <n-space vertical :size="8" style="width: 100%;">
          <n-input v-model:value="emailForm.smtp_host" :placeholder="t('adminSettings.smtpHostPlaceholder')" />
          <n-text depth="3" style="font-size: 12px;">{{ t('adminSettings.smtpHostHint') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpPort')">
        <n-space vertical :size="8" style="width: 100%;">
          <n-space align="center" :wrap="true">
            <n-input-number v-model:value="emailForm.smtp_port" :min="1" :max="65535" style="width: 160px;" />
            <n-button
              size="small"
              :type="smtpPort === 587 ? 'primary' : 'default'"
              secondary
              @click="applyPortPreset(587)"
            >
              587
            </n-button>
            <n-button
              size="small"
              :type="smtpPort === 465 ? 'primary' : 'default'"
              secondary
              @click="applyPortPreset(465)"
            >
              465
            </n-button>
            <n-button
              size="small"
              :type="smtpPort === 25 ? 'primary' : 'default'"
              secondary
              @click="applyPortPreset(25)"
            >
              25
            </n-button>
            <n-text depth="3" style="font-size: 12px;">{{ t('adminSettings.smtpPortPresetHint') }}</n-text>
          </n-space>
          <n-alert type="info" :bordered="false" style="padding: 8px 12px;">
            <div style="white-space: pre-line; line-height: 1.65;">{{ t('adminSettings.smtpPortHint') }}</div>
          </n-alert>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpUsername')">
        <n-space vertical :size="4" style="width: 100%;">
          <n-input v-model:value="emailForm.smtp_username" :placeholder="t('adminSettings.smtpUsernamePlaceholder')" />
          <n-text depth="3" style="font-size: 12px;">{{ t('adminSettings.smtpUsernameHint') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpPassword')">
        <n-space vertical :size="4" style="width: 100%;">
          <n-input
            v-model:value="emailForm.smtp_password"
            type="password"
            show-password-on="click"
            :placeholder="t('adminSettings.smtpPasswordPlaceholder')"
          />
          <n-text depth="3" style="font-size: 12px;">{{ t('adminSettings.smtpPasswordHint') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.systemEmailAddress')">
        <n-space vertical :size="4" style="width: 100%;">
          <n-input
            v-model:value="emailForm.system_email_address"
            :placeholder="t('adminSettings.systemEmailAddressPlaceholder')"
          />
          <n-text depth="3" style="font-size: 12px;">{{ t('adminSettings.systemEmailAddressHint') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.systemEmailName')">
        <n-input v-model:value="emailForm.system_email_name" :placeholder="t('adminSettings.systemEmailNamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.smtpSSL')">
        <n-space vertical :size="8" style="width: 100%;">
          <n-space align="center">
            <n-switch :value="emailForm.smtp_ssl" :loading="switchLoading.smtp_ssl" @update:value="handleUpdateSmtpSSL" />
            <n-text depth="3">{{ emailForm.smtp_ssl ? t('adminSettings.sslEnabled') : t('adminSettings.sslDisabled') }}</n-text>
          </n-space>
          <n-alert type="info" :bordered="false" style="padding: 8px 12px;">
            <div style="white-space: pre-line; line-height: 1.65;">{{ t('adminSettings.smtpSSLHint') }}</div>
          </n-alert>
          <n-alert v-if="showSslOnConflict" type="warning" :bordered="false" style="padding: 8px 12px;">
            {{ t('adminSettings.smtpSSLPortConflict587') }}
          </n-alert>
          <n-alert v-if="showSslOffConflict" type="warning" :bordered="false" style="padding: 8px 12px;">
            {{ t('adminSettings.smtpSSLPortConflict465') }}
          </n-alert>
        </n-space>
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.smtpProxy')">
        <n-space vertical :size="4" style="width: 100%;">
          <n-space align="center">
            <n-switch
              :value="emailForm.smtp_proxy_enabled"
              :loading="switchLoading.smtp_proxy_enabled"
              @update:value="handleUpdateSmtpProxyEnabled"
            />
            <n-text depth="3">{{ emailForm.smtp_proxy_enabled ? t('adminSettings.smtpProxyOn') : t('adminSettings.smtpProxyOff') }}</n-text>
          </n-space>
          <n-text depth="3" style="font-size: 12px;">{{ t('adminSettings.smtpProxyHint') }}</n-text>
        </n-space>
      </n-form-item>
      <template v-if="emailForm.smtp_proxy_enabled">
        <n-form-item :label="t('adminSettings.smtpProxyType')">
          <n-select v-model:value="emailForm.smtp_proxy_type" :options="proxyTypeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.smtpProxyHost')">
          <n-input v-model:value="emailForm.smtp_proxy_host" :placeholder="t('adminSettings.smtpProxyHostPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.smtpProxyPort')">
          <n-input-number v-model:value="emailForm.smtp_proxy_port" :min="1" :max="65535" style="width: 100%;" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.smtpProxyUsername')">
          <n-input v-model:value="emailForm.smtp_proxy_username" :placeholder="t('adminSettings.smtpProxyUsernamePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.smtpProxyPassword')">
          <n-input
            v-model:value="emailForm.smtp_proxy_password"
            type="password"
            show-password-on="click"
            :placeholder="t('adminSettings.smtpProxyPasswordPlaceholder')"
          />
        </n-form-item>
      </template>
      <n-divider />
      <n-form-item :label="t('adminSettings.testEmailTo')">
        <n-input
          v-model:value="testEmailTo"
          :placeholder="t('adminSettings.testEmailToPlaceholder')"
          clearable
        />
      </n-form-item>
      <n-form-item>
        <n-space>
          <n-button type="primary" :loading="savingEmail" @click="() => handleSaveEmail()">{{ t('adminSettings.save') }}</n-button>
          <n-button
            :loading="testingEmail"
            :disabled="!String(testEmailTo || '').trim()"
            @click="handleTestEmail"
          >
            {{ t('adminSettings.sendTestEmail') }}
          </n-button>
        </n-space>
      </n-form-item>
    </n-form>

    <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
      {{ t('adminSettings.emailTestHint') }}
    </n-alert>
    <!-- 快速对照放底部，不顶占首屏 -->
    <n-alert type="info" :title="t('adminSettings.smtpGuideTitle')">
      <div style="font-size: 13px; line-height: 1.75; white-space: pre-line;">{{ t('adminSettings.smtpGuideBody') }}</div>
    </n-alert>
  </n-space>
</template>
