import { blankInstance, request } from '../http'

// ==================== Demo 请求示例 ====================

/** GET 请求示例 */
export function fetchGet(params: Record<string, any>) {
  return request.Get<Service.ResponseResult<any>>('/api/v1/demo/get', { params })
}

/** POST 请求示例 */
export function fetchPost(data: Record<string, any>) {
  return request.Post<Service.ResponseResult<any>>('/api/v1/demo/post', data)
}

/** PUT 请求示例 */
export function fetchPut(data: Record<string, any>) {
  return request.Put<Service.ResponseResult<any>>('/api/v1/demo/put', data)
}

/** DELETE 请求示例 */
export function fetchDelete() {
  return request.Delete<Service.ResponseResult<any>>('/api/v1/demo/delete')
}

/** Form POST 请求示例 */
export function fetchFormPost(data: Record<string, any>) {
  const methodInstance = request.Post<Service.ResponseResult<any>>('/api/v1/demo/form-post', data)
  methodInstance.meta = {
    isFormPost: true,
  }
  return methodInstance
}

/** 不携带 Token 的请求示例 */
export function withoutToken() {
  const methodInstance = request.Get<Service.ResponseResult<any>>('/api/v1/demo/no-token')
  methodInstance.meta = {
    authRole: null,
  }
  return methodInstance
}

/** Token 过期请求示例 */
export function expiredTokenRequest() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/demo/expired-token')
}

/** 获取 Blob 数据 */
export function getBlob(url: string) {
  const methodInstance = blankInstance.Get<Blob>(url)
  methodInstance.meta = {
    isBlob: true,
  }
  return methodInstance
}

/** 下载文件（带进度） */
export function downloadFile(url: string) {
  const methodInstance = blankInstance.Get<Blob>(url)
  methodInstance.meta = {
    isBlob: true,
  }
  return methodInstance
}

/** 失败请求 - 服务器错误 */
export function FailedRequest() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/demo/failed-request')
}

/** 失败请求 - 业务操作错误 */
export function FailedResponse() {
  return request.Get<Service.ResponseResult<any>>('/api/v1/demo/failed-response')
}

/** 失败请求 - 业务操作错误（无提示） */
export function FailedResponseWithoutTip() {
  const methodInstance = request.Get<Service.ResponseResult<any>>('/api/v1/demo/failed-response-no-tip')
  methodInstance.meta = {
    noErrorTip: true,
  }
  return methodInstance
}
