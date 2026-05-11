import type { FastifyInstance } from 'fastify';
import { z } from 'zod';

export async function healthRoutes(app: FastifyInstance): Promise<void> {
  app.get('/healthz', {
    schema: {
      response: {
        200: z.object({ ok: z.literal(true), version: z.string() }),
      },
    },
  }, async () => ({ ok: true as const, version: '0.1.0' }));

  app.get('/readyz', {
    schema: {
      response: {
        200: z.object({ ok: z.literal(true) }),
        503: z.object({ ok: z.literal(false), checks: z.record(z.boolean()) }),
      },
    },
  }, async (_req, reply) => {
    const checks: Record<string, boolean> = {};
    try {
      // Use raw SQL via pg pool for the simplest health check; Drizzle's
      // db.execute API can vary by version, so a direct pool query is
      // the most stable surface for /readyz.
      await app.pool.query('SELECT 1');
      checks.postgres = true;
    } catch {
      checks.postgres = false;
    }
    try {
      checks.embedding = await app.embedding.ping();
    } catch {
      checks.embedding = false;
    }
    const ok = Object.values(checks).every(Boolean);
    if (ok) return { ok: true as const };
    return reply.code(503).send({ ok: false as const, checks });
  });
}
