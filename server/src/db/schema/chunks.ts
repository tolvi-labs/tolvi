import { pgTable, uuid, text, integer, customType, index, uniqueIndex } from 'drizzle-orm/pg-core';
import { sql } from 'drizzle-orm';
import { workspaces } from './workspaces.js';
import { documents } from './documents.js';

// Read EMBEDDING_DIM from env at module load with strict validation matching
// config.ts's zod schema. The dim is locked into the migration when generated
// AND used to construct vector literals at runtime — the two MUST agree, or
// every embedding insert silently corrupts. Task 7's db/client.ts adds a
// runtime check against the actual chunks.embedding column type to close the
// remaining drift window if EMBEDDING_DIM changes between migrate and run.
const EMBEDDING_DIM = (() => {
  const raw = process.env.EMBEDDING_DIM;
  if (raw == null || raw === '') return 768;
  const n = Number(raw);
  if (!Number.isInteger(n) || n <= 0) {
    throw new Error(
      `Invalid EMBEDDING_DIM: ${JSON.stringify(raw)} (expected positive integer)`
    );
  }
  return n;
})();

const vector = customType<{ data: number[]; driverData: string }>({
  dataType() {
    return `vector(${EMBEDDING_DIM})`;
  },
  toDriver(value: number[]): string {
    return `[${value.join(',')}]`;
  },
  // Relies on pgvector's `[a,b,c]` JSON-compatible text format.
  fromDriver(value: string): number[] {
    return JSON.parse(value);
  },
});

export const chunks = pgTable(
  'chunks',
  {
    id: uuid('id').primaryKey().default(sql`gen_random_uuid()`),
    workspaceId: uuid('workspace_id').notNull().references(() => workspaces.id),
    documentId: uuid('document_id').notNull().references(() => documents.id, { onDelete: 'cascade' }),
    position: integer('position').notNull(),
    content: text('content').notNull(),
    embedding: vector('embedding').notNull(),
    headingPath: text('heading_path').array(),
  },
  (table) => ({
    documentPositionUnique: uniqueIndex('chunks_document_position_unique').on(table.documentId, table.position),
    workspaceIdx: index('chunks_workspace').on(table.workspaceId),
  })
);

export type Chunk = typeof chunks.$inferSelect;
export type NewChunk = typeof chunks.$inferInsert;
