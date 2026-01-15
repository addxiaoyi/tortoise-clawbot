import type { HermesRuntimeApi } from "../hermes/runtime/types";
import type { TohelpOpenClawApi } from "./openclaw-api-types";

const DEFAULT_MS = 120_000;
const MIN_MS = 100;
const MAX_MS = 600_000;

/**
 * Wall-clock limit for `tohelp_invoke_skill` / `invokeSkillTool`.
 * Per-call override (`timeoutMs` tool arg) wins, then `pluginConfig.invokeTimeoutMs`, else 120s.
 */
export function resolveInvokeTimeoutMs(
  api: TohelpOpenClawApi | HermesRuntimeApi,
  override?: unknown,
): number {
  let raw = DEFAULT_MS;
  if (typeof override === "number" && Number.isFinite(override)) {
    raw = override;
  } else {
    const cfg = "pluginConfig" in api ? api.pluginConfig?.invokeTimeoutMs : api.config?.invokeTimeoutMs;
    if (typeof cfg === "number" && Number.isFinite(cfg)) {
      raw = cfg;
    }
  }
  return Math.min(MAX_MS, Math.max(MIN_MS, raw));
}
