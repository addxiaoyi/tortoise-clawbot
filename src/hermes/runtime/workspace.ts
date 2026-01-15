import { resolveWorkspacePath } from "../../utils/safe-path.js";
import type { HermesRuntimeApi } from "./types";

export function createHermesRuntimeApi(options: {
  workspaceRoot?: string;
  config?: Record<string, unknown>;
  logger?: HermesRuntimeApi["logger"];
} = {}): HermesRuntimeApi {
  const workspaceRoot = options.workspaceRoot ?? process.env.TOHELP_WORKSPACE_ROOT ?? process.cwd();
  return {
    resolvePath: (input: string) => resolveWorkspacePath(workspaceRoot, input),
    logger: options.logger ?? {
      debug: (m, ...a) => console.error("[hermes]", m, ...a),
      info: (m, ...a) => console.error("[hermes]", m, ...a),
      warn: (m, ...a) => console.error("[hermes]", m, ...a),
      error: (m, e) => console.error("[hermes]", m, e ?? ""),
    },
    config: options.config,
  };
}

export function createHermesRuntimeApiFromEnv(): HermesRuntimeApi {
  let config: Record<string, unknown> | undefined;
  const raw = process.env.TOHELP_PLUGIN_CONFIG_JSON ?? process.env.HERMES_CONFIG_JSON;
  if (raw) {
    try {
      config = JSON.parse(raw) as Record<string, unknown>;
    } catch {
      config = undefined;
    }
  }
  return createHermesRuntimeApi({ config });
}
