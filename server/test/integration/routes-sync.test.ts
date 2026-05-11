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

const validDoc = (slug: string) =>
  `---\ntags: [decision]\nstatus: active\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Why\n\n${slug}\n`;

describe('routes/sync (integration)', () => {
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

  it('ingests a batch and reports per-doc status', async () => {
    const res = await app.inject({
      method: 'POST', url: '/v1/sync',
      headers: { authorization: `Bearer ${key}` },
      payload: {
        repo: 'tolvi',
        documents: [
          { path: 'decisions/2026-04-12-a.md', content: validDoc('a') },
          { path: 'decisions/2026-04-12-b.md', content: validDoc('b') },
        ],
      },
    });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.summary.created).toBe(2);
    expect(body.results).toHaveLength(2);
  });

  it('reports per-doc failures without aborting the batch', async () => {
    const res = await app.inject({
      method: 'POST', url: '/v1/sync',
      headers: { authorization: `Bearer ${key}` },
      payload: {
        repo: 'tolvi',
        documents: [
          { path: 'decisions/2026-04-12-good.md', content: validDoc('g') },
          { path: 'memos/bad.md', content: validDoc('b') },        // wrong dir prefix
        ],
      },
    });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.summary.created).toBe(1);
    expect(body.summary.failed).toBe(1);
  });

  it('returns "unchanged" on idempotent re-sync', async () => {
    const payload = {
      repo: 'tolvi',
      documents: [{ path: 'decisions/2026-04-12-a.md', content: validDoc('a') }],
    };
    const headers = { authorization: `Bearer ${key}` };
    await app.inject({ method: 'POST', url: '/v1/sync', headers, payload });
    const res2 = await app.inject({ method: 'POST', url: '/v1/sync', headers, payload });
    expect(res2.json().summary.unchanged).toBe(1);
  });

  it('rejects batches over 500 docs at the schema layer', async () => {
    const docs = Array.from({ length: 501 }, (_, i) => ({
      path: `decisions/2026-04-12-d${i}.md`, content: validDoc(`d${i}`),
    }));
    const res = await app.inject({
      method: 'POST', url: '/v1/sync',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', documents: docs },
    });
    expect(res.statusCode).toBe(400);
  });
});
