/**
 * 常用国家区号（与后端 utils/dial_codes.go 对齐）
 * label 由组件按当前语言拼装
 */
export interface DialCountry {
  code: string
  dialCode: string
  nameZh: string
  nameEn: string
}

export const DEFAULT_DIAL_COUNTRY_CODE = 'US'

export const DIAL_COUNTRIES: DialCountry[] = [
  { code: 'CN', dialCode: '86', nameZh: '中国大陆', nameEn: 'China Mainland' },
  { code: 'HK', dialCode: '852', nameZh: '中国香港', nameEn: 'Hong Kong' },
  { code: 'MO', dialCode: '853', nameZh: '中国澳门', nameEn: 'Macao' },
  { code: 'TW', dialCode: '886', nameZh: '中国台湾', nameEn: 'Taiwan' },
  { code: 'US', dialCode: '1', nameZh: '美国', nameEn: 'United States' },
  { code: 'CA', dialCode: '1', nameZh: '加拿大', nameEn: 'Canada' },
  { code: 'GB', dialCode: '44', nameZh: '英国', nameEn: 'United Kingdom' },
  { code: 'AU', dialCode: '61', nameZh: '澳大利亚', nameEn: 'Australia' },
  { code: 'NZ', dialCode: '64', nameZh: '新西兰', nameEn: 'New Zealand' },
  { code: 'JP', dialCode: '81', nameZh: '日本', nameEn: 'Japan' },
  { code: 'KR', dialCode: '82', nameZh: '韩国', nameEn: 'South Korea' },
  { code: 'SG', dialCode: '65', nameZh: '新加坡', nameEn: 'Singapore' },
  { code: 'MY', dialCode: '60', nameZh: '马来西亚', nameEn: 'Malaysia' },
  { code: 'TH', dialCode: '66', nameZh: '泰国', nameEn: 'Thailand' },
  { code: 'VN', dialCode: '84', nameZh: '越南', nameEn: 'Vietnam' },
  { code: 'PH', dialCode: '63', nameZh: '菲律宾', nameEn: 'Philippines' },
  { code: 'ID', dialCode: '62', nameZh: '印度尼西亚', nameEn: 'Indonesia' },
  { code: 'IN', dialCode: '91', nameZh: '印度', nameEn: 'India' },
  { code: 'AE', dialCode: '971', nameZh: '阿联酋', nameEn: 'United Arab Emirates' },
  { code: 'SA', dialCode: '966', nameZh: '沙特阿拉伯', nameEn: 'Saudi Arabia' },
  { code: 'TR', dialCode: '90', nameZh: '土耳其', nameEn: 'Turkey' },
  { code: 'RU', dialCode: '7', nameZh: '俄罗斯', nameEn: 'Russia' },
  { code: 'DE', dialCode: '49', nameZh: '德国', nameEn: 'Germany' },
  { code: 'FR', dialCode: '33', nameZh: '法国', nameEn: 'France' },
  { code: 'IT', dialCode: '39', nameZh: '意大利', nameEn: 'Italy' },
  { code: 'ES', dialCode: '34', nameZh: '西班牙', nameEn: 'Spain' },
  { code: 'PT', dialCode: '351', nameZh: '葡萄牙', nameEn: 'Portugal' },
  { code: 'NL', dialCode: '31', nameZh: '荷兰', nameEn: 'Netherlands' },
  { code: 'BE', dialCode: '32', nameZh: '比利时', nameEn: 'Belgium' },
  { code: 'CH', dialCode: '41', nameZh: '瑞士', nameEn: 'Switzerland' },
  { code: 'AT', dialCode: '43', nameZh: '奥地利', nameEn: 'Austria' },
  { code: 'SE', dialCode: '46', nameZh: '瑞典', nameEn: 'Sweden' },
  { code: 'NO', dialCode: '47', nameZh: '挪威', nameEn: 'Norway' },
  { code: 'DK', dialCode: '45', nameZh: '丹麦', nameEn: 'Denmark' },
  { code: 'FI', dialCode: '358', nameZh: '芬兰', nameEn: 'Finland' },
  { code: 'IE', dialCode: '353', nameZh: '爱尔兰', nameEn: 'Ireland' },
  { code: 'PL', dialCode: '48', nameZh: '波兰', nameEn: 'Poland' },
  { code: 'CZ', dialCode: '420', nameZh: '捷克', nameEn: 'Czechia' },
  { code: 'BR', dialCode: '55', nameZh: '巴西', nameEn: 'Brazil' },
  { code: 'MX', dialCode: '52', nameZh: '墨西哥', nameEn: 'Mexico' },
  { code: 'AR', dialCode: '54', nameZh: '阿根廷', nameEn: 'Argentina' },
  { code: 'CL', dialCode: '56', nameZh: '智利', nameEn: 'Chile' },
  { code: 'CO', dialCode: '57', nameZh: '哥伦比亚', nameEn: 'Colombia' },
  { code: 'ZA', dialCode: '27', nameZh: '南非', nameEn: 'South Africa' },
  { code: 'EG', dialCode: '20', nameZh: '埃及', nameEn: 'Egypt' },
  { code: 'NG', dialCode: '234', nameZh: '尼日利亚', nameEn: 'Nigeria' },
  { code: 'IL', dialCode: '972', nameZh: '以色列', nameEn: 'Israel' },
  { code: 'PK', dialCode: '92', nameZh: '巴基斯坦', nameEn: 'Pakistan' },
  { code: 'BD', dialCode: '880', nameZh: '孟加拉国', nameEn: 'Bangladesh' },
  { code: 'UA', dialCode: '380', nameZh: '乌克兰', nameEn: 'Ukraine' },
]

export function getDialCountry(code: string): DialCountry | undefined {
  const c = String(code || '').toUpperCase()
  return DIAL_COUNTRIES.find(x => x.code === c)
}

/** 界面语言 → 默认国家：中文 CN，其它保底 US */
export function countryFromLanguage(lang: string): string {
  const l = String(lang || '').toLowerCase().replace('_', '-')
  if (l.startsWith('zh'))
    return 'CN'
  return DEFAULT_DIAL_COUNTRY_CODE
}

export function dialCountryLabel(c: DialCountry, isZh: boolean): string {
  const name = isZh ? c.nameZh : c.nameEn
  return `${name} (+${c.dialCode})`
}
