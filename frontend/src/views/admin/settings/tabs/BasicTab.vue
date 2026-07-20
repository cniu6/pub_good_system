<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  basicForm,
  langOptions,
  switchLoading,
  savingBasic,
  handleUpdateAllowRegister,
  handleUpdateAnnouncementEnabled,
  handleSaveBasic,
} = useAdminSettings()
</script>

<template>
  <n-space vertical>
    <n-form :model="basicForm" label-placement="left" label-width="120px" style="max-width: 640px;">
      <n-form-item :label="t('adminSettings.siteName')">
        <n-input v-model:value="basicForm.site_name" :placeholder="t('adminSettings.siteNamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.siteDesc')">
        <n-input v-model:value="basicForm.site_desc" type="textarea" :placeholder="t('adminSettings.siteDescPlaceholder')" :rows="3" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.siteLogo')">
        <n-input v-model:value="basicForm.site_logo" :placeholder="t('adminSettings.siteLogoPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.copyright')">
        <n-input v-model:value="basicForm.copyright" :placeholder="t('adminSettings.copyrightPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.icp')">
        <n-input v-model:value="basicForm.icp" :placeholder="t('adminSettings.icpPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.version')">
        <n-input v-model:value="basicForm.version" :placeholder="t('adminSettings.versionPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.defaultLang')">
        <n-select v-model:value="basicForm.default_lang" :options="langOptions" :placeholder="t('adminSettings.defaultLangPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.frontendUrl')">
        <n-input v-model:value="basicForm.frontend_url" :placeholder="t('adminSettings.frontendUrlPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.backendApiUrl')">
        <n-input v-model:value="basicForm.backend_api_url" :placeholder="t('adminSettings.backendApiUrlPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.userRegistration')">
        <n-space align="center">
          <n-switch
            :value="basicForm.allow_register"
            :loading="switchLoading.allow_register"
            @update:value="handleUpdateAllowRegister"
          />
          <n-text depth="3">{{ basicForm.allow_register ? t('adminSettings.allowRegister') : t('adminSettings.disallowRegister') }}</n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.announcementEnabled')">
        <n-switch
          :value="basicForm.announcement_enabled"
          :loading="switchLoading.announcement_enabled"
          @update:value="handleUpdateAnnouncementEnabled"
        />
      </n-form-item>
      <n-form-item>
        <n-button type="primary" :loading="savingBasic" @click="handleSaveBasic">{{ t('adminSettings.saveSettings') }}</n-button>
      </n-form-item>
    </n-form>
  </n-space>
</template>
