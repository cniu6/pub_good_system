/**
 * 管理端 API 统一导出（异步按需加载）
 *
 * 每个子模块只在首次调用时才会动态 import()，
 * 避免进入管理���时一次性同步加载全部 API 模块造成卡顿。
 *
 * 使用方式（与同步版本完全一致）：
 *   import { adminApi } from '@/service/api/admin'
 *   const res = await adminApi.user.list({ page: 1 })
 */

/**
 * 创建懒加载 API 代理
 * 首次调用任意方法时才 import() 对应模块，之后从缓存读取
 */
function createLazyModule<T extends Record<string, (...args: any[]) => any>>(
  loader: () => Promise<T>,
): T {
  let cached: T | null = null
  let loading: Promise<T> | null = null

  return new Proxy({} as T, {
    get(_, method: string) {
      return async (...args: any[]) => {
        if (!cached) {
          if (!loading) loading = loader()
          cached = await loading
        }
        return (cached as any)[method](...args)
      }
    },
  })
}

export const adminApi = {
  user: createLazyModule(() => import('./user').then(m => m.adminUserApi)),
  log: createLazyModule(() => import('./log').then(m => m.adminLogApi)),
}

