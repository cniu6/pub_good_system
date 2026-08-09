/**
 * 自动任务管理器 API
 * 对应后端 /api/v1/{ADMIN}/auto-jobs/*
 */
import { request } from '@/service/http'
import { getAdminApiBase } from './base'

function baseUrl() {
  return `${getAdminApiBase()}/auto-jobs`
}

export interface AutoJobOverview {
  enabled_jobs: number
  total_jobs: number
  running_count: number
  today_success: number
  today_failed: number
  lifetime_run_total: string
  run_row_count: number
  run_max_count: number
  scheduler_running: boolean
  scheduler_uptime_sec: number
  last_tick_at: number
  global_enabled: boolean
}

export interface AutoJobGlobalConfig {
  auto_job_enabled: boolean
  auto_job_run_max_count: number
  auto_job_retain_errors: boolean
  auto_job_auto_prune: boolean
  auto_job_stuck_after_sec: number
  auto_job_auto_keep_job_codes: string[]
  auto_job_auto_keep_categories: string[]
}

export interface AutoJobDefinition {
  job_code: string
  name: string
  description: string
  category: string
  handler_key: string
  cron_expr: string
  interval_seconds: number
  timezone: string
  enabled: number
  timeout_sec: number
  max_concurrency: number
  params_json: string
  last_status: string
  last_started_at: number
  last_finished_at: number
  last_error: string
  lifetime_run_count: string
  lifetime_success_count: string
  lifetime_fail_count: string
  create_time: number
  update_time: number
}

export interface AutoJobRun {
  id: number
  run_uid: string
  job_code: string
  category: string
  trigger: string
  status: string
  started_at: number
  finished_at: number
  duration_ms: number
  message: string
  detail_json: string
  error_text: string
  keep_forever: number
  operator: string
}

export interface AutoJobRunKeep {
  id: number
  run_uid: string
  job_code: string
  category: string
  trigger: string
  status: string
  started_at: number
  finished_at: number
  duration_ms: number
  message: string
  detail_json: string
  error_text: string
  operator: string
  source_run_id: number
  kept_at: number
  run_timestamp: number
}

export interface AutoJobUpdateRequest {
  name?: string
  description?: string
  cron_expr?: string
  interval_seconds?: number
  timezone?: string
  enabled?: boolean
  timeout_sec?: number
  params_json?: string
}

export const adminAutoJobApi = {
  overview() {
    return request.Get<Service.ResponseResult<AutoJobOverview>>(`${baseUrl()}/overview`)
  },
  getConfig() {
    return request.Get<Service.ResponseResult<AutoJobGlobalConfig>>(`${baseUrl()}/config`)
  },
  saveConfig(data: AutoJobGlobalConfig) {
    return request.Put<Service.ResponseResult<AutoJobGlobalConfig>>(`${baseUrl()}/config`, data)
  },
  importPresets(mode: 'skip' | 'update' = 'skip') {
    return request.Post<Service.ResponseResult<Record<string, number>>>(`${baseUrl()}/presets/import`, { mode })
  },
  listHandlers() {
    return request.Get<Service.ResponseResult<{ handlers: string[] }>>(`${baseUrl()}/handlers`)
  },
  listJobs(params?: { keyword?: string, category?: string, enabled?: boolean | string }) {
    return request.Get<Service.ResponseResult<{ list: AutoJobDefinition[], total: number }>>(`${baseUrl()}`, { params })
  },
  getJob(jobCode: string) {
    return request.Get<Service.ResponseResult<AutoJobDefinition>>(`${baseUrl()}/${encodeURIComponent(jobCode)}`)
  },
  updateJob(jobCode: string, data: AutoJobUpdateRequest) {
    return request.Put<Service.ResponseResult<AutoJobDefinition>>(`${baseUrl()}/${encodeURIComponent(jobCode)}`, data)
  },
  runJob(jobCode: string) {
    return request.Post<Service.ResponseResult<AutoJobRun>>(`${baseUrl()}/${encodeURIComponent(jobCode)}/run`)
  },
  enableJob(jobCode: string) {
    return request.Post<Service.ResponseResult<{ job_code: string, enabled: boolean }>>(`${baseUrl()}/${encodeURIComponent(jobCode)}/enable`)
  },
  disableJob(jobCode: string) {
    return request.Post<Service.ResponseResult<{ job_code: string, enabled: boolean }>>(`${baseUrl()}/${encodeURIComponent(jobCode)}/disable`)
  },
  listRuns(params?: Record<string, string | number | undefined>) {
    return request.Get<Service.ResponseResult<{ list: AutoJobRun[], total: number, page: number, page_size: number }>>(`${baseUrl()}/runs`, { params })
  },
  listKeptRuns(params?: Record<string, string | number | undefined>) {
    return request.Get<Service.ResponseResult<{ list: AutoJobRunKeep[], total: number, page: number, page_size: number }>>(`${baseUrl()}/runs/kept`, { params })
  },
  listRunning() {
    return request.Get<Service.ResponseResult<{ list: AutoJobDefinition[], total: number }>>(`${baseUrl()}/running`)
  },
  getRun(id: number) {
    return request.Get<Service.ResponseResult<AutoJobRun>>(`${baseUrl()}/runs/${id}`)
  },
  cleanRuns(data: Record<string, unknown>) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/runs/clean`, data)
  },
  markKeep(ids: number[], keepForever: boolean) {
    return request.Post<Service.ResponseResult<{ affected: number }>>(`${baseUrl()}/runs/mark-keep`, {
      ids,
      keep_forever: keepForever,
    })
  },
}
