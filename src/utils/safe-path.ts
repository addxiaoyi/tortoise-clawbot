import path from "node:path";

function isPathInsideRoot(resolvedRoot: string, resolvedTarget: string): boolean {
  const rel = path.relative(resolvedRoot, resolvedTarget);
  return !rel.startsWith("..") && !path.isAbsolute(rel);
}

/**
 * 将若干相对路径段解析到 root 之下；若结果逃出 root 则抛出。
 */
export function resolvePathUnderRoot(root: string, ...relativeSegments: string[]): string {
  const resolvedRoot = path.resolve(root);
  const resolved =
    relativeSegments.length === 0
      ? resolvedRoot
      : path.resolve(resolvedRoot, ...relativeSegments);
  if (!isPathInsideRoot(resolvedRoot, resolved)) {
    throw new Error("Path escapes root directory");
  }
  return resolved;
}

/**
 * 用户提供的相对文件路径（可含子目录），禁止绝对路径与空字节。
 */
export function resolveSafeRelativeFile(root: string, filename: string): string {
  if (filename.includes("\0")) {
    throw new Error("Invalid filename");
  }
  const trimmed = filename.trim();
  if (!trimmed) {
    throw new Error("Invalid filename");
  }
  if (path.isAbsolute(trimmed)) {
    throw new Error("Invalid filename: absolute path not allowed");
  }
  const parts = trimmed.split(/[/\\]+/).filter((p) => p.length > 0);
  return resolvePathUnderRoot(root, ...parts);
}

/**
 * HTTP 静态服务：解码 URL pathname 并解析为 root 下的绝对路径。
 */
export function resolveUrlPathnameToSafePath(root: string, pathname: string): string {
  let decoded: string;
  try {
    decoded = decodeURIComponent(pathname);
  } catch {
    throw new Error("Invalid URL path");
  }
  if (decoded.includes("\0")) {
    throw new Error("Invalid URL path");
  }
  const rel = decoded.replace(/^[/\\]+/, "");
  if (path.isAbsolute(rel)) {
    throw new Error("Invalid URL path");
  }
  const parts = rel.split(/[/\\]+/).filter((p) => p.length > 0);
  return resolvePathUnderRoot(root, ...parts);
}

/**
 * MCP / 工作区路径解析：解析结果必须落在 workspaceRoot 内。
 * - 相对路径：相对 workspaceRoot；
 * - 绝对路径：仅当规范化后仍位于 workspaceRoot 之下时允许。
 */
export function resolveWorkspacePath(workspaceRoot: string, input: string): string {
  const root = path.resolve(workspaceRoot);
  const t = input.trim();
  if (!t) {
    return root;
  }
  if (path.isAbsolute(t)) {
    const normalized = path.resolve(t);
    if (!isPathInsideRoot(root, normalized)) {
      throw new Error("Path escapes workspace root");
    }
    return normalized;
  }
  return resolveSafeRelativeFile(root, t);
}
