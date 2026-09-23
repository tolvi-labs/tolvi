#!/usr/bin/env bash
# integration-copy-parity-check.sh — fails (exit 1) if skills/tolvi/ and the
# copy the CLI embeds at cli/internal/integrations/files/ have drifted apart.
#
# The duplication is forced, not chosen. Claude Code's plugin loader requires
# skills/ at the repo root, and Go's //go:embed cannot traverse upward out of
# the cli/ module, so the files cannot sit in one place that serves both. See
# vault/decisions/2026-09-23-duplicate-and-pin-is-a-mitigation-not-a-default.md:
# a forced copy is allowed only when a check pins it, and the check lives in
# the repo that has the copies.
#
# skills/tolvi/ is canonical. Edit there, then re-run this to sync:
#   rsync -a --delete skills/tolvi/ cli/internal/integrations/files/
set -euo pipefail

src="skills/tolvi"
dst="cli/internal/integrations/files"

if [ ! -d "$dst" ]; then
  echo "✗ $dst does not exist; the CLI embeds nothing"
  exit 1
fi

# -a compares content, and the file list both ways catches an addition or a
# deletion on either side, which a content-only diff of common files misses.
if ! diff -r "$src" "$dst" >/dev/null 2>&1; then
  echo "✗ skills/tolvi and its embedded copy have drifted:"
  diff -r "$src" "$dst" | head -20 || true
  echo ""
  echo "skills/tolvi is canonical. Sync with:"
  echo "  rsync -a --delete skills/tolvi/ cli/internal/integrations/files/"
  exit 1
fi

echo "✓ skills/tolvi: embedded copy agrees ($(find "$src" -type f | wc -l | tr -d ' ') files)"
