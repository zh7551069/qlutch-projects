import { fileURLToPath } from 'node:url'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [react()],
  envDir: fileURLToPath(new URL('../..', import.meta.url)),
  server: {
    host: '127.0.0.1',
    port: 4712,
    strictPort: true,
  },
  preview: {
    host: '127.0.0.1',
    port: 4712,
    strictPort: true,
  },
  test: {
    environment: 'jsdom',
  },
})
