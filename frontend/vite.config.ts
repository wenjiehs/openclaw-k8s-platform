import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// Vite 配置文件
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(), // React JSX 支持
  ],

  // 路径别名：@/ 指向 src/ 目录
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },

  // 开发服务器配置
  server: {
    port: 3000,
    // 代理 API 请求到后端服务，避免跨域问题
    proxy: {
      '/api': {
        target: process.env.VITE_API_BASE_URL || 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path, // 保留 /api 前缀
      },
    },
  },

  // 构建配置
  build: {
    outDir: 'dist',
    // 代码分割优化：vendor 库单独打包
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          antd: ['antd', '@ant-design/icons'],
          axios: ['axios'],
        },
      },
    },
  },
})
