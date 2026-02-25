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

  static getInstance(): GeetestManager {
    if (!GeetestManager.instance) {
      GeetestManager.instance = new GeetestManager()
    }
    return GeetestManager.instance
  }

  // 检查验证结果是否存在（用于判断是否已完成验证）
  hasCaptchaResult(): boolean {
    return this.currentCaptchaResult !== null
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

    if (!result) {
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
