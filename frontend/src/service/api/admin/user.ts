/**
 * 管理端 API 服务 - 用户管理
 * 此文件会被打包到 admin-api chunk
 */
import { request } from '@/service/http'

// 管理端 API 路径前缀（需与后端 ADMIN_PATH 保持一致）
// 说明：
// - 前端路由前缀使用 VITE_ADMIN_BASE_PATH
// - 后端管理端 API 也使用同一前缀：/api/v1{ADMIN_PATH}/...
// - 为了每次打包可动态修改路径，这里不允许硬编码 /admin
const ADMIN_PATH = import.meta.env.VITE_ADMIN_API_PATH || import.meta.env.VITE_ADMIN_BASE_PATH || '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/users`

export const adminUserApi = {
  // 用户列表
  list(params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: number | null
    role?: string
  }) {
    return request.Get(BASE_URL, { params })
  },

  // 用户详情
  detail(id: number) {
    return request.Get(`${BASE_URL}/${id}`)
  },

  // 创建用户
  create(data: {
    username: string
    password: string
    email: string
    nickname?: string
    mobile?: string
    role?: string
    status?: number
  }) {
    return request.Post(BASE_URL, data)
  },

  // 更新用户
  update(id: number, data: {
    nickname?: string
    email?: string
    mobile?: string
    role?: string
    status?: number
  }) {
    return request.Put(`${BASE_URL}/${id}`, data)
  },

  // 删除用户
  delete(id: number) {
    return request.Delete(`${BASE_URL}/${id}`)
  },

  // 更新用户状态
  updateStatus(id: number, status: number) {
    return request.Put(`${BASE_URL}/${id}/status`, { status })
  },

  // 重置用户密码
  resetPassword(id: number, password: string) {
    return request.Put(`${BASE_URL}/${id}/password`, { password })
  },
}
