<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useTabStore } from '@/store'

const { t } = useI18n()
const { modifyTab } = useTabStore()

const { fullPath, query } = useRoute()

modifyTab(fullPath, (target) => {
  target.meta.title = t('demo.multi.detailTitle', { id: query.id })
})
</script>

<template>
  <n-space vertical>
    <n-alert :title="t('demo.multi.publicInfo')" type="warning">
      {{ t('demo.multi.detailDesc') }}
    </n-alert>
    <n-alert :title="t('demo.multi.publicInfo')" type="info">
      {{ t('demo.multi.publicAccessDesc') }}
    </n-alert>

    <n-h2>
      {{ t('demo.multi.detailPageId', { id: query.id }) }}
    </n-h2>
  </n-space>
</template>

<style scoped></style>
