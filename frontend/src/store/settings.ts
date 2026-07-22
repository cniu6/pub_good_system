import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchAppConfig } from '@/service/api/app-config'
import type { AppConfig } from '@/service/api/app-config'
import { getAdminApiPath, setRuntimeAdminApiPath } from '@/service/api/admin/base'
import { geetestManager } from '@/utils/geetest'
import { i18n } from '@/modules/i18n'

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
  const siteDesc = computed(() => config.value?.site_desc ?? i18n.global.t('settings.defaultSiteDesc'))

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

  // 站内公告总开关（关闭后用户端不展示铃铛/工作台公告区）
  const announcementEnabled = computed(() => config.value?.announcement_enabled ?? true)

  // 是否允许注销账号
  const allowDeleteAccount = computed(() => config.value?.allow_delete_account ?? false)

  // 是否禁止普通用户网页端登录（默认 false；仅小程序/App 场景使用，管理员登录不受影响）
  const webLoginDisabled = computed(() => config.value?.web_login_disabled ?? false)

  // 默认语言
  const defaultLang = computed(() => config.value?.default_lang ?? import.meta.env.VITE_DEFAULT_LANG ?? 'zhCN')

  // 极验是否启用（从后端配置获取）
  const geetestEnabled = computed(() => config.value?.geetest_enabled ?? false)

  // 极验验证码ID（从后端配置获取）
  const geetestCaptchaId = computed(() => config.value?.geetest_captcha_id ?? '')

  // 邮箱验证码是否启用
  const emailVerifyEnabled = computed(() => config.value?.email_verify_enabled ?? true)

  // 短信验证码是否启用
  const smsVerifyEnabled = computed(() => config.value?.sms_verify_enabled ?? false)

  /** 是否仅允许中国大陆手机号（默认 true，与后端一致） */
  const mobileCnOnly = computed(() => config.value?.mobile_cn_only ?? true)

  /** 国际号模式下是否按 IP 自动匹配国家（默认 false） */
  const mobileIpCountryDetect = computed(() => config.value?.mobile_ip_country_detect ?? false)

  // 实名认证功能是否启用
  const realnameEnabled = computed(() => config.value?.realname_enabled ?? true)

  // 实名认证提示语
  const realnameNotifyText = computed(() => config.value?.realname_notify_text ?? '')

  // 提现功能是否启用
  const withdrawEnabled = computed(() => config.value?.withdraw_enabled ?? true)

  // 最低提现金额
  const withdrawMinAmount = computed(() => config.value?.withdraw_min_amount ?? 10)

  // 提现提示语
  const withdrawNotifyText = computed(() => config.value?.withdraw_notify_text ?? '')

  // 支持的提现收款方式
  const withdrawAccountTypes = computed(() => config.value?.withdraw_account_types ?? ['bank', 'alipay', 'wechat', 'usdt'])

  // 提现前是否必须已完成实名认证并通过审核
  const withdrawRequireRealname = computed(() => config.value?.withdraw_require_realname ?? false)

  // 管理端 REST API 前缀（运行时注入后与后端 ADMIN_API_PATH 一致）
  const adminApiPath = computed(() => getAdminApiPath())

  // 在线心跳上报周期（秒），默认30秒；由管理端「在线用户」页可配置
  const onlineReportIntervalSeconds = computed(() => config.value?.online_report_interval_seconds ?? 30)

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
      if (response.isSuccess && response.data) {
        config.value = response.data
        isLoaded.value = true

        // 管理端 API 前缀：后端 ADMIN_API_PATH 注入前端，请求路径与路由一致
        setRuntimeAdminApiPath(response.data.admin_api_path)

        // 注册极验启用检查函数
        geetestManager.setEnabledChecker(() => geetestEnabled.value)
      }
      else {
        // 业务失败但未抛异常（如后端返回非 200 业务码）：与 catch 分支一致的兜底，
        // 避免 admin_api_path 一直未注入、isLoaded 停留在 false 导致应用卡在加载态
        loadError.value = response.message || 'Failed to load app config'
        setRuntimeAdminApiPath(import.meta.env.VITE_ADMIN_API_PATH)
        isLoaded.value = true
      }
    }
    catch (error: unknown) {
      loadError.value = error instanceof Error ? error.message : 'Failed to load app config'

      // 加载失败时使用默认值（VITE_ADMIN_API_PATH /admin），不影响应用启动
      setRuntimeAdminApiPath(import.meta.env.VITE_ADMIN_API_PATH)
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
      if (newConfig.admin_api_path !== undefined)
        setRuntimeAdminApiPath(newConfig.admin_api_path)
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
    announcementEnabled,
    allowDeleteAccount,
    webLoginDisabled,
    defaultLang,
    geetestEnabled,
    geetestCaptchaId,
    emailVerifyEnabled,
    smsVerifyEnabled,
    mobileCnOnly,
    mobileIpCountryDetect,
    realnameEnabled,
    realnameNotifyText,
    withdrawEnabled,
    withdrawMinAmount,
    withdrawNotifyText,
    withdrawAccountTypes,
    withdrawRequireRealname,
    adminApiPath,
    onlineReportIntervalSeconds,

    // Actions
    loadConfig,
    reloadConfig,
    updateConfig,
  }
})
