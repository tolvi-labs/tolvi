---
tags: [pattern, sample, database, concurrency]
status: active
languages: [sql]
---

# Postgres advisory locks

Postgres advisory locks are application-defined locks that the database tracks but does not interpret. They are useful for coordinating cross-process work where you want mutual exclusion without inventing a separate distributed lock service. Common uses: ensuring only one worker runs a periodic job at a time, serializing access to a non-transactional external resource, or avoiding the thundering-herd problem when many workers wake up to do the same recovery work.

There are two flavors. Session-level advisory locks are released when the connection closes — durable and forgiving, but require careful handling around connection poolers like PgBouncer in transaction-pooling mode (where there is no stable session). Transaction-level advisory locks are released at commit or rollback — pooler-safe, and preferred when your work fits inside a single transaction.

The lock key is two `int4`s or one `int8`. Pick a deterministic key by hashing the resource identifier; that way every worker computing the same key competes for the same lock, regardless of how they were scheduled.

```sql
-- Try to acquire a transaction-scoped advisory lock for "rebuild-search-index".
-- Returns true if acquired, false if another transaction already holds it.
SELECT pg_try_advisory_xact_lock(hashtext('rebuild-search-index'));

-- If true, do the work. The lock releases automatically at COMMIT.
```

When to use this: coordination among processes that already share a Postgres database. When not to use this: cross-region coordination (advisory locks are local to one Postgres cluster), or work that must outlive a single transaction or session.
