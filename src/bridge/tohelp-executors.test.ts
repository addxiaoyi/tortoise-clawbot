import { mkdtempSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  collectSkillSummaries,
  createMcpTohelpApi,
  execTohelpGatewayHealthProbe,
  execTohelpMemory,
  execTohelpResolveWorkspacePath,
} from "./tohelp-executors.js";
import { resetBridgeMemoryForTests } from "./bridge-memory.js";
import type { TohelpOpenClawApi } from "./openclaw-api-types.js";

function mockApi(): TohelpOpenClawApi {
  return {
    resolvePath: (p: string) => `/ws/${p}`,
    registerTool: vi.fn(),
    logger: {
      debug: vi.fn(),
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn(),
    },
  };
}

describe("tohelp-executors", () => {
  beforeEach(() => {
    resetBridgeMemoryForTests();
  });

  it("execTohelpResolveWorkspacePath uses api.resolvePath", () => {
    const out = execTohelpResolveWorkspacePath(mockApi(), "foo");
    expect(out).toEqual({ input: "foo", resolved: "/ws/foo" });
  });

  it("createMcpTohelpApi resolvePath blocks directory traversal", () => {
    const ws = mkdtempSync(path.join(tmpdir(), "tohelp-mcp-ws-"));
    const prev = process.env.TOHELP_WORKSPACE_ROOT;
    process.env.TOHELP_WORKSPACE_ROOT = ws;
    try {
      const api = createMcpTohelpApi();
      expect(() => api.resolvePath("../../outside")).toThrow("escapes");
      expect(api.resolvePath("sub/file.txt")).toBe(path.resolve(ws, "sub", "file.txt"));
    } finally {
      if (prev === undefined) {
        delete process.env.TOHELP_WORKSPACE_ROOT;
      } else {
        process.env.TOHELP_WORKSPACE_ROOT = prev;
      }
    }
  });

  it("execTohelpMemory get/set/list/delete/clear", async () => {
    const api = mockApi();
    await expect(
      execTohelpMemory(api, { action: "get", key: "a" }),
    ).resolves.toMatchObject({ found: false });
    await execTohelpMemory(api, { action: "set", key: "a", value: 1 });
    await expect(
      execTohelpMemory(api, { action: "get", key: "a" }),
    ).resolves.toMatchObject({ found: true, value: 1 });
    await expect(
      execTohelpMemory(api, { action: "list" }),
    ).resolves.toMatchObject({ keys: ["a"] });
    await expect(
      execTohelpMemory(api, { action: "delete", key: "a" }),
    ).resolves.toMatchObject({ removed: true });
    await execTohelpMemory(api, { action: "set", key: "b", value: 2 });
    await expect(
      execTohelpMemory(api, { action: "clear" }),
    ).resolves.toMatchObject({ ok: true, scopedToPrefix: false });
    await expect(
      execTohelpMemory(api, { action: "list" }),
    ).resolves.toMatchObject({ keys: [] });
  });

  it("execTohelpMemory rejects set when value exceeds TOHELP_MEMORY_MAX_VALUE_BYTES", async () => {
    vi.stubEnv("TOHELP_MEMORY_MAX_VALUE_BYTES", "8");
    try {
      const api = mockApi();
      await expect(
        execTohelpMemory(api, {
          action: "set",
          key: "k",
          value: "toolong",
        }),
      ).rejects.toThrow(/exceeds TOHELP_MEMORY_MAX_VALUE_BYTES/);
      await expect(
        execTohelpMemory(api, { action: "set", key: "k", value: "ok" }),
      ).resolves.toMatchObject({ ok: true });
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("execTohelpMemory TOHELP_MEMORY_KEY_PREFIX scopes list/clear", async () => {
    const api = mockApi();
    await execTohelpMemory(api, { action: "set", key: "global", value: 1 });
    vi.stubEnv("TOHELP_MEMORY_KEY_PREFIX", "ns:");
    try {
      await execTohelpMemory(api, { action: "set", key: "x", value: 2 });
      await expect(
        execTohelpMemory(api, { action: "list" }),
      ).resolves.toMatchObject({ keys: ["x"] });
      await expect(
        execTohelpMemory(api, { action: "clear" }),
      ).resolves.toMatchObject({ ok: true, scopedToPrefix: true });
      await expect(
        execTohelpMemory(api, { action: "list" }),
      ).resolves.toMatchObject({ keys: [] });
    } finally {
      vi.unstubAllEnvs();
    }
    await expect(
      execTohelpMemory(api, { action: "list" }),
    ).resolves.toMatchObject({ keys: ["global"] });
  });

  it("execTohelpMemory isolates keys per sessionId when TOHELP_MEMORY_AUTO_SESSION_PREFIX", async () => {
    vi.stubEnv("TOHELP_MEMORY_AUTO_SESSION_PREFIX", "1");
    try {
      const api = mockApi();
      const s1 = { sessionId: "11111111-1111-1111-1111-111111111111" };
      const s2 = { sessionId: "22222222-2222-2222-2222-222222222222" };
      await execTohelpMemory(
        api,
        { action: "set", key: "k", value: 1 },
        { runtime: s1 },
      );
      await expect(
        execTohelpMemory(api, { action: "get", key: "k" }, { runtime: s1 }),
      ).resolves.toMatchObject({ found: true, value: 1, sessionScoped: true });
      await expect(
        execTohelpMemory(api, { action: "get", key: "k" }, { runtime: s2 }),
      ).resolves.toMatchObject({ found: false, sessionScoped: true });
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("execTohelpMemory can isolate by sessionKey and writes can be enforced to require sessionKey", async () => {
    const api = mockApi();
    vi.stubEnv("TOHELP_MEMORY_AUTO_SESSION_PREFIX", "1");
    vi.stubEnv("TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE", "1");
    try {
      await expect(
        execTohelpMemory(api, { action: "set", key: "k", value: 1 }),
      ).rejects.toThrow(/sessionKey is required/);
      await expect(
        execTohelpMemory(
          api,
          { action: "set", key: "k", value: 1 },
          { runtime: { sessionKey: "sess-a" } },
        ),
      ).resolves.toMatchObject({ ok: true, sessionScoped: true });
      await expect(
        execTohelpMemory(
          api,
          { action: "get", key: "k" },
          { runtime: { sessionKey: "sess-a" } },
        ),
      ).resolves.toMatchObject({ found: true, value: 1, sessionScoped: true });
      await expect(
        execTohelpMemory(
          api,
          { action: "get", key: "k" },
          { runtime: { sessionKey: "sess-b" } },
        ),
      ).resolves.toMatchObject({ found: false, sessionScoped: true });
      await expect(
        execTohelpMemory(api, { action: "delete", key: "k" }),
      ).rejects.toThrow(/sessionKey is required/);
      await expect(
        execTohelpMemory(
          api,
          { action: "delete", key: "k" },
          { runtime: { sessionKey: "sess-a" } },
        ),
      ).resolves.toMatchObject({ removed: true });
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("execTohelpMemory can enforce sessionKey for reads", async () => {
    const api = mockApi();
    vi.stubEnv("TOHELP_MEMORY_AUTO_SESSION_PREFIX", "1");
    vi.stubEnv("TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ", "1");
    try {
      await execTohelpMemory(
        api,
        { action: "set", key: "k", value: 1 },
        { runtime: { sessionKey: "sess-a" } },
      );
      await expect(
        execTohelpMemory(api, { action: "get", key: "k" }),
      ).rejects.toThrow(/sessionKey is required for read actions/);
      await expect(
        execTohelpMemory(api, { action: "list" }),
      ).rejects.toThrow(/sessionKey is required for read actions/);
      await expect(
        execTohelpMemory(
          api,
          { action: "get", key: "k" },
          { runtime: { sessionKey: "sess-a" } },
        ),
      ).resolves.toMatchObject({ found: true, value: 1 });
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("execTohelpGatewayHealthProbe returns ok and health json", async () => {
    const out = await execTohelpGatewayHealthProbe();
    expect(out.ok).toBe(true);
    expect(out.health).toMatchObject({ status: "ok", version: "probe" });
  });

  it("collectSkillSummaries includes prerequisites for skills that need external setup", () => {
    const rows = collectSkillSummaries();
    expect(rows.find((r) => r.skill === "github")?.prerequisites).toContain(
      "gh",
    );
    expect(rows.find((r) => r.skill === "slack")?.prerequisites).toBeTruthy();
  });
});
