import type { PluginConfig, PluginContext, PluginLogger } from "../../plugins/new-core/types";
import type { HermesRuntimeApi, HermesRuntimeLogger } from "./types";

function adaptHermesLogger(log: HermesRuntimeLogger): PluginLogger {
  return {
    debug: (message: string, ...args: unknown[]) => log.debug?.(message, ...args),
    info: (message: string, ...args: unknown[]) => log.info(message, ...args),
    warn: (message: string, ...args: unknown[]) => log.warn(message, ...args),
    error: (message: string, error?: Error, ...args: unknown[]) =>
      log.error(message, error, ...args),
  };
}

/**
 * Builds a PluginContext for running new-core skill tools under Hermes.
 *
 * `config.skills[skillId]` is merged over top-level keys (excluding `skills`) so
 * each skill can read flat fields like `token` via `getConfig()`.
 */
export function createHermesPluginContext(
  api: HermesRuntimeApi,
  skillId: string,
): PluginContext {
  const store = new Map<string, string>();
  const listeners = new Map<string, Set<(payload: unknown) => void>>();

  const getConfig = <T extends PluginConfig>(): T => {
    const root = (api.config ?? {}) as Record<string, unknown>;
    const { skills: skillsRaw, ...restRoot } = root;
    let skillOverlay: Record<string, unknown> = {};
    if (
      skillsRaw !== null &&
      typeof skillsRaw === "object" &&
      !Array.isArray(skillsRaw) &&
      skillId in skillsRaw
    ) {
      const inner = (skillsRaw as Record<string, unknown>)[skillId];
      if (inner !== null && typeof inner === "object" && !Array.isArray(inner)) {
        skillOverlay = inner as Record<string, unknown>;
      }
    }
    return { ...restRoot, ...skillOverlay } as T;
  };

  return {
    meta: {
      id: `hermes:${skillId}`,
      name: "Hermes Agent Runtime",
      version: "0.1.0",
    },
    logger: adaptHermesLogger(api.logger),
    storage: {
      async getItem<T>(key: string): Promise<T | null> {
        const raw = store.get(key);
        if (raw === undefined) {
          return null;
        }
        try {
          return JSON.parse(raw) as T;
        } catch {
          return null;
        }
      },
      async setItem<T>(key: string, value: T): Promise<void> {
        store.set(key, JSON.stringify(value));
      },
      async removeItem(key: string): Promise<void> {
        store.delete(key);
      },
      async clear(): Promise<void> {
        store.clear();
      },
    },
    events: {
      emit(event: string, payload: unknown): void {
        const set = listeners.get(event);
        if (!set) return;
        for (const fn of set) {
          fn(payload);
        }
      },
      on(event: string, handler: (payload: unknown) => void): () => void {
        let set = listeners.get(event);
        if (!set) {
          set = new Set();
          listeners.set(event, set);
        }
        set.add(handler);
        return () => {
          set?.delete(handler);
        };
      },
      off(event: string, handler: (payload: unknown) => void): void {
        listeners.get(event)?.delete(handler);
      },
    },
    getConfig,
  };
}
