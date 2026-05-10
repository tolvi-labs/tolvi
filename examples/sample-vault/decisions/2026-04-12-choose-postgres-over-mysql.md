---
tags: [decision, sample, database]
date: 2026-04-12
status: active
repo: tolvi-examples
ticket: none
---

## Why

We needed a primary relational store for the application's transactional data. The two finalists were PostgreSQL and MySQL. Both are mature and operationally well understood, but our workload has three characteristics that pushed us off MySQL: heavy use of JSON columns for evolving event payloads, several reporting queries that lean on window functions and CTEs, and a need for partial and expression indexes to keep a few hot tables small. We also expect to use logical replication to fan out changes to a search index without bolting on a third-party CDC pipeline.

## How

We chose PostgreSQL 16. The default tablespace lives on local NVMe with daily base backups and continuous WAL shipping to object storage. Connections go through PgBouncer in transaction-pooling mode. Schema migrations follow our [[idempotent-migrations]] pattern so that they can be re-run safely on partial failures. We deliberately kept the surface area small: no extensions beyond `pg_stat_statements`, `pgcrypto`, and `pg_trgm` until a concrete need shows up.

## Outcome

The store has been in production for three weeks. JSONB indexes on the event table cut a hot dashboard query from 1.4s to 60ms. The team is comfortable with the operational story; we have not yet needed to revisit the choice.
