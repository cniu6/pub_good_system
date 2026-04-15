<script lang="ts">
</script>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AMap from './components/AMap.vue'
import BMap from './components/BMap.vue'

const { t } = useI18n()

defineOptions({
  name: 'map',
})

const maps = [
  {
    id: 'BMap',
    label: t('demo.map.bmap'),
    component: BMap,
  },
  {
    id: 'AMap',
    label: t('demo.map.amap'),
    component: AMap,
  },
]
</script>

<template>
  <n-card :title="t('demo.map.title')">
    <n-tabs
      type="line"
      animated
    >
      <n-tab-pane
        v-for="item in maps"
        :key="item.id"
        :name="item.id"
        :tab="item.label"
        class="h-600px"
      >
        <component :is="item.component" />
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>

<style scoped></style>
