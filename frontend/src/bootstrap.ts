import { createApp } from 'vue'
import type { App as AppInstance } from 'vue'
import { Geetest } from 'vue3-geetest'
import App from './App.vue'
import type { AppRouteMode } from './router'
import { installRouter } from './router'
import { installPinia, useSettingsStore } from './store'
import { install as setupDirectives } from './modules/directives'
import { install as setupI18n } from './modules/i18n'
import { install as setupAssets } from './modules/assets'
import { setupIconifyOffline } from './modules/iconify-offline'
import { setRuntimeRouteMode } from './router/runtime-mode'
import { authStorage } from './utils'
import './styles/index.css'

async function setupApp(app: AppInstance<Element>, mode: AppRouteMode) {
  // 1. 首先安装 Pinia（其他模块依赖它）
  installPinia(app)

  // 1.1 跨标签页同步 token 刷新结果，避免多标签并发轮换 refresh token
  {
    const { initTokenRefreshSync } = await import('./service/http/token-refresh')
    const { useAuthStore } = await import('./store/auth')
    initTokenRefreshSync((payload) => {
      const authStore = useAuthStore()
      // 登出过程的 token 仍可能尚未从 Pinia 清空；此时绝不能接受其它标签的刷新广播，
      // 否则会把刚刚清掉的会话重新写回 storage。
      if (!authStore.isLogin || authStore.isLoggingOut)
        return
      authStore.applyRefreshedLoginInfo(payload)
    })
  }

  // user 模式的普通登录使用 localStorage：其它标签登录/退出时，当前标签的 Pinia 内存态不会自动更新。
  // admin/login-as 用 sessionStorage 隔离，绝不能跟随这里的跨标签同步。
  if (mode === 'user') {
    const { useAuthStore } = await import('./store/auth')
    let authStorageSyncTimer: number | null = null
    window.addEventListener('storage', (event) => {
      if (event.storageArea !== window.localStorage || authStorage.getActiveScope() !== 'local')
        return
      if (authStorageSyncTimer)
        window.clearTimeout(authStorageSyncTimer)
      // 一次登录会连续写入多项 auth 数据，合并到当前事件队列末尾后再读取完整快照。
      authStorageSyncTimer = window.setTimeout(() => {
        authStorageSyncTimer = null
        const authStore = useAuthStore()
        const storedToken = authStorage.get('accessToken')
        if (!storedToken) {
          if (authStore.isLogin)
            authStore.requireReauthentication()
          return
        }
        if (!authStore.isLogin || authStore.token !== storedToken)
          authStore.hydrateFromStorage()
      }, 0)
    })
  }

  // 2. 加载运行时配置（在安装其他模块之前）
  // loadConfig 会注入 admin_api_path，管理端 API 前缀与后端 ADMIN_API_PATH 对齐
  const settingsStore = useSettingsStore()
  await settingsStore.loadConfig()

  // 2.1 本地无语言偏好时，用后端 app-config.default_lang 同步到 appStore（已有 localStorage.lang 不覆盖）
  {
    const { useAppStore } = await import('./store/app')
    const { local } = await import('./utils')
    if (!local.get('lang')) {
      const raw = String(settingsStore.defaultLang || 'zhCN').trim()
      // 后端 default_lang 存的是前端格式 zhCN / enUS
      const resolved: App.lang = (raw === 'enUS' || raw === 'en-US') ? 'enUS' : 'zhCN'
      const appStore = useAppStore()
      if (appStore.lang !== resolved)
        appStore.setAppLang(resolved)
    }
  }

  // 3. 安装其他模块
  setupI18n(app)
  await installRouter(app, mode)
  setupDirectives(app)
  setupAssets()
}

export async function bootstrap(mode: AppRouteMode) {
  try {
    // 图标必须最先注册：侧栏/菜单用字符串图标名，离线集合就绪后才不会空白
    setupIconifyOffline()

    setRuntimeRouteMode(mode)

    // admin 模式下自动启用 sessionStorage 隔离，
    // 避免和普通用户 localStorage 里的 token 互相覆盖
    if (mode === 'admin') {
      authStorage.enableSessionIsolation()
    }

    const app = createApp(App)
    await setupApp(app, mode)

    const settingsStore = useSettingsStore()

    // 极验配置：captchaId 和 enabled 都从后端配置读取（运行时可变）
    const geetestCaptchaId = settingsStore.geetestCaptchaId
    const geetestEnabled = settingsStore.geetestEnabled

    // 只有当有 captchaId 时才初始化极验插件
    if (geetestCaptchaId) {
      app.use(Geetest, {
        captchaId: geetestCaptchaId,
        language: 'zho',
        product: 'popup',
      })
    }

    // 全局配置：告知组件极验是否启用
    app.config.globalProperties.$geetestConfig = {
      geetest_enabled: geetestEnabled.toString(),
      geetest_captcha_id: geetestCaptchaId,
    }

    // 设置全局应用配置
    app.config.globalProperties.$appConfig = {
      siteName: settingsStore.siteName,
      siteDesc: settingsStore.siteDesc,
      copyright: settingsStore.copyright,
      version: settingsStore.version,
      allowRegister: settingsStore.allowRegister,
      geetestEnabled,
    }

    app.mount('#app')
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[Vue 3] 应用启动失败:', error)
    throw error
  }
}
