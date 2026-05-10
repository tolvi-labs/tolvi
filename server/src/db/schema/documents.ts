import { pgTable, uuid, text, timestamp, jsonb, date, index, uniqueIndex } from 'drizzle-orm/pg-core';
import { sql } from 'drizzle-orm';
import { workspaces } from './workspaces.js';
import { repos } from './repos.js';

export const documents = pgTable(
  'documents',
  {
    id: uuid('id').primaryKey().default(sql`gen_random_uuid()`),
    workspaceId: uuid('workspace_id').notNull().references(() => workspaces.id),
    repoId: uuid('repo_id').notNull().references(() => repos.id, { onDelete: 'cascade' }),
    docType: text('doc_type').notNull(),                  // 'decision' | 'session' | 'pattern'
    path: text('path').notNull(),
    slug: text('slug').notNull(),
    status: text('status').notNull().default('active'),
    title: text('title').notNull(),
    body: text('body').notNull(),
    frontmatter: jsonb('frontmatter').notNull().default({}),
    date: date('date'),                                   // nullable: patterns
    contentHash: text('content_hash').notNull(),
    deletedAt: timestamp('deleted_at', { withTimezone: true }),
    createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow(),
  },
  (table) => ({
    repoPathUnique: uniqueIndex('documents_repo_path_unique').on(table.repoId, table.path),
    workspaceStatus: index('documents_workspace_status').on(table.workspaceId, table.status).where(sql`deleted_at IS NULL`),
    workspaceDocType: index('documents_workspace_doc_type').on(table.workspaceId, table.docType).where(sql`deleted_at IS NULL`),
    repoIdx: index('documents_repo').on(table.repoId).where(sql`deleted_at IS NULL`),
    slugIdx: index('documents_slug').on(table.workspaceId, table.slug),
  })
);

export type Document = typeof documents.$inferSelect;
export type NewDocument = typeof documents.$inferInsert;
