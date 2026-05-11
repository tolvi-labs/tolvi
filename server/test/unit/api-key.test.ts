import { describe, it, expect } from 'vitest';
import { generateApiKey, hashApiKey, verifyApiKey, extractKeyPrefix } from '../../src/auth/api-key.js';

describe('api-key', () => {
  it('generates a key with the tlv_ prefix and ~46 chars total', () => {
    const key = generateApiKey();
    expect(key).toMatch(/^tlv_[A-Za-z0-9_-]{40,}$/);
    expect(key.length).toBeGreaterThan(40);
    expect(key.length).toBeLessThan(60);
  });

  it('extracts the 8-char prefix from a generated key', () => {
    const key = 'tlv_abcdefgh1234567890XYZ';
    expect(extractKeyPrefix(key)).toBe('abcdefgh');
  });

  it('throws when extracting prefix from a malformed key', () => {
    expect(() => extractKeyPrefix('not-a-key')).toThrow(/prefix/i);
    expect(() => extractKeyPrefix('tlv_abc')).toThrow(/prefix/i);  // too short after prefix
  });

  it('hashes a key and verifies it round-trip', async () => {
    const key = generateApiKey();
    const hash = await hashApiKey(key);
    expect(hash).not.toBe(key);
    expect(hash).toMatch(/^\$argon2id\$/);
    expect(await verifyApiKey(key, hash)).toBe(true);
  });

  it('rejects a different key against the same hash', async () => {
    const key1 = generateApiKey();
    const key2 = generateApiKey();
    const hash = await hashApiKey(key1);
    expect(await verifyApiKey(key2, hash)).toBe(false);
  });
});
