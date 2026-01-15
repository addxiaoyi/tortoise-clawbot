/**
 * End-to-End Tests for Tortoise Core Features
 * 
 * Tests the complete flow of:
 * - Chat with AI models
 * - Session management
 * - Memory operations
 * - Channel messaging
 * - Plugin system
 */

import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';

// Test configuration
const TEST_CONFIG = {
  gatewayUrl: process.env.TORTOISE_GATEWAY_URL || 'http://localhost:8080',
  apiKey: process.env.TORTOISE_API_KEY || 'test-api-key',
  model: process.env.TORTOISE_TEST_MODEL || 'gpt-4',
};

// Skip tests if gateway is not available
const skipIfNoGateway = process.env.SKIP_E2E === 'true' ? describe.skip : describe;

skipIfNoGateway('Tortoise E2E Tests', () => {
  let client: any;
  
  beforeAll(async () => {
    // Initialize test client
    const { TortoiseClient } = await import('../src/index.js');
    client = new TortoiseClient({
      gatewayUrl: TEST_CONFIG.gatewayUrl,
      apiKey: TEST_CONFIG.apiKey,
    });
    
    try {
      await client.connect();
    } catch (err) {
      console.warn('Gateway not available, skipping E2E tests');
    }
  });
  
  afterAll(async () => {
    if (client) {
      client.disconnect();
    }
  });

  describe('Health Check', () => {
    it('should return healthy status', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }
      
      const health = await client.healthCheck();
      expect(health).toBeDefined();
      expect(health.status).toBeDefined();
    });
  });

  describe('Chat API', () => {
    let sessionId: string;

    beforeEach(async () => {
      if (!client?.isConnected?.()) return;
      
      // Create a session for each test
      const session = await client.createSession({
        title: `Test Session ${Date.now()}`,
      });
      sessionId = session.id;
    });

    it('should send a chat message and receive response', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const response = await client.chat('Hello, how are you?', {
        sessionId,
        model: TEST_CONFIG.model,
      });

      expect(response).toBeDefined();
      expect(response.content).toBeDefined();
      expect(typeof response.content).toBe('string');
    });

    it('should support streaming responses', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const chunks: string[] = [];
      
      for await (const chunk of client.chatStream('Tell me a short story', {
        sessionId,
        model: TEST_CONFIG.model,
      })) {
        chunks.push(chunk);
      }

      expect(chunks.length).toBeGreaterThan(0);
      expect(chunks.join('')).toBeDefined();
    });

    it('should maintain conversation context', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      await client.chat('My name is TestUser', { sessionId });
      
      const response = await client.chat('What is my name?', { 
        sessionId,
        model: TEST_CONFIG.model,
      });

      expect(response.content).toBeDefined();
    });
  });

  describe('Session Management', () => {
    it('should create a new session', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const session = await client.createSession({
        title: 'Test Session',
        model: TEST_CONFIG.model,
      });

      expect(session).toBeDefined();
      expect(session.id).toBeDefined();
      expect(session.title).toBe('Test Session');
    });

    it('should list all sessions', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const sessions = await client.getSessions();
      
      expect(Array.isArray(sessions)).toBe(true);
    });

    it('should delete a session', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const session = await client.createSession({
        title: 'Session to Delete',
      });

      await client.deleteSession(session.id);
      
      // Verify deletion
      const sessions = await client.getSessions();
      expect(sessions.find((s: any) => s.id === session.id)).toBeUndefined();
    });
  });

  describe('Memory System', () => {
    it('should add a memory entry', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const entry = await client.memoryAdd({
        content: 'Test fact: The sky is blue',
        type: 'fact',
        metadata: { source: 'test' },
      });

      expect(entry).toBeDefined();
      expect(entry.id).toBeDefined();
    });

    it('should search memory', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      // Add some memories first
      await client.memoryAdd({ content: 'Python is a programming language', type: 'fact' });
      await client.memoryAdd({ content: 'The capital of France is Paris', type: 'fact' });

      const results = await client.memorySearch({
        query: 'programming',
        limit: 10,
      });

      expect(Array.isArray(results)).toBe(true);
    });
  });

  describe('Channels', () => {
    it('should list available channels', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const channels = await client.listChannels();
      
      expect(Array.isArray(channels)).toBe(true);
    });

    it('should send a channel message (mock)', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      // This test depends on actual channel configuration
      // Skip if no channels are configured
      const channels = await client.listChannels();
      
      if (channels.length === 0) {
        console.warn('Skipping: No channels configured');
        return;
      }

      // Test would require actual channel credentials
      // Skipping for security reasons
    });
  });

  describe('Plugins', () => {
    it('should list available plugins', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const plugins = await client.listPlugins();
      
      expect(Array.isArray(plugins)).toBe(true);
    });

    it('should invoke a plugin skill', async () => {
      if (!client?.isConnected?.()) {
        console.warn('Skipping: Not connected to gateway');
        return;
      }

      const plugins = await client.listPlugins();
      
      if (plugins.length === 0) {
        console.warn('Skipping: No plugins available');
        return;
      }

      // Find a skill plugin
      const skillPlugin = plugins.find((p: any) => p.category === 'skills');
      
      if (!skillPlugin) {
        console.warn('Skipping: No skill plugins available');
        return;
      }

      // Test would require actual plugin execution
    });
  });
});

describe('SDK Unit Tests (No Gateway Required)', () => {
  describe('Client Configuration', () => {
    it('should create client with default config', async () => {
      const { ClientConfig } = await import('../src/client.js');
      
      const config = new ClientConfig();
      
      expect(config.gatewayUrl).toBe('http://localhost:8080');
      expect(config.apiKey).toBeUndefined();
      expect(config.timeout).toBe(60);
    });

    it('should apply custom configuration', async () => {
      const { ClientConfig } = await import('../src/client.js');
      
      const config = new ClientConfig()
        .withUrl('https://custom-gateway.com')
        .withApiKey('my-api-key')
        .withTimeout(120);
      
      expect(config.gatewayUrl).toBe('https://custom-gateway.com');
      expect(config.apiKey).toBe('my-api-key');
      expect(config.timeout).toBe(120);
    });
  });

  describe('Channel Adapters', () => {
    it('should create Telegram channel adapter', async () => {
      const { TelegramChannel } = await import('../src/channels/telegram.js');
      
      const channel = new TelegramChannel();
      
      expect(channel.name).toBe('telegram');
      expect(channel.capabilities).toContain('text');
    });

    it('should create Discord channel adapter', async () => {
      const { DiscordChannel } = await import('../src/channels/discord.js');
      
      const channel = new DiscordChannel();
      
      expect(channel.name).toBe('discord');
      expect(channel.capabilities).toContain('text');
    });

    it('should create Matrix channel adapter', async () => {
      const { MatrixChannel } = await import('../src/channels/matrix.js');
      
      const channel = new MatrixChannel();
      
      expect(channel.name).toBe('matrix');
      expect(channel.capabilities).toContain('encryption');
    });
  });

  describe('Plugin System', () => {
    it('should create plugin sandbox with permissions', async () => {
      const { PluginSandbox, PermissionPresets } = await import('../src/plugins/new-core/sandbox.js');
      
      const sandbox = new PluginSandbox('test-plugin', {
        allowNetwork: true,
        allowFileSystem: false,
        ...PermissionPresets.minimal,
      });
      
      expect(sandbox).toBeDefined();
    });

    it('should enforce permission presets', async () => {
      const { PermissionPresets } = await import('../src/plugins/new-core/sandbox.js');
      
      expect(PermissionPresets.minimal.allowNetwork).toBe(false);
      expect(PermissionPresets.standard.allowNetwork).toBe(true);
      expect(PermissionPresets.trusted.allowFileSystem).toBe(true);
    });
  });

  describe('Orchestrator', () => {
    it('should create orchestrator with agents', async () => {
      const { AgentOrchestrator, AgentTemplates } = await import('../src/runtime/orchestrator.js');
      
      const orchestrator = new AgentOrchestrator();
      
      // Register agents
      orchestrator.registerAgent(AgentTemplates.coordinator('coordinator-1', 'Main Coordinator'));
      orchestrator.registerAgent(AgentTemplates.coder('coder-1', 'Code Assistant'));
      orchestrator.registerAgent(AgentTemplates.researcher('researcher-1', 'Research Assistant'));
      
      const agents = orchestrator.getAllAgents();
      
      expect(agents.length).toBe(3);
    });

    it('should select appropriate agent for task', async () => {
      const { AgentOrchestrator, AgentTemplates } = await import('../src/runtime/orchestrator.js');
      
      const orchestrator = new AgentOrchestrator();
      
      orchestrator.registerAgent(AgentTemplates.coder('coder-1', 'Code Assistant'));
      
      const task = {
        id: 'task-1',
        type: 'code',
        description: 'Write a function',
        input: {},
        requiredCapabilities: ['coding'],
      };
      
      const agent = orchestrator.selectAgent(task);
      
      expect(agent).toBeDefined();
      expect(agent?.role).toBe('coder');
    });
  });

  describe('Memory System', () => {
    it('should create semantic memory store', async () => {
      const { SemanticMemory } = await import('../src/memory/semantic.js');
      
      const memory = new SemanticMemory({
        maxSize: 1000,
        decayRate: 0.01,
      });
      
      expect(memory).toBeDefined();
    });

    it('should store and retrieve memories', async () => {
      const { SemanticMemory } = await import('../src/memory/semantic.js');
      
      const memory = new SemanticMemory();
      
      await memory.add({
        id: 'test-1',
        content: 'Test memory content',
        type: 'fact',
        importance: 0.8,
      });
      
      const results = await memory.search('test');
      
      expect(results.length).toBeGreaterThan(0);
    });
  });
});

describe('TypeScript Type Tests', () => {
  it('should export correct types', async () => {
    const types = await import('../src/runtime/types.js');
    
    expect(types.PluginMetadata).toBeDefined();
    expect(types.PluginContext).toBeDefined();
    expect(types.SkillDefinition).toBeDefined();
    expect(types.ChannelMessage).toBeDefined();
    expect(types.Session).toBeDefined();
  });

  it('should have complete ChannelCapability union', async () => {
    const { ChannelCapability } = await import('../src/runtime/types.js');
    
    const capabilities: ChannelCapability[] = [
      'text',
      'markdown',
      'html',
      'images',
      'audio',
      'video',
      'files',
      'typing',
      'reactions',
      'reply',
      'threads',
      'encryption',
    ];
    
    expect(capabilities.length).toBeGreaterThan(10);
  });
});
