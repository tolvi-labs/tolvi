# Tolvi

This extension adds the Tolvi skill, which reads, writes and asks questions of a Tolvi engineering vault: decisions, sessions and patterns stored as Markdown under `vault/` in a repository.

Use the Tolvi skill when the current repository contains `vault/.vault-meta.json`, or when the user asks about past decisions, session notes or patterns. In a repository without a vault, offer to run `tolvi init` and wait for confirmation.

The `tolvi` CLI is optional but recommended. Without it, the skill reads the vault files directly. To install it:

```bash
go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

This file is context for people using the extension.
