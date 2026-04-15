/**
 * 用户端 API 服务 - 实名认证
 */
import { i18n } from '@/modules/i18n'
import { request } from '../../http'

const BASE_URL = '/api/v1/user/realname'

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

// 证件类型
export type CertificateType = 1 | 2 | 3

export interface RealnameStatusResponse {
  hasVerification: boolean
  id?: number
  status?: RealnameStatus
  realName?: string
  certificateType?: CertificateType
  certificateNo?: string
  certificateFront?: string
  certificateBack?: string
  rejectReason?: string
  submittedAt?: number | null
  reviewedAt?: number | null
}

export interface SubmitRealnameRequest {
  real_name: string
  certificate_type: CertificateType
  certificate_no: string
  certificate_front: string
  certificate_back: string
}

// 证件类型选项（向后兼容，推荐使用 getCertificateTypeOptions 函数）
export const certificateTypeOptions = getCertificateTypeOptions()

// 实名认证状态选项（向后兼容，推荐使用 getRealnameStatusOptions 函数）
export const realnameStatusOptions = getRealnameStatusOptions()

/**
 * 获取我的实名认证状态
 */
export function fetchMyRealnameStatus() {
  return request.Get<Service.ResponseResult<RealnameStatusResponse>>(BASE_URL)
}

/**
 * 提交实名认证
 */
export function submitRealname(data: SubmitRealnameRequest) {
  return request.Post<Service.ResponseResult<{ message?: string }>>(BASE_URL, data)
}
