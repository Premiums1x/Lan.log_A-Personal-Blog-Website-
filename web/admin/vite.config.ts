import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: '/admin/',
  resolve: { alias: { '@': '/src' } },
  server: {
    port: 5174,
    proxy: {
      '/api': 'http://localhost:8080',
      '/static': 'http://localhost:8080',
    },
  },
  build: {
    outDir: '../admin-dist',
    emptyOutDir: true,
  },
  publicDir: false,
})