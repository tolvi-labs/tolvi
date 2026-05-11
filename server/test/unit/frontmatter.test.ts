import { describe, it, expect } from 'vitest';
import { parseDocument, validateFrontmatter } from '../../src/format/frontmatter.js';

describe('parseDocument', () => {
  it('splits frontmatter and body', () => {
    const text = `---\ntags: [decision]\nstatus: active\ndate: 2026-05-10\nrepo: tolvi\n---\n\n## Why\n\nBecause.`;
    const result = parseDocument(text);
    expect(result.frontmatter).toEqual({
      tags: ['decision'],
      status: 'active',
      date: '2026-05-10',
      repo: 'tolvi',
    });
    expect(result.body.trim()).toBe('## Why\n\nBecause.');
  });

  it('throws when no frontmatter is present', () => {
    expect(() => parseDocument('# Just a heading')).toThrow(/frontmatter/i);
  });

  it('extracts the title from the first H1, falling back to the first H2', () => {
    const text = `---\ntags: [decision]\nstatus: active\ndate: 2026-05-10\nrepo: tolvi\n---\n\n## Why\n\nBecause.`;
    expect(parseDocument(text).title).toBe('Why');

    const withH1 = `---\ntags: [decision]\nstatus: active\ndate: 2026-05-10\nrepo: tolvi\n---\n\n# The Real Title\n\n## Why`;
    expect(parseDocument(withH1).title).toBe('The Real Title');
  });

  it('falls back to "Untitled" when no heading is present', () => {
    const text = `---\ntags: [decision]\nstatus: active\ndate: 2026-05-10\nrepo: tolvi\n---\n\nJust prose, no heading.`;
    expect(parseDocument(text).title).toBe('Untitled');
  });
});

describe('validateFrontmatter', () => {
  it('validates a conformant decision frontmatter', () => {
    const fm = { tags: ['decision'], status: 'active', date: '2026-05-10', repo: 'tolvi' };
    const result = validateFrontmatter('decision', fm);
    expect(result.valid).toBe(true);
    expect(result.errors).toBeUndefined();
  });

  it('rejects a decision missing required repo', () => {
    const fm = { tags: ['decision'], status: 'active', date: '2026-05-10' };
    const result = validateFrontmatter('decision', fm);
    expect(result.valid).toBe(false);
    expect(result.errors).toBeDefined();
    expect(result.errors!.some((e) => e.message.includes('repo'))).toBe(true);
  });

  it('rejects a decision with an unknown status value', () => {
    const fm = { tags: ['decision'], status: 'pending', date: '2026-05-10', repo: 'tolvi' };
    const result = validateFrontmatter('decision', fm);
    expect(result.valid).toBe(false);
  });

  it('accepts a pattern with no date or repo', () => {
    const fm = { tags: ['pattern'], status: 'active' };
    const result = validateFrontmatter('pattern', fm);
    expect(result.valid).toBe(true);
  });

  it('accepts an x-* extension field on any doc type', () => {
    const fm = { tags: ['decision'], status: 'active', date: '2026-05-10', repo: 'tolvi', 'x-priority': 'high' };
    const result = validateFrontmatter('decision', fm);
    expect(result.valid).toBe(true);
  });

  it('rejects an unknown non-x-* field (typo guard)', () => {
    const fm = { tags: ['decision'], status: 'active', date: '2026-05-10', repo: 'tolvi', statuss: 'active' };
    const result = validateFrontmatter('decision', fm);
    expect(result.valid).toBe(false);
  });
});
