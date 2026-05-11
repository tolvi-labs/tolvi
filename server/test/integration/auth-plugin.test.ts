import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { startTestDb, stopTestDb, resetTestDb, type TestDb } from '../helpers/postgres.js';
import { workspaces, apiKeys } from '../../src/db/schema/index.js';
import { generateApiKey, hashApiKey, extractKeyPrefix } from '../../src/auth/api-key.js';
import { authPlugin } from '../../src/auth/plugin.js';
import Fastify, { type FastifyInstance } from 'fastify';
import { eq } from 'drizzle-orm';
import type pg from 'pg';

describe('auth plugin (integration)', () => {
  let db: TestDb;
  let pool: pg.Pool;

  beforeAll(async () => {
    const result = await startTestDb();
    db = result.db;
    pool = result.pool;
  });

  afterAll(async () => {
    await stopTestDb();
  });

  beforeEach(async () => {
    await resetTestDb(pool);
  });

  async function buildTestApp(): Promise<FastifyInstance> {
    const app = Fastify({ logger: false });
    app.decorate('db', db);
    await app.register(authPlugin);
    app.get('/v1/test', { preHandler: app.requireAuth }, async (req) => ({ workspaceId: req.workspaceId }));
    return app;
  }

  it('returns 401 when no Authorization header is present', async () => {
    const app = await buildTestApp();
    const res = await app.inject({ method: 'GET', url: '/v1/test' });
    expect(res.statusCode).toBe(401);
    expect(res.json().error.code).toBe('unauthorized');
  });

  it('returns 401 when bearer token is malformed', async () => {
    const app = await buildTestApp();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/test',
      headers: { authorization: 'Bearer not-a-real-key' },
    });
    expect(res.statusCode).toBe(401);
  });

  it('returns 401 when key is unknown to the system', async () => {
    const app = await buildTestApp();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/test',
      headers: { authorization: `Bearer ${generateApiKey()}` },
    });
    expect(res.statusCode).toBe(401);
  });

  it('sets request.workspaceId when key is valid', async () => {
    const [w] = await db.insert(workspaces).values({ slug: 'test', name: 'Test' }).returning();
    const key = generateApiKey();
    await db.insert(apiKeys).values({
      workspaceId: w!.id,
      keyHash: await hashApiKey(key),
      keyPrefix: extractKeyPrefix(key),
      name: 'test key',
    });

    const app = await buildTestApp();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/test',
      headers: { authorization: `Bearer ${key}` },
    });
    expect(res.statusCode).toBe(200);
    expect(res.json().workspaceId).toBe(w!.id);
  });

  it('returns 401 when key has been revoked', async () => {
    const [w] = await db.insert(workspaces).values({ slug: 'test', name: 'Test' }).returning();
    const key = generateApiKey();
    await db.insert(apiKeys).values({
      workspaceId: w!.id,
      keyHash: await hashApiKey(key),
      keyPrefix: extractKeyPrefix(key),
      name: 'revoked key',
      revokedAt: new Date(),
    });

    const app = await buildTestApp();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/test',
      headers: { authorization: `Bearer ${key}` },
    });
    expect(res.statusCode).toBe(401);
  });

  it('updates lastUsedAt after successful auth', async () => {
    const [w] = await db.insert(workspaces).values({ slug: 'test', name: 'Test' }).returning();
    const key = generateApiKey();
    const [inserted] = await db.insert(apiKeys).values({
      workspaceId: w!.id,
      keyHash: await hashApiKey(key),
      keyPrefix: extractKeyPrefix(key),
      name: 'tracking key',
    }).returning();
    expect(inserted!.lastUsedAt).toBeNull();

    const app = await buildTestApp();
    const before = Date.now();
    const res = await app.inject({
      method: 'GET',
      url: '/v1/test',
      headers: { authorization: `Bearer ${key}` },
    });
    expect(res.statusCode).toBe(200);

    // The lastUsedAt update is fire-and-forget; poll briefly for it to land.
    let updated: typeof inserted | undefined;
    for (let i = 0; i < 20; i++) {
      const rows = await db.select().from(apiKeys).where(eq(apiKeys.id, inserted!.id));
      if (rows[0]?.lastUsedAt != null) {
        updated = rows[0];
        break;
      }
      await new Promise((r) => setTimeout(r, 50));
    }
    expect(updated?.lastUsedAt).toBeInstanceOf(Date);
    expect(updated!.lastUsedAt!.getTime()).toBeGreaterThanOrEqual(before);
  });
});
