import { describe, it, expect } from 'vitest';
import { buildEmbeddingProvider } from '../../src/embedding/factory.js';
import { OllamaEmbeddingProvider } from '../../src/embedding/ollama.js';
import { OpenAIEmbeddingProvider } from '../../src/embedding/openai.js';
import type { Config } from '../../src/config.js';

const baseConfig: Config = {
  databaseUrl: 'postgresql://test',
  embeddingProvider: 'ollama',
  embeddingBaseUrl: 'http://localhost:11434',
  embeddingModel: 'nomic-embed-text',
  embeddingDim: 768,
  openaiApiKey: undefined,
  llmProvider: 'anthropic',
  anthropicApiKey: 'test',
  anthropicModel: 'claude-sonnet-4-7',
  port: 3000,
  logLevel: 'info',
  nodeEnv: 'test',
};

describe('buildEmbeddingProvider', () => {
  it('returns Ollama provider when EMBEDDING_PROVIDER=ollama', () => {
    expect(buildEmbeddingProvider(baseConfig)).toBeInstanceOf(OllamaEmbeddingProvider);
  });

  it('returns OpenAI provider when EMBEDDING_PROVIDER=openai with key', () => {
    const cfg = { ...baseConfig, embeddingProvider: 'openai' as const, openaiApiKey: 'sk-test' };
    expect(buildEmbeddingProvider(cfg)).toBeInstanceOf(OpenAIEmbeddingProvider);
  });

  it('throws when EMBEDDING_PROVIDER=openai but no key set', () => {
    const cfg = { ...baseConfig, embeddingProvider: 'openai' as const, openaiApiKey: undefined };
    expect(() => buildEmbeddingProvider(cfg)).toThrow(/OPENAI_API_KEY/);
  });

  it('passes the configured dimension through to the provider', () => {
    const provider = buildEmbeddingProvider({ ...baseConfig, embeddingDim: 1024 });
    expect(provider.dimension).toBe(1024);
  });
});
