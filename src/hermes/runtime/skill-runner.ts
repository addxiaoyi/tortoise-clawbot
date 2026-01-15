import { resolveInvokeTimeoutMs } from "../../bridge/invoke-timeout";
import {
  createSkillPlugin,
  isSkillId,
  SKILL_IDS,
  type SkillId,
} from "../../bridge/skill-registry";
import { withDeadline } from "../../bridge/with-deadline";
import { createHermesPluginContext } from "./context";
import type { HermesInvokeOptions, HermesRuntimeApi } from "./types";

/** Stable codes for log filtering / metrics (not HTTP status). */
export function classifyHermesInvokeError(err: unknown): string {
  if (!(err instanceof Error)) {
    return "UNKNOWN";
  }
  const m = err.message;
  if (m.includes("timeout after")) return "TIMEOUT";
  if (m.includes("aborted")) return "ABORTED";
  if (m.includes("Unknown skill")) return "UNKNOWN_SKILL";
  if (m.includes("Unknown tool")) return "UNKNOWN_TOOL";
  if (m.startsWith("[hermes_invoke_skill]") || m.startsWith("[tohelp_invoke_skill]")) {
    return "TOOL_EXEC_FAILED";
  }
  return "UNKNOWN";
}

/**
 * Runs `onInit` -> `onStart` -> skill tool `execute` -> `onStop` for a new-core skill.
 */
export async function invokeHermesSkillTool(
  api: HermesRuntimeApi,
  skill: string,
  tool: string,
  args: Record<string, unknown>,
  options?: HermesInvokeOptions,
): Promise<unknown> {
  if (!isSkillId(skill)) {
    throw new Error(
      `Unknown skill "${skill}". Valid: ${SKILL_IDS.join(", ")}`,
    );
  }

  const skillId = skill as SkillId;
  const toolName = tool.trim();
  if (!toolName) {
    throw new Error("tool is required");
  }

  const timeoutMs = resolveInvokeTimeoutMs(api, options?.timeoutMs);
  const argKeys = Object.keys(args).sort().join(",") || "(none)";
  const t0 = Date.now();
  api.logger.info(
    `[hermes] invoke start ${skillId}/${toolName} timeoutMs=${timeoutMs} argKeys=${argKeys}`,
  );

  try {
    const result = await withDeadline(
      () => invokeHermesSkillToolUnbounded(api, skillId, toolName, args),
      timeoutMs,
      options?.signal,
    );
    const durationMs = Date.now() - t0;
    api.logger.info(
      `[hermes] invoke done ${skillId}/${toolName} durationMs=${durationMs}`,
    );
    return result;
  } catch (err) {
    const durationMs = Date.now() - t0;
    const errorCode = classifyHermesInvokeError(err);
    api.logger.error(
      `[hermes] invoke error ${skillId}/${toolName} durationMs=${durationMs} errorCode=${errorCode}`,
      err instanceof Error ? err : undefined,
    );
    throw err;
  }
}

async function invokeHermesSkillToolUnbounded(
  api: HermesRuntimeApi,
  skill: SkillId,
  toolName: string,
  args: Record<string, unknown>,
): Promise<unknown> {
  const plugin = createSkillPlugin(skill);
  const ctx = createHermesPluginContext(api, skill);
  await plugin.onInit(ctx);

  try {
    if (plugin.onStart) {
      await plugin.onStart();
    }
    const def = plugin.getSkillDefinition();
    const entry = def.tools.find((t) => t.name === toolName);
    if (!entry) {
      throw new Error(
        `[hermes_invoke_skill] Unknown tool "${toolName}" for skill "${skill}". Known: ${def.tools.map((t) => t.name).join(", ")}`,
      );
    }
    try {
      return await entry.execute(args as Record<string, unknown>, ctx);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      throw new Error(`[hermes_invoke_skill] ${skill}/${toolName}: ${msg}`, {
        cause: err,
      });
    }
  } finally {
    if (plugin.onStop) {
      await plugin.onStop();
    }
  }
}
