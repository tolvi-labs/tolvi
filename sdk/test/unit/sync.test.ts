import { describe, it, expect, vi } from "vitest";
import { Http } from "../../src/http";
import { SyncResource } from "../../src/resources/sync";

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    headers: { "content-type": "application/json" },
  });
}

function makeResource(fetch: typeof globalThis.fetch) {
  const http = new Http({ apiKey: "test", baseUrl: "https://api.example.com", fetch });
  return new SyncResource(http);
}

describe("sync.batch", () => {
  it("POSTs to /v1/sync with the body and returns the typed response", async () => {
    const responseBody = {
      results: [
        { path: "a.md", status: "created", document_id: "id1", error: null },
        { path: "b.md", status: "unchanged", document_id: "id2", error: null },
      ],
      summary: { created: 1, updated: 0, unchanged: 1, failed: 0 },
    };
    const fetch = vi.fn().mockResolvedValue(jsonResponse(responseBody));
    const sync = makeResource(fetch);

    const result = await sync.batch({
      repo: "myrepo",
      documents: [
        { path: "a.md", content: "alpha" },
        { path: "b.md", content: "beta" },
      ],
    });

    expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/sync");
    expect(fetch.mock.calls[0]![1].method).toBe("POST");
    expect(JSON.parse(fetch.mock.calls[0]![1].body as string)).toEqual({
      repo: "myrepo",
      documents: [
        { path: "a.md", content: "alpha" },
        { path: "b.md", content: "beta" },
      ],
    });
    expect(result).toEqual(responseBody);
  });
});
