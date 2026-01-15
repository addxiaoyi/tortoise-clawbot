import { describe, it, expect } from "vitest";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import type { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";
import {
  registerTohelpMcpTools,
  TOHELP_MCP_TOOL_NAMES,
} from "./register-tohelp-tools.js";

describe("registerTohelpMcpTools", () => {
  it("exposes the canonical tool id list", () => {
    expect([...TOHELP_MCP_TOOL_NAMES].sort()).toEqual([
      "tohelp_gateway_health_probe",
      "tohelp_invoke_skill",
      "tohelp_list_new_core_skills",
      "tohelp_memory",
      "tohelp_ping",
      "tohelp_resolve_workspace_path",
    ]);
  });

  it("registers tools reachable via InMemoryTransport + Client", async () => {
    const [clientSide, serverSide] = InMemoryTransport.createLinkedPair();
    const mcp = new McpServer({ name: "tohelp-test", version: "0" });
    registerTohelpMcpTools(mcp);
    await mcp.connect(serverSide);

    const client = new Client(
      { name: "tohelp-test-client", version: "0" },
      { capabilities: {} },
    );
    await client.connect(clientSide);

    const listed = await client.listTools();
    const names = listed.tools.map((t) => t.name).sort();
    expect(names).toEqual([...TOHELP_MCP_TOOL_NAMES].sort());

    const ping = (await client.callTool({
      name: "tohelp_ping",
      arguments: { message: "ci" },
    })) as CallToolResult;
    const first = ping.content[0];
    expect(first && "type" in first && first.type === "text").toBe(true);
    const text =
      first && "text" in first && typeof first.text === "string"
        ? first.text
        : "";
    expect(JSON.parse(text)).toMatchObject({ ok: true, echo: "ci" });

    await client.close();
    await mcp.close();
  });
});
