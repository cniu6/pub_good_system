import type { App as AppInstance } from 'vue'
import App from './App.vue'
import { installRouter } from '@/router'
import { installPinia } from '@/store'
import { Geetest } from 'vue3-geetest'

// 打印编译时间
if (typeof __BUILD_TIMESTAMP__ !== 'undefined') {
  // eslint-disable-next-line no-console
  console.info(
    `%c 🐔 golxc.com %c (Build: ${__BUILD_TIMESTAMP__}) `,
    'color:#333333; background:#fff; padding:5px 10px; font-weight: 600; border-radius: 4px;',
    'color:#eee; background: conic-gradient(from 180deg at 50.03% 32.23%, #165dff 37.1706104279deg, #5679ff 136.8750035763deg, #45adf9 163.9506268501deg, #5f94f4 0.5765463114turn, #fb923c 228.0485773087deg, #c82ee1 304.4084072113deg, #fb923c 333.1886172295deg); padding:5px 10px; border-radius: 4px;',
  )
}

async function setupApp() {
  // 创建应用实例
  const app = createApp(App)

  // 注册 Pinia 状态管理
  await installPinia(app)

  // 注册 Vue Router
  await installRouter(app)

  // 极验配置 - 从环境变量读取
  const geetestConfig = {
    captchaId: import.meta.env.VITE_GEETEST_CAPTCHA_ID || import.meta.env.VITE_GEETEST_ID || '',
    language: 'eng' as const,
    product: 'popup' as 'popup' | 'float' | 'bind',
    nativeButton: {
      width: '100%',
      height: '3rem',
    },
    enabled: import.meta.env.VITE_GEETEST_ENABLED !== 'false', // 默认启用
  }

  // 注入全局配置供组件使用
  app.provide('geetestConfig', geetestConfig)

  // 注册极验插件
  app.use(Geetest, {
    captchaId: geetestConfig.captchaId,
    language: geetestConfig.language,
    product: geetestConfig.product,
  })

  // 配置极验验证（用于全局访问）
  app.config.globalProperties.$geetestConfig = {
    geetest_enabled: geetestConfig.enabled.toString(),
    geetest_captcha_id: geetestConfig.captchaId,
  }

  // 注册其他模块（指令、静态资源等）
  const modules = import.meta.glob<{ install: (app: AppInstance) => void }>('./modules/*.ts', {
    eager: true,
  })
  Object.values(modules).forEach(module => app.use(module))

  // 挂载应用
  app.mount('#app')
}

// 启动应用
setupApp().catch(console.error)
