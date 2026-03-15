import { defineStore } from 'pinia'
import { unref } from 'vue'
import { router } from '@/router'
import { buildAdminEntryUrl, getAdminBasePath } from '@/router/constants'
import { getRuntimeRouteMode } from '@/router/runtime-mode'
import { fetchLogin, fetchUpdateToken, fetchUserSettings } from '@/service'
import { authStorage, langToFrontendFormat } from '@/utils'
import { useRouteStore } from './router'
import { useTabStore } from './tab'

interface AuthStatus {
  userInfo: Api.Login.Info | null
  token: string
  accessTokenExpiresAt: number | null
  refreshTimer: ReturnType<typeof setTimeout> | null
}
export const useAuthStore = defineStore('auth-store', {
  state: (): AuthStatus => {
    return {
      userInfo: authStorage.get('userInfo'),
      token: authStorage.get('accessToken') || '',
      accessTokenExpiresAt: authStorage.get('accessTokenExpiresAt') || null,
      refreshTimer: null,
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
      } else {
        this.userInfo = info as Api.Login.Info
      }
      authStorage.setActive('userInfo', this.userInfo)
    },

    /* 登录退出，重置用户信息等 */
    async logout() {
      // 清除自动刷新定时器
      this.clearRefreshTimer()
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
      // 始终重定向到首页
      router.push({
        path: '/',
      })
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

    /* 用户登录 */
    async login(userName: string, password: string) {
      try {
        const mode = getRuntimeRouteMode()
        const authGuard = mode === 'admin' ? 'admin' : 'user'
        const { isSuccess, data } = await fetchLogin({ userName, password, authGuard })
        if (!isSuccess)
          return

        // 处理登录信息
        await this.handleLoginInfo(data as any)
      }
      catch (e) {
        console.warn('[Login Error]:', e)
      }
    },

    /* 处理登录返回的数据 */
    async handleLoginInfo(data: Api.Login.Info & { expiresAt?: number }) {
      // 将token和userInfo保存下来
      authStorage.setActive('userInfo', data)
      authStorage.setActive('accessToken', data.accessToken)
      authStorage.setActive('refreshToken', data.refreshToken)
      authStorage.setActive('role', data.role?.length ? data.role : ['user'])

      const isAdmin = data.role.includes('admin')

      if (data.expiresAt) {
        authStorage.setActive('accessTokenExpiresAt', data.expiresAt)
        this.accessTokenExpiresAt = data.expiresAt
      }

      this.token = data.accessToken
      this.userInfo = data

      // 添加路由和菜单
      const routeStore = useRouteStore()
      const routeMode = getRuntimeRouteMode()
      await routeStore.initAuthRoute(routeMode)

      // 进行重定向跳转
      const route = unref(router.currentRoute)
      const query = route.query as { redirect: string }
      const redirectPath = query.redirect || '/'

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
      } catch {}
    },

    /**
     * 设置自动刷新 Token 定时器
     */
    setupAutoRefresh() {
      const autoRefresh = import.meta.env.VITE_AUTO_REFRESH_TOKEN === 'Y'
      if (!autoRefresh) return

      this.clearRefreshTimer()

      const expiresAt = this.accessTokenExpiresAt || authStorage.get('accessTokenExpiresAt')
      const refreshToken = authStorage.get('refreshToken')
      if (!expiresAt || !refreshToken) return

      const aheadSeconds = Number(import.meta.env.VITE_TOKEN_REFRESH_AHEAD || 60)
      const now = Math.floor(Date.now() / 1000)
      const delaySeconds = expiresAt - now - aheadSeconds

      console.log(`[Auth] Token 将在 ${delaySeconds} 秒后尝试自动刷新`)

      if (delaySeconds <= 0) {
        // 已经到期或即将到期，立即刷新
        this.refreshTokenSilently()
      } else {
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
      try {
        const refreshToken = authStorage.get('refreshToken')
        if (!refreshToken) return

        const mode = getRuntimeRouteMode()
        const authGuard = mode === 'admin' ? 'admin' : 'user'

        const { isSuccess, data } = await fetchUpdateToken({ refreshToken, authGuard })
        if (!isSuccess) {
          console.warn('[Auth] 自动刷新 Token 失败，可能是 refresh token 已过期')
          return
        }

        // 更新存储
        authStorage.setActive('accessToken', data.accessToken)
        authStorage.setActive('refreshToken', data.refreshToken)
        if ((data as any).expiresAt) {
          authStorage.setActive('accessTokenExpiresAt', (data as any).expiresAt)
          this.accessTokenExpiresAt = (data as any).expiresAt
        }

        this.token = data.accessToken
        console.log('[Auth] Token 自动刷新成功')

        // 安排下一次刷新
        this.setupAutoRefresh()
      } catch (error) {
        console.error('[Auth] 自动刷新 Token 异常:', error)
      }
    }
  },
})
