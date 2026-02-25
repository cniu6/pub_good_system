/**
 * 管理端 API 服务 - 用户管理
 * 此文件会被打包到 admin-api chunk
 */
import { request } from '@/service/http'

// 管理端 API 路径固定为 /admin（与后端保持一致）
const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/users`

// 用户简要信息类型
export interface UserSimpleInfo {
  id: number
  username: string
  nickname: string
  email: string
  role: string
  status: number
}

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

  // 批量获取用户简要信息
  // 返回 map[id]UserSimpleInfo，方便通过 ID 快速查找
  async batchSimpleInfo(ids: number[]): Promise<Record<number, UserSimpleInfo>> {
    if (!ids.length) return {}
    const res = await request.Post(`${BASE_URL}/batch-simple`, { ids })
    return res.data?.users || {}
  },
}
