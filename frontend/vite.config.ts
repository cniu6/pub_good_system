import { copyFileSync, existsSync, mkdirSync } from 'node:fs'
import { resolve } from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import { createVitePlugins } from './build/plugins'

function normalizeAdminEntryDir(path: string | undefined): string {
  const raw = (path || '/system-mgr').trim()
  const normalized = raw.replace(/^\/+|\/+$/g, '')
  return normalized || 'system-mgr'
}

function ensureAdminEntryHtml(rootDir: string, adminEntryDir: string): string {
  const templatePath = resolve(rootDir, '_admin-entry/index.html')
  const targetPath = resolve(rootDir, adminEntryDir, 'index.html')

  if (!existsSync(targetPath)) {
    if (!existsSync(templatePath))
      throw new Error(`[vite] Missing admin entry template: ${templatePath}`)
    mkdirSync(resolve(rootDir, adminEntryDir), { recursive: true })
    copyFileSync(templatePath, targetPath)
  }

  return targetPath
}

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // 根据当前工作目录中的 `mode` 加载 .env 文件
  const env = loadEnv(mode, __dirname, '') as ImportMetaEnv
  const adminEntryDir = normalizeAdminEntryDir(env.VITE_ADMIN_BASE_PATH)
  const adminEntryHtml = ensureAdminEntryHtml(__dirname, adminEntryDir)

  return {
    base: env.VITE_BASE_URL,
    define: {
      __BUILD_TIMESTAMP__: JSON.stringify(new Date().toLocaleString()),
    },
    plugins: createVitePlugins(env),
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      host: '0.0.0.0',
      proxy: {
        '/api': {
          target: env.VITE_API_URL || 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
    build: {
      target: 'esnext',
      // 输出到项目根目录 dist/，供 go:embed 嵌入
      outDir: resolve(__dirname, '../dist'),
      // 构建前自动清空根目录 dist/ 旧产物
      emptyOutDir: true,
      reportCompressedSize: false,
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
        input: {
          user: resolve(__dirname, 'index.html'),
          admin: adminEntryHtml,
        },
        output: {
          /**
           * 管理端 JS 隔离打包策略
           *
           * 安全目的：
           * 1. 将管理端代码打包到独立的 chunk
           * 2. 普通用户无法通过查看 JS 获取管理端代码结构
           * 3. 只有管理员登录后才会动态加载这些 chunk
           *
           * 打包结果：
           * - assets/m/admin-views-[hash].js  管理端视图组件
           * - assets/m/admin-core-[hash].js   管理端路由配置
           * - assets/m/admin-api-[hash].js    管理端 API 调用
           */
          manualChunks(id) {
            // 管理端路由配置 -> admin-core chunk
            if (id.includes('src/router/admin.routes')) {
              return 'admin-core'
            }
            // 第三方库分包优化
            if (id.includes('node_modules')) {
              if (id.includes('naive-ui')) {
                return 'vendor-naive'
              }
              if (id.includes('echarts')) {
                return 'vendor-echarts'
              }
            }
          },
          // 自定义 chunk 文件名
          chunkFileNames: (chunkInfo) => {
            const name = chunkInfo.name || 'chunk'
            // 管理端 chunk 使用特殊目录，增加混淆
            if (name.startsWith('admin-')) {
              return `assets/m/${name}-[hash].js`
            }
            return `assets/js/${name}-[hash].js`
          },
          entryFileNames: 'assets/js/[name]-[hash].js',
          assetFileNames: (assetInfo) => {
            const name = assetInfo.name || ''
            if (/\.(?:png|jpe?g|gif|svg|webp|ico)$/.test(name)) {
              return 'assets/img/[name]-[hash][extname]'
            }
            if (/\.(?:woff2?|eot|ttf|otf)$/.test(name)) {
              return 'assets/fonts/[name]-[hash][extname]'
            }
            if (/\.css$/.test(name)) {
              return 'assets/css/[name]-[hash][extname]'
            }
            return 'assets/[name]-[hash][extname]'
          },
        },
      },
    },
    optimizeDeps: {
      include: ['echarts', 'md-editor-v3', 'quill'],
    },
  }
})
