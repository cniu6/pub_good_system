import type { App } from 'vue'
import { createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import { setupRouterGuard } from './guard'
import { routes } from './routes.inner'

const { VITE_ROUTE_MODE = 'hash', VITE_BASE_URL } = import.meta.env
export const router = createRouter({
  // 统一切换为 Hash 模式，解决 History 模式下的 404 问题
  history: createWebHashHistory(VITE_BASE_URL),
  routes,
})
// 安装vue路由
export async function installRouter(app: App) {
  // 添加路由守卫
  setupRouterGuard(router)
  app.use(router)
  await router.isReady() // https://router.vuejs.org/zh/api/index.html#isready
}
