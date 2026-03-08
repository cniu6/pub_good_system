import type { NDateLocale, NLocale } from 'naive-ui'
import { i18n } from '@/modules/i18n'
import { dateZhCN, zhCN } from 'naive-ui'

export function setLocale(locale: App.lang) {
  i18n.global.locale.value = locale
}

export const $t = i18n.global.t

const langToBackend: Record<App.lang, string> = { zhCN: 'zh-CN', enUS: 'en-US' }
const langToFrontend: Record<string, App.lang> = { 'zh-CN': 'zhCN', 'en-US': 'enUS' }

export function langToBackendFormat(lang: App.lang): string {
  return langToBackend[lang] || 'zh-CN'
}

export function langToFrontendFormat(backendLang: string): App.lang {
  return langToFrontend[backendLang] || 'zhCN'
}

export const naiveI18nOptions: Record<App.lang, { locale: NLocale | null, dateLocale: NDateLocale | null }> = {
  zhCN: {
    locale: zhCN,
    dateLocale: dateZhCN,
  },
  enUS: {
    locale: null,
    dateLocale: null,
  },
}
