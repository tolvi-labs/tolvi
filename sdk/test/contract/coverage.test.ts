import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { Tolvi } from "../../src/client";

const __dirname = dirname(fileURLToPath(import.meta.url));
const openapi = JSON.parse(
  readFileSync(join(__dirname, "../../../spec/openapi.json"), "utf8"),
) as { paths: Record<string, Record<string, unknown>> };

const map: Array<{ op: string; sdk: (c: Tolvi) => boolean }> = [
  { op: "POST /v1/documents",        sdk: (c) => typeof c.documents.create === "function" },
  { op: "GET /v1/documents",         sdk: (c) => typeof c.documents.list   === "function" },
  { op: "GET /v1/documents/{id}",    sdk: (c) => typeof c.documents.get    === "function" },
  { op: "DELETE /v1/documents/{id}", sdk: (c) => typeof c.documents.delete === "function" },
  { op: "POST /v1/sync",             sdk: (c) => typeof c.sync.batch       === "function" },
  { op: "GET /v1/repos",             sdk: (c) => typeof c.repos.list       === "function" },
  { op: "POST /v1/search",           sdk: (c) => typeof c.search.query     === "function" },
  { op: "POST /v1/ask",              sdk: (c) => typeof c.ask              === "function" },
];

const intentionallyUnexposed = new Set(["GET /healthz", "GET /readyz"]);

describe("contract", () => {
  const documented = Object.entries(openapi.paths).flatMap(([path, methods]) =>
    Object.keys(methods).map((m) => `${m.toUpperCase()} ${path}`),
  );

  it("every documented op is either mapped or intentionally unexposed", () => {
    const unmapped = documented.filter(
      (op) => !intentionallyUnexposed.has(op) && !map.find((m) => m.op === op),
    );
    expect(unmapped, `unmapped ops: ${unmapped.join(", ")}`).toEqual([]);
  });

  it("every mapped op has an SDK method", () => {
    const client = new Tolvi({ apiKey: "x", baseUrl: "https://x" });
    for (const m of map) {
      expect(m.sdk(client), `SDK missing method for ${m.op}`).toBe(true);
    }
  });

  it("intentionallyUnexposed entries actually exist in the spec", () => {
    for (const op of intentionallyUnexposed) {
      expect(documented, `intentionallyUnexposed contains ${op} but spec does not document it`).toContain(op);
    }
  });
});
