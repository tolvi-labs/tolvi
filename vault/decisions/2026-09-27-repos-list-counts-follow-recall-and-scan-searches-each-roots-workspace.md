---
tags: [decision, tolvi, registry, repos, json-contract, testing]
date: 2026-09-27
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: CLI / repos registry
---

# repos list counts a repo's docs with recall's own rule, and the default scan searches each root's workspace

**Date:** 2026-09-27
**Repo:** tolvi

## TL;DR

`tolvi repos list --json` now reports `decisions`, `patterns`, `sessions` and `last_session` for each repo, counted from the sources and filters `tolvi recall` already uses, so a consumer never re-derives routing to show them. A no-argument `tolvi repos scan` now searches the workspace directory each declared root sits in, which finds every repo in it rather than only the root's host. The CLI's integration tests stopped writing temp vaults into the developer's real `~/.config/tolvi/repos.json`.

## Why

Anything that lists repos wants to say how much is in each and when it was last worked on. Before this, the only way to know was to walk the vault roots and apply the routing rule yourself, which is exactly the re-derivation the resolver exists to prevent. Separately, a no-argument scan registered one repo per workspace, and every run of the test suite quietly added temp vaults to the developer's own registry.

## How

- **The counts reuse recall, not a copy of it.** `countRepoDocs` reads `recallSources`, the repo's own vault plus each shared root in its chain, and applies recall's filters: `sessionBelongsTo` for session filenames, and a new `decisionBelongsTo`, lifted out of recall's decision loader so both commands call one predicate. Recall and `repos list` therefore cannot disagree about which docs belong to a repo.
- **Decisions count every status.** The figure is an inventory; recall's active-only filter is about what is relevant to surface at session start, and applying it here would make the count shrink whenever a decision is superseded.
- **Patterns count only the repo's own vault.** A pattern has no `repo` field, so a pattern in a shared root cannot honestly be attributed to any one repo.
- **Absent is not zero.** All four fields are omitted for a stale entry and whenever the vault cannot be read, like `session_note`; a vault with nothing in it yet prints real zeros. The wire struct uses pointers for the three counts so the encoder can tell the two apart.
- **The schema grows additively.** `spec/schemas/repos-list.json` and its embedded copy in `cli/internal/format/schemas/` gain the four optional properties, which `2026-09-23-output-schemas-version-with-the-cli-not-the-format` permits within a package major. The existing schema-validation test covers them, and it failed first, because the schema forbids unknown properties.
- **Checked against a real machine.** For a vault with 107 decision files and 31 of its own session notes, the counts were 107 and 31, and routed notes in a shared root were attributed to the repos they name.
- **The default scan goes two levels up.** A declared root is a repo's vault, `<workspace>/<repo>/vault`, so the repo holding it is one level up and the workspace its siblings share is two. `2026-09-23-the-repo-registry-is-a-cache-and-scan-searches-what-it-is-told` set the default to the declared roots; this keeps that principle and corrects the directory it derives. Workspaces also nest, a client or product workspace inside an org's, so a search directory inside another is dropped rather than walked twice. On a machine with four declared roots the default now registers 24 repos from 2 directories in under half a second, where it found 4.
- **Test isolation is the default, not opt-in.** 18 of the 27 places the integration tests run the built binary set no environment, so they inherited the developer's `XDG_CONFIG_HOME`, and `tolvi init` registered each temp vault into the real registry. A `TestMain` in `cmd/tolvi` now points `XDG_CONFIG_HOME` at a temp directory for the whole test process, so every spawned binary inherits it. `HOME` is left alone because the tests' git commits read the global git identity. A new test asserts the isolation before running anything, so its own failure cannot write to the real registry. After a full run of every integration test the developer's registry was unchanged.
- **Found on the way, not fixed here.** `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest` resolves to v0.1.0, because the module lives in `cli/` and Go resolves such a module's versions from `cli/v*` tags only; v0.1.1 through v0.2.0 were tagged `v*`. Every go-install instruction, including doctor's own fix for the `path` check, installs a build without `repos`. The fix is a `cli/vX.Y.Z` tag beside each release tag.

## Outcome

`repos list --json` answers how much each repo holds and when it was last worked on, by the same rule recall uses; a bare `repos scan` finds every repo in each declared workspace; and running the test suite no longer touches the developer's registry.
