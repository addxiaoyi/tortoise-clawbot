
# Canvas Plugin

Migrated from `src/canvas-host` to MCP architecture.

## Configuration
```typescript
interface CanvasConfig {
  port: number; // e.g., 18793
  root: string; // e.g., '/path/to/canvas/root'
}
```

## Features
- Serves static files from `root` directory.
- Security checks to prevent directory traversal.
- Starts/Stops with plugin lifecycle.

## Future Improvements
- Add WebSocket support for live reload.
- Add `chokidar` for file watching.
