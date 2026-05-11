import { describe, it, expect } from 'vitest';
import { chunkMarkdown, estimateTokens } from '../../src/format/chunking.js';

describe('estimateTokens', () => {
  it('approximates tokens via 4-chars-per-token heuristic', () => {
    expect(estimateTokens('hello')).toBe(2);                            // 5 chars / 4 = 1.25, ceil = 2
    expect(estimateTokens('the quick brown fox')).toBe(5);              // 19 chars / 4 = 4.75, ceil = 5
    expect(estimateTokens('')).toBe(0);
  });
});

describe('chunkMarkdown', () => {
  it('returns a single chunk for very short content', () => {
    const chunks = chunkMarkdown('# Title\n\nShort body.', { maxTokens: 512 });
    expect(chunks).toHaveLength(1);
    expect(chunks[0]?.content).toContain('Short body');
    expect(chunks[0]?.position).toBe(0);
    expect(chunks[0]?.headingPath).toEqual(['Title']);
  });

  it('splits at heading boundaries when content exceeds maxTokens', () => {
    const longSection = 'word '.repeat(300);                            // ~1500 chars = ~375 tokens
    const md = `## Section A\n\n${longSection}\n\n## Section B\n\n${longSection}`;
    const chunks = chunkMarkdown(md, { maxTokens: 200 });
    expect(chunks.length).toBeGreaterThanOrEqual(2);
    // Each chunk should carry its heading path
    const hasA = chunks.some((c) => c.headingPath.includes('Section A'));
    const hasB = chunks.some((c) => c.headingPath.includes('Section B'));
    expect(hasA).toBe(true);
    expect(hasB).toBe(true);
  });

  it('preserves heading path breadcrumb across nested headings', () => {
    const md = `# Top\n\n## Middle\n\n### Inner\n\nContent here.`;
    const chunks = chunkMarkdown(md, { maxTokens: 512 });
    expect(chunks[0]?.headingPath).toEqual(['Top', 'Middle', 'Inner']);
  });

  it('assigns sequential 0-based positions', () => {
    const md = '## A\n\n' + 'word '.repeat(300) + '\n\n## B\n\n' + 'word '.repeat(300);
    const chunks = chunkMarkdown(md, { maxTokens: 100 });
    chunks.forEach((c, i) => expect(c.position).toBe(i));
  });

  it('produces non-empty chunks only', () => {
    const chunks = chunkMarkdown('## A\n\n## B\n\nReal content.', { maxTokens: 512 });
    chunks.forEach((c) => expect(c.content.trim().length).toBeGreaterThan(0));
  });
});
