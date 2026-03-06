import { bootstrap } from './bootstrap'

// 根据当前 URL pathname 自动判断运行模式
// 管理端使用 hash 路由，访问路径为 VITE_ADMIN_BASE_PATH（如 /system-mgr）
const adminBase = (import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr').replace(/\/+$/, '')
const pathname = window.location.pathname
const isAdmin = pathname === adminBase || pathname.startsWith(`${adminBase}/`)

bootstrap(isAdmin ? 'admin' : 'user')
