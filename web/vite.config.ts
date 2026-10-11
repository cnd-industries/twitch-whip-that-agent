import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// In dev, the Go backend runs separately and serves the event stream.
const backend = 'http://127.0.0.1:8080'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/events': backend,
      '/api': backend,
    },
  },
})
