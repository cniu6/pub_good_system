import { request } from '../../http'

interface UserMoneyLogListResponse {
  list: Entity.UserMoneyLog[]
  total: number
}

interface UserScoreLogListResponse {
  list: Entity.UserScoreLog[]
  total: number
}

// ========================================
// 用户设置
// ========================================

/** 获取用户设置 */
export function fetchUserSettings() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/user/settings')
}

/** 更新用户设置 */
export function updateUserSettings(data: { language?: string, theme?: string, notify_email?: boolean }) {
  return request.Put<Service.ResponseResult<any>>('/api/v1/user/settings', data)
}

// ========================================
// 用户统计
// ========================================

/** 获取用户统计 */
export function fetchUserStats() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/user/stats')
}

// ========================================
// 邮箱变更
// ========================================

/** 发送修改邮箱验证码 */
export function sendEmailChangeCode(data: { new_email: string, lang?: string }) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/user/email/send-code', data)
}

/** 验证并修改邮箱 */
export function verifyEmailChange(data: { new_email: string, code: string }) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/user/email/verify', data)
}

// ========================================
// 手机变更
// ========================================

/** 发送修改手机号验证码 */
export function sendPhoneChangeCode(data: { new_mobile: string }) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/user/phone/send-code', data)
}

/** 验证并修改手机号 */
export function verifyPhoneChange(data: { new_mobile: string, code: string }) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/user/phone/verify', data)
}

// ========================================
// 账号注销
// ========================================

/** 注销账号 */
export function deactivateAccount(data: { password: string, reason?: string }) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/user/deactivate', data)
}

// ========================================
// 会话管理
// ========================================

/** 获取登录会话列表 */
export function fetchUserSessions() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/user/sessions')
}

/** 踢出指定会话 */
export function revokeSession(sessionId: number | string) {
  return request.Delete<Service.ResponseResult<any>>(`/api/v1/user/sessions/${sessionId}`)
}

/** 踢出所有其他会话 */
export function revokeAllSessions() {
  return request.Post<Service.ResponseResult<any>>('/api/v1/user/sessions/revoke-all')
}

// ========================================
// 余额/积分日志
// ========================================

/** 获取我的余额变动日志 */
export function fetchMyMoneyLogs(params: { page?: number, page_size?: number, keyword?: string }) {
  return request.Get<Service.ResponseResult<UserMoneyLogListResponse>>('/api/v1/user/money-logs', { params })
}

/** 获取我的积分变动日志 */
export function fetchMyScoreLogs(params: { page?: number, page_size?: number, keyword?: string }) {
  return request.Get<Service.ResponseResult<UserScoreLogListResponse>>('/api/v1/user/score-logs', { params })
}

// ========================================
// 用户仪表盘
// ========================================

/** 获取用户仪表盘数据 */
export function fetchDashboard() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/user/dashboard')
}
