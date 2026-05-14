import { describe, it, expect } from 'vitest';
import { buildLlmProvider } from '../../src/llm/factory.js';
import { AnthropicLlmProvider } from '../../src/llm/anthropic.js';
import type { Config } from '../../src/config.js';

const cfg: Config = {
  databaseUrl: 'postgresql://test',
  embeddingProvider: 'ollama',
  embeddingBaseUrl: 'http://localhost:11434',
  embeddingModel: 'nomic-embed-text',
  embeddingDim: 768,
  openaiApiKey: undefined,
  llmProvider: 'anthropic',
  anthropicApiKey: 'sk-ant-test',
  anthropicModel: 'claude-sonnet-4-7',
  port: 3000,
  logLevel: 'info',
  nodeEnv: 'test',
};

describe('buildLlmProvider', () => {
  it('returns Anthropic provider when LLM_PROVIDER=anthropic', () => {
    expect(buildLlmProvider(cfg)).toBeInstanceOf(AnthropicLlmProvider);
  });

  it('throws when LLM_PROVIDER=anthropic but no key set', () => {
    expect(() => buildLlmProvider({ ...cfg, anthropicApiKey: undefined })).toThrow(/ANTHROPIC_API_KEY/);
  });

  it('passes the configured default model through', () => {
    const provider = buildLlmProvider({ ...cfg, anthropicModel: 'claude-haiku-4-7' });
    expect(provider.defaultModel).toBe('claude-haiku-4-7');
  });
});
