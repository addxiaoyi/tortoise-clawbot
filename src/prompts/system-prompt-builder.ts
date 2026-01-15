import { PromptTemplate } from './engine.js';

export type PromptMode = "full" | "minimal" | "none";
export type OwnerIdDisplay = "raw" | "hash";
export type MemoryCitationsMode = "auto" | "off" | "on"; // Assuming these from usage

export class SystemPromptBuilder {
  private template: PromptTemplate;

  constructor(_templatePath?: string) {
    // 模板内容由下方 PromptTemplate 默认字符串提供；路径加载可由调用方扩展。
    this.template = new PromptTemplate(
      // Fallback/Default template content if file reading fails or in test
      `{{#if skillsPrompt}}
## Skills (mandatory)
Before replying: scan <available_skills> <description> entries.
- If exactly one skill clearly applies: read its SKILL.md at <location> with \`{{readToolName}}\`, then follow it.
- If multiple could apply: choose the most specific one, then read/follow it.
- If none clearly apply: do not read any SKILL.md.
Constraints: never read more than one skill up front; only read after selecting.
- When a skill drives external API writes, assume rate limits: prefer fewer larger writes, avoid tight one-item loops, serialize bursts when possible, and respect 429/Retry-After.
{{skillsPrompt}}
{{/if}}

{{#if showMemory}}
## Memory Recall
Before answering anything about prior work, decisions, dates, people, preferences, or todos: run memory_search on MEMORY.md + memory/*.md; then use memory_get to pull only the needed lines. If low confidence after search, say you checked.
{{#if citationsDisabled}}
Citations are disabled: do not mention file paths or line numbers in replies unless the user explicitly asks.
{{/if}}
{{#if citationsEnabled}}
Citations: include Source: <path#line> when it helps the user verify memory snippets.
{{/if}}
{{/if}}

{{#if showUserIdentity}}
## Authorized Senders
{{ownerLine}}
{{/if}}

{{#if userTimezone}}
## Time
Current time: {{currentTime}} ({{userTimezone}})
{{/if}}`
    );
  }

  public build(params: {
    mode: PromptMode;
    skillsPrompt?: string;
    readToolName?: string;
    availableTools?: Set<string>;
    citationsMode?: MemoryCitationsMode;
    ownerLine?: string;
    userTimezone?: string;
    currentTime?: string;
  }): string {
    if (params.mode === 'none') return '';

    const isMinimal = params.mode === 'minimal';
    const showMemory = !isMinimal && 
      params.availableTools && 
      (params.availableTools.has("memory_search") || params.availableTools.has("memory_get"));

    const context = {
      skillsPrompt: params.skillsPrompt?.trim(),
      readToolName: params.readToolName || 'read_file',
      showMemory,
      citationsDisabled: params.citationsMode === 'off',
      citationsEnabled: params.citationsMode !== 'off',
      showUserIdentity: !isMinimal && !!params.ownerLine,
      ownerLine: params.ownerLine,
      userTimezone: params.userTimezone,
      currentTime: params.currentTime || new Date().toLocaleString()
    };

    return this.template.render(context);
  }
}
