<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  fetchGet,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function pinterEnv() {
  const res = await fetchGet({ a: 112211, b: false })
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.getMethod')" size="small">
    <n-button @click="pinterEnv">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
