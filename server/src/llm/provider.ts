export type LlmMessage = {
  role: 'user';
  content: string;
};

export type LlmRequest = {
  systemPrompt: string;
  messages: LlmMessage[];
  model?: string;
  maxTokens?: number;
};

export type LlmResponse = {
  text: string;
  model: string;
  tokens: { input: number; output: number; cacheRead: number };
};

export type LlmProvider = {
  synthesize(req: LlmRequest): Promise<LlmResponse>;
  ping(): Promise<boolean>;
  readonly defaultModel: string;
};
