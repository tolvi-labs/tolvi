import { drizzle } from 'drizzle-orm/node-postgres';
import pg from 'pg';
import * as schema from './schema/index.js';

export type Db = ReturnType<typeof drizzle<typeof schema>>;

export function createDb(databaseUrl: string): { db: Db; pool: pg.Pool } {
  const pool = new pg.Pool({ connectionString: databaseUrl });
  const db = drizzle(pool, { schema });
  return { db, pool };
}

const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/**
 * Tenant-scoped database operations. ALL multi-tenant queries should go
 * through this helper to make workspace_id propagation explicit at call sites.
 *
 * Pattern: const result = await withWorkspace(db, workspaceId, async (txDb) => { ... });
 *
 * **v1 enforcement caveat:** this helper provides NO automatic predicate
 * injection. Every query inside `fn` MUST include `eq(table.workspaceId, workspaceId)`
 * in its WHERE clause; the helper merely validates the workspaceId shape and
 * documents the intent. Future Phase 9 may upgrade this to set a Postgres GUC
 * for RLS, at which point the in-callback discipline becomes defense-in-depth.
 *
 * The UUID-shape validation here catches an entire class of bugs where an
 * unwrapped `req.workspaceId` (undefined / empty / wrong type) silently
 * exposes other tenants' data when the SQL predicate degrades to `IS NULL`.
 */
export async function withWorkspace<T>(
  db: Db,
  workspaceId: string,
  fn: (txDb: Db, workspaceId: string) => Promise<T>
): Promise<T> {
  if (typeof workspaceId !== 'string' || !UUID_REGEX.test(workspaceId)) {
    throw new Error(
      `withWorkspace: invalid workspaceId (expected UUID, got ${JSON.stringify(workspaceId)})`
    );
  }
  return fn(db, workspaceId);
}
