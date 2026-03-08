/**
 * 管理端 API 服务 - 仪表盘
 */
import { request } from '@/service/http'

const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/dashboard`

export const adminDashboardApi = {
  // 获取仪表盘统计数据
  getStatistics() {
    return request.Get(BASE_URL)
  },
}
