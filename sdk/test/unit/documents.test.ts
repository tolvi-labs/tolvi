import { describe, it, expect, vi } from "vitest";
import { Http } from "../../src/http";
import { DocumentsResource } from "../../src/resources/documents";

function jsonResponse(body: unknown, init: ResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function makeResource(fetch: typeof globalThis.fetch) {
  const http = new Http({ apiKey: "test", baseUrl: "https://api.example.com", fetch });
  return new DocumentsResource(http);
}

const sampleDoc = {
  id: "11111111-1111-1111-1111-111111111111",
  repo_id: "22222222-2222-2222-2222-222222222222",
  doc_type: "decision",
  slug: "foo",
  status: "active",
  title: "Foo",
  content_hash: "abc123",
  chunks: 3,
  embedded_at: "2026-05-22T00:00:00Z",
};

describe("documents.create", () => {
  it("POSTs to /v1/documents with the body", async () => {
    const fetch = vi.fn().mockResolvedValue(jsonResponse({ document: sampleDoc }));
    const documents = makeResource(fetch);
    const result = await documents.create({ repo: "r", path: "p.md", content: "c" });

    expect(fetch).toHaveBeenCalledOnce();
    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe("https://api.example.com/v1/documents");
    expect(init.method).toBe("POST");
    expect(init.body).toBe(JSON.stringify({ repo: "r", path: "p.md", content: "c" }));
    expect(result).toEqual({ document: sampleDoc });
  });

  it("forwards the AbortSignal", async () => {
    const fetch = vi.fn().mockResolvedValue(jsonResponse({ document: sampleDoc }));
    const documents = makeResource(fetch);
    const ctrl = new AbortController();
    await documents.create({ repo: "r", path: "p.md", content: "c" }, { signal: ctrl.signal });
    expect(fetch.mock.calls[0]![1].signal).toBe(ctrl.signal);
  });
});

describe("documents.list", () => {
  it("GETs /v1/documents with no query when omitted", async () => {
    const fetch = vi.fn().mockResolvedValue(jsonResponse({ documents: [], next_cursor: null }));
    const documents = makeResource(fetch);
    const result = await documents.list();
    expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/documents");
    expect(result).toEqual({ documents: [], next_cursor: null });
  });

  it("appends query parameters when provided", async () => {
    const fetch = vi.fn().mockResolvedValue(jsonResponse({ documents: [], next_cursor: null }));
    const documents = makeResource(fetch);
    await documents.list({ repo: "x", doc_type: "decision", limit: 25 });
    expect(fetch.mock.calls[0]![0]).toBe(
      "https://api.example.com/v1/documents?repo=x&doc_type=decision&limit=25",
    );
  });
});

describe("documents.get", () => {
  it("GETs /v1/documents/:id", async () => {
    const fetch = vi.fn().mockResolvedValue(jsonResponse({ document: sampleDoc }));
    const documents = makeResource(fetch);
    const result = await documents.get(sampleDoc.id);
    expect(fetch.mock.calls[0]![0]).toBe(`https://api.example.com/v1/documents/${sampleDoc.id}`);
    expect(fetch.mock.calls[0]![1].method).toBe("GET");
    expect(result).toEqual({ document: sampleDoc });
  });
});

describe("documents.delete", () => {
  it("DELETEs /v1/documents/:id and returns void", async () => {
    const fetch = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    const documents = makeResource(fetch);
    const result = await documents.delete(sampleDoc.id);
    expect(fetch.mock.calls[0]![1].method).toBe("DELETE");
    expect(result).toBeUndefined();
  });
});
