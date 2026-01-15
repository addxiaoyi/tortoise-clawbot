"""
Tortoise Framework - Python Usage Examples
"""

import asyncio
import json
from typing import Optional, List, Dict, Any

# Example 1: Using Tortoise HTTP API
async def api_example():
    """Basic API usage example"""
    import aiohttp
    
    base_url = "http://127.0.0.1:8080"
    
    async with aiohttp.ClientSession() as session:
        # Health check
        async with session.get(f"{base_url}/health") as resp:
            health = await resp.json()
            print(f"Gateway: {health['version']}")
        
        # Create agent
        agent_data = {
            "name": "py-agent",
            "model_provider": "openai",
            "model": "gpt-4",
            "skills": ["github"]
        }
        async with session.post(
            f"{base_url}/api/v1/agents",
            json=agent_data
        ) as resp:
            result = await resp.json()
            agent_id = result["id"]
            print(f"Created agent: {agent_id}")
        
        # List agents
        async with session.get(f"{base_url}/api/v1/agents") as resp:
            result = await resp.json()
            for agent in result["agents"]:
                print(f"  - {agent['name']}: {agent['state']}")
        
        # Store memory
        memory_data = {
            "key": "python-example",
            "value": {"message": "Hello from Python!"},
            "memory_type": "episodic"
        }
        async with session.post(
            f"{base_url}/api/v1/memory",
            json=memory_data
        ) as resp:
            print(f"Memory stored: {resp.status}")
        
        # List MCP tools
        async with session.get(f"{base_url}/api/v1/mcp/tools") as resp:
            result = await resp.json()
            print(f"Available tools: {len(result['tools'])}")

# Example 2: Using Tortoise WebSocket
async def websocket_example():
    """WebSocket event handling"""
    import aiohttp
    
    ws_url = "ws://127.0.0.1:8080/ws"
    
    async with aiohttp.ClientSession() as session:
        async with session.ws_connect(ws_url) as ws:
            # Subscribe to events
            await ws.send_json({
                "type": "subscribe",
                "data": {
                    "events": ["agent:created", "agent:stateChanged"]
                }
            })
            
            # Receive confirmation
            msg = await ws.receive_json()
            print(f"Subscribed: {msg}")
            
            # Send ping
            await ws.send_json({"type": "ping"})
            
            # Receive pong
            msg = await ws.receive_json()
            print(f"Pong received: {msg}")
            
            # Close
            await ws.close()

# Example 3: Direct memory operations
class TortoiseMemory:
    """Memory operations wrapper"""
    
    def __init__(self, base_url: str = "http://127.0.0.1:8080"):
        self.base_url = base_url
    
    async def store(
        self,
        key: str,
        value: Any,
        memory_type: str = "episodic"
    ) -> bool:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/api/v1/memory",
                json={"key": key, "value": value, "memory_type": memory_type}
            ) as resp:
                return resp.status == 201
    
    async def get(self, key: str) -> Optional[Any]:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.base_url}/api/v1/memory/{key}"
            ) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    return data["value"]
                return None
    
    async def search(self, query: str, limit: int = 10) -> List[Dict]:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/api/v1/memory/search",
                json={"query": query, "limit": limit}
            ) as resp:
                data = await resp.json()
                return data.get("results", [])
    
    async def stats(self) -> Dict[str, int]:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.base_url}/api/v1/memory"
            ) as resp:
                return await resp.json()

# Example 4: Agent management
class TortoiseAgent:
    """Agent operations wrapper"""
    
    def __init__(self, base_url: str = "http://127.0.0.1:8080"):
        self.base_url = base_url
    
    async def create(
        self,
        name: str,
        model_provider: str = "openai",
        model: str = "gpt-4",
        skills: List[str] = None
    ) -> str:
        import aiohttp
        
        if skills is None:
            skills = []
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/api/v1/agents",
                json={
                    "name": name,
                    "model_provider": model_provider,
                    "model": model,
                    "skills": skills
                }
            ) as resp:
                data = await resp.json()
                return data["id"]
    
    async def list(self) -> List[Dict]:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.base_url}/api/v1/agents"
            ) as resp:
                data = await resp.json()
                return data.get("agents", [])
    
    async def delete(self, agent_id: str) -> bool:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.delete(
                f"{self.base_url}/api/v1/agents/{agent_id}"
            ) as resp:
                return resp.status == 204

# Example 5: Mesh networking
class TortoiseMesh:
    """Mesh networking wrapper"""
    
    def __init__(self, base_url: str = "http://127.0.0.1:8080"):
        self.base_url = base_url
    
    async def list_nodes(self) -> List[Dict]:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.base_url}/api/v1/mesh/nodes"
            ) as resp:
                data = await resp.json()
                return data.get("nodes", [])
    
    async def connect(self, address: str) -> bool:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/api/v1/mesh/connect",
                json={"address": address}
            ) as resp:
                return resp.status == 200
    
    async def delegate(
        self,
        node_id: str,
        task: str,
        priority: str = "normal"
    ) -> bool:
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/api/v1/mesh/delegate",
                json={
                    "node_id": node_id,
                    "task": task,
                    "priority": priority
                }
            ) as resp:
                return resp.status == 200

# Main
if __name__ == "__main__":
    print("Tortoise Framework - Python Examples")
    print("================================\n")
    
    # Run examples
    asyncio.run(api_example())
