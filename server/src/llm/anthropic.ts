import Anthropic from '@anthropic-ai/sdk';
import type { LlmProvider, LlmRequest, LlmResponse } from './provider.js';

export type AnthropicConfig = {
  apiKey: string;
  defaultModel: string;
};

export class AnthropicLlmProvider implements LlmProvider {
  readonly defaultModel: string;
  private readonly client: Anthropic;

  constructor(cfg: AnthropicConfig) {
    this.defaultModel = cfg.defaultModel;
    this.client = new Anthropic({ apiKey: cfg.apiKey });
  }

  async synthesize(req: LlmRequest): Promise<LlmResponse> {
    const model = req.model ?? this.defaultModel;
    const response = await this.client.messages.create({
      model,
      max_tokens: req.maxTokens ?? 1024,
      system: [
        {
          type: 'text',
          text: req.systemPrompt,
          cache_control: { type: 'ephemeral' },
        },
      ],
      messages: req.messages.map((m) => ({ role: m.role, content: m.content })),
    });

    const text = response.content
      .filter((b): b is Anthropic.TextBlock => b.type === 'text')
      .map((b) => b.text)
      .join('');

    if (text === '') {
      throw new Error(`Anthropic returned no text content (stop_reason: ${response.stop_reason})`);
    }

    return {
      text,
      model: response.model,
      tokens: {
        input: response.usage.input_tokens,
        output: response.usage.output_tokens,
        cacheRead: response.usage.cache_read_input_tokens ?? 0,
      },
    };
  }

  async ping(): Promise<boolean> {
    try {
      await this.client.messages.create(
        {
          model: this.defaultModel,
          max_tokens: 1,
          messages: [{ role: 'user', content: 'ping' }],
        },
        { timeout: 3000 },
      );
      return true;
    } catch {
      return false;
    }
  }
}
