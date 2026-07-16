/**
 *后台服务的环境类型
 * - dev: 后台开发环境
 * - prod: 后台生产环境
 */
type ServiceEnvType = 'dev' | 'production'

interface ImportMetaEnv {
  /** 项目基本地址 */
  readonly VITE_BASE_URL: string
  /** 项目标题 */
  readonly VITE_APP_NAME: string
  /** 开启请求代理 */
  readonly VITE_HTTP_PROXY?: 'Y' | 'N'
  /** 是否开启打包压缩 */
  readonly VITE_BUILD_COMPRESS?: 'Y' | 'N'
  /** 压缩算法类型 */
  readonly VITE_COMPRESS_TYPE?:
    | 'gzip'
    | 'brotliCompress'
    | 'deflate'
    | 'deflateRaw'
  /** 路由模式 */
  readonly VITE_ROUTE_MODE?: 'hash' | 'web'
  /** 路由加载模式 */
  readonly VITE_ROUTE_LOAD_MODE: 'static' | 'dynamic'
  /** 首次加载页面 */
  readonly VITE_HOME_PATH: string
  /** 版权信息 */
  readonly VITE_COPYRIGHT_INFO: string
  /** 是否自动刷新token */
  readonly VITE_AUTO_REFRESH_TOKEN: 'Y' | 'N'
  /** 默认语言 */
  readonly VITE_DEFAULT_LANG: App.lang
  /** 用户协议链接 */
  readonly VITE_USER_AGREEMENT_URL: string
  /** 后端 API 地址（可在 .env.dev 和 .env.production 中分别覆盖） */
  readonly VITE_API_URL: string
  /** 管理后台前端页面入口路径（须与根目录 ADMIN_PATH 一致，如 /system-mgr） */
  readonly VITE_ADMIN_BASE_PATH?: string
  /** 管理端 REST API 在 /api/v1 下的前缀（须与根目录 ADMIN_API_PATH 一致，如 /admin） */
  readonly VITE_ADMIN_API_PATH?: string
  /**
   * 构建模式：embedded 表示嵌入单二进制，生产 API 走同源（VITE_API_URL 可留空）
   * 非 embedded 时生产必须配置可访问的 VITE_API_URL
   */
  readonly VITE_BUILD_MODE?: 'embedded' | 'external' | string
  /** 后端服务的环境类型 */
  readonly MODE: ServiceEnvType
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
