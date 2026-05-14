import Fastify, { type FastifyInstance } from 'fastify';
import { authPlugin } from '../../src/auth/plugin.js';
import { healthRoutes } from '../../src/routes/health.js';
import { documentsRoutes } from '../../src/routes/documents.js';
import { reposRoutes } from '../../src/routes/repos.js';
import { syncRoutes } from '../../src/routes/sync.js';
import { searchRoutes } from '../../src/routes/search.js';
import { askRoutes } from '../../src/routes/ask.js';
import type { Db } from '../../src/db/client.js';
import type { EmbeddingProvider } from '../../src/embedding/provider.js';
import type { LlmProvider } from '../../src/llm/provider.js';
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
    llm: LlmProvider;
  }
}

export async function buildTestApp(deps: {
  db: Db;
  pool: pg.Pool;
  embedding: EmbeddingProvider;
  llm?: LlmProvider;
}): Promise<FastifyInstance> {
  const app = Fastify({ logger: false }).withTypeProvider<ZodTypeProvider>();
  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);
  app.decorate('db', deps.db);
  app.decorate('pool', deps.pool);
  app.decorate('embedding', deps.embedding);
  app.decorate('llm', deps.llm ?? makeNoopLlm());
  await app.register(authPlugin);
  await app.register(healthRoutes);
  await app.register(documentsRoutes);
  await app.register(reposRoutes);
  await app.register(syncRoutes);
  await app.register(searchRoutes);
  await app.register(askRoutes);
  return app;
}

function makeNoopLlm(): LlmProvider {
  return {
    defaultModel: 'noop',
    async synthesize() {
      throw new Error('llm not configured in test');
    },
    async ping() {
      return false;
    },
  };
}
