<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import OperationLogs from '@/views/admin/logs/index.vue'
import SMSLogs from '@/views/admin/sms-logs/index.vue'
import EmailLogs from '@/views/admin/email-logs/index.vue'
import APILogs from '@/views/admin/api-logs/index.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const tabOptions = ['operation', 'sms', 'email', 'api'] as const
const activeTab = ref('operation')

function normalizeActiveTab(value: unknown) {
  return typeof value === 'string' && tabOptions.includes(value as any) ? value : 'operation'
}

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
</script>

<template>
  <n-tabs v-model:value="activeTab" type="line" animated>
    <n-tab-pane name="operation" :tab="t('adminLogManagement.operationLogs')">
      <OperationLogs />
    </n-tab-pane>
    <n-tab-pane name="sms" :tab="t('adminLogManagement.smsLogs')">
      <SMSLogs />
    </n-tab-pane>
    <n-tab-pane name="email" :tab="t('adminLogManagement.emailLogs')">
      <EmailLogs />
    </n-tab-pane>
    <n-tab-pane name="api" :tab="t('adminLogManagement.apiLogs')">
      <APILogs />
    </n-tab-pane>
  </n-tabs>
</template>
