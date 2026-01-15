/**
 * Declarations for `openclaw/plugin-sdk/core` when typechecking the workspace extension
 * without resolving the full `openclaw` package graph.
 */
declare module "openclaw/plugin-sdk/core" {
  export function emptyPluginConfigSchema(): unknown;
  export type OpenClawPluginServiceContext = {
    logger: { info: (message: string) => void };
  };
  export type OpenClawPluginApi = {
    registerService(service: {
      id: string;
      start?: (ctx: OpenClawPluginServiceContext) => void | Promise<void>;
    }): void;
    registerTool(tool: unknown, opts?: { optional?: boolean }): void;
    resolvePath(input: string): string;
    logger: {
      debug?: (message: string, ...args: unknown[]) => void;
      info: (message: string, ...args: unknown[]) => void;
      warn: (message: string, ...args: unknown[]) => void;
      error: (message: string, error?: Error, ...args: unknown[]) => void;
    };
    pluginConfig?: Record<string, unknown>;
  };
}
