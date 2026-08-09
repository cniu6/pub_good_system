/**
 * 管理端 API 服务 - 支付通道管理
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() {
  return `${getAdminApiBase()}/payment/gateways`
}

/** 配置字段 select 选项 */
export interface ChannelConfigFieldOption {
  value: string
  label: string
}

/** 配置字段 schema */
export interface ChannelConfigField {
  name: string
  label: string
  type: 'input' | 'textarea' | 'select'
  required: boolean
  secret: boolean
  placeholder: string
  options?: ChannelConfigFieldOption[]
}

/** 签名算法元数据 */
export interface ChannelSignTypeMeta {
  value: string
  name: string
}

/** 通道版本元数据 */
export interface ChannelVersionMeta {
  version: string
  name: string
  signTypes: ChannelSignTypeMeta[]
  configFields: ChannelConfigField[]
}

/** 支付方式元数据 */
export interface ChannelPayTypeMeta {
  value: string
  name: string
}

/** 设备类型元数据 */
export interface ChannelDeviceMeta {
  value: string
  name: string
}

/** 通道类型元数据 */
export interface ChannelMeta {
  type: string
  name: string
  currency: string
  pay_types: ChannelPayTypeMeta[]
  devices: ChannelDeviceMeta[]
  default_notify_path: string
  versions: ChannelVersionMeta[]
}

/** 支付通道 */
export interface PayGateway {
  id: number
  name: string
  type: string
  pay_type: string
  sign_type: string
  version: string
  device: string
  currency: string
  description: string
  status: number
  api_url: string
  pid: string
  key: string
  ext_config: string
  logo_url: string
  sort_order: number
  min_amount: number
  max_amount: number
  fee_rate: number
  fee_mode: string
  min_level: number
  notify_url: string
  expire_minutes: number
  active_query_enabled: number
  query_interval_seconds: number
  query_batch_size: number
  create_time: number
  update_time: number
}

export interface PayGatewayCreateRequest {
  name: string
  type: string
  pay_type: string
  sign_type?: string
  version?: string
  device?: string
  currency?: string
  description?: string
  status: number
  api_url?: string
  pid?: string
  key?: string
  ext_config?: string
  logo_url?: string
  sort_order?: number
  min_amount?: number
  max_amount?: number
  fee_rate?: number
  fee_mode?: string
  min_level?: number
  notify_url?: string
  expire_minutes?: number
  active_query_enabled?: number
  query_interval_seconds?: number
  query_batch_size?: number
}

export type PayGatewayUpdateRequest = Partial<PayGatewayCreateRequest>

interface PayGatewayListResponse {
  list: PayGateway[]
  total: number
}

/** 获取支付通道列表 */
export function fetchPayGateways(params?: { page?: number, page_size?: number, keyword?: string }) {
  return request.Get<Service.ResponseResult<PayGatewayListResponse>>(baseUrl(), { params })
}

/** 创建支付通道 */
export function createPayGateway(data: PayGatewayCreateRequest) {
  return request.Post<Service.ResponseResult<PayGateway>>(baseUrl(), data)
}

/** 更新支付通道 */
export function updatePayGateway(id: number, data: PayGatewayUpdateRequest) {
  return request.Put<Service.ResponseResult<PayGateway>>(`${baseUrl()}/${id}`, data)
}

/** 删除支付通道 */
export function deletePayGateway(id: number) {
  return request.Delete<Service.ResponseResult<null>>(`${baseUrl()}/${id}`)
}

/** 获取已注册支付通道元数据（版本、签名算法、配置字段） */
export function fetchPaymentChannelMetas() {
  return request.Get<Service.ResponseResult<ChannelMeta[]>>(`${getAdminApiBase()}/payment/channels/metas`)
}
