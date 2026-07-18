/**
 * 通用数值/容量格式化（无业务耦合）
 * dashboard / server-management 共用
 */
import { $t } from './i18n'

/** 将百分比限制在 0–100 */
export function normalizePercent(value?: number | null): number {
  const n = Number(value ?? 0)
  if (!Number.isFinite(n) || n < 0)
    return 0
  if (n > 100)
    return 100
  return n
}

/** 格式化百分比，默认 1 位小数 */
export function formatPercent(value?: number | null, digits = 1): string {
  return `${normalizePercent(value).toFixed(digits)}%`
}

/** 字节数 → B/KB/MB/GB/TB */
export function formatBytes(value?: number | null): string {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount) || amount < 0)
    return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = amount
  let idx = 0
  while (size >= 1024 && idx < units.length - 1) {
    size /= 1024
    idx++
  }
  if (idx === 0)
    return `${size.toFixed(0)} B`
  return `${size.toFixed(2)} ${units[idx]}`
}

/** MB 数值 → MB/GB/TB */
export function formatStorageFromMB(value?: number | null): string {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount) || amount < 0)
    return '-'
  if (amount >= 1024 * 1024)
    return `${(amount / 1024 / 1024).toFixed(2)} TB`
  if (amount >= 1024)
    return `${(amount / 1024).toFixed(2)} GB`
  return `${amount.toFixed(1)} MB`
}

/** GB 数值 → GB/TB */
export function formatStorageFromGB(value?: number | null): string {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount) || amount < 0)
    return '-'
  if (amount >= 1024)
    return `${(amount / 1024).toFixed(2)} TB`
  return `${amount.toFixed(2)} GB`
}

/** 运行时长（秒）→ 本地化「天时分秒」 */
export function formatUptime(seconds?: number | null): string {
  const n = Number(seconds ?? 0)
  if (!Number.isFinite(n) || n < 0)
    return '-'
  const totalSeconds = Math.floor(n)
  const day = Math.floor(totalSeconds / 86400)
  const hour = Math.floor((totalSeconds % 86400) / 3600)
  const minute = Math.floor((totalSeconds % 3600) / 60)
  const second = totalSeconds % 60
  return $t('adminSettings.uptimePreciseFormat', { day, hour, minute, second })
}
