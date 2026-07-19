import { request } from '../http'

/**
 * 尽力撤销当前会话记录，即使 access token 已过期也能清理。
 * 走公开接口（不依赖鉴权中间件），后端只验证签名、不校验过期时间，
 * 专门用于"token 已过期被强退登录"时兜底清理数据库里的会话记录。
 */
export function forceLogoutSession() {
  const methodInstance = request.Post<Service.ResponseResult<{ message: string }>>(
    '/api/v1/public/session/force-logout',
  )
  methodInstance.meta = {
    noErrorTip: true,
  }
  return methodInstance
}
