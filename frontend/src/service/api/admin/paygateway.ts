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
  target_currency: string
  exchange_rate_mode: string
  exchange_rate: number
  exchange_fixed_amount: number
  exchange_rate_source: string
  target_fee_rate: number
  target_fee_fixed: number
  target_fee_mode: string
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
  target_currency?: string
  exchange_rate_mode?: string
  exchange_rate?: number
  exchange_fixed_amount?: number
  exchange_rate_source?: string
  target_fee_rate?: number
  target_fee_fixed?: number
  target_fee_mode?: string
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

/** 测试支付通道连接 */
export function testPayGatewayConnection(id: number) {
  return request.Post<Service.ResponseResult<{ success: boolean, message: string }>>(`${baseUrl()}/${id}/test-connection`)
}

/** 获取已注册支付通道元数据（版本、签名算法、配置字段） */
export function fetchPaymentChannelMetas() {
  return request.Get<Service.ResponseResult<ChannelMeta[]>>(`${getAdminApiBase()}/payment/channels/metas`)
}

/** 全局汇率 */
export interface ExchangeRate {
  id: number
  from_currency: string
  to_currency: string
  rate: number
  rate_type: string
  source: string
  create_time: number
  update_time: number
}

/** 设置本位币 */
export function setBaseCurrency(currency: string) {
  return request.Post<Service.ResponseResult<null>>(`${getAdminApiBase()}/payment/currency/base`, { currency })
}

/** 获取本位币 */
export function getBaseCurrency() {
  return request.Get<Service.ResponseResult<{ base_currency: string }>>(`${getAdminApiBase()}/payment/currency/base`)
}

/** 列出汇率 */
export function fetchExchangeRates(params?: { from?: string, to?: string }) {
  return request.Get<Service.ResponseResult<{ list: ExchangeRate[] }>>(`${getAdminApiBase()}/payment/currency/rates`, { params })
}

/** 创建/更新汇率 */
export function createExchangeRate(data: Omit<ExchangeRate, 'id' | 'create_time' | 'update_time'>) {
  return request.Post<Service.ResponseResult<ExchangeRate>>(`${getAdminApiBase()}/payment/currency/rates`, data)
}

/** 删除汇率 */
export function deleteExchangeRate(id: number) {
  return request.Delete<Service.ResponseResult<null>>(`${getAdminApiBase()}/payment/currency/rates/${id}`)
}

/** 刷新动态汇率 */
export function refreshExchangeRates() {
  return request.Post<Service.ResponseResult<{ rates: Record<string, number> }>>(`${getAdminApiBase()}/payment/currency/rates/refresh`)
}
