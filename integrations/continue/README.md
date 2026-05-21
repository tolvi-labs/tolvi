# Tolvi — Continue integration (skeleton)

A `.continuerules` template that teaches [Continue](https://continue.dev) about Tolvi vault conventions. Continue reads this file from the workspace root and uses it as system-prompt context for every interaction in the project.

## Tier

Continue is a **Tier 3 — skeleton** integration in the [Tolvi integrations tier list](../README.md). The file ships a compact format-spec recap; it does not include the deeper behavioral guidance or worked examples of the Tier 1 [Claude Code skill](../claude-code/).

## Prerequisites

- [Continue](https://continue.dev) installed in your editor (VSCode, JetBrains, etc.).
- The `tolvi` CLI in `$PATH` (see <https://github.com/tolvi-labs/tolvi/releases>).
- A Tolvi vault in your repo (`vault/.vault-meta.json`); run `tolvi init` otherwise.

## Install

Copy `.continuerules` from this directory to the **root of your project**:

```bash
cp /path/to/tolvi-labs/tolvi/integrations/continue/.continuerules .continuerules
```

Or symlink so `git pull` on the Tolvi checkout flows through:

```bash
ln -s /path/to/tolvi-labs/tolvi/integrations/continue/.continuerules .continuerules
```

Commit the file so it's shared with everyone on the team.

## Uninstall

```bash
rm .continuerules
```

## Caveat

Continue's rules-file mechanism evolves. If the install steps above don't match the current docs, defer to <https://docs.continue.dev> for canonical instructions; the file content stays the same.
