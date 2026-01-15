# Tortoise TypeScript SDK

## 安装

```bash
npm install @tortoise/sdk
```

## 快速开始

```typescript
import { TortoiseClient, Session } from '@tortoise/sdk';

// 创建客户端
const client = new TortoiseClient({
  apiKey: process.env.TORTOISE_API_KEY,
  baseUrl: 'http://localhost:18792'
});

// 连接
await client.connect();

// 创建会话
const session = await client.sessions.create({
  userId: 'user_123'
});

// 发送消息
const response = await client.messages.send(session.id, {
  content: 'Hello, Tortoise!'
});

console.log(response.content);
```

## 完整示例

### 流式响应

```typescript
import { TortoiseClient } from '@tortoise/sdk';

const client = new TortoiseClient({
  apiKey: process.env.TORTOISE_API_KEY
});

const stream = await client.messages.send('session_id', {
  content: 'Write a story',
  stream: true
});

for await (const event of stream) {
  switch (event.type) {
    case 'message_start':
      console.log('Message started:', event.messageId);
      break;
    case 'content_chunk':
      process.stdout.write(event.delta);
      break;
    case 'tool_call':
      console.log('Tool call:', event.name, event.arguments);
      break;
    case 'message_end':
      console.log('\nMessage complete');
      break;
  }
}
```

### 工具调用

```typescript
// 定义可用工具
const tools = [
  {
    name: 'web_search',
    description: 'Search the web',
    parameters: {
      type: 'object',
      properties: {
        query: { type: 'string' }
      },
      required: ['query']
    }
  }
];

// 发送带工具的消息
const response = await client.messages.send(session.id, {
  content: 'Search for news about AI',
  tools
});

// 处理工具调用
if (response.toolCalls) {
  for (const call of response.toolCalls) {
    if (call.name === 'web_search') {
      const result = await client.tools.invoke('web_search', {
        arguments: call.arguments
      });
      console.log('Search result:', result);
    }
  }
}
```

### 会话管理

```typescript
// 列出所有会话
const { sessions } = await client.sessions.list({
  limit: 20
});

// 获取会话历史
const messages = await client.messages.list(session.id, {
  before: 'msg_123',
  limit: 50
});

// 删除会话
await client.sessions.delete(session.id);
```

### 记忆系统

```typescript
// 存储记忆
await client.memory.store({
  sessionId: session.id,
  content: 'User prefers dark mode',
  type: 'fact',
  tags: ['preference', 'ui'],
  importance: 0.8
});

// 搜索记忆
const results = await client.memory.search({
  query: 'user interface preferences',
  limit: 5
});
```

## API 参考

### TortoiseClient

```typescript
class TortoiseClient {
  constructor(options: ClientOptions);
  
  // 连接
  connect(): Promise<void>;
  disconnect(): Promise<void>;
  
  // 会话
  sessions: SessionManager;
  
  // 消息
  messages: MessageManager;
  
  // 工具
  tools: ToolManager;
  
  // 记忆
  memory: MemoryManager;
}
```

### ClientOptions

```typescript
interface ClientOptions {
  apiKey?: string;
  baseUrl: string;
  timeout?: number;
  headers?: Record<string, string>;
}
```

### SessionManager

```typescript
class SessionManager {
  create(options: CreateSessionOptions): Promise<Session>;
  get(sessionId: string): Promise<Session>;
  list(options?: ListSessionsOptions): Promise<ListSessionsResponse>;
  update(sessionId: string, updates: UpdateSessionOptions): Promise<Session>;
  delete(sessionId: string): Promise<void>;
}
```

### MessageManager

```typescript
class MessageManager {
  send(sessionId: string, options: SendMessageOptions): Promise<Message>;
  list(sessionId: string, options?: ListMessagesOptions): Promise<ListMessagesResponse>;
  stream(sessionId: string, options: SendMessageOptions): AsyncIterableStream<MessageEvent>;
}
```

## 错误处理

```typescript
import { TortoiseError, RateLimitError, AuthError } from '@tortoise/sdk';

try {
  const response = await client.messages.send(sessionId, {
    content: 'Hello'
  });
} catch (error) {
  if (error instanceof AuthError) {
    console.log('Authentication failed');
  } else if (error instanceof RateLimitError) {
    console.log('Rate limited, retry after:', error.retryAfter);
  } else if (error instanceof TortoiseError) {
    console.log('Tortoise error:', error.code, error.message);
  }
}
```

## 订阅事件

```typescript
// 订阅连接状态
client.on('connected', () => console.log('Connected'));
client.on('disconnected', () => console.log('Disconnected'));
client.on('error', (error) => console.error('Error:', error));

// 订阅消息事件
client.on('message', (message) => {
  console.log('New message:', message);
});
```
