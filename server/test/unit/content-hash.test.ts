import { describe, it, expect } from 'vitest';
import { computeContentHash } from '../../src/ingest/content-hash.js';

describe('computeContentHash', () => {
  it('produces a stable 64-char sha256 hex', () => {
    const hash = computeContentHash({ tags: ['decision'] }, 'body');
    expect(hash).toMatch(/^[a-f0-9]{64}$/);
  });

  it('is identical for the same content+frontmatter', () => {
    const a = computeContentHash({ tags: ['decision'], status: 'active' }, 'body');
    const b = computeContentHash({ tags: ['decision'], status: 'active' }, 'body');
    expect(a).toBe(b);
  });

  it('is invariant under frontmatter key order', () => {
    const a = computeContentHash({ tags: ['decision'], status: 'active' }, 'body');
    const b = computeContentHash({ status: 'active', tags: ['decision'] }, 'body');
    expect(a).toBe(b);
  });

  it('is invariant under trailing whitespace in body', () => {
    const a = computeContentHash({}, 'body content');
    const b = computeContentHash({}, 'body content   \n\n');
    expect(a).toBe(b);
  });

  it('changes when body content actually changes', () => {
    const a = computeContentHash({}, 'body one');
    const b = computeContentHash({}, 'body two');
    expect(a).not.toBe(b);
  });
});
