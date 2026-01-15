import { resolveWorkspacePath } from "../utils/safe-path.js";
import type { TohelpOpenClawApi } from "./openclaw-api-types.js";
import type { InvokeSkillToolOptions } from "./skill-invoke.js";
import type { TohelpMemoryRuntime } from "./memory-policy.js";
import {
  collectHermesSkillSummaries,
  execHermesGatewayHealthProbe,
  execHermesInvokeSkill,
  execHermesListSkills,
  execHermesMemory,
  execHermesPing,
  execHermesResolveWorkspacePath,
  type HermesMemoryToolAction,
} from "../hermes/runtime/tools.js";

export type MemoryToolAction = HermesMemoryToolAction;

function asHermesApi(api: TohelpOpenClawApi) {
  return { ...api, config: api.pluginConfig ?? api.config };
}

export const collectSkillSummaries = collectHermesSkillSummaries;

export function execTohelpPing(message?: string | null) {
  return execHermesPing(message);
}

export function execTohelpListNewCoreSkills() {
  return execHermesListSkills();
}

export function execTohelpResolveWorkspacePath(
  api: TohelpOpenClawApi,
  rawPath: string,
) {
  return execHermesResolveWorkspacePath(asHermesApi(api), rawPath);
}

export async function execTohelpInvokeSkill(
  api: TohelpOpenClawApi,
  params: {
    skill: string;
    tool: string;
    args?: Record<string, unknown>;
    timeoutMs?: number;
  },
  options?: Pick<InvokeSkillToolOptions, "signal">,
) {
  return execHermesInvokeSkill(asHermesApi(api), params, options);
}

export async function execTohelpMemory(
  api: TohelpOpenClawApi,
  params: {
    action: MemoryToolAction;
    key?: string;
    value?: unknown;
  },
  options?: { runtime?: TohelpMemoryRuntime },
) {
  return execHermesMemory(asHermesApi(api), params, options);
}

export const execTohelpGatewayHealthProbe = execHermesGatewayHealthProbe;

export function createMcpTohelpApi(): TohelpOpenClawApi {
  const workspaceRoot = process.env.TOHELP_WORKSPACE_ROOT ?? process.cwd();
  let pluginConfig: Record<string, unknown> | undefined;
  const raw = process.env.TOHELP_PLUGIN_CONFIG_JSON ?? process.env.HERMES_CONFIG_JSON;
  if (raw) {
    try {
      pluginConfig = JSON.parse(raw) as Record<string, unknown>;
    } catch {
      pluginConfig = undefined;
    }
  }

  return {
    resolvePath: (input: string) => resolveWorkspacePath(workspaceRoot, input),
    registerTool: () => {
      throw new Error("registerTool is not used by the standalone Hermes MCP server");
    },
    logger: {
      debug: (m, ...a) => console.error("[hermes-mcp]", m, ...a),
      info: (m, ...a) => console.error("[hermes-mcp]", m, ...a),
      warn: (m, ...a) => console.error("[hermes-mcp]", m, ...a),
      error: (m, e) => console.error("[hermes-mcp]", m, e ?? ""),
    },
    pluginConfig,
    config: pluginConfig,
  };
}
