<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  expiredTokenRequest,
} from '@/service'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

async function expiredToken() {
  const res = await expiredTokenRequest()
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.tokenExpiration')" size="small">
    {{ t('demo.fetch.tokenRefreshNotice') }}
    <n-button type="error" @click="expiredToken">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
