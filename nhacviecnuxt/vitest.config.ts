import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    include: ['server/**/*.test.ts', 'app/**/*.test.ts'],
    environment: 'node',
  },
  resolve: {
    alias: {
      '#shared/types': new URL('./shared/types/index.ts', import.meta.url).pathname,
    },
  },
})
