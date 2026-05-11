import { pgTable, uuid, text, timestamp, uniqueIndex, index } from 'drizzle-orm/pg-core';
import { sql } from 'drizzle-orm';
import { workspaces } from './workspaces.js';

export const repos = pgTable(
  'repos',
  {
    id: uuid('id').primaryKey().default(sql`gen_random_uuid()`),
    workspaceId: uuid('workspace_id').notNull().references(() => workspaces.id, { onDelete: 'cascade' }),
    slug: text('slug').notNull(),
    remoteUrl: text('remote_url'),
    createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
  },
  (table) => ({
    workspaceSlugUnique: uniqueIndex('repos_workspace_slug_unique').on(table.workspaceId, table.slug),
    workspaceIdx: index('repos_workspace').on(table.workspaceId),
  })
);

export type Repo = typeof repos.$inferSelect;
export type NewRepo = typeof repos.$inferInsert;
