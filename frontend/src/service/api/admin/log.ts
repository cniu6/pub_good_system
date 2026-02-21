/**
 * 管理端 API 服务 - 操作日志
 */
import { request } from '@/service/http'

// 管理端 API 路径前缀（需与后端 ADMIN_PATH 保持一致）
const ADMIN_PATH = import.meta.env.VITE_ADMIN_API_PATH || import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/logs`

export const adminLogApi = {
  // 日志列表
  list(params?: {
    page?: number
    page_size?: number
    user_id?: number
    username?: string
    module?: string
    action?: string
    method?: string | null
    path?: string
    ip?: string
    start_time?: number
    end_time?: number
  }) {
    return request.Get(BASE_URL, { params })
  },

  // 日志统计
  stats() {
    return request.Get(`${BASE_URL}/stats`)
  },

  // 清理日志
  clean(before_time: number) {
    return request.Post(`${BASE_URL}/clean`, { before_time })
  },
}
