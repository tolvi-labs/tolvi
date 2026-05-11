import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { workspaces, documents, chunks } from '../../src/db/schema/index.js';
import { ingestDocument } from '../../src/ingest/pipeline.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import { eq } from 'drizzle-orm';
import type pg from 'pg';

// In-test fake embedding provider — deterministic per text
class FakeEmbedding implements EmbeddingProvider {
  readonly dimension = 768;
  async embed(texts: string[]): Promise<number[][]> {
    return texts.map((t) => {
      const v = new Array(this.dimension).fill(0);
      for (let i = 0; i < t.length; i++) {
        v[i % this.dimension] += t.charCodeAt(i) / 1000;
      }
      return v;
    });
  }
  async ping(): Promise<boolean> {
    return true;
  }
}

const sampleDecision = `---
tags: [decision]
status: active
date: 2026-04-12
repo: tolvi
---

## Why

We need a primary store.

## How

Use Postgres.

## Outcome

Postgres adopted.
`;

describe('ingestDocument (integration)', () => {
  let db: TestDb;
  let pool: pg.Pool;
  const embedding = new FakeEmbedding();
  let workspaceId: string;

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
    const [w] = await db.insert(workspaces).values({ slug: 'test', name: 'Test' }).returning();
    workspaceId = w!.id;
  });

  it('creates a new document on first ingest', async () => {
    const result = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: sampleDecision,
    });
    expect(result.status).toBe('created');
    if (result.status !== 'created') return;
    expect(result.document.docType).toBe('decision');
    expect(result.document.slug).toBe('postgres');
    expect(result.document.title).toBe('Why');
    expect(result.chunks).toBeGreaterThan(0);

    const chunkRows = await db.select().from(chunks).where(eq(chunks.documentId, result.document.id));
    expect(chunkRows).toHaveLength(result.chunks);
  });

  it('returns "unchanged" on idempotent re-ingest', async () => {
    await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: sampleDecision,
    });
    const second = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: sampleDecision,
    });
    expect(second.status).toBe('unchanged');
  });

  it('returns "updated" when content changes', async () => {
    await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: sampleDecision,
    });
    const modified = sampleDecision.replace('Postgres adopted.', 'Postgres adopted with HA.');
    const second = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: modified,
    });
    expect(second.status).toBe('updated');
  });

  it('rejects malformed frontmatter with format_validation_failed', async () => {
    const bad = `---\ntags: [decision]\nstatus: pending\ndate: 2026-04-12\nrepo: tolvi\n---\n\n## Why\n`;
    const result = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-bad.md',
      content: bad,
    });
    expect(result.status).toBe('failed');
    if (result.status !== 'failed') return;
    expect(result.error.code).toBe('format_validation_failed');
  });

  it('rejects unknown doc-type prefix', async () => {
    const result = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'memos/something.md',
      content: sampleDecision,
    });
    expect(result.status).toBe('failed');
  });

  it('restores a soft-deleted document on re-ingest with new content', async () => {
    // 1. Ingest a doc
    const first = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: sampleDecision,
    });
    expect(first.status).toBe('created');
    if (first.status !== 'created') return;
    const docId = first.document.id;

    // 2. Soft-delete it directly
    await db.update(documents).set({ deletedAt: new Date() }).where(eq(documents.id, docId));
    const afterDelete = await db.select().from(documents).where(eq(documents.id, docId));
    expect(afterDelete[0]?.deletedAt).toBeInstanceOf(Date);

    // 3. Re-ingest with different content (so the unchanged short-circuit can't fire)
    const modified = sampleDecision.replace('Postgres adopted.', 'Postgres adopted with HA.');
    const second = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12-postgres.md',
      content: modified,
    });

    // 4. Should take the update path AND clear deletedAt
    expect(second.status).toBe('updated');
    if (second.status !== 'updated') return;
    expect(second.document.id).toBe(docId);
    expect(second.document.deletedAt).toBeNull();
  });

  it('rejects a decision with no slug after the date prefix', async () => {
    // decisions/YYYY-MM-DD.md (no -slug suffix) is the session convention.
    // For decisions, the slug after the date is required — silently using the
    // date as the slug would conflate the two doc types.
    const result = await ingestDocument(db, embedding, {
      workspaceId,
      repoSlug: 'tolvi',
      path: 'decisions/2026-04-12.md',
      content: sampleDecision,
    });
    expect(result.status).toBe('failed');
    if (result.status !== 'failed') return;
    expect(result.error.code).toBe('format_validation_failed');
    expect(result.error.message).toMatch(/slug/i);
  });
});
