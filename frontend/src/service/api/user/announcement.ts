/**
 * 用户端站内公告 API
 */
import { request } from '../../http'

export interface UserAnnouncementItem {
  id: number
  title: string
  content: string
  summary?: string
  type: string
  popup?: number
  published_at?: number
  is_read: boolean
}

export const userAnnouncementApi = {
  list(params?: { unread_only?: boolean | string, popup?: boolean | string, limit?: number }) {
    return request.Get<Service.ResponseResult<{ list: UserAnnouncementItem[], enabled: boolean }>>(
      '/api/v1/user/announcements',
      { params },
    )
  },
  detail(id: number) {
    return request.Get<Service.ResponseResult<UserAnnouncementItem>>(`/api/v1/user/announcements/${id}`)
  },
  markRead(id: number) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`/api/v1/user/announcements/${id}/read`, {})
  },
  markAllRead() {
    return request.Post<Service.ResponseResult<{ message: string }>>('/api/v1/user/announcements/read-all', {})
  },
  unreadCount() {
    return request.Get<Service.ResponseResult<{ count: number, enabled: boolean }>>(
      '/api/v1/user/announcements/unread-count',
    )
  },
}
