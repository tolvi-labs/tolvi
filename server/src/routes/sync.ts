import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { ingestDocument, type IngestResult } from '../ingest/pipeline.js';

const SyncRequest = z.object({
  repo: z.string().min(1).max(80),
  documents: z.array(z.object({
    path: z.string().min(1),
    content: z.string().min(1),
  })).min(1).max(500),
});

export async function syncRoutes(app: FastifyInstance): Promise<void> {
  app.post('/v1/sync', {
    preHandler: app.requireAuth,
    schema: { body: SyncRequest },
  }, async (req) => {
    const body = req.body as z.infer<typeof SyncRequest>;
    const summary = { created: 0, updated: 0, unchanged: 0, failed: 0 };
    const results: Array<{
      path: string; status: string; document_id: string | null;
      error: { code: string; message: string } | null;
    }> = [];

    // Serial: Ollama is the bottleneck; parallelism just queues at it
    for (const doc of body.documents) {
      let result: IngestResult;
      try {
        result = await ingestDocument(app.db, app.embedding, {
          workspaceId: req.workspaceId!,
          repoSlug: body.repo,
          path: doc.path,
          content: doc.content,
        });
      } catch (err) {
        summary.failed++;
        results.push({
          path: doc.path, status: 'failed', document_id: null,
          error: { code: 'internal_error', message: (err as Error).message },
        });
        continue;
      }

      if (result.status === 'failed') {
        summary.failed++;
        results.push({
          path: doc.path, status: 'failed', document_id: null,
          error: { code: result.error.code, message: result.error.message },
        });
      } else if (result.status === 'unchanged') {
        summary.unchanged++;
        results.push({ path: doc.path, status: 'unchanged', document_id: result.document.id, error: null });
      } else {
        summary[result.status]++;
        results.push({ path: doc.path, status: result.status, document_id: result.document.id, error: null });
      }
    }

    return { results, summary };
  });
}
