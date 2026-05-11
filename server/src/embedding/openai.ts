import type { EmbeddingProvider } from './provider.js';

export type OpenAIConfig = {
  apiKey: string;
  model: string;
  dimension: number;
};

export class OpenAIEmbeddingProvider implements EmbeddingProvider {
  readonly dimension: number;
  constructor(private readonly cfg: OpenAIConfig) {
    this.dimension = cfg.dimension;
  }

  async embed(texts: string[]): Promise<number[][]> {
    if (texts.length === 0) return [];
    const res = await fetch('https://api.openai.com/v1/embeddings', {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        authorization: `Bearer ${this.cfg.apiKey}`,
      },
      body: JSON.stringify({ model: this.cfg.model, input: texts }),
    });
    if (!res.ok) {
      throw new Error(`OpenAI embedding failed: ${res.status} ${await res.text()}`);
    }
    const json = (await res.json()) as { data: Array<{ embedding: number[] }> };
    return json.data.map((d) => d.embedding);
  }

  async ping(): Promise<boolean> {
    try {
      const res = await fetch('https://api.openai.com/v1/models', {
        headers: { authorization: `Bearer ${this.cfg.apiKey}` },
        signal: AbortSignal.timeout(3000),
      });
      return res.ok;
    } catch {
      return false;
    }
  }
}
