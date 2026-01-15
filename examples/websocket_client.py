"""
WebSocket client example
"""

import asyncio
from tortoise import TortoiseClient, TortoiseWebSocket


async def websocket_example():
    """WebSocket real-time example"""
    # Create HTTP client first
    client = TortoiseClient(base_url="http://localhost:18792")
    await client.connect()

    # Create WebSocket connection
    ws_url = client.base_url.replace("http", "ws") + "/ws"
    ws = TortoiseWebSocket(ws_url)

    # Event handlers
    def on_message(data):
        print(f"Received: {data}")

    def on_connected():
        print("Connected!")

    def on_disconnected():
        print("Disconnected")

    ws.on("message", on_message)
    ws.on("connected", on_connected)
    ws.on("disconnected", on_disconnected)

    # Connect
    await ws.connect()

    # Send message
    ws.send({
        "type": "request",
        "sessionId": "demo",
        "content": "Hello via WebSocket!"
    })

    # Listen for 30 seconds
    await asyncio.sleep(30)

    # Cleanup
    ws.disconnect()
    await client.disconnect()


if __name__ == "__main__":
    asyncio.run(websocket_example())
