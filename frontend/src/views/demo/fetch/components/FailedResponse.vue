<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  FailedResponse,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

async function failedResponse() {
  const res = await FailedResponse()
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.bizError')" size="small">
    <n-button type="error" @click="failedResponse">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
