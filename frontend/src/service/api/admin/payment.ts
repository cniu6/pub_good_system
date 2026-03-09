import { request } from '../../http'

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

const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/payment`

export const adminPaymentApi = {
  /** 订单列表 */
  listOrders(params: { page?: number, page_size?: number, status?: number, user_id?: number, keyword?: string }) {
    return request.Get<Service.ResponseResult<PaymentOrderListResponse>>(`${BASE_URL}/orders`, { params })
  },

  /** 订单详情 */
  orderDetail(id: number) {
    return request.Get<Service.ResponseResult<PaymentOrder>>(`${BASE_URL}/orders/${id}`)
  },

  /** 手动补单 */
  completeOrder(id: number, data?: { memo?: string }) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${BASE_URL}/orders/${id}/complete`, data || {})
  },

  /** 取消订单 */
  cancelOrder(id: number) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${BASE_URL}/orders/${id}/cancel`)
  },

  /** 删除订单 */
  deleteOrder(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${BASE_URL}/orders/${id}`)
  },

  /** 支付统计 */
  getStats() {
    return request.Get<Service.ResponseResult<PaymentStats>>(`${BASE_URL}/stats`)
  },
}
