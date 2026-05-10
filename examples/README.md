# Examples

Synthetic content used for documentation, demos, and CI validation.

- [`sample-vault/`](./sample-vault/) — a complete `tolvi-format-v1` vault with 6 decisions, 3 session-day files, and 4 patterns. All content is fully synthetic; no real-world content is used. The CI workflow validates every file in this vault against the JSON Schemas in [`../spec/schemas/`](../spec/schemas/).

To validate locally:

```bash
npx ajv-cli@5 validate -s ../spec/schemas/decision.json -d 'sample-vault/decisions/*.md' --extract-frontmatter
```
