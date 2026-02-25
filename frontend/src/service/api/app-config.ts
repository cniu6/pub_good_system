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
  geetest_enabled: boolean

  // 极验配置
  geetest_captcha_id: string

  // 语言配置
  default_lang: string
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
