import type { Config } from '../config.js';
import type { LlmProvider } from './provider.js';
import { AnthropicLlmProvider } from './anthropic.js';

export function buildLlmProvider(cfg: Config): LlmProvider {
  switch (cfg.llmProvider) {
    case 'anthropic':
      if (!cfg.anthropicApiKey) {
        throw new Error('ANTHROPIC_API_KEY required for anthropic LLM provider');
      }
      return new AnthropicLlmProvider({
        apiKey: cfg.anthropicApiKey,
        defaultModel: cfg.anthropicModel,
      });
  }
}
