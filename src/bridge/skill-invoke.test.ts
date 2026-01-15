import { describe, it, expect, vi } from "vitest";
import { invokeSkillTool } from "./skill-invoke";
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

describe("invokeSkillTool", () => {
  it("runs debugging.inspect_variable without network", async () => {
    const api = mockApi({});
    const out = await invokeSkillTool(api, "debugging", "inspect_variable", {
      name: "foo",
    });
    expect(out).toMatchObject({
      result: expect.stringContaining("foo"),
    });
    expect(api.logger.info).toHaveBeenCalledWith(
      expect.stringContaining("[tohelp] invoke start debugging/inspect_variable"),
    );
    expect(api.logger.info).toHaveBeenCalledWith(
      expect.stringMatching(
        /\[tohelp\] invoke done debugging\/inspect_variable durationMs=\d+/,
      ),
    );
  });

  it("rejects unknown skill", async () => {
    const api = mockApi();
    await expect(
      invokeSkillTool(api, "nope", "x", {}),
    ).rejects.toThrow(/Unknown skill/);
  });

  it("rejects unknown tool", async () => {
    const api = mockApi();
    await expect(
      invokeSkillTool(api, "debugging", "not_a_real_tool", {}),
    ).rejects.toThrow(/Unknown tool/);
    expect(api.logger.error).toHaveBeenCalledWith(
      expect.stringMatching(
        /\[tohelp\] invoke error debugging\/not_a_real_tool durationMs=\d+ errorCode=UNKNOWN_TOOL/,
      ),
      expect.any(Error),
    );
  });
});
