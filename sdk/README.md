# @tolvi-labs/sdk

TypeScript SDK for the Tolvi API.

## Install

```bash
npm install @tolvi-labs/sdk
```

ESM-only. Works in Node 18+, Bun, Deno, and modern browsers.

## Quickstart

```ts
import { Tolvi } from "@tolvi-labs/sdk";

const client = new Tolvi({
  apiKey: process.env.TOLVI_API_KEY!,
  baseUrl: "https://tolvi.example.com",
});

const { answer, citations } = await client.ask({ query: "Why did we choose Postgres?" });
console.log(answer);
for (const c of citations) console.log(c.slug, "→", c.document_id);
```

## Methods

### `client.documents`

```ts
const { document } = await client.documents.create({ repo, path, content });
const { documents, next_cursor } = await client.documents.list({ repo });
const { document } = await client.documents.get(id);
await client.documents.delete(id);
```

### `client.sync`

```ts
const { results, summary } = await client.sync.batch({
  repo: "my-repo",
  documents: [{ path: "decisions/foo.md", content: "..." }],
});
```

### `client.repos`

```ts
const { repos } = await client.repos.list();
```

### `client.search`

```ts
const { results, total } = await client.search.query({
  query: "why postgres?",
  limit: 5,
  filters: { doc_type: ["decision"] },
});
```

### `client.ask`

```ts
const { answer, citations, search_results, model, tokens } =
  await client.ask({ query: "why postgres?" });
```

## Error handling

Every method throws a typed `TolviError` subclass on failure. The hierarchy:

```
TolviError                              (abstract base)
├── TolviAPIError                       (any non-2xx)
│   ├── TolviValidationError            (400)
│   ├── TolviAuthError                  (401)
│   ├── TolviNotFoundError              (404)
│   ├── TolviEmbeddingUnavailableError  (503 from search/ask)
│   └── TolviUnknownAPIError            (any other status)
├── TolviConnectionError                (network failure before response)
└── TolviAbortError                     (signal aborted)
```

```ts
import { TolviNotFoundError, TolviEmbeddingUnavailableError, TolviAbortError } from "@tolvi-labs/sdk";

try {
  const { document } = await client.documents.get(id);
} catch (err) {
  if (err instanceof TolviNotFoundError) return null;
  if (err instanceof TolviEmbeddingUnavailableError) {
    // back off and retry — embedding backend is down
  }
  if (err instanceof TolviAbortError) return; // user cancelled
  throw err;
}
```

Every `TolviAPIError` exposes:

- `status: number` — HTTP status code
- `code: string` — server error code (e.g. `"validation_error"`)
- `body: ErrorEnvelope` — full server envelope for forward-compat access
- `requestId?: string` — from `x-request-id` response header when present

## Cancellation

The SDK has no default request timeout — set one explicitly via `AbortSignal.timeout()`:

```ts
const result = await client.ask(
  { query: "..." },
  { signal: AbortSignal.timeout(60_000) }
);
```

## Server compatibility

| SDK version | Tested against `tolvi-server` |
| ----------- | ------------------------------ |
| 0.1.x       | 0.1.x                          |

When the server adds new surface (endpoints, required fields), an older SDK will
continue to work but won't expose the new capability. Upgrade the SDK to match.

## Developing the SDK

This package lives inside the tolvi monorepo at `sdk/`. Types are generated from
`../spec/openapi.json`:

```bash
cd sdk
npm install
npm run gen:types     # regenerate src/types.gen.ts from spec/openapi.json
npm run typecheck
npm test
npm run build
```

If the SDK is ever extracted to a standalone repo, `gen:types` will need to source
the OpenAPI document differently — see `docs/OPEN_QUESTIONS.md` #10.

## Releasing

1. Bump `sdk/package.json` `version` and `sdk/src/version.ts` `VERSION` to match.
2. Move the `Unreleased` entries in `sdk/CHANGELOG.md` under a new dated heading.
3. Commit: `chore(sdk): release v<X.Y.Z>`.
4. Tag and push: `git tag sdk-v<X.Y.Z> && git push origin sdk-v<X.Y.Z>`.
5. `.github/workflows/sdk-release.yml` publishes to npm and creates the GitHub release.

No local `npm publish` runs — everything happens in CI.

## License

Apache-2.0. See `LICENSE`.
