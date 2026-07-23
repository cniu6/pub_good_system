<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import EmailTemplates from '@/views/admin/email-templates/index.vue'
import SmsTemplates from '@/views/admin/sms-templates/index.vue'
import { useAdminSettings } from './composables/useAdminSettings'
import BasicTab from './tabs/BasicTab.vue'
import EmailTab from './tabs/EmailTab.vue'
import SmsTab from './tabs/SmsTab.vue'
import SecurityTab from './tabs/SecurityTab.vue'
import RealnameApiTab from './tabs/RealnameApiTab.vue'
import PaymentTab from './tabs/PaymentTab.vue'
import UserLevelsTab from './tabs/UserLevelsTab.vue'
import CustomTab from './tabs/CustomTab.vue'

const { t } = useI18n()
const { loading, systemSubTab, loadSettings } = useAdminSettings()

/** 邮件 / 短信侧栏内的二级顶栏：设置 | 模板 */
const emailInnerTab = ref('settings')
const smsInnerTab = ref('settings')

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <n-card :title="t('adminSettings.title')" :bordered="false">
    <n-spin :show="loading">
      <!-- 左侧：各配置分类；邮件/短信点进去后再出顶栏切「设置 / 模板」 -->
      <n-tabs v-model:value="systemSubTab" type="line" placement="left" animated>
        <n-tab-pane name="basic" :tab="t('adminSettings.basicSettings')">
          <BasicTab />
        </n-tab-pane>

        <n-tab-pane name="email" :tab="t('adminSettings.emailSettings')">
          <n-tabs v-model:value="emailInnerTab" type="line" animated>
            <n-tab-pane name="settings" :tab="t('adminSettings.emailSmtpSettings')">
              <EmailTab />
            </n-tab-pane>
            <n-tab-pane name="templates" :tab="t('adminSettings.emailTemplates')">
              <EmailTemplates />
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>

        <n-tab-pane name="sms" :tab="t('adminSettings.smsSettings')">
          <n-tabs v-model:value="smsInnerTab" type="line" animated>
            <n-tab-pane name="settings" :tab="t('adminSettings.smsProviderSettings')">
              <SmsTab />
            </n-tab-pane>
            <n-tab-pane name="templates" :tab="t('adminSettings.smsTemplates')">
              <SmsTemplates />
            </n-tab-pane>
          </n-tabs>
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
        <n-tab-pane name="user-levels" :tab="t('adminSettings.userLevels')">
          <UserLevelsTab />
        </n-tab-pane>
        <n-tab-pane name="custom" :tab="t('adminSettings.customConfig')">
          <CustomTab />
        </n-tab-pane>
      </n-tabs>
    </n-spin>
  </n-card>
</template>
