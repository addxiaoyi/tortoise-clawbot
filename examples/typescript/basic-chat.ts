// Tortoise - Quick Start Examples
// TypeScript/JavaScript

import { TortoiseClient } from '@tortoise/sdk';

async function main() {
  // Create client
  const client = new TortoiseClient({
    apiKey: process.env.TORTOISE_API_KEY,
    baseUrl: 'http://localhost:18792'
  });

  try {
    // Connect
    await client.connect();
    console.log('Connected to Tortoise!');

    // Create session
    const session = await client.sessions.create({
      userId: 'user@example.com',
      config: {
        model: 'gpt-4o',
        temperature: 0.7
      }
    });
    console.log('Session created:', session.id);

    // Send message
    const response = await client.messages.send(session.id, {
      content: 'Hello! What can you help me with?'
    });
    console.log('Response:', response.content);

    // List sessions
    const { sessions } = await client.sessions.list();
    console.log('All sessions:', sessions.length);

  } finally {
    await client.disconnect();
  }
}

main().catch(console.error);
