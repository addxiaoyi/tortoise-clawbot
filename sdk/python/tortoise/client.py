"""
Tortoise Client
Main client for interacting with Tortoise Gateway.
"""

import asyncio
import json
from typing import Optional, AsyncIterator, Callable, Any
from dataclasses import dataclass, field
from datetime import datetime
import aiohttp

from tortoise.models import ChatMessage, Session
from tortoise.exceptions import TortoiseError, AuthError


@dataclass
class ClientConfig:
    """Client configuration."""
    gateway_url: str = "http://localhost:8080"
    api_key: Optional[str] = None
    timeout: float = 60.0
    max_retries: int = 3
    auto_reconnect: bool = True
    model: str = "gpt-4"
    
    # WebSocket
    enable_websocket: bool = True
    reconnect_delay: float = 1.0
    max_reconnect_attempts: int = 5


class TortoiseClient:
    """
    Main client for Tortoise Gateway.
    
    Example:
        async with TortoiseClient("http://localhost:8080", api_key="...") as client:
            async for response in client.chat_stream("Hello!"):
                print(response.content)
    """
    
    def __init__(self, config: Optional[ClientConfig] = None, **kwargs):
        self.config = config or ClientConfig()
        
        # Apply kwargs to config
        for key, value in kwargs.items():
            if hasattr(self.config, key):
                setattr(self.config, key, value)
        
        self._session: Optional[aiohttp.ClientSession] = None
        self._websocket: Optional[aiohttp.ClientWebSocketResponse] = None
        self._connected = False
        self._message_queue: asyncio.Queue = asyncio.Queue()
        self._listen_task: Optional[asyncio.Task] = None
    
    async def __aenter__(self) -> "TortoiseClient":
        await self.connect()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.disconnect()
    
    async def connect(self) -> None:
        """Connect to the Tortoise Gateway."""
        if self._connected:
            return
        
        headers = {}
        if self.config.api_key:
            headers["Authorization"] = f"Bearer {self.config.api_key}"
        
        self._session = aiohttp.ClientSession(
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=self.config.timeout)
        )
        
        # Verify connection
        try:
            async with self._session.get(f"{self.config.gateway_url}/health") as resp:
                if resp.status != 200:
                    raise TortoiseError(f"Gateway returned status {resp.status}")
        except aiohttp.ClientError as e:
            raise TortoiseError(f"Failed to connect to gateway: {e}")
        
        # Start WebSocket listener if enabled
        if self.config.enable_websocket:
            await self._start_websocket()
        
        self._connected = True
    
    async def disconnect(self) -> None:
        """Disconnect from the Tortoise Gateway."""
        if self._listen_task:
            self._listen_task.cancel()
            try:
                await self._listen_task
            except asyncio.CancelledError:
                pass
        
        if self._websocket:
            await self._websocket.close()
            self._websocket = None
        
        if self._session:
            await self._session.close()
            self._session = None
        
        self._connected = False
    
    async def _start_websocket(self) -> None:
        """Start WebSocket connection for real-time events."""
        ws_url = self.config.gateway_url.replace("http", "ws") + "/ws"
        
        self._websocket = await self._session.ws_connect(ws_url)
        self._listen_task = asyncio.create_task(self._listen_websocket())
    
    async def _listen_websocket(self) -> None:
        """Listen for WebSocket messages."""
        try:
            async for msg in self._websocket:
                if msg.type == aiohttp.WSMsgType.TEXT:
                    try:
                        data = json.loads(msg.data)
                        await self._message_queue.put(data)
                    except json.JSONDecodeError:
                        pass
                elif msg.type == aiohttp.WSMsgType.ERROR:
                    break
        except asyncio.CancelledError:
            pass
    
    @property
    def is_connected(self) -> bool:
        """Check if client is connected."""
        return self._connected
    
    # ==============================
    # Chat API
    # ==============================
    
    async def chat(
        self,
        message: str,
        model: Optional[str] = None,
        session_id: Optional[str] = None,
        stream: bool = False,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        **kwargs
    ) -> ChatMessage:
        """
        Send a chat message and get a response.
        
        Args:
            message: The user message
            model: Model to use (default: config.model)
            session_id: Session ID for conversation context
            stream: Whether to stream the response
            temperature: Response randomness (0-2)
            max_tokens: Maximum tokens in response
        
        Returns:
            ChatMessage object with the response
        """
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        model = model or self.config.model
        
        payload = {
            "model": model,
            "messages": [{"role": "user", "content": message}],
            "temperature": temperature,
            "stream": stream,
        }
        
        if session_id:
            payload["session_id"] = session_id
        if max_tokens:
            payload["max_tokens"] = max_tokens
        
        payload.update(kwargs)
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/chat/completions",
            json=payload
        ) as resp:
            if resp.status == 401:
                raise AuthError("Invalid API key")
            if resp.status != 200:
                raise TortoiseError(f"Chat failed: {resp.status}")
            
            data = await resp.json()
            return ChatMessage.from_dict(data)
    
    async def chat_stream(
        self,
        message: str,
        model: Optional[str] = None,
        session_id: Optional[str] = None,
        temperature: float = 0.7,
        **kwargs
    ) -> AsyncIterator[str]:
        """
        Send a chat message and stream the response.
        
        Yields:
            String chunks of the response
        """
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        model = model or self.config.model
        
        payload = {
            "model": model,
            "messages": [{"role": "user", "content": message}],
            "temperature": temperature,
            "stream": True,
        }
        
        if session_id:
            payload["session_id"] = session_id
        
        payload.update(kwargs)
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/chat/completions",
            json=payload
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Stream failed: {resp.status}")
            
            async for line in resp.content:
                line = line.decode("utf-8").strip()
                if line.startswith("data: "):
                    if line == "data: [DONE]":
                        break
                    data = json.loads(line[6:])
                    if "choices" in data and len(data["choices"]) > 0:
                        delta = data["choices"][0].get("delta", {})
                        content = delta.get("content", "")
                        if content:
                            yield content
    
    # ==============================
    # Session Management
    # ==============================
    
    async def create_session(
        self,
        title: Optional[str] = None,
        model: Optional[str] = None,
        **kwargs
    ) -> Session:
        """Create a new chat session."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        payload = kwargs
        if title:
            payload["title"] = title
        if model:
            payload["model"] = model
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/sessions",
            json=payload
        ) as resp:
            if resp.status != 201:
                raise TortoiseError(f"Failed to create session: {resp.status}")
            
            data = await resp.json()
            return Session.from_dict(data)
    
    async def get_sessions(self) -> list[Session]:
        """Get all chat sessions."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.get(
            f"{self.config.gateway_url}/api/v1/sessions"
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Failed to get sessions: {resp.status}")
            
            data = await resp.json()
            return [Session.from_dict(s) for s in data.get("items", [])]
    
    async def get_session(self, session_id: str) -> Session:
        """Get a specific session by ID."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.get(
            f"{self.config.gateway_url}/api/v1/sessions/{session_id}"
        ) as resp:
            if resp.status == 404:
                raise TortoiseError(f"Session not found: {session_id}")
            if resp.status != 200:
                raise TortoiseError(f"Failed to get session: {resp.status}")
            
            data = await resp.json()
            return Session.from_dict(data)
    
    async def delete_session(self, session_id: str) -> None:
        """Delete a session."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.delete(
            f"{self.config.gateway_url}/api/v1/sessions/{session_id}"
        ) as resp:
            if resp.status != 204:
                raise TortoiseError(f"Failed to delete session: {resp.status}")
    
    # ==============================
    # Memory
    # ==============================
    
    async def memory_search(
        self,
        query: str,
        limit: int = 10,
        session_id: Optional[str] = None
    ) -> list[dict]:
        """Search memory/knowledge base."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        payload = {"query": query, "limit": limit}
        if session_id:
            payload["session_id"] = session_id
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/memory/search",
            json=payload
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Memory search failed: {resp.status}")
            
            data = await resp.json()
            return data.get("results", [])
    
    async def memory_add(
        self,
        content: str,
        metadata: Optional[dict] = None,
        memory_type: str = "fact"
    ) -> dict:
        """Add a memory entry."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        payload = {"content": content, "type": memory_type}
        if metadata:
            payload["metadata"] = metadata
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/memory",
            json=payload
        ) as resp:
            if resp.status != 201:
                raise TortoiseError(f"Failed to add memory: {resp.status}")
            
            return await resp.json()
    
    # ==============================
    # Plugins
    # ==============================
    
    async def list_plugins(self) -> list[dict]:
        """List available plugins."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.get(
            f"{self.config.gateway_url}/api/v1/plugins"
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Failed to list plugins: {resp.status}")
            
            data = await resp.json()
            return data.get("plugins", [])
    
    async def invoke_plugin(
        self,
        plugin_id: str,
        action: str,
        params: Optional[dict] = None
    ) -> Any:
        """Invoke a plugin action."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        payload = {"action": action}
        if params:
            payload["params"] = params
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/plugins/{plugin_id}/invoke",
            json=payload
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Plugin invocation failed: {resp.status}")
            
            return await resp.json()
    
    # ==============================
    # Channels
    # ==============================
    
    async def list_channels(self) -> list[dict]:
        """List available message channels."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.get(
            f"{self.config.gateway_url}/api/v1/channels"
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Failed to list channels: {resp.status}")
            
            data = await resp.json()
            return data.get("channels", [])
    
    async def send_channel_message(
        self,
        channel: str,
        recipient: str,
        message: str,
        **kwargs
    ) -> dict:
        """Send a message through a channel."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        payload = {"to": recipient, "content": message, **kwargs}
        
        async with self._session.post(
            f"{self.config.gateway_url}/api/v1/channels/{channel}/send",
            json=payload
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Failed to send message: {resp.status}")
            
            return await resp.json()
    
    # ==============================
    # Events
    # ==============================
    
    async def on_event(
        self,
        event_type: str,
        callback: Callable[[dict], None]
    ) -> None:
        """
        Register an event callback.
        
        Args:
            event_type: Type of event (e.g., "message", "typing", "channel_update")
            callback: Async function to call when event occurs
        """
        if not self._connected:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async def wrapper(data: dict) -> None:
            if data.get("type") == event_type:
                await callback(data)
        
        # Start listener if not already running
        if not self._listen_task:
            self._listen_task = asyncio.create_task(self._listen_websocket())
    
    # ==============================
    # Health & Stats
    # ==============================
    
    async def health_check(self) -> dict:
        """Check gateway health status."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.get(
            f"{self.config.gateway_url}/health"
        ) as resp:
            if resp.status != 200:
                return {"status": "unhealthy", "code": resp.status}
            
            return await resp.json()
    
    async def get_stats(self) -> dict:
        """Get gateway statistics."""
        if not self._session:
            raise TortoiseError("Not connected. Call connect() first.")
        
        async with self._session.get(
            f"{self.config.gateway_url}/api/v1/stats"
        ) as resp:
            if resp.status != 200:
                raise TortoiseError(f"Failed to get stats: {resp.status}")
            
            return await resp.json()
