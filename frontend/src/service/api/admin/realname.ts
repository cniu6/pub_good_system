/**
 * 管理端 API 服务 - 实名认证管理
 */
import { request } from '@/service/http'

const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/realname`

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

// 证件类型选项
export const certificateTypeOptions = [
  { label: '身份证', value: 1 },
  { label: '护照', value: 2 },
  { label: '军官证', value: 3 },
]

// 实名认证状态选项
export const realnameStatusOptions = [
  { label: '待审核', value: 0 },
  { label: '已通过', value: 1 },
  { label: '已拒绝', value: 2 },
]

export const adminRealnameApi = {
  list(params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: RealnameStatus
    user_id?: number
  }) {
    return request.Get<Service.ResponseResult<RealnameListResponse>>(BASE_URL, { params })
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<RealnameDetailResponse>>(`${BASE_URL}/${id}`)
  },

  review(data: {
    id: number
    status: 1 | 2
    reject_reason?: string
  }) {
    return request.Post<Service.ResponseResult<{}>>(
      `${BASE_URL}/review`,
      data,
    )
  },
}
