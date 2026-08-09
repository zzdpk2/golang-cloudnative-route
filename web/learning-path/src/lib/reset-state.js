/**
 * @typedef {{ affectedFiles: number, sharedWith: string[] }} GateResetMetadata
 * @typedef {{ gate: string, publishedRevision: string }} PendingReset
 * @typedef {{ revision?: string, verified: string[], optionalVerified?: string[] }} PublishedProgress
 */

/**
 * @param {unknown} candidate
 * @param {string[]} gateIds
 * @returns {candidate is Record<string, GateResetMetadata>}
 */
export function validResetMetadata(candidate, gateIds) {
  if (typeof candidate !== "object" || candidate === null || Array.isArray(candidate)) return false;
  const metadata = /** @type {Record<string, GateResetMetadata>} */ (candidate);
  const known = new Set(gateIds);
  return gateIds.every((gateId) => {
    const entry = metadata[gateId];
    return Number.isInteger(entry?.affectedFiles) && entry.affectedFiles > 0 &&
      Array.isArray(entry.sharedWith) &&
      new Set(entry.sharedWith).size === entry.sharedWith.length &&
      entry.sharedWith.every((id) => id !== gateId && known.has(id));
  });
}

/**
 * @param {unknown} candidate
 * @param {PublishedProgress} progress
 * @param {Set<string>} gateIds
 * @param {Set<string>} optionalGateIds
 * @returns {{ remaining: PendingReset[], cleared: string[] }}
 */
export function reconcilePendingResets(candidate, progress, gateIds, optionalGateIds) {
  if (!Array.isArray(candidate) || !progress.revision) return { remaining: [], cleared: [] };
  /** @type {PendingReset[]} */
  const remaining = [];
  /** @type {string[]} */
  const cleared = [];
  for (const item of candidate) {
    if (typeof item !== "object" || item === null) continue;
    const pending = /** @type {PendingReset} */ (item);
    if (!gateIds.has(pending.gate) || typeof pending.publishedRevision !== "string") continue;
    if (progress.revision === pending.publishedRevision) {
      remaining.push(pending);
      continue;
    }
    const published = optionalGateIds.has(pending.gate)
      ? (progress.optionalVerified ?? [])
      : progress.verified;
    if (published.includes(pending.gate)) remaining.push(pending);
    else cleared.push(pending.gate);
  }
  return { remaining, cleared };
}
