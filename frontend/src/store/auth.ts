import { defineStore } from 'pinia'
import { unref } from 'vue'
import { router } from '@/router'
import { fetchLogin, fetchUpdateToken } from '@/service'
import { local } from '@/utils'
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
      userInfo: local.get('userInfo'),
      token: local.get('accessToken') || '',
      accessTokenExpiresAt: local.get('accessTokenExpiresAt') || null,
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
      local.remove('accessToken')
      local.remove('refreshToken')
      local.remove('userInfo')
      local.remove('role')
      local.remove('accessTokenExpiresAt')
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
        const { isSuccess, data } = await fetchLogin({ userName, password })
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
      local.set('userInfo', data)
      local.set('accessToken', data.accessToken)
      local.set('refreshToken', data.refreshToken)
      local.set('role', (data as any).role || 'user') // 保存角色信息用于路由守卫

      if (data.expiresAt) {
        local.set('accessTokenExpiresAt', data.expiresAt)
        this.accessTokenExpiresAt = data.expiresAt
      }

      this.token = data.accessToken
      this.userInfo = data

      // 添加路由和菜单
      const routeStore = useRouteStore()
      await routeStore.initAuthRoute()

      // 进行重定向跳转
      const route = unref(router.currentRoute)
      const query = route.query as { redirect: string }
      const redirectPath = query.redirect || '/'

      // 如果重定向路径是根路径，且用户是管理员，可以重定向到管理端
      // 否则重定向到主页
      if (redirectPath === '/' && (data as any).role === 'admin') {
        const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr'
        router.push({ path: adminPath })
      } else {
        router.push({ path: redirectPath })
      }

      // 启动自动刷新
      this.setupAutoRefresh()
    },

    /**
     * 设置自动刷新 Token 定时器
     */
    setupAutoRefresh() {
      const autoRefresh = import.meta.env.VITE_AUTO_REFRESH_TOKEN === 'true'
      if (!autoRefresh) return

      this.clearRefreshTimer()

      const expiresAt = this.accessTokenExpiresAt || local.get('accessTokenExpiresAt')
      const refreshToken = local.get('refreshToken')
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
        const refreshToken = local.get('refreshToken')
        if (!refreshToken) return

        const { isSuccess, data } = await fetchUpdateToken({ refreshToken })
        if (!isSuccess) {
          console.warn('[Auth] 自动刷新 Token 失败，可能是 refresh token 已过期')
          return
        }

        // 更新存储
        local.set('accessToken', data.accessToken)
        local.set('refreshToken', data.refreshToken)
        if ((data as any).expiresAt) {
          local.set('accessTokenExpiresAt', (data as any).expiresAt)
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
