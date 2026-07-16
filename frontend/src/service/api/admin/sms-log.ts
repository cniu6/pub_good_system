/**
 * Admin SMS log API
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() { return `${getAdminApiBase()}/sms-logs` }

export interface SMSLog {
  id: number
  phone: string
  provider: string
  template_code: string
  template_name: string
  lang: string
  content: string
  status: number
  error_msg: string
  request_id: string
  response: string
  created_at: string
}

export interface SMSLogListParams {
  page?: number
  page_size?: number
  phone?: string
  provider?: string
  template_name?: string
  lang?: string
  status?: number
  start_time?: string
  end_time?: string
}

export const adminSMSLogApi = {
  list(params?: SMSLogListParams) {
    return request.Get<Service.ResponseResult<{ list: SMSLog[]; total: number; page: number; page_size: number }>>(baseUrl(), { params })
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<SMSLog>>(`${baseUrl()}/${id}`)
  },

  stats() {
    return request.Get<Service.ResponseResult<{ total: number; success: number; fail: number }>>(`${baseUrl()}/stats`)
  },

  templateNames() {
    return request.Get<Service.ResponseResult<string[]>>(`${baseUrl()}/template-names`)
  },

  clean(before: string) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/clean`, { before })
  },
}
