import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() {
  return `${getAdminApiBase()}/api-logs`
}

export interface APIAccessLog {
  id: number
  request_id: string
  user_id: number
  username: string
  role: string
  auth_method?: string
  scene: string
  method: string
  transport?: string
  protocol?: string
  path: string
  route_path: string
  handler_name?: string
  request_content_type?: string
  response_content_type?: string
  query_string: string
  path_params?: string
  ip: string
  source_ip?: string
  x_ip?: string
  x_forwarded_for?: string
  x_real_ip?: string
  user_agent?: string
  referer?: string
  request_headers?: string
  request_body?: string
  response_body?: string
  status_code: number
  duration: number
  request_size: number
  response_size: number
  create_time?: number
}

export interface APIAccessLogListParams {
  page?: number
  page_size?: number
  keyword?: string
  request_id?: string
  scene?: string
  auth_method?: string
  transport?: string
  user_id?: number
  username?: string
  method?: string
  path?: string
  ip?: string
  status_code?: number
  start_time?: number
  end_time?: number
}

export interface APIAccessLogStats {
  total_count: number
  today_count: number
  success_count: number
  client_error_count: number
  server_error_count: number
  distinct_ip_count: number
  avg_duration: number
  top_paths: Array<{ route_path: string, count: number, avg_duration: number }>
  method_stats: Array<{ method: string, count: number }>
  scene_stats: Array<{ scene: string, count: number }>
}

export const adminAPILogApi = {
  list(params?: APIAccessLogListParams) {
    return request.Get<Service.ResponseResult<{ list: APIAccessLog[], total: number, page: number, page_size: number }>>(baseUrl(), { params })
  },

  detail(id: number | string) {
    return request.Get<Service.ResponseResult<APIAccessLog>>(`${baseUrl()}/${id}`)
  },

  stats() {
    return request.Get<Service.ResponseResult<APIAccessLogStats>>(`${baseUrl()}/stats`)
  },

  clean(before_time: number) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/clean`, { before_time })
  },
}
