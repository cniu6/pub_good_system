/**
 * 管理端 API 服务 - 实名认证管理
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

// 选项常量统一从 constants/realname 导出，避免与用户端重复
export {
  getCertificateTypeOptions,
  getRealnameStatusOptions,
  certificateTypeOptions,
  realnameStatusOptions,
} from '@/constants/realname'

function baseUrl() { return `${getAdminApiBase()}/realname` }

// 实名认证状态
export type RealnameStatus = 0 | 1 | 2

export interface RealnameVerification {
  id: number
  user_id: number
  real_name: string
  certificate_type: 1 | 2 | 3
  certificate_no: string
  certificate_front: string
  certificate_back: string
  status: RealnameStatus
  reject_reason: string
  submitted_at: number | null
  reviewed_at: number | null
  reviewed_by: number | null
  create_time: number | null
  update_time: number | null
}

export interface RealnameListResponse {
  list: RealnameVerification[]
  total: number
  page: number
  page_size: number
}

export interface RealnameDetailResponse {
  verification: RealnameVerification
}

export const adminRealnameApi = {
  list(params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: RealnameStatus
    user_id?: number
  }) {
    return request.Get<Service.ResponseResult<RealnameListResponse>>(baseUrl(), { params })
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<RealnameDetailResponse>>(`${baseUrl()}/${id}`)
  },

  review(data: {
    id: number
    status: 1 | 2
    reject_reason?: string
  }) {
    return request.Post<Service.ResponseResult<{}>>(
      `${baseUrl()}/review`,
      data,
    )
  },
}
