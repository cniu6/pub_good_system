import type { MenuOption } from 'naive-ui'
import type { RouteRecordNameGeneric, RouteRecordRaw } from 'vue-router'
import Layout from '@/layouts/index.vue'
import { $t, arrayToTree, renderIcon } from '@/utils'
import { clone, min, omit, pick } from 'radash'
import { h } from 'vue'
import { RouterLink, RouterView } from 'vue-router'

/** 目录路由透传组件：对齐管理端，保证嵌套子路由在侧边栏切换时 Content 层始终有可渲染组件 */
const PassThrough = { name: 'RoutePassThrough', render: () => h(RouterView) }

export interface AdminMenuRoute extends Omit<RouteRecordRaw, 'children' | 'meta' | 'name'> {
  name?: RouteRecordNameGeneric
  meta?: Partial<AppRoute.RouteMeta>
  children?: AdminMenuRoute[]
}

function resolveRouteDisplayText(name?: RouteRecordNameGeneric, title?: string) {
  const routeName = typeof name === 'string' ? name : ''
  const fallback = title || routeName
  return routeName ? $t(`route.${routeName}`, fallback) : fallback
}

function joinRoutePath(basePath: string, path: string) {
  if (path.startsWith('/'))
    return path

  const normalizedBase = basePath.endsWith('/') ? basePath.slice(0, -1) : basePath
  return `${normalizedBase}/${path}`.replace(/\/+/g, '/')
}

const metaFields: AppRoute.MetaKeys[]
  = ['title', 'icon', 'requiresAuth', 'roles', 'keepAlive', 'hide', 'order', 'href', 'activeMenu', 'withoutTab', 'pinTab', 'menuType']

function standardizedRoutes(route: AppRoute.RowRoute[]) {
  const clonedRoutes = clone(route) as AppRoute.RowRoute[]
  return clonedRoutes.map((item: AppRoute.RowRoute) => {
    const route = omit(item, metaFields)

    Reflect.set(route, 'meta', pick(item, metaFields))
    return route
  }) as AppRoute.Route[]
}

export function createRoutes(routes: AppRoute.RowRoute[]) {
  // Structure the meta field
  let resultRouter = standardizedRoutes(routes)

  // Generate routes, no need to import files for those with redirect
  // 排除 admin 目录：管理端页面统一由 admin.routes.ts 用直接 import() 静态声明并单独分包，
  // 这份 glob 只服务 user/静态/动态路由的 componentPath 映射，收窄范围避免管理端 chunk
  // 被此处的全量扫描连带打进用户端可解析的依赖图，破坏管理端/用户端代码隔离。
  const modules = import.meta.glob(['@/views/**/*.vue', '!@/views/admin/**'])
  resultRouter = resultRouter.map((item: AppRoute.Route) => {
    if (item.componentPath && !item.redirect)
      item.component = modules[`/src/views${item.componentPath}`]
    // 目录节点无页面组件时补 PassThrough，避免跨分组切换时 router-view 的 Component 为空导致空白
    if (item.meta?.menuType === 'dir' && !item.component)
      item.component = PassThrough
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
          target = min(orderChilds, (child: AppRoute.Route) => child.meta.order!) as AppRoute.Route

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
export function createAdminMenus(adminRoutes: AdminMenuRoute[]): MenuOption[] {
  const menus: MenuOption[] = []

  for (const route of adminRoutes) {
    if (!route.children)
      continue

    for (const child of route.children) {
      if (child.meta?.hide)
        continue

      const basePath = route.path

      // 目录类型：生成带 children 的分组菜单
      if (child.meta?.menuType === 'dir' && child.children?.length) {
        const dirPath = joinRoutePath(basePath, child.path)
        const subMenus: MenuOption[] = child.children
          .filter(sub => !sub.meta?.hide)
          .map((sub) => {
            const fullPath = joinRoutePath(dirPath, sub.path)
            return {
              label: () => h(RouterLink, { to: { path: fullPath } }, { default: () => resolveRouteDisplayText(sub.name, sub.meta?.title) }),
              key: fullPath,
              icon: typeof sub.meta?.icon === 'string' ? renderIcon(sub.meta.icon) : undefined,
            }
          })

        menus.push({
          label: () => resolveRouteDisplayText(child.name, child.meta?.title),
          key: dirPath,
          icon: typeof child.meta?.icon === 'string' ? renderIcon(child.meta.icon) : undefined,
          children: subMenus,
        })
      }
      else {
        // 普通页面：生成可点击的菜单项
        const fullPath = joinRoutePath(basePath, child.path)
        menus.push({
          label: () => h(RouterLink, { to: { path: fullPath } }, { default: () => resolveRouteDisplayText(child.name, child.meta?.title) }),
          key: fullPath,
          icon: typeof child.meta?.icon === 'string' ? renderIcon(child.meta.icon) : undefined,
        })
      }
    }
  }

  return menus
}
