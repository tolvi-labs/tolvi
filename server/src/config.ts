import { z } from 'zod';

const ConfigSchema = z.object({
  // Database
  DATABASE_URL: z.string().min(1, 'DATABASE_URL is required'),

  // Embedding provider
  EMBEDDING_PROVIDER: z.enum(['ollama', 'openai']),
  EMBEDDING_BASE_URL: z.string().default('http://localhost:11434'),
  EMBEDDING_MODEL: z.string().default('nomic-embed-text'),
  EMBEDDING_DIM: z.coerce.number().int().positive().default(768),
  OPENAI_API_KEY: z.string().optional(),

  // LLM provider
  LLM_PROVIDER: z.enum(['anthropic']).default('anthropic'),
  ANTHROPIC_API_KEY: z.string().optional(),
  ANTHROPIC_MODEL: z.string().default('claude-sonnet-4-7'),

  // Server
  PORT: z.coerce.number().int().positive().default(3000),
  LOG_LEVEL: z.enum(['fatal', 'error', 'warn', 'info', 'debug', 'trace']).default('info'),
  NODE_ENV: z.enum(['development', 'test', 'production']).default('development'),
});

export type Config = {
  databaseUrl: string;
  embeddingProvider: 'ollama' | 'openai';
  embeddingBaseUrl: string;
  embeddingModel: string;
  embeddingDim: number;
  openaiApiKey: string | undefined;
  llmProvider: 'anthropic';
  anthropicApiKey: string | undefined;
  anthropicModel: string;
  port: number;
  logLevel: 'fatal' | 'error' | 'warn' | 'info' | 'debug' | 'trace';
  nodeEnv: 'development' | 'test' | 'production';
};

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const parsed = ConfigSchema.safeParse(env);
  if (!parsed.success) {
    const issues = parsed.error.issues
      .map((i) => `  ${i.path.join('.')}: ${i.message}`)
      .join('\n');
    throw new Error(`Invalid configuration:\n${issues}`);
  }

  const cfg = parsed.data;

  // Cross-field validation
  if (cfg.LLM_PROVIDER === 'anthropic' && !cfg.ANTHROPIC_API_KEY) {
    throw new Error('ANTHROPIC_API_KEY is required when LLM_PROVIDER=anthropic');
  }
  if (cfg.EMBEDDING_PROVIDER === 'openai' && !cfg.OPENAI_API_KEY) {
    throw new Error('OPENAI_API_KEY is required when EMBEDDING_PROVIDER=openai');
  }

  return {
    databaseUrl: cfg.DATABASE_URL,
    embeddingProvider: cfg.EMBEDDING_PROVIDER,
    embeddingBaseUrl: cfg.EMBEDDING_BASE_URL,
    embeddingModel: cfg.EMBEDDING_MODEL,
    embeddingDim: cfg.EMBEDDING_DIM,
    openaiApiKey: cfg.OPENAI_API_KEY,
    llmProvider: cfg.LLM_PROVIDER,
    anthropicApiKey: cfg.ANTHROPIC_API_KEY,
    anthropicModel: cfg.ANTHROPIC_MODEL,
    port: cfg.PORT,
    logLevel: cfg.LOG_LEVEL,
    nodeEnv: cfg.NODE_ENV,
  };
}
