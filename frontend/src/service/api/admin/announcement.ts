/**
 * 管理端站内公告 API
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

function baseUrl() {
  return `${getAdminApiBase()}/announcements`
}

export interface AdminAnnouncement {
  id: number
  title: string
  summary: string
  content: string
  type: string
  status: number
  priority: number
  popup: number
  target_type: string
  target_value: string
  start_at: number
  end_at: number
  published_at: number
  created_by: number
  updated_by: number
  created_at: number
  updated_at: number
  deleted_at: number
}

export interface AnnouncementUpsertPayload {
  title: string
  summary?: string
  content: string
  type?: string
  priority?: number
  popup?: number
  // 注：产品定调公告面向全体登录用户，不做管理员/用户分层定向，服务端固定写 target_type=all，
  // 创建/编辑接口不再接收 target_type/target_value（传了也会被忽略）。
  start_at?: number
  end_at?: number
}

export const adminAnnouncementApi = {
  list(params?: { page?: number, page_size?: number, status?: number, type?: string, keyword?: string }) {
    return request.Get<Service.ResponseResult<{
      list: AdminAnnouncement[]
      total: number
      page: number
      page_size: number
    }>>(baseUrl(), { params })
  },
  detail(id: number) {
    return request.Get<Service.ResponseResult<AdminAnnouncement>>(`${baseUrl()}/${id}`)
  },
  create(data: AnnouncementUpsertPayload) {
    return request.Post<Service.ResponseResult<AdminAnnouncement>>(baseUrl(), data)
  },
  update(id: number, data: AnnouncementUpsertPayload) {
    return request.Put<Service.ResponseResult<AdminAnnouncement>>(`${baseUrl()}/${id}`, data)
  },
  publish(id: number) {
    return request.Post<Service.ResponseResult<AdminAnnouncement>>(`${baseUrl()}/${id}/publish`, {})
  },
  unpublish(id: number) {
    return request.Post<Service.ResponseResult<AdminAnnouncement>>(`${baseUrl()}/${id}/unpublish`, {})
  },
  remove(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/${id}`)
  },
}
