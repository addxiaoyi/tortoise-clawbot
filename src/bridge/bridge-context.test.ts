import { describe, it, expect, vi } from "vitest";
import { createBridgePluginContext } from "./bridge-context";
import type { TohelpOpenClawApi } from "./openclaw-api-types";

function mockApi(
  pluginConfig?: Record<string, unknown>,
): TohelpOpenClawApi {
  return {
    resolvePath: (p: string) => p,
    registerTool: vi.fn(),
    logger: {
      debug: vi.fn(),
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn(),
    },
    pluginConfig,
  };
}

describe("createBridgePluginContext", () => {
  it("merges skills.<skillId> over top-level config for getConfig", () => {
    const ctx = createBridgePluginContext(
      mockApi({
        globalFlag: true,
        skills: {
          slack: { token: "skill-token", extra: 1 },
        },
      }),
      "slack",
    );
    const cfg = ctx.getConfig<{
      globalFlag?: boolean;
      token?: string;
      extra?: number;
      skills?: unknown;
    }>();
    expect(cfg.token).toBe("skill-token");
    expect(cfg.extra).toBe(1);
    expect(cfg.globalFlag).toBe(true);
    expect(cfg.skills).toBeUndefined();
  });

  it("uses top-level only when skill has no overlay", () => {
    const ctx = createBridgePluginContext(
      mockApi({
        token: "top-level",
        skills: {},
      }),
      "slack",
    );
    expect(ctx.getConfig<{ token?: string }>().token).toBe("top-level");
  });

  it("skill-specific keys override top-level keys", () => {
    const ctx = createBridgePluginContext(
      mockApi({
        token: "top",
        skills: {
          slack: { token: "from-skill" },
        },
      }),
      "slack",
    );
    expect(ctx.getConfig<{ token?: string }>().token).toBe("from-skill");
  });
});
