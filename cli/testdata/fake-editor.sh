#!/usr/bin/env bash
# fake-editor.sh — test stub for $EDITOR.
#
# Reads $TOLVI_TEST_BODY env var and writes it as the file content,
# preserving the frontmatter section that the CLI already wrote.

set -euo pipefail
target="$1"
if [ -z "${TOLVI_TEST_BODY:-}" ]; then
  exit 0
fi

# Read existing content, find the closing frontmatter delimiter, then
# replace the body section with $TOLVI_TEST_BODY.
existing=$(cat "$target")
# Use awk to print up to and including the second `---` line, then append.
awk '
  /^---$/ { count++; print; if (count == 2) { exit } next }
  count >= 1 { print }
' <<< "$existing" > "$target.new"
echo "" >> "$target.new"
printf '%s' "${TOLVI_TEST_BODY}" >> "$target.new"
mv "$target.new" "$target"
