/**
 * 管理端 API 服务 - 仪表盘
 */
import { request } from '@/service/http'

const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/dashboard`

export interface AdminDashboardStatistics {
  total_users: number
  today_new_users: number
  active_users_7d: number
  total_money_logs: number
  total_score_logs: number
  total_operation_logs: number
  today_operation_logs: number
  active_sessions: number
}

export interface AdminDashboardRecentUser {
  id: number
  username: string
  nickname: string
  email: string
  role: string
  status: number
  create_time: number
  last_login_time?: number | null
}

export interface AdminDashboardResponse {
  statistics: AdminDashboardStatistics
  recent_users: AdminDashboardRecentUser[]
}

export const adminDashboardApi = {
  // 获取仪表盘统计数据
  getStatistics() {
    return request.Get<Service.ResponseResult<AdminDashboardResponse>>(BASE_URL)
  },
}
