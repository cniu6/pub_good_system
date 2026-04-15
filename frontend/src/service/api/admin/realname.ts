/**
 * 管理端 API 服务 - 实名认证管理
 */
import { i18n } from '@/modules/i18n'
import { request } from '@/service/http'

const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/realname`

// 获取证件类型选项（支持 i18n）
export function getCertificateTypeOptions() {
  return [
    { label: i18n.global.t('realname.idCard'), value: 1 },
    { label: i18n.global.t('realname.passport'), value: 2 },
    { label: i18n.global.t('realname.officer'), value: 3 },
  ]
}

// 获取实名认证状态选项（支持 i18n）
export function getRealnameStatusOptions() {
  return [
    { label: i18n.global.t('realname.pending'), value: 0 },
    { label: i18n.global.t('realname.approved'), value: 1 },
    { label: i18n.global.t('realname.rejected'), value: 2 },
  ]
}

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

// 证件类型选项（向后兼容，推荐使用 getCertificateTypeOptions 函数）
export const certificateTypeOptions = getCertificateTypeOptions()

// 实名认证状态选项（向后兼容，推荐使用 getRealnameStatusOptions 函数）
export const realnameStatusOptions = getRealnameStatusOptions()

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
