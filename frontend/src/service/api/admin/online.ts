import { request } from '@/service/http'
import { getAdminApiBase } from './base'

/** 单条在线设备会话 */
export interface OnlineSession {
  id: number
  user_id: number
  username?: string
  nickname?: string
  auth_guard: string
  ip: string
  user_agent: string
  device: string
  client_type: string
  is_active: boolean
  is_online: boolean
  is_current: boolean
  login_at: number
  last_seen_at: number
  expires_at: number
  created_at: number
}

/** 管理端在线列表：同一用户+登录端归为一行，多设备挂在 devices */
export interface OnlineUserRow {
  user_id: number
  username?: string
  nickname?: string
  auth_guard: string
  device_count: number
  last_seen_at: number
  is_online: boolean
  devices: OnlineSession[]
}

export interface OnlineSessionListParams {
  page?: number
  page_size?: number
  keyword?: string
  client_type?: string
  auth_guard?: string
}

function baseUrl() {
  return `${getAdminApiBase()}/online`
}

export const adminOnlineApi = {
  stats() {
    return request.Get<Service.ResponseResult<{ online_users: number, online_sessions: number }>>(`${baseUrl()}/stats`)
  },

  sessions(params?: OnlineSessionListParams) {
    return request.Get<Service.ResponseResult<{ list: OnlineUserRow[], total: number, page: number, page_size: number }>>(
      `${baseUrl()}/sessions`,
      { params },
    )
  },

  kick(sessionId: number | string) {
    return request.Delete<Service.ResponseResult<{ message?: string }>>(`${baseUrl()}/sessions/${sessionId}`)
  },

  userSessions(userId: number | string, authGuard = 'user') {
    return request.Get<Service.ResponseResult<OnlineSession[]>>(
      `${getAdminApiBase()}/users/${userId}/sessions`,
      { params: { auth_guard: authGuard } },
    )
  },

  revokeAllUserSessions(userId: number | string, authGuard = 'user') {
    return request.Post<Service.ResponseResult<{ message?: string, count?: number }>>(
      `${getAdminApiBase()}/users/${userId}/sessions/revoke-all`,
      { auth_guard: authGuard },
    )
  },
}
