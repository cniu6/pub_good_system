<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  fetchFormPost,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function post() {
  const params = {
    data: '2022-2-2',
    name: 'for',
  }
  const res = await fetchFormPost(params)
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.formPost')" size="small">
    <n-button @click="post">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
