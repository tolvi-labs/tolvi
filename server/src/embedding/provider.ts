export type EmbeddingProvider = {
  /**
   * Embed a batch of texts. Returns one vector per input, in input order.
   * Throws if the provider is unreachable or the model is not loaded.
   */
  embed(texts: string[]): Promise<number[][]>;

  /**
   * Health check — returns true if the provider is reachable AND the model is available.
   * Used by /readyz.
   */
  ping(): Promise<boolean>;

  /**
   * Dimension of vectors this provider produces.
   */
  readonly dimension: number;
};
