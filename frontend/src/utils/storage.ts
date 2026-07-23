const STORAGE_PREFIX = import.meta.env.VITE_STORAGE_PREFIX

type AuthStorageKey = 'userInfo' | 'accessToken' | 'refreshToken' | 'accessTokenExpiresAt' | 'role' | 'authGuard'
type AuthStorageScope = 'local' | 'session'
type AuthStorageValueMap = Pick<Storage.Local, AuthStorageKey>
type AuthStorageSnapshot = Partial<AuthStorageValueMap>

const AUTH_STORAGE_KEYS: AuthStorageKey[] = ['userInfo', 'accessToken', 'refreshToken', 'accessTokenExpiresAt', 'role', 'authGuard']

interface StorageData<T> {
  value: T
  expire: number | null
}
/**
 * LocalStorage部分操作
 */
function createLocalStorage<T extends Record<string, any>>() {
  // 默认缓存期限为7天

  function set<K extends keyof T>(key: K, value: T[K], expire: number = 60 * 60 * 24 * 7) {
    const storageData: StorageData<T[K]> = {
      value,
      expire: new Date().getTime() + expire * 1000,
    }
    const json = JSON.stringify(storageData)
    window.localStorage.setItem(`${STORAGE_PREFIX}${String(key)}`, json)
  }

  function get<K extends keyof T>(key: K) {
    const storageKey = `${STORAGE_PREFIX}${String(key)}`
    const json = window.localStorage.getItem(storageKey)
    if (!json)
      return null

    let storageData: StorageData<T[K]> | null = null
    try {
      storageData = JSON.parse(json) as StorageData<T[K]> | null
    }
    catch {
      // 损坏的 JSON 不应阻塞应用运行，直接丢弃并返回 null
      window.localStorage.removeItem(storageKey)
      return null
    }

    if (storageData) {
      const { value, expire } = storageData
      if (expire === null || expire >= Date.now())
        return value
    }
    remove(key)
    return null
  }

  function remove(key: keyof T) {
    window.localStorage.removeItem(`${STORAGE_PREFIX}${String(key)}`)
  }

  const clear = window.localStorage.clear

  return {
    set,
    get,
    remove,
    clear,
  }
}
/**
 * sessionStorage部分操作
 */

function createSessionStorage<T extends Record<string, any>>() {
  function set<K extends keyof T>(key: K, value: T[K]) {
    const json = JSON.stringify(value)
    window.sessionStorage.setItem(`${STORAGE_PREFIX}${String(key)}`, json)
  }
  function get<K extends keyof T>(key: K) {
    const storageKey = `${STORAGE_PREFIX}${String(key)}`
    const json = sessionStorage.getItem(storageKey)
    if (!json)
      return null

    try {
      const storageData: T[K] | null = JSON.parse(json)
      if (storageData)
        return storageData
    }
    catch {
      // 损坏的 sessionStorage 数据直接丢弃，避免阻塞应用
      window.sessionStorage.removeItem(storageKey)
    }

    return null
  }
  function remove(key: keyof T) {
    window.sessionStorage.removeItem(`${STORAGE_PREFIX}${String(key)}`)
  }
  const clear = window.sessionStorage.clear

  return {
    set,
    get,
    remove,
    clear,
  }
}

export const local = createLocalStorage<Storage.Local>()
export const session = createSessionStorage<Storage.Session>()
const authLocal = createLocalStorage<AuthStorageValueMap>()
const authSession = createSessionStorage<Pick<Storage.Session, AuthStorageKey>>()
const authIsolationSession = createSessionStorage<Pick<Storage.Session, 'authIsolation'>>()

function reportAuthStorageError(message: string, error: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

function getActiveAuthScope(): AuthStorageScope {
  return authIsolationSession.get('authIsolation') ? 'session' : 'local'
}

function setAuthKeyInScope<K extends AuthStorageKey>(key: K, value: AuthStorageValueMap[K], scope: AuthStorageScope) {
  if (scope === 'session') {
    authSession.set(key, value)
    return
  }

  authLocal.set(key, value)
}

function removeAuthKeyFromScope(key: AuthStorageKey, scope: AuthStorageScope) {
  if (scope === 'session') {
    authSession.remove(key)
    return
  }

  authLocal.remove(key)
}

function setAuthSnapshotInScope(snapshot: AuthStorageSnapshot, scope: AuthStorageScope) {
  AUTH_STORAGE_KEYS.forEach((key) => {
    const value = snapshot[key]
    if (value === undefined || value === null || value === '') {
      removeAuthKeyFromScope(key, scope)
      return
    }
    setAuthKeyInScope(key, value, scope)
  })
}

export const authStorage = {
  /**
   * admin 模式启动时调用，自动启用 sessionStorage 隔离，
   * 避免和普通用户 localStorage 里的 token 互相干扰。
   */
  enableSessionIsolation() {
    authIsolationSession.set('authIsolation', true)
  },
  get<K extends AuthStorageKey>(key: K) {
    if (getActiveAuthScope() === 'session') {
      return authSession.get(key) as AuthStorageValueMap[K] | null
    }

    return authLocal.get(key) as AuthStorageValueMap[K] | null
  },
  getActiveScope() {
    return getActiveAuthScope()
  },
  setLocal<K extends AuthStorageKey>(key: K, value: AuthStorageValueMap[K]) {
    setAuthKeyInScope(key, value, 'local')
  },
  setSession<K extends AuthStorageKey>(key: K, value: AuthStorageValueMap[K]) {
    authIsolationSession.set('authIsolation', true)
    setAuthKeyInScope(key, value, 'session')
  },
  setActive<K extends AuthStorageKey>(key: K, value: AuthStorageValueMap[K]) {
    setAuthKeyInScope(key, value, getActiveAuthScope())
  },
  setScope(snapshot: AuthStorageSnapshot, scope: AuthStorageScope) {
    if (scope === 'session') {
      authIsolationSession.set('authIsolation', true)
    }
    setAuthSnapshotInScope(snapshot, scope)
  },
  clearActive() {
    const scope = getActiveAuthScope()
    AUTH_STORAGE_KEYS.forEach(key => removeAuthKeyFromScope(key, scope))
  },
  openSessionWindow(snapshot: AuthStorageSnapshot, targetUrl = '/') {
    const targetLocation = new URL(targetUrl, window.location.origin)
    if (targetLocation.origin !== window.location.origin) {
      reportAuthStorageError('[AuthStorage] Rejected cross-origin session handoff target:', targetLocation.toString())
      return false
    }

    const targetWindow = window.open('about:blank', '_blank')
    if (!targetWindow) {
      return false
    }

    try {
      targetWindow.sessionStorage.setItem(`${STORAGE_PREFIX}authIsolation`, JSON.stringify(true))
      AUTH_STORAGE_KEYS.forEach((key) => {
        const value = snapshot[key]
        const storageKey = `${STORAGE_PREFIX}${String(key)}`
        if (value === undefined || value === null || value === '') {
          targetWindow.sessionStorage.removeItem(storageKey)
          return
        }
        targetWindow.sessionStorage.setItem(storageKey, JSON.stringify(value))
      })
      targetWindow.location.replace(targetLocation.toString())
      return true
    }
    catch (error) {
      reportAuthStorageError('[AuthStorage] Failed to open isolated session window:', error)
      targetWindow.close()
      return false
    }
  },
}
