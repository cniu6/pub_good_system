export function parseBooleanSetting(value: unknown, fallback = false): boolean {
  if (typeof value === 'boolean')
    return value

  if (typeof value === 'number')
    return value !== 0

  const normalized = String(value ?? '').trim().toLowerCase()
  if (!normalized)
    return fallback

  return normalized === 'true' || normalized === '1'
}

export function parseNumberSetting(value: unknown, fallback: number): number {
  if (typeof value === 'number' && Number.isFinite(value))
    return value

  const normalized = String(value ?? '').trim()
  if (!normalized)
    return fallback

  const parsed = Number(normalized)
  return Number.isFinite(parsed) ? parsed : fallback
}

// 以下三个 clamp 函数是管理端「日志保留」运行时配置的通用校验规则，
// operation/api/sms/email 四个日志管理页面原来各自复制了一份完全相同的实现，这里统一。

/** 日志查询时间窗口（天）：1~365 */
export function normalizeLogQueryDays(value: unknown, fallback = 30): number {
  return Math.min(365, Math.max(1, Math.floor(parseNumberSetting(value, fallback))))
}

/** 日志全局保留上限（条）：100~200000 */
export function normalizeLogMaxCount(value: unknown, fallback = 1000): number {
  return Math.min(200000, Math.max(100, Math.floor(parseNumberSetting(value, fallback))))
}

/** API 日志自动清理间隔（秒）：60~86400 */
export function normalizeAPILogCleanupIntervalSeconds(value: unknown, fallback = 600): number {
  return Math.min(86400, Math.max(60, Math.floor(parseNumberSetting(value, fallback))))
}

/** 日志按用户/收件人保留上限（条）：1~200000 */
export function normalizeLogPerUserMaxCount(value: unknown, fallback = 1000): number {
  return Math.min(200000, Math.max(1, Math.floor(parseNumberSetting(value, fallback))))
}
