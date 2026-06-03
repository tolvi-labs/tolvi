#!/usr/bin/env bash
# session-recall.sh — Claude Code SessionStart hook for Tolvi.
#
# Runs on every session start (startup + resume). Emits the
# SessionStart hook JSON blob so Claude receives recent vault context
# as additionalContext before the first user message.
#
# Silently exits 0 on any error so a missing CLI or vault never
# breaks a session.
#
# Installed by: integrations/claude-code/install.sh --with-hooks
# See: https://github.com/tolvi-labs/tolvi

command -v tolvi >/dev/null 2>&1 || exit 0
tolvi recall --format hook-json 2>/dev/null || exit 0
