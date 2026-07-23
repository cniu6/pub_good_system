/**
 * 管理端 API 服务 - 用户管理
 * 此文件会被打包到 admin-api chunk
 */
import { authStorage } from '@/utils'
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

// 管理端 API base：运行时 app-config.admin_api_path，回退 VITE_ADMIN_API_PATH（默认 /admin）
function baseUrl() { return `${getAdminApiBase()}/users` }

/** 生成幂等键：后端资金/积分写接口强制要求 X-Idempotency-Key */
function createIdempotencyKey(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

export interface AdminUser {
  id: number
  group_id: number
  username: string
  nickname: string
  email: string
  mobile: string
  avatar: string
  back_ground: string
  gender: number
  birthday?: number | null
  money: number
  score: number
  level: number
  role: string
  last_login_time?: number | null
  last_login_ip: string
  login_failure: number
  join_ip: string
  join_time?: number | null
  motto: string
  admin_remark?: string
  status: number
  apikey?: string | null
  update_time?: number | null
  create_time?: number | null
  language: string
  country: string
  token: string
  realname_status?: 0 | 1 | 2 | null
  total_paid_amount?: number
  balance_paid_ratio?: number
  realname?: Api.Login.RealnameSummary
  /** 最近一次会话心跳时间（跨全部设备取最大值），来自 user_sessions；无会话记录时为空 */
  last_seen_at?: number | null
  /** 当前是否在线（依据 last_seen_at 与在线心跳容忍窗口判定，口径与在线用户页一致） */
  is_online?: boolean
}

export interface AdminUserRealnameSummary {
  has_verification: boolean
  id?: number
  status?: 0 | 1 | 2
  real_name?: string
  certificate_type?: 1 | 2 | 3
  certificate_no?: string
  submitted_at?: number | null
  reviewed_at?: number | null
  reject_reason?: string
}

interface UserListResponse {
  list: AdminUser[]
  total: number
  page: number
  page_size: number
}

interface UserDetailResponse {
  user: AdminUser
  realname?: AdminUserRealnameSummary
}

interface LoginAsUserResponse {
  user: AdminUser
  token: string
  refreshToken?: string
  expiresAt?: number
  refreshExpiresAt?: number
}

interface ResetApiKeyResponse {
  apikey: string
}

interface UserMoneyChangeResponse {
  message: string
  log: Entity.UserMoneyLog
}

interface UserScoreChangeResponse {
  message: string
  log: Entity.UserScoreLog
}

interface AdminUserUpdatePayload {
  nickname?: string
  email?: string
  mobile?: string
  avatar?: string
  gender?: number | null
  birthday?: number | null
  motto?: string
  language?: string
  country?: string
  admin_remark?: string
  level?: number
  role?: string
  status?: number
}

/** 与后端对齐：仅 admin / user；历史 super 一律归一为 admin */
export function normalizeAdminUserRole(role?: string): Entity.RoleType {
  if (role === 'admin' || role === 'super') {
    return 'admin'
  }
  return 'user'
}

/**
 * @param sessionRole 本次会话实际生效的角色（对应登录令牌的 auth_guard），
 *   不传则回退到用户自身的 DB role；用于 login-as 场景避免「令牌是 user guard，
 *   但展示的角色却是被登录用户的真实 admin 身份」的不一致
 */
export function toLoginInfo(user: AdminUser, token: string, sessionRole?: Entity.RoleType): Api.Login.Info {
  return {
    id: user.id,
    userName: user.username,
    nickname: user.nickname,
    email: user.email,
    mobile: user.mobile,
    avatar: user.avatar,
    backGround: user.back_ground,
    gender: user.gender as 0 | 1 | 2,
    birthday: user.birthday ?? null,
    money: user.money,
    score: user.score,
    level: user.level,
    role: [sessionRole ?? normalizeAdminUserRole(user.role)],
    lastLoginTime: user.last_login_time ?? null,
    lastLoginIp: user.last_login_ip,
    loginFailure: user.login_failure,
    joinIp: user.join_ip,
    joinTime: user.join_time ?? null,
    motto: user.motto,
    status: user.status === 1 ? 1 : 0,
    apikey: user.apikey ?? null,
    language: user.language,
    country: user.country,
    token: user.token,
    updateTime: user.update_time ?? null,
    createTime: user.create_time ?? null,
    realname: user.realname,
    accessToken: token,
    refreshToken: '',
  }
}

export type LoginAsAuthGuard = 'user' | 'admin'

/**
 * @param authGuard 本次 login-as 令牌实际签发的 auth_guard；session 里的 role 必须与它一致，
 *   不能用被登录用户的 DB role（否则用「user guard」登录一个 DB role=admin 的用户时，
 *   前端会误判 role=admin 触发管理端跳转，而后端 token 其实只是 user guard，两边状态不一致）
 */
export function openLoginAsUserWindow(
  user: AdminUser,
  token: string,
  refreshToken?: string,
  expiresAt?: number,
  targetUrl = '/',
  authGuard: LoginAsAuthGuard = 'user',
) {
  const sessionRole: Entity.RoleType = authGuard === 'admin' ? 'admin' : 'user'
  return authStorage.openSessionWindow({
    accessToken: token,
    refreshToken,
    accessTokenExpiresAt: expiresAt,
    role: [sessionRole],
    userInfo: toLoginInfo(user, token, sessionRole),
  }, targetUrl)
}

// 用户简要信息类型
export interface UserSimpleInfo {
  id: number
  username: string
  nickname: string
  email: string
  role: string
  status: number
}

interface UserBatchSimpleInfoResponse {
  users: Record<number, UserSimpleInfo>
}

interface UserMoneyLogListResponse {
  list: Entity.UserMoneyLog[]
  total: number
}

interface UserScoreLogListResponse {
  list: Entity.UserScoreLog[]
  total: number
}

export const adminUserApi = {
  // 用户列表
  list(params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: number | null
    role?: string
    realname_status?: 0 | 1 | 2 | null
  }) {
    return request.Get<Service.ResponseResult<UserListResponse>>(baseUrl(), { params })
  },

  // 用户详情
  detail(id: number) {
    return request.Get<Service.ResponseResult<UserDetailResponse>>(`${baseUrl()}/${id}`)
  },

  // 创建用户
  create(data: {
    username: string
    password: string
    email: string
    nickname?: string
    mobile?: string
    language?: string
    country?: string
    admin_remark?: string
    level?: number
    role?: string
    status?: number
  }) {
    return request.Post<Service.ResponseResult<AdminUser>>(baseUrl(), data)
  },

  // 更新用户
  update(id: number, data: AdminUserUpdatePayload) {
    return request.Put<Service.ResponseResult<null>>(`${baseUrl()}/${id}`, data)
  },

  // 删除用户
  delete(id: number) {
    return request.Delete<Service.ResponseResult<null>>(`${baseUrl()}/${id}`)
  },

  // 更新用户状态
  updateStatus(id: number, status: number) {
    return request.Put<Service.ResponseResult<null>>(`${baseUrl()}/${id}/status`, { status })
  },

  // 重置用户密码
  resetPassword(id: number, password: string) {
    return request.Put<Service.ResponseResult<null>>(`${baseUrl()}/${id}/password`, { password })
  },

  // 批量获取用户简要信息
  // 返回 map[id]UserSimpleInfo，方便通过 ID 快速查找
  async batchSimpleInfo(ids: number[]): Promise<Record<number, UserSimpleInfo>> {
    if (!ids.length)
      return {}
    const res = await request.Post<Service.ResponseResult<UserBatchSimpleInfoResponse>>(`${baseUrl()}/batch-simple`, { ids })
    return res.isSuccess ? (res.data?.users || {}) : {}
  },

  // 按标识查找用户（ID/用户名/邮箱）
  lookup(keyword: string) {
    return request.Get<Service.ResponseResult<UserDetailResponse>>(`${baseUrl()}/lookup`, { params: { keyword } })
  },

  // 管理员登录指定用户（生成该用户的JWT token）
  // auth_guard: user=用户前端，admin=管理后台（仅目标为管理员时可用）
  loginAsUser(id: number, data?: { auth_guard?: LoginAsAuthGuard }) {
    return request.Post<Service.ResponseResult<LoginAsUserResponse>>(`${baseUrl()}/${id}/login-as`, data || {})
  },

  // 重置指定用户的 API Key
  resetApiKey(id: number) {
    return request.Post<Service.ResponseResult<ResetApiKeyResponse>>(`${baseUrl()}/${id}/reset-apikey`)
  },

  // 变更用户余额（增减）；启用 TOTP 时传 totpCode
  changeMoney(id: number, data: { money: number, memo?: string }, totpCode?: string) {
    const headers: Record<string, string> = { 'X-Idempotency-Key': createIdempotencyKey(`money-change-${id}`) }
    if (totpCode)
      headers['X-Totp-Code'] = totpCode
    return request.Post<Service.ResponseResult<UserMoneyChangeResponse>>(`${baseUrl()}/${id}/money/change`, data, {
      headers,
    })
  },

  // 直接设置用户余额
  setMoney(id: number, data: { money: number, memo?: string }, totpCode?: string) {
    const headers: Record<string, string> = { 'X-Idempotency-Key': createIdempotencyKey(`money-set-${id}`) }
    if (totpCode)
      headers['X-Totp-Code'] = totpCode
    return request.Put<Service.ResponseResult<UserMoneyChangeResponse>>(`${baseUrl()}/${id}/money`, data, {
      headers,
    })
  },

  // 变更用户积分（增减）
  changeScore(id: number, data: { score: number, memo?: string }) {
    return request.Post<Service.ResponseResult<UserScoreChangeResponse>>(`${baseUrl()}/${id}/score/change`, data, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`score-change-${id}`) },
    })
  },

  // 直接设置用户积分
  setScore(id: number, data: { score: number, memo?: string }) {
    return request.Put<Service.ResponseResult<UserScoreChangeResponse>>(`${baseUrl()}/${id}/score`, data, {
      headers: { 'X-Idempotency-Key': createIdempotencyKey(`score-set-${id}`) },
    })
  },
}

// 余额日志管理 API
function moneyLogsUrl() { return `${getAdminApiBase()}/money-logs` }

export const adminMoneyLogApi = {
  list(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
    return request.Get<Service.ResponseResult<UserMoneyLogListResponse>>(moneyLogsUrl(), { params })
  },
  detail(id: number) {
    return request.Get<Service.ResponseResult<Entity.UserMoneyLog>>(`${moneyLogsUrl()}/${id}`)
  },
  delete(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${moneyLogsUrl()}/${id}`)
  },
}

// 积分日志管理 API
function scoreLogsUrl() { return `${getAdminApiBase()}/score-logs` }

export const adminScoreLogApi = {
  list(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
    return request.Get<Service.ResponseResult<UserScoreLogListResponse>>(scoreLogsUrl(), { params })
  },
  detail(id: number) {
    return request.Get<Service.ResponseResult<Entity.UserScoreLog>>(`${scoreLogsUrl()}/${id}`)
  },
  delete(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${scoreLogsUrl()}/${id}`)
  },
}

// =====================================================
// 兼容资料/vue 的命名导出（给 views/admin/users 直接引用）
// =====================================================

export function fetchAdminUserPage(params: {
  page?: number
  page_size?: number
  keyword?: string
  status?: number | null
  role?: string
  realname_status?: 0 | 1 | 2 | null
}) {
  return adminUserApi.list(params)
}

export function createUser(data: Parameters<typeof adminUserApi.create>[0]) {
  return adminUserApi.create(data)
}

export function deleteUser(userId: number) {
  return adminUserApi.delete(userId)
}

export function updateUserStatus(userId: number, data: { status: number }) {
  return adminUserApi.updateStatus(userId, Number(data.status))
}

export function updateAdminUserProfile(userId: number, data: AdminUserUpdatePayload) {
  return adminUserApi.update(userId, data)
}

export function loginAsUser(userId: number, data?: { auth_guard?: LoginAsAuthGuard }) {
  return adminUserApi.loginAsUser(userId, data)
}

export function resetUserApikey(userId: number) {
  return adminUserApi.resetApiKey(userId)
}

export function resetUserPassword(
  arg1: number | { user_id: number, password: string },
  arg2?: { password: string },
) {
  if (typeof arg1 === 'number')
    return adminUserApi.resetPassword(arg1, arg2?.password || '')
  return adminUserApi.resetPassword(arg1.user_id, arg1.password)
}
