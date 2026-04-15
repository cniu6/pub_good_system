<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  FailedRequest,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

async function failedRequest() {
  const res = await FailedRequest()
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.serverError')" size="small">
    <n-button type="error" @click="failedRequest">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
