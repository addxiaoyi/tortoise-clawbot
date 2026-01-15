import fs from 'node:fs';
import path from 'node:path';
import { EventEmitter } from 'node:events';

/**
 * Options for ConfigWatcher
 */
export interface ConfigWatcherOptions {
  configPath: string;
  pollInterval?: number;
  debounceInterval?: number;
  maxReconnectAttempts?: number;
  reconnectDelayMs?: number;
}

/**
 * Watches a configuration file for changes and emits events.
 * Supports hot-reloading by reading the new config on change.
 */
export class ConfigWatcher extends EventEmitter {
  private configPath: string;
  private currentConfig: any = null;
  private watcher: fs.FSWatcher | null = null;
  private debounceTimer: NodeJS.Timeout | null = null;
  private readonly debounceInterval: number;
  private isClosed = false;
  private reconnectAttempts = 0;
  private readonly maxReconnectAttempts: number;
  private readonly reconnectDelayMs: number;

  constructor(options: ConfigWatcherOptions) {
    super();
    this.configPath = path.resolve(options.configPath);
    this.debounceInterval = options.debounceInterval || 100;
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 3;
    this.reconnectDelayMs = options.reconnectDelayMs ?? 1000;
  }

  /**
   * Start watching the configuration file.
   * Emits 'change' event with the new config object.
   * Emits 'error' event on failure.
   */
  public async start(): Promise<void> {
    if (this.watcher) return;

    await this.loadConfig();
    this.setupWatcher();
  }

  private setupWatcher(): void {
    if (this.isClosed) return;

    try {
      if (this.watcher) {
        this.watcher.close();
      }
      
      this.watcher = fs.watch(this.configPath, (eventType) => {
        if (eventType === 'change' || eventType === 'rename') {
          this.handleFileChange();
        }
      });
      
      this.watcher.on('error', (error) => {
        this.reconnectAttempts++;
        const errorMessage = error instanceof Error ? error.message : String(error);
        if (this.reconnectAttempts <= this.maxReconnectAttempts) {
          this.emit('error', new Error(`Watcher error, reconnecting (${this.reconnectAttempts}/${this.maxReconnectAttempts}): ${errorMessage}`));
          setTimeout(() => this.setupWatcher(), this.reconnectDelayMs);
        } else {
          this.emit('error', new Error(`Watcher error after ${this.maxReconnectAttempts} attempts: ${errorMessage}`));
        }
      });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      this.emit('error', new Error(message));
    }
  }

  /**
   * Stop watching.
   */
  public stop(): void {
    this.isClosed = true;
    if (this.watcher) {
      this.watcher.close();
      this.watcher = null;
    }
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
      this.debounceTimer = null;
    }
  }

  /**
   * Get the current configuration.
   */
  public getConfig(): any {
    return this.currentConfig;
  }

  private handleFileChange() {
    if (this.isClosed) return;

    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }

    this.debounceTimer = setTimeout(() => {
      this.loadConfig();
    }, this.debounceInterval);
  }

  private async loadConfig() {
    try {
      // Check if file exists (it might be briefly missing during atomic writes)
      if (!fs.existsSync(this.configPath)) {
        return;
      }

      const content = await fs.promises.readFile(this.configPath, 'utf-8');
      
      // Parse JSON
      let newConfig;
      try {
        newConfig = JSON.parse(content);
      } catch (parseError) {
        const message = parseError instanceof Error ? parseError.message : String(parseError);
        this.emit('error', new Error(`Failed to parse config JSON: ${message}`));
        return;
      }

      // Deep compare could be added here to avoid emitting if no actual change
      // For now, assume change if file changed
      this.currentConfig = newConfig;
      this.emit('change', newConfig);
      
    } catch (error) {
      this.emit('error', error);
    }
  }
}
