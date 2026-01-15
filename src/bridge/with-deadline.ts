/**
 * Runs `run()` under a wall-clock deadline and optional external {@link AbortSignal}.
 * The first of: success, rejection, timeout, or abort wins; timers and listeners are cleaned up.
 */
export async function withDeadline<T>(
  run: () => Promise<T>,
  deadlineMs: number,
  external?: AbortSignal,
): Promise<T> {
  if (deadlineMs <= 0) {
    throw new Error("[tohelp_invoke_skill] deadlineMs must be positive");
  }
  if (external?.aborted) {
    throw new Error("[tohelp_invoke_skill] aborted");
  }

  return new Promise<T>((resolve, reject) => {
    let settled = false;
    const timer = setTimeout(() => {
      if (settled) {
        return;
      }
      settled = true;
      external?.removeEventListener("abort", onAbort);
      reject(new Error(`[tohelp_invoke_skill] timeout after ${deadlineMs}ms`));
    }, deadlineMs);

    const onAbort = () => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timer);
      reject(new Error("[tohelp_invoke_skill] aborted"));
    };
    external?.addEventListener("abort", onAbort, { once: true });

    run()
      .then(
        (v) => {
          if (settled) {
            return;
          }
          settled = true;
          clearTimeout(timer);
          external?.removeEventListener("abort", onAbort);
          resolve(v);
        },
        (e) => {
          if (settled) {
            return;
          }
          settled = true;
          clearTimeout(timer);
          external?.removeEventListener("abort", onAbort);
          reject(e);
        },
      );
  });
}
