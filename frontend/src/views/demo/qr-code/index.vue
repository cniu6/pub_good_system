<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const text = ref('nova-admin')
</script>

<template>
  <n-card :title="t('demo.qrcode.title')">
    <n-alert :show-icon="false" type="info">
      {{ t('demo.qrcode.desc') }}
    </n-alert>
    <n-qr-code :value="text" />
    <n-input v-model:value="text" :maxlength="60" type="text" />
  </n-card>
</template>

<style scoped></style>
