/**
 * 管理端用户等级能力 API
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

export interface UserLevelCap {
  level: number
  name: string
  allow_api_key: boolean
  allow_recharge: boolean
  allow_withdraw: boolean
  menu_flags: string
  create_time: number
}

export const adminUserLevelApi = {
  list() {
    return request.Get<Service.ResponseResult<{ list: UserLevelCap[] }>>(`${getAdminApiBase()}/user-levels`)
  },
  update(data: Partial<UserLevelCap> & { level: number }) {
    return request.Put<Service.ResponseResult<{ item: UserLevelCap }>>(`${getAdminApiBase()}/user-levels`, data)
  },
}
