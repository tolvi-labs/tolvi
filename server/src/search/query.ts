import { sql } from 'drizzle-orm';
import type { Db } from '../db/client.js';
import type { EmbeddingProvider } from '../embedding/provider.js';
import {
  RECENCY_FLOOR, RECENCY_AMPLITUDE, RECENCY_HALF_LIFE_DAYS,
  SESSION_DOWN_WEIGHT, DEFAULT_SURFACED_STATUSES, ALL_STATUSES,
} from './ranking.js';

export type SearchFilters = {
  repo?: string | string[] | null;
  docType?: Array<'decision' | 'session' | 'pattern'> | null;
  status?: string[] | 'any' | null;
};

export type SearchResult = {
  documentId: string;
  docType: string;
  slug: string;
  title: string;
  score: number;
  rawSimilarity: number;
  matchedChunk: {
    position: number;
    content: string;
    headingPath: string[] | null;
  };
};

function pgVectorLiteral(v: number[]): string {
  return `[${v.join(',')}]`;
}

export async function search(
  db: Db,
  embedding: EmbeddingProvider,
  workspaceId: string,
  query: string,
  filters: SearchFilters,
  limit: number
): Promise<{ results: SearchResult[]; total: number }> {
  const [queryVec] = await embedding.embed([query]);
  if (!queryVec) throw new Error('Embedding returned no vector');
  if (queryVec.some((x) => !Number.isFinite(x))) {
    throw new Error('Embedding vector contains non-finite values');
  }

  const statusList: string[] =
    filters.status === null || filters.status === undefined
      ? [...DEFAULT_SURFACED_STATUSES]
      : filters.status === 'any'
        ? [...ALL_STATUSES]
        : filters.status;

  const repoList: string[] | null =
    filters.repo == null ? null : Array.isArray(filters.repo) ? filters.repo : [filters.repo];

  const docTypeList: string[] | null = filters.docType ?? null;

  const result = await db.execute(sql`
    WITH q AS (SELECT ${pgVectorLiteral(queryVec)}::vector AS v)
    SELECT
      d.id              AS document_id,
      d.doc_type        AS doc_type,
      d.slug            AS slug,
      d.title           AS title,
      c.position        AS chunk_position,
      c.content         AS chunk_content,
      c.heading_path    AS heading_path,
      1 - (c.embedding <=> q.v)                                                    AS raw_similarity,
      (1 - (c.embedding <=> q.v)) *
        (${RECENCY_FLOOR} + ${RECENCY_AMPLITUDE} *
         exp(-EXTRACT(EPOCH FROM (now() - COALESCE(d.date::timestamp, d.created_at)))
             / 86400.0 / ${RECENCY_HALF_LIFE_DAYS})) *
        CASE WHEN d.doc_type = 'session' THEN ${SESSION_DOWN_WEIGHT} ELSE 1.0 END  AS score
    FROM chunks c
    JOIN documents d ON d.id = c.document_id
    JOIN repos r     ON r.id = d.repo_id, q
    WHERE d.workspace_id = ${workspaceId}
      AND d.deleted_at IS NULL
      AND d.status = ANY(${statusList}::text[])
      AND (${repoList === null ? sql`true` : sql`r.slug = ANY(${repoList}::text[])`})
      AND (${docTypeList === null ? sql`true` : sql`d.doc_type = ANY(${docTypeList}::text[])`})
    ORDER BY score DESC
    LIMIT ${limit * 3}
  `);

  // Dedupe by document — highest-scoring chunk per doc wins
  const byDoc = new Map<string, SearchResult>();
  for (const row of result.rows as Array<Record<string, unknown>>) {
    const docId = row.document_id as string;
    if (byDoc.has(docId)) continue;
    byDoc.set(docId, {
      documentId: docId,
      docType: row.doc_type as string,
      slug: row.slug as string,
      title: row.title as string,
      score: Number(row.score),
      rawSimilarity: Number(row.raw_similarity),
      matchedChunk: {
        position: Number(row.chunk_position),
        content: row.chunk_content as string,
        headingPath: row.heading_path as string[] | null,
      },
    });
    if (byDoc.size >= limit) break;
  }

  return { results: Array.from(byDoc.values()), total: byDoc.size };
}
