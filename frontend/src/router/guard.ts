import type { Router } from 'vue-router'
import type { AppRouteMode } from './index'
import { useAppStore, useAuthStore, useRouteStore, useTabStore } from '@/store'
import { i18n } from '@/modules/i18n'
import { authStorage, resolveI18nText } from '@/utils'
import { buildAdminEntryUrl, getAdminBasePath } from './constants'

function reportGuardError(message: string, error: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

async function loadAdminRoutesDynamic() {
  try {
    const { getAdminRoutes } = await import(
      /* webpackChunkName: "admin-core" */
      '@/router/admin.routes',
    )
    const { createAdminMenus } = await import('@/store/router/helper')
    return { getAdminRoutes, createAdminMenus }
  }
  catch (error) {
    reportGuardError('[Security] Failed to load admin routes dynamically:', error)
    return null
  }
}

const ADMIN_LOGIN_PATH = '/admin/login'
const ADMIN_PUBLIC_PATHS = new Set([ADMIN_LOGIN_PATH, '/403', '/404', '/500', '/loading'])

function isAdminRoutePath(path: string, mode: AppRouteMode, adminPath: string) {
  if (mode === 'admin')
    return !ADMIN_PUBLIC_PATHS.has(path)
  return path === adminPath || path.startsWith(`${adminPath}/`)
}

export function setupRouterGuard(router: Router, mode: AppRouteMode = 'user') {
  const appStore = useAppStore()
  const routeStore = useRouteStore()
  const tabStore = useTabStore()

  router.beforeEach(async (to, _from, next) => {
    const adminPath = getAdminBasePath()
    const isAdminRoute = isAdminRoutePath(to.path, mode, adminPath)
    const authStore = useAuthStore()

    // /admin/login 是管理端 hash 内部路由，只能从独立管理入口加载。
    // 用户端 history 路由访问它时，保持在普通用户登录页，不能借此切换认证上下文。
    if (mode === 'user' && to.path === ADMIN_LOGIN_PATH) {
      next({ path: '/user/login', replace: true })
      return
    }

    if (to.meta.href) {
      window.open(to.meta.href)
      next(false)
      return
    }

    // 会话恢复弹窗显示时保持当前路由，禁止守卫把用户带去独立登录页。
    if (authStore.needsReauthentication) {
      next(false)
      return
    }

    appStore.showProgress && window.$loadingBar?.start()

    const isLogin = Boolean(authStorage.get('accessToken'))
    const roleValue = authStorage.get('role')
    const roles = Array.isArray(roleValue) ? roleValue : (roleValue ? [roleValue] : [])
    // 管理端准入优先看会话 JWT auth_guard；旧会话没有该字段时回退到 role。
    const storedGuard = authStorage.get('authGuard')
    const hasAdminRole = storedGuard
      ? storedGuard === 'admin'
      : roles.includes('admin')
    // 用户端可能在其它标签刚完成登录；路由守卫读到 localStorage 后要同步 Pinia，
    // 否则路由虽放行，页面仍没有 userInfo / Presence / 自动刷新状态。
    if (mode === 'user' && isLogin) {
      if (!authStore.isLogin || authStore.token !== authStorage.get('accessToken'))
        authStore.hydrateFromStorage()
    }
    // 管理端已登录但会话 guard 不是 admin：进不了受保护页，避免用 user token 打 admin API。
    if (mode === 'admin' && isLogin && !authStore.isLogin)
      authStore.hydrateFromStorage()

    routeStore.setMenuMode(to.path, mode)

    if (mode === 'user' && isAdminRoute && isLogin && hasAdminRole) {
      window.location.replace(buildAdminEntryUrl(to.fullPath))
      next(false)
      return
    }

    if (isAdminRoute && (!isLogin || !hasAdminRole)) {
      if (isLogin) {
        next({ path: '/403', replace: true })
      }
      else {
        // 管理端未登录：进入专用管理员登录页，提交固定使用 admin guard。
        next({ path: ADMIN_LOGIN_PATH, query: { redirect: to.fullPath } })
      }
      return
    }

    // 未登录：用户业务路径与显式 requiresAuth 路由一律拦到登录
    // 注意：动态路由尚未挂载时 meta.requiresAuth 可能为空，故额外用 /user/* 前缀兜底
    const isUserAuthArea = mode === 'user'
      && to.path.startsWith('/user/')
      && to.path !== '/user/login'
      && to.path !== '/user/register'
    if (!isLogin && (to.meta.requiresAuth === true || isUserAuthArea)) {
      const redirect = to.name === '404' || to.name === 'notFoundCatchAll' ? undefined : to.fullPath
      next({ path: '/user/login', query: redirect ? { redirect } : undefined })
      return
    }

    // 仅已登录（或明确放行的公开页）才初始化动态路由，避免未登录刷用户路由
    if (!routeStore.isInitAuthRoute && to.name !== 'login' && to.name !== 'register' && to.name !== 'admin-login') {
      if (!isLogin && mode === 'user') {
        next()
        return
      }
      try {
        await routeStore.initAuthRoute(mode)
        next({
          path: to.fullPath,
          replace: true,
          query: to.query,
          hash: to.hash,
        })
        return
      }
      catch {
        const redirect = to.fullPath !== '/' ? to.fullPath : undefined
        const loginPath = mode === 'admin' ? ADMIN_LOGIN_PATH : '/user/login'
        next({ path: loginPath, query: redirect ? { redirect } : undefined })
        return
      }
    }

    if (routeStore.isInitAuthRoute && (to.name === '404' || to.name === 'notFoundCatchAll')) {
      if (mode === 'user' && isAdminRoute && hasAdminRole) {
        window.location.replace(buildAdminEntryUrl(to.fullPath))
        next(false)
        return
      }

      if (mode === 'admin' && isAdminRoute && hasAdminRole) {
        try {
          if (!router.hasRoute('admin-root')) {
            const adminModule = await loadAdminRoutesDynamic()
            if (adminModule) {
              const adminRoutes = adminModule.getAdminRoutes()
              adminRoutes.forEach(route => router.addRoute(route))
              routeStore.setAdminMenus(adminModule.createAdminMenus(adminRoutes))
              next({
                path: to.fullPath,
                replace: true,
                query: to.query,
                hash: to.hash,
              })
              return
            }
          }
        }
        catch (error) {
          reportGuardError('[Security] Failed to load admin routes in guard:', error)
        }
      }

      if (to.fullPath !== '/404') {
        const resolved = router.resolve(to.fullPath)
        if (resolved.name && resolved.name !== '404' && resolved.name !== 'notFoundCatchAll') {
          next({
            path: to.fullPath,
            replace: true,
            query: to.query,
            hash: to.hash,
          })
          return
        }
      }

      next()
      return
    }

    if ((to.name === 'login' || to.name === 'register' || to.name === 'admin-login') && isLogin) {
      if (hasAdminRole) {
        if (mode === 'user') {
          window.location.replace(buildAdminEntryUrl(adminPath))
          next(false)
          return
        }
        next({ path: '/dashboard' })
      }
      else {
        next({ path: import.meta.env.VITE_HOME_PATH || '/user/dashboard/workbench' })
      }
      return
    }

    next()
  })

  router.beforeResolve((to) => {
    routeStore.setActiveMenu(to.meta.activeMenu ?? to.fullPath)
    tabStore.addTab(to)
    tabStore.setCurrentTab(to.fullPath as string)
  })

  router.afterEach((to) => {
    const { t } = i18n.global
    const appTitle = resolveI18nText(t, 'app.title') || import.meta.env.VITE_APP_TITLE || import.meta.env.VITE_APP_NAME
    const pageTitle = resolveI18nText(t, to.meta.title)
    document.title = pageTitle ? `${pageTitle} - ${appTitle}` : appTitle
    appStore.showProgress && window.$loadingBar?.finish()
  })

  window.addEventListener('app:locale-changed', () => {
    const { t } = i18n.global
    const appTitle = resolveI18nText(t, 'app.title') || import.meta.env.VITE_APP_TITLE || import.meta.env.VITE_APP_NAME
    const currentRoute = router.currentRoute.value
    const pageTitle = resolveI18nText(t, currentRoute.meta.title)
    document.title = pageTitle ? `${pageTitle} - ${appTitle}` : appTitle
  })
}
