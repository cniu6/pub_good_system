import { request } from '@/service/http'
import { adminMoneyLogApi, adminScoreLogApi } from './user'
import { getAdminApiBase } from './base'

function usersBaseUrl() { return `${getAdminApiBase()}/users` }

interface ScoreChangeResponse {
  message: string
  log: Entity.UserScoreLog
}

interface MoneyOperateResponse {
  message: string
  result: unknown
}

export interface MoneyOperationPayload {
  money: number
  memo?: string
  operation: 'balance_only' | 'log_only' | 'order_only' | 'balance_log' | 'balance_order' | 'log_order' | 'both'
  order_no?: string
  trade_no?: string
  order_status?: number
}

export function operateUserMoney(userId: number, data: MoneyOperationPayload) {
  return request.Post<Service.ResponseResult<MoneyOperateResponse>>(`${usersBaseUrl()}/${userId}/money/operate`, data)
}

/** 后端生成订单号和交易号 */
export function generateNos() {
  return request.Get<Service.ResponseResult<{ order_no: string, trade_no: string }>>(`${getAdminApiBase()}/generate-nos`)
}

export function addScoreLog(userId: number, data: { score: number, memo?: string }) {
  return request.Post<Service.ResponseResult<ScoreChangeResponse>>(`${usersBaseUrl()}/${userId}/score/log`, data)
}

export function fetchAllMoneyLogs(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
  return adminMoneyLogApi.list(params)
}

export function fetchAllScoreLogs(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
  return adminScoreLogApi.list(params)
}

export interface WithdrawRecord {
  id: number
  user_id: number
  amount: number
  account_type: string
  account_name: string
  account_no: string
  real_name: string
  remark: string
  status: number
  review_remark: string
  transfer_remark: string
  reviewed_at: number | null
  reviewed_by: number | null
  paid_at: number | null
  paid_by: number | null
  create_time: number
  update_time: number
}

export interface WithdrawStats {
  pending_count: number
  approved_count: number
  rejected_count: number
  paid_count: number
  paid_amount: number
}

function withdrawUrl() { return `${getAdminApiBase()}/withdraw` }

function createIdempotencyKey(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

export function fetchWithdrawRecords(params: { page?: number, page_size?: number, keyword?: string, user_id?: number, status?: number }) {
  return request.Get<Service.ResponseResult<{ list: WithdrawRecord[], total: number, page: number, page_size: number }>>(withdrawUrl(), { params })
}

export function fetchWithdrawStats(params: { keyword?: string, user_id?: number, status?: number }) {
  return request.Get<Service.ResponseResult<WithdrawStats>>(`${withdrawUrl()}/stats`, { params })
}

export function fetchWithdrawDetail(id: number) {
  return request.Get<Service.ResponseResult<WithdrawRecord>>(`${withdrawUrl()}/${id}`)
}

export function reviewWithdraw(id: number, data: { status: 1 | 2, review_remark?: string }) {
  return request.Post<Service.ResponseResult<{ message: string }>>(`${withdrawUrl()}/${id}/review`, data, {
    headers: { 'X-Idempotency-Key': createIdempotencyKey(`withdraw-review-${id}`) },
  })
}

export function payWithdraw(id: number, data?: { transfer_remark?: string }) {
  return request.Post<Service.ResponseResult<{ message: string }>>(`${withdrawUrl()}/${id}/pay`, data || {}, {
    headers: { 'X-Idempotency-Key': createIdempotencyKey(`withdraw-pay-${id}`) },
  })
}
