/**
 * 敏感字段脱敏工具（与后端 utils / models.MaskPhone 大致对齐）
 * 用于管理端列表/详情展示，避免明文完整回显。
 */

/** 手机号脱敏：保留前 3（国际号前 4）+ 后 4，中间 */
export function maskPhone(phone: string): string {
  const s = String(phone || '').trim()
  if (s.length < 7)
    return s
  // 国际号（以 + 开头且足够长）多留一位国家码，可读性更好
  const prefix = s.startsWith('+') && s.length >= 12 ? 4 : 3
  return `${s.slice(0, prefix)}****${s.slice(-4)}`
}

/**
 * 证件号脱敏（按字符数，兼容护照等含字母场景）：
 * - 空：原样
 * - <=4：仅保留末位
 * - 5~10：保留首尾各 1 位
 * - >10：保留前 6 + 后 4
 */
export function maskCertificateNo(no: string): string {
  const s = String(no || '').trim()
  if (!s)
    return ''
  const chars = Array.from(s)
  const n = chars.length

  if (n <= 4)
    return `${'*'.repeat(n - 1)}${chars[n - 1]}`
  if (n <= 10)
    return `${chars[0]}${'*'.repeat(n - 2)}${chars[n - 1]}`
  const prefixLen = 6
  const suffixLen = 4
  const maskLen = n - prefixLen - suffixLen
  return `${chars.slice(0, prefixLen).join('')}${'*'.repeat(maskLen)}${chars.slice(n - suffixLen).join('')}`
}

/**
 * 银行卡/收款账号脱敏：
 * - <=4：全掩码
 * - <=8：保留首尾各 1 位
 * - 更长：保留前 4 + 后 4
 */
export function maskAccountNo(no: string): string {
  const s = String(no || '').trim()
  if (!s)
    return ''
  const chars = Array.from(s)
  const n = chars.length

  if (n <= 4)
    return '*'.repeat(n)
  if (n <= 8)
    return `${chars[0]}${'*'.repeat(n - 2)}${chars[n - 1]}`
  return `${chars.slice(0, 4).join('')}${'*'.repeat(n - 8)}${chars.slice(n - 4).join('')}`
}
