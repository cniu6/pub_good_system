import { authStorage, getBrowserId } from '@/utils'
import { geetestManager } from '@/utils/geetest'
import { createAlova } from 'alova'
import { createServerTokenAuthentication } from 'alova/client'
import adapterFetch from 'alova/fetch'
import VueHook from 'alova/vue'
import type { VueHookType } from 'alova/vue'
import {
  DEFAULT_ALOVA_OPTIONS,
  DEFAULT_BACKEND_OPTIONS,
} from './config'
import {
  handleBusinessError,
  handleRefreshToken,
  handleResponseError,
  handleServiceResult,
  localizeBackendMessagePayload,
  normalizeRequestError,
} from './handle'
import {
  canRequestDuringSessionRecovery,
  isSessionExpired,
  isSessionRequestSuspendedError,
  releaseProtectedRequest,
  SessionRequestSuspendedError,
  trackProtectedRequest,
} from './auth-expiration'
import { useAuthStore } from '@/store'

const { onAuthRequired, onResponseRefreshToken } = createServerTokenAuthentication<VueHookType>({
  // 服务端判定token过期
  refreshTokenOnSuccess: {
    // 当服务端返回401时，表示token过期
    isExpired: async (response, method) => {
      let businessCode: unknown
      try {
        const res = await response.clone().json() as Record<string, unknown>
        businessCode = res.code
      }
      catch {
        businessCode = undefined
      }

      const isExpired = method.meta && method.meta.isExpired
      return (Number(businessCode) === 401 || response.status === 401) && !isExpired
    },

    // 当token过期时触发，在此函数中触发刷新token
    handler: async (_response, method) => {
      // 此处采取限制，防止过期请求无限循环重发
      if (!method.meta)
        method.meta = { isExpired: true }
      else
        method.meta.isExpired = true

      // 记录触发失效判定时的 access token：多标签共享 localStorage 时，
      // 即便刷新失败或重试后仍 401，也只清「仍然是这一枚」的会话，
      // 避免误清期间已重新登录或其它标签已写入的新会话（见 store/auth.ts requireReauthentication）。
      method.meta.failedAccessToken = authStorage.get('accessToken') || undefined

      const refreshed = await handleRefreshToken()
      if (!refreshed)
        method.meta.sessionExpired = true
    },
  },
  // 添加token到请求头
  assignToken: (method) => {
    method.config.headers.Authorization = `Bearer ${authStorage.get('accessToken')}`
  },
})

// docs path of alova.js https://alova.js.org/
export function createAlovaInstance(
  alovaConfig: Service.AlovaConfig,
  backendConfig?: Service.BackendConfig,
) {
  const _backendConfig = { ...DEFAULT_BACKEND_OPTIONS, ...backendConfig }
  const _alovaConfig = { ...DEFAULT_ALOVA_OPTIONS, ...alovaConfig }

  return createAlova({
    statesHook: VueHook,
    requestAdapter: adapterFetch(),
    cacheFor: null,
    baseURL: _alovaConfig.baseURL,
    timeout: _alovaConfig.timeout,

    beforeRequest: onAuthRequired((method) => {
      // 会话失效弹窗显示期间，所有受保护请求均在发出前终止，避免轮询继续消耗流量。
      if (isSessionExpired() && !canRequestDuringSessionRecovery(method))
        throw new SessionRequestSuspendedError()

      trackProtectedRequest(method)
      // 自动添加极验验证头
      const geetestHeaders = geetestManager.getValidGeetestHeaders()
      Object.assign(method.config.headers, geetestHeaders)
      // 同浏览器实例 ID：登录/刷新时后端用来合并多标签重复会话
      method.config.headers['X-Browser-Id'] = getBrowserId()

      if (method.meta?.isFormPost) {
        method.config.headers['Content-Type'] = 'application/x-www-form-urlencoded'
        method.data = new URLSearchParams(method.data as URLSearchParams).toString()
      }
      alovaConfig.beforeRequest?.(method)
    }),
    responded: onResponseRefreshToken({
      // 请求成功的拦截器
      onSuccess: async (response, method) => {
        const { status } = response

        if (status === 200) {
          // 返回blob数据
          if (method.meta?.isBlob)
            return response.blob()

          // 返回json数据
          const apiData = await response.json() as Record<string, unknown>
          const localizedApiData = localizeBackendMessagePayload(apiData, _backendConfig)
          // 请求成功
          if (localizedApiData[_backendConfig.codeKey] === _backendConfig.successCode)
            return handleServiceResult(localizedApiData)

          // 刷新+重试后仍业务 401：无论上游判断如何，强制进入登录恢复弹窗，堵死静默失败。
          const stillUnauthorized = Boolean(method.meta?.isExpired)
            && Number(localizedApiData[_backendConfig.codeKey]) === 401
          if (stillUnauthorized)
            useAuthStore().requireReauthentication(method.meta?.failedAccessToken)

          // 业务请求失败
          const errorResult = handleBusinessError(
            localizedApiData,
            _backendConfig,
            method.meta?.noErrorTip || method.meta?.sessionExpired || stillUnauthorized,
          )
          return handleServiceResult(errorResult, false)
        }
        // 刷新+重试后仍 HTTP 401：同上，强制恢复弹窗。
        const stillUnauthorized = Boolean(method.meta?.isExpired) && status === 401
        if (stillUnauthorized)
          useAuthStore().requireReauthentication(method.meta?.failedAccessToken)
        // 接口请求失败
        const errorResult = await handleResponseError(
          response,
          method.meta?.sessionExpired || stillUnauthorized,
        )
        return handleServiceResult(errorResult, false)
      },
      onError: async (error, method) => {
        releaseProtectedRequest(method)
        const requestSuspended = isSessionRequestSuspendedError(error)
        const normalizedError = requestSuspended
          ? {
              errorType: 'Response Error' as const,
              code: 401,
              message: '',
              data: null,
            }
          : normalizeRequestError(error, method.url)
        if (!requestSuspended && !isSessionExpired() && !method.meta?.sessionExpired)
          window.$message?.error(normalizedError.message)
        throw handleServiceResult(normalizedError, false)
      },

      onComplete: async (method) => {
        releaseProtectedRequest(method)
      },
    }),
  })
}
