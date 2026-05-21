import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { eq, sql } from 'drizzle-orm';
import { repos, documents } from '../db/schema/index.js';
import { IsoTimestamp } from './_responses.js';

const ListReposResponse = z.object({
  repos: z.array(z.object({
    id: z.string().uuid(),
    slug: z.string(),
    remote_url: z.string().nullable(),
    // Postgres COUNT() comes back as a string from pg's BIGINT handling
    // even though Drizzle's sql<number> type-asserts otherwise. Accept both.
    document_count: z.union([z.number(), z.string()]),
    last_synced_at: IsoTimestamp.nullable(),
  })),
});

export async function reposRoutes(app: FastifyInstance): Promise<void> {
  app.get('/v1/repos', {
    preHandler: app.requireAuth,
    schema: { response: { 200: ListReposResponse } },
  }, async (req) => {
    const rows = await app.db
      .select({
        id: repos.id,
        slug: repos.slug,
        remote_url: repos.remoteUrl,
        document_count: sql<number>`count(${documents.id}) FILTER (WHERE ${documents.deletedAt} IS NULL)`.as('document_count'),
        last_synced_at: sql<Date | null>`max(${documents.updatedAt})`.as('last_synced_at'),
      })
      .from(repos)
      .leftJoin(documents, eq(documents.repoId, repos.id))
      .where(eq(repos.workspaceId, req.workspaceId!))
      .groupBy(repos.id);
    return { repos: rows };
  });
}
