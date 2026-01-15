"""
Tortoise Python SDK - Channel Messaging Example
"""

import asyncio
from tortoise import TortoiseClient


async def main():
    async with TortoiseClient(
        gateway_url="http://localhost:8080",
        api_key="your-api-key"
    ) as client:
        # List available channels
        channels = await client.list_channels()
        print("Available channels:")
        for ch in channels:
            status = "enabled" if ch.get("enabled") else "disabled"
            print(f"  - {ch['name']}: {status}")
        
        # Send message via Telegram
        print("\nSending message via Telegram...")
        result = await client.send_channel_message(
            channel="telegram",
            recipient="123456789",  # Chat ID
            message="Hello from Tortoise SDK!"
        )
        print(f"Message sent: {result}")


if __name__ == "__main__":
    asyncio.run(main())
