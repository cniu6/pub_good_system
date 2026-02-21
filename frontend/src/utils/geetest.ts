// 极验结果接口（vue3-geetest 格式）
interface GeetestResult {
  lot_number: string
  captcha_output: string
  pass_token: string
  gen_time: string
  captcha_id: string
}

class GeetestManager {
  private static instance: GeetestManager
  private currentCaptchaResult: GeetestResult | null = null
  private isGeetestEnabled: boolean = (() => {
    const id = import.meta.env.VITE_GEETEST_CAPTCHA_ID || import.meta.env.VITE_GEETEST_ID
    if (!id)
      return false
    if (/your[_-]?geetest[_-]?id/i.test(id))
      return false
    return true
  })()

  static getInstance(): GeetestManager {
    if (!GeetestManager.instance) {
      GeetestManager.instance = new GeetestManager()
    }
    return GeetestManager.instance
  }

  // 检查极验是否启用
  isEnabled(): boolean {
    return this.isGeetestEnabled
  }

  // 设置验证结果
  setCaptchaResult(result: GeetestResult | null): void {
    this.currentCaptchaResult = result
  }

  // 获取验证结果
  getCaptchaResult(): GeetestResult | null {
    return this.currentCaptchaResult
  }

  // 清除验证结果
  clearCaptchaResult(): void {
    this.currentCaptchaResult = null
  }

  // 生成极验请求头
  generateGeetestHeaders(): Record<string, string> {
    const result = this.getCaptchaResult()

    if (!result || !this.isGeetestEnabled) {
      return {}
    }

    return {
      'X-Geetest-Lot-Number': result.lot_number,
      'X-Geetest-Captcha-Output': result.captcha_output,
      'X-Geetest-Pass-Token': result.pass_token,
      'X-Geetest-Gen-Time': result.gen_time,
      'X-Geetest-Captcha-Id': result.captcha_id,
    }
  }

  // 检查验证是否有效（简单的时间戳验证）
  isCaptchaValid(): boolean {
    const result = this.getCaptchaResult()
    if (!result)
      return false

    // 验证结果5分钟内有效
    const currentTime = Math.floor(Date.now() / 1000)
    const captchaTime = Number.parseInt(result.gen_time)
    const timeDiff = currentTime - captchaTime

    return timeDiff < 300
  }

  // 验证并返回请求头，如果验证无效则清除结果
  getValidGeetestHeaders(): Record<string, string> {
    if (!this.isGeetestEnabled) {
      return {}
    }

    if (!this.isCaptchaValid()) {
      this.clearCaptchaResult()
      return {}
    }

    return this.generateGeetestHeaders()
  }
}

export const geetestManager = GeetestManager.getInstance()

// 导出类型供组件使用
export type { GeetestResult }
