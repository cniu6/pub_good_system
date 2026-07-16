/**
 * 管理端 REST API 基础路径统一 helper
 *
 * - 页面入口路径：VITE_ADMIN_BASE_PATH / ADMIN_PATH（如 /system-mgr），与 API 无关
 * - API 前缀：运行时优先用后端 app-config 的 admin_api_path（与 ADMIN_API_PATH 一致）
 * - 回退：构建时 VITE_ADMIN_API_PATH（默认 /admin）
 * - 完整 base：/api/v1{adminApiPath}，例如 /api/v1/admin
 *
 * 换 API 路径：只改根 .env 的 ADMIN_API_PATH 并重启后端；
 * 前端启动时会从 /api/v1/public/app-config 注入，无需为改 API 前缀单独重建前端。
 * （VITE_ADMIN_API_PATH 仅作加载 app-config 前的回退）
 */

/** 运行时注入的管理端 API 前缀（来自后端 app-config） */
let runtimeAdminApiPath: string | null = null

/** 规范化管理 API 前缀：保证以 / 开头、去掉尾斜杠；空则默认 /admin */
export function normalizeAdminApiPath(path?: string | null): string {
  let p = (path ?? '').trim()
  if (!p)
    return '/admin'
  if (!p.startsWith('/'))
    p = `/${p}`
  return p.replace(/\/+$/g, '') || '/admin'
}

/**
 * 由 bootstrap / settingsStore 在拿到 app-config 后调用。
 * 传入后端返回的 admin_api_path，实现前端管理端请求与后端路由一致。
 */
export function setRuntimeAdminApiPath(path?: string | null): string {
  runtimeAdminApiPath = normalizeAdminApiPath(path)
  return runtimeAdminApiPath
}

/** 当前生效的管理端 API 前缀（如 /admin、/mgr-api） */
export function getAdminApiPath(): string {
  if (runtimeAdminApiPath)
    return runtimeAdminApiPath
  return normalizeAdminApiPath(import.meta.env.VITE_ADMIN_API_PATH as string | undefined)
}

/**
 * 返回管理端 API 完整 base，形如 `/api/v1/admin`
 * 必须在请求时调用（不要在模块顶层 const 固化），以便运行时注入后生效
 */
export function getAdminApiBase(): string {
  return `/api/v1${getAdminApiPath()}`
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
