---
tags: [decision, tolvi, integrations, plugins, hooks, codex, cursor, gemini]
date: 2026-09-16
repo: tolvi
status: active
ticket: none
user_impact: high
product_area: Agent integrations
---

# Codex, Cursor and Gemini CLI install Tolvi as a plugin or extension, and each README states which install was verified

**Date:** 2026-09-16
**Repo:** tolvi

## Why

The repository already shipped plugin manifests for Codex, Cursor and Gemini CLI, but none of them could be used as documented: Codex had no marketplace file to install from, Cursor's session recall answered in a format Cursor does not read, and the Gemini extension loaded contributor instructions as context for every user. Plugin installs are the simplest path for users, so each needed to actually work, and the docs needed to say plainly which ones have been run.

## How

- **Codex:** `.agents/plugins/marketplace.json` defines marketplace `tolvi` with one plugin `tolvi` whose source is local `./`, the repository root that holds `.codex-plugin/plugin.json`. This is the layout Codex's bundled marketplaces use (`./plugins/<name>`), and a marketplace added from `tolvi-labs/tolvi` is a clone of the repository, so `./` resolves the same way. Users run `codex plugin marketplace add tolvi-labs/tolvi` then `codex plugin add tolvi@tolvi`.
- **Codex verification:** with Codex CLI 0.154.0 and the marketplace added from a local checkout, the agent did not list the tolvi skill before the install and did list it after. Updating with `codex plugin marketplace upgrade tolvi` was not exercised (it exits 1 for a local marketplace, and applies to a marketplace added from GitHub), and the README says so in its Update section.
- **Cursor recall hook:** `skills/tolvi/hooks/tolvi-recall` serves both Claude Code and the Cursor plugin (`hooks-cursor.json`). It treats a payload containing `cursor_version` or `composer_mode` as Cursor and answers `{"additional_context": "..."}`, the field Cursor's `sessionStart` reads, re-emitting CLI recall output rather than exec'ing it. Any other payload is Claude Code, and its output is unchanged byte for byte.
- **Cursor vault scope:** in Cursor mode the hook reads only the vault of the first `workspace_roots` entry, which must be a list whose first entry is an existing directory; otherwise it prints nothing. A plugin hook's working directory is not documented and may be the plugin folder itself, which contains this repository's own `vault/`, so falling back to the working directory could put another project's decisions into a user's chat.
- **Test:** `.github/scripts/test-recall-hook-formats.sh`, run in CI, pins the Claude Code startup and `/clear` outputs as exact bytes, checks the Cursor shape found through `workspace_roots` from an unrelated directory, checks CLI recall re-emitted for Cursor, and checks that a Cursor payload without a usable workspace root prints nothing even from inside a vault. The fixture hides the CLI with a minimal PATH; because the hook also looks in `/usr/local/bin` and `/opt/homebrew/bin`, the test fails with an explanation on a machine with a `tolvi` binary there rather than skipping.
- **Cursor docs:** the README documents a local plugin install (a copy under `~/.cursor/plugins/local/tolvi`, then reload), after the project install and marked not yet verified, and warns that a user who also installed the Claude Code session hooks and lets Cursor include third-party configs should use one or the other, or recall arrives twice.
- **Gemini CLI:** `gemini-extension.json` now loads `GEMINI-EXTENSION.md`, a short user-facing file about the Tolvi skill; `GEMINI.md` stays the contributor pointer. Installing from `https://github.com/tolvi-labs/tolvi` without a ref resolves to the release marked Latest, whose tag does not contain the extension, so the README installs with `--ref main`, where Gemini CLI offers a git clone install that tracks `main`. That command has not been run: a local-path install stopped at an interactive folder-trust prompt that `--consent` does not answer. The integrations table marks Gemini CLI install as not yet verified.
- **Guards:** `plugin-manifest-check.sh` fails if the Codex marketplace is missing, does not list the Codex plugin, or does not use the local `./` source, and if `gemini-extension.json` points back at `GEMINI.md`.
- **Rejected:** a `tolvi recall --format cursor-json` CLI flag, because the hook script can reshape the output without a CLI release; a Git URL source in the Codex marketplace, because the local source lets the install be tested before anything is pushed; documenting Cursor team marketplace import, because it likely needs a Cursor marketplace file this repository does not have and has not been tried.
- **Known limits:** Devin and Hermes manifests are unchanged and untested. With an existing vault whose sessions and decisions folders contain no Markdown files, the recall hook's direct read exits non-zero for both agents; this predates the change and fixing it would change Claude Code behavior.

## Outcome

Codex installs Tolvi as a plugin and that path is verified; Cursor receives recall in its own format and only for the open workspace; Gemini CLI users get user-facing context and an install command that targets `main`; and every Tier 2 README says exactly which install paths have been run.
