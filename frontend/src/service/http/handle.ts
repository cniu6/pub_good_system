import { refreshAuthToken } from './token-refresh'
import { useAuthStore } from '@/store'
import { $t, authStorage } from '@/utils'
import {
  ERROR_NO_TIP_STATUS,
  ERROR_STATUS,
} from './config'

type ErrorStatus = keyof typeof ERROR_STATUS
type BackendBusinessPayload = Record<string, unknown> & { data?: unknown }
type BackendResponsePayload = Record<string, unknown>
interface LocalizedBackendMessageRule {
  aliases?: string[]
  pattern?: RegExp
  localeKey?: string
  resolve?: (match: RegExpMatchArray) => string
}

const BACKEND_MESSAGE_RULES: LocalizedBackendMessageRule[] = [
  {
    aliases: ['Invalid or expired token', 'Session expired or revoked', 'Invalid user session', 'Invalid or expired refresh token', 'Refresh session expired or revoked'],
    localeKey: 'http.backendMessage.loginExpired',
  },
  {
    aliases: ['Authorization header is required', 'Authorization header format must be Bearer {token}', 'User not logged in', '用户未登录'],
    localeKey: 'http.backendMessage.loginRequired',
  },
  {
    aliases: ['User not found', 'user not found', '用户不存在'],
    localeKey: 'http.backendMessage.userNotFound',
  },
  {
    aliases: ['Admin access only'],
    localeKey: 'http.backendMessage.adminOnly',
  },
  {
    aliases: ['Insufficient permissions', 'Role not found', 'Invalid role type'],
    localeKey: 'http.backendMessage.insufficientPermissions',
  },
  {
    aliases: ['Captcha validation failed'],
    localeKey: 'http.backendMessage.captchaValidationFailed',
  },
  {
    aliases: ['Email already in use', '邮箱已存在', '邮箱已被使用'],
    localeKey: 'http.backendMessage.emailAlreadyInUse',
  },
  {
    aliases: ['Mobile already in use', '手机号已存在', '手机号已被使用'],
    localeKey: 'http.backendMessage.mobileAlreadyInUse',
  },
  {
    aliases: ['Registration is disabled'],
    localeKey: 'http.backendMessage.registrationDisabled',
  },
  {
    aliases: ['Account deletion is currently disabled'],
    localeKey: 'http.backendMessage.accountDeletionDisabled',
  },
  {
    aliases: ['Invalid or expired verification code', '验证码错误或已过期'],
    localeKey: 'http.backendMessage.invalidVerificationCode',
  },
  {
    aliases: ['Please wait before requesting another verification code'],
    localeKey: 'http.backendMessage.verificationCodeCooldown',
  },
  {
    aliases: ['Invalid account or password'],
    localeKey: 'http.backendMessage.invalidAccountOrPassword',
  },
  {
    aliases: ['Account is inactive'],
    localeKey: 'http.backendMessage.accountInactive',
  },
  {
    aliases: ['Web login is disabled'],
    localeKey: 'http.backendMessage.webLoginDisabled',
  },
  {
    aliases: ['incorrect old password'],
    localeKey: 'http.backendMessage.incorrectOldPassword',
  },
  {
    aliases: ['Verification code sent'],
    localeKey: 'profile.codeSent',
  },
  {
    aliases: ['Verification code sent to new email'],
    localeKey: 'profile.emailCodeSent',
  },
  {
    aliases: ['User registered successfully'],
    localeKey: 'login.registerSuccess',
  },
  {
    aliases: ['提现功能暂未开启'],
    localeKey: 'moneyScore.withdrawDisabled',
  },
  {
    aliases: ['提现申请已提交，等待管理员审核'],
    localeKey: 'moneyScore.withdrawSubmitted',
  },
  {
    aliases: ['请完整填写收款信息'],
    localeKey: 'moneyScore.completeAccountInfo',
  },
  {
    aliases: ['当前收款方式不可用'],
    localeKey: 'moneyScore.selectValidAccountType',
  },
  {
    aliases: ['账户余额不足'],
    localeKey: 'http.backendMessage.insufficientBalance',
  },
  {
    aliases: ['Failed to update profile'],
    localeKey: 'profile.profileSaveFailed',
  },
  {
    aliases: ['Failed to change password'],
    localeKey: 'profile.changePasswordFailed',
  },
  {
    aliases: ['Failed to update avatar'],
    localeKey: 'userCenter.avatarUpdateFailed',
  },
  {
    aliases: ['Failed to update settings', 'Failed to save settings'],
    localeKey: 'settingsTab.saveFailed',
  },
  {
    aliases: ['Failed to change email'],
    localeKey: 'profile.emailVerifyFailed',
  },
  {
    aliases: ['Failed to change phone'],
    localeKey: 'profile.phoneVerifyFailed',
  },
  {
    aliases: ['Failed to deactivate account'],
    localeKey: 'securityTab.deactivateFailed',
  },
  {
    aliases: ['Failed to load sessions'],
    localeKey: 'securityTab.loadSessionsFailed',
  },
  {
    aliases: ['Failed to revoke session'],
    localeKey: 'securityTab.revokeFailed',
  },
  {
    aliases: ['Failed to revoke sessions'],
    localeKey: 'securityTab.revokeAllFailed',
  },
  {
    aliases: ['Failed to reset API key'],
    localeKey: 'apiTab.apiKeyResetFailed',
  },
  {
    pattern: /^Account is locked\. Please try again in (\d+) minutes$/i,
    resolve: match => $t('http.backendMessage.accountLocked', { minutes: match[1] }),
  },
  {
    pattern: /^Failed to send .+$/i,
    localeKey: 'profile.sendCodeFailed',
  },
  {
    pattern: /^Failed to check verification cooldown$/i,
    localeKey: 'profile.sendCodeFailed',
  },
  {
    pattern: /^Failed to generate verification code$/i,
    localeKey: 'profile.sendCodeFailed',
  },
  {
    pattern: /^Failed to (load|fetch|get) .+$/i,
    localeKey: 'http.backendMessage.loadFailed',
  },
  {
    pattern: /^Failed to (update|save) .+$/i,
    localeKey: 'http.backendMessage.saveFailed',
  },
  {
    pattern: /^Failed to change .+$/i,
    localeKey: 'http.backendMessage.changeFailed',
  },
  {
    pattern: /^Failed to reset .+$/i,
    localeKey: 'http.backendMessage.resetFailed',
  },
  {
    pattern: /^Failed to revoke .+$/i,
    localeKey: 'http.backendMessage.revokeFailed',
  },
  {
    pattern: /^Failed to .+$/i,
    localeKey: 'http.backendMessage.operationFailed',
  },
]

function normalizeBackendMessage(message: string) {
  return message.trim().replace(/\s+/g, ' ').toLowerCase()
}

function renderLocalizedBackendMessage(rule: LocalizedBackendMessageRule, match?: RegExpMatchArray) {
  if (rule.resolve && match)
    return rule.resolve(match)

  if (rule.localeKey)
    return $t(rule.localeKey)

  return ERROR_STATUS.default
}

function localizeBackendMessage(rawMessage: string) {
  const message = rawMessage.trim()
  if (!message)
    return ERROR_STATUS.default

  const normalizedMessage = normalizeBackendMessage(message)

  for (const rule of BACKEND_MESSAGE_RULES) {
    if (rule.aliases?.some(alias => normalizeBackendMessage(alias) === normalizedMessage))
      return renderLocalizedBackendMessage(rule)

    if (rule.pattern) {
      const match = message.match(rule.pattern)
      if (match)
        return renderLocalizedBackendMessage(rule, match)
    }
  }

  return message
}

async function extractResponseMessage(response: Response) {
  try {
    const payload = await response.clone().json() as BackendResponsePayload
    if (typeof payload.message === 'string' && payload.message.trim())
      return payload.message
  }
  catch {
  }

  try {
    const text = (await response.clone().text()).trim()
    if (text && !/^<!doctype html/i.test(text) && !/^<html/i.test(text))
      return text
  }
  catch {
  }

  return ''
}

export function localizeBackendMessagePayload<T extends BackendResponsePayload>(data: T, config: Required<Service.BackendConfig>) {
  const { msgKey } = config
  const rawMessage = data[msgKey]
  if (typeof rawMessage !== 'string' || !rawMessage.trim())
    return data

  return {
    ...data,
    [msgKey]: localizeBackendMessage(rawMessage),
  } as T
}

export function normalizeRequestError(error: unknown, requestUrl?: string): Service.RequestError {
  const rawMessage = error instanceof Error ? error.message : String(error || '')
  const lowerMessage = rawMessage.toLowerCase()
  const isNetworkFailure = lowerMessage.includes('failed to fetch')
    || lowerMessage.includes('networkerror')
    || lowerMessage.includes('load failed')
    || lowerMessage.includes('fetch failed')
    || lowerMessage.includes('network request failed')

  const message = isNetworkFailure
    ? `${ERROR_STATUS.network}: ${requestUrl || ERROR_STATUS.default}`
    : `${ERROR_STATUS.unknown}: ${rawMessage || ERROR_STATUS.default}`

  return {
    errorType: 'Response Error',
    code: isNetworkFailure ? 'NETWORK_ERROR' : 'UNKNOWN_ERROR',
    message,
    data: null,
  }
}

/**
 * @description: 处理请求成功，但返回后端服务器报错
 * @param {Response} response
 * @return {*}
 */
export async function handleResponseError(response: Response) {
  const error: Service.RequestError = {
    errorType: 'Response Error',
    code: 0,
    message: ERROR_STATUS.default,
    data: null,
  }
  const errorCode: ErrorStatus = response.status as ErrorStatus
  const rawMessage = await extractResponseMessage(response)
  const message = rawMessage ? localizeBackendMessage(rawMessage) : (ERROR_STATUS[errorCode] || ERROR_STATUS.default)
  Object.assign(error, { code: errorCode, message })

  showError(error)

  return error
}

/**
 * @description:
 * @param {Record} data 接口返回的后台数据
 * @param {Service} config 后台字段配置
 * @param {boolean} noErrorTip 是否不显示错误提示
 * @return {*}
 */
export function handleBusinessError(data: BackendBusinessPayload, config: Required<Service.BackendConfig>, noErrorTip?: boolean) {
  const { codeKey, msgKey } = config
  const rawCode = data[codeKey]
  const rawMessage = data[msgKey]
  const error: Service.RequestError = {
    errorType: 'Business Error',
    code: typeof rawCode === 'number' || typeof rawCode === 'string' ? rawCode : 0,
    message: typeof rawMessage === 'string' ? localizeBackendMessage(rawMessage) : ERROR_STATUS.default,
    data: data.data,
  }

  if (!noErrorTip) {
    showError(error)
  }

  return error
}

/**
 * @description: 统一成功和失败返回类型
 * @param {any} data
 * @param {boolean} isSuccess
 * @return {*} result
 */
export function handleServiceResult<T extends object>(data: T, isSuccess: boolean = true) {
  const result = {
    isSuccess,
    errorType: null,
    ...data,
  } as T & { isSuccess: boolean, errorType: Service.RequestErrorType }
  return result
}

/**
 * @description: 处理接口token刷新
 * @return {*}
 */
export async function handleRefreshToken() {
  const authStore = useAuthStore()
  const isAutoRefresh = import.meta.env.VITE_AUTO_REFRESH_TOKEN === 'Y'
  if (!isAutoRefresh) {
    await authStore.logout()
    return
  }

  // 记录发起刷新时的会话代际，避免"请求发出后用户登出/重新登录，刷新结果晚点才回来"
  // 时把已登出的会话重新救活（详见 store/auth.ts 的 authGeneration 说明）。
  const generation = authStore.authGeneration

  // 刷新token
  const data = await refreshAuthToken(authStorage.get('refreshToken'))
  if (authStore.authGeneration !== generation)
    return

  if (data) {
    authStore.applyRefreshedLoginInfo(data)
    return
  }

  // 刷新失败，退出
  await authStore.logout()
}

export function showError(error: Service.RequestError) {
  // 如果error不需要提示,则跳过
  const code = Number(error.code)
  if (ERROR_NO_TIP_STATUS.includes(code))
    return

  window.$message.error(error.message)
}
