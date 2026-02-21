import { resolve } from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import { createVitePlugins } from './build/plugins'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // 根据当前工作目录中的 `mode` 加载 .env 文件
  const env = loadEnv(mode, __dirname, '') as ImportMetaEnv

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
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
    build: {
      target: 'esnext',
      reportCompressedSize: false,
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
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
            // 管理端视图组件 -> admin-views chunk
            if (id.includes('src/views/admin')) {
              return 'admin-views'
            }
            // 管理端路由配置 -> admin-core chunk
            if (id.includes('src/router/admin.routes')) {
              return 'admin-core'
            }
            // 管理端 API 服务 -> admin-api chunk
            if (id.includes('src/service/api/admin')) {
              return 'admin-api'
            }
            // 第三方库分包优化
            if (id.includes('node_modules')) {
              if (id.includes('naive-ui')) {
                return 'vendor-naive'
              }
              if (id.includes('echarts')) {
                return 'vendor-echarts'
              }
              if (id.includes('md-editor') || id.includes('quill')) {
                return 'vendor-editor'
              }
              if (id.includes('vue') || id.includes('@vue')) {
                return 'vendor-vue'
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
            if (/\.(png|jpe?g|gif|svg|webp|ico)$/.test(name)) {
              return 'assets/img/[name]-[hash][extname]'
            }
            if (/\.(woff2?|eot|ttf|otf)$/.test(name)) {
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
