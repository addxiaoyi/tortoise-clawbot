# Tortoise Python SDK

Python SDK for Tortoise AI Agent Framework.

## Installation

```bash
pip install tortoise-sdk
```

## Quick Start

```python
from tortoise import Tortoise, Message, Agent

# Initialize
tortoise = Tortoise(api_key="your-api-key")

# Create session
session = tortoise.create_session()

# Send message
response = session.chat("Hello, Tortoise!")
print(response.content)

# List agents
agents = tortoise.list_agents()
for agent in agents:
    print(f"{agent.name}: {agent.status}")
```

## Features

- Session management
- Multi-agent orchestration
- Plugin marketplace
- Memory system
- Channel integration
- Enterprise authentication

## Documentation

See [docs/](docs/) for full documentation.
