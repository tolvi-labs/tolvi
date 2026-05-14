# Embedded schemas

These four files are byte-identical copies of `<repo-root>/spec/schemas/*.json`.

Why duplicated: Go's `//go:embed` directive can only reach files at or
below the source file's directory. The CLI module lives at `cli/` and
cannot embed `../spec/schemas/*` (path traversal forbidden by go:embed).

Sync policy: when the canonical schemas at `spec/schemas/` change, copy
them into this directory and re-build. CI verifies the copies match the
canonical sources (see `.github/workflows/cli.yml`, added in Task 25).
