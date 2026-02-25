import type { AppRouteMode } from './index'

declare global {
  interface Window {
    __APP_ROUTE_MODE__?: AppRouteMode
  }
}

export function setRuntimeRouteMode(mode: AppRouteMode) {
  window.__APP_ROUTE_MODE__ = mode
}

export function getRuntimeRouteMode(): AppRouteMode {
  return window.__APP_ROUTE_MODE__ || 'user'
}
