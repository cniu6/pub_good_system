/**
 * 管理端 API 服务 - 操作日志
 * 操作日志主要用于审计，提供分页浏览、详情、统计与清理功能
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() { return `${getAdminApiBase()}/logs` }

export interface OperationLogStats {
  total_count: number
  today_count: number
  success_count: number
  client_error_count: number
  server_error_count: number
  avg_duration: number
  top_modules: Array<{ module: string; count: number }>
  top_actions: Array<{ action: string; count: number }>
  method_stats: Array<{ method: string; count: number }>
}

export const adminLogApi = {
  /**
   * 获取日志列表（分页）
   */
  list(params?: { page?: number; page_size?: number; start_time?: number; end_time?: number }) {
    return request.Get<Service.ResponseResult<{ list: any[]; total: number; page: number; page_size: number }>>(baseUrl(), { params })
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<any>>(`${baseUrl()}/${id}`)
  },

  /** 操作日志详细统计（独立聚合，不受明细清理影响） */
  stats() {
    return request.Get<Service.ResponseResult<OperationLogStats>>(`${baseUrl()}/stats`)
  },

  /**
   * 清理日志
   * @param before_time 清理此时间戳之前的日志
   */
  clean(before_time: number) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/clean`, { before_time })
  },
}
