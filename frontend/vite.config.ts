import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    allowedHosts: [
      'iot.netr0.com',
      '.netr0.com', // 允许所有 netr0.com 的子域名
    ],
    // API代理配置：将/api请求转发到本地后端
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        // 可选：打印代理日志
        configure: (proxy, options) => {
          proxy.on('proxyReq', (_proxyReq, req) => {
            const target = options.target || 'http://localhost:8080';
            const url = req.url || '';
            console.log('[Proxy]', req.method, url, '→', target + url);
          });
        }
      }
    }
  },
  build: {
    // 代码分割优化
    rollupOptions: {
      output: {
        // 手动分割代码块
        manualChunks: {
          // React核心库单独打包
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          // 图表库单独打包（较大）
          'charts': ['recharts'],
          // 地图库单独打包（较大）
          'maps': ['leaflet', 'react-leaflet'],
          // 图标库单独打包
          'icons': ['lucide-react'],
          // 工具库
          'utils': ['axios', 'date-fns'],
        },
        // 优化chunk文件名
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
      }
    },
    // 提高chunk大小警告阈值
    chunkSizeWarningLimit: 600,
    // 启用CSS代码分割
    cssCodeSplit: true,
    // sourcemap在生产环境禁用以减小体积
    sourcemap: false,
    // 使用esbuild压缩（更快，效果也好）
    minify: 'esbuild',
    // 目标浏览器
    target: 'es2015',
  },
  // 性能优化
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      'recharts',
      'leaflet',
      'react-leaflet',
      'lucide-react',
      'axios',
      'date-fns',
    ],
  },
})
