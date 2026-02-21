import type { RouteRecordRaw } from 'vue-router'

/* 页面中的一些固定路由，错误页等 */
export const routes: RouteRecordRaw[] = [
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
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/build-in/login/index.vue'),
    meta: {
      title: 'login.signInTitle',
      withoutTab: true,
    },
  },
  {
    path: '/loading',
    name: 'loading',
    component: () => import('@/components/common/AppLoading.vue'),
    meta: {
      title: '加载中',
      withoutTab: true,
    },
  },
  {
    path: '/public',
    name: 'publicAccess',
    component: () => import('@/views/build-in/public-access/index.vue'),
    meta: {
      title: '公共访问示例',
      requiresAuth: false,
      withoutTab: true,
    },
  },
  {
    path: '/403',
    name: '403',
    component: () => import('@/views/build-in/error/403/index.vue'),
    meta: {
      title: '无权访问',
      icon: 'icon-park-outline:forbidden',
      withoutTab: true,
      requiresAuth: false,
    },
  },
  {
    path: '/500',
    name: '500',
    component: () => import('@/views/build-in/error/500/index.vue'),
    meta: {
      title: '服务器错误',
      icon: 'icon-park-outline:error',
      withoutTab: true,
      requiresAuth: false,
    },
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/build-in/error/404/index.vue'),
    meta: {
      title: '找不到页面',
      icon: 'icon-park-outline:ghost',
      withoutTab: true,
      requiresAuth: false,
    },
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('@/views/build-in/error/404/index.vue'),
    name: 'notFoundCatchAll',
    meta: {
      title: '找不到页面',
      icon: 'icon-park-outline:ghost',
      withoutTab: true,
      requiresAuth: false,
    },
  },
]

