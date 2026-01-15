
import { createHash } from 'crypto';
import { PluginLogger } from '../plugins/new-core/types';
import { noopLogger } from '../utils/logger-noop';

export interface Variant {
  name: string;
  weight: number; // 0-100
}

export interface Experiment {
  id: string;
  variants: Variant[];
}

export class ABTestManager {
  private readonly experiments = new Map<string, Experiment>();
  private logger: PluginLogger;

  constructor(logger: PluginLogger = noopLogger) {
    this.logger = logger;
  }

  public registerExperiment(experiment: Experiment): void {
    // Validate weights sum to 100 (optional but good practice)
    const totalWeight = experiment.variants.reduce((sum, v) => sum + v.weight, 0);
    if (totalWeight !== 100) {
      this.logger.warn(`Experiment ${experiment.id} weights do not sum to 100. Total: ${totalWeight}`);
    }
    this.experiments.set(experiment.id, experiment);
  }

  public getExperiment(id: string): Experiment | undefined {
    return this.experiments.get(id);
  }

  public assignVariant(experimentId: string, userId: string): string | null {
    const experiment = this.experiments.get(experimentId);
    if (!experiment) return null;

    const hash = this.hashString(`${experimentId}:${userId}`);
    const normalizedHash = hash % 100;

    let cumulativeWeight = 0;
    for (const variant of experiment.variants) {
      cumulativeWeight += variant.weight;
      if (normalizedHash < cumulativeWeight) {
        return variant.name;
      }
    }

    // Fallback (should not happen if weights sum to 100)
    return experiment.variants.at(-1)?.name ?? null;
  }

  private hashString(str: string): number {
    // Simple hash function using crypto
    const hash = createHash('sha256').update(str).digest('hex');
    // Take first 8 chars and parse as int
    return parseInt(hash.substring(0, 8), 16);
  }
}
