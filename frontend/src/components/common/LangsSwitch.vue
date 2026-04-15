<script setup lang="ts">
import { useAppStore, useAuthStore } from '@/store'
import { updateUserSettings } from '@/service'
import { langToBackendFormat } from '@/utils'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const options = [
  {
    label: t('langs.english'),
    value: 'enUS',
  },
  {
    label: t('langs.chinese'),
    value: 'zhCN',
  },
]

function handleLangChange(lang: App.lang) {
  appStore.setAppLang(lang)
  if (authStore.isLogin) {
    updateUserSettings({ language: langToBackendFormat(lang) }).catch(() => {})
  }
}
</script>

<template>
  <n-popselect :value="appStore.lang" :options="options" trigger="click" @update:value="handleLangChange">
    <CommonWrapper>
      <icon-park-outline-translate />
    </CommonWrapper>
  </n-popselect>
</template>

<style scoped></style>
