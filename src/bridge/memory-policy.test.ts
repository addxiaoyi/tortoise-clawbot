import { describe, it, expect, vi } from "vitest";
import {
  assertMemoryValueSizeWithinLimit,
  listLogicalMemoryKeys,
  loadMemoryRequireSessionKeyForRead,
  loadMemoryRequireSessionKeyForWrite,
  physicalMemoryKey,
  resolveEffectiveMemoryPrefix,
  sanitizeSessionIdForKey,
} from "./memory-policy.js";

describe("memory-policy", () => {
  it("physicalMemoryKey prepends prefix when set", () => {
    expect(physicalMemoryKey("k", "")).toBe("k");
    expect(physicalMemoryKey("k", "ns:")).toBe("ns:k");
  });

  it("listLogicalMemoryKeys strips prefix and sorts", () => {
    expect(listLogicalMemoryKeys(["ns:b", "ns:a", "other"], "ns:")).toEqual([
      "a",
      "b",
    ]);
    expect(listLogicalMemoryKeys(["x", "y"], "")).toEqual(["x", "y"]);
  });

  it("assertMemoryValueSizeWithinLimit rejects oversized JSON", () => {
    expect(() =>
      assertMemoryValueSizeWithinLimit({ a: "hello" }, 5),
    ).toThrow(/exceeds TOHELP_MEMORY_MAX_VALUE_BYTES/);
    expect(() =>
      assertMemoryValueSizeWithinLimit("hi", 100),
    ).not.toThrow();
  });

  it("sanitizeSessionIdForKey allows UUID-like ids", () => {
    expect(sanitizeSessionIdForKey("550e8400-e29b-41d4-a716-446655440000")).toBe(
      "550e8400-e29b-41d4-a716-446655440000",
    );
    expect(sanitizeSessionIdForKey("a/b")).toBe("a_b");
  });

  it("resolveEffectiveMemoryPrefix adds sess segment when auto + sessionId", () => {
    vi.stubEnv("TOHELP_MEMORY_KEY_PREFIX", "app:");
    vi.stubEnv("TOHELP_MEMORY_AUTO_SESSION_PREFIX", "1");
    try {
      const r = resolveEffectiveMemoryPrefix({
        sessionId: "550e8400-e29b-41d4-a716-446655440000",
      });
      expect(r.sessionScoped).toBe(true);
      expect(r.effectivePrefix).toBe(
        "app:sess:550e8400-e29b-41d4-a716-446655440000:",
      );
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("resolveEffectiveMemoryPrefix prefers sessionKey over sessionId", () => {
    vi.stubEnv("TOHELP_MEMORY_KEY_PREFIX", "app:");
    vi.stubEnv("TOHELP_MEMORY_AUTO_SESSION_PREFIX", "1");
    try {
      const r = resolveEffectiveMemoryPrefix({
        sessionKey: "key-123",
        sessionId: "id-456",
      });
      expect(r.sessionScoped).toBe(true);
      expect(r.effectivePrefix).toBe("app:sess:key-123:");
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("resolveEffectiveMemoryPrefix skips session segment without sessionId", () => {
    vi.stubEnv("TOHELP_MEMORY_AUTO_SESSION_PREFIX", "true");
    try {
      const r = resolveEffectiveMemoryPrefix({});
      expect(r.sessionScoped).toBe(false);
      expect(r.sessionScopeSkipped).toBe(true);
      expect(r.effectivePrefix).toBe("");
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("loadMemoryRequireSessionKeyForWrite parses env flag", () => {
    vi.stubEnv("TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE", "1");
    try {
      expect(loadMemoryRequireSessionKeyForWrite()).toBe(true);
    } finally {
      vi.unstubAllEnvs();
    }
  });

  it("loadMemoryRequireSessionKeyForRead parses env flag", () => {
    vi.stubEnv("TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ", "true");
    try {
      expect(loadMemoryRequireSessionKeyForRead()).toBe(true);
    } finally {
      vi.unstubAllEnvs();
    }
  });
});
