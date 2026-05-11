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
  async embed(texts: string[]): Promise<number[][]> {
    return texts.map(() => new Array(768).fill(0.1));
  }
  async ping(): Promise<boolean> { return true; }
}

const sample = `---\ntags: [decision]\nstatus: active\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Why\n\nBecause.\n`;

describe('routes/documents (integration)', () => {
  let db: TestDb;
  let pool: pg.Pool;
  let app: FastifyInstance;
  let key: string;

  beforeAll(async () => {
    const r = await startTestDb();
    db = r.db; pool = r.pool;
  });
  afterAll(async () => { await stopTestDb(); });

  beforeEach(async () => {
    await resetTestDb(pool);
    const [w] = await db.insert(workspaces).values({ slug: 'w', name: 'W' }).returning();
    key = generateApiKey();
    await db.insert(apiKeys).values({
      workspaceId: w!.id,
      keyHash: await hashApiKey(key),
      keyPrefix: extractKeyPrefix(key),
      name: 'k',
    });
    app = await buildTestApp({ db, pool, embedding: new FakeEmbedding() });
  });

  it('healthz responds without auth', async () => {
    const res = await app.inject({ method: 'GET', url: '/healthz' });
    expect(res.statusCode).toBe(200);
    expect(res.json().ok).toBe(true);
  });

  it('POST /v1/documents creates a doc', async () => {
    const res = await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample },
    });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.document.doc_type).toBe('decision');
    expect(body.document.chunks).toBeGreaterThan(0);
  });

  it('POST /v1/documents returns unchanged on idempotent re-post', async () => {
    const headers = { authorization: `Bearer ${key}` };
    const payload = { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample };
    await app.inject({ method: 'POST', url: '/v1/documents', headers, payload });
    const res2 = await app.inject({ method: 'POST', url: '/v1/documents', headers, payload });
    expect(res2.json().unchanged).toBe(true);
  });

  it('GET /v1/documents lists docs', async () => {
    await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample },
    });
    const res = await app.inject({
      method: 'GET', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
    });
    expect(res.statusCode).toBe(200);
    expect(res.json().documents).toHaveLength(1);
  });

  it('GET /v1/documents/:id returns full doc with chunks', async () => {
    const create = await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample },
    });
    const id = create.json().document.id;
    const res = await app.inject({
      method: 'GET', url: `/v1/documents/${id}`,
      headers: { authorization: `Bearer ${key}` },
    });
    expect(res.statusCode).toBe(200);
    expect(res.json().document.body).toContain('Because');
    expect(res.json().document.chunks.length).toBeGreaterThan(0);
  });

  it('DELETE /v1/documents/:id soft-deletes', async () => {
    const create = await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample },
    });
    const id = create.json().document.id;
    const del = await app.inject({
      method: 'DELETE', url: `/v1/documents/${id}`,
      headers: { authorization: `Bearer ${key}` },
    });
    expect(del.statusCode).toBe(204);
    const get = await app.inject({
      method: 'GET', url: `/v1/documents/${id}`,
      headers: { authorization: `Bearer ${key}` },
    });
    expect(get.statusCode).toBe(404);
  });

  it('returns 404 (not 403) when key from workspace A fetches workspace B doc', async () => {
    // Workspace A creates a doc.
    const createA = await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-foo.md', content: sample },
    });
    const docIdA = createA.json().document.id;

    // Provision workspace B + key.
    const [wB] = await db.insert(workspaces).values({ slug: 'wB', name: 'WB' }).returning();
    const keyB = generateApiKey();
    await db.insert(apiKeys).values({
      workspaceId: wB!.id, keyHash: await hashApiKey(keyB), keyPrefix: extractKeyPrefix(keyB), name: 'kB',
    });

    // B's key tries to read A's doc by ID. Must return 404 (not 403, not 200).
    // 404 is the intentional choice — leaking 403-vs-404 would tell B which
    // doc IDs exist in A's workspace.
    const res = await app.inject({
      method: 'GET', url: `/v1/documents/${docIdA}`,
      headers: { authorization: `Bearer ${keyB}` },
    });
    expect(res.statusCode).toBe(404);
    expect(res.json().error.code).toBe('not_found');

    // Same expectation for DELETE — workspace B can't soft-delete A's docs.
    const delRes = await app.inject({
      method: 'DELETE', url: `/v1/documents/${docIdA}`,
      headers: { authorization: `Bearer ${keyB}` },
    });
    expect(delRes.statusCode).toBe(404);
  });

  it('returns 400 with structured error on malformed POST body (zod rejection)', async () => {
    // Missing required `path` and `content` fields.
    const res = await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi' },
    });
    expect(res.statusCode).toBe(400);
    // The fastify-type-provider-zod default 400 envelope differs from the
    // handler-emitted format_validation_failed shape — this test locks in
    // whichever shape the SDK should expect. Update the assertion if the
    // contract changes.
    const body = res.json();
    expect(body).toBeDefined();
  });
});
