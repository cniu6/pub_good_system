import { fetchUpdateToken } from '../api/user/login'
import { useAuthStore } from '@/store'
import { getRuntimeRouteMode } from '@/router/runtime-mode'
import { authStorage } from '@/utils'
import {
  ERROR_NO_TIP_STATUS,
  ERROR_STATUS,
} from './config'

type ErrorStatus = keyof typeof ERROR_STATUS
type BackendBusinessPayload = Record<string, unknown> & { data?: unknown }
type LoginTokenPayload = Api.Login.Info & { expiresAt?: number }

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
export function handleResponseError(response: Response) {
  const error: Service.RequestError = {
    errorType: 'Response Error',
    code: 0,
    message: ERROR_STATUS.default,
    data: null,
  }
  const errorCode: ErrorStatus = response.status as ErrorStatus
  const message = ERROR_STATUS[errorCode] || ERROR_STATUS.default
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
    message: typeof rawMessage === 'string' ? rawMessage : ERROR_STATUS.default,
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

  // 刷新token
  const mode = getRuntimeRouteMode()
  const authGuard = mode === 'admin' ? 'admin' : 'user'
  try {
    const result = await fetchUpdateToken({ refreshToken: authStorage.get('refreshToken'), authGuard })
    const data = result.data as LoginTokenPayload | null
    if (result.isSuccess && data) {
      authStorage.setActive('accessToken', data.accessToken)
      authStorage.setActive('refreshToken', data.refreshToken)
      if (data.expiresAt) {
        authStorage.setActive('accessTokenExpiresAt', data.expiresAt)
      }
      return
    }
  }
  catch {
    // noop: 统一走退出逻辑
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
