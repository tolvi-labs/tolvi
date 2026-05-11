import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { buildTestApp } from '../helpers/app.js';
import { workspaces, apiKeys } from '../../src/db/schema/index.js';
import { generateApiKey, hashApiKey, extractKeyPrefix } from '../../src/auth/api-key.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import type { FastifyInstance } from 'fastify';
import type pg from 'pg';

class FakeEmbedding implements EmbeddingProvider {
  readonly dimension = 768;
  async embed(texts: string[]): Promise<number[][]> { return texts.map(() => new Array(768).fill(0.1)); }
  async ping(): Promise<boolean> { return true; }
}

const sample = `---\ntags: [decision]\nstatus: active\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Why\n\nBecause.\n`;

describe('routes/repos (integration)', () => {
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
    app = await buildTestApp({ db, pool, embedding: new FakeEmbedding() });
  });

  it('returns empty list when no repos exist', async () => {
    const res = await app.inject({ method: 'GET', url: '/v1/repos', headers: { authorization: `Bearer ${key}` } });
    expect(res.statusCode).toBe(200);
    expect(res.json().repos).toEqual([]);
  });

  it('returns repo list with document counts after ingest', async () => {
    await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample },
    });
    const res = await app.inject({ method: 'GET', url: '/v1/repos', headers: { authorization: `Bearer ${key}` } });
    expect(res.statusCode).toBe(200);
    expect(res.json().repos).toHaveLength(1);
    expect(res.json().repos[0].slug).toBe('tolvi');
    expect(Number(res.json().repos[0].document_count)).toBe(1);
  });
});
