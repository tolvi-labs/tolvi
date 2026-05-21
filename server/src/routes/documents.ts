import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { eq, and, isNull, desc, inArray } from 'drizzle-orm';
import { documents, repos, chunks } from '../db/schema/index.js';
import { ingestDocument } from '../ingest/pipeline.js';
import { ErrorEnvelope, IsoTimestamp, HeadingPath } from './_responses.js';

const IngestRequest = z.object({
  repo: z.string().min(1).max(80),
  path: z.string().min(1),
  content: z.string().min(1),
});

const ListQuery = z.object({
  repo: z.string().optional(),
  doc_type: z.enum(['decision', 'session', 'pattern']).optional(),
  status: z.string().optional(),                              // CSV
  limit: z.coerce.number().int().positive().max(200).default(50),
});

// POST /v1/documents — single response schema covers both "created/updated"
// (has `chunks` count) and "unchanged" (has `unchanged: true`, no chunks).
const PostDocumentResponse = z.object({
  document: z.object({
    id: z.string().uuid(),
    repo_id: z.string().uuid(),
    doc_type: z.string(),
    slug: z.string(),
    status: z.string(),
    title: z.string().nullable(),
    content_hash: z.string(),
    chunks: z.number().int().nonnegative().optional(),
    embedded_at: IsoTimestamp,
  }),
  unchanged: z.literal(true).optional(),
});

// GET /v1/documents — list response (lightweight per-doc shape; no body, no chunks).
const ListDocumentsResponse = z.object({
  documents: z.array(z.object({
    id: z.string().uuid(),
    repo_id: z.string().uuid(),
    doc_type: z.string(),
    slug: z.string(),
    status: z.string(),
    title: z.string().nullable(),
    date: IsoTimestamp.nullable(),
    updated_at: IsoTimestamp,
  })),
  next_cursor: z.string().nullable(),
});

// GET /v1/documents/:id — full doc with body, frontmatter, and chunks.
const GetDocumentResponse = z.object({
  document: z.object({
    id: z.string().uuid(),
    repo_id: z.string().uuid(),
    doc_type: z.string(),
    slug: z.string(),
    status: z.string(),
    title: z.string().nullable(),
    body: z.string(),
    frontmatter: z.record(z.string(), z.unknown()),
    date: IsoTimestamp.nullable(),
    content_hash: z.string(),
    created_at: IsoTimestamp,
    updated_at: IsoTimestamp,
    chunks: z.array(z.object({
      position: z.number().int().nonnegative(),
      content: z.string(),
      heading_path: HeadingPath,
    })),
  }),
});

export async function documentsRoutes(app: FastifyInstance): Promise<void> {
  // POST /v1/documents — single ingest
  app.post('/v1/documents', {
    preHandler: app.requireAuth,
    schema: {
      body: IngestRequest,
      response: {
        200: PostDocumentResponse,
        400: ErrorEnvelope,
        503: ErrorEnvelope,
      },
    },
  }, async (req, reply) => {
    const body = req.body as z.infer<typeof IngestRequest>;
    let result;
    try {
      result = await ingestDocument(app.db, app.embedding, {
        workspaceId: req.workspaceId!,
        repoSlug: body.repo,
        path: body.path,
        content: body.content,
      });
    } catch (err) {
      // The pipeline throws on embedding-provider failures (Task 13's
      // intentional design — the slow remote call lives outside the typed
      // IngestResult discriminator). Map embedding-related errors to the
      // documented 503 envelope; everything else propagates to Fastify's
      // default 500 handler with a logged error.
      const msg = err instanceof Error ? err.message : String(err);
      if (msg.toLowerCase().includes('embedding') || msg.toLowerCase().includes('ollama')) {
        return reply.code(503).send({
          error: { code: 'embedding_unavailable', message: msg },
        });
      }
      throw err;
    }
    if (result.status === 'failed') {
      return reply.code(400).send({ error: result.error });
    }
    if (result.status === 'unchanged') {
      return reply.send({
        document: {
          id: result.document.id,
          repo_id: result.document.repoId,
          doc_type: result.document.docType,
          slug: result.document.slug,
          status: result.document.status,
          title: result.document.title,
          content_hash: result.document.contentHash,
          embedded_at: result.document.updatedAt,
        },
        unchanged: true,
      });
    }
    return reply.send({
      document: {
        id: result.document.id,
        repo_id: result.document.repoId,
        doc_type: result.document.docType,
        slug: result.document.slug,
        status: result.document.status,
        title: result.document.title,
        content_hash: result.document.contentHash,
        chunks: result.chunks,
        embedded_at: result.document.updatedAt,
      },
    });
  });

  // GET /v1/documents — list
  app.get('/v1/documents', {
    preHandler: app.requireAuth,
    schema: {
      querystring: ListQuery,
      response: { 200: ListDocumentsResponse },
    },
  }, async (req) => {
    const q = req.query as z.infer<typeof ListQuery>;
    const conditions = [eq(documents.workspaceId, req.workspaceId!), isNull(documents.deletedAt)];
    if (q.doc_type) conditions.push(eq(documents.docType, q.doc_type));
    if (q.status) {
      const statuses = q.status.split(',');
      conditions.push(inArray(documents.status, statuses));
    }
    if (q.repo) {
      const r = await app.db
        .select()
        .from(repos)
        .where(and(eq(repos.workspaceId, req.workspaceId!), eq(repos.slug, q.repo)));
      if (r[0]) conditions.push(eq(documents.repoId, r[0].id));
      else return { documents: [], next_cursor: null };
    }
    const rows = await app.db
      .select({
        id: documents.id,
        repo_id: documents.repoId,
        doc_type: documents.docType,
        slug: documents.slug,
        status: documents.status,
        title: documents.title,
        date: documents.date,
        updated_at: documents.updatedAt,
      })
      .from(documents)
      .where(and(...conditions))
      .orderBy(desc(documents.updatedAt))
      .limit(q.limit);
    return { documents: rows, next_cursor: null };
  });

  // GET /v1/documents/:id
  app.get('/v1/documents/:id', {
    preHandler: app.requireAuth,
    schema: {
      params: z.object({ id: z.string().uuid() }),
      response: {
        200: GetDocumentResponse,
        404: ErrorEnvelope,
      },
    },
  }, async (req, reply) => {
    const { id } = req.params as { id: string };
    const docs = await app.db
      .select()
      .from(documents)
      .where(and(
        eq(documents.id, id),
        eq(documents.workspaceId, req.workspaceId!),
        isNull(documents.deletedAt),
      ));
    const doc = docs[0];
    if (!doc) return reply.code(404).send({ error: { code: 'not_found', message: 'Document not found' } });
    const chunkRows = await app.db
      .select({
        position: chunks.position,
        content: chunks.content,
        heading_path: chunks.headingPath,
      })
      .from(chunks)
      .where(eq(chunks.documentId, doc.id))
      .orderBy(chunks.position);
    return {
      document: {
        id: doc.id,
        repo_id: doc.repoId,
        doc_type: doc.docType,
        slug: doc.slug,
        status: doc.status,
        title: doc.title,
        body: doc.body,
        frontmatter: doc.frontmatter,
        date: doc.date,
        content_hash: doc.contentHash,
        created_at: doc.createdAt,
        updated_at: doc.updatedAt,
        chunks: chunkRows,
      },
    };
  });

  // DELETE /v1/documents/:id — soft delete
  app.delete('/v1/documents/:id', {
    preHandler: app.requireAuth,
    schema: {
      params: z.object({ id: z.string().uuid() }),
      // 204 = empty body on success; z.null() is the conventional shape
      // for "no content" in zod schemas paired with fastify-type-provider-zod.
      response: {
        204: z.null(),
        404: ErrorEnvelope,
      },
    },
  }, async (req, reply) => {
    const { id } = req.params as { id: string };
    const updated = await app.db
      .update(documents)
      .set({ deletedAt: new Date() })
      .where(and(
        eq(documents.id, id),
        eq(documents.workspaceId, req.workspaceId!),
        isNull(documents.deletedAt),
      ))
      .returning({ id: documents.id });
    if (updated.length === 0) {
      return reply.code(404).send({ error: { code: 'not_found', message: 'Document not found' } });
    }
    return reply.code(204).send();
  });
}
