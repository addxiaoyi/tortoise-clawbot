
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Container } from '../../plugins/new-core/container';
import { PluginRegistry } from '../../plugins/new-core/registry';
import { GitHubPlugin } from '../../plugins/new-core/github/plugin';
import { CanvasPlugin } from '../../plugins/new-core/canvas/plugin';
import { SlackPlugin } from '../../plugins/new-core/slack/plugin';
import { NotionPlugin } from '../../plugins/new-core/notion/plugin';
import { PromptRegistry } from '../../prompts/registry';
import { PluginContext } from '../../plugins/new-core/types';
import { NotionService } from '../../plugins/new-core/notion/service';
import http from 'http';
import fs from 'fs/promises';

// Mock dependencies
vi.mock('http');
vi.mock('fs/promises');
vi.mock('../../plugins/new-core/github/service');
vi.mock('../../plugins/new-core/slack/service');
vi.mock('../../plugins/new-core/notion/service');

describe('Full System Integration', () => {
  let container: Container;
  let registry: PluginRegistry;
  let githubPlugin: GitHubPlugin;
  let canvasPlugin: CanvasPlugin;
  let slackPlugin: SlackPlugin;
  let notionPlugin: NotionPlugin;
  let promptRegistry: PromptRegistry;
  let mockBaseContext: PluginContext;

  beforeEach(() => {
    container = new Container();
    registry = new PluginRegistry(container);
    
    githubPlugin = new GitHubPlugin();
    canvasPlugin = new CanvasPlugin();
    slackPlugin = new SlackPlugin();
    notionPlugin = new NotionPlugin();
    promptRegistry = new PromptRegistry();

    // Mock GitHub Service on the plugin instance
    const mockGitHubService = {
      checkAuth: vi.fn().mockResolvedValue(true),
      listIssues: vi.fn().mockResolvedValue([{ title: 'Issue 1' }, { title: 'Issue 2' }])
    };
    // @ts-ignore
    githubPlugin.service = mockGitHubService;

    // Mock Slack Service
    const mockSlackService = {
      checkAuth: vi.fn().mockResolvedValue(true),
      sendMessage: vi.fn().mockResolvedValue({ ok: true, ts: '1234.5678' })
    };
    // @ts-ignore
    slackPlugin.service = mockSlackService;

    // Mock Notion Service
    const mockNotionService = {
      checkAuth: vi.fn().mockResolvedValue(true),
      search: vi.fn().mockResolvedValue({
        results: [{ id: 'page-1', properties: { title: 'My Page' } }]
      })
    };
    // @ts-ignore
    NotionService.mockImplementation(function () {
      return mockNotionService;
    });

    // Mock Canvas Server
    const mockServer = {
      listen: vi.fn((port, _host, cb) => cb && cb()),
      close: vi.fn((cb) => cb && cb()),
      on: vi.fn()
    };
    vi.mocked(http.createServer).mockReturnValue(mockServer as any);

    // Register plugins
    registry.register({
      meta: { id: 'github', name: 'GitHub Plugin', version: '1.0.0' },
      lifecycle: githubPlugin
    });
    registry.register({
      meta: { id: 'canvas', name: 'Canvas Plugin', version: '1.0.0' },
      lifecycle: canvasPlugin
    });
    registry.register({
      meta: { id: 'slack', name: 'Slack Plugin', version: '1.0.0' },
      lifecycle: slackPlugin
    });
    registry.register({
      meta: { id: 'notion', name: 'Notion Plugin', version: '1.0.0' },
      lifecycle: notionPlugin
    });

    // Create base context
    mockBaseContext = {
      meta: { id: 'system', name: 'System', version: '0.0.0' },
      logger: { debug: vi.fn(), info: vi.fn(), warn: vi.fn(), error: vi.fn() },
      storage: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() },
      events: { emit: vi.fn(), on: vi.fn(), off: vi.fn() },
      getConfig: () =>
        ({
          port: 18793,
          root: '/tmp/canvas',
          token: 'mock-slack-token', // Shared config hack for tests
          apiKey: 'mock-notion-key'
        } as any)
    } as unknown as PluginContext;
  });

  afterEach(async () => {
    await registry.stopAll();
    vi.clearAllMocks();
  });

  it('should flow data from GitHub to Prompt to Canvas/Slack/Notion', async () => {
    // 1. Initialize and Start System
    await registry.initAll(mockBaseContext);
    await registry.startAll();

    // Verify Plugins started and checked auth
    // @ts-ignore
    expect(githubPlugin.service.checkAuth).toHaveBeenCalled();
    // @ts-ignore
    expect(slackPlugin.service.checkAuth).toHaveBeenCalled();
    // @ts-ignore
    expect(notionPlugin.service.checkAuth).toHaveBeenCalled();
    expect(http.createServer).toHaveBeenCalled();

    // 2. Fetch Data (simulate agent action)
    // @ts-ignore
    const issues = await githubPlugin.service.listIssues('owner/repo');
    expect(issues).toHaveLength(2);

    // 3. Process with Prompt System
    const templateContent = "Found {{count}} issues: {{#each issues}}- {{title}}\n{{/each}}";
    const { PromptTemplate } = await import('../../prompts/engine');
    const template = new PromptTemplate(templateContent);
    const rendered = template.render({ count: issues.length, issues });
    
    expect(rendered).toContain('Found 2 issues');

    // 4. "Send" to Slack
    // @ts-ignore
    await slackPlugin.service.sendMessage({ channel: 'C123', text: rendered });
    // @ts-ignore
    expect(slackPlugin.service.sendMessage).toHaveBeenCalledWith({ channel: 'C123', text: rendered });

    // 5. "Search" in Notion (just to verify capability)
    // @ts-ignore
    const notionResults = await notionPlugin.service.search('test');
    expect(notionResults.results[0].id).toBe('page-1');
  });
});
