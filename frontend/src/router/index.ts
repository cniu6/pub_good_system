import type { App } from 'vue'
import type { Router } from 'vue-router'
import { createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import { getAdminEntryBase, getUserBase } from './constants'
import { setupRouterGuard } from './guard'
import { routes } from './routes.inner'

export type AppRouteMode = 'user' | 'admin'

function createHistory(mode: AppRouteMode) {
  if (mode === 'admin') {
    return createWebHashHistory(getAdminEntryBase())
  }
  return createWebHistory(getUserBase())
}

function createAppRouter(mode: AppRouteMode): Router {
  return createRouter({
    history: createHistory(mode),
    routes,
  })
}

export let router: Router = createAppRouter('user')

export async function installRouter(app: App, mode: AppRouteMode = 'user') {
  router = createAppRouter(mode)
  if (mode === 'admin') {
    // 管理端 hash 入口由 admin-root 接管，移除与管理端冲突的用户端路由
    if (router.hasRoute('root')) router.removeRoute('root')
    if (router.hasRoute('user-redirect')) router.removeRoute('user-redirect')
    if (router.hasRoute('dashboard-redirect')) router.removeRoute('dashboard-redirect')
  }
  setupRouterGuard(router, mode)
  app.use(router)
  await router.isReady()
}
