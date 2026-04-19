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
