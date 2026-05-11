import { unified } from 'unified';
import remarkParse from 'remark-parse';
import remarkStringify from 'remark-stringify';
import type { Root, Heading, Content } from 'mdast';

export type Chunk = {
  position: number;
  content: string;
  headingPath: string[];
};

export type ChunkOptions = {
  maxTokens: number;
};

export function estimateTokens(text: string): number {
  if (text.length === 0) return 0;
  return Math.ceil(text.length / 4);
}

function headingText(node: Heading): string {
  return node.children
    .map((c) => ('value' in c ? c.value : ''))
    .join('')
    .trim();
}

function nodeToMarkdown(node: Content): string {
  return unified().use(remarkStringify).stringify({ type: 'root', children: [node] } as Root).trim();
}

/**
 * Split a markdown body into ≤maxTokens chunks. Splits at heading boundaries
 * when possible. Each chunk carries the breadcrumb of headings that lead to it.
 */
export function chunkMarkdown(markdown: string, opts: ChunkOptions): Chunk[] {
  const tree = unified().use(remarkParse).parse(markdown) as Root;

  // Walk the AST, grouping content under each heading sequence.
  type Section = { headingPath: string[]; nodes: Content[] };
  const sections: Section[] = [];
  let stack: { depth: number; text: string }[] = [];

  for (const node of tree.children) {
    if (node.type === 'heading') {
      const h = node as Heading;
      const text = headingText(h);
      // Pop deeper-or-equal headings off the stack
      while (stack.length > 0 && stack[stack.length - 1]!.depth >= h.depth) {
        stack.pop();
      }
      stack.push({ depth: h.depth, text });
      sections.push({ headingPath: stack.map((s) => s.text), nodes: [] });
    } else {
      if (sections.length === 0) {
        // Content before any heading
        sections.push({ headingPath: [], nodes: [] });
      }
      sections[sections.length - 1]!.nodes.push(node);
    }
  }

  // Now turn sections into chunks, splitting any section whose serialized
  // length exceeds maxTokens into smaller pieces at paragraph boundaries.
  const chunks: Chunk[] = [];
  let position = 0;
  for (const section of sections) {
    if (section.nodes.length === 0) continue;
    const serialized = section.nodes.map(nodeToMarkdown).filter((s) => s.length > 0);
    let buffer = '';
    for (const piece of serialized) {
      const candidate = buffer ? `${buffer}\n\n${piece}` : piece;
      if (estimateTokens(candidate) > opts.maxTokens && buffer.length > 0) {
        chunks.push({ position: position++, content: buffer, headingPath: section.headingPath });
        buffer = piece;
      } else {
        buffer = candidate;
      }
    }
    if (buffer.length > 0) {
      chunks.push({ position: position++, content: buffer, headingPath: section.headingPath });
    }
  }

  return chunks;
}
