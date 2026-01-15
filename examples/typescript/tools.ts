// Tool usage example

import { TortoiseClient } from '@tortoise/sdk';

async function toolExample() {
  const client = new TortoiseClient({
    baseUrl: 'http://localhost:18792'
  });

  await client.connect();

  // List available tools
  const { tools } = await client.tools.list();
  console.log('Available tools:');
  tools.forEach(tool => {
    console.log(`  - ${tool.name}: ${tool.description}`);
  });

  // Invoke a tool directly
  const result = await client.tools.invoke('calculator', {
    arguments: { expression: '2 + 2 * 3' }
  });
  console.log('\nCalculator result:', result.result);

  // Send message with tool
  const session = await client.sessions.create();
  
  const response = await client.messages.send(session.id, {
    content: 'What is 15% of 200? Use the calculator.',
    tools: tools.filter(t => t.name === 'calculator')
  });

  console.log('\nResponse:', response.content);
  
  if (response.toolCalls) {
    console.log('Tool calls:', response.toolCalls);
  }

  await client.disconnect();
}

toolExample().catch(console.error);
