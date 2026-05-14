import { describe, it, expect } from 'vitest';
import {
  RECENCY_FLOOR, RECENCY_AMPLITUDE, RECENCY_HALF_LIFE_DAYS,
  SESSION_DOWN_WEIGHT, DEFAULT_SURFACED_STATUSES, ALL_STATUSES,
} from '../../src/search/ranking.js';

describe('ranking constants', () => {
  it('matches spec tolvi-format-v1 §9 numbers exactly', () => {
    expect(RECENCY_FLOOR).toBe(0.8);
    expect(RECENCY_AMPLITUDE).toBe(0.2);
    expect(RECENCY_HALF_LIFE_DAYS).toBe(180);
    expect(SESSION_DOWN_WEIGHT).toBe(0.7);
  });

  it('default surfaced statuses match spec §6 (excludes superseded/deprecated/draft)', () => {
    expect([...DEFAULT_SURFACED_STATUSES]).toEqual(['active', 'in-progress', 'historical']);
  });

  it('ALL_STATUSES has exactly the six values from spec §6', () => {
    expect(ALL_STATUSES).toHaveLength(6);
    expect([...ALL_STATUSES]).toEqual([
      'active', 'in-progress', 'superseded', 'deprecated', 'draft', 'historical',
    ]);
  });
});
