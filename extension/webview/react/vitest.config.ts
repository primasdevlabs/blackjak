import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.smoke.test.ts', 'src/**/*.test.ts'],
  },
});
