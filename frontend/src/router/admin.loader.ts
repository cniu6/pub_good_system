import type { RouteRecordRaw } from 'vue-router'
import { getAdminRoutes } from './admin.routes'

/**
 * 获取管理端路径前缀
 */
export function getAdminPath(): string {
  return import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
}

/**
 * 动态加载管理端路由
 * 通过动态 import 实现代码分割，普通用户不会加载管理端代码
 */
export async function loadAdminRoutes(): Promise<RouteRecordRaw[]> {
  const adminPath = getAdminPath()
  return getAdminRoutes(adminPath)
}
