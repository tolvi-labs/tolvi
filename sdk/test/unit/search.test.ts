import { describe, it, expect, vi } from "vitest";
import { Http } from "../../src/http";
import { SearchResource } from "../../src/resources/search";
import { TolviEmbeddingUnavailableError } from "../../src/errors";

function jsonResponse(body: unknown, init: ResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function makeResource(fetch: typeof globalThis.fetch) {
  const http = new Http({ apiKey: "test", baseUrl: "https://api.example.com", fetch });
  return new SearchResource(http);
}

describe("search.query", () => {
  it("POSTs to /v1/search with the body", async () => {
    const responseBody = {
      results: [
        {
          document_id: "11111111-1111-1111-1111-111111111111",
          doc_type: "decision",
          slug: "postgres",
          title: "Postgres",
          score: 0.91,
          raw_similarity: 0.88,
          matched_chunk: { position: 0, content: "...", heading_path: ["Why"] },
        },
      ],
      total: 1,
    };
    const fetch = vi.fn().mockResolvedValue(jsonResponse(responseBody));
    const search = makeResource(fetch);

    const result = await search.query({ query: "why postgres", limit: 5 });

    expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/search");
    expect(fetch.mock.calls[0]![1].method).toBe("POST");
    expect(JSON.parse(fetch.mock.calls[0]![1].body as string)).toEqual({
      query: "why postgres",
      limit: 5,
    });
    expect(result).toEqual(responseBody);
  });

  it("surfaces 503 as TolviEmbeddingUnavailableError", async () => {
    const fetch = vi.fn().mockResolvedValue(
      jsonResponse({ error: { code: "embedding_unavailable", message: "ollama down" } }, { status: 503 }),
    );
    const search = makeResource(fetch);
    await expect(search.query({ query: "anything" })).rejects.toBeInstanceOf(TolviEmbeddingUnavailableError);
  });
});
