<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  withoutToken,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function withoutTokenRequest() {
  const res = await withoutToken()
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.noToken')" size="small">
    <n-button @click="withoutTokenRequest">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
