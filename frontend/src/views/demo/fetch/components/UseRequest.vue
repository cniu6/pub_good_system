<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  fetchGet,
} from '@/service'

import { useRequest } from 'alova/client'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

const { data: fetchGetData, send: sendFetchGet } = useRequest(fetchGet({ a: 112211 }), {
  immediate: false,
})

async function handleRequestHook() {
  await sendFetchGet()
  emit('update', fetchGetData.value)
}
</script>

<template>
  <n-card :title="t('demo.fetch.useRequestStyle')" size="small">
    <n-button @click="handleRequestHook">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
