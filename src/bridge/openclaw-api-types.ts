/**
 * Logger shape accepted by the bridge, aligned with OpenClaw `PluginLogger` (optional `debug`).
 */
export type TohelpBridgeLogger = {
  debug?: (message: string, ...args: unknown[]) => void;
  info: (message: string, ...args: unknown[]) => void;
  warn: (message: string, ...args: unknown[]) => void;
  error: (message: string, error?: Error, ...args: unknown[]) => void;
};

/**
 * Subset of OpenClaw `OpenClawPluginApi` used by the Tohelp bridge (no dependency on `openclaw` package).
 */
export type TohelpOpenClawApi = {
  resolvePath: (input: string) => string;
  registerTool: (tool: unknown, opts?: { optional?: boolean }) => void;
  logger: TohelpBridgeLogger;
  pluginConfig?: Record<string, unknown>;
  config?: Record<string, unknown>;
};
