// WebSocket real-time example

import { TortoiseClient } from '@tortoise/sdk';

async function websocketExample() {
  const client = new TortoiseClient({
    baseUrl: 'http://localhost:18792'
  });

  await client.connect();

  // Create WebSocket connection
  const ws = await client.createWebSocket();
  await ws.connect();

  console.log('WebSocket connected!');

  // Listen for events
  ws.on('message', (event) => {
    console.log('Received:', event.type, event.data);
  });

  ws.on('connected', () => {
    console.log('WebSocket ready!');
  });

  ws.on('disconnected', () => {
    console.log('WebSocket disconnected');
  });

  // Send a message
  ws.send({
    type: 'request',
    sessionId: 'demo-session',
    content: 'Hello via WebSocket!'
  });

  // Keep connection alive for 30 seconds
  await new Promise(resolve => setTimeout(resolve, 30000));

  ws.disconnect();
  await client.disconnect();
}

websocketExample().catch(console.error);
