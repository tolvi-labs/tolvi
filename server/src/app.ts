import Fastify, { type FastifyInstance } from 'fastify';
import fastifySwagger from '@fastify/swagger';
import fastifySwaggerUi from '@fastify/swagger-ui';
import {
  jsonSchemaTransform,
  serializerCompiler,
  validatorCompiler,
  type ZodTypeProvider,
} from 'fastify-type-provider-zod';
import { ZodError } from 'zod';
import type pg from 'pg';
import { type Config } from './config.js';
import { buildLoggerOptions } from './observability/logging.js';
import { createDb } from './db/client.js';
import { buildEmbeddingProvider } from './embedding/factory.js';
import type { EmbeddingProvider } from './embedding/provider.js';
import { buildLlmProvider } from './llm/factory.js';
import type { LlmProvider } from './llm/provider.js';
import { authPlugin } from './auth/plugin.js';
import { healthRoutes } from './routes/health.js';
import { documentsRoutes } from './routes/documents.js';
import { reposRoutes } from './routes/repos.js';
import { syncRoutes } from './routes/sync.js';
import { searchRoutes } from './routes/search.js';
import { askRoutes } from './routes/ask.js';

declare module 'fastify' {
  interface FastifyInstance {
    config: Config;
    pool: pg.Pool;
    embedding: EmbeddingProvider;
    llm: LlmProvider;
  }
}

export async function buildApp(cfg: Config): Promise<FastifyInstance> {
  const app = Fastify({
    logger: buildLoggerOptions(cfg),
    genReqId: () => crypto.randomUUID(),
  }).withTypeProvider<ZodTypeProvider>();

  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);

  // Normalize all validation errors (from fastify-type-provider-zod's
  // body/querystring/params validators) to the canonical ErrorEnvelope
  // shape — `{ error: { code, message } }` — so the response matches the
  // `400: ErrorEnvelope` schema declarations on data-plane routes. Without
  // this normalization, fastify-type-provider-zod throws a raw ZodError
  // (which does NOT carry `.validation` or `.code === 'FST_ERR_VALIDATION'`),
  // the default error handler renders the ZodError as JSON, and the
  // serializer rejects that shape against the declared 400 schema —
  // surfacing as a 500 to the client.
  app.setErrorHandler((error, request, reply) => {
    const fastifyError = error as { validation?: unknown; code?: string };
    const isZodError = error instanceof ZodError;
    if (isZodError || fastifyError.validation || fastifyError.code === 'FST_ERR_VALIDATION') {
      const message = isZodError
        ? error.issues
            .map((issue) => `${issue.path.join('.') || '<root>'}: ${issue.message}`)
            .join('; ')
        : (error as Error).message;
      return reply.code(400).send({
        error: { code: 'validation_failed', message },
      });
    }
    request.log.error({ err: error }, 'unhandled route error');
    return reply.send(error);
  });

  const { db, pool } = createDb(cfg.databaseUrl);
  const embedding = buildEmbeddingProvider(cfg);
  const llm = buildLlmProvider(cfg);

  app.decorate('config', cfg);
  app.decorate('db', db);
  app.decorate('pool', pool);
  app.decorate('embedding', embedding);
  app.decorate('llm', llm);

  app.addHook('onClose', async () => { await pool.end(); });

  await app.register(fastifySwagger, {
    openapi: {
      info: {
        title: 'Tolvi Server API',
        description:
          'HTTP API for the Tolvi server. See spec/tolvi-format-v1.md for the vault format contract.',
        version: '0.1.0',
      },
      servers: [{ url: 'http://localhost:3000', description: 'local dev' }],
      components: {
        securitySchemes: {
          bearerAuth: { type: 'http', scheme: 'bearer', bearerFormat: 'tlv_<random>' },
        },
      },
      security: [{ bearerAuth: [] }],
    },
    transform: jsonSchemaTransform,
  });

  if (cfg.nodeEnv === 'development') {
    await app.register(fastifySwaggerUi, { routePrefix: '/docs' });
  }

  await app.register(authPlugin);
  await app.register(healthRoutes);
  await app.register(documentsRoutes);
  await app.register(reposRoutes);
  await app.register(syncRoutes);
  await app.register(searchRoutes);
  await app.register(askRoutes);

  return app;
}
