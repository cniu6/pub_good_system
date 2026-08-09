/**
 * 管理端 API 服务 - 货币/汇率管理
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() {
  return `${getAdminApiBase()}/payment/currency`
}

/** 汇率记录 */
export interface ExchangeRate {
  id: number
  from_currency: string
  to_currency: string
  rate: number
  fixed_amount: number
  rate_type: string
  source: string
  create_time: number
  update_time: number
}

/** 货币全局配置 */
export interface CurrencyConfig {
  base_currency: string
  currency_dynamic_source: string
  currency_dynamic_source_url: string
}

/** 获取本位币 */
export function getBaseCurrency() {
  return request.Get<Service.ResponseResult<{ base_currency: string }>>(`${baseUrl()}/base`)
}

/** 设置本位币 */
export function setBaseCurrency(currency: string) {
  return request.Post<Service.ResponseResult<null>>(`${baseUrl()}/base`, { currency })
}

/** 获取货币全局配置 */
export function getCurrencyConfig() {
  return request.Get<Service.ResponseResult<CurrencyConfig>>(`${baseUrl()}/config`)
}

/** 更新货币全局配置 */
export function updateCurrencyConfig(data: CurrencyConfig) {
  return request.Put<Service.ResponseResult<null>>(`${baseUrl()}/config`, data)
}

/** 列出汇率 */
export function fetchExchangeRates(params?: { from?: string, to?: string }) {
  return request.Get<Service.ResponseResult<{ list: ExchangeRate[] }>>(`${baseUrl()}/rates`, { params })
}

/** 创建汇率 */
export function createExchangeRate(data: Omit<ExchangeRate, 'id' | 'create_time' | 'update_time'>) {
  return request.Post<Service.ResponseResult<ExchangeRate>>(`${baseUrl()}/rates`, data)
}

/** 更新汇率 */
export function updateExchangeRate(id: number, data: Omit<ExchangeRate, 'id' | 'create_time' | 'update_time'>) {
  return request.Put<Service.ResponseResult<ExchangeRate>>(`${baseUrl()}/rates/${id}`, data)
}

/** 删除汇率 */
export function deleteExchangeRate(id: number) {
  return request.Delete<Service.ResponseResult<null>>(`${baseUrl()}/rates/${id}`)
}

/** 刷新单条汇率 */
export function refreshExchangeRate(id: number) {
  return request.Post<Service.ResponseResult<ExchangeRate>>(`${baseUrl()}/rates/${id}/refresh`)
}

/** 批量刷新汇率预览 */
export interface BatchRefreshPreviewItem {
  id: number
  from_currency: string
  to_currency: string
  old_rate: number
  new_rate: number
  source: string
  error?: string
}
export function batchRefreshExchangeRatesPreview(ids: number[]) {
  return request.Post<Service.ResponseResult<{ items: BatchRefreshPreviewItem[] }>>(`${baseUrl()}/rates/batch-refresh/preview`, { ids })
}

/** 批量确认刷新汇率 */
export function batchRefreshExchangeRates(ids: number[]) {
  return request.Post<Service.ResponseResult<{ items: BatchRefreshPreviewItem[] }>>(`${baseUrl()}/rates/batch-refresh`, { ids })
}

/** 刷新动态汇率 */
export function refreshExchangeRates() {
  return request.Post<Service.ResponseResult<{ rates: Record<string, number> }>>(`${baseUrl()}/rates/refresh`)
}

/** 管理端货币/汇率 API 聚合导出 */
export const adminCurrencyApi = {
  getBaseCurrency,
  setBaseCurrency,
  getCurrencyConfig,
  updateCurrencyConfig,
  fetchExchangeRates,
  createExchangeRate,
  updateExchangeRate,
  deleteExchangeRate,
  refreshExchangeRate,
  batchRefreshExchangeRatesPreview,
  batchRefreshExchangeRates,
  refreshExchangeRates,
}
