import type { RouteRecordRaw } from 'vue-router'

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

/**
 * 获取管理端路由配置
 * 管理端只使用 hash 内部路由（/#/dashboard）
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
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: () => import('@/views/admin/dashboard/index.vue'),
          meta: {
            title: '仪表盘',
            icon: 'icon-park-outline:dashboard',
          },
        },
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
        {
          path: 'email-templates',
          name: 'admin-email-templates',
          component: () => import('@/views/admin/email-templates/index.vue'),
          meta: {
            title: '邮件模板',
            icon: 'icon-park-outline:mail',
          },
        },
        {
          path: 'logs',
          name: 'admin-logs',
          component: () => import('@/views/admin/logs/index.vue'),
          meta: {
            title: '操作日志',
            icon: 'icon-park-outline:log',
          },
        },
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
