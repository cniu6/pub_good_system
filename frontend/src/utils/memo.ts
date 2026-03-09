import { local } from '@/utils'

/**
 * 解析备注字段：支持多语言 JSON 格式
 * 如果 memo 是 JSON 格式 {"zhCN":"xxx","enUS":"yyy"}，返回当前语言对应的文本
 * 否则原样返回纯文本
 */
export function parseMemo(memo: string): string {
  if (!memo || memo[0] !== '{') return memo || ''
  try {
    const i18n = JSON.parse(memo) as Record<string, string>
    const lang = (local.get('lang') as string) || import.meta.env.VITE_DEFAULT_LANG || 'zhCN'
    // 精确匹配 → 回退中文 → 第一个可用
    if (i18n[lang]) return i18n[lang]
    if (i18n.zhCN) return i18n.zhCN
    const first = Object.values(i18n)[0]
    return first || memo
  }
  catch {
    return memo
  }
}
