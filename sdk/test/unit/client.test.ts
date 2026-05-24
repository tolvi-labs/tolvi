import { describe, it, expect, vi } from "vitest";
import { Tolvi } from "../../src/client";
import { DocumentsResource } from "../../src/resources/documents";
import { SyncResource } from "../../src/resources/sync";
import { ReposResource } from "../../src/resources/repos";
import { SearchResource } from "../../src/resources/search";

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    headers: { "content-type": "application/json" },
  });
}

describe("Tolvi", () => {
  it("constructs all four resource instances eagerly", () => {
    const client = new Tolvi({ apiKey: "k", baseUrl: "https://api" });
    expect(client.documents).toBeInstanceOf(DocumentsResource);
    expect(client.sync).toBeInstanceOf(SyncResource);
    expect(client.repos).toBeInstanceOf(ReposResource);
    expect(client.search).toBeInstanceOf(SearchResource);
  });

  it("exposes ask as a top-level method", async () => {
    const fetch = vi.fn().mockResolvedValue(
      jsonResponse({
        answer: "a",
        citations: [],
        search_results: [],
        model: "m",
        tokens: { input: 0, output: 0, cache_read: 0 },
      }),
    );
    const client = new Tolvi({ apiKey: "k", baseUrl: "https://api.example.com", fetch });
    const result = await client.ask({ query: "anything" });
    expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/ask");
    expect(result.answer).toBe("a");
  });

  it("shares a single Http instance across all resources", async () => {
    const fetch = vi.fn().mockResolvedValue(jsonResponse({ repos: [] }));
    const client = new Tolvi({
      apiKey: "shared-key",
      baseUrl: "https://api.example.com",
      fetch,
      userAgent: "shared-ua/1.0",
    });

    await client.repos.list();
    const [, init] = fetch.mock.calls[0]!;
    expect(init.headers).toMatchObject({
      authorization: "Bearer shared-key",
      "user-agent": "shared-ua/1.0",
    });
  });
});
