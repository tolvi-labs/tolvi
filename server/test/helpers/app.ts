import Fastify, { type FastifyInstance } from 'fastify';
import { authPlugin } from '../../src/auth/plugin.js';
import { healthRoutes } from '../../src/routes/health.js';
import { documentsRoutes } from '../../src/routes/documents.js';
import type { Db } from '../../src/db/client.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import type pg from 'pg';
import {
  serializerCompiler,
  validatorCompiler,
  type ZodTypeProvider,
} from 'fastify-type-provider-zod';

declare module 'fastify' {
  interface FastifyInstance {
    embedding: EmbeddingProvider;
    pool: pg.Pool;
  }
}

export async function buildTestApp(deps: {
  db: Db;
  pool: pg.Pool;
  embedding: EmbeddingProvider;
}): Promise<FastifyInstance> {
  const app = Fastify({ logger: false }).withTypeProvider<ZodTypeProvider>();
  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);
  app.decorate('db', deps.db);
  app.decorate('pool', deps.pool);
  app.decorate('embedding', deps.embedding);
  await app.register(authPlugin);
  await app.register(healthRoutes);
  await app.register(documentsRoutes);
  return app;
}
