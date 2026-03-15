import { request } from '@/service/http'
import { adminMoneyLogApi, adminScoreLogApi } from './user'

const ADMIN_PATH = '/admin'

// 余额/积分“仅写日志”接口（后端新增）
const USERS_BASE_URL = `/api/v1${ADMIN_PATH}/users`

export function updateUserMoney(userId: number, data: { money: number, memo?: string }) {
  return request.Put<Service.ResponseResult<any>>(`${USERS_BASE_URL}/${userId}/money`, data)
}

export function updateUserScore(userId: number, data: { score: number, memo?: string }) {
  return request.Put<Service.ResponseResult<any>>(`${USERS_BASE_URL}/${userId}/score`, data)
}

export function addMoneyLog(userId: number, data: { money: number, memo?: string }) {
  return request.Post<Service.ResponseResult<any>>(`${USERS_BASE_URL}/${userId}/money/log`, data)
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
  return request.Post<Service.ResponseResult<any>>(`${USERS_BASE_URL}/${userId}/money/operate`, data)
}

/** 后端生成订单号和交易号 */
export function generateNos() {
  return request.Get<Service.ResponseResult<{ order_no: string; trade_no: string }>>(`/api/v1${ADMIN_PATH}/generate-nos`)
}

export function addScoreLog(userId: number, data: { score: number, memo?: string }) {
  return request.Post<Service.ResponseResult<any>>(`${USERS_BASE_URL}/${userId}/score/log`, data)
}

export function fetchAllMoneyLogs(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
  return adminMoneyLogApi.list(params)
}

export function fetchAllScoreLogs(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
  return adminScoreLogApi.list(params)
}

export function deleteMoneyRecord(id: number) {
  return adminMoneyLogApi.delete(id)
}

export function deleteScoreRecord(id: number) {
  return adminScoreLogApi.delete(id)
}

// 当前项目暂无提现系统：先做空实现，保证页面可编译/可用
export async function fetchWithdrawRecords(_params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
  return {
    isSuccess: true,
    code: 200,
    message: 'ok',
    data: { list: [], total: 0 },
  } as const
}

export async function deleteWithdrawRecord(_id: number) {
  return {
    isSuccess: false,
    code: 501,
    message: 'withdraw not implemented',
    data: null,
  } as const
}
