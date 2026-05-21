# Tolvi — OpenHands integration (skeleton)

A `.openhands_instructions` template that teaches [OpenHands](https://www.all-hands.dev) about Tolvi vault conventions. OpenHands reads this file from the project root as part of its repo-aware context.

## Tier

OpenHands is a **Tier 3 — skeleton** integration in the [Tolvi integrations tier list](../README.md). The file ships a compact format-spec recap; it does not include the deeper behavioral guidance or worked examples of the Tier 1 [Claude Code skill](../claude-code/).

## Prerequisites

- [OpenHands](https://www.all-hands.dev) installed and configured.
- The `tolvi` CLI in `$PATH` (see <https://github.com/tolvi-labs/tolvi/releases>).
- A Tolvi vault in your repo (`vault/.vault-meta.json`); run `tolvi init` otherwise.

## Install

Copy `.openhands_instructions` from this directory to the **root of your project**:

```bash
cp /path/to/tolvi-labs/tolvi/integrations/openhands/.openhands_instructions .openhands_instructions
```

Or symlink so `git pull` on the Tolvi checkout flows through:

```bash
ln -s /path/to/tolvi-labs/tolvi/integrations/openhands/.openhands_instructions .openhands_instructions
```

Commit the file so it's shared with everyone on the team.

## Uninstall

```bash
rm .openhands_instructions
```

## Caveat

OpenHands's repo-instructions mechanism evolves (the project is under active development as the OpenDevin → OpenHands rebrand settles). If the install steps above don't match the current docs, defer to <https://docs.all-hands.dev> for canonical instructions; the file content stays the same.
