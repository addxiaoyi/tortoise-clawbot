"""
Tortoise SDK - Python SDK
"""

import asyncio
import aiohttp
import json
from typing import Optional, Dict, Any, List, AsyncGenerator


class TortoiseClient:
    """Tortoise AI Agent Python SDK"""
    
    def __init__(
        self,
        base_url: str = "http://localhost:18792",
        api_key: Optional[str] = None
    ):
        self.base_url = base_url
        self.api_key = api_key
        self.session_id: Optional[str] = None
        self._session: Optional[aiohttp.ClientSession] = None
    
    async def _get_session(self) -> aiohttp.ClientSession:
        if self._session is None or self._session.closed:
            headers = {"Content-Type": "application/json"}
            if self.api_key:
                headers["Authorization"] = f"Bearer {self.api_key}"
            self._session = aiohttp.ClientSession(headers=headers)
        return self._session
    
    async def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict] = None
    ) -> Dict[str, Any]:
        session = await self._get_session()
        url = f"{self.base_url}{endpoint}"
        
        async with session.request(method, url, json=data) as response:
            if not response.ok:
                raise Exception(f"HTTP error! status: {response.status}")
            return await response.json()
    
    async def close(self):
        """Close the client session"""
        if self._session and not self._session.closed:
            await self._session.close()
    
    # Session Management
    async def create_session(
        self,
        user_id: str,
        metadata: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """Create a new session"""
        return await self._request(
            "POST",
            "/sessions",
            {"user_id": user_id, "metadata": metadata or {}}
        )
    
    async def get_session(self, session_id: str) -> Dict[str, Any]:
        """Get session by ID"""
        return await self._request("GET", f"/sessions/{session_id}")
    
    async def delete_session(self, session_id: str) -> None:
        """Delete a session"""
        await self._request("DELETE", f"/sessions/{session_id}")
    
    async def list_sessions(self, user_id: Optional[str] = None) -> Dict[str, Any]:
        """List all sessions"""
        endpoint = f"/sessions?user_id={user_id}" if user_id else "/sessions"
        return await self._request("GET", endpoint)
    
    # Messages
    async def send_message(
        self,
        session_id: str,
        content: str,
        msg_type: str = "text",
        format: str = "plain",
        stream: bool = False
    ) -> Dict[str, Any]:
        """Send a message"""
        self.session_id = session_id
        return await self._request(
            "POST",
            "/messages",
            {
                "session_id": session_id,
                "content": content,
                "type": msg_type,
                "format": format,
                "stream": stream
            }
        )
    
    async def send_message_stream(
        self,
        session_id: str,
        content: str
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """Send a message with streaming response"""
        session = await self._get_session()
        url = f"{self.base_url}/messages/stream"
        
        async with session.post(
            url,
            json={"session_id": session_id, "content": content, "stream": True}
        ) as response:
            if not response.ok:
                raise Exception(f"HTTP error! status: {response.status}")
            
            async for line in response.content:
                line = line.decode("utf-8").strip()
                if line:
                    yield json.loads(line)
    
    async def get_messages(
        self,
        session_id: str,
        limit: int = 50,
        offset: int = 0
    ) -> Dict[str, Any]:
        """Get messages for a session"""
        return await self._request(
            "GET",
            f"/sessions/{session_id}/messages?limit={limit}&offset={offset}"
        )
    
    # Tools
    async def list_tools(self) -> List[Dict[str, Any]]:
        """List all available tools"""
        return await self._request("GET", "/tools")
    
    async def execute_tool(
        self,
        plugin_id: str,
        tool_name: str,
        arguments: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Execute a tool"""
        return await self._request(
            "POST",
            "/tools/execute",
            {
                "plugin_id": plugin_id,
                "tool_name": tool_name,
                "arguments": json.dumps(arguments)
            }
        )
    
    # Memory
    async def save_memory(
        self,
        memory_type: str,
        content: str,
        importance: float = 0.5,
        metadata: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """Save a memory"""
        return await self._request(
            "POST",
            "/memory",
            {
                "type": memory_type,
                "content": content,
                "importance": importance,
                "metadata": metadata or {}
            }
        )
    
    async def query_memory(
        self,
        query: str,
        memory_type: Optional[str] = None,
        limit: int = 10,
        similarity_threshold: float = 0.7
    ) -> Dict[str, Any]:
        """Query memories"""
        return await self._request(
            "POST",
            "/memory/query",
            {
                "query": query,
                "type": memory_type,
                "limit": limit,
                "similarity_threshold": similarity_threshold
            }
        )
    
    async def delete_memory(self, memory_id: str) -> None:
        """Delete a memory"""
        await self._request("DELETE", f"/memory/{memory_id}")
    
    # Plugins
    async def list_plugins(self) -> List[Dict[str, Any]]:
        """List all plugins"""
        return await self._request("GET", "/plugins")
    
    async def install_plugin(
        self,
        source: str,
        config: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """Install a plugin"""
        return await self._request(
            "POST",
            "/plugins/install",
            {"source": source, "config": config or {}}
        )
    
    async def uninstall_plugin(self, plugin_id: str, force: bool = False) -> None:
        """Uninstall a plugin"""
        await self._request(
            "DELETE",
            f"/plugins/{plugin_id}?force={force}"
        )
    
    # Channels
    async def list_channels(self) -> List[Dict[str, Any]]:
        """List all channels"""
        return await self._request("GET", "/channels")
    
    async def connect_channel(
        self,
        channel_type: str,
        credentials: Dict[str, str],
        config: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """Connect a channel"""
        return await self._request(
            "POST",
            "/channels/connect",
            {"type": channel_type, "credentials": credentials, "config": config or {}}
        )
    
    async def disconnect_channel(self, channel_id: str) -> None:
        """Disconnect a channel"""
        await self._request("DELETE", f"/channels/{channel_id}")
    
    # Configuration
    async def get_config(self) -> Dict[str, Any]:
        """Get configuration"""
        return await self._request("GET", "/config")
    
    async def update_config(self, config: Dict[str, Any]) -> None:
        """Update configuration"""
        await self._request("PUT", "/config", config)
    
    # Health
    async def health_check(self) -> Dict[str, Any]:
        """Check health status"""
        return await self._request("GET", "/health")
    
    # Metrics
    async def get_metrics(self) -> Dict[str, Any]:
        """Get system metrics"""
        return await self._request("GET", "/metrics")


# Synchronous wrapper
class TortoiseClientSync:
    """Synchronous wrapper for TortoiseClient"""
    
    def __init__(
        self,
        base_url: str = "http://localhost:18792",
        api_key: Optional[str] = None
    ):
        self.client = TortoiseClient(base_url, api_key)
    
    def create_session(self, user_id: str, metadata: Optional[Dict[str, str]] = None):
        return asyncio.run(self.client.create_session(user_id, metadata))
    
    def close(self):
        asyncio.run(self.client.close())
    
    # Add more sync methods as needed...


__all__ = ["TortoiseClient", "TortoiseClientSync"]
