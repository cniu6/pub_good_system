<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

function testMsg() {
  window.$message.error(t('demo.multi.testMessage'))
}
</script>

<template>
  <div
    text-center
    c-red
  >
    {{ t('demo.multi.thirdLevelPage') }}
    <n-button
      strong
      secondary
      type="success"
      @click="testMsg"
    >
      {{ t('demo.multi.testMessage') }}
    </n-button>
  </div>
</template>

<style scoped></style>
