---
tags: [decision, tolvi, registry, repos, json-contract]
date: 2026-09-23
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: CLI / repos registry
---

# The repo registry is a cache, and scan searches what it is told rather than guessing

**Date:** 2026-09-23
**Repo:** tolvi

## TL;DR

`tolvi repos list | scan | forget` maintains `~/.config/tolvi/repos.json`, the machine-local index of repos that have a vault. It is a cache over each repo's own `.vault-meta.json`: entries are verified when read, a repo that moved is reported stale rather than trusted or deleted, and a scan rebuilds the file from nothing. `scan` searches directories it is given, defaulting to the declared roots, and refuses to invent a search path when there are none.

## Why

Nothing could enumerate the repos on a machine that have vaults. `roots.json` declares org and product roots and explicitly refuses to name repo roots, since a repo root resolves from the working directory, so there was no data source for "which repos exist here" at all. That made any per-machine view of the vaults impossible to build without each consumer inventing its own discovery, which is the drift this codebase has already paid for once.

## How

- **The truth is each repo's `.vault-meta.json`; the registry is an index over them.** Entries are verified on read, and `Verify` never rewrites the file. What to do about a stale entry is the caller's decision, and the answer is to report it.
- **Stale, never silently deleted.** A repo that moved keeps its entry, marked `stale`, because deleting it destroys the only record that it was ever registered, and a path that is absent today may be a mounted volume tomorrow. `tolvi repos forget` is the deliberate removal.
- **A stale entry gets no routing answers.** `vault_path` and `session_note` are omitted rather than computed from the cached identity, because a repo that is not there cannot be resolved and a guess would look like a fact.
- **No `schema_version`, matching `roots.json`.** A cache that a scan rebuilds needs no migration path.
- **`scan` searches what it is given.** With arguments it searches those directories. With none it searches the directories holding the declared roots, which is the set this machine already knows about. With neither it fails and says to name a directory. Rejected: a configurable search-paths key, which invents configuration before there is evidence anyone needs it; and an implicit home-directory walk, which is slow and surprising.
- **Scans skip dot-directories and vendor trees** (`node_modules`, `vendor`, `dist`, `build`, `target`). A checked-out or vendored copy of another repo is not this machine's repo. An unreadable directory is skipped rather than fatal, because a scan that dies on one permission error finds nothing.
- **Re-registering keeps the original timestamp**, so `tolvi init` on an existing repo does not churn the file.
- **`list` returns what the resolver computed, not raw entries.** `vault_path` and `session_note` come from `vault.ChainFor` and `SessionNotePath`, so a consumer reads routing rather than re-deriving it. That is the failure this exists to prevent: a fifth consumer re-derived the routing rule and had never indexed a routed session note.
- **A refusal is not a stale entry.** Where a workspace declares no private root, the resolver refuses to route a session; `session_note` is then absent while `status` stays `ok`. The two cases stay distinguishable by `status`. The refusal reason is not carried in the summary; `tolvi roots` gives it.
- **The shape is published** as `spec/schemas/repos-list.json` under the `outputs` key, per `2026-09-23-output-schemas-version-with-the-cli-not-the-format`, and validated from this repo's own tests.

## Outcome

A machine can be asked which repos have vaults, and the answer carries each repo's resolved routing without any consumer re-deriving it, while remaining a cache that is safe to delete.
