import { hash, verify, Algorithm } from '@node-rs/argon2';
import { randomBytes } from 'node:crypto';

const KEY_PREFIX = 'tlv_';
const RANDOM_BYTES = 32;          // 32 bytes → 43 base64url chars
const PREFIX_DISPLAY_LEN = 8;     // first 8 chars after `tlv_` for indexed lookup

/**
 * Generate a new API key. Format: `tlv_` + 32 random bytes (base64url, no padding).
 * Total length ~47 chars.
 */
export function generateApiKey(): string {
  const random = randomBytes(RANDOM_BYTES).toString('base64url');
  return `${KEY_PREFIX}${random}`;
}

/**
 * Extract the 8-character prefix used for indexed key lookup.
 * Throws if the input doesn't look like a Tolvi API key.
 */
export function extractKeyPrefix(key: string): string {
  if (!key.startsWith(KEY_PREFIX)) {
    throw new Error(`Invalid API key: missing ${KEY_PREFIX} prefix`);
  }
  const afterPrefix = key.slice(KEY_PREFIX.length);
  if (afterPrefix.length < PREFIX_DISPLAY_LEN) {
    throw new Error(`Invalid API key: too short for prefix extraction`);
  }
  return afterPrefix.slice(0, PREFIX_DISPLAY_LEN);
}

/**
 * Hash an API key with argon2id for storage.
 */
export async function hashApiKey(key: string): Promise<string> {
  return hash(key, {
    algorithm: Algorithm.Argon2id,
    memoryCost: 19456,     // 19 MiB — OWASP 2024 recommendation for argon2id
    timeCost: 2,
    parallelism: 1,
  });
}

/**
 * Verify a plaintext key against its argon2id hash.
 * Returns false on mismatch; throws only on hash format errors.
 */
export async function verifyApiKey(key: string, keyHash: string): Promise<boolean> {
  try {
    return await verify(keyHash, key);
  } catch {
    return false;
  }
}
