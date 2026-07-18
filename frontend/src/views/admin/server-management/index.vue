<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  normalizeActiveTab,
  useServerManagement,
} from './composables/useServerManagement'
import MonitorTab from './tabs/MonitorTab.vue'
import OpsTab from './tabs/OpsTab.vue'
import DebugTab from './tabs/DebugTab.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { activeTab, initPage, bindLifecycleOnce } = useServerManagement()

bindLifecycleOnce()

watch(
  () => route.query.tab,
  (value) => {
    activeTab.value = normalizeActiveTab(value)
  },
  { immediate: true },
)

watch(activeTab, (value) => {
  const nextTab = normalizeActiveTab(value)
  if (route.query.tab === nextTab)
    return
  router.replace({ query: { ...route.query, tab: nextTab } })
})

onMounted(() => {
  void initPage()
})
</script>

<template>
  <NTabs v-model:value="activeTab" type="line" animated>
    <NTabPane name="monitor" :tab="t('adminSettings.systemMonitor')">
      <MonitorTab />
    </NTabPane>
    <NTabPane name="ops" :tab="t('adminServer.operationsTab')">
      <OpsTab />
    </NTabPane>
    <NTabPane name="debug" :tab="t('adminSettings.debugTools')">
      <DebugTab />
    </NTabPane>
  </NTabs>
</template>
