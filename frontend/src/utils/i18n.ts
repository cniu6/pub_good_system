import type { NDateLocale, NLocale } from 'naive-ui'
import type { RouteRecordNameGeneric } from 'vue-router'
import { i18n } from '@/modules/i18n'
import { dateZhCN, zhCN } from 'naive-ui'

export function setLocale(locale: App.lang) {
  i18n.global.locale.value = locale
}

export const $t = i18n.global.t

export function isI18nKey(key: string) {
  return key.includes('.') && /^[\w.-]+$/.test(key)
}

export function resolveI18nText(translate: (key: string) => string, key: unknown, fallback = '') {
  if (!key || typeof key !== 'string')
    return fallback
  if (!isI18nKey(key))
    return key
  const text = translate(key)
  return text === key ? (fallback || key) : text
}

export function resolveRouteTitle(translate: (key: string) => string, name?: RouteRecordNameGeneric | null, title?: string) {
  const routeName = typeof name === 'string' ? name : ''
  const fallback = title || routeName
  return routeName ? resolveI18nText(translate, `route.${routeName}`, fallback) : fallback
}

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
