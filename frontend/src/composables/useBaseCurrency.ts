import { computed } from 'vue'
import { useSettingsStore } from '@/store'
import { formatCurrency as formatCurrencyUtil, getCurrencySymbol } from '@/utils'

/**
 * 全局本位币展示组合函数
 *
 * 从 settingsStore 读取当前本位币配置（默认 CNY），
 * 提供货币符号映射与格式化输出。
 */
export function useBaseCurrency() {
  const settingsStore = useSettingsStore()

  // 当前本位币，默认回退 CNY
  const baseCurrency = computed(() => settingsStore.baseCurrency)

  // 根据本位币返回对应符号
  const currencySymbol = computed(() => getCurrencySymbol(baseCurrency.value))

  /**
   * 格式化金额
   * @param value 原始数值，支持 null/undefined
   * @returns 带符号的两位小数字符串
   */
  function formatCurrency(value?: number | string | null): string {
    return formatCurrencyUtil(baseCurrency.value, value)
  }

  return {
    baseCurrency,
    currencySymbol,
    formatCurrency,
  }
}
