import { describe, it, expect } from "vitest";
import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const testDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(testDir, "../..");

describe("scripts/dev-doctor.mjs", () => {
  it("exits 0 for the Hermes-first layout", () => {
    const out = execFileSync(process.execPath, [path.join(root, "scripts", "dev-doctor.mjs")], {
      cwd: root,
      encoding: "utf-8",
    });
    expect(out).toContain("Hermes Agent dev layout looks good.");
    expect(out).toContain("tip openclaw-main: optional compatibility layer");
  });
});
