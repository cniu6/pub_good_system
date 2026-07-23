import type { RouteRecordRaw } from 'vue-router'

/* 页面中的一些固定路由，错误页等 */
export const routes: RouteRecordRaw[] = [
  // ========================================
  // 首页（公开）
  // ========================================
  {
    path: '/',
    name: 'root',
    component: () => import('@/views/index/index.vue'),
    meta: {
      title: 'app.title',
      requiresAuth: false,
      withoutTab: true,
    },
  },

  // ========================================
  // 用户认证页面
  // ========================================
  {
    path: '/user/login',
    name: 'login',
    component: () => import('@/views/_builtin/login/index.vue'),
    meta: {
      title: 'login.signInTitle',
      withoutTab: true,
    },
  },
  {
    path: '/user/register',
    name: 'register',
    component: () => import('@/views/_builtin/login/index.vue'),
    meta: {
      title: 'login.signUpTitle',
      withoutTab: true,
    },
    props: { defaultForm: 'register' },
  },

  // ========================================
  // 便捷重定向
  // ========================================
  {
    path: '/login',
    redirect: '/user/login',
  },
  {
    path: '/register',
    redirect: '/user/register',
  },
  {
    path: '/user',
    name: 'user-redirect',
    redirect: '/user/dashboard/workbench',
  },
  {
    path: '/dashboard',
    name: 'dashboard-redirect',
    redirect: '/user/dashboard/workbench',
  },
  {
    path: '/user/recharge',
    name: 'legacy-user-recharge-redirect',
    redirect: to => ({
      path: '/user/account/recharge',
      query: to.query,
      hash: to.hash,
    }),
  },

  // ========================================
  // 公共 / 工具页面
  // ========================================
  {
    path: '/loading',
    name: 'loading',
    component: () => import('@/components/common/AppLoading.vue'),
    meta: {
      title: 'route.loading',
      withoutTab: true,
    },
  },
  // ========================================
  // 错误页面
  // ========================================
  {
    path: '/403',
    name: '403',
    component: () => import('@/views/_builtin/error/403/index.vue'),
    meta: {
      title: 'route.forbidden',
      icon: 'icon-park-outline:forbidden',
      withoutTab: true,
      requiresAuth: false,
    },
  },
  {
    path: '/500',
    name: '500',
    component: () => import('@/views/_builtin/error/500/index.vue'),
    meta: {
      title: 'route.serverError',
      icon: 'icon-park-outline:error',
      withoutTab: true,
      requiresAuth: false,
    },
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/_builtin/error/404/index.vue'),
    meta: {
      title: 'route.notFound',
      icon: 'icon-park-outline:ghost',
      withoutTab: true,
      requiresAuth: false,
    },
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('@/views/_builtin/error/404/index.vue'),
    name: 'notFoundCatchAll',
    meta: {
      title: 'route.notFound',
      icon: 'icon-park-outline:ghost',
      withoutTab: true,
      requiresAuth: false,
    },
  },
]
