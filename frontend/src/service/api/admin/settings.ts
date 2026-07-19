import { request } from '../../http'
import { getAdminApiBase } from './base'

/**
 * 系统配置管理 API
 * 用于管理端对系统配置进行 CRUD 操作
 */

// 配置项类型（password 语义等同 string，仅用于渲染带查看/隐藏眼睛的密码输入框）
export type SettingType = 'string' | 'number' | 'boolean' | 'json' | 'password'

// 配置分类
export type SettingCategory = 'basic' | 'security' | 'email' | 'sms' | 'payment' | 'custom'

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
  return request.Get<Service.ResponseResult<SettingsListResponse>>(`${getAdminApiBase()}/settings`)
}

/**
 * 获取指定分类的配置
 */
export function fetchSettingsByCategory(category: SettingCategory) {
  return request.Get<Service.ResponseResult<SettingDTO[]>>(`${getAdminApiBase()}/settings/category/${category}`)
}

/**
 * 获取单个配置
 */
export function fetchSetting(key: string) {
  return request.Get<Service.ResponseResult<SettingDTO>>(`${getAdminApiBase()}/settings/${key}`)
}

/**
 * 更新单个配置值
 * 这里故意保留通用返回类型：
 * 配置项是动态的，后端这组接口主要承载通用 message/data 包装，
 * 业务层应以 isSuccess/message 为准，不在 API 层对 data 做过度收窄。
 */
export function updateSetting(key: string, value: string) {
  return request.Put<Service.ResponseResult<any>>(`${getAdminApiBase()}/settings/${key}`, { value })
}

/**
 * 更新配置元数据
 * 这里故意保留通用返回类型：
 * 元数据编辑面向自定义配置 CRUD，后端返回体以通用成功消息为主，
 * 由调用页面按实际场景消费，避免把动态配置结构错误固化到类型层。
 */
export function updateSettingMeta(key: string, data: UpdateSettingMetaRequest) {
  return request.Put<Service.ResponseResult<any>>(`${getAdminApiBase()}/settings/${key}/meta`, data)
}

/**
 * 批量更新配置
 * 这里故意保留通用返回类型：
 * batch 接口服务多个配置分类，返回体没有稳定业务实体，
 * 当前只需要统一判断 isSuccess/message。
 */
export function batchUpdateSettings(settings: Record<string, string>) {
  return request.Put<Service.ResponseResult<any>>(`${getAdminApiBase()}/settings/batch`, { settings })
}

/**
 * 创建新配置
 */
export function createSetting(data: CreateSettingRequest) {
  return request.Post<Service.ResponseResult<{ message: string, key: string }>>(`${getAdminApiBase()}/settings`, data)
}

/**
 * 删除配置（仅限自定义配置）
 * 这里故意保留通用返回类型：
 * 删除后前端只依赖通用成功/失败语义，不依赖固定 data 结构。
 */
export function deleteSetting(key: string) {
  return request.Delete<Service.ResponseResult<any>>(`${getAdminApiBase()}/settings/${key}`)
}

/**
 * 重启后端服务
 * 这里故意保留通用返回类型：
 * 重启接口本质是一次命令触发，前端只消费通用 message，不绑定具体 data 结构。
 */
export function restartBackend() {
  return request.Post<Service.ResponseResult<any>>(`${getAdminApiBase()}/settings/restart-backend`)
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
}
