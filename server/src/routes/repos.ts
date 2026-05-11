import type { FastifyInstance } from 'fastify';
import { eq, sql } from 'drizzle-orm';
import { repos, documents } from '../db/schema/index.js';

export async function reposRoutes(app: FastifyInstance): Promise<void> {
  app.get('/v1/repos', { preHandler: app.requireAuth }, async (req) => {
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
