import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod";
import { SKILL_IDS } from "../../bridge/skill-registry.js";
import {
  createHermesRuntimeApiFromEnv,
} from "../runtime/workspace.js";
import {
  execHermesGatewayHealthProbe,
  execHermesInvokeSkill,
  execHermesListSkills,
  execHermesMemory,
  execHermesPing,
  execHermesResolveWorkspacePath,
} from "../runtime/tools.js";

/** Stable compatibility tool ids exposed by Hermes MCP. */
export const HERMES_MCP_TOOL_NAMES = [
  "tohelp_ping",
  "tohelp_list_new_core_skills",
  "tohelp_resolve_workspace_path",
  "tohelp_invoke_skill",
  "tohelp_memory",
  "tohelp_gateway_health_probe",
] as const;

export type HermesMcpToolName = (typeof HERMES_MCP_TOOL_NAMES)[number];

const firstSkill = SKILL_IDS[0];
if (!firstSkill) {
  throw new Error("SKILL_IDS is empty");
}
const skillEnum = z.enum([firstSkill, ...SKILL_IDS.slice(1)] as [
  string,
  ...string[],
]);

function textResult(obj: unknown): CallToolResult {
  return {
    content: [{ type: "text", text: JSON.stringify(obj, null, 2) }],
  };
}

function errorResult(message: string): CallToolResult {
  return {
    isError: true,
    content: [{ type: "text", text: message }],
  };
}

export function registerHermesMcpTools(mcp: McpServer): void {
  const api = createHermesRuntimeApiFromEnv();

  mcp.registerTool(
    "tohelp_ping",
    {
      description:
        "Health check for the Hermes Agent MCP runtime. Returns JSON with node version and optional echo.",
      inputSchema: z.object({
        message: z.string().optional(),
      }),
    },
    async (args) => textResult(execHermesPing(args.message)),
  );

  mcp.registerTool(
    "tohelp_list_new_core_skills",
    {
      description:
        "Lists Hermes/new-core SkillPlugin definitions and tool names (no network).",
    },
    async () => textResult(execHermesListSkills()),
  );

  mcp.registerTool(
    "tohelp_resolve_workspace_path",
    {
      description:
        "解析路径到 TOHELP_WORKSPACE_ROOT（或当前工作目录）；禁止逃出工作区根目录（相对 ../ 与越界绝对路径均拒绝）。",
      inputSchema: z.object({
        path: z.string().min(1),
      }),
    },
    async (args) =>
      textResult(execHermesResolveWorkspacePath(api, args.path.trim())),
  );

  mcp.registerTool(
    "tohelp_invoke_skill",
    {
      description:
        "Run a Hermes/new-core skill tool. Secrets: set HERMES_CONFIG_JSON or TOHELP_PLUGIN_CONFIG_JSON.",
      inputSchema: z.object({
        skill: skillEnum,
        tool: z.string().min(1),
        args: z.record(z.string(), z.unknown()).optional(),
        timeoutMs: z.number().min(100).max(600_000).optional(),
      }),
    },
    async (args, extra) => {
      try {
        const out = await execHermesInvokeSkill(
          api,
          {
            skill: args.skill,
            tool: args.tool,
            args: args.args,
            timeoutMs: args.timeoutMs,
          },
          { signal: extra.signal },
        );
        return textResult(out);
      } catch (e) {
        return errorResult(e instanceof Error ? e.message : String(e));
      }
    },
  );

  mcp.registerTool(
    "tohelp_memory",
    {
      description:
        "Hermes in-process MemoryPlugin KV: get, set, list, delete, clear. " +
        "Env: TOHELP_MEMORY_KEY_PREFIX, TOHELP_MEMORY_MAX_VALUE_BYTES, TOHELP_MEMORY_AUTO_SESSION_PREFIX, " +
        "TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE, TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ.",
      inputSchema: z.object({
        action: z.enum(["get", "set", "list", "delete", "clear"]),
        key: z.string().optional(),
        value: z.unknown().optional(),
      }),
    },
    async (args) => {
      try {
        return textResult(
          await execHermesMemory(api, {
            action: args.action,
            key: args.key,
            value: args.value,
          }),
        );
      } catch (e) {
        return errorResult(e instanceof Error ? e.message : String(e));
      }
    },
  );

  mcp.registerTool(
    "tohelp_gateway_health_probe",
    {
      description:
        "Start Hermes GatewayPlugin on a free port, GET /health, then stop.",
    },
    async () => textResult(await execHermesGatewayHealthProbe()),
  );
}
