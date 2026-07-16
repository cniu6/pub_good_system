import { bootstrap } from './bootstrap'

// 根据当前 URL pathname 自动判断运行模式
// 管理端使用 hash 路由，访问路径为 VITE_ADMIN_BASE_PATH（如 /system-mgr，与 ADMIN_PATH 一致）
const adminBase = (import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr').replace(/\/+$/, '') || '/system-mgr'
const pathname = window.location.pathname
const isAdmin = pathname === adminBase || pathname.startsWith(`${adminBase}/`)

bootstrap(isAdmin ? 'admin' : 'user')
