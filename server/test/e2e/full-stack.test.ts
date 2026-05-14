import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { GenericContainer, Network, Wait, type StartedTestContainer, type StartedNetwork } from 'testcontainers';
import pg from 'pg';
import { drizzle } from 'drizzle-orm/node-postgres';
import { migrate } from 'drizzle-orm/node-postgres/migrator';
import * as schema from '../../src/db/schema/index.js';
import { workspaces, apiKeys } from '../../src/db/schema/index.js';
import { generateApiKey, hashApiKey, extractKeyPrefix } from '../../src/auth/api-key.js';
import { buildApp } from '../../src/app.js';
import type { FastifyInstance } from 'fastify';
import { loadConfig } from '../../src/config.js';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('full-stack e2e (slow)', () => {
  let network: StartedNetwork;
  let pgContainer: StartedTestContainer;
  let ollamaContainer: StartedTestContainer;
  let app: FastifyInstance;
  let key: string;

  beforeAll(async () => {
    network = await new Network().start();

    pgContainer = await new GenericContainer('pgvector/pgvector:pg16')
      .withNetwork(network)
      .withNetworkAliases('postgres')
      .withEnvironment({
        POSTGRES_USER: 'tolvi', POSTGRES_PASSWORD: 'tolvi', POSTGRES_DB: 'tolvi_e2e',
      })
      .withExposedPorts(5432)
      .withWaitStrategy(Wait.forLogMessage('database system is ready', 2))
      .start();

    ollamaContainer = await new GenericContainer('ollama/ollama:latest')
      .withNetwork(network)
      .withNetworkAliases('ollama')
      .withExposedPorts(11434)
      .withWaitStrategy(Wait.forHttp('/api/tags', 11434))
      .start();

    // Pull the model
    await ollamaContainer.exec(['ollama', 'pull', 'nomic-embed-text']);

    const databaseUrl = `postgresql://tolvi:tolvi@${pgContainer.getHost()}:${pgContainer.getMappedPort(5432)}/tolvi_e2e`;
    const ollamaUrl = `http://${ollamaContainer.getHost()}:${ollamaContainer.getMappedPort(11434)}`;

    // Run migrations
    const pool = new pg.Pool({ connectionString: databaseUrl });
    await migrate(drizzle(pool, { schema }), {
      migrationsFolder: path.resolve(__dirname, '../../src/db/migrations'),
    });
    await pool.end();

    // Build and start the app pointing at the real services
    process.env.DATABASE_URL = databaseUrl;
    process.env.EMBEDDING_PROVIDER = 'ollama';
    process.env.EMBEDDING_BASE_URL = ollamaUrl;
    process.env.LLM_PROVIDER = 'anthropic';
    process.env.ANTHROPIC_API_KEY = 'sk-ant-stub';   // /v1/ask not exercised in this test
    process.env.NODE_ENV = 'test';

    app = await buildApp(loadConfig());
    await app.ready();

    // Seed a workspace + key
    const [w] = await app.db.insert(workspaces).values({ slug: 'e2e', name: 'E2E' }).returning();
    key = generateApiKey();
    await app.db.insert(apiKeys).values({
      workspaceId: w!.id, keyHash: await hashApiKey(key), keyPrefix: extractKeyPrefix(key), name: 'e2e',
    });
  }, 300_000);

  afterAll(async () => {
    await app?.close();
    await ollamaContainer?.stop();
    await pgContainer?.stop();
    await network?.stop();
  });

  it('ingests a doc end-to-end and finds it via search', async () => {
    const sample = `---\ntags: [decision]\nstatus: active\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Postgres\n\nWe chose postgres because it has pgvector and json support.\n`;

    const ingest = await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-postgres.md', content: sample },
    });
    expect(ingest.statusCode).toBe(200);

    const search = await app.inject({
      method: 'POST', url: '/v1/search',
      headers: { authorization: `Bearer ${key}` },
      payload: { query: 'database choice', limit: 10 },
    });
    expect(search.statusCode).toBe(200);
    expect(search.json().results.length).toBeGreaterThan(0);
    expect(search.json().results[0].slug).toBe('postgres');
  }, 60_000);
});
