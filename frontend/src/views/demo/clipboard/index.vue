<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const text = ref('Hello nova-admin')
</script>

<template>
  <n-card :title="t('demo.clipboard.title')">
    <n-h3>{{ t('demo.clipboard.vCopyDirective') }}</n-h3>
    <n-input-group>
      <n-input v-model:value="text" :placeholder="t('demo.clipboard.placeholder')" />
      <n-button v-copy="text" type="primary">
        {{ t('demo.clipboard.vCopyBtn') }}
      </n-button>
    </n-input-group>
    <n-h3>{{ t('demo.clipboard.copyTextComponent') }}</n-h3>
    <copy-text v-model:value="text" />
  </n-card>
</template>

<style scoped></style>
