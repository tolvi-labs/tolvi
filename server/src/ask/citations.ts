const CITATION_REGEX = /\[\[([a-z0-9](?:[a-z0-9-]*[a-z0-9])?)\]\]/g;

export function extractCitations(text: string): string[] {
  const seen = new Set<string>();
  const out: string[] = [];
  let m: RegExpExecArray | null;
  // Reset state by creating fresh regex each call (the global flag is stateful)
  const re = new RegExp(CITATION_REGEX.source, 'g');
  while ((m = re.exec(text)) !== null) {
    const slug = m[1]!;
    if (!seen.has(slug)) {
      seen.add(slug);
      out.push(slug);
    }
  }
  return out;
}

export function scrubUnverifiedCitations(text: string, verifiedSlugs: Set<string>): string {
  const re = new RegExp(CITATION_REGEX.source, 'g');
  return text.replace(re, (full, slug: string) => {
    return verifiedSlugs.has(slug) ? full : `[unverified citation: ${slug}]`;
  });
}
