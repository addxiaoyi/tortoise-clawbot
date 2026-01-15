import { PluginLifecycle, PluginContext } from './types.js';
import { readdir, readFile, stat } from 'node:fs/promises';
import { join } from 'node:path';

export interface LegacySkill {
  name: string;
  description: string;
  path: string;
}

export class LegacySkillAdapter implements PluginLifecycle {
  private context?: PluginContext;
  private skillsDir: string;
  private skills: LegacySkill[] = [];

  constructor(skillsDir?: string) {
    // Default to openclaw-main/skills relative to CWD if not provided
    this.skillsDir = skillsDir || join(process.cwd(), 'openclaw-main/skills');
  }

  async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    ctx.logger.info(`Initializing Legacy Skill Adapter with dir: ${this.skillsDir}`);
    await this.loadSkills();
  }

  async onStart(): Promise<void> {
    this.context?.logger.info(`Legacy Adapter started. Loaded ${this.skills.length} skills.`);
    // Emit event for other components to consume
    this.context?.events.emit('skills:loaded', this.skills);
  }

  private async loadSkills(): Promise<void> {
    try {
      // Check if dir exists first
      try {
        await stat(this.skillsDir);
      } catch {
        this.context?.logger.warn(`Skills directory not found: ${this.skillsDir}`);
        return;
      }

      const entries = await readdir(this.skillsDir, { withFileTypes: true });
      for (const entry of entries) {
        const name = entry.name;
        if (
          name === '.' ||
          name === '..' ||
          name.includes('/') ||
          name.includes('\\') ||
          name.includes('\0')
        ) {
          continue;
        }
        if (entry.isDirectory()) {
          const skillPath = join(this.skillsDir, name, 'SKILL.md');
          try {
            const stats = await stat(skillPath);
            if (stats.isFile()) {
              const content = await readFile(skillPath, 'utf-8');
              const description = this.extractDescription(content);
              this.skills.push({
                name,
                description,
                path: skillPath
              });
            }
          } catch {
            this.context?.logger.debug(`Skipping ${name}: No SKILL.md found`);
          }
        }
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : String(e);
      this.context?.logger.error(`Failed to load legacy skills: ${message}`);
    }
  }

  private extractDescription(content: string): string {
    // Attempt to find the first meaningful paragraph
    // 1. Remove comments
    const cleanContent = content.replace(/<!--[\s\S]*?-->/g, '');
    // 2. Find lines that are not headers (#) and not empty
    const lines = cleanContent.split('\n');
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed && !trimmed.startsWith('#') && !trimmed.startsWith('```')) {
        return trimmed;
      }
    }
    return 'No description provided';
  }

  public getSkills(): LegacySkill[] {
    return this.skills;
  }
}
