/**
 * 管理端 API 服务 - 操作日志
 * 操作日志主要用于审计，不需要复杂查询，只提供分页浏览和清理功能
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() { return `${getAdminApiBase()}/logs` }

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

  /**
   * 清理日志
   * @param before_time 清理此时间戳之前的日志
   */
  clean(before_time: number) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/clean`, { before_time })
  },
}
