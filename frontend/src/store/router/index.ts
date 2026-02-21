import { defineStore } from 'pinia'
import { useRouter } from 'vue-router'
import type { MenuOption } from 'naive-ui'
import { router } from '@/router'
import { staticRoutes } from '@/router/routes.static'
import { getAdminRoutes } from '@/router/admin.routes'
import { fetchUserRoutes } from '@/service'
import { local, $t } from '@/utils'
import { createAdminMenus, createMenus, createRoutes, generateCacheRoutes } from './helper'

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
    },
    // set the currently highlighted menu key
    setActiveMenu(key: string) {
      this.activeMenu = key
    },
    // 设置菜单模式（用户端/管理端）
    setMenuMode(path: string) {
      const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
      this.menuMode = path.startsWith(adminPath) ? 'admin' : 'user'
    },

    async initRouteInfo() {
      if (import.meta.env.VITE_ROUTE_LOAD_MODE === 'dynamic') {
        try {
          // Get user's route
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
      else {
        this.rowRoutes = staticRoutes
        return staticRoutes
      }
    },
    async initAuthRoute() {
      this.isInitAuthRoute = false

      try {
        // Initialize route information
        const rowRoutes = await this.initRouteInfo()
        if (!rowRoutes) {
          const error = new Error('Failed to get route information')
          window.$message.error($t(`app.getRouteError`))
          throw error
        }
        this.rowRoutes = rowRoutes

// Generate actual route and insert
        const routes = createRoutes(rowRoutes)
        router.addRoute(routes)

        // Generate side menu
        this.menus = createMenus(rowRoutes)

        // Generate the route cache
        this.cacheRoutes = generateCacheRoutes(rowRoutes)

        // 如果用户是管理员，加载管理端路由和菜单
        const roleValue = local.get('role')
        const roles = Array.isArray(roleValue) ? roleValue : (roleValue ? [roleValue] : [])
        if (roles.includes('admin')) {
          const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
          const adminRoutes = getAdminRoutes(adminPath)
          adminRoutes.forEach(route => router.addRoute(route))
          this.adminMenus = createAdminMenus(adminRoutes)
        }

        this.isInitAuthRoute = true
      }
      catch (error) {
        // 重置状态并重新抛出错误
        this.isInitAuthRoute = false
        throw error
      }
    },
  },
})
