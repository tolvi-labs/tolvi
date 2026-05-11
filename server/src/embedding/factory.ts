import type { Config } from '../config.js';
import type { EmbeddingProvider } from './provider.js';
import { OllamaEmbeddingProvider } from './ollama.js';
import { OpenAIEmbeddingProvider } from './openai.js';

export function buildEmbeddingProvider(cfg: Config): EmbeddingProvider {
  switch (cfg.embeddingProvider) {
    case 'ollama':
      return new OllamaEmbeddingProvider({
        baseUrl: cfg.embeddingBaseUrl,
        model: cfg.embeddingModel,
        dimension: cfg.embeddingDim,
      });
    case 'openai':
      if (!cfg.openaiApiKey) {
        throw new Error('OPENAI_API_KEY required for openai embedding provider');
      }
      return new OpenAIEmbeddingProvider({
        apiKey: cfg.openaiApiKey,
        model: cfg.embeddingModel,
        dimension: cfg.embeddingDim,
      });
  }
}
