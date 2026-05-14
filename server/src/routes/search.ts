import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { search } from '../search/query.js';

const SearchRequest = z.object({
  query: z.string().min(1).max(2000),
  limit: z.number().int().positive().max(50).default(10),
  filters: z.object({
    repo: z.union([z.string(), z.array(z.string())]).nullable().optional(),
    doc_type: z.array(z.enum(['decision', 'session', 'pattern'])).nullable().optional(),
    status: z.union([z.array(z.string()), z.literal('any')]).nullable().optional(),
  }).default({}),
});

export async function searchRoutes(app: FastifyInstance): Promise<void> {
  app.post('/v1/search', {
    preHandler: app.requireAuth,
    schema: { body: SearchRequest },
  }, async (req, reply) => {
    const body = req.body as z.infer<typeof SearchRequest>;
    try {
      const { results, total } = await search(
        app.db, app.embedding, req.workspaceId!, body.query,
        {
          repo: body.filters.repo ?? null,
          docType: body.filters.doc_type ?? null,
          status: body.filters.status ?? null,
        },
        body.limit
      );
      return {
        results: results.map((r) => ({
          document_id: r.documentId, doc_type: r.docType, slug: r.slug, title: r.title,
          score: r.score, raw_similarity: r.rawSimilarity,
          matched_chunk: {
            position: r.matchedChunk.position,
            content: r.matchedChunk.content,
            heading_path: r.matchedChunk.headingPath,
          },
        })),
        total,
      };
    } catch (err) {
      const msg = (err as Error).message;
      if (msg.toLowerCase().includes('embedding') || msg.toLowerCase().includes('ollama')) {
        return reply.code(503).send({ error: { code: 'embedding_unavailable', message: msg } });
      }
      throw err;
    }
  });
}
