import { describe, it, expect, vi } from "vitest";
import { DocumentationService } from "./service.js";

describe("DocumentationService", () => {
  const logger = {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  };

  it("slugifies titles and rejects path-like keys on getDocContent", async () => {
    const svc = new DocumentationService({ format: "markdown" }, logger as any);
    await svc.start();
    const name = await svc.generateDoc("../../etc/passwd", "body");
    expect(name).not.toContain("/");
    expect(name).toMatch(/^[-a-z0-9]+\.markdown$/);

    await expect(svc.getDocContent("../evil.md")).rejects.toThrow("Invalid document filename");
    await expect(svc.getDocContent("safe-name.markdown")).resolves.toBeUndefined();
  });
});
