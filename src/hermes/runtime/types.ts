export type HermesRuntimeLogger = {
  debug?: (message: string, ...args: unknown[]) => void;
  info: (message: string, ...args: unknown[]) => void;
  warn: (message: string, ...args: unknown[]) => void;
  error: (message: string, error?: Error, ...args: unknown[]) => void;
};

export type HermesRuntimeConfig = Record<string, unknown>;

export type HermesRuntimeApi = {
  resolvePath: (input: string) => string;
  logger: HermesRuntimeLogger;
  config?: HermesRuntimeConfig;
  registerTool?: (tool: unknown, opts?: { optional?: boolean }) => void;
};

export type HermesSession = {
  sessionId?: string;
  sessionKey?: string;
};

export type HermesInvokeOptions = {
  signal?: AbortSignal;
  timeoutMs?: number;
};
