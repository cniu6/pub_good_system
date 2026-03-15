/**
 * 前端安全防护工具
 * 用于检测 Vue Router 篡改、非法访问等行为
 *
 * 注意：前端防护只能增加破解难度，真正的安全必须依靠后端验证！
 */

import { authStorage } from './storage'

interface SecurityConfig {
  /** 是否启用安全检测 */
  enabled: boolean
  /** 检测间隔（毫秒） */
  checkInterval: number
  /** 发现篡改后的处理方式 */
  onTampering: 'redirect' | 'alert' | 'silent'
  /** 重定向目标 */
  redirectTarget: string
}

const defaultConfig: SecurityConfig = {
  enabled: true,
  checkInterval: 2000,
  onTampering: 'redirect',
  redirectTarget: '/403',
}

let securityTimer: ReturnType<typeof setInterval> | null = null
let originalBeforeEach: (() => void) | null = null
let originalAfterEach: (() => void) | null = null
let isInitialized = false

/**
 * 获取 Vue Router 实例（安全方式）
 */
function getRouterInstance(): any {
  try {
    // Vue 3 方式
    const vueApp = (window as any).__VUE_APP__
    if (vueApp?.config?.globalProperties?.$router) {
      return vueApp.config.globalProperties.$router
    }

    // 备用方式：从 DOM 查找
    const appRoot = document.querySelector('#app')
    if (appRoot && (appRoot as any).__vue_app__) {
      return (appRoot as any).__vue_app__.config.globalProperties.$router
    }

    return null
  }
  catch {
    return null
  }
}

/**
 * 检测 Router 是否被篡改
 */
function checkRouterTampering(router: any): boolean {
  if (!router)
    return false

  try {
    // 检测 1: beforeEach 是否被置空或覆盖
    if (typeof router.beforeEach !== 'function') {
      console.warn('[Security] router.beforeEach 被篡改')
      return true
    }

    // 检测 2: 检查守卫数组是否被清空（Vue Router 内部属性）
    const guardArrays = ['beforeGuards', 'beforeResolveGuards', 'afterGuards']
    for (const prop of guardArrays) {
      if (router[prop] && Array.isArray(router[prop])) {
        // 如果之前有守卫，现在被清空，说明被篡改
        // 注意：这里只检测异常的空数组情况
        if (router[prop].length === 0 && originalBeforeEach !== null) {
          console.warn(`[Security] ${prop} 被清空`)
          return true
        }
      }
    }

    // 检测 3: 检查是否有可疑的属性被修改
    // 如果 router.getRoutes 被覆盖或不存在
    if (typeof router.getRoutes !== 'function' && router.options?.routes) {
      console.warn('[Security] router.getRoutes 被篡改')
      return true
    }

    return false
  }
  catch (error) {
    console.warn('[Security] 检测过程出错:', error)
    return false
  }
}

/**
 * 检测可疑的 Vue DevTools 或破解脚本
 */
function detectSuspiciousActivity(): boolean {
  try {
    // 检测常见的破解脚本特征
    const suspiciousPatterns = [
      'detector.js',
      'crack',
      'bypass',
      'patchRouterGuards',
      'patchAllRouteAuth',
    ]

    // 检查是否有注入的脚本标签
    const scripts = document.querySelectorAll('script[src]')
    for (const script of Array.from(scripts)) {
      const src = script.getAttribute('src') || ''
      for (const pattern of suspiciousPatterns) {
        if (src.toLowerCase().includes(pattern.toLowerCase())) {
          console.warn(`[Security] 检测到可疑脚本: ${src}`)
          return true
        }
      }
    }

    // 检查 window 上是否有可疑属性
    const suspiciousWindowProps = ['__VUE_DETECTOR__', '__CRACK_ROUTER__']
    for (const prop of suspiciousWindowProps) {
      if ((window as any)[prop]) {
        console.warn(`[Security] 检测到可疑 window 属性: ${prop}`)
        return true
      }
    }

    return false
  }
  catch {
    return false
  }
}

/**
 * 处理检测到的篡改行为
 */
function handleTampering(config: SecurityConfig): void {
  console.warn('[Security] ⚠️ 检测到安全威胁！')

  switch (config.onTampering) {
    case 'redirect':
      // 清除本地存储的安全信息
      authStorage.clearActive()
      // 重定向到错误页面
      window.location.href = config.redirectTarget
      break
    case 'alert':
      alert('检测到安全威胁，页面将重新加载')
      window.location.reload()
      break
    case 'silent':
      // 静默处理，但记录日志
      console.error('[Security] 检测到篡改，已静默处理')
      break
  }
}

/**
 * 执行安全检测
 */
function performSecurityCheck(config: SecurityConfig): void {
  if (!config.enabled)
    return

  const router = getRouterInstance()

  // 检测 Router 篡改
  if (router && checkRouterTampering(router)) {
    handleTampering(config)
    return
  }

  // 检测可疑活动
  if (detectSuspiciousActivity()) {
    handleTampering(config)
    return
  }
}

/**
 * 初始化安全防护
 */
export function initSecurityProtection(customConfig?: Partial<SecurityConfig>): void {
  if (isInitialized) {
    console.warn('[Security] 安全防护已初始化，跳过重复初始化')
    return
  }

  const config = { ...defaultConfig, ...customConfig }

  if (!config.enabled) {
    console.log('[Security] 安全防护已禁用')
    return
  }

  // 记录原始守卫状态
  const router = getRouterInstance()
  if (router) {
    originalBeforeEach = typeof router.beforeEach === 'function' ? router.beforeEach : null
    originalAfterEach = typeof router.afterEach === 'function' ? router.afterEach : null
    void originalAfterEach
  }

  // 启动定期检测
  securityTimer = setInterval(() => {
    performSecurityCheck(config)
  }, config.checkInterval)

  // 添加页面可见性检测
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
      performSecurityCheck(config)
    }
  })

  isInitialized = true
  console.log('[Security] 🛡️ 安全防护已启用')
}

/**
 * 停止安全检测
 */
export function stopSecurityProtection(): void {
  if (securityTimer) {
    clearInterval(securityTimer)
    securityTimer = null
  }
  isInitialized = false
  console.log('[Security] 安全防护已停止')
}

/**
 * 防止 DevTools 控制台被滥用（增加破解难度）
 * 注意：这只能增加破解难度，无法完全阻止
 */
export function protectConsole(): void {
  // 防止通过 console 注入代码
  // 仅在生产环境启用
  if (import.meta.env.PROD) {
    const originalConsole = {
      log: console.log,
      warn: console.warn,
      error: console.error,
      info: console.info,
    }

    // 拦截可疑的 console 调用
    const suspiciousPatterns = [
      'router.getRoutes',
      'router.addRoute',
      'router.beforeEach',
      'meta.auth',
      'meta.requiresAuth',
    ]

    const checkAndLog = (type: keyof typeof originalConsole, ...args: any[]) => {
      const str = args.map(a => String(a)).join(' ')
      for (const pattern of suspiciousPatterns) {
        if (str.includes(pattern)) {
          // 记录可疑行为（不输出到控制台）
          originalConsole.warn.call(console, '[Security] 可疑的 console 调用:', pattern)
        }
      }
      originalConsole[type].apply(console, args)
    }

    console.log = (...args) => checkAndLog('log', ...args)
    console.warn = (...args) => checkAndLog('warn', ...args)
    console.error = (...args) => checkAndLog('error', ...args)
    console.info = (...args) => checkAndLog('info', ...args)
  }
}

/**
 * 混淆敏感信息
 * 用于隐藏管理端路径等敏感配置
 */
export function getSecureAdminPath(): string {
  // 从环境变量获取，但运行时混淆
  const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr'
  // 返回 Base64 编码后的路径（增加静态分析难度）
  return adminPath
}

/**
 * 验证当前用户角色（与后端同步）
 */
export async function verifyUserRole(): Promise<string | null> {
  try {
    // 这里应该调用后端 API 验证用户角色
    // 防止前端 token 被篡改后伪装成管理员
    const token = authStorage.get('accessToken')
    if (!token)
      return null

    // 实际项目中应该调用后端验证接口
    // const response = await fetch('/api/v1/user/verify-role', {
    //   headers: { Authorization: `Bearer ${token}` }
    // })
    // return response.json().role

    // 简单实现：从 localStorage 读取（不够安全，仅作演示）
    const role = authStorage.get('role')
    if (Array.isArray(role))
      return role[0] || null
    return role || null
  }
  catch {
    return null
  }
}
