<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import EmailTemplates from '@/views/admin/email-templates/index.vue'
import { useAdminSettings } from './composables/useAdminSettings'
import BasicTab from './tabs/BasicTab.vue'
import EmailTab from './tabs/EmailTab.vue'
import SmsTab from './tabs/SmsTab.vue'
import SecurityTab from './tabs/SecurityTab.vue'
import RealnameApiTab from './tabs/RealnameApiTab.vue'
import PaymentTab from './tabs/PaymentTab.vue'
import CustomTab from './tabs/CustomTab.vue'

const { t } = useI18n()
const { loading, topTab, systemSubTab, loadSettings } = useAdminSettings()

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <n-card :title="t('adminSettings.title')" :bordered="false">
    <n-spin :show="loading">
      <n-tabs v-model:value="topTab" type="line" animated>
        <n-tab-pane name="system-config" :tab="t('adminSettings.systemConfig')">
          <n-tabs v-model:value="systemSubTab" type="line" placement="left" animated>
            <n-tab-pane name="basic" :tab="t('adminSettings.basicSettings')">
              <BasicTab />
            </n-tab-pane>
            <n-tab-pane name="email" :tab="t('adminSettings.emailSettings')">
              <EmailTab />
            </n-tab-pane>
            <n-tab-pane name="sms" :tab="t('adminSettings.smsSettings')">
              <SmsTab />
            </n-tab-pane>
            <n-tab-pane name="security" :tab="t('adminSettings.securitySettings')">
              <SecurityTab />
            </n-tab-pane>
            <n-tab-pane name="realname-api" :tab="t('adminSettings.realnameApiSettings')">
              <RealnameApiTab />
            </n-tab-pane>
            <n-tab-pane name="payment" :tab="t('adminSettings.paymentSettings')">
              <PaymentTab />
            </n-tab-pane>
            <n-tab-pane name="custom" :tab="t('adminSettings.customConfig')">
              <CustomTab />
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>

        <n-tab-pane name="email-templates" :tab="t('adminSettings.emailTemplates')">
          <EmailTemplates />
        </n-tab-pane>
      </n-tabs>
    </n-spin>
  </n-card>
</template>
