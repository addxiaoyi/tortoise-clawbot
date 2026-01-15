/**
 * MCP Tools Module
 * 
 * Model Context Protocol tool management
 */

import { AxiosInstance } from 'axios';

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
  annotations?: {
    readOnlyHint?: boolean;
    destructiveHint?: boolean;
    idempotentHint?: boolean;
    openWorldHint?: boolean;
  };
}

export interface ToolResult {
  content: Array<{
    type: 'text' | 'image' | 'resource';
    text?: string;
    data?: string;
    mimeType?: string;
  }>;
  isError?: boolean;
}

export class McpTools {
  constructor(private http: AxiosInstance) {}
  
  /**
   * List all available tools
   */
  async list(): Promise<ToolDefinition[]> {
    const response = await this.http.get<{ tools: ToolDefinition[] }>('/api/v1/mcp/tools');
    return response.data.tools;
  }
  
  /**
   * Call a tool
   */
  async call(toolName: string, arguments_: Record<string, unknown> = {}): Promise<ToolResult> {
    const response = await this.http.post<ToolResult>('/api/v1/mcp/tools/call', {
      name: toolName,
      arguments: arguments_,
    });
    return response.data;
  }
  
  /**
   * Get tool definition
   */
  async get(toolName: string): Promise<ToolDefinition | undefined> {
    const tools = await this.list();
    return tools.find(t => t.name === toolName);
  }
  
  /**
   * Call tool and extract text result
   */
  async callAndExtractText(toolName: string, args: Record<string, unknown> = {}): Promise<string> {
    const result = await this.call(toolName, args);
    const textContent = result.content.find(c => c.type === 'text');
    return textContent?.text || '';
  }
}
