import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { search } from '../search/query.js';
import { SYSTEM_PROMPT, buildUserMessage } from '../ask/prompt.js';
import { extractCitations, scrubUnverifiedCitations } from '../ask/citations.js';

const AskRequest = z.object({
  query: z.string().min(1).max(2000),
  filters: z.object({
    repo: z.union([z.string(), z.array(z.string())]).nullable().optional(),
    doc_type: z.array(z.enum(['decision', 'session', 'pattern'])).nullable().optional(),
    status: z.union([z.array(z.string()), z.literal('any')]).nullable().optional(),
  }).default({}),
  model: z.string().nullable().optional(),
});

const ASK_SEARCH_LIMIT = 8;

export async function askRoutes(app: FastifyInstance): Promise<void> {
  app.post('/v1/ask', {
    preHandler: app.requireAuth,
    schema: { body: AskRequest },
  }, async (req, reply) => {
    const body = req.body as z.infer<typeof AskRequest>;

    let results;
    try {
      const r = await search(
        app.db, app.embedding, req.workspaceId!, body.query,
        {
          repo: body.filters.repo ?? null,
          docType: body.filters.doc_type ?? null,
          status: body.filters.status ?? null,
        },
        ASK_SEARCH_LIMIT
      );
      results = r.results;
    } catch (err) {
      const msg = (err as Error).message;
      if (msg.toLowerCase().includes('embedding') || msg.toLowerCase().includes('ollama')) {
        return reply.code(503).send({ error: { code: 'embedding_unavailable', message: msg } });
      }
      throw err;
    }

    const userMessage = buildUserMessage(body.query, results);

    let llmResponse;
    try {
      llmResponse = await app.llm.synthesize({
        systemPrompt: SYSTEM_PROMPT,
        messages: [{ role: 'user', content: userMessage }],
        model: body.model ?? undefined,
      });
    } catch (err) {
      return reply.code(503).send({
        error: { code: 'llm_unavailable', message: (err as Error).message },
      });
    }

    const verifiedSlugs = new Set(results.map((r) => r.slug));
    const cleanAnswer = scrubUnverifiedCitations(llmResponse.text, verifiedSlugs);
    const citedSlugs = extractCitations(llmResponse.text).filter((s) => verifiedSlugs.has(s));
    const citations = citedSlugs.flatMap((slug) =>
      results
        .filter((r) => r.slug === slug)
        .map((r) => ({ slug: r.slug, doc_type: r.docType, document_id: r.documentId })),
    );

    return {
      answer: cleanAnswer,
      citations,
      search_results: results.map((r) => ({
        document_id: r.documentId, doc_type: r.docType, slug: r.slug, title: r.title,
        score: r.score, raw_similarity: r.rawSimilarity,
        matched_chunk: { position: r.matchedChunk.position, content: r.matchedChunk.content, heading_path: r.matchedChunk.headingPath },
      })),
      model: llmResponse.model,
      tokens: { input: llmResponse.tokens.input, output: llmResponse.tokens.output, cache_read: llmResponse.tokens.cacheRead },
    };
  });
}
