import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { fetchAppConfig, type AppConfig } from '@/service/api/app-config'

/**
 * 运行时配置 Store
 *
 * 用于存储从后端获取的动态配置信息，替代部分 VITE_ 环境变量。
 * 这些配置可以在管理后台动态修改，无需重新构建前端。
 */
export const useSettingsStore = defineStore('settings-store', () => {
  // ========================================
  // 状态
  // ========================================

  // 应用配置
  const config = ref<AppConfig | null>(null)

  // 是否已加载
  const isLoaded = ref(false)

  // 是否正在加载
  const isLoading = ref(false)

  // 加载错误
  const loadError = ref<string | null>(null)

  // ========================================
  // 计算属性
  // ========================================

  // 站点名称
  const siteName = computed(() => config.value?.site_name ?? import.meta.env.VITE_APP_NAME ?? 'F.st')

  // 站点描述
  const siteDesc = computed(() => config.value?.site_desc ?? '基于 Go + Vue 3 的全栈管理系统模板')

  // 站点Logo
  const siteLogo = computed(() => config.value?.site_logo ?? '')

  // 版权信息
  const copyright = computed(() => config.value?.copyright ?? import.meta.env.VITE_COPYRIGHT_INFO ?? '© 2024 F.st')

  // ICP备案号
  const icp = computed(() => config.value?.icp ?? '')

  // 系统版本
  const version = computed(() => config.value?.version ?? '1.0.0')

  // 是否允许注册
  const allowRegister = computed(() => config.value?.allow_register ?? true)

  // 默认语言
  const defaultLang = computed(() => config.value?.default_lang ?? import.meta.env.VITE_DEFAULT_LANG ?? 'zhCN')

  // 极验是否启用（从后端配置获取）
  const geetestEnabled = computed(() => config.value?.geetest_enabled ?? false)

  // 极验验证码ID（从后端配置获取）
  const geetestCaptchaId = computed(() => config.value?.geetest_captcha_id ?? '')

  // ========================================
  // Actions
  // ========================================

  /**
   * 从后端加载应用配置
   * 应在应用启动时调用
   */
  async function loadConfig() {
    if (isLoading.value || isLoaded.value) {
      return
    }

    isLoading.value = true
    loadError.value = null

    try {
      const response = await fetchAppConfig()
      if (response.data) {
        config.value = response.data
        isLoaded.value = true

        // 输出调试信息
        if (import.meta.env.DEV) {
          console.log('[Settings] App config loaded:', config.value)
        }
      }
    }
    catch (error: any) {
      console.warn('[Settings] Failed to load app config:', error)
      loadError.value = error.message || 'Failed to load app config'

      // 加载失败时使用默认值，不影响应用启动
      isLoaded.value = true
    }
    finally {
      isLoading.value = false
    }
  }

  /**
   * 强制重新加载配置
   */
  async function reloadConfig() {
    isLoaded.value = false
    config.value = null
    await loadConfig()
  }

  /**
   * 更新配置（用于管理端修改后同步更新）
   */
  function updateConfig(newConfig: Partial<AppConfig>) {
    if (config.value) {
      config.value = { ...config.value, ...newConfig }
    }
  }

  return {
    // 状态
    config,
    isLoaded,
    isLoading,
    loadError,

    // 计算属性
    siteName,
    siteDesc,
    siteLogo,
    copyright,
    icp,
    version,
    allowRegister,
    defaultLang,
    geetestEnabled,
    geetestCaptchaId,

    // Actions
    loadConfig,
    reloadConfig,
    updateConfig,
  }
})
