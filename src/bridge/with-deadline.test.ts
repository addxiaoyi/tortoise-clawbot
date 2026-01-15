import { describe, it, expect } from "vitest";
import { withDeadline } from "./with-deadline";

describe("withDeadline", () => {
  it("resolves when run finishes before deadline", async () => {
    await expect(
      withDeadline(async () => 42, 5000),
    ).resolves.toBe(42);
  });

  it("rejects on timeout", async () => {
    await expect(
      withDeadline(() => new Promise<number>(() => {}), 20),
    ).rejects.toThrow(/timeout after 20ms/);
  });

  it("rejects when external signal is already aborted", async () => {
    const ac = new AbortController();
    ac.abort();
    await expect(
      withDeadline(async () => 1, 5000, ac.signal),
    ).rejects.toThrow(/aborted/);
  });

  it("rejects when aborted during run", async () => {
    const ac = new AbortController();
    const p = withDeadline(async () => {
      await new Promise((r) => setTimeout(r, 500));
      return 1;
    }, 10_000, ac.signal);
    queueMicrotask(() => ac.abort());
    await expect(p).rejects.toThrow(/aborted/);
  });
});
