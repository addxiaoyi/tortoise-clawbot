import { PluginLogger } from '../plugins/new-core/types';

export const noopLogger: PluginLogger = {
  debug: () => {},
  info: () => {},
  warn: () => {},
  error: () => {},
};
