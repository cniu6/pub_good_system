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
  // 充值中心
  // ----------------------------------------
  {
    name: 'recharge',
    path: '/user/recharge',
    title: '充值中心',
    requiresAuth: true,
    icon: 'icon-park-outline:add-one',
    componentPath: '/user/recharge/index.vue',
    id: 10,
    pid: null,
  },

  // ----------------------------------------
  // 余额与积分
  // ----------------------------------------
  {
    name: 'moneyScoreLogs',
    path: '/user/money-score-logs',
    title: '余额与积分',
    requiresAuth: true,
    icon: 'icon-park-outline:wallet',
    componentPath: '/user/money-score-logs/index.vue',
    id: 9,
    pid: null,
  },

  // ----------------------------------------
  // 个人中心（侧边栏最后一项）
  // ----------------------------------------
  {
    name: 'userCenter',
    path: '/user/user-center',
    title: 'route.userCenter',
    requiresAuth: true,
    icon: 'carbon:user-avatar-filled-alt',
    componentPath: '/user/user-center/index.vue',
    id: 999,
    pid: null,
  },

  // ----------------------------------------
  // 关于（隐藏菜单，侧边栏底部小链接）
  // ----------------------------------------
  {
    name: 'about',
    path: '/user/about',
    title: 'route.about',
    requiresAuth: true,
    hide: true,
    icon: 'icon-park-outline:info',
    componentPath: '/demo/about/index.vue',
    id: 8,
    pid: null,
  },
]
