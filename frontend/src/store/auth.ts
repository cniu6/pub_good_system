import { defineStore } from 'pinia'
import { unref } from 'vue'
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
  if (/[\x00-\x1F\x7F]/.test(trimmed))
    return false
  return true
}

interface AuthStatus {
  userInfo: Api.Login.Info | null
  token: string
  accessTokenExpiresAt: number | null
  refreshTimer: ReturnType<typeof setTimeout> | null
  isLoggingOut: boolean
  /**
   * 会话代际计数器：每次登出/登录都会递增。
   * 用于丢弃「登出前已发出、登出后才返回」的过期 token 刷新结果，
   * 避免刷新竟态把已登出的会话重新救活（重新写回 storage / 重启 Presence 与定时器）。
   */
  authGeneration: number
}
export const useAuthStore = defineStore('auth-store', {
  state: (): AuthStatus => {
    return {
      userInfo: authStorage.get('userInfo'),
      token: authStorage.get('accessToken') || '',
      accessTokenExpiresAt: authStorage.get('accessTokenExpiresAt') || null,
      refreshTimer: null,
      isLoggingOut: false,
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
    clearRefreshTimer() {
      if (this.refreshTimer) {
        clearTimeout(this.refreshTimer)
        this.refreshTimer = null
      }
    },

    /* 用户登录；管理端若启用 TOTP，返回 need_totp 而不落正式会话 */
    async login(userName: string, password: string): Promise<{ status: 'ok' | 'need_totp' | 'fail', tempToken?: string }> {
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

        if (loginData.need_totp && loginData.temp_token) {
          return { status: 'need_totp', tempToken: loginData.temp_token }
        }

        await this.handleLoginInfo(loginData)
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

    /** 管理端 TOTP 第二步：用临时令牌 + 动态码换正式会话 */
    async loginWithTotp(tempToken: string, code: string): Promise<boolean> {
      try {
        const { fetchLoginTotp } = await import('@/service/api/user/login')
        const result = await fetchLoginTotp({ temp_token: tempToken, code, clientType: 'web' })
        if (!result.isSuccess || !result.data) {
          window.$message?.error(result.message || $t('login.totpFailed'))
          return false
        }
        await this.handleLoginInfo(result.data as LoginInfoPayload)
        return true
      }
      catch (error: unknown) {
        const tip = error && typeof error === 'object' && 'message' in error
          && typeof (error as { message?: unknown }).message === 'string'
          ? (error as { message: string }).message
          : $t('login.totpFailed')
        window.$message?.error(tip)
        return false
      }
    },

    /* 处理登录返回的数据 */
    async handleLoginInfo(data: LoginInfoPayload) {
      // 递增会话代际：避免上一个会话遗留的旧刷新请求晚于本次登录 resolve 后覆盖新会话
      this.authGeneration += 1
      // 与后端对齐：仅 admin/user；历史 super 视为 admin
      const rawRoles: string[] = Array.isArray(data.role) && data.role.length ? data.role as string[] : ['user']
      const roles: Entity.RoleType[] = rawRoles.map((r) => (r === 'admin' || r === 'super' ? 'admin' : 'user'))
      const userInfo: LoginInfoPayload = { ...data, role: roles }

      // 将token和userInfo保存下来
      authStorage.setActive('userInfo', userInfo)
      authStorage.setActive('accessToken', userInfo.accessToken)
      authStorage.setActive('refreshToken', userInfo.refreshToken)
      authStorage.setActive('role', roles)

      const isAdmin = roles.includes('admin')

      if (userInfo.expiresAt) {
        authStorage.setActive('accessTokenExpiresAt', userInfo.expiresAt)
        this.accessTokenExpiresAt = userInfo.expiresAt
      }

      this.token = userInfo.accessToken
      this.userInfo = userInfo
      this.startPresence()

      // 添加路由和菜单
      const routeStore = useRouteStore()
      const routeMode = getRuntimeRouteMode()
      await routeStore.initAuthRoute(routeMode)

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
      const rawRoles: string[] = Array.isArray(data.role) && data.role.length
        ? data.role as string[]
        : ((this.userInfo?.role || ['user']) as string[])
      const roles: Entity.RoleType[] = rawRoles.map((r) => (r === 'admin' || r === 'super' ? 'admin' : 'user'))
      const nextUserInfo: LoginInfoPayload = {
        ...(this.userInfo || {} as LoginInfoPayload),
        ...data,
        role: roles,
      }

      authStorage.setActive('accessToken', nextUserInfo.accessToken)
      authStorage.setActive('refreshToken', nextUserInfo.refreshToken)
      authStorage.setActive('role', roles)
      if (nextUserInfo.expiresAt) {
        authStorage.setActive('accessTokenExpiresAt', nextUserInfo.expiresAt)
        this.accessTokenExpiresAt = nextUserInfo.expiresAt
      }

      this.token = nextUserInfo.accessToken
      this.userInfo = nextUserInfo
      authStorage.setActive('userInfo', nextUserInfo)
      this.startPresence()
      this.setupAutoRefresh()
    },

    startPresence() {
      if (!this.token)
        return
      // 上报周期取管理端可配置的「在线心跳上报周期」（默认30秒），未加载完成时组合式函数内部兜底为30秒。
      const settingsStore = useSettingsStore()
      const intervalMs = settingsStore.onlineReportIntervalSeconds * 1000
      startPresence(this.token, () => {
        window.$message.warning($t('securityTab.forcedLogout'))
        void this.logout(false)
      }, intervalMs)
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
        const refreshToken = authStorage.get('refreshToken')
        if (!refreshToken) {
          await this.logout()
          return
        }

        const nextLoginInfo = await refreshAuthToken(refreshToken)
        if (this.authGeneration !== generation)
          return
        if (!nextLoginInfo) {
          await this.logout()
          return
        }

        this.applyRefreshedLoginInfo(nextLoginInfo)
      }
      catch {
        if (this.authGeneration !== generation)
          return
        await this.logout()
      }
    },
  },
})
