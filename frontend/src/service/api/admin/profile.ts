/**
 * 管理端个人设置 API
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

function meUrl(path = '') {
  return `${getAdminApiBase()}/me${path}`
}

export interface AdminMeInfo {
  id?: number
  username?: string
  role?: string
  [key: string]: unknown
}

export const adminProfileApi = {
  me() {
    return request.Get<Service.ResponseResult<AdminMeInfo>>(meUrl())
  },
  changePassword(data: { old_password: string, new_password: string }) {
    return request.Put<Service.ResponseResult<{ message?: string }>>(meUrl('/password'), data)
  },
}
