/**
 * Score-formula constants matching tolvi-format-v1 §9 (RAG defaults).
 * Single edit point if the spec evolves to v1.x.
 */
export const RECENCY_FLOOR = 0.8;
export const RECENCY_AMPLITUDE = 0.2;
export const RECENCY_HALF_LIFE_DAYS = 180;
export const SESSION_DOWN_WEIGHT = 0.7;

/**
 * Default status filter applied when caller doesn't override.
 * Per spec §6: exclude superseded, deprecated, draft.
 */
export const DEFAULT_SURFACED_STATUSES = ['active', 'in-progress', 'historical'] as const;
export type SurfacedStatus = (typeof DEFAULT_SURFACED_STATUSES)[number];

export const ALL_STATUSES = [
  'active', 'in-progress', 'superseded', 'deprecated', 'draft', 'historical',
] as const;
