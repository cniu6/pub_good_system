/**
 * 管理端 API 服务 - 支付通道管理
 */
import { request } from '@/service/http'

const BASE_URL = '/api/v1/admin/payment/gateways'

/** 支付通道 */
export interface PayGateway {
  id: number
  name: string
  type: string
  pay_type: string
  description: string
  status: number
  api_url: string
  pid: string
  key: string
  logo_url: string
  sort_order: number
  min_amount: number
  max_amount: number
  fee_rate: number
  fee_mode: string
  min_level: number
  notify_url: string
  create_time: number
  update_time: number
}

export interface PayGatewayCreateRequest {
  name: string
  type: string
  pay_type: string
  description?: string
  status: number
  api_url?: string
  pid?: string
  key?: string
  logo_url?: string
  sort_order?: number
  min_amount?: number
  max_amount?: number
  fee_rate?: number
  fee_mode?: string
  min_level?: number
  notify_url?: string
}

export type PayGatewayUpdateRequest = Partial<PayGatewayCreateRequest>

interface PayGatewayListResponse {
  list: PayGateway[]
  total: number
}

/** 获取支付通道列表 */
export function fetchPayGateways(params?: { page?: number, page_size?: number, keyword?: string }) {
  return request.Get<Service.ResponseResult<PayGatewayListResponse>>(BASE_URL, { params })
}

/** 获取支付通道详情 */
export function fetchPayGatewayDetail(id: number) {
  return request.Get<Service.ResponseResult<PayGateway>>(`${BASE_URL}/${id}`)
}

/** 创建支付通道 */
export function createPayGateway(data: PayGatewayCreateRequest) {
  return request.Post<Service.ResponseResult<PayGateway>>(BASE_URL, data)
}

/** 更新支付通道 */
export function updatePayGateway(id: number, data: PayGatewayUpdateRequest) {
  return request.Put<Service.ResponseResult<PayGateway>>(`${BASE_URL}/${id}`, data)
}

/** 删除支付通道 */
export function deletePayGateway(id: number) {
  return request.Delete<Service.ResponseResult<null>>(`${BASE_URL}/${id}`)
}
