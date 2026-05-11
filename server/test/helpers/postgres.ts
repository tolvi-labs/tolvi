import { GenericContainer, type StartedTestContainer, Wait } from 'testcontainers';
import pg from 'pg';
import { drizzle } from 'drizzle-orm/node-postgres';
import { migrate } from 'drizzle-orm/node-postgres/migrator';
import * as schema from '../../src/db/schema/index.js';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

let container: StartedTestContainer | null = null;
let pool: pg.Pool | null = null;

export type TestDb = ReturnType<typeof drizzle<typeof schema>>;

export async function startTestDb(): Promise<{ db: TestDb; pool: pg.Pool; databaseUrl: string }> {
  if (!container) {
    container = await new GenericContainer('pgvector/pgvector:pg16')
      .withEnvironment({
        POSTGRES_USER: 'tolvi',
        POSTGRES_PASSWORD: 'tolvi',
        POSTGRES_DB: 'tolvi_test',
      })
      .withExposedPorts(5432)
      .withWaitStrategy(Wait.forLogMessage('database system is ready to accept connections', 2))
      .start();
  }

  const databaseUrl = `postgresql://tolvi:tolvi@${container.getHost()}:${container.getMappedPort(5432)}/tolvi_test`;
  pool = new pg.Pool({ connectionString: databaseUrl });
  const db = drizzle(pool, { schema });

  // Run migrations
  await migrate(db, {
    migrationsFolder: path.resolve(__dirname, '../../src/db/migrations'),
  });

  return { db, pool, databaseUrl };
}

export async function stopTestDb(): Promise<void> {
  if (pool) {
    await pool.end();
    pool = null;
  }
  if (container) {
    await container.stop();
    container = null;
  }
}

/**
 * Truncate all tables between tests for clean state.
 * Introspection-driven: discovers tables from information_schema so this stays
 * correct as the schema grows. Skips drizzle's __drizzle_migrations bookkeeping.
 */
export async function resetTestDb(pool: pg.Pool): Promise<void> {
  const result = await pool.query<{ tables: string }>(`
    SELECT string_agg(quote_ident(tablename), ', ') AS tables
    FROM pg_tables
    WHERE schemaname = 'public' AND tablename != '__drizzle_migrations'
  `);
  const tables = result.rows[0]?.tables;
  if (tables) {
    await pool.query(`TRUNCATE ${tables} RESTART IDENTITY CASCADE`);
  }
}
