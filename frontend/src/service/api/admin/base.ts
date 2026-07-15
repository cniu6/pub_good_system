/**
 * 管理端 REST API 基础路径统一 helper
 *
 * - 页面入口路径：VITE_ADMIN_BASE_PATH / ADMIN_PATH（如 /system-mgr），与 API 无关
 * - API 前缀：VITE_ADMIN_API_PATH / ADMIN_API_PATH（默认 /admin）
 * - 完整 base：/api/v1{VITE_ADMIN_API_PATH}，例如 /api/v1/admin
 *
 * 换 API 路径：同时改根 .env 的 ADMIN_API_PATH 与 frontend 的 VITE_ADMIN_API_PATH 后重建
 * 换页面入口：同时改 ADMIN_PATH 与 VITE_ADMIN_BASE_PATH
 */

/** 规范化管理 API 前缀：保证以 / 开头、去掉尾斜杠；空则默认 /admin */
export function normalizeAdminApiPath(path?: string | null): string {
  let p = (path ?? '').trim()
  if (!p)
    return '/admin'
  if (!p.startsWith('/'))
    p = `/${p}`
  return p.replace(/\/+$/, '') || '/admin'
}

/**
 * 返回管理端 API 完整 base，形如 `/api/v1/admin`
 * 路径来自 import.meta.env.VITE_ADMIN_API_PATH，默认 /admin
 */
export function getAdminApiBase(): string {
  const path = normalizeAdminApiPath(import.meta.env.VITE_ADMIN_API_PATH as string | undefined)
  return `/api/v1${path}`
}

/**
 * 拼接管理端 API 子路径，例如 adminApiUrl('/users') → /api/v1/admin/users
 */
export function adminApiUrl(subPath = ''): string {
  const base = getAdminApiBase()
  if (!subPath)
    return base
  return `${base}${subPath.startsWith('/') ? subPath : `/${subPath}`}`
}
