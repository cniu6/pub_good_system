import { request } from '../../http'

interface Ilogin {
  userName: string
  password: string
}

/** 用户登录 */
export function fetchLogin(data: Ilogin) {
  const methodInstance = request.Post<Service.ResponseResult<Api.Login.Info>>('/api/v1/public/login', data)
  methodInstance.meta = {
    authRole: null,
  }
  return methodInstance
}

/** 刷新Token */
export function fetchUpdateToken(data: any) {
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
  return request.Post<Service.ResponseResult<any>>('/api/v1/public/send-register-code', data)
}

/** 用户注册 */
export function fetchRegister(data: any) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/public/register', data)
}

/** 发送重置密码邮件 */
export function fetchSendResetEmail(data: { email: string, lang: string }) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/public/forgot-password', data)
}

/** 确认重置密码 */
export function fetchResetPasswordConfirm(data: any) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/public/reset-password', data)
}

/** 获取用户信息 */
export function fetchUserProfile() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/user/profile')
}

/** 获取当前用户 API Key */
export function fetchUserApiKey() {
  return request.Get<Service.ResponseResult<{ apikey: string | null }>>('/api/v1/user/apikey')
}

/** 更新用户信息 */
export function fetchUpdateProfile(data: any) {
  return request.Put<Service.ResponseResult<any>>('/api/v1/user/profile', data)
}

/** 修改密码 */
export function fetchChangePassword(data: { old_password: string, new_password: string }) {
  return request.Put<Service.ResponseResult<any>>('/api/v1/user/password', data)
}

/** 重置API密钥 */
export function fetchResetApiKey() {
  return request.Post<Service.ResponseResult<{ apikey: string }>>('/api/v1/user/resetapikey')
}
