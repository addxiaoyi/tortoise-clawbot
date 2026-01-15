import { describe, it, expect, vi } from "vitest";
import { resolveInvokeTimeoutMs } from "./invoke-timeout";
import type { TohelpOpenClawApi } from "./openclaw-api-types";

function api(pluginConfig?: Record<string, unknown>): TohelpOpenClawApi {
  return {
    resolvePath: (p) => p,
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

describe("resolveInvokeTimeoutMs", () => {
  it("uses default 120s", () => {
    expect(resolveInvokeTimeoutMs(api())).toBe(120_000);
  });

  it("reads pluginConfig.invokeTimeoutMs", () => {
    expect(resolveInvokeTimeoutMs(api({ invokeTimeoutMs: 30_000 }))).toBe(30_000);
  });

  it("override wins and is clamped", () => {
    expect(resolveInvokeTimeoutMs(api({ invokeTimeoutMs: 30_000 }), 5_000)).toBe(5_000);
    expect(resolveInvokeTimeoutMs(api(), 50)).toBe(100);
    expect(resolveInvokeTimeoutMs(api(), 999_999)).toBe(600_000);
  });
});
