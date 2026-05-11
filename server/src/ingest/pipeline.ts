import { eq, and } from 'drizzle-orm';
import { documents, repos, chunks, type Document } from '../db/schema/index.js';
import type { Db } from '../db/client.js';
import type { EmbeddingProvider } from '../embedding/provider.js';
import { parseDocument, validateFrontmatter, type ValidationError } from '../format/frontmatter.js';
import { chunkMarkdown } from '../format/chunking.js';
import { computeContentHash } from './content-hash.js';

export type IngestResult =
  | { status: 'created' | 'updated'; document: Document; chunks: number }
  | { status: 'unchanged'; document: Document }
  | { status: 'failed'; error: { code: string; message: string; details?: ValidationError[] } };

export type IngestInput = {
  workspaceId: string;
  repoSlug: string;
  path: string;
  content: string;
};

const DOC_TYPES = ['decision', 'session', 'pattern'] as const;
type DocType = (typeof DOC_TYPES)[number];

const SLUG_REGEX = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
const DATE_REGEX = /^\d{4}-\d{2}-\d{2}$/;

function detectDocType(path: string): DocType | null {
  for (const t of DOC_TYPES) {
    if (path.startsWith(`${t}s/`)) return t;
  }
  return null;
}

function extractSlugAndDate(docType: DocType, path: string): { slug: string; date: string | null } {
  const filename = path.replace(/^[^/]+\//, '').replace(/\.md$/, '');
  // Patterns are timeless: filename IS the slug; no date prefix expected.
  if (docType === 'pattern') {
    return { slug: filename, date: null };
  }
  // Sessions: filename is YYYY-MM-DD; the date IS the slug (one session file per day).
  // Decisions: filename is YYYY-MM-DD-slug; the slug after the date prefix is required.
  const match = filename.match(/^(\d{4}-\d{2}-\d{2})(?:-(.+))?$/);
  if (!match) return { slug: filename, date: null };
  const datePart = match[1] ?? null;
  const slugPart = match[2];
  if (docType === 'decision') {
    // Required slug after date — without it, slug would silently equal the date,
    // which is the session-only convention and wrong for decisions.
    return { slug: slugPart ?? '', date: datePart };
  }
  // session: date IS the slug
  return { slug: slugPart ?? datePart ?? filename, date: datePart };
}

export async function ingestDocument(
  db: Db,
  embedding: EmbeddingProvider,
  input: IngestInput
): Promise<IngestResult> {
  const { workspaceId, repoSlug, path, content } = input;

  // 1. Detect doc type
  const docType = detectDocType(path);
  if (!docType) {
    return {
      status: 'failed',
      error: {
        code: 'format_validation_failed',
        message: `Path must start with decisions/, sessions/, or patterns/`,
      },
    };
  }

  // 2. Parse frontmatter + body + title
  let parsed;
  try {
    parsed = parseDocument(content);
  } catch (err) {
    return {
      status: 'failed',
      error: { code: 'format_validation_failed', message: (err as Error).message },
    };
  }

  // 3. Validate frontmatter against schema
  const validation = validateFrontmatter(docType, parsed.frontmatter);
  if (!validation.valid) {
    return {
      status: 'failed',
      error: {
        code: 'format_validation_failed',
        message: 'Frontmatter validation failed',
        details: validation.errors,
      },
    };
  }

  // 4. Validate filename
  const { slug, date } = extractSlugAndDate(docType, path);
  if (!SLUG_REGEX.test(slug) || slug.length > 80) {
    return {
      status: 'failed',
      error: { code: 'format_validation_failed', message: `Invalid slug: ${slug}` },
    };
  }
  if (date && !DATE_REGEX.test(date)) {
    return {
      status: 'failed',
      error: { code: 'format_validation_failed', message: `Invalid date prefix: ${date}` },
    };
  }

  // 5. Compute content hash
  const contentHash = computeContentHash(parsed.frontmatter, parsed.body);

  // 6. Upsert repo
  const existingRepos = await db.select().from(repos).where(and(eq(repos.workspaceId, workspaceId), eq(repos.slug, repoSlug)));
  let repo = existingRepos[0];
  if (!repo) {
    const inserted = await db.insert(repos).values({ workspaceId, slug: repoSlug }).returning();
    repo = inserted[0]!;
  }

  // 7. Check for existing document with matching content_hash (idempotency).
  // TODO(phase-9): two concurrent ingests at the same (repoId, path) can both
  // miss this SELECT and race to INSERT — second one fails on the
  // documents_repo_path_unique constraint with Postgres error 23505. The unique
  // index is the right defense; future work should catch the violation and
  // retry as an update so callers see a normal 200 instead of a 500.
  const existing = await db
    .select()
    .from(documents)
    .where(and(eq(documents.repoId, repo.id), eq(documents.path, path)));

  if (existing[0] && existing[0].contentHash === contentHash && !existing[0].deletedAt) {
    return { status: 'unchanged', document: existing[0] };
  }

  // 8. Chunk the body
  const newChunks = chunkMarkdown(parsed.body, { maxTokens: 512 });

  // 9. Embed chunks (one batch call)
  const embeddings = await embedding.embed(newChunks.map((c) => c.content));

  // 10. Atomic upsert + chunk replace
  const result = await db.transaction(async (tx) => {
    const status = (parsed.frontmatter as { status?: string }).status ?? 'active';
    let docRow: Document;

    if (existing[0]) {
      const updated = await tx
        .update(documents)
        .set({
          docType,
          slug,
          status,
          title: parsed.title,
          body: parsed.body,
          frontmatter: parsed.frontmatter as object,
          date,
          contentHash,
          deletedAt: null,
          updatedAt: new Date(),
        })
        .where(eq(documents.id, existing[0].id))
        .returning();
      docRow = updated[0]!;
    } else {
      const inserted = await tx
        .insert(documents)
        .values({
          workspaceId,
          repoId: repo.id,
          docType,
          path,
          slug,
          status,
          title: parsed.title,
          body: parsed.body,
          frontmatter: parsed.frontmatter as object,
          date,
          contentHash,
        })
        .returning();
      docRow = inserted[0]!;
    }

    // Replace chunks
    await tx.delete(chunks).where(eq(chunks.documentId, docRow.id));
    if (newChunks.length > 0) {
      await tx.insert(chunks).values(
        newChunks.map((c, i) => ({
          workspaceId,
          documentId: docRow.id,
          position: c.position,
          content: c.content,
          embedding: embeddings[i]!,
          headingPath: c.headingPath,
        }))
      );
    }

    return docRow;
  });

  return {
    status: existing[0] ? 'updated' : 'created',
    document: result,
    chunks: newChunks.length,
  };
}
