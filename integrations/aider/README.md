# Tolvi — Aider integration (skeleton)

A `CONVENTIONS.md` template that teaches [Aider](https://aider.chat) about Tolvi vault conventions. Aider reads convention files at session start (when invoked with `--read CONVENTIONS.md`, or auto-loaded if listed in `.aider.conf.yml`).

## Tier

Aider is a **Tier 3 — skeleton** integration in the [Tolvi integrations tier list](../README.md). The file ships a compact format-spec recap that should let Aider participate in vault read/write flows; it does not include the deeper behavioral guidance or worked examples of the Tier 1 [Claude Code skill](../claude-code/). If you find yourself wanting more, copy from there.

## Prerequisites

- [Aider](https://aider.chat) installed.
- The `tolvi` CLI in `$PATH` (see <https://github.com/tolvi-labs/tolvi/releases>).
- A Tolvi vault in your repo (`vault/.vault-meta.json`); run `tolvi init` otherwise.

## Install

From the root of your project (the repo where you want Aider to know about Tolvi):

```bash
# Copy:
cp /path/to/tolvi-labs/tolvi/integrations/aider/CONVENTIONS.md CONVENTIONS.md
```

Or symlink so `git pull` on the Tolvi checkout flows through:

```bash
ln -s /path/to/tolvi-labs/tolvi/integrations/aider/CONVENTIONS.md CONVENTIONS.md
```

Then either invoke Aider with `--read CONVENTIONS.md` per-session:

```bash
aider --read CONVENTIONS.md
```

Or add the convention to `.aider.conf.yml` so it auto-loads:

```yaml
read:
  - CONVENTIONS.md
```

Commit `CONVENTIONS.md` (and `.aider.conf.yml` if you use it) so the rules are shared with everyone on the team.

## Uninstall

```bash
rm CONVENTIONS.md
# and remove any `read: - CONVENTIONS.md` entry from .aider.conf.yml
```

## Caveat

Aider's convention-loading mechanism evolves. If the install steps above don't match the current Aider docs, defer to <https://aider.chat/docs/usage/conventions.html> for the canonical instructions; the file content stays the same.
