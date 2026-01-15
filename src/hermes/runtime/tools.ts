import http from "node:http";
import { GatewayPlugin } from "../../plugins/new-core/gateway/plugin.js";
import type { PluginContext } from "../../plugins/new-core/types.js";
import { reserveFreePort } from "../../bridge/free-port.js";
import { getOrInitHermesMemory } from "./memory.js";
import {
  assertMemoryValueSizeWithinLimit,
  listLogicalMemoryKeys,
  loadMemoryMaxValueBytes,
  loadMemoryRequireSessionKeyForRead,
  loadMemoryRequireSessionKeyForWrite,
  physicalMemoryKey,
  resolveEffectiveMemoryPrefix,
  type HermesMemoryRuntime,
} from "./memory.js";
import {
  createSkillPlugin,
  SKILL_IDS,
  type SkillId,
} from "../../bridge/skill-registry.js";
import { invokeHermesSkillTool } from "./skill-runner.js";
import type { HermesInvokeOptions, HermesRuntimeApi } from "./types.js";

export type HermesMemoryToolAction = "get" | "set" | "list" | "delete" | "clear";

/** 部署/调用前置条件，供技能列表与文档对齐。 */
const SKILL_PREREQUISITES: Partial<Record<SkillId, string>> = {
  github:
    "需要本机已安装并已认证的 GitHub CLI（`gh`）；与 runtime config 里的键无必然对应关系。",
  slack:
    "需要在 Hermes/OpenClaw/MCP 配置中提供 Slack Bot Token（如 config.skills.slack）。",
  notion:
    "需要在配置中提供 Notion API 密钥（如 skills.notion / NOTION_API_KEY）。",
};

export function collectHermesSkillSummaries(): Array<{
  skill: string;
  version: string;
  description: string;
  prerequisites?: string;
  tools: Array<{ name: string; description: string }>;
}> {
  return SKILL_IDS.map((id) => {
    const d = createSkillPlugin(id).getSkillDefinition();
    const prerequisites = SKILL_PREREQUISITES[id];
    return {
      skill: d.name,
      version: d.version,
      description: d.description,
      ...(prerequisites ? { prerequisites } : {}),
      tools: d.tools.map((t) => ({ name: t.name, description: t.description })),
    };
  });
}

export function execHermesPing(message?: string | null) {
  const echo =
    typeof message === "string" && message.trim() ? message.trim() : null;
  return {
    ok: true,
    runtime: "hermes-agent",
    bridge: "tohelp-compat",
    echo,
    node: process.version,
  };
}

export function execHermesListSkills() {
  return {
    runtime: "hermes-agent",
    skills: collectHermesSkillSummaries(),
    lifecyclePlugins: ["memory", "gateway"],
  };
}

export function execHermesResolveWorkspacePath(
  api: HermesRuntimeApi,
  rawPath: string,
) {
  const resolved = api.resolvePath(rawPath);
  return { input: rawPath, resolved };
}

export async function execHermesInvokeSkill(
  api: HermesRuntimeApi,
  params: {
    skill: string;
    tool: string;
    args?: Record<string, unknown>;
    timeoutMs?: number;
  },
  options?: Pick<HermesInvokeOptions, "signal">,
) {
  const skill = typeof params.skill === "string" ? params.skill : "";
  const tool = typeof params.tool === "string" ? params.tool : "";
  const args =
    params.args !== null &&
    typeof params.args === "object" &&
    !Array.isArray(params.args)
      ? params.args
      : {};
  const timeoutOverride =
    typeof params.timeoutMs === "number" && Number.isFinite(params.timeoutMs)
      ? params.timeoutMs
      : undefined;
  const result = await invokeHermesSkillTool(api, skill, tool, args, {
    signal: options?.signal,
    timeoutMs: timeoutOverride,
  });
  return { skill, tool, result };
}

export async function execHermesMemory(
  api: HermesRuntimeApi,
  params: {
    action: HermesMemoryToolAction;
    key?: string;
    value?: unknown;
  },
  options?: { runtime?: HermesMemoryRuntime },
) {
  const mem = await getOrInitHermesMemory(api);
  const { action } = params;
  const { effectivePrefix: keyPrefix, sessionScoped, sessionScopeSkipped } =
    resolveEffectiveMemoryPrefix(options?.runtime);
  const maxValueBytes = loadMemoryMaxValueBytes();
  const requireSessionKeyForRead = loadMemoryRequireSessionKeyForRead();
  const requireSessionKeyForWrite = loadMemoryRequireSessionKeyForWrite();
  const sessionKey = options?.runtime?.sessionKey?.trim() ?? "";
  const isReadAction = action === "get" || action === "list";
  const isWriteAction = action === "set" || action === "delete" || action === "clear";
  if (requireSessionKeyForRead && isReadAction && !sessionKey) {
    throw new Error(
      "sessionKey is required for read actions when TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ=1",
    );
  }
  if (requireSessionKeyForWrite && isWriteAction && !sessionKey) {
    throw new Error(
      "sessionKey is required for write actions when TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE=1",
    );
  }

  const scopeMeta = {
    ...(sessionScoped ? { sessionScoped: true as const } : {}),
    ...(sessionScopeSkipped ? { sessionScopeSkipped: true as const } : {}),
  };

  switch (action) {
    case "get": {
      const key = typeof params.key === "string" ? params.key.trim() : "";
      if (!key) throw new Error("key is required for get");
      const pkey = physicalMemoryKey(key, keyPrefix);
      const value = mem.get(pkey);
      return {
        action: "get",
        key,
        found: value !== undefined,
        value,
        ...scopeMeta,
      };
    }
    case "set": {
      const key = typeof params.key === "string" ? params.key.trim() : "";
      if (!key) throw new Error("key is required for set");
      assertMemoryValueSizeWithinLimit(params.value, maxValueBytes);
      const pkey = physicalMemoryKey(key, keyPrefix);
      mem.set(pkey, params.value);
      return { action: "set", key, ok: true, ...scopeMeta };
    }
    case "list":
      return {
        action: "list",
        keys: listLogicalMemoryKeys(mem.listKeys(), keyPrefix),
        ...scopeMeta,
      };
    case "delete": {
      const key = typeof params.key === "string" ? params.key.trim() : "";
      if (!key) throw new Error("key is required for delete");
      const pkey = physicalMemoryKey(key, keyPrefix);
      const removed = mem.removeKey(pkey);
      return { action: "delete", key, removed, ...scopeMeta };
    }
    case "clear":
      if (keyPrefix) {
        for (const k of [...mem.listKeys()]) {
          if (k.startsWith(keyPrefix)) {
            mem.removeKey(k);
          }
        }
        return {
          action: "clear",
          ok: true,
          scopedToPrefix: true,
          ...scopeMeta,
        };
      }
      mem.clearAll();
      return {
        action: "clear",
        ok: true,
        scopedToPrefix: false,
        ...scopeMeta,
      };
    default: {
      const _exhaustive: never = action;
      throw new Error(`Unknown action: ${_exhaustive}`);
    }
  }
}

function fetchGatewayHealthJson(port: number): Promise<unknown> {
  return new Promise((resolve, reject) => {
    http
      .get(`http://127.0.0.1:${port}/health`, (res) => {
        let data = "";
        res.on("data", (c) => {
          data += c;
        });
        res.on("end", () => {
          try {
            resolve(JSON.parse(data));
          } catch {
            resolve(data);
          }
        });
      })
      .on("error", reject);
  });
}

/**
 * Starts GatewayPlugin on a free localhost port, GETs `/health`, then stops.
 * Useful for verifying the Hermes lifecycle plugin without a long-running gateway.
 */
export async function execHermesGatewayHealthProbe(): Promise<{
  ok: boolean;
  port: number;
  health: unknown;
  error?: string;
}> {
  const port = await reserveFreePort();
  const plugin = new GatewayPlugin();
  const mockContext = {
    meta: {
      id: "hermes-gateway-probe",
      name: "Hermes gateway probe",
      version: "probe",
    },
    logger: {
      info: () => {},
      warn: () => {},
      debug: () => {},
      error: () => {},
    },
    storage: {
      async getItem() {
        return null;
      },
      async setItem() {},
      async removeItem() {},
      async clear() {},
    },
    events: {
      emit: () => {},
      on: () => () => {},
      off: () => {},
    },
    getConfig: () => ({ port, host: "127.0.0.1" as const }),
  } as unknown as PluginContext;

  await plugin.onInit(mockContext);
  try {
    await plugin.onStart();
    const health = await fetchGatewayHealthJson(port);
    await plugin.onStop();
    return { ok: true, port, health };
  } catch (e) {
    await plugin.onStop().catch(() => {});
    return {
      ok: false,
      port,
      health: null,
      error: e instanceof Error ? e.message : String(e),
    };
  }
}
