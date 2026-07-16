// ========================================
// 用户端静态路由（所有路径以 /user 为前缀）
// ========================================
export const staticRoutes: AppRoute.RowRoute[] = [
  // ----------------------------------------
  // 仪表盘
  // ----------------------------------------
  {
    name: 'dashboard',
    path: '/user/dashboard',
    title: 'route.dashboard',
    requiresAuth: true,
    icon: 'icon-park-outline:analysis',
    menuType: 'dir',
    componentPath: null,
    id: 1,
    pid: null,
  },
  {
    name: 'workbench',
    path: '/user/dashboard/workbench',
    title: 'route.workbench',
    requiresAuth: true,
    icon: 'icon-park-outline:alarm',
    pinTab: true,
    menuType: 'page',
    componentPath: '/user/dashboard/workbench/index.vue',
    id: 101,
    pid: 1,
  },
  {
    name: 'monitor',
    path: '/user/dashboard/monitor',
    title: 'route.monitor',
    requiresAuth: true,
    icon: 'icon-park-outline:anchor',
    menuType: 'page',
    componentPath: '/user/dashboard/monitor/index.vue',
    id: 102,
    pid: 1,
  },

  // ----------------------------------------
  // 我的账户（分组目录）
  // ----------------------------------------
  {
    name: 'account',
    path: '/user/account',
    title: 'route.myAccount',
    requiresAuth: true,
    icon: 'icon-park-outline:peoples',
    menuType: 'dir',
    componentPath: null,
    id: 9,
    pid: null,
  },
  {
    name: 'recharge',
    path: '/user/account/recharge',
    title: 'route.rechargeCenter',
    requiresAuth: true,
    icon: 'icon-park-outline:add-one',
    menuType: 'page',
    componentPath: '/user/recharge/index.vue',
    id: 10,
    pid: 9,
  },
  {
    name: 'userCenter',
    path: '/user/account/user-center',
    title: 'route.userCenter',
    requiresAuth: true,
    icon: 'carbon:user-avatar-filled-alt',
    menuType: 'page',
    componentPath: '/user/user-center/index.vue',
    id: 999,
    pid: 9,
  },

]
