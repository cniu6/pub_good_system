<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  fetchPost,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function post() {
  const params = {
    data: '2022-2-2',
    data1: [],
    data2: {},
    data3: '',
    data4: null,
    data5: undefined,
  }
  const res = await fetchPost(params)
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.postMethod')" size="small">
    <n-button @click="post">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
