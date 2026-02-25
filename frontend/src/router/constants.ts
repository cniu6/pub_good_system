const USER_BASE = import.meta.env.VITE_BASE_URL || '/'
const ADMIN_BASE_PATH = import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr'

function ensureLeadingSlash(path: string): string {
  if (!path)
    return '/'
  return path.startsWith('/') ? path : `/${path}`
}

function trimTrailingSlash(path: string): string {
  if (path === '/')
    return '/'
  return path.replace(/\/+$/, '') || '/'
}

function normalizeBase(path: string): string {
  const normalized = trimTrailingSlash(ensureLeadingSlash(path))
  return normalized.endsWith('/') ? normalized : `${normalized}/`
}

function normalizePath(path: string): string {
  return trimTrailingSlash(ensureLeadingSlash(path))
}

function normalizeHashTarget(targetPath: string): string {
  const [rawPathWithQuery, rawFragment = ''] = targetPath.split('#')
  const [rawPath = '', rawQuery = ''] = rawPathWithQuery.split('?')

  const adminBase = getAdminBasePath()
  const normalizedRawPath = ensureLeadingSlash(rawPath || '/')

  let hashPath = normalizedRawPath
  if (normalizedRawPath === adminBase || normalizedRawPath === `${adminBase}/`) {
    hashPath = '/dashboard'
  }
  else if (normalizedRawPath.startsWith(`${adminBase}/`)) {
    hashPath = normalizedRawPath.slice(adminBase.length) || '/dashboard'
  }

  const query = rawQuery ? `?${rawQuery}` : ''
  const fragment = rawFragment ? `#${rawFragment}` : ''
  return `${hashPath}${query}${fragment}`
}

export function getUserBase() {
  return normalizeBase(USER_BASE)
}

export function getAdminBasePath() {
  return normalizePath(ADMIN_BASE_PATH)
}

export function getAdminEntryBase() {
  return normalizeBase(getAdminBasePath())
}

export function toAdminHashPath(targetPath: string) {
  return normalizeHashTarget(targetPath)
}

export function buildAdminEntryUrl(targetPath: string) {
  return `${getAdminEntryBase()}#${toAdminHashPath(targetPath)}`
}
