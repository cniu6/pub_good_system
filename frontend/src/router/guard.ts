import type { Router } from 'vue-router'
import { useAppStore, useRouteStore, useTabStore } from '@/store'
import { local } from '@/utils'
import { i18n } from '@/modules/i18n'

export function setupRouterGuard(router: Router) {
  const appStore = useAppStore()
  const routeStore = useRouteStore()
  const tabStore = useTabStore()

  router.beforeEach(async (to, from, next) => {
    // 获取管理端自定义路径，默认为 /admin
    const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
    const isAdminRoute = to.path.startsWith(adminPath)

    // 判断是否是外链，如果是直接打开网页并拦截跳转
    if (to.meta.href) {
      window.open(to.meta.href)
      next(false) // 取消当前导航
      return
    }
    // 开始 loadingBar
    appStore.showProgress && window.$loadingBar?.start()

    // 判断有无TOKEN,登录鉴权
    const isLogin = Boolean(local.get('accessToken'))
    const roleValue = local.get('role')
    // 兼容数组格式 {"value":["user"]} 和 字符串格式
    const roles = Array.isArray(roleValue) ? roleValue : (roleValue ? [roleValue] : [])
    const hasAdminRole = roles.includes('admin')

    // 每次跳转前，根据目标路径设置菜单模式 (用户端/管理端)
    routeStore.setMenuMode(to.path)

    // 处理管理端路由访问权限
    if (isAdminRoute && (!isLogin || !hasAdminRole)) {
      console.warn('[Route Guard] 权限不足，拦截管理端访问. Roles:', roles)
      // 已登录但无权限 -> 403；未登录 -> 登录页
      if (isLogin) {
        next({ path: '/403', replace: true })
      }
      else {
        next({ path: '/login', query: { redirect: to.fullPath } })
      }
      return
    }

    // 如果是login路由，直接放行
    if (to.name === 'login') {
      // login页面不需要任何认证检查，直接放行
      // 继续执行后面的逻辑
    }
    // 如果路由明确设置了requiresAuth为false，直接放行
    else if (to.meta.requiresAuth === false) {
      // 明确设置为false的路由直接放行
      // 继续执行后面的逻辑
    }
    // 如果路由设置了requiresAuth为true，且用户未登录，重定向到登录页
    else if (to.meta.requiresAuth === true && !isLogin) {
      const redirect = to.name === '404' ? undefined : to.fullPath
      next({ path: '/login', query: { redirect } })
      return
    }

    // 判断路由有无进行初始化
    if (!routeStore.isInitAuthRoute && to.name !== 'login') {
      try {
        await routeStore.initAuthRoute()
        // 路由初始化完成后，重新导航到目标路由
        // 这样可以确保动态加载的路由能够正确匹配
        console.log('[Route Guard] 路由初始化完成，重新导航到:', to.fullPath)
        next({
          path: to.fullPath,
          replace: true,
          query: to.query,
          hash: to.hash,
        })
        return
      }
      catch {
        // 如果路由初始化失败（比如 401 错误），重定向到登录页
        const redirect = to.fullPath !== '/' ? to.fullPath : undefined
        next({ path: '/login', query: redirect ? { redirect } : undefined })
        return
      }
    }

    // 如果路由已初始化，但是访问的路由不存在（404），检查是否需要重新加载
    if (routeStore.isInitAuthRoute && (to.name === '404' || to.name === 'notFoundCatchAll')) {
      console.log('[Route Guard] 检测到 404，当前路径:', to.fullPath, '路由名称:', to.name)

      // 检查是否是管理端路由
      if (isAdminRoute && hasAdminRole) {
        try {
          // 检查管理端路由是否已加载
          const adminRouteExists = router.hasRoute('admin-root')
          console.log('[Admin Routes] 管理端路由是否存在:', adminRouteExists)

          if (!adminRouteExists) {
            console.log('[Admin Routes] 管理端路由未加载，重新加载')
            // 管理端路由未加载，重新加载
            const { getAdminRoutes } = await import('@/router/admin.routes')
            const { createAdminMenus } = await import('@/store/router/helper')
            const adminRoutes = getAdminRoutes(adminPath)
            adminRoutes.forEach(route => {
              router.addRoute(route)
            })
            // 同步更新管理端侧边栏菜单
            ;(routeStore as any).adminMenus = createAdminMenus(adminRoutes)
            console.log('[Admin Routes] 管理端路由已重新加载，重新导航到:', to.fullPath)

            // 重新导航到目标路由
            next({
              path: to.fullPath,
              replace: true,
              query: to.query,
              hash: to.hash,
            })
            return
          } else {
            // 管理端路由已加载，但仍然 404，可能是路径问题
            console.warn('[Admin Routes] 管理端路由已加载，但仍然 404，路径:', to.fullPath)
          }
        }
        catch (error) {
          console.warn('[Admin Routes] 重新加载管理端路由失败:', error)
        }
      }

      // 检查路由是否真的不存在（排除 404 路由本身）
      if (to.fullPath !== '/404') {
        // 尝试解析路由
        const resolved = router.resolve(to.fullPath)
        console.log('[Route Guard] 路由解析结果:', resolved.name, resolved.path)

        // 如果解析后的路由不是 404，说明路由存在，重新导航
        if (resolved.name && resolved.name !== '404' && resolved.name !== 'notFoundCatchAll') {
          console.log('[Route Guard] 路由存在，重新导航')
          next({
            path: to.fullPath,
            replace: true,
            query: to.query,
            hash: to.hash,
          })
          return
        }
      }

      // 路由确实不存在，显示 404 页面
      console.warn('[Route Guard] 路由不存在，显示 404 页面:', to.fullPath)
      console.log('[Route Guard] 当前所有路由:', router.getRoutes().map(r => ({ name: r.name, path: r.path })))
      next()
      return
    }

    // 如果用户已登录且访问login页面，重定向到首页或管理端
    if (to.name === 'login' && isLogin) {
      // 如果是管理员，重定向到管理端；否则重定向到主页
      if (hasAdminRole) {
        const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
        next({ path: adminPath })
      } else {
        next({ path: import.meta.env.VITE_HOME_PATH || '/' })
      }
      return
    }

    next()
  })
  router.beforeResolve((to) => {
    // 设置菜单高亮
    routeStore.setActiveMenu(to.meta.activeMenu ?? to.fullPath)
    // 添加tabs
    tabStore.addTab(to)
    // 设置高亮标签;
    tabStore.setCurrentTab(to.fullPath as string)
  })

  router.afterEach((to) => {
    // 修改网页标题
    const { t } = i18n.global
    const appTitle = t('app.title') || import.meta.env.VITE_APP_TITLE || import.meta.env.VITE_APP_NAME
    const pageTitle = to.meta.title ? t(to.meta.title as string) : ''
    document.title = pageTitle ? `${pageTitle} - ${appTitle}` : appTitle
    // 结束 loadingBar
    appStore.showProgress && window.$loadingBar?.finish()
  })

  // 监听语言切换事件，更新标题
  window.addEventListener('app:locale-changed', () => {
    const { t } = i18n.global
    const appTitle = t('app.title') || import.meta.env.VITE_APP_TITLE || import.meta.env.VITE_APP_NAME
    const currentRoute = router.currentRoute.value
    const pageTitle = currentRoute.meta.title ? t(currentRoute.meta.title as string) : ''
    document.title = pageTitle ? `${pageTitle} - ${appTitle}` : appTitle
  })
}
