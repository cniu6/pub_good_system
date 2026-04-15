<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  fetchUpdateToken,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function updataToken() {
  const res = await fetchUpdateToken({ refreshToken: 'test token' })
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.refreshToken')" size="small">
    <n-button @click="updataToken">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
