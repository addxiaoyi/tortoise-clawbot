import { MemoryPlugin } from "../plugins/new-core/memory/plugin.js";
import { createHermesPluginContext } from "../hermes/runtime/context.js";
import type { HermesRuntimeApi } from "../hermes/runtime/types.js";
import type { TohelpOpenClawApi } from "./openclaw-api-types.js";

let plugin: MemoryPlugin | null = null;
let ready: Promise<MemoryPlugin> | null = null;

/**
 * Lazily initializes a single {@link MemoryPlugin} for the OpenClaw bridge / MCP process
 * (in-memory key-value scoped to the process lifetime).
 */
export async function getOrInitBridgeMemory(
  api: TohelpOpenClawApi | HermesRuntimeApi,
): Promise<MemoryPlugin> {
  if (plugin) {
    return plugin;
  }
  if (!ready) {
    ready = (async () => {
      const p = new MemoryPlugin();
      const ctx = createHermesPluginContext(
        "pluginConfig" in api
          ? { ...api, config: api.pluginConfig ?? api.config }
          : api,
        "memory",
      );
      await p.onInit(ctx);
      if (p.onStart) {
        await p.onStart();
      }
      plugin = p;
      return p;
    })();
  }
  return ready;
}

/** Vitest-only: drop singleton between tests. */
export function resetBridgeMemoryForTests(): void {
  plugin = null;
  ready = null;
}
