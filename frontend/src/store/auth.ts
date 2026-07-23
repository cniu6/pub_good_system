import { defineStore } from 'pinia'
import { unref } from 'vue'
import { markSessionExpired, resetSessionExpired } from '@/service/http/auth-expiration'
import { refreshAuthToken } from '@/service/http/token-refresh'
import { router } from '@/router'
import { buildAdminEntryUrl, getAdminBasePath } from '@/router/constants'
import { getRuntimeRouteMode } from '@/router/runtime-mode'
import { fetchLogin, fetchUserSettings, logoutCurrentSession } from '@/service'
import { forceLogoutSession } from '@/service/api/session'
import { startPresence, stopPresence } from '@/composables/usePresence'
import { $t, authStorage, langToFrontendFormat } from '@/utils'
import { useRouteStore } from './router'
import { useTabStore } from './tab'
import { useSettingsStore } from './settings'

type LoginInfoPayload = Api.Login.Info & { expiresAt?: number }

/**
 * 校验登录后 redirect 是否为安全的站内相对路径。
 * 拒绝 //evil.com、http(s)://、反斜杠等协议相对/绝对外链，防止 open redirect。
 */
function isSafeInternalRedirect(path: string): boolean {
  if (!path || typeof path !== 'string')
    return false
  const trimmed = path.trim()
  if (!trimmed.startsWith('/'))
    return false
  if (trimmed.startsWith('//') || trimmed.startsWith('/\\'))
    return false
  if (trimmed.includes('://'))
    return false
  for (let i = 0; i < trimmed.length; i++) {
    const code = trimmed.charCodeAt(i)
    if ((code >= 0 && code <= 31) || code === 127)
      return false
  }
  return true
}

/** 规范化后端/本地 authGuard；非法值返回 null。 */
function normalizeAuthGuard(value: unknown): Entity.AuthGuardType | null {
  if (value === 'admin' || value === 'user')
    return value
  return null
}

/**
 * 从登录/刷新载荷解析 authGuard。
 * 优先用后端显式字段；缺失时（旧会话/过渡期）回退到当前页面模式，避免强制老用户重新登录。
 */
function resolveAuthGuardFromPayload(data: Partial<LoginInfoPayload>): Entity.AuthGuardType {
  return normalizeAuthGuard(data.authGuard) ?? (getRuntimeRouteMode() === 'admin' ? 'admin' : 'user')
}

interface AuthStatus {
  userInfo: Api.Login.Info | null
  token: string
  accessTokenExpiresAt: number | null
  /** 本次会话 JWT auth_guard，刷新与 Presence 以此为准 */
  authGuard: Entity.AuthGuardType | null
  refreshTimer: ReturnType<typeof setTimeout> | null
  isLoggingOut: boolean
  /** 被动失效后显示全局登录恢复弹窗，保持当前页面不跳转。 */
  needsReauthentication: boolean
  /** 触发被动失效前的账号 ID，用于避免新账号看到旧页面中的敏感数据。 */
  reauthenticationUserId: number | null
  /**
   * 会话代际计数器：每次登出/登录都会递增。
   * 用于丢弃「登出前已发出、登出后才返回」的过期 token 刷新结果，
   * 避免刷新竟态把已登出的会话重新救活（重新写回 storage / 重启定时器）。
   */
  authGeneration: number
}
export const useAuthStore = defineStore('auth-store', {
  state: (): AuthStatus => {
    return {
      userInfo: authStorage.get('userInfo'),
      token: authStorage.get('accessToken') || '',
      accessTokenExpiresAt: authStorage.get('accessTokenExpiresAt') || null,
      authGuard: normalizeAuthGuard(authStorage.get('authGuard')),
      refreshTimer: null,
      isLoggingOut: false,
      needsReauthentication: false,
      reauthenticationUserId: null,
      authGeneration: 0,
    }
  },
  getters: {
    /** 是否登录 */
    isLogin(state) {
      return Boolean(state.token)
    },
  },
  actions: {
    /** 更新本地用户信息 */
    updateUserInfo(info: Partial<Api.Login.Info>) {
      if (this.userInfo) {
        this.userInfo = { ...this.userInfo, ...info }
      }
      else {
        this.userInfo = info as Api.Login.Info
      }
      authStorage.setActive('userInfo', this.userInfo)
    },

    /* 登录退出，重置用户信息等 */
    async logout(revokeRemote = true) {
      if (this.isLoggingOut)
        return
      this.isLoggingOut = true
      // 递增会话代际：让登出前已发出、登出后才 resolve 的旧刷新结果失效，防止把已登出的会话救活
      this.authGeneration += 1
      // 清除自动刷新定时器
      this.clearRefreshTimer()
      // 本地先断开，避免退出过程中仍继续发送 Presence 心跳。
      stopPresence()
      // 撤销远程会话放后台异步执行、不等待结果：
      // 若是因 token 已失效被强制退出，这个请求本身也会 401，
      // 若在这里 await 它，会先卡进一轮 401 处理再回来，导致本地清理和跳转登录页被延迟甚至卡住。
      if (revokeRemote && this.token) {
        // 正常路径：token 仍有效时，走鉴权接口撤销会话 + 顺带踢掉 Presence 连接。
        logoutCurrentSession().catch(() => {
          // 网络中断或会话已失效时忽略即可，不影响本地退出。
        })
        // 兜底路径：token 已过期时上面那个会 401 什么也不做，
        // 这里额外调一个"只验签名不验过期"的公开接口，确保过期会话也能被标记撤销，
        // 不用一直等定时清理任务扫。两边都是各自忽略失败，互不影响。
        forceLogoutSession().catch(() => {})
      }
      // 清除本地缓存
      this.clearAuthStorage()
      // 清空路由、菜单等数据
      const routeStore = useRouteStore()
      routeStore.resetRouteStore()
      // 清空标签栏数据
      const tabStore = useTabStore()
      tabStore.clearAllTabs()
      // 重置当前存储库
      this.$reset()
      // 立即跳转登录页（管理端/用户端共用 /user/login，管理端 hash 路由下也可正常解析）
      router.replace({ path: '/user/login' })
    },
    clearAuthStorage() {
      authStorage.clearActive()
    },
    /**
     * 被动会话失效：停止所有认证能力并打开全局登录恢复弹窗。
     * 不重置路由、标签和页面组件，避免用户正在编辑的数据丢失。
     */
    requireReauthentication() {
      if (this.needsReauthentication || this.isLoggingOut)
        return

      markSessionExpired()
      this.authGeneration += 1
      this.clearRefreshTimer()
      stopPresence()
      this.reauthenticationUserId = this.userInfo?.id ?? null
      this.clearAuthStorage()
      this.userInfo = null
      this.token = ''
      this.accessTokenExpiresAt = null
      this.authGuard = null
      this.needsReauthentication = true
    },
    /**
     * 用户端 localStorage 被其它标签登录/退出后，Pinia 不会自动更新。
     * 此处只从当前 active scope 回填内存态；管理端/login-as 的 sessionStorage 隔离不参与跨标签同步。
     */
    hydrateFromStorage() {
      if (this.isLoggingOut)
        return false

      const accessToken = authStorage.get('accessToken')
      const refreshToken = authStorage.get('refreshToken')
      const userInfo = authStorage.get('userInfo')
      if (!accessToken || !refreshToken || !userInfo)
        return false

      // 旧会话没有 authGuard 字段：按当前页面模式回填一次，避免强制老用户重新登录。
      let authGuard = normalizeAuthGuard(authStorage.get('authGuard'))
        ?? normalizeAuthGuard(userInfo.authGuard)
      if (!authGuard) {
        authGuard = getRuntimeRouteMode() === 'admin' ? 'admin' : 'user'
        authStorage.setActive('authGuard', authGuard)
      }

      this.token = accessToken
      this.userInfo = { ...userInfo, authGuard }
      this.accessTokenExpiresAt = authStorage.get('accessTokenExpiresAt') || null
      this.authGuard = authGuard
      this.needsReauthentication = false
      this.reauthenticationUserId = null
      resetSessionExpired()
      this.startPresence()
      this.setupAutoRefresh()
      return true
    },
    clearRefreshTimer() {
      if (this.refreshTimer) {
        clearTimeout(this.refreshTimer)
        this.refreshTimer = null
      }
    },

    /* 用户登录 */
    async login(userName: string, password: string, options?: { preserveCurrentPage?: boolean }): Promise<{ status: 'ok' | 'fail' }> {
      try {
        const mode = getRuntimeRouteMode()
        const authGuard = mode === 'admin' ? 'admin' : 'user'
        const result = await fetchLogin({ userName, password, authGuard })
        const { isSuccess, data } = result
        const loginData = data as LoginInfoPayload | undefined
        if (!isSuccess || !loginData) {
          const tip = typeof (result as { message?: string }).message === 'string'
            && (result as { message?: string }).message
            ? (result as { message: string }).message
            : $t('login.loginFailed')
          window.$message?.error(tip)
          return { status: 'fail' }
        }

        await this.handleLoginInfo(loginData, options)
        return { status: 'ok' }
      }
      catch (error: unknown) {
        const tip = error && typeof error === 'object' && 'message' in error
          && typeof (error as { message?: unknown }).message === 'string'
          && (error as { message: string }).message
          ? (error as { message: string }).message
          : $t('login.loginFailed')
        window.$message?.error(tip)
        return { status: 'fail' }
      }
    },

    /* 处理登录返回的数据 */
    async handleLoginInfo(data: LoginInfoPayload, options?: { preserveCurrentPage?: boolean }) {
      // 递增会话代际：避免上一个会话遗留的旧刷新请求晚于本次登录 resolve 后覆盖新会话
      this.authGeneration += 1
      // 新会话已建立，解除旧会话过期期间对并发请求的保护状态。
      resetSessionExpired()
      // 与后端对齐：仅 admin/user；历史 super 视为 admin
      const rawRoles: string[] = Array.isArray(data.role) && data.role.length ? data.role as string[] : ['user']
      const roles: Entity.RoleType[] = rawRoles.map(r => (r === 'admin' || r === 'super' ? 'admin' : 'user'))
      const authGuard = resolveAuthGuardFromPayload(data)
      const userInfo: LoginInfoPayload = { ...data, role: roles, authGuard }

      // 将token和userInfo保存下来
      authStorage.setActive('userInfo', userInfo)
      authStorage.setActive('accessToken', userInfo.accessToken)
      authStorage.setActive('refreshToken', userInfo.refreshToken)
      authStorage.setActive('role', roles)
      authStorage.setActive('authGuard', authGuard)

      const isAdmin = authGuard === 'admin' || roles.includes('admin')
      const routeMode = getRuntimeRouteMode()
      const isSameUser = this.reauthenticationUserId === userInfo.id
      // 只有同一账号且具备当前端权限时，才允许保留旧页面，避免泄露旧账号数据。
      const preserveCurrentPage = options?.preserveCurrentPage
        && this.needsReauthentication
        && isSameUser
        && (routeMode !== 'admin' || authGuard === 'admin')

      if (userInfo.expiresAt) {
        authStorage.setActive('accessTokenExpiresAt', userInfo.expiresAt)
        this.accessTokenExpiresAt = userInfo.expiresAt
      }

      this.token = userInfo.accessToken
      this.userInfo = userInfo
      this.authGuard = authGuard
      this.needsReauthentication = false
      this.reauthenticationUserId = null
      this.startPresence()

      // 添加路由和菜单
      const routeStore = useRouteStore()
      if (options?.preserveCurrentPage && !preserveCurrentPage) {
        routeStore.resetRouteStore()
        const tabStore = useTabStore()
        tabStore.clearAllTabs()
      }
      await routeStore.initAuthRoute(routeMode)

      if (preserveCurrentPage) {
        this.restoreLanguageFromBackend()
        this.setupAutoRefresh()
        return
      }

      // 进行重定向跳转（仅允许站内相对路径，防 open redirect）
      const route = unref(router.currentRoute)
      const query = route.query as { redirect: string }
      const rawRedirect = typeof query.redirect === 'string' ? query.redirect : ''
      const redirectPath = isSafeInternalRedirect(rawRedirect) ? rawRedirect : '/'

      // 如果重定向路径是根路径，且用户是管理员，可以重定向到管理端
      // 否则重定向到主页
      if (redirectPath === '/' && isAdmin) {
        const adminPath = getAdminBasePath()
        if (routeMode === 'user') {
          window.location.replace(buildAdminEntryUrl(adminPath))
        }
        else {
          router.push({ path: '/dashboard' })
        }
      }
      else if (redirectPath === '/') {
        router.push({ path: import.meta.env.VITE_HOME_PATH || '/user/dashboard/workbench' })
      }
      else {
        router.push({ path: redirectPath })
      }

      // 从后端恢复用户语言偏好
      this.restoreLanguageFromBackend()

      // 启动自动刷新
      this.setupAutoRefresh()
    },

    applyRefreshedLoginInfo(data: LoginInfoPayload) {
      // 广播消息可能与本页登出交错到达；登出中的会话不能被任何迟到刷新结果重新写回。
      if (this.isLoggingOut || this.needsReauthentication || !data.accessToken || !data.refreshToken)
        return

      const rawRoles: string[] = Array.isArray(data.role) && data.role.length
        ? data.role as string[]
        : ((this.userInfo?.role || ['user']) as string[])
      const roles: Entity.RoleType[] = rawRoles.map(r => (r === 'admin' || r === 'super' ? 'admin' : 'user'))
      const authGuard = resolveAuthGuardFromPayload({
        authGuard: data.authGuard ?? this.authGuard ?? this.userInfo?.authGuard,
      })
      const nextUserInfo: LoginInfoPayload = {
        ...(this.userInfo || {} as LoginInfoPayload),
        ...data,
        role: roles,
        authGuard,
      }

      authStorage.setActive('accessToken', nextUserInfo.accessToken)
      authStorage.setActive('refreshToken', nextUserInfo.refreshToken)
      authStorage.setActive('role', roles)
      authStorage.setActive('authGuard', authGuard)
      if (nextUserInfo.expiresAt) {
        authStorage.setActive('accessTokenExpiresAt', nextUserInfo.expiresAt)
        this.accessTokenExpiresAt = nextUserInfo.expiresAt
      }

      this.token = nextUserInfo.accessToken
      this.userInfo = nextUserInfo
      this.authGuard = authGuard
      authStorage.setActive('userInfo', nextUserInfo)
      this.startPresence()
      this.setupAutoRefresh()
    },

    startPresence() {
      if (!this.token)
        return
      const settingsStore = useSettingsStore()
      // 总开关关闭：确保断开已有连接，且不再取 ticket / 建 WS / 定时 ping
      if (!settingsStore.presenceEnabled) {
        stopPresence()
        return
      }
      // 上报周期取管理端可配置的「在线心跳上报周期」（默认30秒），未加载完成时组合式函数内部兜底为30秒。
      const intervalMs = settingsStore.onlineReportIntervalSeconds * 1000
      const userID = this.userInfo?.id
      if (!userID)
        return
      // Presence 必须跟 JWT auth_guard 一致；缺失时才回退到页面模式。
      const guard = this.authGuard
        ?? normalizeAuthGuard(authStorage.get('authGuard'))
        ?? (getRuntimeRouteMode() === 'admin' ? 'admin' : 'user')
      startPresence(this.token, () => {
        this.requireReauthentication()
      }, userID, guard, intervalMs)
    },

    async restoreLanguageFromBackend() {
      try {
        const res = await fetchUserSettings()
        if (res.isSuccess && res.data?.language) {
          const { useAppStore } = await import('./app')
          const appStore = useAppStore()
          const frontendLang = langToFrontendFormat(res.data.language)
          if (frontendLang !== appStore.lang) {
            appStore.setAppLang(frontendLang)
          }
        }
      }
      catch {
        // do nothing
      }
    },

    /**
     * 设置自动刷新 Token 定时器
     */
    setupAutoRefresh() {
      const autoRefresh = import.meta.env.VITE_AUTO_REFRESH_TOKEN === 'Y'
      if (!autoRefresh)
        return

      this.clearRefreshTimer()

      const expiresAt = this.accessTokenExpiresAt || authStorage.get('accessTokenExpiresAt')
      const refreshToken = authStorage.get('refreshToken')
      if (!expiresAt || !refreshToken)
        return

      const aheadSeconds = Number(import.meta.env.VITE_TOKEN_REFRESH_AHEAD || 60)
      const now = Math.floor(Date.now() / 1000)
      const delaySeconds = expiresAt - now - aheadSeconds

      if (delaySeconds <= 0) {
        // 已经到期或即将到期，立即刷新
        this.refreshTokenSilently()
      }
      else {
        // 开启定时器
        this.refreshTimer = setTimeout(() => {
          this.refreshTokenSilently()
        }, delaySeconds * 1000)
      }
    },

    /**
     * 静默刷新 Token
     */
    async refreshTokenSilently() {
      // 记录发起刷新时的会话代际，resolve 后若代际已变（期间发生了登出/重新登录），
      // 说明这是一次过期结果，直接丢弃，不能再写回 storage 或重启定时器/Presence。
      const generation = this.authGeneration
      try {
        if (!authStorage.get('refreshToken')) {
          this.requireReauthentication()
          return
        }

        const nextLoginInfo = await refreshAuthToken()
        if (this.authGeneration !== generation)
          return
        if (!nextLoginInfo) {
          this.requireReauthentication()
          return
        }

        this.applyRefreshedLoginInfo(nextLoginInfo)
      }
      catch {
        if (this.authGeneration !== generation)
          return
        this.requireReauthentication()
      }
    },
  },
})
