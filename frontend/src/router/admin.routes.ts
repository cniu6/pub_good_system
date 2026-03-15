import type { RouteRecordRaw } from 'vue-router'
import { h } from 'vue'
import { RouterView } from 'vue-router'

/**
 * 管理端路由定义
 * 此文件会被打包到独立的 chunk (admin-core)
 * 只有管理员登录后才会动态加载
 *
 * 安全说明：
 * - 管理端路由通过环境变量 VITE_ADMIN_BASE_PATH 配置路径前缀
 * - 打包时会被分离到独立的 JS chunk，普通用户无法通过查看源码获取路由信息
 * - 只有在验证用户为管理员后才会动态加载这些路由
 */

// 透传组件，用于目录类型路由渲染子路由
const PassThrough = { render: () => h(RouterView) }

/**
 * 获取管理端路由配置
 * 管理端只使用 hash 内部路由（/#/dashboard）
 * 侧边栏采用分组层级结构：仪表盘、用户管理、财务中心（含支付）、系统设置
 */
export function getAdminRoutes(): RouteRecordRaw[] {
  return [
    {
      path: '/',
      name: 'admin-root',
      redirect: '/dashboard',
      component: () => import('@/layouts/index.vue'),
      meta: {
        title: '管理后台',
        requiresAuth: true,
        requiresAdmin: true,
      },
      children: [
        // ── 仪表盘（独立页面）──
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: () => import('@/views/admin/dashboard/index.vue'),
          meta: {
            title: '仪表盘',
            icon: 'icon-park-outline:dashboard',
          },
        },
        // ── 用户管理（独立页面）──
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('@/views/admin/users/index.vue'),
          meta: {
            title: '用户管理',
            icon: 'icon-park-outline:user',
          },
        },
        {
          path: 'users/:id',
          name: 'admin-user-detail',
          component: () => import('@/views/admin/users/detail.vue'),
          meta: {
            title: '用户详情',
            hide: true,
            activeMenu: '/users',
          },
        },
        // ── 财务中心（分组目录）──
        {
          path: 'finance',
          name: 'admin-finance',
          component: PassThrough,
          redirect: { name: 'admin-money-logs' },
          meta: {
            title: '财务中心',
            icon: 'icon-park-outline:finance',
            menuType: 'dir',
          },
          children: [
            {
              path: 'money-logs',
              name: 'admin-money-logs',
              component: () => import('@/views/admin/money-logs/index.vue'),
              meta: {
                title: '余额日志',
                icon: 'icon-park-outline:wallet',
              },
            },
            {
              path: 'score-logs',
              name: 'admin-score-logs',
              component: () => import('@/views/admin/score-logs/index.vue'),
              meta: {
                title: '积分日志',
                icon: 'icon-park-outline:diamond',
              },
            },
            {
              path: 'pay-gateways',
              name: 'admin-pay-gateways',
              component: () => import('@/views/admin/pay-gateways/index.vue'),
              meta: {
                title: '支付通道',
                icon: 'icon-park-outline:pay-code-one',
              },
            },
            {
              path: 'payment-orders',
              name: 'admin-payment-orders',
              component: () => import('@/views/admin/payment-orders/index.vue'),
              meta: {
                title: '支付订单',
                icon: 'icon-park-outline:transaction-order',
              },
            },
          ],
        },
        // ── 系统设置（独立页面）──
        {
          path: 'settings',
          name: 'admin-settings',
          component: () => import('@/views/admin/settings/index.vue'),
          meta: {
            title: '系统设置',
            icon: 'icon-park-outline:setting-two',
          },
        },
      ],
    },
  ]
}
