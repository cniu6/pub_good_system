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
        title: 'route.adminDashboard',
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
            title: 'route.dashboard',
            icon: 'icon-park-outline:dashboard',
          },
        },
        // ── 用户管理（分组目录：用户列表 + 在线用户 + 实名认证）──
        {
          path: 'users',
          name: 'admin-users-group',
          component: PassThrough,
          redirect: { name: 'admin-users' },
          meta: {
            title: 'route.userManagement',
            icon: 'icon-park-outline:user',
            menuType: 'dir',
          },
          children: [
            {
              path: 'list',
              name: 'admin-users',
              component: () => import('@/views/admin/users/index.vue'),
              meta: {
                title: 'route.userList',
                icon: 'icon-park-outline:peoples',
              },
            },
            {
              path: 'online',
              name: 'admin-online-users',
              component: () => import('@/views/admin/online-users/index.vue'),
              meta: {
                title: 'route.onlineUsers',
                icon: 'icon-park-outline:online-meeting',
              },
            },
            {
              path: 'realname',
              name: 'admin-realname',
              component: () => import('@/views/admin/realname/index.vue'),
              meta: {
                title: 'route.realnameVerify',
                icon: 'icon-park-outline:id-card',
              },
            },
          ],
        },
        {
          path: 'users/:id',
          name: 'admin-user-detail',
          component: () => import('@/views/admin/users/detail.vue'),
          meta: {
            title: 'route.userDetail',
            hide: true,
            activeMenu: '/users/list',
          },
        },
        // ── 财务中心（分组目录）──
        {
          path: 'finance',
          name: 'admin-finance',
          component: PassThrough,
          redirect: { name: 'admin-money-logs' },
          meta: {
            title: 'route.financeCenter',
            icon: 'icon-park-outline:finance',
            menuType: 'dir',
          },
          children: [
            {
              path: 'money-logs',
              name: 'admin-money-logs',
              component: () => import('@/views/admin/money-logs/index.vue'),
              meta: {
                title: 'route.moneyLogs',
                icon: 'icon-park-outline:wallet',
              },
            },
            {
              path: 'score-logs',
              name: 'admin-score-logs',
              component: () => import('@/views/admin/score-logs/index.vue'),
              meta: {
                title: 'route.scoreLogs',
                icon: 'icon-park-outline:diamond',
              },
            },
            {
              path: 'pay-gateways',
              name: 'admin-pay-gateways',
              component: () => import('@/views/admin/pay-gateways/index.vue'),
              meta: {
                title: 'route.payGateways',
                icon: 'icon-park-outline:pay-code-one',
              },
            },
            {
              path: 'payment-orders',
              name: 'admin-payment-orders',
              component: () => import('@/views/admin/payment-orders/index.vue'),
              meta: {
                title: 'route.paymentOrders',
                icon: 'icon-park-outline:transaction-order',
              },
            },
            {
              path: 'withdraw',
              name: 'admin-withdraw',
              component: () => import('@/views/admin/withdraw/index.vue'),
              meta: {
                title: 'route.withdrawManagement',
                icon: 'icon-park-outline:wallet-two',
              },
            },
          ],
        },
        // ── 系统设置（分组目录）──
        {
          path: 'settings',
          name: 'admin-settings',
          component: PassThrough,
          redirect: { name: 'admin-settings-config' },
          meta: {
            title: 'route.systemSettings',
            icon: 'icon-park-outline:setting-two',
            menuType: 'dir',
          },
          children: [
            {
              path: 'config',
              name: 'admin-settings-config',
              component: () => import('@/views/admin/settings/index.vue'),
              meta: {
                title: 'route.systemConfig',
                icon: 'icon-park-outline:setting-config',
              },
            },
            {
              path: 'announcements',
              name: 'admin-announcements',
              component: () => import('@/views/admin/announcements/index.vue'),
              meta: {
                title: 'route.announcements',
                icon: 'icon-park-outline:volume-notice',
              },
            },
            {
              // 短信模板已合并进「系统配置 → 短信设置」内层 Tab，此路由仅保留供日志页深链跳转，不在侧边栏展示。
              path: 'sms-templates',
              name: 'admin-sms-templates',
              component: () => import('@/views/admin/sms-templates/index.vue'),
              meta: {
                title: 'route.smsTemplates',
                icon: 'icon-park-outline:message',
                hide: true,
                activeMenu: '/settings/config',
              },
            },
            {
              // 邮件模板已合并进「系统配置 → 邮件设置」内层 Tab，此路由仅保留供日志页深链跳转，不在侧边栏展示。
              path: 'email-templates',
              name: 'admin-email-templates',
              component: () => import('@/views/admin/email-templates/index.vue'),
              meta: {
                title: 'route.emailTemplates',
                icon: 'icon-park-outline:mail',
                hide: true,
                activeMenu: '/settings/config',
              },
            },
            {
              path: 'log-management',
              name: 'admin-log-management',
              component: () => import('@/views/admin/log-management/index.vue'),
              meta: {
                title: 'route.logManagement',
                icon: 'icon-park-outline:log',
              },
            },
            {
              path: 'server-management',
              name: 'admin-server-management',
              component: () => import('@/views/admin/server-management/index.vue'),
              meta: {
                title: 'route.serverManagement',
                icon: 'icon-park-outline:server',
              },
            },
            {
              path: 'auto-jobs',
              name: 'admin-auto-jobs',
              component: () => import('@/views/admin/auto-jobs/index.vue'),
              meta: {
                title: 'route.autoJobManager',
                icon: 'icon-park-outline:timer',
              },
            },
          ],
        },
        {
          path: 'log-management',
          name: 'admin-log-management-legacy',
          redirect: '/settings/log-management',
          meta: {
            title: 'route.logManagement',
            hide: true,
            activeMenu: '/settings/log-management',
          },
        },
        {
          path: 'server-management',
          name: 'admin-server-management-legacy',
          redirect: '/settings/server-management',
          meta: {
            title: 'route.serverManagement',
            hide: true,
            activeMenu: '/settings/server-management',
          },
        },
      ],
    },
  ]
}
