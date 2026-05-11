import { createHash } from 'node:crypto';

function canonicalJson(value: unknown): string {
  if (value === null || typeof value !== 'object') return JSON.stringify(value);
  if (Array.isArray(value)) return `[${value.map(canonicalJson).join(',')}]`;
  const keys = Object.keys(value as Record<string, unknown>).sort();
  const entries = keys.map((k) => `${JSON.stringify(k)}:${canonicalJson((value as Record<string, unknown>)[k])}`);
  return `{${entries.join(',')}}`;
}

export function computeContentHash(frontmatter: unknown, body: string): string {
  const fm = canonicalJson(frontmatter);
  const trimmedBody = body.replace(/\s+$/g, '');
  return createHash('sha256').update(fm).update('\n').update(trimmedBody).digest('hex');
}
