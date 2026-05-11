import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    projects: [
      {
        test: {
          name: 'unit',
          include: ['test/unit/**/*.test.ts'],
          environment: 'node',
        },
      },
      {
        test: {
          name: 'integration',
          include: ['test/integration/**/*.test.ts'],
          environment: 'node',
          testTimeout: 30000,
          hookTimeout: 60000,
          // Integration tests share a testcontainers Postgres+pgvector instance
          // declared at module scope in test/helpers/postgres.ts. Running test
          // files in parallel would spawn N containers concurrently. singleFork
          // serializes integration files within one process so the singleton
          // actually behaves as one. Unit + e2e remain unconstrained.
          pool: 'forks',
          poolOptions: { forks: { singleFork: true } },
        },
      },
      {
        test: {
          name: 'e2e',
          include: ['test/e2e/**/*.test.ts'],
          environment: 'node',
          testTimeout: 120000,
          hookTimeout: 180000,
        },
      },
    ],
  },
});
