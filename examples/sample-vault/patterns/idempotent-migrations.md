---
tags: [pattern, sample, database, migrations]
status: active
languages: [sql]
---

# Idempotent migrations

Database migrations should be safe to re-run. A migration that has already been applied should be a no-op the second time, not a duplicate-key error or a "column already exists" crash. This becomes load-bearing the moment a migration partially fails (network blip, deploy aborted halfway through, the runner OOM-kills itself between two statements) and somebody has to re-run it without first manually inspecting which statements landed.

The pattern is: every DDL statement uses the existence-aware form your database supports, every DML statement uses an upsert or a `WHERE NOT EXISTS` guard, and every migration is wrapped in a transaction where the engine permits it. For databases that do not allow transactional DDL (older versions of some engines), break the migration into the smallest possible self-contained units and rely on the migration runner's tracking table for ordering.

A small example:

```sql
-- Safe to re-run.
CREATE TABLE IF NOT EXISTS audit_event (
  id          BIGSERIAL PRIMARY KEY,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  payload     JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS audit_event_occurred_at_idx
  ON audit_event (occurred_at);

INSERT INTO config (key, value)
VALUES ('audit.retention_days', '30')
ON CONFLICT (key) DO NOTHING;
```

When to use this: every schema change, without exception. When not to use it: never; even one-off backfills benefit from being re-runnable.
