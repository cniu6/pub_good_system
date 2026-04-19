import { request } from '@/service/http'

const BASE_URL = '/api/v1/admin/email-logs'

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
    return request.Get<Service.ResponseResult<{ list: EmailLog[]; total: number; page: number; page_size: number }>>(BASE_URL, { params })
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<EmailLog>>(`${BASE_URL}/${id}`)
  },

  stats() {
    return request.Get<Service.ResponseResult<EmailLogStats>>(`${BASE_URL}/stats`)
  },

  templateNames() {
    return request.Get<Service.ResponseResult<string[]>>(`${BASE_URL}/template-names`)
  },

  clean(before: string) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${BASE_URL}/clean`, { before })
  },
}
