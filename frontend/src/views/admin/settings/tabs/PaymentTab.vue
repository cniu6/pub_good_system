<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import CurrencyTab from './CurrencyTab.vue'
import PaymentBasicTab from './PaymentBasicTab.vue'

const { t } = useI18n()

const innerTab = ref('basic')
</script>

<template>
  <n-tabs v-model:value="innerTab" type="line" animated>
    <n-tab-pane name="basic" :tab="t('adminSettings.paymentBasicSettings')">
      <PaymentBasicTab />
    </n-tab-pane>
    <n-tab-pane name="currency" :tab="t('adminSettings.currencySettings')">
      <CurrencyTab />
    </n-tab-pane>
  </n-tabs>
</template>
