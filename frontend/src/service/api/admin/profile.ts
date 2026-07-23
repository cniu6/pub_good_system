/**
 * 管理端个人设置 / TOTP API
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

function meUrl(path = '') {
  return `${getAdminApiBase()}/me${path}`
}

export interface AdminMeInfo {
  totp_enabled?: boolean
  rbac_roles?: Array<{ id: number, code: string, name: string }>
  [key: string]: unknown
}

export const adminProfileApi = {
  me() {
    return request.Get<Service.ResponseResult<AdminMeInfo>>(meUrl())
  },
  changePassword(data: { old_password: string, new_password: string }) {
    return request.Put<Service.ResponseResult<{ message?: string }>>(meUrl('/password'), data)
  },
  setupTotp() {
    return request.Post<Service.ResponseResult<{ secret?: string, otpauth_url?: string }>>(meUrl('/totp/setup'))
  },
  enableTotp(data: { code: string }) {
    return request.Post<Service.ResponseResult<{ message?: string }>>(meUrl('/totp/enable'), data)
  },
  disableTotp(data: { code: string }) {
    return request.Post<Service.ResponseResult<{ message?: string }>>(meUrl('/totp/disable'), data)
  },
}
