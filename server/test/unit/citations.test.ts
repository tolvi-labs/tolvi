import { describe, it, expect } from 'vitest';
import { extractCitations, scrubUnverifiedCitations } from '../../src/ask/citations.js';

describe('extractCitations', () => {
  it('extracts unique [[slug]] references from text', () => {
    const text = 'See [[idempotent-migrations]] and [[postgres-locks]] for details.';
    expect(extractCitations(text)).toEqual(['idempotent-migrations', 'postgres-locks']);
  });
  it('dedupes repeated citations', () => {
    const text = '[[a]] then [[b]] then [[a]] again.';
    expect(extractCitations(text)).toEqual(['a', 'b']);
  });
  it('ignores non-wiki-link bracket patterns', () => {
    const text = 'Single [brackets] and [[but-not-this stop early.';
    expect(extractCitations(text)).toEqual([]);
  });
  it('only matches slug-shaped contents', () => {
    expect(extractCitations('[[foo bar]]')).toEqual([]);    // space disallowed
    expect(extractCitations('[[FOO]]')).toEqual([]);        // uppercase disallowed
    expect(extractCitations('[[foo-bar]]')).toEqual(['foo-bar']);
  });
});

describe('scrubUnverifiedCitations', () => {
  it('replaces unverified citations with a marker, leaves verified ones alone', () => {
    const verified = new Set(['real-slug']);
    const text = 'See [[real-slug]] and [[fake-slug]] for details.';
    const scrubbed = scrubUnverifiedCitations(text, verified);
    expect(scrubbed).toContain('[[real-slug]]');
    expect(scrubbed).toContain('[unverified citation: fake-slug]');
    expect(scrubbed).not.toContain('[[fake-slug]]');
  });
});
