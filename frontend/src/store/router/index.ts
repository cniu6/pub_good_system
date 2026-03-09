import { h } from 'vue'
import { defineStore } from 'pinia'
import type { MenuOption } from 'naive-ui'
import { RouterLink, type RouteRecordRaw } from 'vue-router'
import type { AppRouteMode } from '@/router'
import { router } from '@/router'
import { getAdminBasePath } from '@/router/constants'
import { staticRoutes } from '@/router/routes.static'
import { fetchUserRoutes } from '@/service'
import { local, $t, renderIcon } from '@/utils'
import { createMenus, createRoutes, generateCacheRoutes } from './helper'

interface RoutesStatus {
  isInitAuthRoute: boolean
  menus: MenuOption[]
  adminMenus: MenuOption[]
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
  getters: {
    currentMenus(state) {
      return state.menuMode === 'admin' ? state.adminMenus : state.menus
    },
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

          return result.data
        }
        catch (error) {
          console.error('Failed to initialize route info:', error)
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

        const roleValue = local.get('role')
        const roles = Array.isArray(roleValue) ? roleValue : (roleValue ? [roleValue] : [])

        if (mode === 'admin' && roles.includes('admin')) {
          try {
            const adminModule = await import(
              /* webpackChunkName: "admin-core" */
              /* vite: {"chunkName": "admin-core"} */
              '@/router/admin.routes'
            ) as { getAdminRoutes: () => unknown[] }
            const adminRoutes = adminModule.getAdminRoutes()
            for (const route of adminRoutes) {
              router.addRoute(route as RouteRecordRaw)
            }
            this.adminMenus = adminRoutes.flatMap((route) => {
              const parent = route as {
                path?: string
                children?: Array<{
                  path?: string
                  name?: string | symbol | null
                  meta?: {
                    hide?: boolean
                    title?: string
                    icon?: string
                  }
                }>
              }
              if (!parent.children || !parent.path) {
                return []
              }
              return parent.children
                .filter(child => !child.meta?.hide && !!child.path)
                .map((child) => {
                  const fullPath = parent.path!.endsWith('/')
                    ? `${parent.path}${child.path}`
                    : `${parent.path}/${child.path}`
                  return {
                    label: () => h(RouterLink, { to: { path: fullPath } }, { default: () => child.meta?.title || String(child.name || child.path) }),
                    key: fullPath,
                    icon: child.meta?.icon ? renderIcon(child.meta.icon) : undefined,
                  } satisfies MenuOption
                })
            })
          }
          catch (error) {
            console.error('[Security] Failed to load admin routes:', error)
          }
        }
        else {
          this.adminMenus = []
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
