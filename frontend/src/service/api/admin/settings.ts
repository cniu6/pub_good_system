import { request } from '../../http'

/**
 * 系统配置管理 API
 * 用于管理端对系统配置进行 CRUD 操作
 */

// 配置项类型
export type SettingType = 'string' | 'number' | 'boolean' | 'json'

// 配置分类
export type SettingCategory = 'basic' | 'security' | 'email' | 'custom'

// 配置项 DTO
export interface SettingDTO {
  key: string
  value: any // 根据类型不同，可能是 string | number | boolean | object
  type: SettingType
  category: SettingCategory
  label: string
  description: string
  is_public: boolean
  is_editable: boolean
}

// 配置分组
export interface SettingsGroup {
  category: SettingCategory
  label: string
  items: SettingDTO[]
}

// 配置列表响应
export interface SettingsListResponse {
  categories: SettingsGroup[]
}

export interface ServerMonitoringAppInfo {
  name: string
  mode: string
  port: string
  go_version: string
}

export interface ServerMonitoringProcessInfo {
  pid: number
  goroutines: number
  process_cpu: number
  process_rss_mb: number
  memory_alloc_mb: number
  memory_sys_mb: number
  heap_alloc_mb: number
  heap_inuse_mb: number
  heap_idle_mb: number
  stack_inuse_mb: number
  gc_count: number
  gc_cpu_fraction: number
}

export interface ServerMonitoringCpuMetrics {
  usage_percent: number
  core_count: number
}

export interface ServerMonitoringMemoryMetrics {
  total_mb: number
  used_mb: number
  used_percent: number
}

export interface ServerMonitoringDiskMetrics {
  path: string
  total_gb: number
  used_gb: number
  used_percent: number
}

export interface ServerMonitoringNetworkMetrics {
  bytes_sent: number
  bytes_recv: number
  packets_sent: number
  packets_recv: number
}

export interface ServerMonitoringSwapMetrics {
  total_mb: number
  used_mb: number
  used_percent: number
}

export interface ServerMonitoringMetrics {
  cpu: ServerMonitoringCpuMetrics
  memory: ServerMonitoringMemoryMetrics
  swap: ServerMonitoringSwapMetrics
  disk: ServerMonitoringDiskMetrics
  network: ServerMonitoringNetworkMetrics
}

export interface ServerMonitoringService {
  name: string
  status: 'up' | 'down' | 'warning'
  message: string
  configured?: boolean
  host?: string
  port?: string
  open_connections?: number
  in_use?: number
  idle?: number
}

export interface ServerMonitoringStatusResponse {
  generated_at: string
  uptime_seconds: number
  app: ServerMonitoringAppInfo
  metrics: ServerMonitoringMetrics
  process: ServerMonitoringProcessInfo
  services: ServerMonitoringService[]
}

// 创建配置请求
export interface CreateSettingRequest {
  key: string
  value: string
  type?: SettingType
  category?: SettingCategory
  label: string
  description?: string
  is_public?: boolean
  is_editable?: boolean
  sort_order?: number
}

// 更新配置元数据请求
export interface UpdateSettingMetaRequest {
  value: string
  type?: SettingType
  category?: SettingCategory
  label?: string
  description?: string
  is_public?: boolean
  is_editable?: boolean
  sort_order?: number
}

// 批量更新配置请求
export interface BatchUpdateSettingsRequest {
  settings: Record<string, string>
}

/**
 * 获取所有配置（按分类分组）
 */
export function fetchSettings() {
  return request.Get<Service.ResponseResult<SettingsListResponse>>('/api/v1/admin/settings')
}

/**
 * 获取指定分类的配置
 */
export function fetchSettingsByCategory(category: SettingCategory) {
  return request.Get<Service.ResponseResult<SettingDTO[]>>(`/api/v1/admin/settings/category/${category}`)
}

/**
 * 获取单个配置
 */
export function fetchSetting(key: string) {
  return request.Get<Service.ResponseResult<SettingDTO>>(`/api/v1/admin/settings/${key}`)
}

/**
 * 更新单个配置值
 */
export function updateSetting(key: string, value: string) {
  return request.Put<Service.ResponseResult<any>>(`/api/v1/admin/settings/${key}`, { value })
}

/**
 * 更新配置元数据
 */
export function updateSettingMeta(key: string, data: UpdateSettingMetaRequest) {
  return request.Put<Service.ResponseResult<any>>(`/api/v1/admin/settings/${key}/meta`, data)
}

/**
 * 批量更新配置
 */
export function batchUpdateSettings(settings: Record<string, string>) {
  return request.Put<Service.ResponseResult<any>>('/api/v1/admin/settings/batch', { settings })
}

/**
 * 创建新配置
 */
export function createSetting(data: CreateSettingRequest) {
  return request.Post<Service.ResponseResult<{ message: string, key: string }>>('/api/v1/admin/settings', data)
}

/**
 * 删除配置（仅限自定义配置）
 */
export function deleteSetting(key: string) {
  return request.Delete<Service.ResponseResult<any>>(`/api/v1/admin/settings/${key}`)
}

/**
 * 重启后端服务
 */
export function restartBackend() {
  return request.Post<Service.ResponseResult<any>>('/api/v1/admin/settings/restart-backend')
}

/**
 * 获取本项目服务器监控状态
 */
export function fetchServerMonitoringStatus() {
  return request.Get<Service.ResponseResult<ServerMonitoringStatusResponse>>('/api/v1/admin/settings/server-monitoring')
}

// 导出管理端配置 API 对象（用于懒加载模式）
export const adminSettingsApi = {
  list: fetchSettings,
  listByCategory: fetchSettingsByCategory,
  get: fetchSetting,
  update: updateSetting,
  updateMeta: updateSettingMeta,
  batchUpdate: batchUpdateSettings,
  create: createSetting,
  delete: deleteSetting,
  restartBackend,
  serverMonitoring: fetchServerMonitoringStatus,
}
