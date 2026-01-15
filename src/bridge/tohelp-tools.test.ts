import { describe, it, expect, vi } from "vitest";
import { registerTohelpTools } from "./tohelp-tools";

describe("registerTohelpTools", () => {
  it("registers optional Tohelp tools including invoke", () => {
    const registerTool = vi.fn();
    const api = {
      resolvePath: (p: string) => `/mock/${p}`,
      registerTool,
      logger: {
        debug: vi.fn(),
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
      },
    };
    registerTohelpTools(api);
    expect(registerTool).toHaveBeenCalledTimes(6);
    const names = registerTool.mock.calls.map((c) => {
      const first = c[0] as { name: string } | ((ctx: object) => { name: string });
      if (typeof first === "function") {
        return first({}).name;
      }
      return first.name;
    });
    expect(names).toEqual([
      "tohelp_ping",
      "tohelp_list_new_core_skills",
      "tohelp_resolve_workspace_path",
      "tohelp_invoke_skill",
      "tohelp_memory",
      "tohelp_gateway_health_probe",
    ]);
    expect(registerTool.mock.calls.every((c) => c[1]?.optional === true)).toBe(
      true,
    );
  });
});
