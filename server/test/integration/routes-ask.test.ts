import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { buildTestApp } from '../helpers/app.js';
import { workspaces, apiKeys } from '../../src/db/schema/index.js';
import { generateApiKey, hashApiKey, extractKeyPrefix } from '../../src/auth/api-key.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import type { LlmProvider, LlmRequest, LlmResponse } from '../../src/llm/provider.js';
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

class FakeLlm implements LlmProvider {
  readonly defaultModel = 'fake-model';
  public lastReq: LlmRequest | null = null;
  constructor(private readonly responseText: string) {}
  async synthesize(req: LlmRequest): Promise<LlmResponse> {
    this.lastReq = req;
    return {
      text: this.responseText,
      model: this.defaultModel,
      tokens: { input: 100, output: 20, cacheRead: 50 },
    };
  }
  async ping(): Promise<boolean> { return true; }
}

const sample = `---\ntags: [decision]\nstatus: active\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Postgres\n\nWe chose postgres for its pgvector and json support.\n`;

describe('routes/ask (integration)', () => {
  let db: TestDb; let pool: pg.Pool; let app: FastifyInstance; let key: string; let llm: FakeLlm;
  beforeAll(async () => { const r = await startTestDb(); db = r.db; pool = r.pool; });
  afterAll(async () => { await stopTestDb(); });

  beforeEach(async () => {
    await resetTestDb(pool);
    const [w] = await db.insert(workspaces).values({ slug: 'w', name: 'W' }).returning();
    key = generateApiKey();
    await db.insert(apiKeys).values({
      workspaceId: w!.id, keyHash: await hashApiKey(key), keyPrefix: extractKeyPrefix(key), name: 'k',
    });
    llm = new FakeLlm('We chose Postgres because of pgvector. See [[postgres]].');
    app = await buildTestApp({ db, pool, embedding: new DetEmbedding(), llm });
    await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-postgres.md', content: sample },
    });
  });

  it('returns answer with verified citations', async () => {
    const res = await app.inject({
      method: 'POST', url: '/v1/ask',
      headers: { authorization: `Bearer ${key}` },
      payload: { query: 'why postgres' },
    });
    expect(res.statusCode).toBe(200);
    const body = res.json();
    expect(body.answer).toContain('[[postgres]]');
    expect(body.citations).toHaveLength(1);
    expect(body.citations[0].slug).toBe('postgres');
    expect(body.tokens.cache_read).toBe(50);
  });

  it('scrubs unverified citations', async () => {
    llm = new FakeLlm('Hallucinated [[fake-slug]] reference.');
    app = await buildTestApp({ db, pool, embedding: new DetEmbedding(), llm });
    await app.inject({
      method: 'POST', url: '/v1/documents',
      headers: { authorization: `Bearer ${key}` },
      payload: { repo: 'tolvi', path: 'decisions/2026-04-12-postgres.md', content: sample },
    });
    const res = await app.inject({
      method: 'POST', url: '/v1/ask',
      headers: { authorization: `Bearer ${key}` },
      payload: { query: 'anything' },
    });
    expect(res.json().answer).toContain('[unverified citation: fake-slug]');
    expect(res.json().citations).toEqual([]);
  });
});
