// Streaming response example

import { TortoiseClient } from '@tortoise/sdk';

async function streamChat() {
  const client = new TortoiseClient({
    baseUrl: 'http://localhost:18792'
  });

  await client.connect();

  const session = await client.sessions.create();

  console.log('Starting streaming chat...\n');

  const stream = await client.messages.send(session.id, {
    content: 'Write a short poem about artificial intelligence',
    stream: true
  });

  for await (const event of stream) {
    switch (event.type) {
      case 'message_start':
        console.log('Message started:', event.data.messageId);
        break;
      case 'content_chunk':
        process.stdout.write(event.data.delta);
        break;
      case 'tool_call':
        console.log('\n\n[Tool call detected]');
        console.log('Tool:', event.data.name);
        console.log('Arguments:', JSON.stringify(event.data.arguments));
        break;
      case 'message_end':
        console.log('\n\n[Message complete]');
        console.log('Metadata:', JSON.stringify(event.data.metadata));
        break;
    }
  }

  await client.disconnect();
}

streamChat().catch(console.error);
