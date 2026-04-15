<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <n-card class="h-130vh">
    {{ t('demo.multi.multi2Desc') }}
    <n-button @click="router.push({ path: '/multi/multi2/detail', query: { id: 1 } })">
      {{ t('demo.multi.gotoDetail1') }}
    </n-button>
    <n-button @click="router.push({ path: '/multi/multi2/detail', query: { id: 2 } })">
      {{ t('demo.multi.gotoDetail2') }}
    </n-button>
    <n-button @click="router.push({ path: '/multi/multi2/detail', query: { id: 3 } })">
      {{ t('demo.multi.gotoDetail3') }}
    </n-button>
  </n-card>
</template>

<style scoped></style>
