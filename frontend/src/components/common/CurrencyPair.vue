<script setup lang="ts">
/**
 * 货币对展示组件：国旗 + 货币代码
 * 国旗使用 flag-icons（本地 SVG，仅打包需要的国家）
 */

defineProps<{
  from: string
  to: string
}>()

const currencyToCountry: Record<string, string> = {
  CNY: 'cn',
  USD: 'us',
  EUR: 'eu',
  JPY: 'jp',
  GBP: 'gb',
  HKD: 'hk',
  TWD: 'tw',
  KRW: 'kr',
  AUD: 'au',
  CAD: 'ca',
  CHF: 'ch',
  SGD: 'sg',
  INR: 'in',
  RUB: 'ru',
  THB: 'th',
  VND: 'vn',
  MYR: 'my',
  PHP: 'ph',
  IDR: 'id',
  MXN: 'mx',
  BRL: 'br',
  NZD: 'nz',
  ZAR: 'za',
  SEK: 'se',
  NOK: 'no',
  DKK: 'dk',
  PLN: 'pl',
  TRY: 'tr',
  AED: 'ae',
  SAR: 'sa',
  ILS: 'il',
  EGP: 'eg',
}

function getCountryCode(currency: string): string {
  return currencyToCountry[currency.toUpperCase()] || ''
}
</script>

<template>
  <span class="inline-flex items-center gap-1.5 whitespace-nowrap">
    <span
      v-if="getCountryCode(from)"
      class="fi"
      :class="`fi-${getCountryCode(from)}`"
      style="width: 20px; height: 15px; border-radius: 2px;"
    />
    <span class="text-sm">{{ from.toUpperCase() }}</span>
    <span class="text-xs opacity-60">-></span>
    <span
      v-if="getCountryCode(to)"
      class="fi"
      :class="`fi-${getCountryCode(to)}`"
      style="width: 20px; height: 15px; border-radius: 2px;"
    />
    <span class="text-sm">{{ to.toUpperCase() }}</span>
  </span>
</template>

<style scoped lang="scss">
@use '@/styles/flags';
</style>
