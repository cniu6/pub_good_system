<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

async function get() {
  const res = import.meta.env
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.checkEnv')" size="small">
    <n-button @click="get">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
