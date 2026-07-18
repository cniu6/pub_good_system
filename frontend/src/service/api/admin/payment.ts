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

  /** 手动补单 */
  completeOrder(id: number, data?: { memo?: string }) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/orders/${id}/complete`, data || {}, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`payment-complete-${id}`) },
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
}
