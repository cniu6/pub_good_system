import { request } from '@/service/http'
import type { ServerMonitoringStatusResponse } from './settings'

export interface BackgroundTaskInfo {
  key: string
  label: string
  running: boolean
  interval_secs: number
  last_run_time: string
  next_run_time: string
  last_status: string
  last_message: string
  last_duration_ms: number
}

export interface DynamicRateLimitSnapshot {
  name: string
  enabled: boolean
  rate: number
  burst: number
  allowed_count: number
  blocked_count: number
  total_count: number
  active_visitors: number
  last_config_reload: string
  cleanup_interval_ms: number
}

export interface ServerOperationsStatusResponse {
  tasks: BackgroundTaskInfo[]
  rate_limits: DynamicRateLimitSnapshot[]
  api_log: {
    enabled: boolean
    query_days: number
    max_count: number
  }
}

export const adminServerApi = {
  monitoring() {
    return request.Get<Service.ResponseResult<ServerMonitoringStatusResponse>>('/api/v1/admin/settings/server-monitoring')
  },

  operations() {
    return request.Get<Service.ResponseResult<ServerOperationsStatusResponse>>('/api/v1/admin/settings/server-ops')
  },

  runTask(key: string) {
    return request.Post<Service.ResponseResult<{ message: string }>>('/api/v1/admin/settings/background-tasks/run', { key })
  },
}
