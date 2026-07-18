/**
 * 用户端 API 服务 - 实名认证
 */
import { request } from '../../http'

// 选项常量统一从 constants/realname 导出，避免与管理端重复
export {
  getCertificateTypeOptions,
  getRealnameStatusOptions,
  certificateTypeOptions,
  realnameStatusOptions,
} from '@/constants/realname'

const BASE_URL = '/api/v1/user/realname'

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
