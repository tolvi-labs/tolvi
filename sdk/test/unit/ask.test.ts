import { describe, it, expect, vi } from "vitest";
import { Http } from "../../src/http";
import { AskResource } from "../../src/resources/ask";
import { TolviEmbeddingUnavailableError } from "../../src/errors";

function jsonResponse(body: unknown, init: ResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function makeResource(fetch: typeof globalThis.fetch) {
  const http = new Http({ apiKey: "test", baseUrl: "https://api.example.com", fetch });
  return new AskResource(http);
}

describe("ask", () => {
  it("POSTs to /v1/ask with the body", async () => {
    const responseBody = {
      answer: "Because of session pooling.",
      citations: [
        { slug: "postgres", doc_type: "decision", document_id: "11111111-1111-1111-1111-111111111111" },
      ],
      search_results: [],
      model: "claude-sonnet-4-6",
      tokens: { input: 100, output: 50, cache_read: 0 },
    };
    const fetch = vi.fn().mockResolvedValue(jsonResponse(responseBody));
    const ask = makeResource(fetch);

    const result = await ask.ask({ query: "why postgres" });

    expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/ask");
    expect(fetch.mock.calls[0]![1].method).toBe("POST");
    expect(JSON.parse(fetch.mock.calls[0]![1].body as string)).toEqual({ query: "why postgres" });
    expect(result).toEqual(responseBody);
  });

  it("surfaces 503 as TolviEmbeddingUnavailableError", async () => {
    const fetch = vi.fn().mockResolvedValue(
      jsonResponse({ error: { code: "embedding_unavailable", message: "ollama down" } }, { status: 503 }),
    );
    const ask = makeResource(fetch);
    await expect(ask.ask({ query: "anything" })).rejects.toBeInstanceOf(TolviEmbeddingUnavailableError);
  });
});
