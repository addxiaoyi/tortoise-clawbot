/**
 * `tohelp_memory` 可选安全策略（环境变量）。
 * 网关 / MCP 进程需在启动前注入；OpenClaw 若从 shell 启动且已 `source .env` 则生效。
 */

/** OpenClaw 在解析插件工具时传入的上下文字段（子集，避免依赖 openclaw 类型包）。 */
export type TohelpMemoryRuntime = {
  sessionId?: string;
  sessionKey?: string;
};

/** 启用后：网关侧优先取 `sessionKey`，其次 `sessionId`，在静态前缀后追加 `sess:<id>:`。 */
export function loadMemoryAutoSessionPrefix(): boolean {
  const v = process.env.TOHELP_MEMORY_AUTO_SESSION_PREFIX;
  if (v === undefined) return false;
  const t = v.trim().toLowerCase();
  return t === "1" || t === "true" || t === "yes" || t === "on";
}

/** 启用后：写操作（set/delete/clear）必须携带 `sessionKey`，否则拒绝。 */
export function loadMemoryRequireSessionKeyForWrite(): boolean {
  const v = process.env.TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE;
  if (v === undefined) return false;
  const t = v.trim().toLowerCase();
  return t === "1" || t === "true" || t === "yes" || t === "on";
}

/** 启用后：读操作（get/list）也必须携带 `sessionKey`，否则拒绝。 */
export function loadMemoryRequireSessionKeyForRead(): boolean {
  const v = process.env.TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ;
  if (v === undefined) return false;
  const t = v.trim().toLowerCase();
  return t === "1" || t === "true" || t === "yes" || t === "on";
}

const SESSION_SCOPE_SEGMENT_MAX = 128;

/** 限制字符集与长度，避免键名注入或过长。 */
export function sanitizeSessionIdForKey(sessionId: string): string {
  const s = sessionId.trim().slice(0, SESSION_SCOPE_SEGMENT_MAX);
  if (!s) return "unknown";
  return s.replace(/[^a-zA-Z0-9_.-]/g, "_");
}

/**
 * 最终物理键前缀 = `TOHELP_MEMORY_KEY_PREFIX` +（可选）`sess:<sessionScopeId>:`。
 * `sessionScopeId` 选择顺序：`sessionKey` > `sessionId`。
 */
export function resolveEffectiveMemoryPrefix(
  runtime?: TohelpMemoryRuntime,
): {
  effectivePrefix: string;
  sessionScoped: boolean;
  sessionScopeSkipped: boolean;
} {
  const staticPrefix = loadMemoryKeyPrefix();
  const auto = loadMemoryAutoSessionPrefix();
  const sid = runtime?.sessionKey?.trim() || runtime?.sessionId?.trim();
  if (auto && sid) {
    const seg = sanitizeSessionIdForKey(sid);
    return {
      effectivePrefix: `${staticPrefix}sess:${seg}:`,
      sessionScoped: true,
      sessionScopeSkipped: false,
    };
  }
  return {
    effectivePrefix: staticPrefix,
    sessionScoped: false,
    sessionScopeSkipped: auto && !sid,
  };
}

/** 逻辑 key → 物理 key；未设置前缀时二者相同。 */
export function loadMemoryKeyPrefix(): string {
  const raw = process.env.TOHELP_MEMORY_KEY_PREFIX;
  if (raw === undefined) return "";
  const t = raw.trim();
  return t;
}

/**
 * 单条 `set` 的 value 经 JSON 序列化后的 UTF-8 字节上限。
 * 未设置或非正数：不限制。
 */
export function loadMemoryMaxValueBytes(): number | null {
  const raw = process.env.TOHELP_MEMORY_MAX_VALUE_BYTES;
  if (raw === undefined) return null;
  const t = raw.trim();
  if (t === "") return null;
  const n = Number.parseInt(t, 10);
  if (!Number.isFinite(n) || n <= 0) return null;
  return n;
}

export function physicalMemoryKey(logicalKey: string, prefix: string): string {
  return prefix ? `${prefix}${logicalKey}` : logicalKey;
}

/** list：在前缀模式下仅列出此前缀下的逻辑 key（已去掉前缀），并排序以稳定输出。 */
export function listLogicalMemoryKeys(
  allKeys: string[],
  prefix: string,
): string[] {
  if (!prefix) {
    return [...allKeys];
  }
  const out: string[] = [];
  for (const k of allKeys) {
    if (k.startsWith(prefix)) {
      out.push(k.slice(prefix.length));
    }
  }
  out.sort();
  return out;
}

export function assertMemoryValueSizeWithinLimit(
  value: unknown,
  maxBytes: number | null,
): void {
  if (maxBytes === null) return;
  let serialized: string | undefined;
  try {
    serialized = JSON.stringify(value);
  } catch {
    throw new Error("value is not JSON-serializable for size check");
  }
  if (serialized === undefined) {
    return;
  }
  const bytes = Buffer.byteLength(serialized, "utf8");
  if (bytes > maxBytes) {
    throw new Error(
      `value exceeds TOHELP_MEMORY_MAX_VALUE_BYTES (${maxBytes} bytes, got ${bytes})`,
    );
  }
}
