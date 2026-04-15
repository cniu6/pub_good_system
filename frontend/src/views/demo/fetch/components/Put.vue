<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  fetchPut,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function put() {
  const res = await fetchPut({ a: 112211, b: false })
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.putMethod')" size="small">
    <n-button @click="put">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
