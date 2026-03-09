import { request } from '../../http'

// ========================================
// 类型定义
// ========================================

/** 支付通道 */
export interface PayGateway {
  id: number
  name: string
  type: string
  pay_type: string
  description: string
  status: number
  logo_url: string
  sort_order: number
  min_amount: number
  max_amount: number
  fee_rate: number
  fee_mode: string
  min_level: number
  create_time: number
  update_time: number
}

export interface CreateOrderResponse {
  order_no: string
  pay_url: string
  amount: number
  fee: number
  pay_amount: number
  expire_at: number
  gateway_name: string
  payment_type: string
}

export interface PaymentOrder {
  id: number
  order_no: string
  user_id: number
  gateway_id: number
  trade_no: string
  payment_channel: string
  payment_type: string
  amount: number
  fee: number
  pay_amount: number
  subject: string
  status: number
  notify_count: number
  pay_url: string
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

interface PayGatewayListResponse {
  list: PayGateway[]
}

interface OrderStatusResponse {
  order_no: string
  status: number
  paid_at: number | null
}

// ========================================
// 用户端支付 API
// ========================================

/** 获取可用支付通道列表 */
export function fetchPayGateways() {
  return request.Get<Service.ResponseResult<PayGatewayListResponse>>('/api/v1/user/payment/gateways')
}

/** 创建充值订单 */
export function createPaymentOrder(data: { gateway_id: number, amount: number, subject?: string }) {
  return request.Post<Service.ResponseResult<CreateOrderResponse>>('/api/v1/user/payment/create', data)
}

/** 获取充值订单列表 */
export function fetchPaymentOrders(params: { page?: number, page_size?: number, status?: number }) {
  return request.Get<Service.ResponseResult<PaymentOrderListResponse>>('/api/v1/user/payment/orders', { params })
}

/** 获取订单详情 */
export function fetchPaymentOrderDetail(id: number) {
  return request.Get<Service.ResponseResult<PaymentOrder>>(`/api/v1/user/payment/orders/${id}`)
}

/** 轮询订单支付状态 */
export function checkPaymentOrderStatus(id: number) {
  return request.Get<Service.ResponseResult<OrderStatusResponse>>(`/api/v1/user/payment/orders/${id}/status`)
}
