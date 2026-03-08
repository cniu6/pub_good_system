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
    const res = await request.Post<Service.ResponseResult<UserBatchSimpleInfoResponse>>(`${BASE_URL}/batch-simple`, { ids })
    return res.isSuccess ? (res.data?.users || {}) : {}
  },

  // 按标识查找用户（ID/用户名/邮箱）
  lookup(keyword: string) {
    return request.Get(`${BASE_URL}/lookup`, { params: { keyword } })
  },

  // 管理员登录指定用户（生成该用户的JWT token）
  loginAsUser(id: number) {
    return request.Post(`${BASE_URL}/${id}/login-as`)
  },

  // 重置指定用户的 API Key
  resetApiKey(id: number) {
    return request.Post(`${BASE_URL}/${id}/reset-apikey`)
  },

  // 变更用户余额（增减）
  changeMoney(id: number, data: { money: number, memo?: string }) {
    return request.Post(`${BASE_URL}/${id}/money/change`, data)
  },

  // 直接设置用户余额
  setMoney(id: number, data: { money: number, memo?: string }) {
    return request.Put(`${BASE_URL}/${id}/money`, data)
  },

  // 变更用户积分（增减）
  changeScore(id: number, data: { score: number, memo?: string }) {
    return request.Post(`${BASE_URL}/${id}/score/change`, data)
  },

  // 直接设置用户积分
  setScore(id: number, data: { score: number, memo?: string }) {
    return request.Put(`${BASE_URL}/${id}/score`, data)
  },
}

// 余额日志管理 API
const MONEY_LOGS_URL = `/api/v1${ADMIN_PATH}/money-logs`

export const adminMoneyLogApi = {
  list(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
    return request.Get<Service.ResponseResult<UserMoneyLogListResponse>>(MONEY_LOGS_URL, { params })
  },
  detail(id: number) {
    return request.Get<Service.ResponseResult<Entity.UserMoneyLog>>(`${MONEY_LOGS_URL}/${id}`)
  },
  delete(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${MONEY_LOGS_URL}/${id}`)
  },
}

// 积分日志管理 API
const SCORE_LOGS_URL = `/api/v1${ADMIN_PATH}/score-logs`

export const adminScoreLogApi = {
  list(params: { page?: number, page_size?: number, keyword?: string, user_id?: number }) {
    return request.Get<Service.ResponseResult<UserScoreLogListResponse>>(SCORE_LOGS_URL, { params })
  },
  detail(id: number) {
    return request.Get<Service.ResponseResult<Entity.UserScoreLog>>(`${SCORE_LOGS_URL}/${id}`)
  },
  delete(id: number) {
    return request.Delete<Service.ResponseResult<{ message: string }>>(`${SCORE_LOGS_URL}/${id}`)
  },
}
