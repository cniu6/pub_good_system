/**
 * 手机号校验 / 规范化（与后端 utils/phone.go 对齐）
 *
 * 后台开关 mobile_cn_only：
 * - true（默认）：仅中国大陆 +86（存 11 位）
 * - false：允许国际 E.164（存 +国家码...）；UI 为国家选择 + 本地号码
 */
import { Regex } from '@/constants/Regex'
import { DIAL_COUNTRIES, getDialCountry, type DialCountry } from '@/constants/dialCodes'

/** 去掉空格/横线/括号，保留开头的 + */
export function normalizePhoneInput(raw: string): string {
  const s = String(raw || '').trim()
  if (!s)
    return ''
  let out = ''
  for (let i = 0; i < s.length; i++) {
    const ch = s[i]
    if (/\d/.test(ch))
      out += ch
    else if (ch === '+' && out.length === 0)
      out += ch
  }
  return out
}

/** 区号 + 国内号码 → E.164（去掉国内号开头的 0） */
export function composeE164(dialCode: string, national: string): string {
  const dial = String(dialCode || '').replace(/^\+/, '').replace(/\D/g, '')
  let n = String(national || '').replace(/\D/g, '')
  if (n.startsWith('0'))
    n = n.slice(1)
  if (!dial || !n)
    return ''
  return `+${dial}${n}`
}

/**
 * 规范化并校验。成功返回规范化后的号码，失败返回 null。
 */
export function normalizeAndValidateMobile(raw: string, cnOnly: boolean): string | null {
  let s = normalizePhoneInput(raw)
  if (!s)
    return null

  if (cnOnly) {
    if (s.startsWith('+86'))
      s = s.slice(3)
    else if (s.startsWith('86') && s.length === 13)
      s = s.slice(2)
    else if (s.startsWith('0086') && s.length === 15)
      s = s.slice(4)
    return new RegExp(Regex.MobileCN).test(s) ? s : null
  }

  // 国际模式
  if (s.startsWith('+'))
    return new RegExp(Regex.MobileE164).test(s) ? s : null
  // 大陆 11 位便捷输入 → +86（北美号请用区号选择 +1）
  if (new RegExp(Regex.MobileCN).test(s))
    return `+86${s}`
  if (s.startsWith('86') && s.length === 13 && new RegExp(Regex.MobileCN).test(s.slice(2)))
    return `+${s}`
  if (s.startsWith('0086') && s.length === 15 && new RegExp(Regex.MobileCN).test(s.slice(4)))
    return `+86${s.slice(4)}`
  return null
}

export function isValidMobile(raw: string, cnOnly: boolean): boolean {
  return normalizeAndValidateMobile(raw, cnOnly) !== null
}

/**
 * 把已存号码拆成「国家 + 本地号」，供区号选择器回显。
 * 优先最长区号匹配；匹配失败时回退到 fallbackCode。
 */
export function splitMobileToCountryNational(
  raw: string,
  fallbackCode = 'US',
): { country: DialCountry, national: string } {
  const s = normalizePhoneInput(raw)
  const fallback = getDialCountry(fallbackCode) || getDialCountry('US')!

  if (!s) {
    return { country: fallback, national: '' }
  }

  // 纯大陆 11 位
  if (new RegExp(Regex.MobileCN).test(s)) {
    return { country: getDialCountry('CN')!, national: s }
  }

  let e164 = s
  if (!e164.startsWith('+')) {
    if (s.startsWith('86') && s.length === 13 && new RegExp(Regex.MobileCN).test(s.slice(2)))
      e164 = `+${s}`
    else
      return { country: fallback, national: s.replace(/\D/g, '') }
  }

  const digits = e164.slice(1)
  // 最长区号优先（避免 1 吃掉 18x）
  const sorted = [...DIAL_COUNTRIES].sort((a, b) => b.dialCode.length - a.dialCode.length)
  for (const c of sorted) {
    if (digits.startsWith(c.dialCode)) {
      return { country: c, national: digits.slice(c.dialCode.length) }
    }
  }
  return { country: fallback, national: digits }
}
