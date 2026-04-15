<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

async function getDictData() {
  const res = { data: t('demo.fetch.sampleData') }
  emit('update', res)
}
</script>

<template>
  <n-card :title="t('demo.fetch.transformData')" size="small">
    <n-button @click="getDictData">
      {{ t('demo.fetch.click') }}
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
