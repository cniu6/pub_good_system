<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const text = ref('# Hello Editor ![' + t('demo.editor.imageDesc') + '](https://via.placeholder.com/350x150)')
</script>

<template>
  <n-card :title="t('demo.editor.markdownTitle')">
    <n-space vertical :size="12">
      <n-alert :title="t('demo.editor.markdownAlert')" type="success" />
      <MarkDownEditor v-model="text" />
    </n-space>
  </n-card>
</template>

<style scoped></style>
