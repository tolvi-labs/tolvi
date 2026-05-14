import Fastify, { type FastifyInstance } from 'fastify';
import {
  serializerCompiler,
  validatorCompiler,
  type ZodTypeProvider,
} from 'fastify-type-provider-zod';
import type pg from 'pg';
import { type Config } from './config.js';
import { buildLoggerOptions } from './observability/logging.js';
import { createDb } from './db/client.js';
import { buildEmbeddingProvider } from './embedding/factory.js';
import type { EmbeddingProvider } from './embedding/provider.js';
import { authPlugin } from './auth/plugin.js';
import { healthRoutes } from './routes/health.js';
import { documentsRoutes } from './routes/documents.js';
import { reposRoutes } from './routes/repos.js';
import { syncRoutes } from './routes/sync.js';
import { searchRoutes } from './routes/search.js';

declare module 'fastify' {
  interface FastifyInstance {
    config: Config;
    pool: pg.Pool;
    embedding: EmbeddingProvider;
  }
}

export async function buildApp(cfg: Config): Promise<FastifyInstance> {
  const app = Fastify({
    logger: buildLoggerOptions(cfg),
    genReqId: () => crypto.randomUUID(),
  }).withTypeProvider<ZodTypeProvider>();

  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);

  const { db, pool } = createDb(cfg.databaseUrl);
  const embedding = buildEmbeddingProvider(cfg);

  app.decorate('config', cfg);
  app.decorate('db', db);
  app.decorate('pool', pool);
  app.decorate('embedding', embedding);

  app.addHook('onClose', async () => { await pool.end(); });

  await app.register(authPlugin);
  await app.register(healthRoutes);
  await app.register(documentsRoutes);
  await app.register(reposRoutes);
  await app.register(syncRoutes);
  await app.register(searchRoutes);

  return app;
}
