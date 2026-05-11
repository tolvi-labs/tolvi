import type { EmbeddingProvider } from './provider.js';

export type OllamaConfig = {
  baseUrl: string;
  model: string;
  dimension: number;
};

export class OllamaEmbeddingProvider implements EmbeddingProvider {
  readonly dimension: number;
  constructor(private readonly cfg: OllamaConfig) {
    this.dimension = cfg.dimension;
  }

  async embed(texts: string[]): Promise<number[][]> {
    if (texts.length === 0) return [];
    const results: number[][] = [];
    // Ollama's /api/embeddings takes one input at a time in the public API.
    // /api/embed (newer) supports batches; we use /api/embed with fallback.
    for (const text of texts) {
      const res = await fetch(`${this.cfg.baseUrl}/api/embeddings`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ model: this.cfg.model, prompt: text }),
      });
      if (!res.ok) {
        throw new Error(`Ollama embedding failed: ${res.status} ${await res.text()}`);
      }
      const json = (await res.json()) as { embedding: number[] };
      if (!Array.isArray(json.embedding)) {
        throw new Error(`Ollama returned no embedding for text length ${text.length}`);
      }
      if (json.embedding.length !== this.dimension) {
        throw new Error(
          `Ollama returned ${json.embedding.length}-dim vector but config expects ${this.dimension}`
        );
      }
      results.push(json.embedding);
    }
    return results;
  }

  async ping(): Promise<boolean> {
    try {
      const res = await fetch(`${this.cfg.baseUrl}/api/tags`, { signal: AbortSignal.timeout(2000) });
      if (!res.ok) return false;
      const json = (await res.json()) as { models?: Array<{ name: string }> };
      return (json.models ?? []).some((m) => m.name.startsWith(this.cfg.model));
    } catch {
      return false;
    }
  }
}
