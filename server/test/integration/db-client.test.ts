import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { workspaces, repos, documents, chunks } from '../../src/db/schema/index.js';
import { eq } from 'drizzle-orm';
import type pg from 'pg';

describe('db client + schema (integration)', () => {
  let db: TestDb;
  let pool: pg.Pool;

  beforeAll(async () => {
    const result = await startTestDb();
    db = result.db;
    pool = result.pool;
  });

  afterAll(async () => {
    await stopTestDb();
  });

  beforeEach(async () => {
    await resetTestDb(pool);
  });

  it('inserts and queries a workspace', async () => {
    const [w] = await db.insert(workspaces).values({ slug: 'test', name: 'Test Workspace' }).returning();
    expect(w?.slug).toBe('test');
    expect(w?.id).toMatch(/^[0-9a-f-]{36}$/);

    const found = await db.select().from(workspaces).where(eq(workspaces.id, w!.id));
    expect(found).toHaveLength(1);
    expect(found[0]?.name).toBe('Test Workspace');
  });

  it('cascades repos when workspace is deleted', async () => {
    const [w] = await db.insert(workspaces).values({ slug: 'a', name: 'A' }).returning();
    await db.insert(repos).values({ workspaceId: w!.id, slug: 'tolvi' });

    await db.delete(workspaces).where(eq(workspaces.id, w!.id));

    const reposAfter = await db.select().from(repos).where(eq(repos.workspaceId, w!.id));
    expect(reposAfter).toHaveLength(0);
  });

  it('enforces workspace+repo slug uniqueness', async () => {
    const [w] = await db.insert(workspaces).values({ slug: 'a', name: 'A' }).returning();
    await db.insert(repos).values({ workspaceId: w!.id, slug: 'tolvi' });

    await expect(
      db.insert(repos).values({ workspaceId: w!.id, slug: 'tolvi' })
    ).rejects.toThrow(/duplicate key|unique/i);
  });

  it('proves pgvector loaded and embedding dimension matches EMBEDDING_DIM', async () => {
    // End-to-end smoke test for the RAG foundation: insert a chunk with a
    // 768-dim vector (the default for nomic-embed-text), read it back, verify
    // round-trip. If pgvector isn't loaded the migration fails before we get
    // here; if EMBEDDING_DIM disagrees with the column's vector(N), the insert
    // fails. Closes the silent-corruption window flagged in the Task 5 review.
    const [w] = await db.insert(workspaces).values({ slug: 'v', name: 'V' }).returning();
    const [r] = await db.insert(repos).values({ workspaceId: w!.id, slug: 'tolvi' }).returning();
    const [doc] = await db.insert(documents).values({
      workspaceId: w!.id,
      repoId: r!.id,
      docType: 'decision',
      path: 'decisions/2026-01-01-test.md',
      slug: 'test',
      title: 'Test',
      body: 'body',
      contentHash: 'sha256-fake',
    }).returning();

    const fakeEmbedding = new Array(768).fill(0.1);
    await db.insert(chunks).values({
      workspaceId: w!.id,
      documentId: doc!.id,
      position: 0,
      content: 'chunk content',
      embedding: fakeEmbedding,
      headingPath: ['Test'],
    });

    const found = await db.select().from(chunks).where(eq(chunks.documentId, doc!.id));
    expect(found).toHaveLength(1);
    expect(found[0]?.embedding).toHaveLength(768);
    expect(found[0]?.embedding?.[0]).toBeCloseTo(0.1, 4);
  });
});
