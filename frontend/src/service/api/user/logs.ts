/**
 * 用户端 - 本人操作日志 / API 访问日志
 * 后端强制按当前登录 user_id 过滤，前端无需也不能传别人的 user_id
 */
import { request } from '../../http'

export interface UserOperationLog {
  id: number
  user_id: number
  username?: string
  module?: string
  action?: string
  method?: string
  path?: string
  ip?: string
  handler_name?: string
  request_body?: string
  response_body?: string
  status_code?: number
  duration?: number
  create_time?: number
}

export interface UserAPIAccessLog {
  id: number
  request_id?: string
  user_id: number
  username?: string
  scene?: string
  method?: string
  path?: string
  route_path?: string
  query_string?: string
  path_params?: string
  ip?: string
  request_headers?: string
  request_body?: string
  response_body?: string
  status_code?: number
  duration?: number
  create_time?: number
  user_agent?: string
}

export function fetchMyOperationLogs(params?: {
  page?: number
  page_size?: number
  start_time?: number
  end_time?: number
}) {
  return request.Get<Service.ResponseResult<{ list: UserOperationLog[], total: number, page: number, page_size: number }>>(
    '/api/v1/user/logs',
    { params },
  )
}

export function fetchMyOperationLogDetail(id: number) {
  return request.Get<Service.ResponseResult<UserOperationLog>>(`/api/v1/user/logs/${id}`)
}

export function fetchMyAPILogs(params?: {
  page?: number
  page_size?: number
  keyword?: string
  method?: string
  path?: string
  status_code?: number
  start_time?: number
  end_time?: number
}) {
  return request.Get<Service.ResponseResult<{ list: UserAPIAccessLog[], total: number, page: number, page_size: number }>>(
    '/api/v1/user/api-logs',
    { params },
  )
}

export function fetchMyAPILogDetail(id: number | string) {
  return request.Get<Service.ResponseResult<UserAPIAccessLog>>(`/api/v1/user/api-logs/${id}`)
}
