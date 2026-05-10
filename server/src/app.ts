import Fastify, { type FastifyInstance } from 'fastify';
import {
  serializerCompiler,
  validatorCompiler,
  type ZodTypeProvider,
} from 'fastify-type-provider-zod';
import { type Config } from './config.js';
import { buildLoggerOptions } from './observability/logging.js';

export async function buildApp(cfg: Config): Promise<FastifyInstance> {
  const app = Fastify({
    logger: buildLoggerOptions(cfg),
    genReqId: () => crypto.randomUUID(),
  }).withTypeProvider<ZodTypeProvider>();

  app.setValidatorCompiler(validatorCompiler);
  app.setSerializerCompiler(serializerCompiler);

  // Make config available to routes via app.config
  app.decorate('config', cfg);

  return app;
}

declare module 'fastify' {
  interface FastifyInstance {
    config: Config;
  }
}
