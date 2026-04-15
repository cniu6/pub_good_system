import { defineStore } from 'pinia'
import type { RouteRecordRaw } from 'vue-router'
import type { AppRouteMode } from '@/router'
import { router } from '@/router'
import { getAdminBasePath } from '@/router/constants'
import { staticRoutes } from '@/router/routes.static'
import { fetchUserRoutes } from '@/service'
import { $t, authStorage } from '@/utils'
import { createAdminMenus, createMenus, createRoutes, generateCacheRoutes } from './helper'
import type { AdminMenuRoute } from './helper'

function reportRouteInitError(message: string, error: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

interface RouteMenuStateItem {
  key?: string | number
  label?: unknown
  icon?: unknown
  children?: unknown
  id?: number
  pid?: number | null
}

interface RoutesStatus {
  isInitAuthRoute: boolean
  menus: RouteMenuStateItem[]
  adminMenus: RouteMenuStateItem[]
  rowRoutes: AppRoute.RowRoute[]
  activeMenu: string | null
  cacheRoutes: string[]
  menuMode: 'user' | 'admin'
}

export const useRouteStore = defineStore('route-store', {
  state: (): RoutesStatus => {
    return {
      isInitAuthRoute: false,
      activeMenu: null,
      menus: [],
      adminMenus: [],
      rowRoutes: [],
      cacheRoutes: [],
      menuMode: 'user',
    }
  },
  actions: {
    resetRouteStore() {
      this.resetRoutes()
      this.$reset()
    },
    resetRoutes() {
      if (router.hasRoute('appRoot'))
        router.removeRoute('appRoot')
      if (router.hasRoute('admin-root'))
        router.removeRoute('admin-root')
    },
    setActiveMenu(key: string) {
      this.activeMenu = key
    },
    setAdminMenus(menus: RouteMenuStateItem[]) {
      Reflect.set(this, 'adminMenus', menus)
    },
    setMenuMode(path: string, mode: AppRouteMode = 'user') {
      if (mode === 'admin') {
        this.menuMode = 'admin'
        return
      }
      const adminPath = getAdminBasePath()
      this.menuMode = (path === adminPath || path.startsWith(`${adminPath}/`)) ? 'admin' : 'user'
    },
    async initRouteInfo() {
      if (import.meta.env.VITE_ROUTE_LOAD_MODE === 'dynamic') {
        try {
          const result = await fetchUserRoutes({
            id: 1,
          })

          if (!result.isSuccess || !result.data) {
            throw new Error('Failed to fetch user routes')
          }

          return result.data as AppRoute.RowRoute[]
        }
        catch (error) {
          reportRouteInitError('Failed to initialize route info:', error)
          throw error
        }
      }

      this.rowRoutes = staticRoutes
      return staticRoutes
    },
    async initAuthRoute(mode: AppRouteMode = 'user') {
      this.isInitAuthRoute = false

      try {
        if (mode === 'user') {
          const rowRoutes = await this.initRouteInfo()
          if (!rowRoutes) {
            const error = new Error('Failed to get route information')
            window.$message.error($t('app.getRouteError'))
            throw error
          }
          this.rowRoutes = rowRoutes

          const routes = createRoutes(rowRoutes)
          router.addRoute(routes)

          this.menus = createMenus(rowRoutes)
          this.cacheRoutes = generateCacheRoutes(rowRoutes)
        }
        else {
          // 管理端入口仅加载管理端路由，避免与普通用户路由冲突
          this.rowRoutes = []
          this.menus = []
          this.cacheRoutes = []
        }

        const roleValue = authStorage.get('role')
        const roles = Array.isArray(roleValue) ? roleValue : (roleValue ? [roleValue] : [])

        if (mode === 'admin' && roles.includes('admin')) {
          try {
            const adminModule = await import(
              /* webpackChunkName: "admin-core" */
              /* vite: {"chunkName": "admin-core"} */
              '@/router/admin.routes',
            ) as { getAdminRoutes: () => RouteRecordRaw[] }
            const adminRoutes: RouteRecordRaw[] = adminModule.getAdminRoutes()
            for (const route of adminRoutes) {
              router.addRoute(route)
            }
            // 使用 helper 生成支持嵌套层级的管理端菜单
            const adminMenus = createAdminMenus(adminRoutes as AdminMenuRoute[])
            this.setAdminMenus(adminMenus)
          }
          catch (error) {
            reportRouteInitError('[Security] Failed to load admin routes:', error)
          }
        }
        else {
          this.setAdminMenus([])
        }

        this.isInitAuthRoute = true
      }
      catch (error) {
        this.isInitAuthRoute = false
        throw error
      }
    },
  },
})
