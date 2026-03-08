<script setup lang="ts">
import { useAppStore, useAuthStore } from '@/store'
import { updateUserSettings } from '@/service'
import { langToBackendFormat } from '@/utils'

const appStore = useAppStore()
const authStore = useAuthStore()
const options = [
  {
    label: 'English',
    value: 'enUS',
  },
  {
    label: '中文',
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
