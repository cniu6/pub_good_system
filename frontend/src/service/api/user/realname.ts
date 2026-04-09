/**
 * 用户端 API 服务 - 实名认证
 */
import { request } from '../../http'

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
