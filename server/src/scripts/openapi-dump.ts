/**
 * Generates spec/openapi.json from the registered Fastify routes.
 *
 * Known v1 limitation: response schemas are not yet declared on data-plane
 * routes (documents, sync, repos, search, ask) — only `body` schemas are
 * defined. As a result the generated spec describes request shapes but
 * not response shapes. Adding response zod schemas to each route is a
 * follow-up before the spec is consumed by SDK generation.
 */
import { writeFileSync } from 'node:fs';
import path from 'node:path';
import { buildApp } from '../app.js';
import type { Config } from '../config.js';

// Boot the app with a stub config — no real DB/embedding/LLM needed since we
// only extract the swagger document. `pg.Pool` is lazy (no connection on
// construction), and route handlers never run during boot, so the stub
// suffices for emitting the OpenAPI document.
const stubConfig: Config = {
  databaseUrl: 'postgresql://stub',
  embeddingProvider: 'ollama',
  embeddingBaseUrl: 'http://localhost:11434',
  embeddingModel: 'nomic-embed-text',
  embeddingDim: 768,
  openaiApiKey: undefined,
  llmProvider: 'anthropic',
  anthropicApiKey: 'sk-stub',
  anthropicModel: 'claude-sonnet-4-7',
  port: 3000,
  logLevel: 'fatal',
  nodeEnv: 'production',
};

async function dump(): Promise<void> {
  const app = await buildApp(stubConfig);
  await app.ready();
  try {
    const doc = app.swagger();
    const outPath = path.resolve(process.cwd(), '../spec/openapi.json');
    writeFileSync(outPath, JSON.stringify(doc, null, 2) + '\n', 'utf8');
    console.log(`Wrote ${outPath}`);
  } finally {
    await app.close();
  }
}

dump().catch((err) => {
  console.error('openapi-dump failed:', err);
  process.exit(1);
});
