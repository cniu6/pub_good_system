import { request } from '../http'

/**
 * 应用配置接口
 * 用于前端启动时获取运行时配置
 */

// 应用配置类型定义
export interface AppConfig {
  // 基本配置
  site_name: string
  site_desc: string
  site_logo: string
  copyright: string
  icp: string
  version: string

  // 功能开关
  allow_register: boolean
  announcement_enabled: boolean
  allow_delete_account: boolean
  geetest_enabled: boolean

  // 极验配置
  geetest_captcha_id: string

  // 语言配置
  default_lang: string

  // 验证码开关
  email_verify_enabled: boolean
  sms_verify_enabled: boolean
  /** true=仅中国大陆手机号(+86)；false=允许国际 E.164 */
  mobile_cn_only: boolean
  /** 国际号模式下是否按 IP/CDN 预选国家区号 */
  mobile_ip_country_detect: boolean

  // 实名认证配置
  realname_enabled: boolean
  realname_notify_text: string

  // 提现配置
  withdraw_enabled: boolean
  withdraw_min_amount: number
  withdraw_notify_text: string
  withdraw_account_types: string[]

  /**
   * 管理端 REST API 在 /api/v1 下的前缀（来自后端 ADMIN_API_PATH，默认 /admin）
   * 前端应注入到 getAdminApiBase()，勿写死 /admin
   */
  admin_api_path?: string

  /** 在线心跳上报周期（秒），Presence 心跳按此间隔发送，默认30秒 */
  online_report_interval_seconds?: number
}

/**
 * 获取应用配置
 * 此接口无需登录，在应用启动时调用
 */
export function fetchAppConfig() {
  const methodInstance = request.Get<Service.ResponseResult<AppConfig>>('/api/v1/public/app-config')
  methodInstance.meta = {
    authRole: null, // 无需认证
    noErrorTip: true, // 静默失败
  }
  return methodInstance
}
