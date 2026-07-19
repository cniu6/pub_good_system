import { request } from '../http'

export interface PhoneCountryDetectResult {
  country_code: string
  dial_code: string
  name_zh: string
  name_en: string
  source: string
  mobile_cn_only: boolean
  ip_detect_enabled: boolean
}

/** 探测默认手机号国家（语言 / IP / CDN / 保底 US） */
export function fetchPhoneCountryDetect(lang: string) {
  const methodInstance = request.Get<Service.ResponseResult<PhoneCountryDetectResult>>(
    '/api/v1/public/geo/phone-country',
    { params: { lang } },
  )
  methodInstance.meta = {
    authRole: null,
    noErrorTip: true,
  }
  return methodInstance
}
