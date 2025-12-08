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
})
