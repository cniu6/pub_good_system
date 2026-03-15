import type { MenuOption } from 'naive-ui'
import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/layouts/index.vue'
import { $t, arrayToTree, renderIcon } from '@/utils'
import { clone, min, omit, pick } from 'radash'
import { RouterLink } from 'vue-router'

const metaFields: AppRoute.MetaKeys[]
  = ['title', 'icon', 'requiresAuth', 'roles', 'keepAlive', 'hide', 'order', 'href', 'activeMenu', 'withoutTab', 'pinTab', 'menuType']

function standardizedRoutes(route: AppRoute.RowRoute[]) {
  return clone(route).map((i) => {
    const route = omit(i, metaFields)

    Reflect.set(route, 'meta', pick(i, metaFields))
    return route
  }) as AppRoute.Route[]
}

export function createRoutes(routes: AppRoute.RowRoute[]) {
  // Structure the meta field
  let resultRouter = standardizedRoutes(routes)

  // Generate routes, no need to import files for those with redirect
  const modules = import.meta.glob('@/views/**/*.vue')
  resultRouter = resultRouter.map((item: AppRoute.Route) => {
    if (item.componentPath && !item.redirect)
      item.component = modules[`/src/views${item.componentPath}`]
    return item
  })

  // Generate route tree
  resultRouter = arrayToTree(resultRouter) as AppRoute.Route[]

  const appRootRoute: RouteRecordRaw = {
    path: '/',
    name: 'appRoot',
    component: Layout,
    meta: {
      title: '',
      icon: 'icon-park-outline:home',
    },
    children: [],
  }

  // Set the correct redirect path for the route
  setRedirect(resultRouter)

  // Insert the processed route into the root route
  appRootRoute.children = resultRouter as unknown as RouteRecordRaw[]
  return appRootRoute
}

// Generate an array of route names that need to be kept alive
export function generateCacheRoutes(routes: AppRoute.RowRoute[]) {
  return routes
    .filter(i => i.keepAlive)
    .map(i => i.name)
}

function setRedirect(routes: AppRoute.Route[]) {
  routes.forEach((route) => {
    if (route.children) {
      if (!route.redirect) {
        // Filter out a collection of child elements that are not hidden
        const visibleChilds = route.children.filter(child => !child.meta.hide)

        // Redirect page to the path of the first child element by default
        let target = visibleChilds[0]

        // Filter out pages with the order attribute
        const orderChilds = visibleChilds.filter(child => child.meta.order)

        if (orderChilds.length > 0)
          target = min(orderChilds, i => i.meta.order!) as AppRoute.Route

        if (target)
          route.redirect = target.path
      }

      setRedirect(route.children)
    }
  })
}

/* 生成侧边菜单的数据 */
export function createMenus(userRoutes: AppRoute.RowRoute[]) {
  const resultMenus = standardizedRoutes(userRoutes)

  // filter menus that do not need to be displayed
  const visibleMenus = resultMenus.filter(route => !route.meta.hide)

  // generate side menu
  return arrayToTree(transformAuthRoutesToMenus(visibleMenus))
}

// render the returned routing table as a sidebar
function transformAuthRoutesToMenus(userRoutes: AppRoute.Route[]) {
  return userRoutes
    //  Sort the menu according to the order size
    .sort((a, b) => {
      if (a.meta && a.meta.order && b.meta && b.meta.order)
        return a.meta.order - b.meta.order
      else if (a.meta && a.meta.order)
        return -1
      else if (b.meta && b.meta.order)
        return 1
      else return 0
    })
    // Convert to side menu data structure
    .map((item) => {
      const target: MenuOption = {
        id: item.id,
        pid: item.pid,
        label:
          (!item.meta.menuType || item.meta.menuType === 'page')
            ? () =>
                h(
                  RouterLink,
                  {
                    to: {
                      path: item.path,
                    },
                  },
                  { default: () => $t(`route.${String(item.name)}`, item.meta.title) },
                )
            : () => $t(`route.${String(item.name)}`, item.meta.title),
        key: item.path,
        icon: item.meta.icon ? renderIcon(item.meta.icon) : undefined,
      }
      return target
    })
}

/**
 * 从 RouteRecordRaw 格式的管理端路由生成侧边栏菜单
 * 支持嵌套层级：menuType === 'dir' 的路由作为分组目录，其子路由作为子菜单项
 */
export function createAdminMenus(adminRoutes: Array<{
  path: string
  children?: Array<{
    path: string
    name?: string | symbol | null
    meta?: {
      hide?: boolean
      title?: string
      icon?: string
      menuType?: string
    }
    children?: Array<{
      path: string
      name?: string | symbol | null
      meta?: {
        hide?: boolean
        title?: string
        icon?: string
      }
    }>
  }>
}>): MenuOption[] {
  const menus: MenuOption[] = []

  for (const route of adminRoutes) {
    if (!route.children) continue

    for (const child of route.children) {
      if (child.meta?.hide) continue

      const basePath = route.path.endsWith('/') ? route.path : `${route.path}/`

      // 目录类型：生成带 children 的分组菜单
      if (child.meta?.menuType === 'dir' && child.children?.length) {
        const dirPath = `${basePath}${child.path}`
        const subMenus: MenuOption[] = child.children
          .filter(sub => !sub.meta?.hide)
          .map((sub) => {
            const fullPath = `${dirPath}/${sub.path}`
            return {
              label: () => h(RouterLink, { to: { path: fullPath } }, { default: () => (sub.meta?.title as string) || sub.name }),
              key: fullPath,
              icon: sub.meta?.icon ? renderIcon(sub.meta.icon as string) : undefined,
            }
          })

        menus.push({
          label: (child.meta?.title as string) || String(child.name),
          key: dirPath,
          icon: child.meta?.icon ? renderIcon(child.meta.icon as string) : undefined,
          children: subMenus,
        })
      }
      else {
        // 普通页面：生成可点击的菜单项
        const fullPath = `${basePath}${child.path}`
        menus.push({
          label: () => h(RouterLink, { to: { path: fullPath } }, { default: () => (child.meta?.title as string) || child.name }),
          key: fullPath,
          icon: child.meta?.icon ? renderIcon(child.meta.icon as string) : undefined,
        })
      }
    }
  }

  return menus
}
