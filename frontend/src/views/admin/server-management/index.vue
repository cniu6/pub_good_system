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
import DatabaseTab from './tabs/DatabaseTab.vue'
import TerminalTab from './tabs/TerminalTab.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { activeTab, handleActiveTabChanged, initPage, bindLifecycleOnce } = useServerManagement()

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
  handleActiveTabChanged(nextTab)
  if (route.query.tab === nextTab)
    return
  router.replace({ query: { ...route.query, tab: nextTab } })
})

onMounted(() => {
  void initPage()
  // 首次 route.query.tab 的 immediate watch 发生在本 watch 注册之前，需主动处理一次。
  handleActiveTabChanged(activeTab.value)
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
    <NTabPane name="database" :tab="t('adminServer.dbTab')">
      <DatabaseTab />
    </NTabPane>
    <NTabPane name="terminal" :tab="t('adminServer.terminalTab')">
      <TerminalTab />
    </NTabPane>
  </NTabs>
</template>
