import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { loadConfig } from '../../src/config.js';

describe('loadConfig', () => {
  let originalEnv: NodeJS.ProcessEnv;

  beforeEach(() => {
    originalEnv = { ...process.env };
    // Clear env so tests start clean
    for (const key of Object.keys(process.env)) {
      if (key.startsWith('DATABASE_') || key.startsWith('EMBEDDING_') ||
          key.startsWith('LLM_') || key.startsWith('ANTHROPIC_') ||
          ['PORT', 'LOG_LEVEL', 'NODE_ENV'].includes(key)) {
        delete process.env[key];
      }
    }
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('parses a valid config with all defaults', () => {
    process.env.DATABASE_URL = 'postgresql://test';
    process.env.EMBEDDING_PROVIDER = 'ollama';
    process.env.LLM_PROVIDER = 'anthropic';
    process.env.ANTHROPIC_API_KEY = 'test-key';

    const cfg = loadConfig();

    expect(cfg.databaseUrl).toBe('postgresql://test');
    expect(cfg.embeddingProvider).toBe('ollama');
    expect(cfg.embeddingBaseUrl).toBe('http://localhost:11434');
    expect(cfg.embeddingModel).toBe('nomic-embed-text');
    expect(cfg.embeddingDim).toBe(768);
    expect(cfg.llmProvider).toBe('anthropic');
    expect(cfg.port).toBe(3000);
    expect(cfg.logLevel).toBe('info');
    expect(cfg.nodeEnv).toBe('development');
  });

  it('throws a clear error when required env var is missing', () => {
    process.env.EMBEDDING_PROVIDER = 'ollama';
    // DATABASE_URL deliberately missing

    expect(() => loadConfig()).toThrow(/DATABASE_URL/);
  });

  it('throws when LLM_PROVIDER is anthropic but ANTHROPIC_API_KEY is missing', () => {
    process.env.DATABASE_URL = 'postgresql://test';
    process.env.EMBEDDING_PROVIDER = 'ollama';
    process.env.LLM_PROVIDER = 'anthropic';
    // ANTHROPIC_API_KEY missing

    expect(() => loadConfig()).toThrow(/ANTHROPIC_API_KEY/);
  });

  it('throws when EMBEDDING_PROVIDER is openai but OPENAI_API_KEY is missing', () => {
    process.env.DATABASE_URL = 'postgresql://test';
    process.env.EMBEDDING_PROVIDER = 'openai';
    process.env.LLM_PROVIDER = 'anthropic';
    process.env.ANTHROPIC_API_KEY = 'test-key';
    // OPENAI_API_KEY missing

    expect(() => loadConfig()).toThrow(/OPENAI_API_KEY/);
  });

  it('respects EMBEDDING_DIM override for non-default providers', () => {
    process.env.DATABASE_URL = 'postgresql://test';
    process.env.EMBEDDING_PROVIDER = 'openai';
    process.env.EMBEDDING_DIM = '1536';
    process.env.OPENAI_API_KEY = 'sk-test';
    process.env.LLM_PROVIDER = 'anthropic';
    process.env.ANTHROPIC_API_KEY = 'test-key';

    const cfg = loadConfig();
    expect(cfg.embeddingDim).toBe(1536);
    expect(cfg.embeddingProvider).toBe('openai');
  });
});
