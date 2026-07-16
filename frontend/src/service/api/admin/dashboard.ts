/**
 * 管理端 API 服务 - 仪表盘
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() { return `${getAdminApiBase()}/dashboard` }

export interface AdminDashboardStatistics {
  total_users: number
  today_new_users: number
  today_active_users: number
  active_users_7d: number
  total_money_logs: number
  total_score_logs: number
  total_operation_logs: number
  today_operation_logs: number
  active_sessions: number
  total_payment_orders: number
  paid_payment_orders: number
  pending_payment_orders: number
  total_payment_amount: number
  today_payment_orders: number
  today_payment_amount: number
  month_payment_amount: number
  year_payment_amount: number
  total_user_balance: number
  pending_withdraw_count: number
  approved_withdraw_count: number
  paid_withdraw_count: number
  paid_withdraw_amount: number
  total_realname_requests: number
  pending_realname_count: number
  approved_realname_count: number
  rejected_realname_count: number
}

export interface AdminDashboardRecentUser {
  id: number
  username: string
  nickname: string
  email: string
  role: string
  status: number
  money: number
  total_paid_amount: number
  balance_paid_ratio: number
  create_time: number
  last_login_time?: number | null
}

export interface AdminDashboardTrendPoint {
  date: string
  label: string
  new_users: number
  active_users: number
  paid_orders: number
  paid_amount: number
  operation_logs: number
}

export interface AdminDashboardResponse {
  statistics: AdminDashboardStatistics
  recent_users: AdminDashboardRecentUser[]
  recent_login_users: AdminDashboardRecentUser[]
  trends: AdminDashboardTrendPoint[]
}

export const adminDashboardApi = {
  // 获取仪表盘统计数据
  getStatistics() {
    return request.Get<Service.ResponseResult<AdminDashboardResponse>>(baseUrl())
  },
}
