"""
Tortoise WebSocket Client
"""

import json
import asyncio
from typing import Callable, Optional, Dict, Any
import websockets
from websockets.client import WebSocketClientProtocol

class TortoiseWebSocket:
    """
    WebSocket client for real-time Tortoise communication
    """
    
    def __init__(self, url: str, api_key: Optional[str] = None):
        self.url = url
        self.api_key = api_key
        self._ws: Optional[WebSocketClientProtocol] = None
        self._handlers: Dict[str, set] = {}
        self._running = False
    
    async def connect(self) -> None:
        """Connect to the WebSocket server"""
        self._ws = await websockets.connect(self.url)
        
        # Send handshake
        await self._ws.send(json.dumps({
            "type": "handshake",
            "apiKey": self.api_key
        }))
        
        self._running = True
        self._emit("connected")
    
    async def disconnect(self) -> None:
        """Disconnect from the WebSocket server"""
        self._running = False
        if self._ws:
            await self._ws.close()
            self._ws = None
        self._emit("disconnected")
    
    async def send(self, message: Dict[str, Any]) -> None:
        """Send a message"""
        if self._ws:
            await self._ws.send(json.dumps(message))
    
    def on(self, event: str, handler: Callable) -> None:
        """Register an event handler"""
        if event not in self._handlers:
            self._handlers[event] = set()
        self._handlers[event].add(handler)
    
    def off(self, event: str, handler: Callable) -> None:
        """Unregister an event handler"""
        if event in self._handlers:
            self._handlers[event].discard(handler)
    
    def _emit(self, event: str, data: Any = None) -> None:
        """Emit an event to all handlers"""
        if event in self._handlers:
            for handler in self._handlers[event]:
                if asyncio.iscoroutinefunction(handler):
                    asyncio.create_task(handler(data))
                else:
                    handler(data)
        
        # Also emit to wildcard handlers
        if "*" in self._handlers:
            for handler in self._handlers["*"]:
                event_data = {"type": event, "data": data}
                if asyncio.iscoroutinefunction(handler):
                    asyncio.create_task(handler(event_data))
                else:
                    handler(event_data)
    
    async def listen(self) -> None:
        """Listen for messages"""
        while self._running and self._ws:
            try:
                message = await self._ws.recv()
                data = json.loads(message)
                
                # Emit to specific event type
                self._emit(data.get("type", ""), data)
                
                # Also emit to message event
                self._emit("message", data)
                
            except websockets.ConnectionClosed:
                self._running = False
                self._emit("disconnected")
                break
            except Exception as e:
                self._emit("error", {"error": str(e)})
    
    @property
    def is_connected(self) -> bool:
        """Check if connected"""
        return self._ws is not None and self._running
