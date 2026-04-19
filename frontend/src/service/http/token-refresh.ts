import { fetchUpdateToken } from '../api/user/login'
import { getRuntimeRouteMode } from '@/router/runtime-mode'

export type LoginTokenPayload = Api.Login.Info & { expiresAt?: number }

let refreshPromise: Promise<LoginTokenPayload | null> | null = null

function getCurrentAuthGuard(): 'user' | 'admin' {
  return getRuntimeRouteMode() === 'admin' ? 'admin' : 'user'
}

// 统一刷新请求，避免多个 401/定时器并发触发时重复轮换 refresh token。
export async function refreshAuthToken(refreshToken: string | null): Promise<LoginTokenPayload | null> {
  if (!refreshToken)
    return null

  if (refreshPromise)
    return refreshPromise

  refreshPromise = (async () => {
    try {
      const result = await fetchUpdateToken({ refreshToken, authGuard: getCurrentAuthGuard() })
      const data = result.data as LoginTokenPayload | null
      if (result.isSuccess && data)
        return data
      return null
    }
    catch {
      return null
    }
    finally {
      refreshPromise = null
    }
  })()

  return refreshPromise
}
