import path from 'node:path';

/**
 * Normalizes a path to POSIX format (forward slashes).
 * Useful for ensuring consistent paths across Windows and macOS/Linux.
 */
export function normalizePath(filePath: string): string {
  if (!filePath) return filePath;
  return filePath.split(path.win32.sep).join(path.posix.sep);
}

/**
 * Joins path segments and normalizes the result.
 */
export function joinPath(...paths: string[]): string {
  return normalizePath(path.join(...paths));
}

/**
 * Resolves path segments and normalizes the result.
 */
export function resolvePath(...paths: string[]): string {
  return normalizePath(path.resolve(...paths));
}
