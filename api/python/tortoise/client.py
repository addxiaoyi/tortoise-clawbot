"""
Tortoise Client

Main client for interacting with the Tortoise agent.
"""

import asyncio
import json
from typing import List, Optional, Callable, Dict, Any
from dataclasses import dataclass, field
from enum import Enum

import aiohttp
import websockets

from .models import Message, ChatOptions, ThinkMode


class TortoiseClient:
    """Main client for Tortoise agent interaction."""

    def __init__(
        self,
        gateway_url: str = "http://localhost:18789",
        api_key: Optional[str] = None,
        timeout: int = 30000,
    ):
        """
        Initialize the Tortoise client.

        Args:
            gateway_url: URL of the Tortoise gateway
            api_key: Optional API key for authentication
            timeout: Request timeout in milliseconds
        """
        self.gateway_url = gateway_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout
        self._websocket: Optional[websockets.WebSocketClientProtocol] = None
        self._session: Optional[aiohttp.ClientSession] = None

    async def __aenter__(self):
        """Async context manager entry."""
        await self.connect()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.disconnect()

    async def connect(self):
        """Connect to the Tortoise gateway."""
        headers = {}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"

        # Create HTTP session
        self._session = aiohttp.ClientSession(
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=self.timeout / 1000),
        )

        # Connect WebSocket for streaming
        ws_url = self.gateway_url.replace("http://", "ws://").replace("https://", "wss://")
        self._websocket = await websockets.connect(f"{ws_url}/ws", extra_headers=headers)

    async def disconnect(self):
        """Disconnect from the Tortoise gateway."""
        if self._websocket:
            await self._websocket.close()
            self._websocket = None
        if self._session:
            await self._session.close()
            self._session = None

    async def chat(
        self,
        messages: List[Message],
        options: Optional[ChatOptions] = None,
        on_chunk: Optional[Callable[[str], None]] = None,
    ) -> str:
        """
        Send a chat message and get a response.

        Args:
            messages: List of conversation messages
            options: Chat options (thinking mode, etc.)
            on_chunk: Callback for streaming response chunks

        Returns:
            Complete response text
        """
        if not self._session:
            raise RuntimeError("Not connected. Call connect() first.")

        options = options or ChatOptions()

        payload = {
            "messages": [msg.to_dict() for msg in messages],
            "thinking_mode": options.thinking_mode.value,
            "temperature": options.temperature,
            "max_tokens": options.max_tokens,
            "stream": on_chunk is not None,
        }

        async with self._session.post(
            f"{self.gateway_url}/api/chat",
            json=payload,
        ) as response:
            if response.status != 200:
                error_text = await response.text()
                raise RuntimeError(f"Chat failed: {response.status} - {error_text}")

            if on_chunk:
                # Streaming response
                full_response = []
                async for line in response.content:
                    if line:
                        data = json.loads(line.decode())
                        if "chunk" in data:
                            chunk = data["chunk"]
                            full_response.append(chunk)
                            on_chunk(chunk)
                return "".join(full_response)
            else:
                # Non-streaming response
                data = await response.json()
                return data.get("content", "")

    async def chat_stream(
        self,
        messages: List[Message],
        options: Optional[ChatOptions] = None,
    ) -> AsyncIterator[str]:
        """
        Stream chat response as an async iterator.

        Args:
            messages: List of conversation messages
            options: Chat options

        Yields:
            Response chunks
        """
        if not self._websocket:
            raise RuntimeError("Not connected. Call connect() first.")

        options = options or ChatOptions()

        await self._websocket.send_json({
            "type": "chat",
            "messages": [msg.to_dict() for msg in messages],
            "options": options.to_dict(),
        })

        async for message in self._websocket:
            data = json.loads(message)
            if data.get("type") == "chunk":
                yield data.get("content", "")
            elif data.get("type") == "done":
                break

    async def remember(self, content: str, importance: float = 0.5) -> str:
        """
        Store a memory.

        Args:
            content: Content to remember
            importance: Importance score (0.0 - 1.0)

        Returns:
            Memory ID
        """
        if not self._session:
            raise RuntimeError("Not connected. Call connect() first.")

        async with self._session.post(
            f"{self.gateway_url}/api/memory",
            json={"content": content, "importance": importance},
        ) as response:
            data = await response.json()
            return data.get("id", "")

    async def recall(self, query: str, limit: int = 10) -> List[Dict[str, Any]]:
        """
        Recall relevant memories.

        Args:
            query: Query string
            limit: Maximum number of results

        Returns:
            List of memory items
        """
        if not self._session:
            raise RuntimeError("Not connected. Call connect() first.")

        async with self._session.get(
            f"{self.gateway_url}/api/memory",
            params={"query": query, "limit": limit},
        ) as response:
            data = await response.json()
            return data.get("memories", [])

    async def memory_stats(self) -> Dict[str, int]:
        """
        Get memory statistics.

        Returns:
            Dictionary with memory counts
        """
        if not self._session:
            raise RuntimeError("Not connected. Call connect() first.")

        async with self._session.get(f"{self.gateway_url}/api/memory/stats") as response:
            data = await response.json()
            return {
                "short_term": data.get("short_term", 0),
                "medium_term": data.get("medium_term", 0),
                "long_term": data.get("long_term", 0),
            }

    async def status(self) -> Dict[str, Any]:
        """
        Get Tortoise status.

        Returns:
            Status information
        """
        if not self._session:
            raise RuntimeError("Not connected. Call connect() first.")

        async with self._session.get(f"{self.gateway_url}/api/status") as response:
            return await response.json()


# Type alias for async iterator
from typing import AsyncIterator
