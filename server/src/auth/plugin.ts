import fp from 'fastify-plugin';
import { type FastifyInstance, type FastifyRequest, type FastifyReply } from 'fastify';
import { eq, and, isNull } from 'drizzle-orm';
import { apiKeys } from '../db/schema/index.js';
import { extractKeyPrefix, verifyApiKey } from './api-key.js';

declare module 'fastify' {
  interface FastifyInstance {
    requireAuth: (req: FastifyRequest, reply: FastifyReply) => Promise<void>;
  }
  interface FastifyRequest {
    workspaceId?: string;
    apiKeyId?: string;
  }
}

const BEARER_PREFIX = 'Bearer ';
const UNAUTHORIZED_CODE = 'unauthorized';

export const authPlugin = fp(
  async (app: FastifyInstance) => {
    // Fail fast at register time if the DB decorator hasn't been set, instead
    // of crashing with `Cannot read properties of undefined` on the first
    // authenticated request. Production wires app.db in buildApp; tests in
    // buildTestApp.
    if (!app.hasDecorator('db')) {
      throw new Error('auth plugin requires app.db decorator (register the db plugin first)');
    }

    app.decorate('requireAuth', async (req: FastifyRequest, reply: FastifyReply) => {
      const header = req.headers.authorization;
      if (!header || !header.startsWith(BEARER_PREFIX)) {
        return reply.code(401).send({
          error: { code: UNAUTHORIZED_CODE, message: 'Missing or malformed Authorization header' },
        });
      }

      const key = header.slice(BEARER_PREFIX.length).trim();
      let prefix: string;
      try {
        prefix = extractKeyPrefix(key);
      } catch {
        return reply.code(401).send({
          error: { code: UNAUTHORIZED_CODE, message: 'Invalid API key format' },
        });
      }

      // Lookup candidates by prefix (typically <5 rows in practice). The
      // sequential argon2 verify of multiple candidates leaks small timing
      // differences across iteration order — acceptable at v1 prefix length
      // (8 chars after `tlv_`) where collisions are rare; revisit if prefix
      // length shrinks or per-workspace key counts grow large.
      const candidates = await app.db
        .select()
        .from(apiKeys)
        .where(and(eq(apiKeys.keyPrefix, prefix), isNull(apiKeys.revokedAt)));

      for (const candidate of candidates) {
        if (await verifyApiKey(key, candidate.keyHash)) {
          req.workspaceId = candidate.workspaceId;
          req.apiKeyId = candidate.id;
          // Update last_used_at asynchronously; failure is non-fatal. Re-checks
          // revocation in the WHERE clause to avoid touching a key revoked
          // between SELECT and UPDATE.
          app.db
            .update(apiKeys)
            .set({ lastUsedAt: new Date() })
            .where(and(eq(apiKeys.id, candidate.id), isNull(apiKeys.revokedAt)))
            .catch((err) => app.log.warn({ err, apiKeyId: candidate.id }, 'last_used_at update failed'));
          return;
        }
      }

      return reply.code(401).send({
        error: { code: UNAUTHORIZED_CODE, message: 'Invalid API key' },
      });
    });
  },
  { name: 'auth' }
);
