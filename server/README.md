# Tolvi Server

Multi-tenant Fastify + Postgres + pgvector implementation of the [`tolvi-format-v1`](../spec/tolvi-format-v1.md) public format contract.

> **Status:** Phase 2 implementation. CRUD + sync land in PR A; search + ask in PR B.

## Quickstart (self-host)

From the repo root:

```bash
cp .env.example .env
# edit .env to set POSTGRES_PASSWORD and ANTHROPIC_API_KEY
docker compose up -d
curl http://localhost:3000/healthz   # { "ok": true, "version": "0.1.0" }
```

The first run takes a few minutes (Ollama pulls the embedding model). Subsequent starts are fast.

## Local development

The server itself runs from your IDE; Postgres + Ollama run in containers.

```bash
cd server
docker compose up -d              # postgres + ollama only
npm install
DATABASE_URL=postgresql://tolvi:tolvi@localhost:5432/tolvi \
  EMBEDDING_PROVIDER=ollama \
  LLM_PROVIDER=anthropic \
  ANTHROPIC_API_KEY=sk-ant-... \
  npm run migrate
DATABASE_URL=... ANTHROPIC_API_KEY=... npm run dev
```

## Configuration

All config is via environment variables (see `src/config.ts` for the zod schema).

| Variable | Required | Default | Notes |
|---|---|---|---|
| `DATABASE_URL` | yes | — | Postgres connection string |
| `EMBEDDING_PROVIDER` | yes | — | `ollama` or `openai` |
| `EMBEDDING_BASE_URL` | no | `http://localhost:11434` | Ollama only |
| `EMBEDDING_MODEL` | no | `nomic-embed-text` | |
| `EMBEDDING_DIM` | no | `768` | Must match the model's actual dim; reindex required if changed |
| `OPENAI_API_KEY` | conditional | — | Required when `EMBEDDING_PROVIDER=openai` |
| `LLM_PROVIDER` | no | `anthropic` | Only `anthropic` for v1 |
| `ANTHROPIC_API_KEY` | conditional | — | Required when `LLM_PROVIDER=anthropic` |
| `ANTHROPIC_MODEL` | no | `claude-sonnet-4-7` | |
| `PORT` | no | `3000` | |
| `LOG_LEVEL` | no | `info` | `fatal` / `error` / `warn` / `info` / `debug` / `trace` |
| `NODE_ENV` | no | `development` | Affects log formatting (pretty in dev, JSON in prod) |

## Creating an API key

Workspaces and API keys are not exposed via HTTP in v1 (no signup flow). Provision directly in the database:

```bash
# 1. Create a workspace
psql $DATABASE_URL -c "INSERT INTO workspaces (slug, name) VALUES ('my-team', 'My Team') RETURNING id;"

# 2. Generate a key (Node REPL)
node -e "
  const { generateApiKey, hashApiKey, extractKeyPrefix } = await import('./dist/auth/api-key.js');
  const key = generateApiKey();
  const hash = await hashApiKey(key);
  console.log('KEY:', key);
  console.log('PREFIX:', extractKeyPrefix(key));
  console.log('HASH:', hash);
"

# 3. Insert (use the workspace UUID from step 1, and outputs from step 2)
psql $DATABASE_URL -c "
  INSERT INTO api_keys (workspace_id, key_hash, key_prefix, name)
  VALUES ('<workspace-uuid>', '<hash>', '<prefix>', 'my key');
"
```

Save the plaintext `KEY` value — it's not recoverable from the database.

A `tolvi keys create` CLI command lands in Phase 3.

## API

The OpenAPI spec is generated from the Fastify route schemas and committed to `spec/openapi.json` in PR B (search + ask routes + the OpenAPI generator land together there). For now, see the route definitions under `src/routes/` for request/response shapes.

In dev, Swagger UI will be available at `http://localhost:3000/docs` once PR B's OpenAPI generation lands.

## Tests

```bash
npm run test           # unit + integration (testcontainers; requires Docker)
npm run test:unit      # fast, no Docker
npm run test:integration
npm run test:e2e       # full stack including real Ollama; slow (PR B)
```

## Operations

### Vacuum soft-deleted documents

Soft-deleted documents accumulate. Periodically:

```sql
DELETE FROM documents WHERE deleted_at < now() - interval '30 days';
```

### Rotate an API key

```sql
UPDATE api_keys SET revoked_at = now() WHERE id = '<key-uuid>';
-- Then create a new one via the steps above.
```

### Reindex after EMBEDDING_DIM change

Changing `EMBEDDING_DIM` requires regenerating the migration and reindexing all chunks:

```bash
TRUNCATE chunks;                            # drops all embeddings
EMBEDDING_DIM=<new> npm run migrate:generate --name embedding_dim_change
npm run migrate
# Then re-ingest all docs by re-POSTing to /v1/sync
```

## License

Apache 2.0. See [`../LICENSE`](../LICENSE) and [`../NOTICE`](../NOTICE).
