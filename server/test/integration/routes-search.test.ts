import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { buildTestApp } from '../helpers/app.js';
import { workspaces, apiKeys } from '../../src/db/schema/index.js';
import { generateApiKey, hashApiKey, extractKeyPrefix } from '../../src/auth/api-key.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import type { FastifyInstance } from 'fastify';
import type pg from 'pg';

class DetEmbedding implements EmbeddingProvider {
  readonly dimension = 768;
  async embed(texts: string[]): Promise<number[][]> {
    return texts.map((t) => {
      const v = new Array(768).fill(0);
      for (let i = 0; i < t.length; i++) v[i % 768] += t.charCodeAt(i) / 1000;
      const n = Math.sqrt(v.reduce((s, x) => s + x * x, 0));
      return n > 0 ? v.map((x) => x / n) : v;
    });
  }
  async ping(): Promise<boolean> { return true; }
}

const sample = `---\ntags: [decision]\nstatus: active\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Postgres\n\nWe chose postgres for its pgvector and json support.\n`;

describe('routes/search (integration)', () => {
  let db: TestDb; let pool: pg.Pool; let app: FastifyInstance; let key: string;
  beforeAll(async () => { const r = await startTestDb(); db = r.db; pool = r.pool; });
  afterAll(async () => { await stopTestDb(); });

  beforeEach(async () => {
    await resetTestDb(pool);
    const [w] = await db.insert(workspaces).values({ slug: 'w', name: 'W' }).returning();
    key = generateApiKey();
    await db.insert(apiKeys).values({
      workspaceId: w!.id, keyHash: await hashApiKey(key), keyPrefix: extractKeyPrefix(key), name: 'k',
    });
    app = await buildTestApp({ db, pool, embedding: new DetEmbedding() });
    await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-postgres.md', content: sample },
    });
  });

  it('returns results for a relevant query', async () => {
    const res = await app.inject({
      method: 'POST', url: '/v1/search',
      headers: { authorization: `Bearer ${key}` },
      payload: { query: 'postgres', limit: 10 },
    });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.results.length).toBeGreaterThan(0);
    expect(body.results[0].slug).toBe('postgres');
    expect(body.results[0].matched_chunk).toBeDefined();
  });

  it('respects limit', async () => {
    const res = await app.inject({
      method: 'POST', url: '/v1/search',
      headers: { authorization: `Bearer ${key}` },
      payload: { query: 'postgres', limit: 1 },
    });
    expect(res.json().results).toHaveLength(1);
  });

  it('returns 401 without auth', async () => {
    const res = await app.inject({ method: 'POST', url: '/v1/search', payload: { query: 'x' } });
    expect(res.statusCode).toBe(401);
  });
});
