import { Type } from "@sinclair/typebox";
import type { TohelpOpenClawApi } from "./openclaw-api-types";
import {
  execTohelpGatewayHealthProbe,
  execTohelpInvokeSkill,
  execTohelpListNewCoreSkills,
  execTohelpMemory,
  execTohelpPing,
  execTohelpResolveWorkspacePath,
} from "./tohelp-executors";
import { SKILL_IDS } from "./skill-registry";
import type { TohelpMemoryRuntime } from "./memory-policy.js";

export type { TohelpOpenClawApi } from "./openclaw-api-types";

type AgentToolResult = {
  content: Array<{ type: string; text: string }>;
  details?: unknown;
};

function json(payload: unknown): AgentToolResult {
  return {
    content: [{ type: "text", text: JSON.stringify(payload, null, 2) }],
    details: payload,
  };
}

function stringEnum<T extends readonly string[]>(
  values: T,
  options: { description?: string } = {},
) {
  return Type.Unsafe<T[number]>({
    type: "string",
    enum: [...values],
    ...options,
  });
}

/**
 * Registers Tohelp / new-core related tools on the OpenClaw gateway.
 */
export function registerTohelpTools(api: TohelpOpenClawApi): void {
  api.registerTool(
    {
      name: "tohelp_ping",
      label: "Tohelp ping",
      description:
        "Health check for the Tohelp OpenClaw bridge. Returns a short JSON status payload.",
      parameters: Type.Object(
        {
          message: Type.Optional(
            Type.String({ description: "Optional echo string for debugging." }),
          ),
        },
        { additionalProperties: false },
      ),
      async execute(_id: string, params: Record<string, unknown>) {
        const message =
          typeof params.message === "string" && params.message.trim()
            ? params.message.trim()
            : undefined;
        return json(execTohelpPing(message));
      },
    },
    { optional: true },
  );

  api.registerTool(
    {
      name: "tohelp_list_new_core_skills",
      label: "Tohelp new-core skills",
      description:
        "Lists skill modules defined under `src/plugins/new-core/*` (SkillPlugin.getSkillDefinition), including tool names and descriptions. Does not call external APIs.",
      parameters: Type.Object({}, { additionalProperties: false }),
      async execute() {
        return json(execTohelpListNewCoreSkills());
      },
    },
    { optional: true },
  );

  api.registerTool(
    {
      name: "tohelp_resolve_workspace_path",
      label: "Tohelp resolve path",
      description:
        "Resolves a path relative to the OpenClaw workspace using the gateway `resolvePath` helper (same rules as other OpenClaw tools).",
      parameters: Type.Object(
        {
          path: Type.String({
            description: "Relative or absolute path input to resolve against the workspace.",
          }),
        },
        { additionalProperties: false },
      ),
      async execute(_id: string, params: Record<string, unknown>) {
        const raw = typeof params.path === "string" ? params.path.trim() : "";
        if (!raw) {
          throw new Error("path is required");
        }
        return json(execTohelpResolveWorkspacePath(api, raw));
      },
    },
    { optional: true },
  );

  api.registerTool(
    {
      name: "tohelp_invoke_skill",
      label: "Tohelp invoke new-core skill",
      description:
        "Runs a tool from `src/plugins/new-core` skills: initializes the skill plugin, calls `onStart`, executes the named tool with `args`, then `onStop`. " +
        "Use `tohelp_list_new_core_skills` to discover `skill` / `tool` names. " +
        "Secrets (e.g. Slack token) should live in OpenClaw plugin config under `skills.<skillId>` — see repo README. " +
        "Default wall-clock limit 120s; override with `timeoutMs` or `plugins.entries.tohelp-openclaw.config.invokeTimeoutMs`.",
      parameters: Type.Object(
        {
          skill: stringEnum(SKILL_IDS, {
            description: "Skill id (matches SkillDefinition.name).",
          }),
          tool: Type.String({
            description: "Tool name inside that skill (e.g. list_issues, send_message).",
          }),
          args: Type.Optional(
            Type.Record(Type.String(), Type.Unknown(), {
              description: "Arguments object passed to the skill tool execute handler.",
            }),
          ),
          timeoutMs: Type.Optional(
            Type.Number({
              minimum: 100,
              maximum: 600_000,
              description:
                "Optional wall-clock limit in ms (100–600000). Overrides plugin config for this call.",
            }),
          ),
        },
        { additionalProperties: false },
      ),
      async execute(
        _id: string,
        params: Record<string, unknown>,
        signal?: AbortSignal,
      ) {
        const skill = typeof params.skill === "string" ? params.skill : "";
        const tool = typeof params.tool === "string" ? params.tool : "";
        const args =
          params.args !== null &&
          typeof params.args === "object" &&
          !Array.isArray(params.args)
            ? (params.args as Record<string, unknown>)
            : {};
        const timeoutOverride =
          typeof params.timeoutMs === "number" && Number.isFinite(params.timeoutMs)
            ? params.timeoutMs
            : undefined;
        const result = await execTohelpInvokeSkill(
          api,
          { skill, tool, args, timeoutMs: timeoutOverride },
          { signal },
        );
        return json(result);
      },
    },
    { optional: true },
  );

  /** OpenClaw 在 `resolvePluginTools` 时传入；可读取 `sessionKey/sessionId` 参与前缀隔离。 */
  api.registerTool(
    (toolCtx: TohelpMemoryRuntime & Record<string, unknown>) => ({
      name: "tohelp_memory",
      label: "Tohelp MemoryPlugin KV",
      description:
        "In-process key-value store backed by `MemoryPlugin` (same lifecycle as the bridge). " +
        "Actions: get, set, list, delete, clear. For `get`/`set`/`delete`, `key` is required. " +
        "Optional env: `TOHELP_MEMORY_KEY_PREFIX`, `TOHELP_MEMORY_MAX_VALUE_BYTES`, " +
        "`TOHELP_MEMORY_AUTO_SESSION_PREFIX=1` (OpenClaw: per-session sub-prefix `sess:<scopeId>:`; `sessionKey` 优先于 `sessionId`), " +
        "`TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE=1` (set/delete/clear 必须有 sessionKey), " +
        "`TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ=1` (get/list 也必须有 sessionKey). Not for secrets.",
      parameters: Type.Object(
        {
          action: stringEnum(
            ["get", "set", "list", "delete", "clear"] as const,
            { description: "Memory operation." },
          ),
          key: Type.Optional(
            Type.String({ description: "Key (required for get, set, delete)." }),
          ),
          value: Type.Optional(
            Type.Unknown({ description: "Value for set (any JSON-serializable shape)." }),
          ),
        },
        { additionalProperties: false },
      ),
      async execute(_id: string, params: Record<string, unknown>) {
        const action = params.action as
          | "get"
          | "set"
          | "list"
          | "delete"
          | "clear";
        const key = typeof params.key === "string" ? params.key : undefined;
        const value = params.value;
        const out = await execTohelpMemory(api, { action, key, value }, {
          runtime: toolCtx,
        });
        return json(out);
      },
    }),
    { optional: true },
  );

  api.registerTool(
    {
      name: "tohelp_gateway_health_probe",
      label: "Tohelp GatewayPlugin /health probe",
      description:
        "Starts `GatewayPlugin` on a random localhost port, requests GET /health, then stops. " +
        "Use to verify the gateway lifecycle plugin without keeping a server running.",
      parameters: Type.Object({}, { additionalProperties: false }),
      async execute() {
        return json(await execTohelpGatewayHealthProbe());
      },
    },
    { optional: true },
  );
}
