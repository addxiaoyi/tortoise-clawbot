import path from 'node:path';

/** 按扩展名返回 Content-Type；未知扩展名使用 octet-stream 以减少浏览器 MIME 嗅探执行脚本。 */
export function contentTypeForFile(filePath: string): string {
  const ext = path.extname(filePath).toLowerCase();
  switch (ext) {
    case '.html':
    case '.htm':
      return 'text/html; charset=utf-8';
    case '.css':
      return 'text/css; charset=utf-8';
    case '.js':
    case '.mjs':
      return 'text/javascript; charset=utf-8';
    case '.json':
      return 'application/json; charset=utf-8';
    case '.svg':
      return 'image/svg+xml; charset=utf-8';
    case '.txt':
    case '.md':
      return 'text/plain; charset=utf-8';
    case '.xml':
      return 'application/xml; charset=utf-8';
    case '.png':
      return 'image/png';
    case '.jpg':
    case '.jpeg':
      return 'image/jpeg';
    case '.gif':
      return 'image/gif';
    case '.webp':
      return 'image/webp';
    case '.ico':
      return 'image/x-icon';
    case '.woff':
      return 'font/woff';
    case '.woff2':
      return 'font/woff2';
    case '.pdf':
      return 'application/pdf';
    default:
      return 'application/octet-stream';
  }
}

const TEXT_LIKE_EXT = new Set([
  '.html',
  '.htm',
  '.css',
  '.js',
  '.mjs',
  '.json',
  '.svg',
  '.txt',
  '.md',
  '.xml',
]);

export function shouldServeFileAsUtf8Text(filePath: string): boolean {
  return TEXT_LIKE_EXT.has(path.extname(filePath).toLowerCase());
}
