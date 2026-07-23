import { defineStore } from 'pinia'
import type { RouteRecordRaw } from 'vue-router'
import type { AppRouteMode } from '@/router'
import { router } from '@/router'
import { getAdminBasePath } from '@/router/constants'
import { staticRoutes } from '@/router/routes.static'
import { fetchUserRoutes } from '@/service'
import { $t, authStorage } from '@/utils'
import { useAuthStore } from '../auth'
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
  /** 管理端路由加载失败时的错误；成功后清空。避免空菜单被当成初始化成功 */
  authRouteError: string | null
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
      authRouteError: null,
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
          // 动态路由模式必须用当前登录用户的 id 去后端换取专属路由，
          // 而不是硬编码 id:1（否则所有用户都会拿到同一份/错误用户的路由）
          const currentUserId = useAuthStore().userInfo?.id
          if (!currentUserId) {
            throw new Error('Failed to initialize route info: missing current user id')
          }
          const result = await fetchUserRoutes({
            id: currentUserId,
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
      this.authRouteError = null

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
            const adminMenus = createAdminMenus(adminRoutes as AdminMenuRoute[])
            this.setAdminMenus(adminMenus)
          }
          catch (error) {
            // 管理端路由加载失败：不能标成初始化成功，否则会出现空侧边栏却继续当成功态
            reportRouteInitError('[Security] Failed to load admin routes:', error)
            this.authRouteError = error instanceof Error ? error.message : String(error)
            this.isInitAuthRoute = false
            this.setAdminMenus([])
            throw error
          }
        }
        else {
          this.setAdminMenus([])
        }

        this.authRouteError = null
        this.isInitAuthRoute = true
      }
      catch (error) {
        this.isInitAuthRoute = false
        throw error
      }
    },
  },
})
