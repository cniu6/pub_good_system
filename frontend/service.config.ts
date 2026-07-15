/**
 * 不同请求服务的环境配置
 * 地址从 .env 文件读取：VITE_API_URL
 * - 开发环境：.env.dev 中的 VITE_API_URL
 * - 生产环境：.env.production 中的 VITE_API_URL
 * Vite 会根据运行模式自动加载对应的 .env 文件
 */
export function getServiceConfig(env: Record<string, string>): Record<ServiceEnvType, Record<string, string>> {
  // 与根目录 .env PORT 默认 8080、frontend/.env.dev 保持一致
  const apiUrl = env.VITE_API_URL || 'http://localhost:8080'
  const buildMode = env.VITE_BUILD_MODE || ''
  // embedded 单二进制：生产走同源，API 基址留空
  const productionApiUrl = buildMode === 'embedded' ? '' : apiUrl

  return {
    dev: {
      url: apiUrl,
    },
    production: {
      url: productionApiUrl,
    },
  }
}
