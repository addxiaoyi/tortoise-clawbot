import { createHermesPluginContext } from "../hermes/runtime/context";
import type { TohelpOpenClawApi } from "./openclaw-api-types";

export function createBridgePluginContext(
  api: TohelpOpenClawApi,
  skillId: string,
) {
  return createHermesPluginContext(
    { ...api, config: api.pluginConfig ?? api.config },
    skillId,
  );
}
