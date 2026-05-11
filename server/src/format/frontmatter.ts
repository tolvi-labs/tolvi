import Ajv from 'ajv/dist/2020.js';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Minimal YAML parser for the Tolvi frontmatter subset (matches the
// extract-frontmatter.js helper from Phase 0+1's CI guard exactly).
function parseYaml(text: string): Record<string, unknown> {
  const obj: Record<string, unknown> = {};
  const lines = text.split('\n');
  let i = 0;
  while (i < lines.length) {
    const line = lines[i] ?? '';
    if (!line.trim() || line.trim().startsWith('#')) {
      i++;
      continue;
    }
    const kv = line.match(/^([a-zA-Z_][a-zA-Z0-9_-]*):\s*(.*)$/);
    if (!kv) {
      i++;
      continue;
    }
    const key = kv[1]!;
    const rawValue = (kv[2] ?? '').trim();

    if (rawValue === '') {
      const arr: string[] = [];
      i++;
      while (i < lines.length && /^\s*-\s+/.test(lines[i] ?? '')) {
        arr.push((lines[i] ?? '').replace(/^\s*-\s+/, '').trim().replace(/^["']|["']$/g, ''));
        i++;
      }
      obj[key] = arr;
    } else if (rawValue.startsWith('[') && rawValue.endsWith(']')) {
      const inner = rawValue.slice(1, -1).trim();
      obj[key] = inner === '' ? [] : inner.split(',').map((s) => s.trim().replace(/^["']|["']$/g, ''));
      i++;
    } else if (/^-?\d+$/.test(rawValue)) {
      obj[key] = parseInt(rawValue, 10);
      i++;
    } else {
      obj[key] = rawValue.replace(/^["']|["']$/g, '');
      i++;
    }
  }
  return obj;
}

export type ParsedDocument = {
  frontmatter: Record<string, unknown>;
  body: string;
  title: string;
};

const FRONTMATTER_REGEX = /^---\n([\s\S]*?)\n---\n?/;

export function parseDocument(text: string): ParsedDocument {
  const match = text.match(FRONTMATTER_REGEX);
  if (!match) {
    throw new Error('No frontmatter found (expected leading --- ... --- block)');
  }
  const frontmatter = parseYaml(match[1] ?? '');
  const body = text.slice(match[0].length);

  // Title: first H1, fall back to first H2, fall back to 'Untitled'
  const h1 = body.match(/^#\s+(.+)$/m);
  const h2 = body.match(/^##\s+(.+)$/m);
  const title = (h1?.[1] ?? h2?.[1] ?? 'Untitled').trim();

  return { frontmatter, body, title };
}

// Load the JSON Schemas committed in spec/schemas/ (relative to the server source location).
// In the built dist/, paths shift, so we resolve via process.cwd() with a fallback.
function loadSchema(name: string): object {
  const candidates = [
    path.resolve(__dirname, '../../../spec/schemas', `${name}.json`),
    path.resolve(__dirname, '../../../../spec/schemas', `${name}.json`),
    path.resolve(process.cwd(), 'spec/schemas', `${name}.json`),
    path.resolve(process.cwd(), '../spec/schemas', `${name}.json`),
  ];
  for (const c of candidates) {
    try {
      return JSON.parse(readFileSync(c, 'utf8'));
    } catch {
      continue;
    }
  }
  throw new Error(`Could not locate spec/schemas/${name}.json (tried: ${candidates.join(', ')})`);
}

const ajv = new Ajv({ strict: true, allErrors: true });
const validators = {
  decision: ajv.compile(loadSchema('decision')),
  session: ajv.compile(loadSchema('session')),
  pattern: ajv.compile(loadSchema('pattern')),
};

export type ValidationError = {
  instancePath: string;
  message: string;
};

export type ValidationResult =
  | { valid: true; errors?: undefined }
  | { valid: false; errors: ValidationError[] };

export function validateFrontmatter(
  docType: 'decision' | 'session' | 'pattern',
  frontmatter: unknown
): ValidationResult {
  const validator = validators[docType];
  const valid = validator(frontmatter);
  if (valid) return { valid: true };
  return {
    valid: false,
    errors: (validator.errors ?? []).map((e) => ({
      instancePath: e.instancePath || '/',
      message: e.message ?? 'unknown error',
    })),
  };
}
