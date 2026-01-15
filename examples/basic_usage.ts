/**
 * Tortoise Framework - Usage Examples
 */

// ============ Basic Agent Usage ============

import { TortoiseClient } from '@tortoise/sdk';

async function basicAgentExample() {
  const client = new TortoiseClient();
  
  // Connect to gateway
  await client.connect();
  
  // Create an agent
  const agentId = await client.agents.create({
    name: 'assistant',
    modelProvider: 'openai',
    model: 'gpt-4',
    skills: ['code-review', 'github'],
  });
  
  console.log('Created agent:', agentId);
  
  // List agents
  const agents = await client.agents.list();
  console.log('Agents:', agents);
  
  // Send message
  const response = await client.agents.sendMessage(agentId, 'Hello!');
  console.log('Response:', response);
  
  // Delete agent
  await client.agents.delete(agentId);
  
  client.disconnect();
}

// ============ Memory Example ============

async function memoryExample() {
  const client = new TortoiseClient();
  await client.connect();
  
  // Store episodic memory (short-term)
  await client.memory.store(
    'conversation-start',
    { message: 'User started a conversation', timestamp: Date.now() },
    'episodic'
  );
  
  // Store semantic memory (knowledge)
  await client.memory.store(
    'user-preferences',
    { theme: 'dark', language: 'en' },
    'semantic'
  );
  
  // Store procedural memory (skills/workflows)
  await client.memory.store(
    'build-workflow',
    { steps: ['lint', 'test', 'build', 'deploy'] },
    'procedural'
  );
  
  // Recall a memory
  const prefs = await client.memory.get('user-preferences');
  console.log('Preferences:', prefs);
  
  // Search memories
  const results = await client.memory.search('preferences', 10);
  console.log('Search results:', results);
  
  // Get memory stats
  const stats = await client.memory.stats();
  console.log('Memory stats:', stats);
  
  client.disconnect();
}

// ============ MCP Tools Example ============

async function mcpToolsExample() {
  const client = new TortoiseClient();
  await client.connect();
  
  // List available tools
  const tools = await client.mcp.list();
  console.log('Available tools:', tools);
  
  // Call a tool
  const result = await client.mcp.call('tortoise_ping', {});
  console.log('Ping result:', result);
  
  // Call tool with arguments
  const ghResult = await client.mcp.call('github_list_issues', {
    owner: 'openclaw',
    repo: 'openclaw',
    limit: 10,
  });
  console.log('GitHub issues:', ghResult);
  
  client.disconnect();
}

// ============ Multi-Agent Mesh Example ============

async function meshExample() {
  const client = new TortoiseClient();
  await client.connect();
  
  // Create multiple agents
  const agent1 = await client.agents.create({
    name: 'coordinator',
    modelProvider: 'openai',
    model: 'gpt-4',
  });
  
  const agent2 = await client.agents.create({
    name: 'researcher',
    modelProvider: 'openai',
    model: 'gpt-4',
  });
  
  // Connect to mesh nodes
  await client.mesh.connect('192.168.1.100:8080');
  await client.mesh.connect('192.168.1.101:8080');
  
  // List nodes
  const nodes = await client.mesh.listNodes();
  console.log('Mesh nodes:', nodes);
  
  // Delegate task to another node
  await client.mesh.delegate({
    nodeId: nodes[0].id,
    task: 'Research latest AI developments',
    priority: 'high',
  });
  
  // Collaborate on task
  await client.mesh.collaborate(
    [nodes[0].id, nodes[1].id],
    'Analyze codebase for security issues'
  );
  
  client.disconnect();
}

// ============ Event Handling Example ============

async function eventExample() {
  const client = new TortoiseClient();
  
  // Event handlers
  client.on('connected', () => {
    console.log('Connected to gateway');
  });
  
  client.on('disconnected', () => {
    console.log('Disconnected from gateway');
  });
  
  client.on('error', (error) => {
    console.error('Error:', error);
  });
  
  client.on('agent:created', (data) => {
    console.log('Agent created:', data);
  });
  
  client.on('agent:stateChanged', (data) => {
    console.log('Agent state changed:', data);
  });
  
  client.on('message', (msg) => {
    console.log('Received message:', msg);
  });
  
  await client.connect();
  
  // Do something...
  
  client.disconnect();
}

// ============ Plugin Example ============

import { PluginLoader } from '@tortoise/sdk';

async function pluginExample() {
  const loader = new PluginLoader();
  
  // Load a plugin
  const plugin = await loader.load('./plugins/discord-channel');
  
  // Get plugin manifest
  console.log('Plugin:', plugin.manifest);
  
  // Enable plugin
  await plugin.enable();
  
  // Configure plugin
  await plugin.configure({
    token: 'your-discord-bot-token',
    channels: ['general', 'ai-talk'],
  });
  
  // Use plugin
  await plugin.sendMessage({
    channel: 'general',
    content: 'Hello from Tortoise!',
  });
  
  // Disable plugin
  await plugin.disable();
}

// Run examples
console.log('Tortoise Framework Examples');
console.log('========================');
