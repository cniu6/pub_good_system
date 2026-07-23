import { request } from '../../http'
import { getAdminApiBase } from './base'

// ========================================
// 类型定义
// ========================================

export interface PaymentOrder {
  id: number
  order_no: string
  user_id: number
  trade_no: string
  payment_channel: string
  payment_type: string
  amount: number
  subject: string
  status: number
  notify_count: number
  paid_at: number | null
  expire_at: number
  client_ip: string
  create_time: number
  update_time: number
}

interface PaymentOrderListResponse {
  list: PaymentOrder[]
  total: number
}

export interface PaymentStats {
  total_orders: number
  paid_orders: number
  total_amount: number
  today_orders: number
  today_amount: number
  pending_orders: number
}

// ========================================
// 管理端支付 API
// ========================================

function baseUrl() { return `${getAdminApiBase()}/payment` }

/** 补单/取消订单写接口需要幂等键 */
function createIdempotencyKey(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

export const adminPaymentApi = {
  /** 订单列表 */
  listOrders(params: { page?: number, page_size?: number, status?: number, user_id?: number, keyword?: string }) {
    return request.Get<Service.ResponseResult<PaymentOrderListResponse>>(`${baseUrl()}/orders`, { params })
  },

  /** 订单详情 */
  orderDetail(id: number) {
    return request.Get<Service.ResponseResult<PaymentOrder>>(`${baseUrl()}/orders/${id}`)
  },

  /** 手动补单（force=true 可对取消/失败单强制补单，须填 memo；启用 TOTP 时传 totpCode） */
  completeOrder(id: number, data?: { memo?: string, force?: boolean }, totpCode?: string) {
    const headers: Record<string, string> = { 'X-Idempotency-Key': createIdempotencyKey(`payment-complete-${id}`) }
    if (totpCode)
      headers['X-Totp-Code'] = totpCode
    return request.Post<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/orders/${id}/complete`, data || {}, {
      headers,
    })
  },

  /** 单笔主动对账 */
  reconcileOrder(id: number) {
    return request.Post<Service.ResponseResult<{ changed: boolean, order: PaymentOrder }>>(`${baseUrl()}/orders/${id}/reconcile`, {}, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`payment-reconcile-${id}`) },
    })
  },

  /** 取消订单 */
  cancelOrder(id: number) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/orders/${id}/cancel`, {}, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`payment-cancel-${id}`) },
    })
  },

  /** 删除订单 */
  deleteOrder(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/orders/${id}`)
  },

  /** 支付统计 */
  getStats() {
    return request.Get<Service.ResponseResult<PaymentStats>>(`${baseUrl()}/stats`)
  },

  /** 支付异常列表 */
  listExceptions(params: { page?: number, page_size?: number, status?: number, exception_type?: string, order_no?: string, user_id?: number }) {
    return request.Get<Service.ResponseResult<{ list: PaymentException[], total: number }>>(`${baseUrl()}/exceptions`, { params })
  },

  /** 处理/忽略异常 */
  resolveException(id: number, data: { action: 'resolve' | 'ignore', remark?: string }) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/exceptions/${id}/resolve`, data, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`payment-exception-${id}`) },
    })
  },
}

export interface PaymentException {
  id: number
  order_no: string
  user_id: number
  gateway_id: number
  exception_type: string
  status: number
  source: string
  message: string
  detail: string
  order_status: number
  trade_no: string
  resolved_by: number
  resolved_at: number | null
  resolve_remark: string
  create_time: number
  update_time: number
}
