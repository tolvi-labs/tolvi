import { describe, it, expect, vi } from "vitest";
import { Http } from "../../src/http";
import {
  TolviValidationError,
  TolviAuthError,
  TolviNotFoundError,
  TolviEmbeddingUnavailableError,
  TolviUnknownAPIError,
  TolviConnectionError,
  TolviAbortError,
} from "../../src/errors";
import { VERSION } from "../../src/version";

function jsonResponse(body: unknown, init: ResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function makeHttp(fetch: typeof globalThis.fetch, overrides: Partial<{ userAgent: string }> = {}) {
  return new Http({
    apiKey: "test-key",
    baseUrl: "https://api.example.com",
    fetch,
    userAgent: overrides.userAgent,
  });
}

describe("Http", () => {
  describe("auth + headers", () => {
    it("sends Authorization: Bearer <apiKey>", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
      const http = makeHttp(fetch);
      await http.request("GET", "/v1/repos");

      expect(fetch).toHaveBeenCalledOnce();
      const [, init] = fetch.mock.calls[0]!;
      expect(init.headers).toMatchObject({
        authorization: "Bearer test-key",
        "user-agent": `tolvi-sdk/${VERSION}`,
      });
    });

    it("uses a custom userAgent when provided", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({}));
      const http = makeHttp(fetch, { userAgent: "my-app/1.0" });
      await http.request("GET", "/v1/repos");
      const [, init] = fetch.mock.calls[0]!;
      expect(init.headers).toMatchObject({ "user-agent": "my-app/1.0" });
    });

    it("constructs URLs by joining baseUrl + path", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({}));
      const http = makeHttp(fetch);
      await http.request("GET", "/v1/repos");
      expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/repos");
    });

    it("appends query parameters when provided", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({}));
      const http = makeHttp(fetch);
      await http.request("GET", "/v1/documents", { query: { repo: "x", limit: 10 } });
      expect(fetch.mock.calls[0]![0]).toBe("https://api.example.com/v1/documents?repo=x&limit=10");
    });

    it("omits the request body and content-type for GET", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({}));
      const http = makeHttp(fetch);
      await http.request("GET", "/v1/repos");
      const [, init] = fetch.mock.calls[0]!;
      expect(init.body).toBeUndefined();
      expect((init.headers as Record<string, string>)["content-type"]).toBeUndefined();
    });

    it("serializes JSON body and sets content-type for POST", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({}));
      const http = makeHttp(fetch);
      await http.request("POST", "/v1/documents", { body: { repo: "x", path: "y.md", content: "z" } });
      const [, init] = fetch.mock.calls[0]!;
      expect(init.body).toBe(JSON.stringify({ repo: "x", path: "y.md", content: "z" }));
      expect(init.headers).toMatchObject({ "content-type": "application/json" });
    });
  });

  describe("success responses", () => {
    it("parses 200 JSON response", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({ documents: [] }));
      const http = makeHttp(fetch);
      const result = await http.request<{ documents: unknown[] }>("GET", "/v1/documents");
      expect(result).toEqual({ documents: [] });
    });

    it("returns undefined on 204", async () => {
      const fetch = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
      const http = makeHttp(fetch);
      const result = await http.request<void>("DELETE", "/v1/documents/abc");
      expect(result).toBeUndefined();
    });
  });

  describe("error mapping", () => {
    it("throws TolviValidationError on 400", async () => {
      const fetch = vi.fn().mockResolvedValue(
        jsonResponse({ error: { code: "validation_error", message: "bad" } }, { status: 400 }),
      );
      const http = makeHttp(fetch);
      await expect(http.request("POST", "/v1/documents")).rejects.toBeInstanceOf(TolviValidationError);
    });

    it("throws TolviAuthError on 401", async () => {
      const fetch = vi.fn().mockResolvedValue(
        jsonResponse({ error: { code: "unauthorized", message: "no key" } }, { status: 401 }),
      );
      const http = makeHttp(fetch);
      await expect(http.request("GET", "/v1/repos")).rejects.toBeInstanceOf(TolviAuthError);
    });

    it("throws TolviNotFoundError on 404", async () => {
      const fetch = vi.fn().mockResolvedValue(
        jsonResponse({ error: { code: "not_found", message: "no" } }, { status: 404 }),
      );
      const http = makeHttp(fetch);
      await expect(http.request("GET", "/v1/documents/abc")).rejects.toBeInstanceOf(TolviNotFoundError);
    });

    it("throws TolviEmbeddingUnavailableError on 503", async () => {
      const fetch = vi.fn().mockResolvedValue(
        jsonResponse({ error: { code: "embedding_unavailable", message: "ollama down" } }, { status: 503 }),
      );
      const http = makeHttp(fetch);
      await expect(http.request("POST", "/v1/search")).rejects.toBeInstanceOf(TolviEmbeddingUnavailableError);
    });

    it("throws TolviUnknownAPIError on 500", async () => {
      const fetch = vi.fn().mockResolvedValue(
        jsonResponse({ error: { code: "internal", message: "boom" } }, { status: 500 }),
      );
      const http = makeHttp(fetch);
      const err = await http.request("GET", "/v1/repos").catch((e: unknown) => e);
      expect(err).toBeInstanceOf(TolviUnknownAPIError);
      expect((err as TolviUnknownAPIError).status).toBe(500);
    });

    it("synthesizes an envelope when the response is not JSON", async () => {
      const fetch = vi.fn().mockResolvedValue(
        new Response("<html>bad gateway</html>", { status: 502, statusText: "Bad Gateway" }),
      );
      const http = makeHttp(fetch);
      const err = await http.request("GET", "/v1/repos").catch((e: unknown) => e);
      expect(err).toBeInstanceOf(TolviUnknownAPIError);
      expect((err as TolviUnknownAPIError).code).toBe("unknown");
      expect((err as TolviUnknownAPIError).body.error.message).toBe("Bad Gateway");
    });

    it("extracts x-request-id from response headers", async () => {
      const fetch = vi.fn().mockResolvedValue(
        jsonResponse({ error: { code: "not_found", message: "no" } }, {
          status: 404,
          headers: { "x-request-id": "req-abc" },
        }),
      );
      const http = makeHttp(fetch);
      const err = await http.request("GET", "/v1/documents/x").catch((e: unknown) => e);
      expect((err as TolviNotFoundError).requestId).toBe("req-abc");
    });
  });

  describe("pre-response failures", () => {
    it("wraps fetch rejections as TolviConnectionError", async () => {
      const fetch = vi.fn().mockRejectedValue(new Error("ECONNREFUSED"));
      const http = makeHttp(fetch);
      const err = await http.request("GET", "/v1/repos").catch((e: unknown) => e);
      expect(err).toBeInstanceOf(TolviConnectionError);
      expect((err as TolviConnectionError).cause.message).toBe("ECONNREFUSED");
    });

    it("wraps AbortError as TolviAbortError", async () => {
      const abortErr = new Error("The operation was aborted");
      abortErr.name = "AbortError";
      const fetch = vi.fn().mockRejectedValue(abortErr);
      const http = makeHttp(fetch);
      const err = await http.request("GET", "/v1/repos").catch((e: unknown) => e);
      expect(err).toBeInstanceOf(TolviAbortError);
    });

    it("forwards AbortSignal to fetch", async () => {
      const fetch = vi.fn().mockResolvedValue(jsonResponse({}));
      const http = makeHttp(fetch);
      const ctrl = new AbortController();
      await http.request("GET", "/v1/repos", { signal: ctrl.signal });
      const [, init] = fetch.mock.calls[0]!;
      expect(init.signal).toBe(ctrl.signal);
    });
  });
});
