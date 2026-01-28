import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    // 监听所有 IP，方便局域网手机测试 (可选)
    host: '0.0.0.0', 
    port: 5173,
    // 👇 核心配置：跨域代理
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // 你的 Go 后端地址
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, '') // 如果你后端路由没带 /api 前缀才需要这行，但你有 /api/v1，所以不需要 rewrite
      }
    }
  }
})