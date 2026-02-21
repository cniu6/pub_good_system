import { request } from '../http'

interface Ilogin {
  userName: string
  password: string
}

export function fetchLogin(data: Ilogin) {
  const methodInstance = request.Post<Service.ResponseResult<Api.Login.Info>>('/api/v1/login', data)
  methodInstance.meta = {
    authRole: null,
  }
  return methodInstance
}
export function fetchUpdateToken(data: any) {
  const method = request.Post<Service.ResponseResult<Api.Login.Info>>('/api/v1/updateToken', data)
  method.meta = {
    authRole: 'refreshToken',
  }
  return method
}

export function fetchUserRoutes(params: { id: number }) {
  return request.Get<Service.ResponseResult<AppRoute.RowRoute[]>>('/api/v1/getUserRoutes', { params })
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
