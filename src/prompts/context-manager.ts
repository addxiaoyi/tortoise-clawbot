
export interface Message {
  role: 'system' | 'user' | 'assistant' | 'tool';
  content: string;
}

export class ContextManager {
  private messages: Message[] = [];
  private readonly tokenLimit: number;

  constructor(tokenLimit: number = 4096) {
    this.tokenLimit = tokenLimit;
  }

  public addMessage(message: Message): void {
    this.messages.push(message);
    this.prune();
  }

  public getMessages(): Message[] {
    return this.messages;
  }

  public estimateTokens(content: string): number {
    // Simple heuristic: 1 token ≈ 4 characters
    return Math.ceil(content.length / 4);
  }

  private prune(): void {
    let currentTokens = this.messages.reduce((sum, msg) => sum + this.estimateTokens(msg.content), 0);

    if (currentTokens <= this.tokenLimit) return;

    // Remove oldest non-system messages
    while (currentTokens > this.tokenLimit && this.messages.length > 0) {
      const indexToRemove = this.messages.findIndex(msg => msg.role !== 'system');
      
      if (indexToRemove === -1) {
        // Only system messages left, but still over limit.
        // We might need to truncate the last system message or just stop.
        // For safety, we stop here to avoid removing critical system instructions.
        break;
      }

      const removedMsg = this.messages.splice(indexToRemove, 1)[0];
      currentTokens -= this.estimateTokens(removedMsg.content);
    }
  }
}
