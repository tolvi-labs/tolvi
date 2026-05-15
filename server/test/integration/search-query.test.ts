import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { workspaces } from '../../src/db/schema/index.js';
import { ingestDocument } from '../../src/ingest/pipeline.js';
import { search } from '../../src/search/query.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import type pg from 'pg';

class DeterministicEmbedding implements EmbeddingProvider {
  readonly dimension = 768;
  async embed(texts: string[]): Promise<number[][]> {
    return texts.map((t) => {
      const v = new Array(this.dimension).fill(0);
      for (let i = 0; i < t.length; i++) v[i % this.dimension] += t.charCodeAt(i) / 1000;
      // Normalize so cosine similarity is meaningful
      const norm = Math.sqrt(v.reduce((s, x) => s + x * x, 0));
      return norm > 0 ? v.map((x) => x / norm) : v;
    });
  }
  async ping(): Promise<boolean> { return true; }
}

const doc = (slug: string, body: string, status = 'active', docType = 'decision') => `---
tags: [${docType}]
status: ${status}
${docType === 'pattern' ? '' : `date: 2026-04-12\nrepo: tolvi`}
---

## ${slug}

${body}
`;

describe('search (integration)', () => {
  let db: TestDb; let pool: pg.Pool;
  const embedding = new DeterministicEmbedding();
  let workspaceId: string;

  beforeAll(async () => { const r = await startTestDb(); db = r.db; pool = r.pool; });
  afterAll(async () => { await stopTestDb(); });

  beforeEach(async () => {
    await resetTestDb(pool);
    const [w] = await db.insert(workspaces).values({ slug: 'w', name: 'W' }).returning();
    workspaceId = w!.id;
  });

  it('returns results from all ingested docs', async () => {
    // The DeterministicEmbedding is char-code-positional, not semantic — so
    // we can't assert that "postgres database" ranks the postgres doc
    // higher than the redis doc (the synthetic similarity has no semantic
    // meaning). Instead, verify the SQL pipeline returns BOTH docs and
    // delivers a properly-shaped SearchResult. Semantic ranking quality is
    // covered by the e2e test (server/test/e2e/full-stack.test.ts) which
    // uses a real Ollama embedding.
    await ingestDocument(db, embedding, {
      workspaceId, repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: doc('postgres choice', 'we use postgres because it has pgvector and json'),
    });
    await ingestDocument(db, embedding, {
      workspaceId, repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-redis.md',
      content: doc('redis cache', 'cache invalidation is hard for stale reads'),
    });
    const { results } = await search(db, embedding, workspaceId, 'postgres database', { repo: null, docType: null, status: null }, 10);
    expect(results).toHaveLength(2);
    const slugs = results.map((r) => r.slug).sort();
    expect(slugs).toEqual(['postgres', 'redis']);
    // Sanity-check the SearchResult shape on the first hit.
    expect(results[0]?.score).toBeTypeOf('number');
    expect(results[0]?.matchedChunk.content).toBeTypeOf('string');
  });

  it('excludes superseded by default', async () => {
    await ingestDocument(db, embedding, {
      workspaceId, repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-old.md',
      content: doc('old choice', 'database choice', 'superseded'),
    });
    const { results } = await search(db, embedding, workspaceId, 'database', { repo: null, docType: null, status: null }, 10);
    expect(results).toHaveLength(0);
  });

  it('includes superseded when status: any is passed', async () => {
    await ingestDocument(db, embedding, {
      workspaceId, repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-old.md',
      content: doc('old choice', 'database choice', 'superseded'),
    });
    const { results } = await search(db, embedding, workspaceId, 'database', { repo: null, docType: null, status: 'any' }, 10);
    expect(results.length).toBeGreaterThan(0);
  });

  it('respects repo filter', async () => {
    await ingestDocument(db, embedding, { workspaceId, repoSlug: 'tolvi', path: 'decisions/2026-04-12-a.md', content: doc('a', 'about postgres') });
    await ingestDocument(db, embedding, { workspaceId, repoSlug: 'other', path: 'decisions/2026-04-12-b.md', content: doc('b', 'about postgres') });
    const { results } = await search(db, embedding, workspaceId, 'postgres', { repo: 'tolvi', docType: null, status: null }, 10);
    expect(results.every((r) => r.slug !== 'b')).toBe(true);
  });
});
