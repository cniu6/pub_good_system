const STORAGE_PREFIX = import.meta.env.VITE_STORAGE_PREFIX
const AUTH_TRANSFER_WINDOW_NAME_PREFIX = '__FST_AUTH_TRANSFER__:'

type AuthStorageKey = 'userInfo' | 'accessToken' | 'refreshToken' | 'accessTokenExpiresAt' | 'role'
type AuthStorageScope = 'local' | 'session'
type AuthStorageValueMap = Pick<Storage.Local, AuthStorageKey>
type AuthStorageSnapshot = Partial<AuthStorageValueMap>

const AUTH_STORAGE_KEYS: AuthStorageKey[] = ['userInfo', 'accessToken', 'refreshToken', 'accessTokenExpiresAt', 'role']

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
    const json = window.localStorage.getItem(`${STORAGE_PREFIX}${String(key)}`)
    if (!json)
      return null

    const storageData: StorageData<T[K]> | null = JSON.parse(json)

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
    const json = sessionStorage.getItem(`${STORAGE_PREFIX}${String(key)}`)
    if (!json)
      return null

    const storageData: T[K] | null = JSON.parse(json)

    if (storageData)
      return storageData

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

function consumeTransferredAuthSession() {
  const rawWindowName = window.name
  if (!rawWindowName || !rawWindowName.startsWith(AUTH_TRANSFER_WINDOW_NAME_PREFIX)) {
    return
  }

  try {
    const snapshot = JSON.parse(rawWindowName.slice(AUTH_TRANSFER_WINDOW_NAME_PREFIX.length)) as AuthStorageSnapshot
    authIsolationSession.set('authIsolation', true)
    setAuthSnapshotInScope(snapshot, 'session')
  }
  catch (error) {
    console.error('[AuthStorage] Failed to consume transferred auth session:', error)
  }
  finally {
    window.name = ''
  }
}

consumeTransferredAuthSession()

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
    const targetWindow = window.open('about:blank', '_blank')
    if (!targetWindow) {
      return false
    }

    targetWindow.name = `${AUTH_TRANSFER_WINDOW_NAME_PREFIX}${JSON.stringify(snapshot)}`
    targetWindow.location.replace(targetUrl)
    return true
  },
}
