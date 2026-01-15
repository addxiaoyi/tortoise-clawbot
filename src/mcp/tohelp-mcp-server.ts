import "dotenv/config";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { registerTohelpMcpTools } from "./register-tohelp-tools.js";

async function main(): Promise<void> {
  const mcp = new McpServer(
    { name: "tohelp-mcp", version: "0.1.0" },
    {
      instructions:
        "Hermes Agent MCP: new-core skills, MemoryPlugin KV, and GatewayPlugin health probe. " +
        "Set TOHELP_WORKSPACE_ROOT and optional HERMES_CONFIG_JSON or TOHELP_PLUGIN_CONFIG_JSON for invoke_skill. " +
        "Optional hardening: TOHELP_MEMORY_KEY_PREFIX, TOHELP_MEMORY_MAX_VALUE_BYTES, TOHELP_MEMORY_AUTO_SESSION_PREFIX, " +
        "TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE, TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ.",
    },
  );
  registerTohelpMcpTools(mcp);
  const transport = new StdioServerTransport();
  await mcp.connect(transport);
}

main().catch((err) => {
  console.error("[tohelp-mcp] fatal", err);
  process.exit(1);
});
