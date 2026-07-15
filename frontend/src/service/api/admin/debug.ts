import { request } from '@/service/http'
import { getAdminApiBase } from './base'

/**
 * 协程统计响应
 */
export interface RuntimeGoroutineInfo {
  id: number
  state: string
  wait_time?: string
  stack: string
  function: string
  locked_to_thread?: boolean
  created_by?: string
  stack_lines: number
}

export interface GoroutineStatsResponse {
  total_count: number
  tracked_count: number
  potential_leaks: number
  potential_leak_stacks?: RuntimeGoroutineInfo[]
  long_running?: RuntimeGoroutineInfo[]
  runtime_stacks?: RuntimeGoroutineInfo[]
  runtime_state_summary?: Record<string, number>
  by_state?: Record<string, number>
  num_cpu: number
  gomaxprocs: number
  num_cgo_call: number
  mem_stats: {
    heap_alloc: number
    total_alloc: number
    heap_sys: number
    heap_inuse: number
    heap_idle: number
    heap_released: number
    heap_objects: number
    stack_inuse: number
    stack_sys: number
    sys: number
    mallocs: number
    frees: number
    next_gc: number
    last_gc: number
    pause_total_ns: number
    num_gc: number
    num_forced_gc: number
    gc_cpu_fraction: number
  }
}

/**
 * GC 响应
 */
export interface GCResponse {
  goroutines_before: number
  goroutines_after: number
  message: string
}

/**
 * 获取协程统计信息
 */
export function fetchGoroutineStats(params?: { stacks?: boolean, min_wait_minutes?: number }) {
  return request.Get<Service.ResponseResult<GoroutineStatsResponse>>(`${getAdminApiBase()}/debug/goroutines/stats`, { params })
}

/**
 * 强制执行垃圾回收
 */
export function forceGC() {
  return request.Post<Service.ResponseResult<GCResponse>>(`${getAdminApiBase()}/debug/gc`)
}

/**
 * 获取 pprof CPU profile
 */
export function fetchCPUProfile(seconds: number = 30) {
  return `${getAdminApiBase()}/debug/pprof/profile?seconds=${seconds}`
}

/**
 * 获取 pprof Heap profile
 */
export function fetchHeapProfile() {
  return `${getAdminApiBase()}/debug/pprof/heap?debug=1`
}

/**
 * 获取 pprof Goroutine profile
 */
export function fetchGoroutineProfile(minWaitMinutes: number = 0) {
  return `${getAdminApiBase()}/debug/pprof/goroutine?debug=2&min_wait_minutes=${minWaitMinutes}`
}

/**
 * 获取 pprof Allocs profile
 */
export function fetchAllocsProfile() {
  return `${getAdminApiBase()}/debug/pprof/allocs?debug=1`
}

/**
 * 获取 pprof Block profile
 */
export function fetchBlockProfile() {
  return `${getAdminApiBase()}/debug/pprof/block?debug=1`
}

/**
 * 获取 pprof Mutex profile
 */
export function fetchMutexProfile() {
  return `${getAdminApiBase()}/debug/pprof/mutex?debug=1`
}

/**
 * 获取 pprof ThreadCreate profile
 */
export function fetchThreadCreateProfile() {
  return `${getAdminApiBase()}/debug/pprof/threadcreate?debug=1`
}

/**
 * 获取 pprof Trace
 */
export function fetchTraceProfile(seconds: number = 5, binary: boolean = false) {
  const query = new URLSearchParams({ seconds: String(seconds) })
  if (binary)
    query.set('binary', '1')
  return `${getAdminApiBase()}/debug/pprof/trace?${query.toString()}`
}

// 导出调试 API 对象
export const adminDebugApi = {
  goroutineStats: fetchGoroutineStats,
  forceGC,
  cpuProfile: fetchCPUProfile,
  heapProfile: fetchHeapProfile,
  goroutineProfile: fetchGoroutineProfile,
  allocsProfile: fetchAllocsProfile,
  blockProfile: fetchBlockProfile,
  mutexProfile: fetchMutexProfile,
  threadcreateProfile: fetchThreadCreateProfile,
  traceProfile: fetchTraceProfile,
}
