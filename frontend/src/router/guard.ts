import type { Router } from 'vue-router'
import type { AppRouteMode } from './index'
import { useAppStore, useRouteStore, useTabStore } from '@/store'
import { i18n } from '@/modules/i18n'
import { local } from '@/utils'
import { buildAdminEntryUrl, getAdminBasePath } from './constants'

async function loadAdminRoutesDynamic() {
  try {
    const { getAdminRoutes } = await import(
      /* webpackChunkName: "admin-core" */
      '@/router/admin.routes'
    )
    const { createAdminMenus } = await import('@/store/router/helper')
    return { getAdminRoutes, createAdminMenus }
  }
  catch (error) {
    console.error('[Security] Failed to load admin routes dynamically:', error)
    return null
  }
}

const ADMIN_PUBLIC_PATHS = new Set(['/login', '/user/login', '/403', '/404', '/500', '/loading', '/public'])

function isAdminRoutePath(path: string, mode: AppRouteMode, adminPath: string) {
  if (mode === 'admin')
    return !ADMIN_PUBLIC_PATHS.has(path)
  return path === adminPath || path.startsWith(`${adminPath}/`)
}

function isI18nKey(key: string) {
  // 仅对形如 "module.key" 的 key 做翻译，避免把中文标题当 key 触发 missing 警告
  return key.includes('.') && /^[A-Za-z0-9_.-]+$/.test(key)
}

function resolveI18nText(t: (key: string) => string, key: unknown) {
  if (!key || typeof key !== 'string')
    return ''
  if (!isI18nKey(key))
    return key
  const text = t(key)
  return text === key ? key : text
}

export function setupRouterGuard(router: Router, mode: AppRouteMode = 'user') {
  const appStore = useAppStore()
  const routeStore = useRouteStore()
  const tabStore = useTabStore()

  router.beforeEach(async (to, _from, next) => {
    const adminPath = getAdminBasePath()
    const isAdminRoute = isAdminRoutePath(to.path, mode, adminPath)

    if (to.meta.href) {
      window.open(to.meta.href)
      next(false)
      return
    }

    appStore.showProgress && window.$loadingBar?.start()

    const isLogin = Boolean(local.get('accessToken'))
    const roleValue = local.get('role')
    const roles = Array.isArray(roleValue) ? roleValue : (roleValue ? [roleValue] : [])
    const hasAdminRole = roles.includes('admin')

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
        next({ path: '/user/login', query: { redirect: to.fullPath } })
      }
      return
    }

    if (to.name !== 'login' && to.name !== 'register' && to.meta.requiresAuth !== false && to.meta.requiresAuth === true && !isLogin) {
      const redirect = to.name === '404' ? undefined : to.fullPath
      next({ path: '/user/login', query: { redirect } })
      return
    }

    if (!routeStore.isInitAuthRoute && to.name !== 'login') {
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
        next({ path: '/user/login', query: redirect ? { redirect } : undefined })
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
              ;(routeStore as any).adminMenus = adminModule.createAdminMenus(adminRoutes)
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
          console.error('[Security] Failed to load admin routes in guard:', error)
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

    if ((to.name === 'login' || to.name === 'register') && isLogin) {
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
