/**
 * Simple Prompt Template Engine
 * Supports {{variable}}, {{#if condition}}...{{/if}}, {{#each list}}...{{/each}}
 */

const MAX_TEMPLATE_LENGTH = 50000;
const MAX_ITERATIONS = 1000;

export class PromptTemplate {
  private template: string;
  private iterationCount = 0;

  constructor(template: string) {
    if (template.length > MAX_TEMPLATE_LENGTH) {
      throw new Error(`Template exceeds maximum length of ${MAX_TEMPLATE_LENGTH}`);
    }
    this.template = template;
  }

  public render(context: Record<string, any>): string {
    this.iterationCount = 0;
    let result = this.template;

    result = this.renderConditionals(result, context);
    result = this.renderEach(result, context);
    result = this.renderVariables(result, context);

    return result;
  }

  private checkIteration(): void {
    this.iterationCount++;
    if (this.iterationCount > MAX_ITERATIONS) {
      throw new Error('Template rendering exceeded maximum iteration count');
    }
  }

  private renderConditionals(result: string, context: Record<string, any>): string {
    const pattern = /\{\{#if\s+([a-zA-Z0-9_.]+)\}\}([\s\S]*?)\{\{\/if\}\}/g;
    return result.replace(pattern, (match, key, content) => {
      this.checkIteration();
      const value = this.getValue(context, key);
      return value ? content : '';
    });
  }

  private renderEach(result: string, context: Record<string, any>): string {
    const pattern = /\{\{#each\s+([a-zA-Z0-9_.]+)\}\}([\s\S]*?)\{\{\/each\}\}/g;
    return result.replace(pattern, (match, key, content) => {
      this.checkIteration();
      const value = this.getValue(context, key);
      if (!Array.isArray(value)) return '';

      return value.map(item => {
        let itemContent = content;
        itemContent = itemContent.replace(/\{\{this\}\}/g, String(item));

        if (typeof item === 'object' && item !== null) {
          itemContent = itemContent.replace(/\{\{([a-zA-Z0-9_.]+)\}\}/g, (m: string, k: string) => {
            if (k === 'this') return String(item);
            const v = this.getValue(item, k);
            return v !== undefined ? String(v) : m;
          });
        }
        return itemContent;
      }).join('');
    });
  }

  private renderVariables(result: string, context: Record<string, any>): string {
    const pattern = /\{\{([a-zA-Z0-9_.]+)\}\}/g;
    return result.replace(pattern, (match, key) => {
      this.checkIteration();
      const value = this.getValue(context, key);
      return value !== undefined ? String(value) : match;
    });
  }

  private getValue(context: any, path: string): any {
    return path.split('.').reduce((obj, key) => obj?.[key], context);
  }
}