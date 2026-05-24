import { describe, it, expect, vi } from "vitest";
import { Http } from "../../src/http";
import { ReposResource } from "../../src/resources/repos";

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    headers: { "content-type": "application/json" },
  });
}

function makeResource(fetch: typeof globalThis.fetch) {
  const http = new Http({ apiKey: "test", baseUrl: "https://api.example.com", fetch });
  return new ReposResource(http);
}

describe("repos.list", () => {
  it("GETs /v1/repos and returns the typed response", async () => {
    const responseBody = {
      repos: [
        {
          id: "11111111-1111-1111-1111-111111111111",
          slug: "myrepo",
          remote_url: "https://github.com/x/y",
          document_count: 12,
          last_synced_at: "2026-05-22T00:00:00Z",
        },
      ],
    };
    const fetch = vi.fn().mockResolvedValue(jsonResponse(responseBody));
    const repos = makeResource(fetch);

    const result = await repos.list();

    expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/repos");
    expect(fetch.mock.calls[0]![1].method).toBe("GET");
    expect(result).toEqual(responseBody);
  });
});
