export { getOrInitBridgeMemory as getOrInitHermesMemory } from "../../bridge/bridge-memory.js";
export type { TohelpMemoryRuntime as HermesMemoryRuntime } from "../../bridge/memory-policy.js";
export {
  assertMemoryValueSizeWithinLimit,
  listLogicalMemoryKeys,
  loadMemoryMaxValueBytes,
  loadMemoryRequireSessionKeyForRead,
  loadMemoryRequireSessionKeyForWrite,
  physicalMemoryKey,
  resolveEffectiveMemoryPrefix,
} from "../../bridge/memory-policy.js";
