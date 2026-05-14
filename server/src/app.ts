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

  const { db, pool } = createDb(cfg.databaseUrl);
  const embedding = buildEmbeddingProvider(cfg);
  const llm = buildLlmProvider(cfg);

  app.decorate('config', cfg);
  app.decorate('db', db);
  app.decorate('pool', pool);
  app.decorate('embedding', embedding);
  app.decorate('llm', llm);

  app.addHook('onClose', async () => { await pool.end(); });

  await app.register(authPlugin);
  await app.register(healthRoutes);
  await app.register(documentsRoutes);
  await app.register(reposRoutes);
  await app.register(syncRoutes);
  await app.register(searchRoutes);
  await app.register(askRoutes);

  return app;
}
