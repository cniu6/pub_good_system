import type { Method } from 'alova'

/**
 * 当前浏览器页签的会话过期状态。
 *
 * 第一个确认失效的请求会开启门闩：中断尚未完成的受保护请求，并拒绝后续
 * 受保护请求进入网络层。登录、注册、找回密码等公开认证请求可显式放行。
 */
let sessionExpired = false
const pendingProtectedMethods = new Set<Method<any>>()

export class SessionRequestSuspendedError extends Error {
  constructor() {
    super('Session reauthentication is required')
    this.name = 'SessionRequestSuspendedError'
  }
}

export function isSessionExpired(): boolean {
  return sessionExpired
}

export function isSessionRequestSuspendedError(error: unknown): boolean {
  return error instanceof SessionRequestSuspendedError
}

export function canRequestDuringSessionRecovery(method: Method<any>): boolean {
  return method.meta?.allowDuringSessionRecovery === true
}

export function trackProtectedRequest(method: Method<any>) {
  if (!canRequestDuringSessionRecovery(method))
    pendingProtectedMethods.add(method)
}

export function releaseProtectedRequest(method: Method<any>) {
  pendingProtectedMethods.delete(method)
}

export function markSessionExpired(): boolean {
  if (sessionExpired)
    return false

  sessionExpired = true
  for (const method of pendingProtectedMethods)
    method.abort()
  pendingProtectedMethods.clear()
  return true
}

/** 成功登录后解除请求门闩，允许新会话正常发起请求。 */
export function resetSessionExpired() {
  sessionExpired = false
}
