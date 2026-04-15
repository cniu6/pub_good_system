<script setup lang="ts">
import {
  fetchDelete,
} from '@/service'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function handleDelete() {
  const res = await fetchDelete()
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.delete')" size="small">
    <n-button @click="handleDelete">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
