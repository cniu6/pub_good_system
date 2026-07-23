import { request } from '../../http'

interface Ilogin {
  userName: string
  password: string
  authGuard?: 'user' | 'admin'
}

interface ActionMessageResponse {
  message?: string
}

interface RegisterPayload {
  username: string
  password: string
  email: string
  code: string
  nickname?: string
  lang?: string
}

interface ResetPasswordConfirmPayload {
  email: string
  code: string
  new_password: string
}

interface UserProfileResponse extends Api.Login.Info {}

/** 用户登录 */
export function fetchLogin(data: Ilogin) {
  const methodInstance = request.Post<Service.ResponseResult<Api.Login.Info>>('/api/v1/public/login', data)
  methodInstance.meta = {
    authRole: null,
    allowDuringSessionRecovery: true,
  }
  return methodInstance
}

/** 刷新Token */
export function fetchUpdateToken(data: { refreshToken: string | null, authGuard?: 'user' | 'admin' }) {
  const method = request.Post<Service.ResponseResult<Api.Login.Info>>('/api/v1/public/refresh-token', data)
  method.meta = {
    authRole: 'refreshToken',
  }
  return method
}

/** 获取用户路由 */
export function fetchUserRoutes(params: { id: number }) {
  return request.Get<Service.ResponseResult<AppRoute.RowRoute[]>>('/api/v1/user/routes', { params })
}

/** 发送注册验证码 */
export function fetchSendRegisterCode(data: { email: string, lang: string }) {
  const methodInstance = request.Post<Service.ResponseResult<ActionMessageResponse>>('/api/v1/public/send-register-code', data)
  methodInstance.meta = {
    authRole: null,
    allowDuringSessionRecovery: true,
  }
  return methodInstance
}

/** 用户注册 */
export function fetchRegister(data: RegisterPayload) {
  const methodInstance = request.Post<Service.ResponseResult<ActionMessageResponse>>('/api/v1/public/register', data)
  methodInstance.meta = {
    authRole: null,
    allowDuringSessionRecovery: true,
  }
  return methodInstance
}

/** 发送重置密码邮件 */
export function fetchSendResetEmail(data: { email: string, lang: string }) {
  const methodInstance = request.Post<Service.ResponseResult<ActionMessageResponse>>('/api/v1/public/forgot-password', data)
  methodInstance.meta = {
    authRole: null,
    allowDuringSessionRecovery: true,
  }
  return methodInstance
}

/** 确认重置密码 */
export function fetchResetPasswordConfirm(data: ResetPasswordConfirmPayload) {
  const methodInstance = request.Post<Service.ResponseResult<ActionMessageResponse>>('/api/v1/public/reset-password', data)
  methodInstance.meta = {
    authRole: null,
    allowDuringSessionRecovery: true,
  }
  return methodInstance
}

/** 获取用户信息 */
export function fetchUserProfile() {
  return request.Get<Service.ResponseResult<UserProfileResponse>>('/api/v1/user/profile')
}

/** 获取当前用户 API Key（明文；前端自行用 password 眼睛显隐） */
export function fetchUserApiKey() {
  return request.Get<Service.ResponseResult<{ apikey: string | null }>>('/api/v1/user/apikey')
}

/** 更新用户信息 */
export function fetchUpdateProfile(data: {
  nickname?: string
  avatar?: string
  gender?: 0 | 1 | 2
  birthday?: number | null
  motto?: string
  mobile?: string
  back_ground?: string
  language?: string
  country?: string
}) {
  return request.Put<Service.ResponseResult<ActionMessageResponse>>('/api/v1/user/profile', data)
}

/** 修改密码 */
export function fetchChangePassword(data: { old_password: string, new_password: string }) {
  return request.Put<Service.ResponseResult<ActionMessageResponse>>('/api/v1/user/password', data)
}

/** 重置API密钥 */
export function fetchResetApiKey() {
  return request.Post<Service.ResponseResult<{ apikey: string }>>('/api/v1/user/resetapikey')
}
