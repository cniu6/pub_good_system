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
