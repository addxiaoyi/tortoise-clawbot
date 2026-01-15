import type { OpenClawPluginApi } from "openclaw/plugin-sdk/core";
import { emptyPluginConfigSchema } from "openclaw/plugin-sdk/core";
import type { TohelpOpenClawApi } from "../../src/bridge/openclaw-api-types.js";
import { registerTohelpTools } from "../../src/bridge/tohelp-tools.js";

const plugin = {
  id: "tohelp-openclaw",
  name: "Tohelp OpenClaw bridge",
  description:
    "Registers Tohelp tools (`tohelp_*`) backed by `src/bridge/tohelp-tools.ts` and `src/plugins/new-core`.",
  configSchema: emptyPluginConfigSchema(),
  register(api: OpenClawPluginApi) {
    api.registerService({
      id: "tohelp-openclaw",
      async start(ctx) {
        ctx.logger.info(
          "tohelp-openclaw: loaded (tohelp_ping, tohelp_list_new_core_skills, tohelp_resolve_workspace_path, tohelp_invoke_skill, tohelp_memory, tohelp_gateway_health_probe).",
        );
      },
    });
    registerTohelpTools(api as TohelpOpenClawApi);
  },
};

export default plugin;
