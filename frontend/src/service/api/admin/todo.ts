/**
 * 管理端待办聚合 API
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

export interface AdminTodoItem {
  type: string
  title: string
  count: number
  link: string
}

export const adminTodoApi = {
  list() {
    return request.Get<Service.ResponseResult<{ list: AdminTodoItem[] }>>(`${getAdminApiBase()}/todos`)
  },
}
