import { PluginContext, PluginConfig } from './types.js';

export class SafePluginContext implements PluginContext {
  private baseContext: PluginContext;
  private errorBoundary: (error: Error) => void;
  private handlerMap: Map<(payload: any) => void, (payload: any) => void> = new Map();

  constructor(baseContext: PluginContext, errorBoundary: (error: Error) => void) {
    this.baseContext = baseContext;
    this.errorBoundary = errorBoundary;
  }

  get meta() { return this.baseContext.meta; }
  get logger() { return this.baseContext.logger; }
  get storage() { return this.baseContext.storage; }

  get events() {
    const originalEvents = this.baseContext.events;
    return {
      emit: originalEvents.emit.bind(originalEvents),
      on: (event: string, handler: (payload: any) => void) => {
        const safeHandler = (payload: any) => {
          try {
            handler(payload);
          } catch (error) {
            this.errorBoundary(
              error instanceof Error ? error : new Error(String(error)),
            );
          }
        };
        this.handlerMap.set(handler, safeHandler);
        return originalEvents.on(event, safeHandler);
      },
      off: (event: string, handler: (payload: any) => void) => {
        const safeHandler = this.handlerMap.get(handler);
        if (safeHandler) {
          originalEvents.off(event, safeHandler);
          this.handlerMap.delete(handler);
        }
      },
    };
  }

  getConfig<T extends PluginConfig>() {
    return this.baseContext.getConfig<T>();
  }
}
