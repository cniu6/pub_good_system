import { request } from '@/service/http'

/**
 * 协程统计响应
 */
export interface GoroutineStatsResponse {
  total_count: number
  mem_stats: {
    heap_alloc: number
    heap_sys: number
    heap_inuse: number
    heap_idle: number
    heap_released: number
    stack_inuse: number
    stack_sys: number
    sys: number
    num_gc: number
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
export function fetchGoroutineStats() {
  return request.Get<Service.ResponseResult<GoroutineStatsResponse>>('/api/v1/admin/debug/goroutines/stats')
}

/**
 * 强制执行垃圾回收
 */
export function forceGC() {
  return request.Post<Service.ResponseResult<GCResponse>>('/api/v1/admin/debug/gc')
}

/**
 * 获取 pprof CPU profile
 */
export function fetchCPUProfile(seconds: number = 30) {
  return `/api/v1/admin/debug/pprof/profile?seconds=${seconds}`
}

/**
 * 获取 pprof Heap profile
 */
export function fetchHeapProfile() {
  return `/api/v1/admin/debug/pprof/heap?debug=1`
}

/**
 * 获取 pprof Goroutine profile
 */
export function fetchGoroutineProfile(minWaitMinutes: number = 0) {
  return `/api/v1/admin/debug/pprof/goroutine?debug=2&min_wait_minutes=${minWaitMinutes}`
}

/**
 * 获取 pprof Allocs profile
 */
export function fetchAllocsProfile() {
  return `/api/v1/admin/debug/pprof/allocs?debug=1`
}

/**
 * 获取 pprof Block profile
 */
export function fetchBlockProfile() {
  return `/api/v1/admin/debug/pprof/block?debug=1`
}

/**
 * 获取 pprof Mutex profile
 */
export function fetchMutexProfile() {
  return `/api/v1/admin/debug/pprof/mutex?debug=1`
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
}
