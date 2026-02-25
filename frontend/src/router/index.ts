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
  if (mode === 'admin' && router.hasRoute('root')) {
    // 管理端 hash 入口由 admin-root 接管，避免与静态首页路由冲突
    router.removeRoute('root')
  }
  setupRouterGuard(router, mode)
  app.use(router)
  await router.isReady()
}
