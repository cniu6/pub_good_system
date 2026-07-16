import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() { return `${getAdminApiBase()}/email-logs` }

export interface EmailLog {
  id: number
  to_email: string
  subject: string
  content?: string
  template_name: string
  status: number
  error_msg: string
  created_at: string
}

export interface EmailLogListParams {
  page?: number
  page_size?: number
  to_email?: string
  template_name?: string
  status?: number
  start_time?: string
  end_time?: string
}

export interface EmailLogStats {
  total: number
  success: number
  fail: number
}

export const adminEmailLogApi = {
  list(params?: EmailLogListParams) {
    return request.Get<Service.ResponseResult<{ list: EmailLog[]; total: number; page: number; page_size: number }>>(baseUrl(), { params })
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<EmailLog>>(`${baseUrl()}/${id}`)
  },

  stats() {
    return request.Get<Service.ResponseResult<EmailLogStats>>(`${baseUrl()}/stats`)
  },

  templateNames() {
    return request.Get<Service.ResponseResult<string[]>>(`${baseUrl()}/template-names`)
  },

  clean(before: string) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/clean`, { before })
  },
}
