/**
 * 管理端财务审批 API（双人复核）
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

function base() {
  return `${getAdminApiBase()}/approvals`
}

function createIdempotencyKey(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

export interface ApprovalRequestItem {
  id: number
  type: string
  payload_json: string
  status: string
  requester_id: number
  reviewer_id?: number | null
  comment: string
  create_time: number
  review_time?: number | null
}

export const adminApprovalApi = {
  listPending() {
    return request.Get<Service.ResponseResult<{ list: ApprovalRequestItem[] }>>(`${base()}/pending`)
  },
  approve(id: number, data?: { comment?: string }) {
    return request.Post<Service.ResponseResult<{ message?: string }>>(`${base()}/${id}/approve`, data || {}, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`approval-approve-${id}`) },
    })
  },
  reject(id: number, data?: { comment?: string }) {
    return request.Post<Service.ResponseResult<{ message?: string }>>(`${base()}/${id}/reject`, data || {})
  },
}
