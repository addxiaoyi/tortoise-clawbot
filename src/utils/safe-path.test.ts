import { mkdtempSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import { describe, expect, it } from "vitest";
import {
  resolvePathUnderRoot,
  resolveSafeRelativeFile,
  resolveUrlPathnameToSafePath,
  resolveWorkspacePath,
} from "./safe-path.js";

describe("safe-path", () => {
  const root = mkdtempSync(path.join(tmpdir(), "vitest-safe-path-"));

  it("resolvePathUnderRoot allows nested file", () => {
    const p = resolvePathUnderRoot(root, "a", "b.txt");
    expect(p).toBe(path.resolve(root, "a", "b.txt"));
  });

  it("resolvePathUnderRoot rejects traversal via segments", () => {
    expect(() => resolvePathUnderRoot(root, "..", "etc", "passwd")).toThrow(
      "Path escapes root directory",
    );
  });

  it("resolveSafeRelativeFile rejects absolute path", () => {
    expect(() => resolveSafeRelativeFile(root, "/etc/passwd")).toThrow(
      "absolute path not allowed",
    );
  });

  it("resolveSafeRelativeFile rejects traversal in string", () => {
    expect(() => resolveSafeRelativeFile(root, "../../outside")).toThrow(
      "Path escapes root directory",
    );
  });

  it("resolveUrlPathnameToSafePath rejects encoded traversal", () => {
    expect(() => resolveUrlPathnameToSafePath(root, "/%2e%2e/%2e%2e/etc/passwd")).toThrow(
      "Path escapes root directory",
    );
  });

  it("resolveUrlPathnameToSafePath allows normal file", () => {
    const p = resolveUrlPathnameToSafePath(root, "/foo/bar.html");
    expect(p).toBe(path.resolve(root, "foo", "bar.html"));
  });

  it("resolveWorkspacePath returns root for blank input", () => {
    const w = mkdtempSync(path.join(tmpdir(), "vitest-wsp-"));
    expect(resolveWorkspacePath(w, "   ")).toBe(path.resolve(w));
  });

  it("resolveWorkspacePath rejects relative traversal", () => {
    const w = mkdtempSync(path.join(tmpdir(), "vitest-wsp-"));
    expect(() => resolveWorkspacePath(w, "../../etc/passwd")).toThrow("escapes");
  });

  it("resolveWorkspacePath allows nested relative path", () => {
    const w = mkdtempSync(path.join(tmpdir(), "vitest-wsp-"));
    const r = resolveWorkspacePath(w, "src/a.ts");
    expect(r).toBe(path.resolve(w, "src", "a.ts"));
  });

  it("resolveWorkspacePath rejects absolute path outside workspace", () => {
    const w = mkdtempSync(path.join(tmpdir(), "vitest-wsp-"));
    const outside = mkdtempSync(path.join(tmpdir(), "vitest-out-"));
    expect(() => resolveWorkspacePath(w, outside)).toThrow("escapes");
  });

  it("resolveWorkspacePath allows absolute path inside workspace", () => {
    const w = mkdtempSync(path.join(tmpdir(), "vitest-wsp-"));
    const inner = path.resolve(w, "pkg", "x.js");
    expect(resolveWorkspacePath(w, inner)).toBe(inner);
  });
});
