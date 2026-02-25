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
import { setRuntimeRouteMode } from './router/runtime-mode'
import './styles/index.css'

async function setupApp(app: AppInstance<Element>, mode: AppRouteMode) {
  // 1. 首先安装 Pinia（其他模块依赖它）
  installPinia(app)

  // 2. 加载运行时配置（在安装其他模块之前）
  const settingsStore = useSettingsStore()
  await settingsStore.loadConfig()

  // 3. 安装其他模块
  setupI18n(app)
  await installRouter(app, mode)
  setupDirectives(app)
  setupAssets()
}

export async function bootstrap(mode: AppRouteMode) {
  try {
    setRuntimeRouteMode(mode)

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
      geetestEnabled: geetestEnabled,
    }

    app.mount('#app')
    console.log('[Vue 3] 应用已成功启动')
    console.log('[Vue 3] 运行时配置已加载:', {
      siteName: settingsStore.siteName,
      version: settingsStore.version,
      allowRegister: settingsStore.allowRegister,
      geetestEnabled: geetestEnabled,
    })
  }
  catch (error) {
    console.error('[Vue 3] 应用启动失败:', error)
    throw error
  }
}
